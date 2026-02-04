package dynamodb

import (
	"github.com/sivchari/awsim/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

func init() {
	service.Register(New(NewMemoryStorage(defaultBaseURL)))
}

// Service implements the DynamoDB service.
type Service struct {
	storage Storage
}

// New creates a new DynamoDB service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "dynamodb"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the DynamoDB routes.
// Note: DynamoDB uses AWS JSON 1.0 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - DynamoDB uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for DynamoDB.
func (s *Service) TargetPrefix() string {
	return "DynamoDB_20120810"
}
