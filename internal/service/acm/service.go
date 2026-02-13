package acm

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the ACM service.
type Service struct {
	storage Storage
}

// New creates a new ACM service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "acm"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the ACM routes.
// Note: ACM uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - ACM uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for ACM.
func (s *Service) TargetPrefix() string {
	return "CertificateManager"
}

// JSONProtocol is a marker method that indicates ACM uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
