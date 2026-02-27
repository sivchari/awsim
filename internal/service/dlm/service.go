// Package dlm provides Data Lifecycle Manager service emulation for awsim.
package dlm

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the DLM service.
type Service struct {
	storage Storage
}

// New creates a new DLM service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "dlm"
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	r.HandleFunc("POST", "/dlm/policies", s.CreateLifecyclePolicy)
	r.HandleFunc("GET", "/dlm/policies", s.GetLifecyclePolicies)
	r.HandleFunc("GET", "/dlm/policies/{policyId}", s.GetLifecyclePolicy)
	r.HandleFunc("PATCH", "/dlm/policies/{policyId}", s.UpdateLifecyclePolicy)
	r.HandleFunc("DELETE", "/dlm/policies/{policyId}", s.DeleteLifecyclePolicy)
}

// Ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)
