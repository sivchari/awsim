package batch

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Batch service.
type Service struct {
	storage Storage
}

// New creates a new Batch service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "batch"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Batch routes.
// Batch uses REST JSON protocol with paths like /v1/createcomputeenvironment.
func (s *Service) RegisterRoutes(r service.Router) {
	// Compute Environment operations
	r.Handle("POST", "/v1/createcomputeenvironment", s.CreateComputeEnvironment)
	r.Handle("POST", "/v1/deletecomputeenvironment", s.DeleteComputeEnvironment)
	r.Handle("POST", "/v1/describecomputeenvironments", s.DescribeComputeEnvironments)

	// Job Queue operations
	r.Handle("POST", "/v1/createjobqueue", s.CreateJobQueue)
	r.Handle("POST", "/v1/deletejobqueue", s.DeleteJobQueue)
	r.Handle("POST", "/v1/describejobqueues", s.DescribeJobQueues)

	// Job Definition operations
	r.Handle("POST", "/v1/registerjobdefinition", s.RegisterJobDefinition)

	// Job operations
	r.Handle("POST", "/v1/submitjob", s.SubmitJob)
	r.Handle("POST", "/v1/describejobs", s.DescribeJobs)
	r.Handle("POST", "/v1/terminatejob", s.TerminateJob)
}
