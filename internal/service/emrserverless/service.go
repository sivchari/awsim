package emrserverless

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the EMR Serverless service.
type Service struct {
	storage Storage
}

// New creates a new EMR Serverless service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Name returns the service name.
func (s *Service) Name() string {
	return "emrserverless"
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Application operations.
	r.Handle("POST", "/applications", s.CreateApplication)
	r.Handle("GET", "/applications/{applicationId}", s.GetApplication)
	r.Handle("GET", "/applications", s.ListApplications)
	r.Handle("PATCH", "/applications/{applicationId}", s.UpdateApplication)
	r.Handle("DELETE", "/applications/{applicationId}", s.DeleteApplication)
	r.Handle("POST", "/applications/{applicationId}/start", s.StartApplication)
	r.Handle("POST", "/applications/{applicationId}/stop", s.StopApplication)

	// Job run operations.
	r.Handle("POST", "/applications/{applicationId}/jobruns", s.StartJobRun)
	r.Handle("GET", "/applications/{applicationId}/jobruns/{jobRunId}", s.GetJobRun)
	r.Handle("GET", "/applications/{applicationId}/jobruns", s.ListJobRuns)
	r.Handle("DELETE", "/applications/{applicationId}/jobruns/{jobRunId}", s.CancelJobRun)
}

// Ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)
