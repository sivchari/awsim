package cloudwatch

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage("")
	service.Register(New(storage))
}

// Service implements the CloudWatch service.
type Service struct {
	storage Storage
}

// New creates a new CloudWatch service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "monitoring"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers routes with the router.
// CloudWatch uses JSON protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// CloudWatch uses JSON protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for CloudWatch.
func (s *Service) TargetPrefix() string {
	return "GraniteServiceVersion20100801"
}

// JSONProtocol is a marker method that indicates CloudWatch uses AWS JSON 1.0 protocol.
func (s *Service) JSONProtocol() {}
