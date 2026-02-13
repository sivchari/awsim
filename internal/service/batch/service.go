package batch

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Batch service.
type Service struct {
	storage Storage
}

// New creates a new Batch service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "batch"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Batch routes.
// Note: Batch uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - Batch uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for Batch.
func (s *Service) TargetPrefix() string {
	return "AWSBatch_V20160810"
}

// JSONProtocol is a marker method that indicates Batch uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
