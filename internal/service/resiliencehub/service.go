package resiliencehub

import "github.com/sivchari/awsim/internal/service"

// Compile-time check to ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Resilience Hub service.
type Service struct {
	storage Storage
}

// New creates a new Resilience Hub service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "resiliencehub"
}

// RegisterRoutes registers the Resilience Hub routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// App operations
	r.Handle("POST", "/apps", s.CreateApp)
	r.Handle("GET", "/apps/{appArn}", s.DescribeApp)
	r.Handle("POST", "/apps/{appArn}", s.UpdateApp)
	r.Handle("DELETE", "/apps/{appArn}", s.DeleteApp)
	r.Handle("GET", "/apps", s.ListApps)

	// ResiliencyPolicy operations
	r.Handle("POST", "/resiliency-policies", s.CreateResiliencyPolicy)
	r.Handle("GET", "/resiliency-policies/{policyArn}", s.DescribeResiliencyPolicy)
	r.Handle("POST", "/resiliency-policies/{policyArn}", s.UpdateResiliencyPolicy)
	r.Handle("DELETE", "/resiliency-policies/{policyArn}", s.DeleteResiliencyPolicy)
	r.Handle("GET", "/resiliency-policies", s.ListResiliencyPolicies)

	// Assessment operations
	r.Handle("POST", "/app-assessments", s.StartAppAssessment)
	r.Handle("GET", "/app-assessments/{assessmentArn}", s.DescribeAppAssessment)
	r.Handle("DELETE", "/app-assessments/{assessmentArn}", s.DeleteAppAssessment)
	r.Handle("GET", "/app-assessments", s.ListAppAssessments)

	// Tag operations
	r.Handle("POST", "/tags/{resourceArn}", s.TagResource)
	r.Handle("DELETE", "/tags/{resourceArn}", s.UntagResource)
	r.Handle("GET", "/tags/{resourceArn}", s.ListTagsForResource)
}

// Prefix returns the URL prefix for Resilience Hub.
func (s *Service) Prefix() string {
	return "/resiliencehub"
}
