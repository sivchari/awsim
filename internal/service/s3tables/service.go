// Package s3tables provides S3 Tables service emulation for awsim.
package s3tables

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the S3 Tables service.
type Service struct {
	storage Storage
}

// New creates a new S3 Tables service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "s3tables"
}

// Prefix returns the URL prefix for the service.
// Note: S3 Tables uses /s3tables prefix to avoid conflicts with S3 wildcard routes.
func (s *Service) Prefix() string {
	return "/s3tables"
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Table bucket operations
	r.HandleFunc("POST", "/s3tables/buckets", s.CreateTableBucket)
	r.HandleFunc("DELETE", "/s3tables/buckets/{tableBucketARN}", s.DeleteTableBucket)
	r.HandleFunc("GET", "/s3tables/buckets/{tableBucketARN}", s.GetTableBucket)
	r.HandleFunc("GET", "/s3tables/buckets", s.ListTableBuckets)

	// Namespace operations
	r.HandleFunc("POST", "/s3tables/namespaces/{tableBucketARN}", s.CreateNamespace)
	r.HandleFunc("DELETE", "/s3tables/namespaces/{tableBucketARN}/{namespace}", s.DeleteNamespace)
	r.HandleFunc("GET", "/s3tables/namespaces/{tableBucketARN}/{namespace}", s.GetNamespace)
	r.HandleFunc("GET", "/s3tables/namespaces/{tableBucketARN}", s.ListNamespaces)

	// Table operations
	r.HandleFunc("POST", "/s3tables/tables/{tableBucketARN}/{namespace}", s.CreateTable)
	r.HandleFunc("DELETE", "/s3tables/tables/{tableBucketARN}/{namespace}/{tableName}", s.DeleteTable)
	r.HandleFunc("GET", "/s3tables/tables/{tableBucketARN}/{namespace}/{tableName}", s.GetTable)
	r.HandleFunc("GET", "/s3tables/tables/{tableBucketARN}/{namespace}", s.ListTables)
}
