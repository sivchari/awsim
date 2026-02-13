// Package eventbridge provides AWS EventBridge service emulation.
package eventbridge

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the EventBridge service.
type Service struct {
	storage Storage
}

// New creates a new EventBridge service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "events"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the X-Amz-Target prefix.
func (s *Service) TargetPrefix() string {
	return "AWSEvents"
}

// JSONProtocol marks this service as using JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(_ service.Router) {
	// Routes are handled via X-Amz-Target header dispatching.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
