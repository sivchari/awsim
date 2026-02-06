package ecs

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the ECS service.
type Service struct {
	storage Storage
}

// New creates a new ECS service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "ecs"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the ECS routes.
// Note: ECS uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - ECS uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for ECS.
func (s *Service) TargetPrefix() string {
	return "AmazonEC2ContainerServiceV20141113"
}

// JSONProtocol is a marker method that indicates ECS uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
