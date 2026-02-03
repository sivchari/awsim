package s3

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 is required for S3 ETag calculation per AWS specification
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// Storage defines the S3 storage interface.
type Storage interface {
	// Bucket operations
	CreateBucket(ctx context.Context, name string) error
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]Bucket, error)
	BucketExists(ctx context.Context, name string) (bool, error)

	// Object operations
	PutObject(ctx context.Context, bucket, key string, body io.Reader, metadata map[string]string) (*Object, error)
	GetObject(ctx context.Context, bucket, key string) (*Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	HeadObject(ctx context.Context, bucket, key string) (*Object, error)
	ListObjects(ctx context.Context, bucket, prefix, delimiter string, maxKeys int) ([]Object, []string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	buckets map[string]*memoryBucket
}

type memoryBucket struct {
	name         string
	creationDate time.Time
	objects      map[string]*Object
}

// NewMemoryStorage creates a new in-memory S3 storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		buckets: make(map[string]*memoryBucket),
	}
}

// CreateBucket creates a new bucket.
func (s *MemoryStorage) CreateBucket(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.buckets[name]; exists {
		return &BucketError{Code: "BucketAlreadyOwnedByYou", Message: "Your previous request to create the named bucket succeeded and you already own it.", BucketName: name}
	}

	s.buckets[name] = &memoryBucket{
		name:         name,
		creationDate: time.Now(),
		objects:      make(map[string]*Object),
	}

	return nil
}

// DeleteBucket deletes a bucket.
func (s *MemoryStorage) DeleteBucket(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[name]
	if !exists {
		return &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: name}
	}

	if len(bucket.objects) > 0 {
		return &BucketError{Code: "BucketNotEmpty", Message: "The bucket you tried to delete is not empty", BucketName: name}
	}

	delete(s.buckets, name)

	return nil
}

// ListBuckets returns all buckets.
func (s *MemoryStorage) ListBuckets(_ context.Context) ([]Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buckets := make([]Bucket, 0, len(s.buckets))
	for _, b := range s.buckets {
		buckets = append(buckets, Bucket{
			Name:         b.name,
			CreationDate: b.creationDate,
		})
	}

	// Sort by name for consistent ordering
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Name < buckets[j].Name
	})

	return buckets, nil
}

// BucketExists checks if a bucket exists.
func (s *MemoryStorage) BucketExists(_ context.Context, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.buckets[name]

	return exists, nil
}

// PutObject stores an object.
func (s *MemoryStorage) PutObject(_ context.Context, bucket, key string, body io.Reader, metadata map[string]string) (*Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
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

	if obj.ContentType == "" {
		obj.ContentType = "application/octet-stream"
	}

	b.objects[key] = obj

	return obj, nil
}

// GetObject retrieves an object.
func (s *MemoryStorage) GetObject(_ context.Context, bucket, key string) (*Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	obj, exists := b.objects[key]
	if !exists {
		return nil, &ObjectError{Code: "NoSuchKey", Message: "The specified key does not exist.", Key: key}
	}

	return obj, nil
}

// DeleteObject deletes an object.
func (s *MemoryStorage) DeleteObject(_ context.Context, bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	// S3 doesn't return error if key doesn't exist
	delete(b.objects, key)

	return nil
}

// HeadObject retrieves object metadata without body.
func (s *MemoryStorage) HeadObject(_ context.Context, bucket, key string) (*Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	obj, exists := b.objects[key]
	if !exists {
		return nil, &ObjectError{Code: "NoSuchKey", Message: "The specified key does not exist.", Key: key}
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

// ListObjects lists objects in a bucket.
func (s *MemoryStorage) ListObjects(_ context.Context, bucket, prefix, delimiter string, maxKeys int) ([]Object, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	if maxKeys <= 0 {
		maxKeys = 1000
	}

	objects := make([]Object, 0)
	commonPrefixes := make(map[string]bool)

	// Collect all matching keys.
	keys := make([]string, 0, len(b.objects))

	for key := range b.objects {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	// Sort keys for consistent ordering
	sort.Strings(keys)

	for _, key := range keys {
		obj := b.objects[key]

		// Handle delimiter
		if delimiter != "" {
			// Find the part after prefix
			remainder := strings.TrimPrefix(key, prefix)
			if idx := strings.Index(remainder, delimiter); idx >= 0 {
				// This is a common prefix
				commonPrefix := prefix + remainder[:idx+len(delimiter)]
				commonPrefixes[commonPrefix] = true

				continue
			}
		}

		objects = append(objects, Object{
			Key:          obj.Key,
			ETag:         obj.ETag,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		})

		if len(objects) >= maxKeys {
			break
		}
	}

	// Convert common prefixes to sorted slice
	prefixList := make([]string, 0, len(commonPrefixes))
	for p := range commonPrefixes {
		prefixList = append(prefixList, p)
	}

	sort.Strings(prefixList)

	return objects, prefixList, nil
}

// BucketError represents an S3 bucket error.
type BucketError struct {
	Code       string
	Message    string
	BucketName string
}

func (e *BucketError) Error() string {
	return fmt.Sprintf("%s: %s (bucket: %s)", e.Code, e.Message, e.BucketName)
}

// ObjectError represents an S3 object error.
type ObjectError struct {
	Code    string
	Message string
	Key     string
}

func (e *ObjectError) Error() string {
	return fmt.Sprintf("%s: %s (key: %s)", e.Code, e.Message, e.Key)
}
