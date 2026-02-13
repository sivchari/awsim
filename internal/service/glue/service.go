package glue

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Glue service.
type Service struct {
	storage Storage
}

// New creates a new Glue service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "glue"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Glue routes.
// Note: Glue uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - Glue uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for Glue.
func (s *Service) TargetPrefix() string {
	return "AWSGlue"
}

// JSONProtocol is a marker method that indicates Glue uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
