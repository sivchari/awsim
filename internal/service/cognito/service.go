// Package cognito provides AWS Cognito Identity Provider service emulation.
package cognito

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the Cognito Identity Provider service.
type Service struct {
	storage Storage
}

// New creates a new Cognito service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "cognito-idp"
}

// Prefix returns the URL prefix for routing.
func (s *Service) Prefix() string {
	return ""
}

// TargetPrefix returns the AWS JSON target prefix.
func (s *Service) TargetPrefix() string {
	return "AWSCognitoIdentityProviderService"
}

// JSONProtocol marks this service as using AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for REST-based operations.
func (s *Service) RegisterRoutes(_ service.Router) {
	// Cognito uses AWS JSON protocol with X-Amz-Target header.
	// Routes are handled by DispatchAction.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
