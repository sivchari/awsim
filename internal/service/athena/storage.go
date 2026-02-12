package athena

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes for Athena.
const (
	errInvalidRequestException = "InvalidRequestException"
)

// Storage defines the interface for Athena storage.
type Storage interface {
	StartQueryExecution(ctx context.Context, query string, workGroup string, context *QueryExecutionContext, resultConfig *ResultConfiguration, executionParams []string) (*QueryExecution, error)
	StopQueryExecution(ctx context.Context, queryExecutionID string) error
	GetQueryExecution(ctx context.Context, queryExecutionID string) (*QueryExecution, error)
	GetQueryResults(ctx context.Context, queryExecutionID string, nextToken string, maxResults int32) (*ResultSet, string, error)
	ListQueryExecutions(ctx context.Context, workGroup string, nextToken string, maxResults int32) ([]string, string, error)
	CreateWorkGroup(ctx context.Context, name string, configuration *WorkGroupConfiguration, description string, tags []Tag) error
	DeleteWorkGroup(ctx context.Context, name string, recursiveDelete bool) error
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu              sync.RWMutex
	queryExecutions map[string]*QueryExecution
	workGroups      map[string]*WorkGroup
	queryResults    map[string]*ResultSet
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	s := &MemoryStorage{
		queryExecutions: make(map[string]*QueryExecution),
		workGroups:      make(map[string]*WorkGroup),
		queryResults:    make(map[string]*ResultSet),
	}

	// Create the default "primary" workgroup.
	s.workGroups["primary"] = &WorkGroup{
		Name:         "primary",
		State:        WorkGroupStateEnabled,
		CreationTime: time.Now(),
	}

	return s
}

// StartQueryExecution starts a new query execution.
func (s *MemoryStorage) StartQueryExecution(_ context.Context, query string, workGroup string, execContext *QueryExecutionContext, resultConfig *ResultConfiguration, executionParams []string) (*QueryExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workGroup == "" {
		workGroup = "primary"
	}

	// Verify workgroup exists.
	if _, ok := s.workGroups[workGroup]; !ok {
		return nil, &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: fmt.Sprintf("WorkGroup %s is not found.", workGroup),
		}
	}

	queryExecutionID := uuid.New().String()
	now := time.Now()

	qe := &QueryExecution{
		QueryExecutionID:      queryExecutionID,
		Query:                 query,
		StatementType:         "DML",
		ResultConfiguration:   resultConfig,
		QueryExecutionContext: execContext,
		Status: &QueryExecutionStatus{
			State:              QueryExecutionStateSucceeded,
			SubmissionDateTime: now,
			CompletionDateTime: &now,
		},
		Statistics: &QueryExecutionStatistics{
			EngineExecutionTimeInMillis:      100,
			DataScannedInBytes:               1024,
			TotalExecutionTimeInMillis:       150,
			QueryQueueTimeInMillis:           10,
			ServicePreProcessingTimeInMillis: 20,
			QueryPlanningTimeInMillis:        10,
			ServiceProcessingTimeInMillis:    10,
		},
		WorkGroup:           workGroup,
		ExecutionParameters: executionParams,
		EngineVersion: &EngineVersion{
			SelectedEngineVersion:  "AUTO",
			EffectiveEngineVersion: "Athena engine version 3",
		},
	}

	s.queryExecutions[queryExecutionID] = qe

	// Create mock result set.
	s.queryResults[queryExecutionID] = &ResultSet{
		Rows: []Row{
			{Data: []Datum{{VarCharValue: "column1"}, {VarCharValue: "column2"}}},
			{Data: []Datum{{VarCharValue: "value1"}, {VarCharValue: "value2"}}},
		},
		ResultSetMetadata: &ResultSetMetadata{
			ColumnInfo: []ColumnInfo{
				{Name: "column1", Type: "varchar", Nullable: "UNKNOWN"},
				{Name: "column2", Type: "varchar", Nullable: "UNKNOWN"},
			},
		},
	}

	return qe, nil
}

// StopQueryExecution stops a running query execution.
func (s *MemoryStorage) StopQueryExecution(_ context.Context, queryExecutionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	qe, ok := s.queryExecutions[queryExecutionID]
	if !ok {
		return &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: fmt.Sprintf("QueryExecution %s is not found.", queryExecutionID),
		}
	}

	// Only running or queued queries can be stopped.
	if qe.Status.State == QueryExecutionStateRunning || qe.Status.State == QueryExecutionStateQueued {
		now := time.Now()
		qe.Status.State = QueryExecutionStateCancelled
		qe.Status.StateChangeReason = "Query was cancelled by user."
		qe.Status.CompletionDateTime = &now
	}

	return nil
}

// GetQueryExecution retrieves a query execution by ID.
func (s *MemoryStorage) GetQueryExecution(_ context.Context, queryExecutionID string) (*QueryExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	qe, ok := s.queryExecutions[queryExecutionID]
	if !ok {
		return nil, &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: fmt.Sprintf("QueryExecution %s is not found.", queryExecutionID),
		}
	}

	return qe, nil
}

// GetQueryResults retrieves results for a query execution.
func (s *MemoryStorage) GetQueryResults(_ context.Context, queryExecutionID string, _ string, _ int32) (*ResultSet, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	qe, ok := s.queryExecutions[queryExecutionID]
	if !ok {
		return nil, "", &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: fmt.Sprintf("QueryExecution %s is not found.", queryExecutionID),
		}
	}

	if qe.Status.State != QueryExecutionStateSucceeded {
		return nil, "", &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: "Query has not yet finished. Current state: " + string(qe.Status.State),
		}
	}

	rs, ok := s.queryResults[queryExecutionID]
	if !ok {
		// Return empty result set.
		return &ResultSet{
			Rows:              []Row{},
			ResultSetMetadata: &ResultSetMetadata{ColumnInfo: []ColumnInfo{}},
		}, "", nil
	}

	return rs, "", nil
}

// ListQueryExecutions lists query execution IDs.
func (s *MemoryStorage) ListQueryExecutions(_ context.Context, workGroup string, _ string, maxResults int32) ([]string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 50
	}

	ids := make([]string, 0)

	for id, qe := range s.queryExecutions {
		if workGroup != "" && qe.WorkGroup != workGroup {
			continue
		}

		ids = append(ids, id)

		if int32(len(ids)) >= maxResults {
			break
		}
	}

	return ids, "", nil
}

// CreateWorkGroup creates a new workgroup.
func (s *MemoryStorage) CreateWorkGroup(_ context.Context, name string, configuration *WorkGroupConfiguration, description string, _ []Tag) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.workGroups[name]; ok {
		return &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: fmt.Sprintf("WorkGroup %s already exists.", name),
		}
	}

	wg := &WorkGroup{
		Name:          name,
		State:         WorkGroupStateEnabled,
		Configuration: configuration,
		Description:   description,
		CreationTime:  time.Now(),
	}

	s.workGroups[name] = wg

	return nil
}

// DeleteWorkGroup deletes a workgroup.
func (s *MemoryStorage) DeleteWorkGroup(_ context.Context, name string, recursiveDelete bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == "primary" {
		return &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: "Cannot delete the primary workgroup.",
		}
	}

	if _, ok := s.workGroups[name]; !ok {
		return &AthenaServiceError{
			Code:    errInvalidRequestException,
			Message: fmt.Sprintf("WorkGroup %s is not found.", name),
		}
	}

	// Check if there are any query executions in this workgroup.
	if !recursiveDelete {
		for _, qe := range s.queryExecutions {
			if qe.WorkGroup == name {
				return &AthenaServiceError{
					Code:    errInvalidRequestException,
					Message: fmt.Sprintf("WorkGroup %s has query executions. Set RecursiveDeleteOption to true to delete.", name),
				}
			}
		}
	}

	// Delete query executions if recursive delete.
	if recursiveDelete {
		for id, qe := range s.queryExecutions {
			if qe.WorkGroup == name {
				delete(s.queryExecutions, id)
				delete(s.queryResults, id)
			}
		}
	}

	delete(s.workGroups, name)

	return nil
}
