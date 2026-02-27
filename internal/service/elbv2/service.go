package elbv2

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the ELB v2 service.
type Service struct {
	storage Storage
}

// New creates a new ELB v2 service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "elasticloadbalancingv2"
}

// RegisterRoutes registers routes with the router.
// ELB uses Query protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// ELB uses Query protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for ELB.
func (s *Service) TargetPrefix() string {
	return "ElasticLoadBalancing"
}

// Actions returns the list of action names this service handles.
func (s *Service) Actions() []string {
	return []string{
		"CreateLoadBalancer",
		"DeleteLoadBalancer",
		"DescribeLoadBalancers",
		"CreateTargetGroup",
		"DeleteTargetGroup",
		"DescribeTargetGroups",
		"RegisterTargets",
		"DeregisterTargets",
		"CreateListener",
		"DeleteListener",
	}
}

// QueryProtocol is a marker method that indicates ELB uses AWS Query protocol.
func (s *Service) QueryProtocol() {}
