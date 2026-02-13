// Package apigateway provides API Gateway service emulation for awsim.
package apigateway

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the API Gateway service.
type Service struct {
	storage Storage
}

// New creates a new API Gateway service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "apigateway"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return "/apigateway"
}

// RegisterRoutes registers the API Gateway routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// REST API routes.
	r.HandleFunc("POST", "/apigateway/restapis", s.CreateRestAPI)
	r.HandleFunc("GET", "/apigateway/restapis", s.GetRestAPIs)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}", s.GetRestAPI)
	r.HandleFunc("DELETE", "/apigateway/restapis/{restApiId}", s.DeleteRestAPI)

	// Resource routes.
	r.HandleFunc("POST", "/apigateway/restapis/{restApiId}/resources/{parentId}", s.CreateResource)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/resources", s.GetResources)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/resources/{resourceId}", s.GetResource)
	r.HandleFunc("DELETE", "/apigateway/restapis/{restApiId}/resources/{resourceId}", s.DeleteResource)

	// Method routes.
	r.HandleFunc("PUT", "/apigateway/restapis/{restApiId}/resources/{resourceId}/methods/{httpMethod}", s.PutMethod)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/resources/{resourceId}/methods/{httpMethod}", s.GetMethod)

	// Integration routes.
	r.HandleFunc("PUT", "/apigateway/restapis/{restApiId}/resources/{resourceId}/methods/{httpMethod}/integration", s.PutIntegration)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/resources/{resourceId}/methods/{httpMethod}/integration", s.GetIntegration)

	// Deployment routes.
	r.HandleFunc("POST", "/apigateway/restapis/{restApiId}/deployments", s.CreateDeployment)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/deployments", s.GetDeployments)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/deployments/{deploymentId}", s.GetDeployment)
	r.HandleFunc("DELETE", "/apigateway/restapis/{restApiId}/deployments/{deploymentId}", s.DeleteDeployment)

	// Stage routes.
	r.HandleFunc("POST", "/apigateway/restapis/{restApiId}/stages", s.CreateStage)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/stages", s.GetStages)
	r.HandleFunc("GET", "/apigateway/restapis/{restApiId}/stages/{stageName}", s.GetStage)
	r.HandleFunc("DELETE", "/apigateway/restapis/{restApiId}/stages/{stageName}", s.DeleteStage)
}
