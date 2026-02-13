package xray

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the X-Ray service.
type Service struct {
	storage Storage
}

// New creates a new X-Ray service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "xray"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return "/xray"
}

// RegisterRoutes registers the X-Ray routes.
// X-Ray uses REST JSON protocol.
func (s *Service) RegisterRoutes(r service.Router) {
	r.HandleFunc("POST", "/TraceSegments", s.PutTraceSegments)
	r.HandleFunc("POST", "/TraceSummaries", s.GetTraceSummaries)
	r.HandleFunc("POST", "/Traces", s.BatchGetTraces)
	r.HandleFunc("POST", "/ServiceGraph", s.GetServiceGraph)
	r.HandleFunc("POST", "/CreateGroup", s.CreateGroup)
	r.HandleFunc("POST", "/DeleteGroup", s.DeleteGroup)
}
