package appsync

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the AppSync service.
type Service struct {
	storage Storage
}

// New creates a new AppSync service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "appsync"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return "/appsync"
}

// RegisterRoutes registers the AppSync routes.
// AppSync uses REST API protocol.
// Note: Routes use /appsync prefix to avoid conflicts with S3 wildcard routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// GraphQL API operations.
	r.HandleFunc("POST", "/appsync/apis", s.CreateGraphqlAPI)
	r.HandleFunc("DELETE", "/appsync/apis/{apiId}", s.DeleteGraphqlAPI)
	r.HandleFunc("GET", "/appsync/apis/{apiId}", s.GetGraphqlAPI)
	r.HandleFunc("GET", "/appsync/apis", s.ListGraphqlAPIs)

	// Data source operations.
	r.HandleFunc("POST", "/appsync/apis/{apiId}/datasources", s.CreateDataSource)

	// Resolver operations.
	r.HandleFunc("POST", "/appsync/apis/{apiId}/types/{typeName}/resolvers", s.CreateResolver)

	// Schema operations.
	r.HandleFunc("POST", "/appsync/apis/{apiId}/schemacreation", s.StartSchemaCreation)
}
