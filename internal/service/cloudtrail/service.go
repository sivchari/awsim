// Package cloudtrail provides a mock implementation of AWS CloudTrail.
package cloudtrail

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service is the CloudTrail service.
type Service struct {
	storage Storage
}

// New creates a new CloudTrail service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "cloudtrail"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the X-Amz-Target prefix.
func (s *Service) TargetPrefix() string {
	return "com.amazonaws.cloudtrail.v20131101.CloudTrail_20131101"
}

// JSONProtocol marks this service as using AWS JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// CloudTrail uses AWS JSON 1.1 protocol with X-Amz-Target header.
	// Routes are dispatched by the server based on X-Amz-Target.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
