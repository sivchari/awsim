package s3control

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the S3 Control service.
type Service struct {
	storage Storage
}

// New creates a new S3 Control service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "s3control"
}

// RegisterRoutes registers the S3 Control routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Public Access Block operations
	r.Handle("GET", "/v20180820/configuration/publicAccessBlock", s.GetPublicAccessBlock)
	r.Handle("PUT", "/v20180820/configuration/publicAccessBlock", s.PutPublicAccessBlock)
	r.Handle("DELETE", "/v20180820/configuration/publicAccessBlock", s.DeletePublicAccessBlock)

	// Access Point operations
	r.Handle("PUT", "/v20180820/accesspoint/{name}", s.CreateAccessPoint)
	r.Handle("GET", "/v20180820/accesspoint/{name}", s.GetAccessPoint)
	r.Handle("DELETE", "/v20180820/accesspoint/{name}", s.DeleteAccessPoint)
	r.Handle("GET", "/v20180820/accesspoint", s.ListAccessPoints)
}
