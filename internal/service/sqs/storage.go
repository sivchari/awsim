package sqs

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 is required by SQS spec for message body hash
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// deduplicationEntry holds deduplication information for FIFO queues.
type deduplicationEntry struct {
	messageID string
	expiresAt time.Time
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

// MemoryStorage implements Storage using in-memory maps.
type MemoryStorage struct {
	mu      sync.RWMutex
	queues  map[string]*queueData
	baseURL string
}

type queueData struct {
	queue              *Queue
	messages           []*Message
	inflight           map[string]*Message           // receiptHandle -> message
	deduplicationCache map[string]deduplicationEntry // deduplicationID -> entry (FIFO only)
	sequenceCounter    uint64                        // Per-queue sequence number (FIFO only)
}

// NewMemoryStorage creates a new in-memory SQS storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		queues:  make(map[string]*queueData),
		baseURL: baseURL,
	}
}

// CreateQueue creates a new queue.
func (s *MemoryStorage) CreateQueue(_ context.Context, name string, attributes map[string]string) (*Queue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queueURL := fmt.Sprintf("%s/000000000000/%s", s.baseURL, name)

	if qd, exists := s.queues[queueURL]; exists {
		return qd.queue, nil
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

	qd := &queueData{
		queue:    queue,
		messages: make([]*Message, 0),
		inflight: make(map[string]*Message),
	}

	if isFifo {
		qd.deduplicationCache = make(map[string]deduplicationEntry)
	}

	s.queues[queueURL] = qd

	return queue, nil
}

// DeleteQueue deletes a queue.
func (s *MemoryStorage) DeleteQueue(_ context.Context, queueURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.queues[queueURL]; !exists {
		return ErrQueueDoesNotExist
	}

	delete(s.queues, queueURL)

	return nil
}

// ListQueues lists all queues.
func (s *MemoryStorage) ListQueues(_ context.Context, prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urls := make([]string, 0, len(s.queues))

	for url, qd := range s.queues {
		if prefix == "" || len(qd.queue.Name) >= len(prefix) && qd.queue.Name[:len(prefix)] == prefix {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

// GetQueueURL gets the URL for a queue by name.
func (s *MemoryStorage) GetQueueURL(_ context.Context, name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for url, qd := range s.queues {
		if qd.queue.Name == name {
			return url, nil
		}
	}

	return "", ErrQueueDoesNotExist
}

// GetQueue gets a queue by URL.
func (s *MemoryStorage) GetQueue(_ context.Context, queueURL string) (*Queue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	qd, exists := s.queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	return qd.queue, nil
}

// fifoResult holds the result of FIFO validation and deduplication.
type fifoResult struct {
	sequenceNumber string
	dedupID        string
	existingMsg    *Message
}

// validateFIFO validates FIFO queue requirements and handles deduplication.
func (qd *queueData) validateFIFO(body, messageGroupID, messageDeduplicationID string, now time.Time) (*fifoResult, error) {
	if messageGroupID == "" {
		return nil, &QueueError{
			Code:    "MissingParameter",
			Message: "The request must contain the parameter MessageGroupId",
		}
	}

	dedupID := messageDeduplicationID
	if dedupID == "" {
		if qd.queue.ContentBasedDeduplication {
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
	for id, entry := range qd.deduplicationCache {
		if now.After(entry.expiresAt) {
			delete(qd.deduplicationCache, id)
		}
	}

	// Check deduplication cache (5-minute window).
	if entry, exists := qd.deduplicationCache[dedupID]; exists {
		for _, msg := range qd.messages {
			if msg.MessageID == entry.messageID {
				return &fifoResult{existingMsg: msg}, nil
			}
		}
	}

	// Generate sequence number.
	qd.sequenceCounter++

	// Add to deduplication cache (5-minute TTL).
	qd.deduplicationCache[dedupID] = deduplicationEntry{
		messageID: "",
		expiresAt: now.Add(5 * time.Minute),
	}

	return &fifoResult{
		sequenceNumber: fmt.Sprintf("%d", qd.sequenceCounter),
		dedupID:        dedupID,
	}, nil
}

// updateFIFOCache updates the deduplication cache with the message ID.
func (qd *queueData) updateFIFOCache(dedupID, messageID string) {
	entry := qd.deduplicationCache[dedupID]
	entry.messageID = messageID
	qd.deduplicationCache[dedupID] = entry
}

// SendMessage sends a message to a queue.
func (s *MemoryStorage) SendMessage(_ context.Context, queueURL, body string, delaySeconds int, messageAttributes map[string]MessageAttributeValue, messageGroupID, messageDeduplicationID string) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	now := time.Now()
	delay := delaySeconds

	if delay == 0 {
		delay = qd.queue.DelaySeconds
	}

	var sequenceNumber, dedupID string

	if qd.queue.FifoQueue {
		result, err := qd.validateFIFO(body, messageGroupID, messageDeduplicationID, now)
		if err != nil {
			return nil, err
		}

		if result.existingMsg != nil {
			return result.existingMsg, nil
		}

		sequenceNumber = result.sequenceNumber
		dedupID = result.dedupID
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

	if qd.queue.FifoQueue {
		qd.updateFIFOCache(dedupID, msg.MessageID)
	}

	qd.messages = append(qd.messages, msg)

	return msg, nil
}

// ReceiveMessage receives messages from a queue.
func (s *MemoryStorage) ReceiveMessage(_ context.Context, queueURL string, maxMessages, visibilityTimeout, _ int) ([]*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	if visibilityTimeout == 0 {
		visibilityTimeout = qd.queue.VisibilityTimeout
	}

	now := time.Now()
	result := make([]*Message, 0, maxMessages)
	remaining := make([]*Message, 0, len(qd.messages))

	for _, msg := range qd.messages {
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

		qd.inflight[msg.ReceiptHandle] = msg
		result = append(result, msg)
	}

	qd.messages = remaining

	return result, nil
}

// DeleteMessage deletes a message from a queue.
func (s *MemoryStorage) DeleteMessage(_ context.Context, queueURL, receiptHandle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.queues[queueURL]
	if !exists {
		return ErrQueueDoesNotExist
	}

	if _, exists := qd.inflight[receiptHandle]; !exists {
		return ErrReceiptHandleInvalid
	}

	delete(qd.inflight, receiptHandle)

	return nil
}

// PurgeQueue purges all messages from a queue.
func (s *MemoryStorage) PurgeQueue(_ context.Context, queueURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	qd, exists := s.queues[queueURL]
	if !exists {
		return ErrQueueDoesNotExist
	}

	qd.messages = make([]*Message, 0)
	qd.inflight = make(map[string]*Message)

	return nil
}

// GetQueueAttributes gets queue attributes.
func (s *MemoryStorage) GetQueueAttributes(_ context.Context, queueURL string, attributeNames []string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	qd, exists := s.queues[queueURL]
	if !exists {
		return nil, ErrQueueDoesNotExist
	}

	q := qd.queue
	allAttrs := map[string]string{
		"QueueArn":                              q.ARN,
		"CreatedTimestamp":                      fmt.Sprintf("%d", q.CreatedTimestamp.Unix()),
		"LastModifiedTimestamp":                 fmt.Sprintf("%d", q.LastModifiedTimestamp.Unix()),
		"VisibilityTimeout":                     fmt.Sprintf("%d", q.VisibilityTimeout),
		"MessageRetentionPeriod":                fmt.Sprintf("%d", q.MessageRetentionPeriod),
		"DelaySeconds":                          fmt.Sprintf("%d", q.DelaySeconds),
		"MaximumMessageSize":                    fmt.Sprintf("%d", q.MaxMessageSize),
		"ReceiveMessageWaitTimeSeconds":         fmt.Sprintf("%d", q.ReceiveWaitTimeSeconds),
		"ApproximateNumberOfMessages":           fmt.Sprintf("%d", len(qd.messages)),
		"ApproximateNumberOfMessagesNotVisible": fmt.Sprintf("%d", len(qd.inflight)),
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

	qd, exists := s.queues[queueURL]
	if !exists {
		return ErrQueueDoesNotExist
	}

	applyQueueAttributes(qd.queue, attributes)
	qd.queue.LastModifiedTimestamp = time.Now()

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
