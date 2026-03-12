package entityresolution

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"
)

// Storage defines the Entity Resolution storage interface.
type Storage interface {
	CreateSchemaMapping(ctx context.Context, req *CreateSchemaMappingRequest) (*SchemaMapping, error)
	GetSchemaMapping(ctx context.Context, schemaName string) (*SchemaMapping, error)
	DeleteSchemaMapping(ctx context.Context, schemaName string) error
	ListSchemaMappings(ctx context.Context) ([]SchemaMappingSummary, error)

	CreateMatchingWorkflow(ctx context.Context, req *CreateMatchingWorkflowRequest) (*MatchingWorkflow, error)
	GetMatchingWorkflow(ctx context.Context, workflowName string) (*MatchingWorkflow, error)
	DeleteMatchingWorkflow(ctx context.Context, workflowName string) error
	ListMatchingWorkflows(ctx context.Context) ([]MatchingWorkflowSummary, error)

	CreateIDMappingWorkflow(ctx context.Context, req *CreateIDMappingWorkflowRequest) (*IDMappingWorkflow, error)
	GetIDMappingWorkflow(ctx context.Context, workflowName string) (*IDMappingWorkflow, error)
	DeleteIDMappingWorkflow(ctx context.Context, workflowName string) error
	ListIDMappingWorkflows(ctx context.Context) ([]IDMappingWorkflowSummary, error)

	ListProviderServices(ctx context.Context) ([]ProviderService, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu                 sync.RWMutex
	schemas            map[string]*SchemaMapping
	matchingWorkflows  map[string]*MatchingWorkflow
	idMappingWorkflows map[string]*IDMappingWorkflow
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		schemas:            make(map[string]*SchemaMapping),
		matchingWorkflows:  make(map[string]*MatchingWorkflow),
		idMappingWorkflows: make(map[string]*IDMappingWorkflow),
	}
}

// CreateSchemaMapping creates a new schema mapping.
func (m *MemoryStorage) CreateSchemaMapping(_ context.Context, req *CreateSchemaMappingRequest) (*SchemaMapping, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schemas[req.SchemaName]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Schema mapping %s already exists", req.SchemaName),
		}
	}

	now := float64(time.Now().Unix())
	schema := &SchemaMapping{
		SchemaName:        req.SchemaName,
		SchemaArn:         fmt.Sprintf("arn:aws:entityresolution:%s:%s:schemamapping/%s", defaultRegion, defaultAccountID, req.SchemaName),
		Description:       req.Description,
		MappedInputFields: req.MappedInputFields,
		CreatedAt:         now,
		UpdatedAt:         now,
		Tags:              req.Tags,
	}

	m.schemas[req.SchemaName] = schema

	return schema, nil
}

// GetSchemaMapping returns a schema mapping.
func (m *MemoryStorage) GetSchemaMapping(_ context.Context, schemaName string) (*SchemaMapping, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schema, exists := m.schemas[schemaName]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Schema mapping %s not found", schemaName),
		}
	}

	return schema, nil
}

// DeleteSchemaMapping deletes a schema mapping.
func (m *MemoryStorage) DeleteSchemaMapping(_ context.Context, schemaName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schemas[schemaName]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Schema mapping %s not found", schemaName),
		}
	}

	delete(m.schemas, schemaName)

	return nil
}

// ListSchemaMappings lists all schema mappings.
func (m *MemoryStorage) ListSchemaMappings(_ context.Context) ([]SchemaMappingSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]SchemaMappingSummary, 0, len(m.schemas))
	for _, s := range m.schemas {
		summaries = append(summaries, SchemaMappingSummary{
			SchemaName: s.SchemaName,
			SchemaArn:  s.SchemaArn,
			CreatedAt:  s.CreatedAt,
			UpdatedAt:  s.UpdatedAt,
		})
	}

	return summaries, nil
}

// CreateMatchingWorkflow creates a new matching workflow.
func (m *MemoryStorage) CreateMatchingWorkflow(_ context.Context, req *CreateMatchingWorkflowRequest) (*MatchingWorkflow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.matchingWorkflows[req.WorkflowName]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Matching workflow %s already exists", req.WorkflowName),
		}
	}

	now := float64(time.Now().Unix())
	workflow := &MatchingWorkflow{
		WorkflowName:         req.WorkflowName,
		WorkflowArn:          fmt.Sprintf("arn:aws:entityresolution:%s:%s:matchingworkflow/%s", defaultRegion, defaultAccountID, req.WorkflowName),
		Description:          req.Description,
		InputSourceConfig:    req.InputSourceConfig,
		OutputSourceConfig:   req.OutputSourceConfig,
		ResolutionTechniques: req.ResolutionTechniques,
		RoleArn:              req.RoleArn,
		CreatedAt:            now,
		UpdatedAt:            now,
		Tags:                 req.Tags,
	}

	m.matchingWorkflows[req.WorkflowName] = workflow

	return workflow, nil
}

// GetMatchingWorkflow returns a matching workflow.
func (m *MemoryStorage) GetMatchingWorkflow(_ context.Context, workflowName string) (*MatchingWorkflow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workflow, exists := m.matchingWorkflows[workflowName]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Matching workflow %s not found", workflowName),
		}
	}

	return workflow, nil
}

// DeleteMatchingWorkflow deletes a matching workflow.
func (m *MemoryStorage) DeleteMatchingWorkflow(_ context.Context, workflowName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.matchingWorkflows[workflowName]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Matching workflow %s not found", workflowName),
		}
	}

	delete(m.matchingWorkflows, workflowName)

	return nil
}

// ListMatchingWorkflows lists all matching workflows.
func (m *MemoryStorage) ListMatchingWorkflows(_ context.Context) ([]MatchingWorkflowSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]MatchingWorkflowSummary, 0, len(m.matchingWorkflows))
	for _, w := range m.matchingWorkflows {
		summaries = append(summaries, MatchingWorkflowSummary{
			WorkflowName: w.WorkflowName,
			WorkflowArn:  w.WorkflowArn,
			CreatedAt:    w.CreatedAt,
			UpdatedAt:    w.UpdatedAt,
		})
	}

	return summaries, nil
}

// CreateIDMappingWorkflow creates a new ID mapping workflow.
func (m *MemoryStorage) CreateIDMappingWorkflow(_ context.Context, req *CreateIDMappingWorkflowRequest) (*IDMappingWorkflow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.idMappingWorkflows[req.WorkflowName]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("ID mapping workflow %s already exists", req.WorkflowName),
		}
	}

	now := float64(time.Now().Unix())
	workflow := &IDMappingWorkflow{
		WorkflowName:        req.WorkflowName,
		WorkflowArn:         fmt.Sprintf("arn:aws:entityresolution:%s:%s:idmappingworkflow/%s", defaultRegion, defaultAccountID, req.WorkflowName),
		Description:         req.Description,
		InputSourceConfig:   req.InputSourceConfig,
		OutputSourceConfig:  req.OutputSourceConfig,
		IDMappingTechniques: req.IDMappingTechniques,
		RoleArn:             req.RoleArn,
		CreatedAt:           now,
		UpdatedAt:           now,
		Tags:                req.Tags,
	}

	m.idMappingWorkflows[req.WorkflowName] = workflow

	return workflow, nil
}

// GetIDMappingWorkflow returns an ID mapping workflow.
func (m *MemoryStorage) GetIDMappingWorkflow(_ context.Context, workflowName string) (*IDMappingWorkflow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workflow, exists := m.idMappingWorkflows[workflowName]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("ID mapping workflow %s not found", workflowName),
		}
	}

	return workflow, nil
}

// DeleteIDMappingWorkflow deletes an ID mapping workflow.
func (m *MemoryStorage) DeleteIDMappingWorkflow(_ context.Context, workflowName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.idMappingWorkflows[workflowName]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("ID mapping workflow %s not found", workflowName),
		}
	}

	delete(m.idMappingWorkflows, workflowName)

	return nil
}

// ListIDMappingWorkflows lists all ID mapping workflows.
func (m *MemoryStorage) ListIDMappingWorkflows(_ context.Context) ([]IDMappingWorkflowSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]IDMappingWorkflowSummary, 0, len(m.idMappingWorkflows))
	for _, w := range m.idMappingWorkflows {
		summaries = append(summaries, IDMappingWorkflowSummary{
			WorkflowName: w.WorkflowName,
			WorkflowArn:  w.WorkflowArn,
			CreatedAt:    w.CreatedAt,
			UpdatedAt:    w.UpdatedAt,
		})
	}

	return summaries, nil
}

// ListProviderServices returns a static list of provider services.
func (m *MemoryStorage) ListProviderServices(_ context.Context) ([]ProviderService, error) {
	return []ProviderService{}, nil
}
