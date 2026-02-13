// Package ecr provides a mock implementation of AWS Elastic Container Registry.
package ecr

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service is the ECR service.
type Service struct {
	storage Storage
}

// New creates a new ECR service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "ecr"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the X-Amz-Target prefix.
func (s *Service) TargetPrefix() string {
	return "AmazonEC2ContainerRegistry_V20150921"
}

// JSONProtocol marks this service as using AWS JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// ECR uses AWS JSON 1.1 protocol with X-Amz-Target header.
	// Routes are dispatched by the server based on X-Amz-Target.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
