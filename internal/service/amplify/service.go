package amplify

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the Amplify service.
type Service struct {
	storage Storage
}

// New creates a new Amplify service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "amplify"
}

// Prefix returns the URL prefix for Amplify.
func (s *Service) Prefix() string {
	return "/apps"
}

// RegisterRoutes registers routes with the router.
func (s *Service) RegisterRoutes(r service.Router) {
	// App operations.
	r.Handle("POST", "/apps", s.CreateApp)
	r.Handle("GET", "/apps", s.ListApps)
	r.Handle("GET", "/apps/{appId}", s.GetApp)
	r.Handle("POST", "/apps/{appId}", s.UpdateApp)
	r.Handle("DELETE", "/apps/{appId}", s.DeleteApp)

	// Branch operations.
	r.Handle("POST", "/apps/{appId}/branches", s.CreateBranch)
	r.Handle("GET", "/apps/{appId}/branches", s.ListBranches)
	r.Handle("GET", "/apps/{appId}/branches/{branchName}", s.GetBranch)
	r.Handle("DELETE", "/apps/{appId}/branches/{branchName}", s.DeleteBranch)
}
