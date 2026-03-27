package s3

import (
	"fmt"
	"io"
	"os"

	"github.com/sivchari/kumo/internal/service"
)

// Compile-time check that Service implements io.Closer.
var _ io.Closer = (*Service)(nil)

func init() {
	var opts []Option
	if dir := os.Getenv("KUMO_DATA_DIR"); dir != "" {
		opts = append(opts, WithDataDir(dir))
	}

	service.Register(New(NewMemoryStorage(opts...)))
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

// RegisterRoutes registers the S3 routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Bucket operations
	r.Handle("GET", "/", s.ListBuckets)
	r.Handle("PUT", "/{bucket}", s.handleBucketPut)
	r.Handle("DELETE", "/{bucket}", s.DeleteBucket)
	r.Handle("HEAD", "/{bucket}", s.HeadBucket)

	// Bucket-level GET handles ListObjects, ListMultipartUploads, versioning queries
	r.Handle("GET", "/{bucket}", s.handleBucketGet)

	// Object operations with multipart upload support
	r.Handle("PUT", "/{bucket}/{key...}", s.handleObjectPut)
	r.Handle("GET", "/{bucket}/{key...}", s.handleObjectGet)
	r.Handle("DELETE", "/{bucket}/{key...}", s.handleObjectDelete)
	r.Handle("HEAD", "/{bucket}/{key...}", s.HeadObject)
	r.Handle("POST", "/{bucket}/{key...}", s.handleObjectPost)
}

// Close saves the storage state if persistence is enabled.
func (s *Service) Close() error {
	if c, ok := s.storage.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close storage: %w", err)
		}
	}

	return nil
}
