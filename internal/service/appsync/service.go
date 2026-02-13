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
func (s *Service) RegisterRoutes(r service.Router) {
	// GraphQL API operations.
	r.HandleFunc("POST", "/apis", s.CreateGraphqlAPI)
	r.HandleFunc("DELETE", "/apis/{apiId}", s.DeleteGraphqlAPI)
	r.HandleFunc("GET", "/apis/{apiId}", s.GetGraphqlAPI)
	r.HandleFunc("GET", "/apis", s.ListGraphqlAPIs)

	// Data source operations.
	r.HandleFunc("POST", "/apis/{apiId}/datasources", s.CreateDataSource)

	// Resolver operations.
	r.HandleFunc("POST", "/apis/{apiId}/types/{typeName}/resolvers", s.CreateResolver)

	// Schema operations.
	r.HandleFunc("POST", "/apis/{apiId}/schemacreation", s.StartSchemaCreation)
}
