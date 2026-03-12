package sts

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the STS service.
type Service struct {
	storage Storage
}

// New creates a new STS service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "sts"
}

// RegisterRoutes registers routes with the router.
// STS uses Query protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// STS uses Query protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for STS.
func (s *Service) TargetPrefix() string {
	return "AWSSecurityTokenServiceV20110615"
}

// Actions returns the list of action names this service handles.
func (s *Service) Actions() []string {
	return []string{
		"AssumeRole",
		"AssumeRoleWithSAML",
		"AssumeRoleWithWebIdentity",
		"GetCallerIdentity",
		"GetSessionToken",
		"GetFederationToken",
	}
}

// QueryProtocol is a marker method that indicates STS uses AWS Query protocol.
func (s *Service) QueryProtocol() {}
