package apigateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errRestAPINotFound    = "NotFoundException"
	errResourceNotFound   = "NotFoundException"
	errMethodNotFound     = "NotFoundException"
	errDeploymentNotFound = "NotFoundException"
	errStageNotFound      = "NotFoundException"
	errBadRequest         = "BadRequestException"
)

// Storage defines the API Gateway storage interface.
type Storage interface {
	CreateRestAPI(ctx context.Context, req *CreateRestAPIRequest) (*RestAPI, error)
	GetRestAPI(ctx context.Context, restAPIID string) (*RestAPI, error)
	GetRestAPIs(ctx context.Context, limit int32, position string) ([]*RestAPI, string, error)
	DeleteRestAPI(ctx context.Context, restAPIID string) error

	CreateResource(ctx context.Context, restAPIID, parentID, pathPart string) (*Resource, error)
	GetResource(ctx context.Context, restAPIID, resourceID string) (*Resource, error)
	GetResources(ctx context.Context, restAPIID string, limit int32, position string) ([]*Resource, string, error)
	DeleteResource(ctx context.Context, restAPIID, resourceID string) error

	PutMethod(ctx context.Context, restAPIID, resourceID, httpMethod string, req *PutMethodRequest) (*Method, error)
	GetMethod(ctx context.Context, restAPIID, resourceID, httpMethod string) (*Method, error)

	PutIntegration(ctx context.Context, restAPIID, resourceID, httpMethod string, req *PutIntegrationRequest) (*Integration, error)
	GetIntegration(ctx context.Context, restAPIID, resourceID, httpMethod string) (*Integration, error)

	CreateDeployment(ctx context.Context, restAPIID string, req *CreateDeploymentRequest) (*Deployment, error)
	GetDeployment(ctx context.Context, restAPIID, deploymentID string) (*Deployment, error)
	GetDeployments(ctx context.Context, restAPIID string, limit int32, position string) ([]*Deployment, string, error)
	DeleteDeployment(ctx context.Context, restAPIID, deploymentID string) error

	CreateStage(ctx context.Context, restAPIID string, req *CreateStageRequest) (*Stage, error)
	GetStage(ctx context.Context, restAPIID, stageName string) (*Stage, error)
	GetStages(ctx context.Context, restAPIID string) ([]*Stage, error)
	DeleteStage(ctx context.Context, restAPIID, stageName string) error
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu       sync.RWMutex
	restAPIs map[string]*restAPIData
}

// restAPIData holds REST API information and its resources.
type restAPIData struct {
	api         *RestAPI
	resources   map[string]*Resource // keyed by resource ID
	deployments map[string]*Deployment
	stages      map[string]*Stage // keyed by stage name
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		restAPIs: make(map[string]*restAPIData),
	}
}

// CreateRestAPI creates a new REST API.
func (s *MemoryStorage) CreateRestAPI(_ context.Context, req *CreateRestAPIRequest) (*RestAPI, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	rootResourceID := generateID()
	now := time.Now()

	api := &RestAPI{
		ID:                     id,
		Name:                   req.Name,
		Description:            req.Description,
		CreatedDate:            now,
		Version:                req.Version,
		APIKeySource:           req.APIKeySource,
		EndpointConfiguration:  req.EndpointConfiguration,
		DisableExecuteAPIEndpt: req.DisableExecuteAPIEndpt,
		Tags:                   req.Tags,
		RootResourceID:         rootResourceID,
	}

	// Create root resource.
	rootResource := &Resource{
		ID:              rootResourceID,
		Path:            "/",
		ResourceMethods: make(map[string]Method),
	}

	s.restAPIs[id] = &restAPIData{
		api:         api,
		resources:   map[string]*Resource{rootResourceID: rootResource},
		deployments: make(map[string]*Deployment),
		stages:      make(map[string]*Stage),
	}

	return api, nil
}

// GetRestAPI returns a REST API by ID.
func (s *MemoryStorage) GetRestAPI(_ context.Context, restAPIID string) (*RestAPI, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	return data.api, nil
}

// GetRestAPIs returns all REST APIs.
func (s *MemoryStorage) GetRestAPIs(_ context.Context, limit int32, _ string) ([]*RestAPI, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 25
	}

	var apis []*RestAPI

	for _, data := range s.restAPIs {
		apis = append(apis, data.api)

		if int32(len(apis)) >= limit { //nolint:gosec // slice length bounded by limit parameter
			break
		}
	}

	return apis, "", nil
}

// DeleteRestAPI deletes a REST API.
func (s *MemoryStorage) DeleteRestAPI(_ context.Context, restAPIID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.restAPIs[restAPIID]; !exists {
		return &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	delete(s.restAPIs, restAPIID)

	return nil
}

// CreateResource creates a new resource.
func (s *MemoryStorage) CreateResource(_ context.Context, restAPIID, parentID, pathPart string) (*Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	parent, exists := data.resources[parentID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	id := generateID()
	path := buildPath(parent.Path, pathPart)

	resource := &Resource{
		ID:              id,
		ParentID:        parentID,
		PathPart:        pathPart,
		Path:            path,
		ResourceMethods: make(map[string]Method),
	}

	data.resources[id] = resource

	return resource, nil
}

// GetResource returns a resource by ID.
func (s *MemoryStorage) GetResource(_ context.Context, restAPIID, resourceID string) (*Resource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	return resource, nil
}

// GetResources returns all resources for a REST API.
func (s *MemoryStorage) GetResources(_ context.Context, restAPIID string, limit int32, _ string) ([]*Resource, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, "", &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	if limit <= 0 {
		limit = 25
	}

	var resources []*Resource

	for _, r := range data.resources {
		resources = append(resources, r)

		if int32(len(resources)) >= limit { //nolint:gosec // slice length bounded by limit parameter
			break
		}
	}

	return resources, "", nil
}

// DeleteResource deletes a resource.
func (s *MemoryStorage) DeleteResource(_ context.Context, restAPIID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	// Cannot delete root resource.
	if resource.Path == "/" {
		return &ServiceError{Code: errBadRequest, Message: "Cannot delete root resource"}
	}

	delete(data.resources, resourceID)

	return nil
}

// PutMethod creates or updates a method.
func (s *MemoryStorage) PutMethod(_ context.Context, restAPIID, resourceID, httpMethod string, req *PutMethodRequest) (*Method, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	method := Method{
		HTTPMethod:        httpMethod,
		AuthorizationType: req.AuthorizationType,
		APIKeyRequired:    req.APIKeyRequired,
		OperationName:     req.OperationName,
	}

	resource.ResourceMethods[httpMethod] = method

	return &method, nil
}

// GetMethod returns a method.
func (s *MemoryStorage) GetMethod(_ context.Context, restAPIID, resourceID, httpMethod string) (*Method, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	method, exists := resource.ResourceMethods[httpMethod]
	if !exists {
		return nil, &ServiceError{Code: errMethodNotFound, Message: "Invalid method identifier specified"}
	}

	return &method, nil
}

// PutIntegration creates or updates an integration.
func (s *MemoryStorage) PutIntegration(_ context.Context, restAPIID, resourceID, httpMethod string, req *PutIntegrationRequest) (*Integration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	method, exists := resource.ResourceMethods[httpMethod]
	if !exists {
		return nil, &ServiceError{Code: errMethodNotFound, Message: "Invalid method identifier specified"}
	}

	integration := &Integration{
		Type:                req.Type,
		HTTPMethod:          req.HTTPMethod,
		URI:                 req.URI,
		ConnectionType:      req.ConnectionType,
		ConnectionID:        req.ConnectionID,
		PassthroughBehavior: req.PassthroughBehavior,
		ContentHandling:     req.ContentHandling,
		TimeoutInMillis:     req.TimeoutInMillis,
		CacheNamespace:      req.CacheNamespace,
		CacheKeyParameters:  req.CacheKeyParameters,
		RequestParameters:   req.RequestParameters,
		RequestTemplates:    req.RequestTemplates,
	}

	method.MethodIntegration = integration
	resource.ResourceMethods[httpMethod] = method

	return integration, nil
}

// GetIntegration returns an integration.
func (s *MemoryStorage) GetIntegration(_ context.Context, restAPIID, resourceID, httpMethod string) (*Integration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	method, exists := resource.ResourceMethods[httpMethod]
	if !exists {
		return nil, &ServiceError{Code: errMethodNotFound, Message: "Invalid method identifier specified"}
	}

	if method.MethodIntegration == nil {
		return nil, &ServiceError{Code: errMethodNotFound, Message: "Integration not found"}
	}

	return method.MethodIntegration, nil
}

// CreateDeployment creates a new deployment.
func (s *MemoryStorage) CreateDeployment(_ context.Context, restAPIID string, req *CreateDeploymentRequest) (*Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	id := generateID()
	now := time.Now()

	deployment := &Deployment{
		ID:          id,
		Description: req.Description,
		CreatedDate: now,
	}

	data.deployments[id] = deployment

	// If stage name is specified, create or update the stage.
	if req.StageName != "" {
		stage := &Stage{
			StageName:       req.StageName,
			DeploymentID:    id,
			CreatedDate:     now,
			LastUpdatedDate: now,
		}
		data.stages[req.StageName] = stage
	}

	return deployment, nil
}

// GetDeployment returns a deployment.
func (s *MemoryStorage) GetDeployment(_ context.Context, restAPIID, deploymentID string) (*Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	deployment, exists := data.deployments[deploymentID]
	if !exists {
		return nil, &ServiceError{Code: errDeploymentNotFound, Message: "Invalid deployment identifier specified"}
	}

	return deployment, nil
}

// GetDeployments returns all deployments.
func (s *MemoryStorage) GetDeployments(_ context.Context, restAPIID string, limit int32, _ string) ([]*Deployment, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, "", &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	if limit <= 0 {
		limit = 25
	}

	var deployments []*Deployment

	for _, d := range data.deployments {
		deployments = append(deployments, d)

		if int32(len(deployments)) >= limit { //nolint:gosec // slice length bounded by limit parameter
			break
		}
	}

	return deployments, "", nil
}

// DeleteDeployment deletes a deployment.
func (s *MemoryStorage) DeleteDeployment(_ context.Context, restAPIID, deploymentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	if _, exists := data.deployments[deploymentID]; !exists {
		return &ServiceError{Code: errDeploymentNotFound, Message: "Invalid deployment identifier specified"}
	}

	delete(data.deployments, deploymentID)

	return nil
}

// CreateStage creates a new stage.
func (s *MemoryStorage) CreateStage(_ context.Context, restAPIID string, req *CreateStageRequest) (*Stage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	// Verify deployment exists.
	if _, exists := data.deployments[req.DeploymentID]; !exists {
		return nil, &ServiceError{Code: errDeploymentNotFound, Message: "Invalid deployment identifier specified"}
	}

	now := time.Now()

	stage := &Stage{
		StageName:           req.StageName,
		DeploymentID:        req.DeploymentID,
		Description:         req.Description,
		CacheClusterEnabled: req.CacheClusterEnabled,
		CacheClusterSize:    req.CacheClusterSize,
		CreatedDate:         now,
		LastUpdatedDate:     now,
		Tags:                req.Tags,
	}

	data.stages[req.StageName] = stage

	return stage, nil
}

// GetStage returns a stage.
func (s *MemoryStorage) GetStage(_ context.Context, restAPIID, stageName string) (*Stage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	stage, exists := data.stages[stageName]
	if !exists {
		return nil, &ServiceError{Code: errStageNotFound, Message: "Invalid stage identifier specified"}
	}

	return stage, nil
}

// GetStages returns all stages.
func (s *MemoryStorage) GetStages(_ context.Context, restAPIID string) ([]*Stage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return nil, &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	var stages []*Stage

	for _, stage := range data.stages {
		stages = append(stages, stage)
	}

	return stages, nil
}

// DeleteStage deletes a stage.
func (s *MemoryStorage) DeleteStage(_ context.Context, restAPIID, stageName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restAPIs[restAPIID]
	if !exists {
		return &ServiceError{Code: errRestAPINotFound, Message: "Invalid REST API identifier specified"}
	}

	if _, exists := data.stages[stageName]; !exists {
		return &ServiceError{Code: errStageNotFound, Message: "Invalid stage identifier specified"}
	}

	delete(data.stages, stageName)

	return nil
}

// generateID generates a unique ID.
func generateID() string {
	return uuid.New().String()[:10]
}

// buildPath builds a full path from parent path and path part.
func buildPath(parentPath, pathPart string) string {
	if parentPath == "/" {
		return "/" + pathPart
	}

	return fmt.Sprintf("%s/%s", parentPath, pathPart)
}
