package storage

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 is required for S3 ETag calculation per AWS specification
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryS3Storage is an in-memory implementation of S3Storage.
type MemoryS3Storage struct {
	mu      sync.RWMutex
	buckets map[string]*memoryBucket
}

type memoryBucket struct {
	name         string
	creationDate time.Time
	objects      map[string]*Object
}

// NewMemoryS3Storage creates a new in-memory S3 storage.
func NewMemoryS3Storage() *MemoryS3Storage {
	return &MemoryS3Storage{
		buckets: make(map[string]*memoryBucket),
	}
}

// CreateBucket creates a new bucket.
func (s *MemoryS3Storage) CreateBucket(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.buckets[name]; exists {
		return fmt.Errorf("bucket already exists: %s", name)
	}

	s.buckets[name] = &memoryBucket{
		name:         name,
		creationDate: time.Now(),
		objects:      make(map[string]*Object),
	}
	return nil
}

// DeleteBucket deletes a bucket.
func (s *MemoryS3Storage) DeleteBucket(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[name]
	if !exists {
		return fmt.Errorf("bucket not found: %s", name)
	}

	if len(bucket.objects) > 0 {
		return fmt.Errorf("bucket not empty: %s", name)
	}

	delete(s.buckets, name)
	return nil
}

// ListBuckets returns all buckets.
func (s *MemoryS3Storage) ListBuckets(_ context.Context) ([]Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buckets := make([]Bucket, 0, len(s.buckets))
	for _, b := range s.buckets {
		buckets = append(buckets, Bucket{
			Name:         b.name,
			CreationDate: b.creationDate,
		})
	}
	return buckets, nil
}

// BucketExists checks if a bucket exists.
func (s *MemoryS3Storage) BucketExists(_ context.Context, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.buckets[name]
	return exists, nil
}

// PutObject stores an object.
func (s *MemoryS3Storage) PutObject(_ context.Context, bucket, key string, body io.Reader, metadata map[string]string) (*Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, fmt.Errorf("bucket not found: %s", bucket)
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	hash := md5.Sum(data) //nolint:gosec // MD5 is required for S3 ETag calculation per AWS specification
	etag := hex.EncodeToString(hash[:])

	obj := &Object{
		Key:          key,
		Body:         data,
		ETag:         fmt.Sprintf("%q", etag),
		Size:         int64(len(data)),
		LastModified: time.Now(),
		Metadata:     metadata,
	}

	if metadata != nil {
		if ct, ok := metadata["Content-Type"]; ok {
			obj.ContentType = ct
		}
	}

	b.objects[key] = obj
	return obj, nil
}

// GetObject retrieves an object.
func (s *MemoryS3Storage) GetObject(_ context.Context, bucket, key string) (*Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, fmt.Errorf("bucket not found: %s", bucket)
	}

	obj, exists := b.objects[key]
	if !exists {
		return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
	}

	return obj, nil
}

// DeleteObject deletes an object.
func (s *MemoryS3Storage) DeleteObject(_ context.Context, bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return fmt.Errorf("bucket not found: %s", bucket)
	}

	delete(b.objects, key)
	return nil
}

// ListObjects lists objects in a bucket.
func (s *MemoryS3Storage) ListObjects(_ context.Context, bucket, prefix string, maxKeys int) ([]Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, fmt.Errorf("bucket not found: %s", bucket)
	}

	objects := make([]Object, 0)
	for key, obj := range b.objects {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			objects = append(objects, *obj)
			if maxKeys > 0 && len(objects) >= maxKeys {
				break
			}
		}
	}

	return objects, nil
}

// HeadObject retrieves object metadata without body.
func (s *MemoryS3Storage) HeadObject(_ context.Context, bucket, key string) (*Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, fmt.Errorf("bucket not found: %s", bucket)
	}

	obj, exists := b.objects[key]
	if !exists {
		return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
	}

	// Return metadata only (no body)
	return &Object{
		Key:          obj.Key,
		ContentType:  obj.ContentType,
		ETag:         obj.ETag,
		Size:         obj.Size,
		LastModified: obj.LastModified,
		Metadata:     obj.Metadata,
	}, nil
}

// MemorySQSStorage is an in-memory implementation of SQSStorage.
type MemorySQSStorage struct {
	mu     sync.RWMutex
	queues map[string]*memoryQueue
}

type memoryQueue struct {
	queue    Queue
	messages []*Message
}

// NewMemorySQSStorage creates a new in-memory SQS storage.
func NewMemorySQSStorage() *MemorySQSStorage {
	return &MemorySQSStorage{
		queues: make(map[string]*memoryQueue),
	}
}

// CreateQueue creates a new queue.
func (s *MemorySQSStorage) CreateQueue(_ context.Context, name string, _ map[string]string) (*Queue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queueURL := fmt.Sprintf("http://localhost:4566/000000000000/%s", name)

	if _, exists := s.queues[queueURL]; exists {
		return nil, fmt.Errorf("queue already exists: %s", name)
	}

	q := &Queue{
		Name: name,
		URL:  queueURL,
		Arn:  fmt.Sprintf("arn:aws:sqs:us-east-1:000000000000:%s", name),
	}

	s.queues[queueURL] = &memoryQueue{
		queue:    *q,
		messages: make([]*Message, 0),
	}

	return q, nil
}

// DeleteQueue deletes a queue.
func (s *MemorySQSStorage) DeleteQueue(_ context.Context, queueURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.queues[queueURL]; !exists {
		return fmt.Errorf("queue not found: %s", queueURL)
	}

	delete(s.queues, queueURL)
	return nil
}

// ListQueues returns all queues.
func (s *MemorySQSStorage) ListQueues(_ context.Context, prefix string) ([]Queue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	queues := make([]Queue, 0, len(s.queues))
	for _, q := range s.queues {
		if prefix == "" || strings.HasPrefix(q.queue.Name, prefix) {
			queues = append(queues, q.queue)
		}
	}
	return queues, nil
}

// GetQueueURL returns the URL for a queue name.
func (s *MemorySQSStorage) GetQueueURL(_ context.Context, name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, q := range s.queues {
		if q.queue.Name == name {
			return q.queue.URL, nil
		}
	}
	return "", fmt.Errorf("queue not found: %s", name)
}

// SendMessage sends a message to a queue.
func (s *MemorySQSStorage) SendMessage(_ context.Context, queueURL, body string, attributes map[string]string) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q, exists := s.queues[queueURL]
	if !exists {
		return nil, fmt.Errorf("queue not found: %s", queueURL)
	}

	msg := &Message{
		MessageID:     uuid.New().String(),
		ReceiptHandle: uuid.New().String(),
		Body:          body,
		Attributes:    attributes,
		SentTimestamp: time.Now(),
	}

	q.messages = append(q.messages, msg)
	return msg, nil
}

// ReceiveMessages receives messages from a queue.
func (s *MemorySQSStorage) ReceiveMessages(_ context.Context, queueURL string, maxMessages int, _ time.Duration) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q, exists := s.queues[queueURL]
	if !exists {
		return nil, fmt.Errorf("queue not found: %s", queueURL)
	}

	if maxMessages <= 0 || maxMessages > 10 {
		maxMessages = 10
	}

	count := min(maxMessages, len(q.messages))
	messages := make([]Message, count)

	for i := range count {
		messages[i] = *q.messages[i]
	}

	return messages, nil
}

// DeleteMessage deletes a message from a queue.
func (s *MemorySQSStorage) DeleteMessage(_ context.Context, queueURL, receiptHandle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	q, exists := s.queues[queueURL]
	if !exists {
		return fmt.Errorf("queue not found: %s", queueURL)
	}

	for i, msg := range q.messages {
		if msg.ReceiptHandle == receiptHandle {
			q.messages = append(q.messages[:i], q.messages[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("message not found: %s", receiptHandle)
}
