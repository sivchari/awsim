package resiliencehub

import (
	"net/http"
	"strings"

	"github.com/sivchari/awsim/internal/service"
)

// Compile-time checks to ensure Service implements required interfaces.
var (
	_ service.Service             = (*Service)(nil)
	_ service.JSONProtocolService = (*Service)(nil)
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Resilience Hub service.
type Service struct {
	storage  Storage
	handlers map[string]http.HandlerFunc
}

// New creates a new Resilience Hub service.
func New(storage Storage) *Service {
	s := &Service{
		storage: storage,
	}
	s.initHandlers()

	return s
}

// initHandlers initializes the action handlers map.
func (s *Service) initHandlers() {
	s.handlers = map[string]http.HandlerFunc{
		// App operations
		"CreateApp":   s.CreateApp,
		"DescribeApp": s.DescribeApp,
		"UpdateApp":   s.UpdateApp,
		"DeleteApp":   s.DeleteApp,
		"ListApps":    s.ListApps,
		// ResiliencyPolicy operations
		"CreateResiliencyPolicy":   s.CreateResiliencyPolicy,
		"DescribeResiliencyPolicy": s.DescribeResiliencyPolicy,
		"UpdateResiliencyPolicy":   s.UpdateResiliencyPolicy,
		"DeleteResiliencyPolicy":   s.DeleteResiliencyPolicy,
		"ListResiliencyPolicies":   s.ListResiliencyPolicies,
		// Assessment operations
		"StartAppAssessment":    s.StartAppAssessment,
		"DescribeAppAssessment": s.DescribeAppAssessment,
		"DeleteAppAssessment":   s.DeleteAppAssessment,
		"ListAppAssessments":    s.ListAppAssessments,
		// Tag operations
		"TagResource":         s.TagResource,
		"UntagResource":       s.UntagResource,
		"ListTagsForResource": s.ListTagsForResource,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "resiliencehub"
}

// RegisterRoutes registers the Resilience Hub routes.
// Note: Resilience Hub uses AWS JSON 1.1 protocol via the JSONProtocolService interface,
// so no direct routes are registered here.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - Resilience Hub uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for Resilience Hub.
func (s *Service) TargetPrefix() string {
	return "AwsResilienceHub"
}

// JSONProtocol is a marker method that indicates Resilience Hub uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}

// DispatchAction handles the JSON protocol request after routing.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	if target == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "MissingAction",
			Message: "X-Amz-Target header is required",
		})

		return
	}

	// Extract the action from the target (e.g., "AwsResilienceHub.CreateApp" -> "CreateApp").
	action := target
	if idx := strings.LastIndex(target, "."); idx != -1 {
		action = target[idx+1:]
	}

	handler, ok := s.handlers[action]
	if !ok {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "UnknownOperationException",
			Message: "Unknown operation: " + action,
		})

		return
	}

	handler(w, r)
}
