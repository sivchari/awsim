// Package entityresolution provides an AWS Entity Resolution service emulator.
package entityresolution

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Entity Resolution service.
type Service struct {
	storage Storage
}

// New creates a new Entity Resolution service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "entityresolution"
}

// RegisterRoutes registers the Entity Resolution routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Schema mapping routes.
	r.Handle("POST", "/schemas", s.CreateSchemaMapping)
	r.Handle("GET", "/schemas", s.ListSchemaMappings)
	r.Handle("GET", "/schemas/{schemaName}", s.GetSchemaMapping)
	r.Handle("DELETE", "/schemas/{schemaName}", s.DeleteSchemaMapping)

	// Matching workflow routes.
	r.Handle("POST", "/matchingworkflows", s.CreateMatchingWorkflow)
	r.Handle("GET", "/matchingworkflows", s.ListMatchingWorkflows)
	r.Handle("GET", "/matchingworkflows/{workflowName}", s.GetMatchingWorkflow)
	r.Handle("DELETE", "/matchingworkflows/{workflowName}", s.DeleteMatchingWorkflow)

	// ID mapping workflow routes.
	r.Handle("POST", "/idmappingworkflows", s.CreateIDMappingWorkflow)
	r.Handle("GET", "/idmappingworkflows", s.ListIDMappingWorkflows)
	r.Handle("GET", "/idmappingworkflows/{workflowName}", s.GetIDMappingWorkflow)
	r.Handle("DELETE", "/idmappingworkflows/{workflowName}", s.DeleteIDMappingWorkflow)

	// Provider service routes.
	r.Handle("GET", "/providerservices", s.ListProviderServices)
}
