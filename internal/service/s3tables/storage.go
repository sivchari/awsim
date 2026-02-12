package s3tables

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAccountID = "000000000000"
	defaultRegion    = "us-east-1"
	defaultMaxItems  = 100
)

// Storage defines the S3 Tables storage interface.
type Storage interface {
	// TableBucket operations
	CreateTableBucket(ctx context.Context, name string) (*TableBucket, error)
	DeleteTableBucket(ctx context.Context, arn string) error
	GetTableBucket(ctx context.Context, arn string) (*TableBucket, error)
	ListTableBuckets(ctx context.Context, prefix string, maxBuckets int) ([]TableBucketSummary, error)

	// Namespace operations
	CreateNamespace(ctx context.Context, tableBucketArn string, namespace []string) (*Namespace, error)
	DeleteNamespace(ctx context.Context, tableBucketArn, namespace string) error
	GetNamespace(ctx context.Context, tableBucketArn, namespace string) (*Namespace, error)
	ListNamespaces(ctx context.Context, tableBucketArn, prefix string, maxNamespaces int) ([]NamespaceSummary, error)

	// Table operations
	CreateTable(ctx context.Context, tableBucketArn, namespace, name, format string) (*Table, error)
	DeleteTable(ctx context.Context, tableBucketArn, namespace, name string) error
	GetTable(ctx context.Context, tableBucketArn, namespace, name string) (*Table, error)
	ListTables(ctx context.Context, tableBucketArn, namespace, prefix string, maxTables int) ([]TableSummary, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu           sync.RWMutex
	tableBuckets map[string]*TableBucket          // ARN -> TableBucket
	namespaces   map[string]map[string]*Namespace // TableBucketARN -> namespace -> Namespace
	tables       map[string]map[string]*Table     // TableBucketARN/namespace -> tableName -> Table
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tableBuckets: make(map[string]*TableBucket),
		namespaces:   make(map[string]map[string]*Namespace),
		tables:       make(map[string]map[string]*Table),
	}
}

// CreateTableBucket creates a new table bucket.
func (s *MemoryStorage) CreateTableBucket(_ context.Context, name string) (*TableBucket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if bucket with same name already exists
	for _, bucket := range s.tableBuckets {
		if bucket.Name == name {
			return nil, &Error{
				Code:    errConflict,
				Message: fmt.Sprintf("Table bucket with name '%s' already exists", name),
			}
		}
	}

	arn := fmt.Sprintf("arn:aws:s3tables:%s:%s:bucket/%s", defaultRegion, defaultAccountID, name)

	bucket := &TableBucket{
		Arn:       arn,
		Name:      name,
		OwnerID:   defaultAccountID,
		CreatedAt: time.Now().UTC(),
	}

	s.tableBuckets[arn] = bucket
	s.namespaces[arn] = make(map[string]*Namespace)

	return bucket, nil
}

// DeleteTableBucket deletes a table bucket.
func (s *MemoryStorage) DeleteTableBucket(_ context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tableBuckets[arn]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", arn),
		}
	}

	// Check if bucket has namespaces
	if namespaces, exists := s.namespaces[arn]; exists && len(namespaces) > 0 {
		return &Error{
			Code:    errConflict,
			Message: "Table bucket contains namespaces and cannot be deleted",
		}
	}

	delete(s.tableBuckets, arn)
	delete(s.namespaces, arn)

	return nil
}

// GetTableBucket retrieves a table bucket.
func (s *MemoryStorage) GetTableBucket(_ context.Context, arn string) (*TableBucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, exists := s.tableBuckets[arn]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", arn),
		}
	}

	return bucket, nil
}

// ListTableBuckets lists all table buckets.
func (s *MemoryStorage) ListTableBuckets(_ context.Context, prefix string, maxBuckets int) ([]TableBucketSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxBuckets <= 0 {
		maxBuckets = defaultMaxItems
	}

	buckets := make([]TableBucketSummary, 0)

	for _, bucket := range s.tableBuckets {
		if prefix != "" && !strings.HasPrefix(bucket.Name, prefix) {
			continue
		}

		buckets = append(buckets, TableBucketSummary{
			Arn:       bucket.Arn,
			Name:      bucket.Name,
			OwnerID:   bucket.OwnerID,
			CreatedAt: bucket.CreatedAt,
		})

		if len(buckets) >= maxBuckets {
			break
		}
	}

	return buckets, nil
}

// CreateNamespace creates a new namespace.
func (s *MemoryStorage) CreateNamespace(_ context.Context, tableBucketArn string, namespace []string) (*Namespace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	namespaceKey := strings.Join(namespace, ".")

	if s.namespaces[tableBucketArn] == nil {
		s.namespaces[tableBucketArn] = make(map[string]*Namespace)
	}

	if _, exists := s.namespaces[tableBucketArn][namespaceKey]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Namespace '%s' already exists", namespaceKey),
		}
	}

	ns := &Namespace{
		Namespace:      namespace,
		TableBucketArn: tableBucketArn,
		OwnerID:        defaultAccountID,
		CreatedAt:      time.Now().UTC(),
		CreatedBy:      defaultAccountID,
	}

	s.namespaces[tableBucketArn][namespaceKey] = ns

	return ns, nil
}

// DeleteNamespace deletes a namespace.
func (s *MemoryStorage) DeleteNamespace(_ context.Context, tableBucketArn, namespace string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	if _, exists := s.namespaces[tableBucketArn][namespace]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Namespace '%s' not found", namespace),
		}
	}

	// Check if namespace has tables
	tableKey := tableBucketArn + "/" + namespace
	if tables, exists := s.tables[tableKey]; exists && len(tables) > 0 {
		return &Error{
			Code:    errConflict,
			Message: "Namespace contains tables and cannot be deleted",
		}
	}

	delete(s.namespaces[tableBucketArn], namespace)

	return nil
}

// GetNamespace retrieves a namespace.
func (s *MemoryStorage) GetNamespace(_ context.Context, tableBucketArn, namespace string) (*Namespace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	ns, exists := s.namespaces[tableBucketArn][namespace]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Namespace '%s' not found", namespace),
		}
	}

	return ns, nil
}

// ListNamespaces lists all namespaces in a table bucket.
func (s *MemoryStorage) ListNamespaces(_ context.Context, tableBucketArn, prefix string, maxNamespaces int) ([]NamespaceSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	if maxNamespaces <= 0 {
		maxNamespaces = defaultMaxItems
	}

	namespaces := make([]NamespaceSummary, 0)

	for _, ns := range s.namespaces[tableBucketArn] {
		nsName := strings.Join(ns.Namespace, ".")
		if prefix != "" && !strings.HasPrefix(nsName, prefix) {
			continue
		}

		namespaces = append(namespaces, NamespaceSummary{
			Namespace: ns.Namespace,
			CreatedAt: ns.CreatedAt,
			CreatedBy: ns.CreatedBy,
			OwnerID:   ns.OwnerID,
		})

		if len(namespaces) >= maxNamespaces {
			break
		}
	}

	return namespaces, nil
}

// CreateTable creates a new table.
func (s *MemoryStorage) CreateTable(_ context.Context, tableBucketArn, namespace, name, format string) (*Table, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	if _, exists := s.namespaces[tableBucketArn][namespace]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Namespace '%s' not found", namespace),
		}
	}

	tableKey := tableBucketArn + "/" + namespace
	if s.tables[tableKey] == nil {
		s.tables[tableKey] = make(map[string]*Table)
	}

	if _, exists := s.tables[tableKey][name]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Table '%s' already exists in namespace '%s'", name, namespace),
		}
	}

	// Extract bucket name from ARN for table ARN
	bucketName := extractBucketNameFromArn(tableBucketArn)
	tableArn := fmt.Sprintf("arn:aws:s3tables:%s:%s:bucket/%s/table/%s/%s",
		defaultRegion, defaultAccountID, bucketName, namespace, name)

	now := time.Now().UTC()
	versionToken := uuid.New().String()

	table := &Table{
		Arn:            tableArn,
		Name:           name,
		Namespace:      namespace,
		TableBucketArn: tableBucketArn,
		Type:           "customer",
		Format:         format,
		VersionToken:   versionToken,
		CreatedAt:      now,
		CreatedBy:      defaultAccountID,
		ModifiedAt:     now,
		ModifiedBy:     defaultAccountID,
		OwnerID:        defaultAccountID,
	}

	s.tables[tableKey][name] = table

	return table, nil
}

// DeleteTable deletes a table.
func (s *MemoryStorage) DeleteTable(_ context.Context, tableBucketArn, namespace, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	tableKey := tableBucketArn + "/" + namespace
	if _, exists := s.tables[tableKey][name]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table '%s' not found in namespace '%s'", name, namespace),
		}
	}

	delete(s.tables[tableKey], name)

	return nil
}

// GetTable retrieves a table.
func (s *MemoryStorage) GetTable(_ context.Context, tableBucketArn, namespace, name string) (*Table, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	tableKey := tableBucketArn + "/" + namespace
	table, exists := s.tables[tableKey][name]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table '%s' not found in namespace '%s'", name, namespace),
		}
	}

	return table, nil
}

// ListTables lists all tables in a namespace.
func (s *MemoryStorage) ListTables(_ context.Context, tableBucketArn, namespace, prefix string, maxTables int) ([]TableSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.tableBuckets[tableBucketArn]; !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Table bucket '%s' not found", tableBucketArn),
		}
	}

	if maxTables <= 0 {
		maxTables = defaultMaxItems
	}

	tables := make([]TableSummary, 0)

	// If namespace is specified, list tables in that namespace only
	if namespace != "" {
		tableKey := tableBucketArn + "/" + namespace
		for _, table := range s.tables[tableKey] {
			if prefix != "" && !strings.HasPrefix(table.Name, prefix) {
				continue
			}

			tables = append(tables, TableSummary{
				Arn:        table.Arn,
				Name:       table.Name,
				Namespace:  []string{table.Namespace},
				Type:       table.Type,
				CreatedAt:  table.CreatedAt,
				ModifiedAt: table.ModifiedAt,
			})

			if len(tables) >= maxTables {
				break
			}
		}
	} else {
		// List all tables across all namespaces
		for key, tablemap := range s.tables {
			if !strings.HasPrefix(key, tableBucketArn+"/") {
				continue
			}

			for _, table := range tablemap {
				if prefix != "" && !strings.HasPrefix(table.Name, prefix) {
					continue
				}

				tables = append(tables, TableSummary{
					Arn:        table.Arn,
					Name:       table.Name,
					Namespace:  []string{table.Namespace},
					Type:       table.Type,
					CreatedAt:  table.CreatedAt,
					ModifiedAt: table.ModifiedAt,
				})

				if len(tables) >= maxTables {
					return tables, nil
				}
			}
		}
	}

	return tables, nil
}

// extractBucketNameFromArn extracts the bucket name from a table bucket ARN.
func extractBucketNameFromArn(arn string) string {
	// ARN format: arn:aws:s3tables:region:account:bucket/bucket-name
	parts := strings.Split(arn, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}

	return ""
}
