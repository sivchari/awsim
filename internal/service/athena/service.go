// Package athena provides Athena service emulation for awsim.
package athena

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Athena service.
type Service struct {
	storage Storage
}

// New creates a new Athena service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "athena"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Athena routes.
// Note: Athena uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - Athena uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for Athena.
func (s *Service) TargetPrefix() string {
	return "AmazonAthena"
}

// JSONProtocol is a marker method that indicates Athena uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
