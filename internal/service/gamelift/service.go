// Package gamelift provides a mock implementation of AWS GameLift.
package gamelift

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service is the GameLift service.
type Service struct {
	storage Storage
}

// New creates a new GameLift service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "gamelift"
}

// TargetPrefix returns the X-Amz-Target prefix.
func (s *Service) TargetPrefix() string {
	return "GameLift"
}

// JSONProtocol marks this service as using AWS JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// GameLift uses AWS JSON 1.0 protocol with X-Amz-Target header.
	// Routes are dispatched by the server based on X-Amz-Target.
}

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Ensure Service implements required interfaces.
var (
	_ service.Service             = (*Service)(nil)
	_ service.JSONProtocolService = (*Service)(nil)
)
