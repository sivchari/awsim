// Package sagemaker provides SageMaker service emulation for awsim.
package sagemaker

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the SageMaker service.
type Service struct {
	storage Storage
}

// New creates a new SageMaker service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "sagemaker"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the SageMaker routes.
// Note: SageMaker uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - SageMaker uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for SageMaker.
func (s *Service) TargetPrefix() string {
	return "SageMaker"
}

// JSONProtocol is a marker method that indicates SageMaker uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}

// Ensure Service implements service.Service and service.JSONProtocolService.
var (
	_ service.Service             = (*Service)(nil)
	_ service.JSONProtocolService = (*Service)(nil)
)
