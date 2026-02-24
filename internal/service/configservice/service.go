// Package configservice provides a mock implementation of AWS Config.
package configservice

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service is the Config service.
type Service struct {
	storage Storage
}

// New creates a new Config service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "configservice"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the X-Amz-Target prefix.
func (s *Service) TargetPrefix() string {
	return "StarlingDoveService"
}

// JSONProtocol marks this service as using AWS JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// Config uses AWS JSON 1.1 protocol with X-Amz-Target header.
	// Routes are dispatched by the server based on X-Amz-Target.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Ensure Service implements required interfaces.
var (
	_ service.Service             = (*Service)(nil)
	_ service.JSONProtocolService = (*Service)(nil)
)
