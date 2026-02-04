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
func (s *Service) RegisterRoutes(r service.Router) {
	// DynamoDB uses POST with X-Amz-Target header for all operations (AWS JSON 1.0 protocol).
	r.HandleFunc("POST", "/", s.dispatchAction)
}
