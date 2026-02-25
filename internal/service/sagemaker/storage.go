package sagemaker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Error codes.
const (
	errResourceNotFound    = "ResourceNotFound"
	errResourceInUse       = "ResourceInUse"
	errValidationException = "ValidationException"
	errInternalFailure     = "InternalFailure"
)

// Default values.
const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "123456789012"
)

// Notebook instance statuses.
const statusInService = "InService"

// Training job statuses.
const trainingStatusCompleted = "Completed"

// Endpoint statuses.
const endpointStatusInService = "InService"

// Storage defines the SageMaker service storage interface.
type Storage interface {
	// Notebook instance operations
	CreateNotebookInstance(ctx context.Context, req *CreateNotebookInstanceRequest) (*NotebookInstance, error)
	DeleteNotebookInstance(ctx context.Context, name string) error
	DescribeNotebookInstance(ctx context.Context, name string) (*NotebookInstance, error)
	ListNotebookInstances(ctx context.Context, maxResults int32, nextToken string) ([]*NotebookInstance, string, error)

	// Training job operations
	CreateTrainingJob(ctx context.Context, req *CreateTrainingJobRequest) (*TrainingJob, error)
	DescribeTrainingJob(ctx context.Context, name string) (*TrainingJob, error)

	// Model operations
	CreateModel(ctx context.Context, req *CreateModelRequest) (*Model, error)
	DeleteModel(ctx context.Context, name string) error

	// Endpoint operations
	CreateEndpoint(ctx context.Context, req *CreateEndpointRequest) (*Endpoint, error)
	DeleteEndpoint(ctx context.Context, name string) error
	DescribeEndpoint(ctx context.Context, name string) (*Endpoint, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu                sync.RWMutex
	notebookInstances map[string]*NotebookInstance
	trainingJobs      map[string]*TrainingJob
	models            map[string]*Model
	endpoints         map[string]*Endpoint
	region            string
	accountID         string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		notebookInstances: make(map[string]*NotebookInstance),
		trainingJobs:      make(map[string]*TrainingJob),
		models:            make(map[string]*Model),
		endpoints:         make(map[string]*Endpoint),
		region:            defaultRegion,
		accountID:         defaultAccountID,
	}
}

// CreateNotebookInstance creates a new notebook instance.
func (m *MemoryStorage) CreateNotebookInstance(_ context.Context, req *CreateNotebookInstanceRequest) (*NotebookInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.notebookInstances[req.NotebookInstanceName]; exists {
		return nil, &Error{
			Code:    errResourceInUse,
			Message: fmt.Sprintf("Notebook instance %s already exists", req.NotebookInstanceName),
		}
	}

	now := time.Now()
	arn := fmt.Sprintf("arn:aws:sagemaker:%s:%s:notebook-instance/%s", m.region, m.accountID, req.NotebookInstanceName)

	instance := &NotebookInstance{
		NotebookInstanceName:   req.NotebookInstanceName,
		NotebookInstanceArn:    arn,
		NotebookInstanceStatus: statusInService,
		InstanceType:           req.InstanceType,
		RoleArn:                req.RoleArn,
		KmsKeyID:               req.KmsKeyID,
		SubnetID:               req.SubnetID,
		SecurityGroups:         req.SecurityGroupIDs,
		DirectInternetAccess:   req.DirectInternetAccess,
		VolumeSizeInGB:         req.VolumeSizeInGB,
		AcceleratorTypes:       req.AcceleratorTypes,
		DefaultCodeRepository:  req.DefaultCodeRepository,
		AdditionalCodeRepos:    req.AdditionalCodeRepos,
		RootAccess:             req.RootAccess,
		PlatformIdentifier:     req.PlatformIdentifier,
		InstanceMetadataConfig: req.InstanceMetadataConfig,
		CreationTime:           now,
		LastModifiedTime:       now,
	}

	// Set default volume size if not specified.
	if instance.VolumeSizeInGB == 0 {
		instance.VolumeSizeInGB = 5
	}

	// Set default direct internet access.
	if instance.DirectInternetAccess == "" {
		instance.DirectInternetAccess = "Enabled"
	}

	// Set default root access.
	if instance.RootAccess == "" {
		instance.RootAccess = "Enabled"
	}

	m.notebookInstances[req.NotebookInstanceName] = instance

	return instance, nil
}

// DeleteNotebookInstance deletes a notebook instance.
func (m *MemoryStorage) DeleteNotebookInstance(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.notebookInstances[name]; !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Notebook instance %s not found", name),
		}
	}

	delete(m.notebookInstances, name)

	return nil
}

// DescribeNotebookInstance describes a notebook instance.
func (m *MemoryStorage) DescribeNotebookInstance(_ context.Context, name string) (*NotebookInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.notebookInstances[name]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Notebook instance %s not found", name),
		}
	}

	return instance, nil
}

// ListNotebookInstances lists notebook instances.
func (m *MemoryStorage) ListNotebookInstances(_ context.Context, maxResults int32, _ string) ([]*NotebookInstance, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	result := make([]*NotebookInstance, 0, len(m.notebookInstances))

	for _, instance := range m.notebookInstances {
		result = append(result, instance)

		if len(result) >= int(maxResults) {
			break
		}
	}

	return result, "", nil
}

// CreateTrainingJob creates a new training job.
func (m *MemoryStorage) CreateTrainingJob(_ context.Context, req *CreateTrainingJobRequest) (*TrainingJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.trainingJobs[req.TrainingJobName]; exists {
		return nil, &Error{
			Code:    errResourceInUse,
			Message: fmt.Sprintf("Training job %s already exists", req.TrainingJobName),
		}
	}

	now := time.Now()
	arn := fmt.Sprintf("arn:aws:sagemaker:%s:%s:training-job/%s", m.region, m.accountID, req.TrainingJobName)

	job := &TrainingJob{
		TrainingJobName:   req.TrainingJobName,
		TrainingJobArn:    arn,
		TrainingJobStatus: trainingStatusCompleted,
		SecondaryStatus:   "Completed",
		AlgorithmSpec:     req.AlgorithmSpec,
		RoleArn:           req.RoleArn,
		InputDataConfig:   req.InputDataConfig,
		OutputDataConfig:  req.OutputDataConfig,
		ResourceConfig:    req.ResourceConfig,
		StoppingCondition: req.StoppingCondition,
		CreationTime:      now,
		TrainingStartTime: &now,
		TrainingEndTime:   &now,
	}

	m.trainingJobs[req.TrainingJobName] = job

	return job, nil
}

// DescribeTrainingJob describes a training job.
func (m *MemoryStorage) DescribeTrainingJob(_ context.Context, name string) (*TrainingJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.trainingJobs[name]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Training job %s not found", name),
		}
	}

	return job, nil
}

// CreateModel creates a new model.
func (m *MemoryStorage) CreateModel(_ context.Context, req *CreateModelRequest) (*Model, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.models[req.ModelName]; exists {
		return nil, &Error{
			Code:    errResourceInUse,
			Message: fmt.Sprintf("Model %s already exists", req.ModelName),
		}
	}

	now := time.Now()
	arn := fmt.Sprintf("arn:aws:sagemaker:%s:%s:model/%s", m.region, m.accountID, req.ModelName)

	model := &Model{
		ModelName:              req.ModelName,
		ModelArn:               arn,
		PrimaryContainer:       req.PrimaryContainer,
		ExecutionRoleArn:       req.ExecutionRoleArn,
		EnableNetworkIsolation: req.EnableNetworkIsolation,
		CreationTime:           now,
	}

	m.models[req.ModelName] = model

	return model, nil
}

// DeleteModel deletes a model.
func (m *MemoryStorage) DeleteModel(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.models[name]; !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Model %s not found", name),
		}
	}

	delete(m.models, name)

	return nil
}

// CreateEndpoint creates a new endpoint.
func (m *MemoryStorage) CreateEndpoint(_ context.Context, req *CreateEndpointRequest) (*Endpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.endpoints[req.EndpointName]; exists {
		return nil, &Error{
			Code:    errResourceInUse,
			Message: fmt.Sprintf("Endpoint %s already exists", req.EndpointName),
		}
	}

	now := time.Now()
	arn := fmt.Sprintf("arn:aws:sagemaker:%s:%s:endpoint/%s", m.region, m.accountID, req.EndpointName)

	endpoint := &Endpoint{
		EndpointName:       req.EndpointName,
		EndpointArn:        arn,
		EndpointConfigName: req.EndpointConfigName,
		EndpointStatus:     endpointStatusInService,
		CreationTime:       now,
		LastModifiedTime:   now,
	}

	m.endpoints[req.EndpointName] = endpoint

	return endpoint, nil
}

// DeleteEndpoint deletes an endpoint.
func (m *MemoryStorage) DeleteEndpoint(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.endpoints[name]; !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Endpoint %s not found", name),
		}
	}

	delete(m.endpoints, name)

	return nil
}

// DescribeEndpoint describes an endpoint.
func (m *MemoryStorage) DescribeEndpoint(_ context.Context, name string) (*Endpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	endpoint, exists := m.endpoints[name]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Endpoint %s not found", name),
		}
	}

	return endpoint, nil
}

// Ensure MemoryStorage implements Storage.
var _ Storage = (*MemoryStorage)(nil)
