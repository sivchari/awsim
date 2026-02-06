package ec2

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the EC2 service.
type Service struct {
	storage Storage
}

// New creates a new EC2 service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "ec2"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers routes with the router.
// EC2 uses Query protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// EC2 uses Query protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for EC2.
// EC2 uses Action parameter instead of X-Amz-Target header.
func (s *Service) TargetPrefix() string {
	return "AmazonEC2"
}

// Actions returns the list of action names this service handles.
func (s *Service) Actions() []string {
	return []string{
		"RunInstances",
		"TerminateInstances",
		"DescribeInstances",
		"StartInstances",
		"StopInstances",
		"CreateSecurityGroup",
		"DeleteSecurityGroup",
		"AuthorizeSecurityGroupIngress",
		"AuthorizeSecurityGroupEgress",
		"CreateKeyPair",
		"DeleteKeyPair",
		"DescribeKeyPairs",
	}
}

// QueryProtocol is a marker method that indicates EC2 uses AWS Query protocol.
func (s *Service) QueryProtocol() {}
