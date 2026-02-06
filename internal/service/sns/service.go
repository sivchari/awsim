package sns

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage("")
	service.Register(New(storage))
}

// Service implements the SNS service.
type Service struct {
	storage Storage
}

// New creates a new SNS service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "sns"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers routes with the router.
// SNS uses Query protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// SNS uses Query protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for SNS.
func (s *Service) TargetPrefix() string {
	return "AmazonSimpleNotificationService"
}

// Actions returns the list of action names this service handles.
func (s *Service) Actions() []string {
	return []string{
		"CreateTopic",
		"DeleteTopic",
		"ListTopics",
		"Subscribe",
		"Unsubscribe",
		"Publish",
		"ListSubscriptions",
		"ListSubscriptionsByTopic",
	}
}

// QueryProtocol is a marker method that indicates SNS uses AWS Query protocol.
func (s *Service) QueryProtocol() {}

// Storage returns the SNS storage.
// This can be used to set up cross-service integration (e.g., SNS to SQS).
func (s *Service) Storage() Storage {
	return s.storage
}
