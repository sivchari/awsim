package sqs

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 is required by SQS spec for message body hash
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sivchari/kumo/internal/storage"
)

// DeduplicationEntry holds deduplication information for FIFO queues.
type DeduplicationEntry struct {
	MessageID string    `json:"messageId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Storage defines the interface for SQS storage operations.
type Storage interface {
	CreateQueue(ctx context.Context, name string, attributes map[string]string) (*Queue, error)
	DeleteQueue(ctx context.Context, queueURL string) error
	ListQueues(ctx context.Context, prefix string) ([]string, error)
	GetQueueURL(ctx context.Context, name string) (string, error)
	GetQueue(ctx context.Context, queueURL string) (*Queue, error)
	SendMessage(ctx context.Context, queueURL, body string, delaySeconds int, messageAttributes map[string]MessageAttributeValue, messageGroupID, messageDeduplicationID string) (*Message, error)
	ReceiveMessage(ctx context.Context, queueURL string, maxMessages, visibilityTimeout, waitTimeSeconds int) ([]*Message, error)
	DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error
	PurgeQueue(ctx context.Context, queueURL string) error
	GetQueueAttributes(ctx context.Context, queueURL string, attributeNames []string) (map[string]string, error)
	SetQueueAttributes(ctx context.Context, queueURL string, attributes map[string]string) error
}

// QueueError represents an SQS queue error.
type QueueError struct {
	Code    string
	Message string
}

func (e *QueueError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Attribute value for boolean true.
const attrValueTrue = "true"

// Common error codes.
var (
	ErrQueueAlreadyExists   = &QueueError{Code: "QueueAlreadyExists", Message: "A queue with this name already exists"}
	ErrQueueDoesNotExist    = &QueueError{Code: "AWS.SimpleQueueService.NonExistentQueue", Message: "The specified queue does not exist"}
	ErrReceiptHandleInvalid = &QueueError{Code: "ReceiptHandleIsInvalid", Message: "The receipt handle is not valid"}
	ErrMessageNotInflight   = &QueueError{Code: "MessageNotInflight", Message: "The message is not in flight"}
)

// Option is a configuration option for MemoryStorage.
type Option func(*MemoryStorage)

// WithDataDir enables persistent storage in the specified directory.
func WithDataDir(dir string) Option {
	return func(s *MemoryStorage) {
		s.dataDir = dir
	}
}

// Compile-time interface checks.
var (
	_ json.Marshaler   = (*MemoryStorage)(nil)
	_ json.Unmarshaler = (*MemoryStorage)(nil)
)

// MemoryStorage implements Storage using in-memory maps.
type MemoryStorage struct {
	mu      sync.RWMutex          `json:"-"`
	Queues  map[string]*QueueData `json:"queues"`
	baseURL string
	dataDir string
}

// QueueData holds all data associated with a single SQS queue.
type QueueData struct {
	Queue              *Queue                        `json:"queue"`
	Messages           []*Message                    `json:"messages"`
	Inflight           map[string]*Message           `json:"-"`               // receiptHandle -> message
	DeduplicationCache map[string]DeduplicationEntry `json:"-"`               // deduplicationID -> entry (FIFO only)
	SequenceCounter    uint64                        `json:"sequenceCounter"` // Per-queue sequence number (FIFO only)
}

// NewMemoryStorage creates a new in-memory SQS storage.
func NewMemoryStorage(baseURL string, opts ...Option) *MemoryStorage {
	s := &MemoryStorage{
		Queues:  make(map[string]*QueueData),
		baseURL: baseURL,
	}
	for _, o := range opts {
		o(s)
	}

	if s.dataDir != "" {
		_ = storage.Load(s.dataDir, "sqs", s)
	}

	return s
}

// MarshalJSON serializes the storage state to JSON.
func (s *MemoryStorage) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type Alias MemoryStorage

	data, err := json.Marshal(&struct{ *Alias }{Alias: (*Alias)(s)})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	return data, nil
}

// UnmarshalJSON restores the storage state from JSON.
func (s *MemoryStorage) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	type Alias MemoryStorage

	aux := &struct{ *Alias }{Alias: (*Alias)(s)}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	if s.Queues == nil {
		s.Queues = make(map[string]*QueueData)
	}

	return nil
}

// Close saves the storage state to disk if persistence is enabled.
func (s *MemoryStorage) Close() error {
	if s.dataDir == "" {
		return nil
	}

	if err := storage.Save(s.dataDir, "sqs", s); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// CreateQueue creates a new queue.
func (s *MemoryStorage) CreateQueue(_ context.Context, name string, attributes map[string]string) (*Queue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queueURL := fmt.Sprintf("%s/000000000000/%s", s.baseURL, name)

	if qd, exists := s.Queues[queueURL]; exists {
		return qd.Queue, nil
	}

	// Check FIFO queue requirements.
	isFifo := strings.HasSuffix(name, ".fifo")
	if attributes["FifoQueue"] == attrValueTrue && !isFifo {
		return nil, &QueueError{
			Code:    "InvalidParameterValue",
			Message: "The queue name must end with .fifo for FIFO queues",
		}
	}

	now := time.Now()
	queue := &Queue{
		Name:                      name,
		URL:                       queueURL,
		ARN:                       fmt.Sprintf("arn:aws:sqs:us-east-1:000000000000:%s", name),
		CreatedTimestamp:          now,
		LastModifiedTimestamp:     now,
		VisibilityTimeout:         30,
		MessageRetentionPeriod:    345600,
		DelaySeconds:              0,
		MaxMessageSize:            262144,
		ReceiveWaitTimeSeconds:    0,
		FifoQueue:                 isFifo,
		ContentBasedDeduplication: attributes["ContentBasedDeduplication"] == attrValueTrue,
	}

	// Apply attributes.
	applyQueueAttributes(queue, attributes)

	qd := &QueueData{
		Queue:    queue,
		Messages: make([]*Message, 0),
		Inflight: make(map[string]*Message),
	}

	if isFifo {
		qd.DeduplicationCache = make(map[string]DeduplicationEntry)
	}

	s.Queues[queueURL] = qd

	return queue, nil
}

// DeleteQueue deletes a queue.
func (s *MemoryStorage) DeleteQueue(_ context.Context, queueURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Queues[queueURL]; !exists {
		return ErrQueueDoesNotExist
	}

	delete(s.Queues, queueURL)

	return nil
}

// ListQueues lists all queues.
func (s *MemoryStorage) ListQueues(_ context.Context, prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urls := make([]string, 0, len(s.Queues))

	for url, qd := range s.Queues {
		if prefix == "" || len(qd.Queue.Name) >= len(prefix) && qd.Queue.Name[:len(prefix)] == prefix {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

// GetQueueURL gets the URL for a queue by name.
func (s *MemoryStorage) GetQueueURL(_ context.Context, name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for url, qd := range s.Queues {
		if qd.Queue.Name == name {
			return url, nil
		}
	}

	return "", ErrQueueDoesNotExist
}

// GetQueue gets a queue by URL.
func (s *MemoryStorage) GetQueue(_ context.Context, queueURL string) (*Queue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	return qd.Queue, nil
}

// FifoResult holds the result of FIFO validation and deduplication.
type FifoResult struct {
	SequenceNumber string   `json:"sequenceNumber"`
	DedupID        string   `json:"dedupId"`
	ExistingMsg    *Message `json:"existingMsg"`
}

// validateFIFO validates FIFO queue requirements and handles deduplication.
func (qd *QueueData) validateFIFO(body, messageGroupID, messageDeduplicationID string, now time.Time) (*FifoResult, error) {
	if messageGroupID == "" {
		return nil, &QueueError{
			Code:    "MissingParameter",
			Message: "The request must contain the parameter MessageGroupId",
		}
	}

	dedupID := messageDeduplicationID
	if dedupID == "" {
		if qd.Queue.ContentBasedDeduplication {
			hash := sha256.Sum256([]byte(body))
			dedupID = hex.EncodeToString(hash[:])
		} else {
			return nil, &QueueError{
				Code:    "InvalidParameterValue",
				Message: "The queue should either have ContentBasedDeduplication enabled or MessageDeduplicationId provided explicitly",
			}
		}
	}

	// Clean up expired deduplication entries.
	for id, entry := range qd.DeduplicationCache {
		if now.After(entry.ExpiresAt) {
			delete(qd.DeduplicationCache, id)
		}
	}

	// Check deduplication cache (5-minute window).
	if entry, exists := qd.DeduplicationCache[dedupID]; exists {
		for _, msg := range qd.Messages {
			if msg.MessageID == entry.MessageID {
				return &FifoResult{ExistingMsg: msg}, nil
			}
		}
	}

	// Generate sequence number.
	qd.SequenceCounter++

	// Add to deduplication cache (5-minute TTL).
	qd.DeduplicationCache[dedupID] = DeduplicationEntry{
		MessageID: "",
		ExpiresAt: now.Add(5 * time.Minute),
	}

	return &FifoResult{
		SequenceNumber: fmt.Sprintf("%d", qd.SequenceCounter),
		DedupID:        dedupID,
	}, nil
}

// updateFIFOCache updates the deduplication cache with the message ID.
func (qd *QueueData) updateFIFOCache(dedupID, messageID string) {
	entry := qd.DeduplicationCache[dedupID]
	entry.MessageID = messageID
	qd.DeduplicationCache[dedupID] = entry
}

// SendMessage sends a message to a queue.
func (s *MemoryStorage) SendMessage(_ context.Context, queueURL, body string, delaySeconds int, messageAttributes map[string]MessageAttributeValue, messageGroupID, messageDeduplicationID string) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	now := time.Now()
	delay := delaySeconds

	if delay == 0 {
		delay = qd.Queue.DelaySeconds
	}

	var sequenceNumber, dedupID string

	if qd.Queue.FifoQueue {
		result, err := qd.validateFIFO(body, messageGroupID, messageDeduplicationID, now)
		if err != nil {
			return nil, err
		}

		if result.ExistingMsg != nil {
			return result.ExistingMsg, nil
		}

		sequenceNumber = result.SequenceNumber
		dedupID = result.DedupID
	}

	// MD5 is required by SQS specification for message body hash.
	md5Hash := md5.Sum([]byte(body)) //nolint:gosec // MD5 is required by SQS spec
	msg := &Message{
		MessageID:              uuid.New().String(),
		Body:                   body,
		MD5OfBody:              hex.EncodeToString(md5Hash[:]),
		MessageAttributes:      messageAttributes,
		SentTimestamp:          now,
		VisibleAt:              now.Add(time.Duration(delay) * time.Second),
		MessageGroupID:         messageGroupID,
		MessageDeduplicationID: messageDeduplicationID,
		SequenceNumber:         sequenceNumber,
		Attributes: map[string]string{
			"SentTimestamp":                    fmt.Sprintf("%d", now.UnixMilli()),
			"ApproximateReceiveCount":          "0",
			"ApproximateFirstReceiveTimestamp": "",
		},
	}

	if qd.Queue.FifoQueue {
		qd.updateFIFOCache(dedupID, msg.MessageID)
	}

	qd.Messages = append(qd.Messages, msg)

	return msg, nil
}

// ReceiveMessage receives messages from a queue.
func (s *MemoryStorage) ReceiveMessage(_ context.Context, queueURL string, maxMessages, visibilityTimeout, _ int) ([]*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	if visibilityTimeout == 0 {
		visibilityTimeout = qd.Queue.VisibilityTimeout
	}

	now := time.Now()
	result := make([]*Message, 0, maxMessages)
	remaining := make([]*Message, 0, len(qd.Messages))

	for _, msg := range qd.Messages {
		if len(result) >= maxMessages {
			remaining = append(remaining, msg)

			continue
		}

		if msg.VisibleAt.After(now) {
			remaining = append(remaining, msg)

			continue
		}

		// Make message invisible and add to inflight.
		msg.ReceiptHandle = uuid.New().String()
		msg.VisibleAt = now.Add(time.Duration(visibilityTimeout) * time.Second)
		msg.ReceiveCount++
		msg.Attributes["ApproximateReceiveCount"] = fmt.Sprintf("%d", msg.ReceiveCount)

		if msg.Attributes["ApproximateFirstReceiveTimestamp"] == "" {
			msg.Attributes["ApproximateFirstReceiveTimestamp"] = fmt.Sprintf("%d", now.UnixMilli())
		}

		qd.Inflight[msg.ReceiptHandle] = msg
		result = append(result, msg)
	}

	qd.Messages = remaining

	return result, nil
}

// DeleteMessage deletes a message from a queue.
func (s *MemoryStorage) DeleteMessage(_ context.Context, queueURL, receiptHandle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return ErrQueueDoesNotExist
	}

	if _, exists := qd.Inflight[receiptHandle]; !exists {
		return ErrReceiptHandleInvalid
	}

	delete(qd.Inflight, receiptHandle)

	return nil
}

// PurgeQueue purges all messages from a queue.
func (s *MemoryStorage) PurgeQueue(_ context.Context, queueURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return ErrQueueDoesNotExist
	}

	qd.Messages = make([]*Message, 0)
	qd.Inflight = make(map[string]*Message)

	return nil
}

// GetQueueAttributes gets queue attributes.
func (s *MemoryStorage) GetQueueAttributes(_ context.Context, queueURL string, attributeNames []string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	q := qd.Queue
	allAttrs := map[string]string{
		"QueueArn":                              q.ARN,
		"CreatedTimestamp":                      fmt.Sprintf("%d", q.CreatedTimestamp.Unix()),
		"LastModifiedTimestamp":                 fmt.Sprintf("%d", q.LastModifiedTimestamp.Unix()),
		"VisibilityTimeout":                     fmt.Sprintf("%d", q.VisibilityTimeout),
		"MessageRetentionPeriod":                fmt.Sprintf("%d", q.MessageRetentionPeriod),
		"DelaySeconds":                          fmt.Sprintf("%d", q.DelaySeconds),
		"MaximumMessageSize":                    fmt.Sprintf("%d", q.MaxMessageSize),
		"ReceiveMessageWaitTimeSeconds":         fmt.Sprintf("%d", q.ReceiveWaitTimeSeconds),
		"ApproximateNumberOfMessages":           fmt.Sprintf("%d", len(qd.Messages)),
		"ApproximateNumberOfMessagesNotVisible": fmt.Sprintf("%d", len(qd.Inflight)),
		"FifoQueue":                             fmt.Sprintf("%t", q.FifoQueue),
		"ContentBasedDeduplication":             fmt.Sprintf("%t", q.ContentBasedDeduplication),
	}

	// Check if "All" is requested.
	if slices.Contains(attributeNames, "All") {
		return allAttrs, nil
	}

	result := make(map[string]string)

	for _, name := range attributeNames {
		if val, ok := allAttrs[name]; ok {
			result[name] = val
		}
	}

	return result, nil
}

// SetQueueAttributes sets queue attributes.
func (s *MemoryStorage) SetQueueAttributes(_ context.Context, queueURL string, attributes map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.Queues[queueURL]
	if !exists {
		return ErrQueueDoesNotExist
	}

	applyQueueAttributes(qd.Queue, attributes)
	qd.Queue.LastModifiedTimestamp = time.Now()

	return nil
}

func applyQueueAttributes(q *Queue, attrs map[string]string) {
	for key, val := range attrs {
		switch key {
		case "VisibilityTimeout":
			_, _ = fmt.Sscanf(val, "%d", &q.VisibilityTimeout)
		case "MessageRetentionPeriod":
			_, _ = fmt.Sscanf(val, "%d", &q.MessageRetentionPeriod)
		case "DelaySeconds":
			_, _ = fmt.Sscanf(val, "%d", &q.DelaySeconds)
		case "MaximumMessageSize":
			_, _ = fmt.Sscanf(val, "%d", &q.MaxMessageSize)
		case "ReceiveMessageWaitTimeSeconds":
			_, _ = fmt.Sscanf(val, "%d", &q.ReceiveWaitTimeSeconds)
		case "ContentBasedDeduplication":
			q.ContentBasedDeduplication = val == "true"
		}
	}
}
