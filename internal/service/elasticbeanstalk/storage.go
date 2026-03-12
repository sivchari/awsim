package elasticbeanstalk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"

	errAppNotFound = "InvalidParameterValue"
	errAppExists   = "InvalidParameterValue"
	errEnvNotFound = "InvalidParameterValue"
)

// ServiceError represents an Elastic Beanstalk service error.
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Storage defines the Elastic Beanstalk storage interface.
type Storage interface {
	CreateApplication(ctx context.Context, req *CreateApplicationInput) (*ApplicationDescription, error)
	DescribeApplications(ctx context.Context, names []string) ([]ApplicationDescription, error)
	UpdateApplication(ctx context.Context, req *UpdateApplicationInput) (*ApplicationDescription, error)
	DeleteApplication(ctx context.Context, name string) error

	CreateEnvironment(ctx context.Context, req *CreateEnvironmentInput) (*EnvironmentDescription, error)
	DescribeEnvironments(ctx context.Context, appName string, envNames []string) ([]EnvironmentDescription, error)
	TerminateEnvironment(ctx context.Context, envName string) (*EnvironmentDescription, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu           sync.RWMutex
	applications map[string]*ApplicationDescription
	environments map[string]*EnvironmentDescription
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		applications: make(map[string]*ApplicationDescription),
		environments: make(map[string]*EnvironmentDescription),
	}
}

// CreateApplication creates a new application.
func (m *MemoryStorage) CreateApplication(_ context.Context, req *CreateApplicationInput) (*ApplicationDescription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.applications[req.ApplicationName]; exists {
		return nil, &ServiceError{
			Code:    errAppExists,
			Message: fmt.Sprintf("Application %s already exists.", req.ApplicationName),
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	app := &ApplicationDescription{
		ApplicationName: req.ApplicationName,
		Description:     req.Description,
		DateCreated:     now,
		DateUpdated:     now,
		ApplicationArn:  fmt.Sprintf("arn:aws:elasticbeanstalk:%s:%s:application/%s", defaultRegion, defaultAccountID, req.ApplicationName),
	}

	m.applications[req.ApplicationName] = app

	return app, nil
}

// DescribeApplications returns applications matching the filter.
func (m *MemoryStorage) DescribeApplications(_ context.Context, names []string) ([]ApplicationDescription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(names) > 0 {
		apps := make([]ApplicationDescription, 0, len(names))

		for _, name := range names {
			if app, exists := m.applications[name]; exists {
				apps = append(apps, *app)
			}
		}

		return apps, nil
	}

	apps := make([]ApplicationDescription, 0, len(m.applications))
	for _, app := range m.applications {
		apps = append(apps, *app)
	}

	return apps, nil
}

// UpdateApplication updates an existing application.
func (m *MemoryStorage) UpdateApplication(_ context.Context, req *UpdateApplicationInput) (*ApplicationDescription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.applications[req.ApplicationName]
	if !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("No Application named '%s' found.", req.ApplicationName),
		}
	}

	if req.Description != "" {
		app.Description = req.Description
	}

	app.DateUpdated = time.Now().UTC().Format(time.RFC3339)

	return app, nil
}

// DeleteApplication deletes an application.
func (m *MemoryStorage) DeleteApplication(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.applications[name]; !exists {
		return &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("No Application named '%s' found.", name),
		}
	}

	delete(m.applications, name)

	return nil
}

// CreateEnvironment creates a new environment.
func (m *MemoryStorage) CreateEnvironment(_ context.Context, req *CreateEnvironmentInput) (*EnvironmentDescription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.environments[req.EnvironmentName]; exists {
		return nil, &ServiceError{
			Code:    errEnvNotFound,
			Message: fmt.Sprintf("Environment %s already exists.", req.EnvironmentName),
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	envID := "e-" + uuid.New().String()[:8]
	env := &EnvironmentDescription{
		ApplicationName:   req.ApplicationName,
		EnvironmentID:     envID,
		EnvironmentName:   req.EnvironmentName,
		Description:       req.Description,
		SolutionStackName: req.SolutionStackName,
		Status:            "Ready",
		Health:            "Green",
		DateCreated:       now,
		DateUpdated:       now,
		EnvironmentArn:    fmt.Sprintf("arn:aws:elasticbeanstalk:%s:%s:environment/%s/%s", defaultRegion, defaultAccountID, req.ApplicationName, req.EnvironmentName),
	}

	m.environments[req.EnvironmentName] = env

	return env, nil
}

// DescribeEnvironments returns environments matching the filter.
func (m *MemoryStorage) DescribeEnvironments(_ context.Context, appName string, envNames []string) ([]EnvironmentDescription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(envNames) > 0 {
		envs := make([]EnvironmentDescription, 0, len(envNames))

		for _, name := range envNames {
			if env, exists := m.environments[name]; exists {
				if appName == "" || env.ApplicationName == appName {
					envs = append(envs, *env)
				}
			}
		}

		return envs, nil
	}

	envs := make([]EnvironmentDescription, 0, len(m.environments))

	for _, env := range m.environments {
		if appName == "" || env.ApplicationName == appName {
			envs = append(envs, *env)
		}
	}

	return envs, nil
}

// TerminateEnvironment terminates an environment.
func (m *MemoryStorage) TerminateEnvironment(_ context.Context, envName string) (*EnvironmentDescription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	env, exists := m.environments[envName]
	if !exists {
		return nil, &ServiceError{
			Code:    errEnvNotFound,
			Message: fmt.Sprintf("No Environment found for EnvironmentName = '%s'.", envName),
		}
	}

	env.Status = "Terminated"

	delete(m.environments, envName)

	return env, nil
}
