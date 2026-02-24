package cloudformation

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for CloudFormation storage operations.
type Storage interface {
	CreateStack(ctx context.Context, req *CreateStackRequest) (*Stack, error)
	DeleteStack(ctx context.Context, stackName string) error
	DescribeStacks(ctx context.Context, stackName string) ([]*Stack, error)
	ListStacks(ctx context.Context, statusFilter []string) ([]*Stack, error)
	UpdateStack(ctx context.Context, req *UpdateStackRequest) (*Stack, error)
	DescribeStackResources(ctx context.Context, stackName, logicalResourceID string) ([]*StackResource, error)
	GetTemplate(ctx context.Context, stackName string) (string, error)
	ValidateTemplate(ctx context.Context, templateBody string) (*TemplateValidationResult, error)
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu     sync.RWMutex
	stacks map[string]*Stack // key: stackName
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		stacks: make(map[string]*Stack),
	}
}

// CreateStack creates a new stack.
func (m *MemoryStorage) CreateStack(_ context.Context, req *CreateStackRequest) (*Stack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.StackName == "" {
		return nil, &Error{Code: "ValidationError", Message: "StackName is required"}
	}

	if _, exists := m.stacks[req.StackName]; exists {
		return nil, &Error{Code: "AlreadyExistsException", Message: "Stack already exists"}
	}

	if req.TemplateBody == "" && req.TemplateURL == "" {
		return nil, &Error{Code: "ValidationError", Message: "Either TemplateBody or TemplateURL must be specified"}
	}

	stackID := generateStackID(req.StackName)
	now := time.Now()

	// Parse template to extract resources.
	resources := parseTemplateResources(req.TemplateBody, stackID, req.StackName)

	stack := &Stack{
		StackID:         stackID,
		StackName:       req.StackName,
		TemplateBody:    req.TemplateBody,
		Parameters:      req.Parameters,
		StackStatus:     StackStatusCreateComplete,
		CreationTime:    now,
		LastUpdatedTime: now,
		Resources:       resources,
	}

	m.stacks[req.StackName] = stack

	return stack, nil
}

// DeleteStack deletes a stack.
func (m *MemoryStorage) DeleteStack(_ context.Context, stackName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stack, exists := m.stacks[stackName]
	if !exists {
		return &Error{Code: "StackNotFoundException", Message: "Stack not found"}
	}

	// Mark stack as deleted.
	stack.StackStatus = StackStatusDeleteComplete
	stack.DeletionTime = time.Now()

	// Keep the stack for ListStacks but prevent DescribeStacks from returning it.
	delete(m.stacks, stackName)

	return nil
}

// DescribeStacks describes stacks.
func (m *MemoryStorage) DescribeStacks(_ context.Context, stackName string) ([]*Stack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stackName != "" {
		stack, exists := m.stacks[stackName]
		if !exists {
			return nil, &Error{Code: "StackNotFoundException", Message: "Stack not found"}
		}

		return []*Stack{stack}, nil
	}

	// Return all stacks.
	result := make([]*Stack, 0, len(m.stacks))
	for _, stack := range m.stacks {
		result = append(result, stack)
	}

	return result, nil
}

// ListStacks lists stacks with optional status filter.
func (m *MemoryStorage) ListStacks(_ context.Context, statusFilter []string) ([]*Stack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Stack, 0, len(m.stacks))

	for _, stack := range m.stacks {
		if len(statusFilter) > 0 {
			if !containsStatus(statusFilter, stack.StackStatus) {
				continue
			}
		}

		result = append(result, stack)
	}

	return result, nil
}

// UpdateStack updates a stack.
func (m *MemoryStorage) UpdateStack(_ context.Context, req *UpdateStackRequest) (*Stack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stack, exists := m.stacks[req.StackName]
	if !exists {
		return nil, &Error{Code: "StackNotFoundException", Message: "Stack not found"}
	}

	if req.TemplateBody == "" {
		return nil, &Error{Code: "ValidationError", Message: "TemplateBody is required for update"}
	}

	now := time.Now()
	stack.TemplateBody = req.TemplateBody
	stack.LastUpdatedTime = now
	stack.StackStatus = StackStatusUpdateComplete

	if req.Parameters != nil {
		stack.Parameters = req.Parameters
	}

	// Re-parse resources from new template.
	stack.Resources = parseTemplateResources(req.TemplateBody, stack.StackID, stack.StackName)

	return stack, nil
}

// DescribeStackResources describes stack resources.
func (m *MemoryStorage) DescribeStackResources(_ context.Context, stackName, logicalResourceID string) ([]*StackResource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stack, exists := m.stacks[stackName]
	if !exists {
		return nil, &Error{Code: "StackNotFoundException", Message: "Stack not found"}
	}

	if logicalResourceID != "" {
		for i := range stack.Resources {
			if stack.Resources[i].LogicalResourceID == logicalResourceID {
				return []*StackResource{&stack.Resources[i]}, nil
			}
		}

		return []*StackResource{}, nil
	}

	result := make([]*StackResource, len(stack.Resources))
	for i := range stack.Resources {
		result[i] = &stack.Resources[i]
	}

	return result, nil
}

// GetTemplate gets the template of a stack.
func (m *MemoryStorage) GetTemplate(_ context.Context, stackName string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stack, exists := m.stacks[stackName]
	if !exists {
		return "", &Error{Code: "StackNotFoundException", Message: "Stack not found"}
	}

	return stack.TemplateBody, nil
}

// ValidateTemplate validates a template.
func (m *MemoryStorage) ValidateTemplate(_ context.Context, templateBody string) (*TemplateValidationResult, error) {
	if templateBody == "" {
		return nil, &Error{Code: "ValidationError", Message: "TemplateBody is required"}
	}

	// Try to parse as JSON.
	var template map[string]any
	if err := json.Unmarshal([]byte(templateBody), &template); err != nil {
		return nil, &Error{Code: "ValidationError", Message: "Template format error: " + err.Error()}
	}

	result := &TemplateValidationResult{
		Parameters:   []TemplateParameter{},
		Capabilities: []string{},
	}

	// Extract description.
	if desc, ok := template["Description"].(string); ok {
		result.Description = desc
	}

	// Extract parameters.
	if params, ok := template["Parameters"].(map[string]any); ok {
		for key, value := range params {
			param := parseTemplateParameter(key, value)
			result.Parameters = append(result.Parameters, param)
		}
	}

	return result, nil
}

// parseTemplateParameter extracts a TemplateParameter from a template parameter definition.
func parseTemplateParameter(key string, value any) TemplateParameter {
	param := TemplateParameter{
		ParameterKey: key,
	}

	paramDef, ok := value.(map[string]any)
	if !ok {
		return param
	}

	if defaultVal, ok := paramDef["Default"].(string); ok {
		param.DefaultValue = defaultVal
	}

	if desc, ok := paramDef["Description"].(string); ok {
		param.Description = desc
	}

	if paramType, ok := paramDef["Type"].(string); ok {
		param.ParameterType = paramType
	}

	if noEcho, ok := paramDef["NoEcho"].(bool); ok {
		param.NoEcho = noEcho
	}

	return param
}

// Helper functions.

func generateStackID(stackName string) string {
	return "arn:aws:cloudformation:us-east-1:123456789012:stack/" + stackName + "/" + uuid.New().String()
}

func parseTemplateResources(templateBody, stackID, stackName string) []StackResource {
	var template map[string]any
	if err := json.Unmarshal([]byte(templateBody), &template); err != nil {
		return []StackResource{}
	}

	resources := []StackResource{}
	now := time.Now()

	if resourcesSection, ok := template["Resources"].(map[string]any); ok {
		for logicalID, resourceDef := range resourcesSection {
			resource := StackResource{
				LogicalResourceID:  logicalID,
				PhysicalResourceID: generatePhysicalResourceID(logicalID),
				ResourceStatus:     ResourceStatusCreateComplete,
				Timestamp:          now,
				StackID:            stackID,
				StackName:          stackName,
			}

			if def, ok := resourceDef.(map[string]any); ok {
				if resourceType, ok := def["Type"].(string); ok {
					resource.ResourceType = resourceType
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources
}

func generatePhysicalResourceID(logicalID string) string {
	return logicalID + "-" + uuid.New().String()[:8]
}

func containsStatus(statusFilter []string, status string) bool {
	return slices.Contains(statusFilter, status)
}
