// Package servicequotas provides AWS Service Quotas service emulation.
package servicequotas

import "github.com/sivchari/awsim/internal/service"

// Service implements the Service Quotas service.
type Service struct {
	storage Storage
}

// New creates a new Service Quotas service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "service-quotas"
}

// TargetPrefix returns the target prefix for AWS JSON protocol.
func (s *Service) TargetPrefix() string {
	return "ServiceQuotasV20190624"
}

// JSONProtocol indicates this service uses AWS JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers HTTP routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {}

func init() {
	service.Register(New(NewMemoryStorage()))
}

var (
	_ service.Service             = (*Service)(nil)
	_ service.JSONProtocolService = (*Service)(nil)
)
