// Package forecast provides Amazon Forecast service emulation.
package forecast

import "github.com/sivchari/awsim/internal/service"

// Service implements the Forecast service.
type Service struct {
	storage Storage
}

// New creates a new Forecast service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "forecast"
}

// TargetPrefix returns the target prefix for AWS JSON protocol.
func (s *Service) TargetPrefix() string {
	return "AmazonForecast"
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
