package dataexchange

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the AWS Data Exchange service.
type Service struct {
	storage Storage
}

// New creates a new Data Exchange service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "dataexchange"
}

// Prefix returns the URL prefix for Data Exchange.
func (s *Service) Prefix() string {
	return "/v1"
}

// RegisterRoutes registers routes with the router.
func (s *Service) RegisterRoutes(r service.Router) {
	// Data set operations.
	r.Handle("POST", "/v1/data-sets", s.CreateDataSet)
	r.Handle("GET", "/v1/data-sets", s.ListDataSets)
	r.Handle("GET", "/v1/data-sets/{dataSetId}", s.GetDataSet)
	r.Handle("PUT", "/v1/data-sets/{dataSetId}", s.UpdateDataSet)
	r.Handle("DELETE", "/v1/data-sets/{dataSetId}", s.DeleteDataSet)

	// Revision operations.
	r.Handle("POST", "/v1/data-sets/{dataSetId}/revisions", s.CreateRevision)
	r.Handle("GET", "/v1/data-sets/{dataSetId}/revisions", s.ListRevisions)
	r.Handle("GET", "/v1/data-sets/{dataSetId}/revisions/{revisionId}", s.GetRevision)
	r.Handle("PUT", "/v1/data-sets/{dataSetId}/revisions/{revisionId}", s.UpdateRevision)
	r.Handle("DELETE", "/v1/data-sets/{dataSetId}/revisions/{revisionId}", s.DeleteRevision)

	// Job operations.
	r.Handle("POST", "/v1/jobs", s.CreateJob)
	r.Handle("GET", "/v1/jobs", s.ListJobs)
	r.Handle("GET", "/v1/jobs/{jobId}", s.GetJob)
}
