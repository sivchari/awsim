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
// S3 Tables routes don't have a common prefix, so we return empty.
// The router handles S3 Tables paths (/buckets, /namespaces, /tables, /get-table) explicitly.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Table bucket operations
	// SDK uses PUT for Create, GET for List/Get, DELETE for Delete
	r.HandleFunc("PUT", "/buckets", s.CreateTableBucket)
	r.HandleFunc("DELETE", "/buckets/{tableBucketARN}", s.DeleteTableBucket)
	r.HandleFunc("GET", "/buckets/{tableBucketARN}", s.GetTableBucket)
	r.HandleFunc("GET", "/buckets", s.ListTableBuckets)

	// Namespace operations
	r.HandleFunc("PUT", "/namespaces/{tableBucketARN}", s.CreateNamespace)
	r.HandleFunc("DELETE", "/namespaces/{tableBucketARN}/{namespace}", s.DeleteNamespace)
	r.HandleFunc("GET", "/namespaces/{tableBucketARN}/{namespace}", s.GetNamespace)
	r.HandleFunc("GET", "/namespaces/{tableBucketARN}", s.ListNamespaces)

	// Table operations
	r.HandleFunc("PUT", "/tables/{tableBucketARN}/{namespace}", s.CreateTable)
	r.HandleFunc("DELETE", "/tables/{tableBucketARN}/{namespace}/{tableName}", s.DeleteTable)
	r.HandleFunc("GET", "/get-table", s.GetTable)
	r.HandleFunc("GET", "/tables/{tableBucketARN}/{namespace}", s.ListTables)
	r.HandleFunc("GET", "/tables/{tableBucketARN}", s.ListTables)
}
