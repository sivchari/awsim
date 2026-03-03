package rekognition

import (
	"github.com/sivchari/awsim/internal/service"
)

// Compile-time check to ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the AWS Rekognition service.
type Service struct {
	storage Storage
}

// New creates a new Rekognition service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "rekognition"
}

// TargetPrefix returns the X-Amz-Target prefix for JSON protocol dispatch.
func (s *Service) TargetPrefix() string {
	return "RekognitionService"
}

// JSONProtocol marks this service as using JSON protocol.
func (s *Service) JSONProtocol() {}

// RegisterRoutes registers the routes for this service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// JSON protocol services use DispatchAction for routing
}
