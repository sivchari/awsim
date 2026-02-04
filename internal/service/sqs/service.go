package sqs

import (
	"github.com/sivchari/awsim/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

func init() {
	service.Register(New(NewMemoryStorage(defaultBaseURL), defaultBaseURL))
}

// Service implements the SQS service.
type Service struct {
	storage Storage
	baseURL string
}

// New creates a new SQS service.
func New(storage Storage, baseURL string) *Service {
	return &Service{
		storage: storage,
		baseURL: baseURL,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "sqs"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the SQS routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// SQS uses POST with X-Amz-Target header for all operations (AWS JSON 1.0 protocol).
	r.HandleFunc("POST", "/", s.dispatchAction)
}
