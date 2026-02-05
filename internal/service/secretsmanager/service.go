package secretsmanager

import (
	"github.com/sivchari/awsim/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

func init() {
	service.Register(New(NewMemoryStorage(defaultBaseURL), defaultBaseURL))
}

// Service implements the Secrets Manager service.
type Service struct {
	storage Storage
	baseURL string
}

// New creates a new Secrets Manager service.
func New(storage Storage, baseURL string) *Service {
	return &Service{
		storage: storage,
		baseURL: baseURL,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "secretsmanager"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Secrets Manager routes.
// Note: Secrets Manager uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - Secrets Manager uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for Secrets Manager.
func (s *Service) TargetPrefix() string {
	return "secretsmanager"
}

// isJSONProtocol is a marker method that indicates Secrets Manager uses AWS JSON 1.1 protocol.
//
//nolint:unused // Marker method for interface compliance.
func (s *Service) isJSONProtocol() {}
