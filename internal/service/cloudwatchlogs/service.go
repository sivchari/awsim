package cloudwatchlogs

import (
	"github.com/sivchari/awsim/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

func init() {
	service.Register(New(NewMemoryStorage(defaultBaseURL), defaultBaseURL))
}

// Service implements the CloudWatch Logs service.
type Service struct {
	storage Storage
	baseURL string
}

// New creates a new CloudWatch Logs service.
func New(storage Storage, baseURL string) *Service {
	return &Service{
		storage: storage,
		baseURL: baseURL,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "logs"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the CloudWatch Logs routes.
// Note: CloudWatch Logs uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - CloudWatch Logs uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for CloudWatch Logs.
func (s *Service) TargetPrefix() string {
	return "Logs_20140328"
}

// JSONProtocol is a marker method that indicates CloudWatch Logs uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
