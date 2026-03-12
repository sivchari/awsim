package amplify

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAccountID = "000000000000"
	defaultRegion    = "us-east-1"
	defaultPlatform  = "WEB"

	errAppNotFound    = "NotFoundException"
	errBranchNotFound = "NotFoundException"
)

// ServiceError represents an Amplify service error.
type ServiceError struct {
	Code    string
	Message string
	Status  int
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Storage defines the Amplify storage interface.
type Storage interface {
	CreateApp(ctx context.Context, input *CreateAppInput) (*App, error)
	GetApp(ctx context.Context, appID string) (*App, error)
	ListApps(ctx context.Context) ([]App, error)
	UpdateApp(ctx context.Context, appID string, input *UpdateAppInput) (*App, error)
	DeleteApp(ctx context.Context, appID string) (*App, error)
	CreateBranch(ctx context.Context, appID string, input *CreateBranchInput) (*Branch, error)
	GetBranch(ctx context.Context, appID, branchName string) (*Branch, error)
	ListBranches(ctx context.Context, appID string) ([]Branch, error)
	DeleteBranch(ctx context.Context, appID, branchName string) (*Branch, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu       sync.RWMutex
	apps     map[string]*App
	branches map[string]map[string]*Branch // appID -> branchName -> Branch
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		apps:     make(map[string]*App),
		branches: make(map[string]map[string]*Branch),
	}
}

// CreateApp creates a new Amplify app.
func (m *MemoryStorage) CreateApp(_ context.Context, input *CreateAppInput) (*App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	appID := uuid.New().String()[:12]
	now := epochNow()

	platform := defaultPlatform
	if input.Platform != "" {
		platform = input.Platform
	}

	app := &App{
		AppArn:                fmt.Sprintf("arn:aws:amplify:%s:%s:apps/%s", defaultRegion, defaultAccountID, appID),
		AppID:                 appID,
		CreateTime:            now,
		DefaultDomain:         fmt.Sprintf("%s.amplifyapp.com", appID),
		Description:           input.Description,
		EnableBasicAuth:       boolValue(input.EnableBasicAuth),
		EnableBranchAutoBuild: boolValue(input.EnableBranchAutoBuild),
		EnvironmentVariables:  ensureMap(input.EnvironmentVariables),
		Name:                  input.Name,
		Platform:              platform,
		Repository:            input.Repository,
		UpdateTime:            now,
		Tags:                  input.Tags,
	}

	m.apps[appID] = app
	m.branches[appID] = make(map[string]*Branch)

	return app, nil
}

// GetApp returns an app by ID.
func (m *MemoryStorage) GetApp(_ context.Context, appID string) (*App, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	app, exists := m.apps[appID]
	if !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	return app, nil
}

// ListApps returns all apps.
func (m *MemoryStorage) ListApps(_ context.Context) ([]App, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	apps := make([]App, 0, len(m.apps))
	for _, app := range m.apps {
		apps = append(apps, *app)
	}

	return apps, nil
}

// UpdateApp updates an existing app.
func (m *MemoryStorage) UpdateApp(_ context.Context, appID string, input *UpdateAppInput) (*App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.apps[appID]
	if !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	if input.Name != "" {
		app.Name = input.Name
	}

	if input.Description != "" {
		app.Description = input.Description
	}

	if input.Platform != "" {
		app.Platform = input.Platform
	}

	if input.EnableBasicAuth != nil {
		app.EnableBasicAuth = *input.EnableBasicAuth
	}

	if input.EnableBranchAutoBuild != nil {
		app.EnableBranchAutoBuild = *input.EnableBranchAutoBuild
	}

	if input.EnvironmentVariables != nil {
		app.EnvironmentVariables = input.EnvironmentVariables
	}

	app.UpdateTime = epochNow()

	return app, nil
}

// DeleteApp deletes an app.
func (m *MemoryStorage) DeleteApp(_ context.Context, appID string) (*App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.apps[appID]
	if !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	delete(m.apps, appID)

	delete(m.branches, appID)

	return app, nil
}

// CreateBranch creates a new branch for an app.
func (m *MemoryStorage) CreateBranch(_ context.Context, appID string, input *CreateBranchInput) (*Branch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.apps[appID]; !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	now := epochNow()

	stage := "NONE"
	if input.Stage != "" {
		stage = input.Stage
	}

	branch := &Branch{
		ActiveJobID:              "",
		BranchArn:                fmt.Sprintf("arn:aws:amplify:%s:%s:apps/%s/branches/%s", defaultRegion, defaultAccountID, appID, input.BranchName),
		BranchName:               input.BranchName,
		CreateTime:               now,
		CustomDomains:            []string{},
		Description:              input.Description,
		DisplayName:              input.BranchName,
		EnableAutoBuild:          boolValue(input.EnableAutoBuild),
		EnableNotification:       boolValue(input.EnableNotification),
		EnablePullRequestPreview: false,
		EnvironmentVariables:     ensureMap(input.EnvironmentVariables),
		Framework:                input.Framework,
		Stage:                    stage,
		TTL:                      "5",
		TotalNumberOfJobs:        "0",
		UpdateTime:               now,
		Tags:                     input.Tags,
	}

	m.branches[appID][input.BranchName] = branch

	return branch, nil
}

// GetBranch returns a branch by name.
func (m *MemoryStorage) GetBranch(_ context.Context, appID, branchName string) (*Branch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	appBranches, exists := m.branches[appID]
	if !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	branch, exists := appBranches[branchName]
	if !exists {
		return nil, &ServiceError{
			Code:    errBranchNotFound,
			Message: fmt.Sprintf("Branch not found for branchName: %s", branchName),
			Status:  404,
		}
	}

	return branch, nil
}

// ListBranches returns all branches for an app.
func (m *MemoryStorage) ListBranches(_ context.Context, appID string) ([]Branch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.apps[appID]; !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	appBranches := m.branches[appID]
	branches := make([]Branch, 0, len(appBranches))

	for _, branch := range appBranches {
		branches = append(branches, *branch)
	}

	return branches, nil
}

// DeleteBranch deletes a branch.
func (m *MemoryStorage) DeleteBranch(_ context.Context, appID, branchName string) (*Branch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	appBranches, exists := m.branches[appID]
	if !exists {
		return nil, &ServiceError{
			Code:    errAppNotFound,
			Message: fmt.Sprintf("App not found for appId: %s", appID),
			Status:  404,
		}
	}

	branch, exists := appBranches[branchName]
	if !exists {
		return nil, &ServiceError{
			Code:    errBranchNotFound,
			Message: fmt.Sprintf("Branch not found for branchName: %s", branchName),
			Status:  404,
		}
	}

	delete(appBranches, branchName)

	return branch, nil
}

// Helper functions.

func boolValue(b *bool) bool {
	if b == nil {
		return false
	}

	return *b
}

func ensureMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}

	return m
}

func epochNow() float64 {
	return float64(time.Now().Unix())
}
