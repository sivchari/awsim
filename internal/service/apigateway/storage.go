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
	errRestApiNotFound    = "NotFoundException"
	errResourceNotFound   = "NotFoundException"
	errMethodNotFound     = "NotFoundException"
	errDeploymentNotFound = "NotFoundException"
	errStageNotFound      = "NotFoundException"
	errBadRequest         = "BadRequestException"
)

// Storage defines the API Gateway storage interface.
type Storage interface {
	CreateRestApi(ctx context.Context, req *CreateRestApiRequest) (*RestApi, error)
	GetRestApi(ctx context.Context, restApiID string) (*RestApi, error)
	GetRestApis(ctx context.Context, limit int32, position string) ([]*RestApi, string, error)
	DeleteRestApi(ctx context.Context, restApiID string) error

	CreateResource(ctx context.Context, restApiID, parentID, pathPart string) (*Resource, error)
	GetResource(ctx context.Context, restApiID, resourceID string) (*Resource, error)
	GetResources(ctx context.Context, restApiID string, limit int32, position string) ([]*Resource, string, error)
	DeleteResource(ctx context.Context, restApiID, resourceID string) error

	PutMethod(ctx context.Context, restApiID, resourceID, httpMethod string, req *PutMethodRequest) (*Method, error)
	GetMethod(ctx context.Context, restApiID, resourceID, httpMethod string) (*Method, error)

	PutIntegration(ctx context.Context, restApiID, resourceID, httpMethod string, req *PutIntegrationRequest) (*Integration, error)
	GetIntegration(ctx context.Context, restApiID, resourceID, httpMethod string) (*Integration, error)

	CreateDeployment(ctx context.Context, restApiID string, req *CreateDeploymentRequest) (*Deployment, error)
	GetDeployment(ctx context.Context, restApiID, deploymentID string) (*Deployment, error)
	GetDeployments(ctx context.Context, restApiID string, limit int32, position string) ([]*Deployment, string, error)
	DeleteDeployment(ctx context.Context, restApiID, deploymentID string) error

	CreateStage(ctx context.Context, restApiID string, req *CreateStageRequest) (*Stage, error)
	GetStage(ctx context.Context, restApiID, stageName string) (*Stage, error)
	GetStages(ctx context.Context, restApiID string) ([]*Stage, error)
	DeleteStage(ctx context.Context, restApiID, stageName string) error
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu       sync.RWMutex
	restApis map[string]*restApiData
}

// restApiData holds REST API information and its resources.
type restApiData struct {
	api         *RestApi
	resources   map[string]*Resource // keyed by resource ID
	deployments map[string]*Deployment
	stages      map[string]*Stage // keyed by stage name
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		restApis: make(map[string]*restApiData),
	}
}

// CreateRestApi creates a new REST API.
func (s *MemoryStorage) CreateRestApi(_ context.Context, req *CreateRestApiRequest) (*RestApi, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	rootResourceID := generateID()
	now := time.Now()

	api := &RestApi{
		ID:                     id,
		Name:                   req.Name,
		Description:            req.Description,
		CreatedDate:            now,
		Version:                req.Version,
		ApiKeySource:           req.ApiKeySource,
		EndpointConfiguration:  req.EndpointConfiguration,
		DisableExecuteApiEndpt: req.DisableExecuteApiEndpt,
		Tags:                   req.Tags,
		RootResourceID:         rootResourceID,
	}

	// Create root resource.
	rootResource := &Resource{
		ID:              rootResourceID,
		Path:            "/",
		ResourceMethods: make(map[string]Method),
	}

	s.restApis[id] = &restApiData{
		api:         api,
		resources:   map[string]*Resource{rootResourceID: rootResource},
		deployments: make(map[string]*Deployment),
		stages:      make(map[string]*Stage),
	}

	return api, nil
}

// GetRestApi returns a REST API by ID.
func (s *MemoryStorage) GetRestApi(_ context.Context, restApiID string) (*RestApi, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	return data.api, nil
}

// GetRestApis returns all REST APIs.
func (s *MemoryStorage) GetRestApis(_ context.Context, limit int32, _ string) ([]*RestApi, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 25
	}

	var apis []*RestApi

	for _, data := range s.restApis {
		apis = append(apis, data.api)

		if int32(len(apis)) >= limit { //nolint:gosec // slice length bounded by limit parameter
			break
		}
	}

	return apis, "", nil
}

// DeleteRestApi deletes a REST API.
func (s *MemoryStorage) DeleteRestApi(_ context.Context, restApiID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.restApis[restApiID]; !exists {
		return &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	delete(s.restApis, restApiID)

	return nil
}

// CreateResource creates a new resource.
func (s *MemoryStorage) CreateResource(_ context.Context, restApiID, parentID, pathPart string) (*Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) GetResource(_ context.Context, restApiID, resourceID string) (*Resource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	return resource, nil
}

// GetResources returns all resources for a REST API.
func (s *MemoryStorage) GetResources(_ context.Context, restApiID string, limit int32, _ string) ([]*Resource, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, "", &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) DeleteResource(_ context.Context, restApiID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) PutMethod(_ context.Context, restApiID, resourceID, httpMethod string, req *PutMethodRequest) (*Method, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	resource, exists := data.resources[resourceID]
	if !exists {
		return nil, &ServiceError{Code: errResourceNotFound, Message: "Invalid resource identifier specified"}
	}

	method := Method{
		HTTPMethod:        httpMethod,
		AuthorizationType: req.AuthorizationType,
		ApiKeyRequired:    req.ApiKeyRequired,
		OperationName:     req.OperationName,
	}

	resource.ResourceMethods[httpMethod] = method

	return &method, nil
}

// GetMethod returns a method.
func (s *MemoryStorage) GetMethod(_ context.Context, restApiID, resourceID, httpMethod string) (*Method, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) PutIntegration(_ context.Context, restApiID, resourceID, httpMethod string, req *PutIntegrationRequest) (*Integration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) GetIntegration(_ context.Context, restApiID, resourceID, httpMethod string) (*Integration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) CreateDeployment(_ context.Context, restApiID string, req *CreateDeploymentRequest) (*Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) GetDeployment(_ context.Context, restApiID, deploymentID string) (*Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	deployment, exists := data.deployments[deploymentID]
	if !exists {
		return nil, &ServiceError{Code: errDeploymentNotFound, Message: "Invalid deployment identifier specified"}
	}

	return deployment, nil
}

// GetDeployments returns all deployments.
func (s *MemoryStorage) GetDeployments(_ context.Context, restApiID string, limit int32, _ string) ([]*Deployment, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, "", &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) DeleteDeployment(_ context.Context, restApiID, deploymentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	if _, exists := data.deployments[deploymentID]; !exists {
		return &ServiceError{Code: errDeploymentNotFound, Message: "Invalid deployment identifier specified"}
	}

	delete(data.deployments, deploymentID)

	return nil
}

// CreateStage creates a new stage.
func (s *MemoryStorage) CreateStage(_ context.Context, restApiID string, req *CreateStageRequest) (*Stage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
func (s *MemoryStorage) GetStage(_ context.Context, restApiID, stageName string) (*Stage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	stage, exists := data.stages[stageName]
	if !exists {
		return nil, &ServiceError{Code: errStageNotFound, Message: "Invalid stage identifier specified"}
	}

	return stage, nil
}

// GetStages returns all stages.
func (s *MemoryStorage) GetStages(_ context.Context, restApiID string) ([]*Stage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return nil, &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
	}

	var stages []*Stage

	for _, stage := range data.stages {
		stages = append(stages, stage)
	}

	return stages, nil
}

// DeleteStage deletes a stage.
func (s *MemoryStorage) DeleteStage(_ context.Context, restApiID, stageName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.restApis[restApiID]
	if !exists {
		return &ServiceError{Code: errRestApiNotFound, Message: "Invalid REST API identifier specified"}
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
