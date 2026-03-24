package codeguruprofiler

import (
	"github.com/sivchari/kumo/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the AWS CodeGuru Profiler service.
type Service struct {
	storage Storage
}

// New creates a new CodeGuru Profiler service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "codeguru-profiler"
}

// Prefix returns the URL prefix for CodeGuru Profiler.
func (s *Service) Prefix() string {
	return "/profilingGroups"
}

// RegisterRoutes registers routes with the router.
func (s *Service) RegisterRoutes(r service.Router) {
	r.Handle("POST", "/profilingGroups", s.CreateProfilingGroup)
	r.Handle("GET", "/profilingGroups", s.ListProfilingGroups)
	r.Handle("GET", "/profilingGroups/{profilingGroupName}", s.DescribeProfilingGroup)
	r.Handle("PUT", "/profilingGroups/{profilingGroupName}", s.UpdateProfilingGroup)
	r.Handle("DELETE", "/profilingGroups/{profilingGroupName}", s.DeleteProfilingGroup)
}
