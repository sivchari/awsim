package memorydb

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the AWS MemoryDB service.
type Service struct {
	storage Storage
}

// New creates a new MemoryDB service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "memorydb"
}

// TargetPrefix returns the X-Amz-Target prefix for JSON protocol dispatch.
func (s *Service) TargetPrefix() string {
	return "AmazonMemoryDB"
}

// JSONProtocol marks this service as using JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers the routes for this service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// JSON protocol services use DispatchAction for routing
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
