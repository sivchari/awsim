package s3

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the S3 service.
type Service struct {
	storage Storage
}

// New creates a new S3 service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "s3"
}

// Prefix returns the URL prefix for the service.
// S3 uses path-style URLs, so no prefix is needed.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the S3 routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Bucket operations
	r.Handle("GET", "/", s.ListBuckets)
	r.Handle("PUT", "/{bucket}", s.handleBucketPut)
	r.Handle("DELETE", "/{bucket}", s.DeleteBucket)
	r.Handle("HEAD", "/{bucket}", s.HeadBucket)

	// Object list (must be before object operations to handle query params)
	r.Handle("GET", "/{bucket}", s.handleBucketGet)

	// Object operations
	r.Handle("PUT", "/{bucket}/{key...}", s.PutObject)
	r.Handle("GET", "/{bucket}/{key...}", s.GetObject)
	r.Handle("DELETE", "/{bucket}/{key...}", s.DeleteObject)
	r.Handle("HEAD", "/{bucket}/{key...}", s.HeadObject)
}
