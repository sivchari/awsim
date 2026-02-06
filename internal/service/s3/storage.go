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

// Versioning status constants.
const (
	VersioningEnabled   = "Enabled"
	VersioningSuspended = "Suspended"
	VersionIDNull       = "null"
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
	GetObjectVersion(ctx context.Context, bucket, key, versionID string) (*Object, error)
	DeleteObject(ctx context.Context, bucket, key string) (*Object, error)
	DeleteObjectVersion(ctx context.Context, bucket, key, versionID string) (*Object, error)
	HeadObject(ctx context.Context, bucket, key string) (*Object, error)
	ListObjects(ctx context.Context, bucket, prefix, delimiter string, maxKeys int) ([]Object, []string, error)

	// Versioning operations
	PutBucketVersioning(ctx context.Context, bucket, status string) error
	GetBucketVersioning(ctx context.Context, bucket string) (string, error)
	ListObjectVersions(ctx context.Context, bucket, prefix, delimiter string, maxKeys int) ([]Object, []string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	buckets map[string]*memoryBucket
}

type memoryBucket struct {
	name             string
	creationDate     time.Time
	objects          map[string]*Object   // current/latest version per key
	versions         map[string][]*Object // all versions per key (newest first)
	versioningStatus string               // "", "Enabled", "Suspended"
	versionIDCounter uint64               // counter for generating version IDs
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
		name:             name,
		creationDate:     time.Now(),
		objects:          make(map[string]*Object),
		versions:         make(map[string][]*Object),
		versioningStatus: "",
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

	// Handle versioning
	switch b.versioningStatus {
	case VersioningEnabled:
		// Generate version ID
		b.versionIDCounter++
		obj.VersionID = fmt.Sprintf("v%d", b.versionIDCounter)

		// Prepend to versions list (newest first)
		b.versions[key] = append([]*Object{obj}, b.versions[key]...)
	case VersioningSuspended:
		// For suspended versioning, use "null" version ID
		obj.VersionID = VersionIDNull

		// Remove any existing "null" version
		versions := b.versions[key]
		newVersions := make([]*Object, 0, len(versions))

		for _, v := range versions {
			if v.VersionID != VersionIDNull {
				newVersions = append(newVersions, v)
			}
		}

		b.versions[key] = append([]*Object{obj}, newVersions...)
	}

	// Always update current object
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

	// Check if current version is a delete marker
	if obj.IsDeleteMarker {
		return nil, &ObjectError{Code: "NoSuchKey", Message: "The specified key does not exist.", Key: key}
	}

	return obj, nil
}

// GetObjectVersion retrieves a specific version of an object.
func (s *MemoryStorage) GetObjectVersion(_ context.Context, bucket, key, versionID string) (*Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	versions := b.versions[key]
	for _, obj := range versions {
		if obj.VersionID == versionID {
			if obj.IsDeleteMarker {
				return nil, &ObjectError{Code: "MethodNotAllowed", Message: "The specified method is not allowed against this resource.", Key: key}
			}

			return obj, nil
		}
	}

	return nil, &ObjectError{Code: "NoSuchVersion", Message: "The specified version does not exist.", Key: key}
}

// DeleteObject deletes an object.
// Returns the deleted object (or delete marker for versioned buckets), or nil if non-versioned delete.
func (s *MemoryStorage) DeleteObject(_ context.Context, bucket, key string) (*Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	// Handle versioning - create delete marker for enabled buckets
	if b.versioningStatus == VersioningEnabled {
		b.versionIDCounter++
		deleteMarker := &Object{
			Key:            key,
			VersionID:      fmt.Sprintf("v%d", b.versionIDCounter),
			IsDeleteMarker: true,
			LastModified:   time.Now(),
		}

		// Prepend delete marker to versions
		b.versions[key] = append([]*Object{deleteMarker}, b.versions[key]...)
		b.objects[key] = deleteMarker

		return deleteMarker, nil
	}

	// For non-versioned or suspended buckets, just delete
	delete(b.objects, key)

	// For suspended buckets, also remove "null" version
	if b.versioningStatus == VersioningSuspended {
		versions := b.versions[key]
		newVersions := make([]*Object, 0, len(versions))

		for _, v := range versions {
			if v.VersionID != VersionIDNull {
				newVersions = append(newVersions, v)
			}
		}

		if len(newVersions) == 0 {
			delete(b.versions, key)
		} else {
			b.versions[key] = newVersions
		}
	}

	// Return empty object for non-versioned delete (S3 returns 204 with no body)
	return &Object{Key: key}, nil //nolint:nilnil // S3 returns empty response for non-versioned delete
}

// DeleteObjectVersion deletes a specific version of an object.
func (s *MemoryStorage) DeleteObjectVersion(_ context.Context, bucket, key, versionID string) (*Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	versions := b.versions[key]
	deletedObj, newVersions := filterOutVersion(versions, versionID)

	// S3 doesn't return error if version doesn't exist, returns empty object
	if deletedObj == nil {
		return &Object{Key: key, VersionID: versionID}, nil //nolint:nilnil // S3 returns empty response for non-existent version
	}

	if len(newVersions) == 0 {
		delete(b.versions, key)
		delete(b.objects, key)
	} else {
		b.versions[key] = newVersions
		// Update current object to the newest version
		b.objects[key] = newVersions[0]
	}

	return deletedObj, nil
}

// filterOutVersion removes a specific version from the versions list.
func filterOutVersion(versions []*Object, versionID string) (*Object, []*Object) {
	var deletedObj *Object

	newVersions := make([]*Object, 0, len(versions))

	for _, v := range versions {
		if v.VersionID == versionID {
			deletedObj = v
		} else {
			newVersions = append(newVersions, v)
		}
	}

	return deletedObj, newVersions
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

// PutBucketVersioning sets the versioning status of a bucket.
func (s *MemoryStorage) PutBucketVersioning(_ context.Context, bucket, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	if status != VersioningEnabled && status != VersioningSuspended && status != "" {
		return &BucketError{Code: "MalformedXML", Message: "Invalid versioning status", BucketName: bucket}
	}

	b.versioningStatus = status

	return nil
}

// GetBucketVersioning returns the versioning status of a bucket.
func (s *MemoryStorage) GetBucketVersioning(_ context.Context, bucket string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return "", &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	return b.versioningStatus, nil
}

// ListObjectVersions lists all versions of objects in a bucket.
func (s *MemoryStorage) ListObjectVersions(_ context.Context, bucket, prefix, delimiter string, maxKeys int) ([]Object, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, nil, &BucketError{Code: "NoSuchBucket", Message: "The specified bucket does not exist", BucketName: bucket}
	}

	if maxKeys <= 0 {
		maxKeys = 1000
	}

	keys := collectVersionKeys(b, prefix)
	sort.Strings(keys)

	objects, commonPrefixes := processVersionKeys(b, keys, prefix, delimiter, maxKeys)
	prefixList := sortedPrefixList(commonPrefixes)

	return objects, prefixList, nil
}

// collectVersionKeys collects all keys that match the prefix from both versions and objects maps.
func collectVersionKeys(b *memoryBucket, prefix string) []string {
	keySet := make(map[string]bool)

	for key := range b.versions {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keySet[key] = true
		}
	}

	for key := range b.objects {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keySet[key] = true
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}

	return keys
}

// processVersionKeys processes keys and returns objects and common prefixes.
func processVersionKeys(b *memoryBucket, keys []string, prefix, delimiter string, maxKeys int) ([]Object, map[string]bool) {
	objects := make([]Object, 0)
	commonPrefixes := make(map[string]bool)
	count := 0

	for _, key := range keys {
		if count >= maxKeys {
			break
		}

		// Handle delimiter for common prefixes
		if delimiter != "" {
			if cp := extractCommonPrefix(key, prefix, delimiter); cp != "" {
				commonPrefixes[cp] = true

				continue
			}
		}

		// Add versions for this key
		added := addKeyVersions(b, key, &objects, maxKeys-count)
		count += added
	}

	return objects, commonPrefixes
}

// extractCommonPrefix extracts common prefix if delimiter is found.
func extractCommonPrefix(key, prefix, delimiter string) string {
	remainder := strings.TrimPrefix(key, prefix)
	if idx := strings.Index(remainder, delimiter); idx >= 0 {
		return prefix + remainder[:idx+len(delimiter)]
	}

	return ""
}

// addKeyVersions adds all versions of a key to the objects slice.
func addKeyVersions(b *memoryBucket, key string, objects *[]Object, limit int) int {
	versions := b.versions[key]
	if len(versions) == 0 {
		// No versioning history, include current object if exists
		if obj, exists := b.objects[key]; exists {
			*objects = append(*objects, objectToVersionInfo(obj))

			return 1
		}

		return 0
	}

	count := 0

	for _, obj := range versions {
		if count >= limit {
			break
		}

		*objects = append(*objects, objectToVersionInfo(obj))
		count++
	}

	return count
}

// objectToVersionInfo converts an Object to version info format.
func objectToVersionInfo(obj *Object) Object {
	return Object{
		Key:            obj.Key,
		VersionID:      obj.VersionID,
		ETag:           obj.ETag,
		Size:           obj.Size,
		LastModified:   obj.LastModified,
		IsDeleteMarker: obj.IsDeleteMarker,
	}
}

// sortedPrefixList converts a map of prefixes to a sorted slice.
func sortedPrefixList(prefixes map[string]bool) []string {
	list := make([]string, 0, len(prefixes))
	for p := range prefixes {
		list = append(list, p)
	}

	sort.Strings(list)

	return list
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
