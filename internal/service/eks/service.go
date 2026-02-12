// Package eks provides an EKS service emulator.
package eks

import (
	"net/http"

	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the EKS service.
type Service struct {
	storage Storage
}

// New creates a new EKS service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "eks"
}

// Prefix returns the URL prefix for the service.
// Note: EKS uses /eks prefix to avoid conflicts with S3 wildcard routes.
func (s *Service) Prefix() string {
	return "/eks"
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Cluster operations
	r.HandleFunc("POST", "/eks/clusters", s.CreateCluster)
	r.HandleFunc("DELETE", "/eks/clusters/{name}", s.DeleteCluster)
	r.HandleFunc("GET", "/eks/clusters/{name}", s.handleClusterGet)
	r.HandleFunc("GET", "/eks/clusters", s.ListClusters)

	// Nodegroup operations
	r.HandleFunc("POST", "/eks/clusters/{name}/node-groups", s.CreateNodegroup)
	r.HandleFunc("DELETE", "/eks/clusters/{name}/node-groups/{nodegroupName}", s.DeleteNodegroup)
	r.HandleFunc("GET", "/eks/clusters/{name}/node-groups/{nodegroupName}", s.DescribeNodegroup)
	r.HandleFunc("GET", "/eks/clusters/{name}/node-groups", s.ListNodegroups)
}

// handleClusterGet handles GET requests to /clusters/{name}.
// This is needed because the router might match both DescribeCluster and ListNodegroups.
func (s *Service) handleClusterGet(w http.ResponseWriter, r *http.Request) {
	s.DescribeCluster(w, r)
}
