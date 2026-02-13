package appsync

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Error codes.
const (
	errNotFound       = "NotFoundException"
	errInvalidRequest = "BadRequestException"
	errConflict       = "ConcurrentModificationException"
)

// Storage defines the interface for AppSync storage operations.
type Storage interface {
	CreateGraphqlAPI(ctx context.Context, input *CreateGraphqlAPIInput) (*GraphqlAPI, error)
	DeleteGraphqlAPI(ctx context.Context, apiID string) error
	GetGraphqlAPI(ctx context.Context, apiID string) (*GraphqlAPI, error)
	ListGraphqlAPIs(ctx context.Context, input *ListGraphqlAPIsInput) ([]GraphqlAPI, string, error)
	CreateDataSource(ctx context.Context, input *CreateDataSourceInput) (*DataSource, error)
	CreateResolver(ctx context.Context, input *CreateResolverInput) (*Resolver, error)
	StartSchemaCreation(ctx context.Context, apiID string, definition []byte) (*SchemaCreationStatus, error)
}

// apiData holds all data associated with a GraphQL API.
type apiData struct {
	api         *GraphqlAPI
	dataSources map[string]*DataSource // key: name
	resolvers   map[string]*Resolver   // key: typeName:fieldName
	schema      *SchemaCreationStatus
}

// MemoryStorage implements Storage with in-memory data structures.
type MemoryStorage struct {
	mu   sync.RWMutex
	apis map[string]*apiData // key: apiID
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		apis: make(map[string]*apiData),
	}
}

// CreateGraphqlAPI creates a new GraphQL API.
func (s *MemoryStorage) CreateGraphqlAPI(_ context.Context, input *CreateGraphqlAPIInput) (*GraphqlAPI, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if input.Name == "" {
		return nil, &Error{
			Code:    errInvalidRequest,
			Message: "Name is required",
		}
	}

	if input.AuthenticationType == "" {
		return nil, &Error{
			Code:    errInvalidRequest,
			Message: "AuthenticationType is required",
		}
	}

	apiID := uuid.New().String()
	apiARN := fmt.Sprintf("arn:aws:appsync:us-east-1:000000000000:apis/%s", apiID)

	api := &GraphqlAPI{
		APIId:              apiID,
		Name:               input.Name,
		AuthenticationType: input.AuthenticationType,
		ARN:                apiARN,
		URIs: map[string]string{
			"GRAPHQL":  fmt.Sprintf("https://%s.appsync-api.us-east-1.amazonaws.com/graphql", apiID),
			"REALTIME": fmt.Sprintf("wss://%s.appsync-realtime-api.us-east-1.amazonaws.com/graphql", apiID),
		},
		Tags:                              input.Tags,
		LogConfig:                         input.LogConfig,
		UserPoolConfig:                    input.UserPoolConfig,
		OpenIDConnectConfig:               input.OpenIDConnectConfig,
		AdditionalAuthenticationProviders: input.AdditionalAuthenticationProviders,
		XrayEnabled:                       input.XrayEnabled,
		LambdaAuthorizerConfig:            input.LambdaAuthorizerConfig,
		Visibility:                        input.Visibility,
		APIType:                           input.APIType,
		MergedAPIExecutionRoleARN:         input.MergedAPIExecutionRoleARN,
		OwnerContact:                      input.OwnerContact,
		IntrospectionConfig:               input.IntrospectionConfig,
		QueryDepthLimit:                   input.QueryDepthLimit,
		ResolverCountLimit:                input.ResolverCountLimit,
		EnhancedMetricsConfig:             input.EnhancedMetricsConfig,
	}

	s.apis[apiID] = &apiData{
		api:         api,
		dataSources: make(map[string]*DataSource),
		resolvers:   make(map[string]*Resolver),
	}

	return api, nil
}

// DeleteGraphqlAPI deletes a GraphQL API.
func (s *MemoryStorage) DeleteGraphqlAPI(_ context.Context, apiID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.apis[apiID]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("GraphQL API %s not found", apiID),
		}
	}

	delete(s.apis, apiID)

	return nil
}

// GetGraphqlAPI retrieves a GraphQL API.
func (s *MemoryStorage) GetGraphqlAPI(_ context.Context, apiID string) (*GraphqlAPI, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.apis[apiID]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("GraphQL API %s not found", apiID),
		}
	}

	return data.api, nil
}

// ListGraphqlAPIs lists all GraphQL APIs.
func (s *MemoryStorage) ListGraphqlAPIs(_ context.Context, input *ListGraphqlAPIsInput) ([]GraphqlAPI, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apis := make([]GraphqlAPI, 0, len(s.apis))

	for _, data := range s.apis {
		// Filter by API type if specified.
		if input.APIType != "" && data.api.APIType != input.APIType {
			continue
		}

		// Filter by owner if specified.
		if input.Owner != "" && data.api.Owner != input.Owner {
			continue
		}

		apis = append(apis, *data.api)
	}

	// Note: Pagination is not implemented in this basic version.
	return apis, "", nil
}

// CreateDataSource creates a new data source for a GraphQL API.
func (s *MemoryStorage) CreateDataSource(_ context.Context, input *CreateDataSourceInput) (*DataSource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.apis[input.APIID]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("GraphQL API %s not found", input.APIID),
		}
	}

	if input.Name == "" {
		return nil, &Error{
			Code:    errInvalidRequest,
			Message: "Name is required",
		}
	}

	if input.Type == "" {
		return nil, &Error{
			Code:    errInvalidRequest,
			Message: "Type is required",
		}
	}

	if _, exists := data.dataSources[input.Name]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Data source %s already exists", input.Name),
		}
	}

	dataSourceARN := fmt.Sprintf("arn:aws:appsync:us-east-1:000000000000:apis/%s/datasources/%s",
		input.APIID, input.Name)

	dataSource := &DataSource{
		DataSourceARN:            dataSourceARN,
		Name:                     input.Name,
		Description:              input.Description,
		Type:                     input.Type,
		ServiceRoleARN:           input.ServiceRoleARN,
		DynamoDBConfig:           input.DynamoDBConfig,
		LambdaConfig:             input.LambdaConfig,
		ElasticsearchConfig:      input.ElasticsearchConfig,
		OpenSearchServiceConfig:  input.OpenSearchServiceConfig,
		HTTPConfig:               input.HTTPConfig,
		RelationalDatabaseConfig: input.RelationalDatabaseConfig,
		EventBridgeConfig:        input.EventBridgeConfig,
		MetricsConfig:            input.MetricsConfig,
	}

	data.dataSources[input.Name] = dataSource

	return dataSource, nil
}

// CreateResolver creates a new resolver for a GraphQL API.
func (s *MemoryStorage) CreateResolver(_ context.Context, input *CreateResolverInput) (*Resolver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.apis[input.APIID]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("GraphQL API %s not found", input.APIID),
		}
	}

	if input.TypeName == "" {
		return nil, &Error{
			Code:    errInvalidRequest,
			Message: "TypeName is required",
		}
	}

	if input.FieldName == "" {
		return nil, &Error{
			Code:    errInvalidRequest,
			Message: "FieldName is required",
		}
	}

	resolverKey := fmt.Sprintf("%s:%s", input.TypeName, input.FieldName)

	if _, exists := data.resolvers[resolverKey]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Resolver for %s.%s already exists", input.TypeName, input.FieldName),
		}
	}

	resolverARN := fmt.Sprintf("arn:aws:appsync:us-east-1:000000000000:apis/%s/types/%s/resolvers/%s",
		input.APIID, input.TypeName, input.FieldName)

	resolver := &Resolver{
		TypeName:                input.TypeName,
		FieldName:               input.FieldName,
		DataSourceName:          input.DataSourceName,
		ResolverARN:             resolverARN,
		RequestMappingTemplate:  input.RequestMappingTemplate,
		ResponseMappingTemplate: input.ResponseMappingTemplate,
		Kind:                    input.Kind,
		PipelineConfig:          input.PipelineConfig,
		SyncConfig:              input.SyncConfig,
		CachingConfig:           input.CachingConfig,
		MaxBatchSize:            input.MaxBatchSize,
		Runtime:                 input.Runtime,
		Code:                    input.Code,
		MetricsConfig:           input.MetricsConfig,
	}

	data.resolvers[resolverKey] = resolver

	return resolver, nil
}

// StartSchemaCreation starts schema creation for a GraphQL API.
func (s *MemoryStorage) StartSchemaCreation(_ context.Context, apiID string, _ []byte) (*SchemaCreationStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.apis[apiID]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("GraphQL API %s not found", apiID),
		}
	}

	// In a real implementation, we would parse and validate the schema.
	// For the emulator, we just mark it as successful.
	status := &SchemaCreationStatus{
		Status:  SchemaStatusSuccess,
		Details: "Schema created successfully",
	}

	data.schema = status

	return status, nil
}
