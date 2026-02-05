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
// Note: SQS uses AWS JSON 1.0 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - SQS uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for SQS.
func (s *Service) TargetPrefix() string {
	return "AmazonSQS"
}

// JSONProtocol is a marker method that indicates SQS uses AWS JSON 1.0 protocol.
func (s *Service) JSONProtocol() {}
