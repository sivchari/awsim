// Package firehose provides a mock implementation of Amazon Data Firehose.
package firehose

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service is the Firehose service.
type Service struct {
	storage Storage
}

// New creates a new Firehose service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "firehose"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the X-Amz-Target prefix.
func (s *Service) TargetPrefix() string {
	return "Firehose_20150804"
}

// JSONProtocol marks this service as using AWS JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// Firehose uses AWS JSON 1.1 protocol with X-Amz-Target header.
	// Routes are dispatched by the server based on X-Amz-Target.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
