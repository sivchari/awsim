package cloudformation

import (
	"net/http"

	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the CloudFormation service.
type Service struct {
	storage Storage
}

// New creates a new CloudFormation service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "cloudformation"
}

// TargetPrefix returns the X-Amz-Target prefix for the service.
func (s *Service) TargetPrefix() string {
	return "CloudFormation"
}

// Actions returns the list of actions supported by this service.
func (s *Service) Actions() []string {
	return []string{
		"CreateStack",
		"DeleteStack",
		"DescribeStacks",
		"ListStacks",
		"UpdateStack",
		"DescribeStackResources",
		"GetTemplate",
		"ValidateTemplate",
	}
}

// QueryProtocol marks this service as using the Query protocol.
func (s *Service) QueryProtocol() {}

// RegisterRoutes registers HTTP routes for the service.
func (s *Service) RegisterRoutes(_ service.Router) {
	// CloudFormation uses Query protocol, routes are handled by the dispatcher.
}

// Compile-time check that Service implements the required interfaces.
var (
	_ service.Service              = (*Service)(nil)
	_ service.QueryProtocolService = (*Service)(nil)
)

// HandleRequest handles HTTP requests for the CloudFormation service.
func (s *Service) HandleRequest(w http.ResponseWriter, r *http.Request) {
	s.DispatchAction(w, r)
}
