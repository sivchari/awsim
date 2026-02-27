package finspace

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	errResourceNotFound = "ResourceNotFoundException"
	errConflict         = "ConflictException"
	errValidation       = "ValidationException"

	statusCreated = "CREATED"
)

// Storage is the interface for FinSpace storage operations.
type Storage interface {
	// KxEnvironment operations
	CreateKxEnvironment(ctx context.Context, req *CreateKxEnvironmentRequest) (*CreateKxEnvironmentResponse, error)
	GetKxEnvironment(ctx context.Context, environmentID string) (*GetKxEnvironmentResponse, error)
	DeleteKxEnvironment(ctx context.Context, environmentID string) error
	ListKxEnvironments(ctx context.Context, maxResults int, nextToken string) (*ListKxEnvironmentsResponse, error)
	UpdateKxEnvironment(ctx context.Context, req *UpdateKxEnvironmentRequest) (*UpdateKxEnvironmentResponse, error)

	// KxDatabase operations
	CreateKxDatabase(ctx context.Context, req *CreateKxDatabaseRequest) (*CreateKxDatabaseResponse, error)
	GetKxDatabase(ctx context.Context, environmentID, databaseName string) (*GetKxDatabaseResponse, error)
	DeleteKxDatabase(ctx context.Context, environmentID, databaseName string) error
	ListKxDatabases(ctx context.Context, environmentID string, maxResults int, nextToken string) (*ListKxDatabasesResponse, error)
	UpdateKxDatabase(ctx context.Context, req *UpdateKxDatabaseRequest) (*UpdateKxDatabaseResponse, error)

	// KxUser operations
	CreateKxUser(ctx context.Context, req *CreateKxUserRequest) (*CreateKxUserResponse, error)
	GetKxUser(ctx context.Context, environmentID, userName string) (*GetKxUserResponse, error)
	DeleteKxUser(ctx context.Context, environmentID, userName string) error
	ListKxUsers(ctx context.Context, environmentID string, maxResults int, nextToken string) (*ListKxUsersResponse, error)
	UpdateKxUser(ctx context.Context, req *UpdateKxUserRequest) (*UpdateKxUserResponse, error)

	// Tag operations
	TagResource(ctx context.Context, resourceARN string, tags map[string]string) error
	UntagResource(ctx context.Context, resourceARN string, tagKeys []string) error
	ListTagsForResource(ctx context.Context, resourceARN string) (map[string]string, error)
}

// MemoryStorage implements in-memory storage for FinSpace.
type MemoryStorage struct {
	mu           sync.RWMutex
	environments map[string]*KxEnvironment
	databases    map[string]*KxDatabase // key: environmentID/databaseName
	users        map[string]*KxUser     // key: environmentID/userName
	tags         map[string]map[string]string
	accountID    string
	region       string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		environments: make(map[string]*KxEnvironment),
		databases:    make(map[string]*KxDatabase),
		users:        make(map[string]*KxUser),
		tags:         make(map[string]map[string]string),
		accountID:    "123456789012",
		region:       "us-east-1",
	}
}

// CreateKxEnvironment creates a new kdb environment.
func (s *MemoryStorage) CreateKxEnvironment(_ context.Context, req *CreateKxEnvironmentRequest) (*CreateKxEnvironmentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate name
	for _, env := range s.environments {
		if env.Name == req.Name {
			return nil, &Error{
				Code:    errConflict,
				Message: fmt.Sprintf("Environment with name %s already exists", req.Name),
			}
		}
	}

	now := float64(time.Now().Unix())
	environmentID := uuid.New().String()
	arn := fmt.Sprintf("arn:aws:finspace:%s:%s:kxEnvironment/%s", s.region, s.accountID, environmentID)

	env := &KxEnvironment{
		AwsAccountID:      s.accountID,
		CreationTimestamp: now,
		Description:       req.Description,
		EnvironmentARN:    arn,
		EnvironmentID:     environmentID,
		KmsKeyID:          req.KmsKeyID,
		Name:              req.Name,
		Status:            statusCreated,
		UpdateTimestamp:   now,
	}

	s.environments[environmentID] = env

	if len(req.Tags) > 0 {
		s.tags[arn] = req.Tags
	}

	return &CreateKxEnvironmentResponse{
		CreationTimestamp: env.CreationTimestamp,
		Description:       env.Description,
		EnvironmentARN:    env.EnvironmentARN,
		EnvironmentID:     env.EnvironmentID,
		KmsKeyID:          env.KmsKeyID,
		Name:              env.Name,
		Status:            env.Status,
	}, nil
}

// GetKxEnvironment retrieves a kdb environment by ID.
func (s *MemoryStorage) GetKxEnvironment(_ context.Context, environmentID string) (*GetKxEnvironmentResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	env, exists := s.environments[environmentID]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Environment with ID %s not found", environmentID),
		}
	}

	return &GetKxEnvironmentResponse{
		AvailabilityZoneIDs:       env.AvailabilityZoneIDs,
		AwsAccountID:              env.AwsAccountID,
		CertificateARN:            env.CertificateARN,
		CreationTimestamp:         env.CreationTimestamp,
		CustomDNSConfiguration:    env.CustomDNSConfiguration,
		DedicatedServiceAccountID: env.DedicatedServiceAccountID,
		Description:               env.Description,
		DNSStatus:                 env.DNSStatus,
		EnvironmentARN:            env.EnvironmentARN,
		EnvironmentID:             env.EnvironmentID,
		ErrorMessage:              env.ErrorMessage,
		KmsKeyID:                  env.KmsKeyID,
		Name:                      env.Name,
		Status:                    env.Status,
		TgwStatus:                 env.TgwStatus,
		UpdateTimestamp:           env.UpdateTimestamp,
	}, nil
}

// DeleteKxEnvironment deletes a kdb environment.
func (s *MemoryStorage) DeleteKxEnvironment(_ context.Context, environmentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	env, exists := s.environments[environmentID]
	if !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Environment with ID %s not found", environmentID),
		}
	}

	delete(s.tags, env.EnvironmentARN)
	delete(s.environments, environmentID)

	return nil
}

// ListKxEnvironments lists kdb environments.
func (s *MemoryStorage) ListKxEnvironments(_ context.Context, maxResults int, _ string) (*ListKxEnvironmentsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	environments := make([]*KxEnvironment, 0, len(s.environments))

	for _, env := range s.environments {
		environments = append(environments, env)

		if len(environments) >= maxResults {
			break
		}
	}

	return &ListKxEnvironmentsResponse{
		Environments: environments,
	}, nil
}

// UpdateKxEnvironment updates a kdb environment.
func (s *MemoryStorage) UpdateKxEnvironment(_ context.Context, req *UpdateKxEnvironmentRequest) (*UpdateKxEnvironmentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	env, exists := s.environments[req.EnvironmentID]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Environment with ID %s not found", req.EnvironmentID),
		}
	}

	if req.Name != "" {
		env.Name = req.Name
	}

	if req.Description != "" {
		env.Description = req.Description
	}

	env.UpdateTimestamp = float64(time.Now().Unix())

	return &UpdateKxEnvironmentResponse{
		AvailabilityZoneIDs:       env.AvailabilityZoneIDs,
		AwsAccountID:              env.AwsAccountID,
		CreationTimestamp:         env.CreationTimestamp,
		DedicatedServiceAccountID: env.DedicatedServiceAccountID,
		Description:               env.Description,
		DNSStatus:                 env.DNSStatus,
		EnvironmentARN:            env.EnvironmentARN,
		EnvironmentID:             env.EnvironmentID,
		KmsKeyID:                  env.KmsKeyID,
		Name:                      env.Name,
		Status:                    env.Status,
		TgwStatus:                 env.TgwStatus,
		UpdateTimestamp:           env.UpdateTimestamp,
	}, nil
}

// CreateKxDatabase creates a new kdb database.
func (s *MemoryStorage) CreateKxDatabase(_ context.Context, req *CreateKxDatabaseRequest) (*CreateKxDatabaseResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if environment exists
	if _, exists := s.environments[req.EnvironmentID]; !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Environment with ID %s not found", req.EnvironmentID),
		}
	}

	key := fmt.Sprintf("%s/%s", req.EnvironmentID, req.DatabaseName)

	// Check for duplicate
	if _, exists := s.databases[key]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("Database with name %s already exists in environment %s", req.DatabaseName, req.EnvironmentID),
		}
	}

	now := float64(time.Now().Unix())
	arn := fmt.Sprintf("arn:aws:finspace:%s:%s:kxEnvironment/%s/kxDatabase/%s", s.region, s.accountID, req.EnvironmentID, req.DatabaseName)

	db := &KxDatabase{
		CreatedTimestamp:      now,
		DatabaseARN:           arn,
		DatabaseName:          req.DatabaseName,
		Description:           req.Description,
		EnvironmentID:         req.EnvironmentID,
		LastModifiedTimestamp: now,
	}

	s.databases[key] = db

	if len(req.Tags) > 0 {
		s.tags[arn] = req.Tags
	}

	return &CreateKxDatabaseResponse{
		CreatedTimestamp: db.CreatedTimestamp,
		DatabaseARN:      db.DatabaseARN,
		DatabaseName:     db.DatabaseName,
		Description:      db.Description,
		EnvironmentID:    db.EnvironmentID,
	}, nil
}

// GetKxDatabase retrieves a kdb database.
func (s *MemoryStorage) GetKxDatabase(_ context.Context, environmentID, databaseName string) (*GetKxDatabaseResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", environmentID, databaseName)

	db, exists := s.databases[key]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Database with name %s not found in environment %s", databaseName, environmentID),
		}
	}

	return &GetKxDatabaseResponse{
		CreatedTimestamp:         db.CreatedTimestamp,
		DatabaseARN:              db.DatabaseARN,
		DatabaseName:             db.DatabaseName,
		Description:              db.Description,
		EnvironmentID:            db.EnvironmentID,
		LastCompletedChangesetID: db.LastCompletedChangesetID,
		LastModifiedTimestamp:    db.LastModifiedTimestamp,
		NumBytes:                 db.NumBytes,
		NumChangesets:            db.NumChangesets,
		NumFiles:                 db.NumFiles,
	}, nil
}

// DeleteKxDatabase deletes a kdb database.
func (s *MemoryStorage) DeleteKxDatabase(_ context.Context, environmentID, databaseName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s/%s", environmentID, databaseName)

	db, exists := s.databases[key]
	if !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Database with name %s not found in environment %s", databaseName, environmentID),
		}
	}

	delete(s.tags, db.DatabaseARN)
	delete(s.databases, key)

	return nil
}

// ListKxDatabases lists kdb databases.
func (s *MemoryStorage) ListKxDatabases(_ context.Context, environmentID string, maxResults int, _ string) (*ListKxDatabasesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	databases := make([]*KxDatabase, 0)

	for key, db := range s.databases {
		if db.EnvironmentID == environmentID {
			databases = append(databases, db)

			if len(databases) >= maxResults {
				break
			}
		}

		_ = key
	}

	return &ListKxDatabasesResponse{
		KxDatabases: databases,
	}, nil
}

// UpdateKxDatabase updates a kdb database.
func (s *MemoryStorage) UpdateKxDatabase(_ context.Context, req *UpdateKxDatabaseRequest) (*UpdateKxDatabaseResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s/%s", req.EnvironmentID, req.DatabaseName)

	db, exists := s.databases[key]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Database with name %s not found in environment %s", req.DatabaseName, req.EnvironmentID),
		}
	}

	if req.Description != "" {
		db.Description = req.Description
	}

	db.LastModifiedTimestamp = float64(time.Now().Unix())

	return &UpdateKxDatabaseResponse{
		DatabaseARN:           db.DatabaseARN,
		DatabaseName:          db.DatabaseName,
		Description:           db.Description,
		EnvironmentID:         db.EnvironmentID,
		LastModifiedTimestamp: db.LastModifiedTimestamp,
	}, nil
}

// CreateKxUser creates a new kdb user.
func (s *MemoryStorage) CreateKxUser(_ context.Context, req *CreateKxUserRequest) (*CreateKxUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if environment exists
	if _, exists := s.environments[req.EnvironmentID]; !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Environment with ID %s not found", req.EnvironmentID),
		}
	}

	key := fmt.Sprintf("%s/%s", req.EnvironmentID, req.UserName)

	// Check for duplicate
	if _, exists := s.users[key]; exists {
		return nil, &Error{
			Code:    errConflict,
			Message: fmt.Sprintf("User with name %s already exists in environment %s", req.UserName, req.EnvironmentID),
		}
	}

	now := float64(time.Now().Unix())
	arn := fmt.Sprintf("arn:aws:finspace:%s:%s:kxEnvironment/%s/kxUser/%s", s.region, s.accountID, req.EnvironmentID, req.UserName)

	user := &KxUser{
		CreateTimestamp: now,
		IamRole:         req.IamRole,
		UpdateTimestamp: now,
		UserARN:         arn,
		UserName:        req.UserName,
	}

	s.users[key] = user

	if len(req.Tags) > 0 {
		s.tags[arn] = req.Tags
	}

	return &CreateKxUserResponse{
		EnvironmentID: req.EnvironmentID,
		IamRole:       user.IamRole,
		UserARN:       user.UserARN,
		UserName:      user.UserName,
	}, nil
}

// GetKxUser retrieves a kdb user.
func (s *MemoryStorage) GetKxUser(_ context.Context, environmentID, userName string) (*GetKxUserResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", environmentID, userName)

	user, exists := s.users[key]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("User with name %s not found in environment %s", userName, environmentID),
		}
	}

	return &GetKxUserResponse{
		EnvironmentID: environmentID,
		IamRole:       user.IamRole,
		UserARN:       user.UserARN,
		UserName:      user.UserName,
	}, nil
}

// DeleteKxUser deletes a kdb user.
func (s *MemoryStorage) DeleteKxUser(_ context.Context, environmentID, userName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s/%s", environmentID, userName)

	user, exists := s.users[key]
	if !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("User with name %s not found in environment %s", userName, environmentID),
		}
	}

	delete(s.tags, user.UserARN)
	delete(s.users, key)

	return nil
}

// ListKxUsers lists kdb users.
func (s *MemoryStorage) ListKxUsers(_ context.Context, environmentID string, maxResults int, _ string) (*ListKxUsersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	users := make([]*KxUser, 0)

	for key, user := range s.users {
		// Check if the key starts with the environmentID
		if len(key) > len(environmentID) && key[:len(environmentID)] == environmentID && key[len(environmentID)] == '/' {
			users = append(users, user)

			if len(users) >= maxResults {
				break
			}
		}
	}

	return &ListKxUsersResponse{
		Users: users,
	}, nil
}

// UpdateKxUser updates a kdb user.
func (s *MemoryStorage) UpdateKxUser(_ context.Context, req *UpdateKxUserRequest) (*UpdateKxUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s/%s", req.EnvironmentID, req.UserName)

	user, exists := s.users[key]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("User with name %s not found in environment %s", req.UserName, req.EnvironmentID),
		}
	}

	if req.IamRole != "" {
		user.IamRole = req.IamRole
	}

	user.UpdateTimestamp = float64(time.Now().Unix())

	return &UpdateKxUserResponse{
		EnvironmentID: req.EnvironmentID,
		IamRole:       user.IamRole,
		UserARN:       user.UserARN,
		UserName:      user.UserName,
	}, nil
}

// TagResource adds tags to a resource.
func (s *MemoryStorage) TagResource(_ context.Context, resourceARN string, tags map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tags[resourceARN] == nil {
		s.tags[resourceARN] = make(map[string]string)
	}

	maps.Copy(s.tags[resourceARN], tags)

	return nil
}

// UntagResource removes tags from a resource.
func (s *MemoryStorage) UntagResource(_ context.Context, resourceARN string, tagKeys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tags[resourceARN] == nil {
		return nil
	}

	for _, key := range tagKeys {
		delete(s.tags[resourceARN], key)
	}

	return nil
}

// ListTagsForResource lists tags for a resource.
func (s *MemoryStorage) ListTagsForResource(_ context.Context, resourceARN string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tags := s.tags[resourceARN]
	if tags == nil {
		tags = make(map[string]string)
	}

	return tags, nil
}
