package cloudfront

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the CloudFront service.
type Service struct {
	storage Storage
}

// New creates a new CloudFront service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "cloudfront"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the CloudFront routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Distribution operations.
	r.Handle("POST", "/2020-05-31/distribution", s.CreateDistribution)
	r.Handle("GET", "/2020-05-31/distribution", s.ListDistributions)
	r.Handle("GET", "/2020-05-31/distribution/{id}", s.GetDistribution)
	r.Handle("GET", "/2020-05-31/distribution/{id}/config", s.GetDistributionConfig)
	r.Handle("PUT", "/2020-05-31/distribution/{id}/config", s.UpdateDistribution)
	r.Handle("DELETE", "/2020-05-31/distribution/{id}", s.DeleteDistribution)

	// Invalidation operations.
	r.Handle("POST", "/2020-05-31/distribution/{id}/invalidation", s.CreateInvalidation)
	r.Handle("GET", "/2020-05-31/distribution/{id}/invalidation/{invalidationId}", s.GetInvalidation)
}
