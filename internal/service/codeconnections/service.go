package codeconnections

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the AWS CodeConnections service.
type Service struct {
	storage Storage
}

// New creates a new CodeConnections service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "codeconnections"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the X-Amz-Target prefix for JSON protocol dispatch.
func (s *Service) TargetPrefix() string {
	return "CodeConnections_20231201"
}

// JSONProtocol marks this service as using JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers the routes for this service.
func (s *Service) RegisterRoutes(r service.Router) {
	// JSON protocol services use DispatchAction for routing
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
