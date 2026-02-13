package glue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errEntityNotFound = "EntityNotFoundException"
	errAlreadyExists  = "AlreadyExistsException"
	errInvalidInput   = "InvalidInputException"
)

const defaultCatalogID = "default"

// Storage defines the interface for Glue storage operations.
type Storage interface {
	CreateDatabase(ctx context.Context, catalogID string, input *DatabaseInput) error
	GetDatabase(ctx context.Context, catalogID, name string) (*Database, error)
	GetDatabases(ctx context.Context, catalogID string, maxResults int32, nextToken string) ([]*Database, string, error)
	DeleteDatabase(ctx context.Context, catalogID, name string) error

	CreateTable(ctx context.Context, catalogID, databaseName string, input *TableInput) error
	GetTable(ctx context.Context, catalogID, databaseName, name string) (*Table, error)
	GetTables(ctx context.Context, catalogID, databaseName string, maxResults int32, nextToken string) ([]*Table, string, error)
	DeleteTable(ctx context.Context, catalogID, databaseName, name string) error

	CreateJob(ctx context.Context, input *CreateJobInput) (*Job, error)
	DeleteJob(ctx context.Context, jobName string) error
	StartJobRun(ctx context.Context, input *StartJobRunInput) (*JobRun, error)
}

// MemoryStorage implements Storage with in-memory data structures.
type MemoryStorage struct {
	mu        sync.RWMutex
	databases map[string]*Database // key: catalogID/databaseName
	tables    map[string]*Table    // key: catalogID/databaseName/tableName
	jobs      map[string]*Job      // key: jobName
	jobRuns   map[string]*JobRun   // key: jobRunID
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		databases: make(map[string]*Database),
		tables:    make(map[string]*Table),
		jobs:      make(map[string]*Job),
		jobRuns:   make(map[string]*JobRun),
	}
}

func databaseKey(catalogID, name string) string {
	if catalogID == "" {
		catalogID = defaultCatalogID
	}

	return catalogID + "/" + name
}

func tableKey(catalogID, databaseName, tableName string) string {
	if catalogID == "" {
		catalogID = defaultCatalogID
	}

	return catalogID + "/" + databaseName + "/" + tableName
}

// CreateDatabase creates a new database.
func (s *MemoryStorage) CreateDatabase(_ context.Context, catalogID string, input *DatabaseInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if input.Name == "" {
		return &Error{
			Code:    errInvalidInput,
			Message: "Database name is required",
		}
	}

	key := databaseKey(catalogID, input.Name)

	if _, exists := s.databases[key]; exists {
		return &Error{
			Code:    errAlreadyExists,
			Message: fmt.Sprintf("Database %s already exists", input.Name),
		}
	}

	db := &Database{
		Name:            input.Name,
		Description:     input.Description,
		LocationURI:     input.LocationURI,
		Parameters:      input.Parameters,
		CreateTime:      time.Now(),
		CatalogID:       catalogID,
		CreateTableMode: input.CreateTableMode,
	}

	s.databases[key] = db

	return nil
}

// GetDatabase retrieves a database.
func (s *MemoryStorage) GetDatabase(_ context.Context, catalogID, name string) (*Database, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := databaseKey(catalogID, name)
	db, exists := s.databases[key]

	if !exists {
		return nil, &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Database %s not found", name),
		}
	}

	return db, nil
}

// GetDatabases lists all databases.
func (s *MemoryStorage) GetDatabases(_ context.Context, catalogID string, maxResults int32, _ string) ([]*Database, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	if catalogID == "" {
		catalogID = defaultCatalogID
	}

	prefix := catalogID + "/"
	databases := make([]*Database, 0)

	for key, db := range s.databases {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			databases = append(databases, db)

			if len(databases) >= int(maxResults) {
				break
			}
		}
	}

	return databases, "", nil
}

// DeleteDatabase deletes a database.
func (s *MemoryStorage) DeleteDatabase(_ context.Context, catalogID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := databaseKey(catalogID, name)

	if _, exists := s.databases[key]; !exists {
		return &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Database %s not found", name),
		}
	}

	delete(s.databases, key)

	return nil
}

// CreateTable creates a new table.
func (s *MemoryStorage) CreateTable(_ context.Context, catalogID, databaseName string, input *TableInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if input.Name == "" {
		return &Error{
			Code:    errInvalidInput,
			Message: "Table name is required",
		}
	}

	// Check if database exists.
	dbKey := databaseKey(catalogID, databaseName)
	if _, exists := s.databases[dbKey]; !exists {
		return &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Database %s not found", databaseName),
		}
	}

	key := tableKey(catalogID, databaseName, input.Name)

	if _, exists := s.tables[key]; exists {
		return &Error{
			Code:    errAlreadyExists,
			Message: fmt.Sprintf("Table %s already exists", input.Name),
		}
	}

	now := time.Now()
	table := &Table{
		Name:              input.Name,
		DatabaseName:      databaseName,
		Description:       input.Description,
		Owner:             input.Owner,
		CreateTime:        now,
		UpdateTime:        now,
		Retention:         input.Retention,
		StorageDescriptor: input.StorageDescriptor,
		PartitionKeys:     input.PartitionKeys,
		ViewOriginalText:  input.ViewOriginalText,
		ViewExpandedText:  input.ViewExpandedText,
		TableType:         input.TableType,
		Parameters:        input.Parameters,
		CatalogID:         catalogID,
	}

	s.tables[key] = table

	return nil
}

// GetTable retrieves a table.
func (s *MemoryStorage) GetTable(_ context.Context, catalogID, databaseName, name string) (*Table, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := tableKey(catalogID, databaseName, name)
	table, exists := s.tables[key]

	if !exists {
		return nil, &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Table %s not found", name),
		}
	}

	return table, nil
}

// GetTables lists tables in a database.
func (s *MemoryStorage) GetTables(_ context.Context, catalogID, databaseName string, maxResults int32, _ string) ([]*Table, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	if catalogID == "" {
		catalogID = defaultCatalogID
	}

	prefix := catalogID + "/" + databaseName + "/"
	tables := make([]*Table, 0)

	for key, table := range s.tables {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			tables = append(tables, table)

			if len(tables) >= int(maxResults) {
				break
			}
		}
	}

	return tables, "", nil
}

// DeleteTable deletes a table.
func (s *MemoryStorage) DeleteTable(_ context.Context, catalogID, databaseName, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := tableKey(catalogID, databaseName, name)

	if _, exists := s.tables[key]; !exists {
		return &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Table %s not found", name),
		}
	}

	delete(s.tables, key)

	return nil
}

// CreateJob creates a new job.
func (s *MemoryStorage) CreateJob(_ context.Context, input *CreateJobInput) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if input.Name == "" {
		return nil, &Error{
			Code:    errInvalidInput,
			Message: "Job name is required",
		}
	}

	if input.Role == "" {
		return nil, &Error{
			Code:    errInvalidInput,
			Message: "Role is required",
		}
	}

	if _, exists := s.jobs[input.Name]; exists {
		return nil, &Error{
			Code:    errAlreadyExists,
			Message: fmt.Sprintf("Job %s already exists", input.Name),
		}
	}

	now := time.Now()
	job := &Job{
		Name:                    input.Name,
		Description:             input.Description,
		Role:                    input.Role,
		Command:                 input.Command,
		DefaultArguments:        input.DefaultArguments,
		NonOverridableArguments: input.NonOverridableArguments,
		MaxRetries:              input.MaxRetries,
		AllocatedCapacity:       input.AllocatedCapacity,
		Timeout:                 input.Timeout,
		MaxCapacity:             input.MaxCapacity,
		WorkerType:              input.WorkerType,
		NumberOfWorkers:         input.NumberOfWorkers,
		GlueVersion:             input.GlueVersion,
		CreatedOn:               now,
		LastModifiedOn:          now,
		ExecutionProperty:       input.ExecutionProperty,
	}

	s.jobs[input.Name] = job

	return job, nil
}

// DeleteJob deletes a job.
func (s *MemoryStorage) DeleteJob(_ context.Context, jobName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[jobName]; !exists {
		return &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Job %s not found", jobName),
		}
	}

	delete(s.jobs, jobName)

	return nil
}

// StartJobRun starts a job run.
func (s *MemoryStorage) StartJobRun(_ context.Context, input *StartJobRunInput) (*JobRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[input.JobName]
	if !exists {
		return nil, &Error{
			Code:    errEntityNotFound,
			Message: fmt.Sprintf("Job %s not found", input.JobName),
		}
	}

	runID := input.JobRunID
	if runID == "" {
		runID = "jr_" + uuid.New().String()
	}

	now := time.Now()
	jobRun := &JobRun{
		ID:                runID,
		Attempt:           0,
		JobName:           input.JobName,
		StartedOn:         now,
		LastModifiedOn:    now,
		JobRunState:       "RUNNING",
		Arguments:         input.Arguments,
		AllocatedCapacity: input.AllocatedCapacity,
		Timeout:           input.Timeout,
		MaxCapacity:       input.MaxCapacity,
		WorkerType:        input.WorkerType,
		NumberOfWorkers:   input.NumberOfWorkers,
		GlueVersion:       job.GlueVersion,
	}

	s.jobRuns[runID] = jobRun

	return jobRun, nil
}
