// Package globalaccelerator provides AWS Global Accelerator service emulation.
package globalaccelerator

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the Global Accelerator service.
type Service struct {
	storage Storage
}

// New creates a new Global Accelerator service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "globalaccelerator"
}

// Prefix returns the URL prefix for routing.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the AWS JSON target prefix.
func (s *Service) TargetPrefix() string {
	return "GlobalAccelerator_V20180706"
}

// JSONProtocol marks this service as using AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for REST-based operations.
func (s *Service) RegisterRoutes(_ service.Router) {
	// Global Accelerator uses AWS JSON protocol with X-Amz-Target header.
	// Routes are handled by DispatchAction.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
