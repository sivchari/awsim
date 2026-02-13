package sfn

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errStateMachineDoesNotExist  = "StateMachineDoesNotExist"
	errStateMachineAlreadyExists = "StateMachineAlreadyExists"
	errExecutionDoesNotExist     = "ExecutionDoesNotExist"
	errExecutionAlreadyExists    = "ExecutionAlreadyExists"
	errInvalidArn                = "InvalidArn"
	errInvalidDefinition         = "InvalidDefinition"
)

// Storage defines the Step Functions storage interface.
type Storage interface {
	// State machine operations.
	CreateStateMachine(ctx context.Context, req *CreateStateMachineRequest) (*StateMachine, error)
	DeleteStateMachine(ctx context.Context, arn string) error
	DescribeStateMachine(ctx context.Context, arn string) (*StateMachine, error)
	ListStateMachines(ctx context.Context, maxResults int32, nextToken string) ([]*StateMachine, string, error)

	// Execution operations.
	StartExecution(ctx context.Context, stateMachineArn, name, input, traceHeader string) (*Execution, error)
	StopExecution(ctx context.Context, executionArn, errorCode, cause string) (*Execution, error)
	DescribeExecution(ctx context.Context, executionArn string) (*Execution, error)
	ListExecutions(ctx context.Context, stateMachineArn, statusFilter string, maxResults int32, nextToken string) ([]*Execution, string, error)
	GetExecutionHistory(ctx context.Context, executionArn string, maxResults int32, nextToken string, reverseOrder bool) ([]*HistoryEvent, string, error)

	// DispatchAction dispatches the request to the appropriate handler.
	DispatchAction(action string) bool
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu            sync.RWMutex
	stateMachines map[string]*StateMachine
	executions    map[string]*executionData
	region        string
	accountID     string
	eventCounter  int64
}

// executionData holds execution information and its history.
type executionData struct {
	execution *Execution
	history   []*HistoryEvent
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		stateMachines: make(map[string]*StateMachine),
		executions:    make(map[string]*executionData),
		region:        "us-east-1",
		accountID:     "000000000000",
	}
}

// CreateStateMachine creates a new state machine.
func (s *MemoryStorage) CreateStateMachine(_ context.Context, req *CreateStateMachineRequest) (*StateMachine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	arn := fmt.Sprintf("arn:aws:states:%s:%s:stateMachine:%s", s.region, s.accountID, req.Name)

	if _, exists := s.stateMachines[arn]; exists {
		return nil, &ServiceError{Code: errStateMachineAlreadyExists, Message: "State machine already exists"}
	}

	smType := StateMachineTypeStandard
	if req.Type == "EXPRESS" {
		smType = StateMachineTypeExpress
	}

	now := time.Now()
	sm := &StateMachine{
		StateMachineArn:      arn,
		Name:                 req.Name,
		Definition:           req.Definition,
		RoleArn:              req.RoleArn,
		Type:                 smType,
		Status:               StateMachineStatusActive,
		CreationDate:         now,
		LoggingConfiguration: req.LoggingConfiguration,
		TracingConfiguration: req.TracingConfiguration,
		RevisionID:           uuid.New().String(),
	}

	s.stateMachines[arn] = sm

	return sm, nil
}

// DeleteStateMachine deletes a state machine.
func (s *MemoryStorage) DeleteStateMachine(_ context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stateMachines[arn]; !exists {
		return &ServiceError{Code: errStateMachineDoesNotExist, Message: "State machine does not exist"}
	}

	delete(s.stateMachines, arn)

	return nil
}

// DescribeStateMachine describes a state machine.
func (s *MemoryStorage) DescribeStateMachine(_ context.Context, arn string) (*StateMachine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sm, exists := s.stateMachines[arn]
	if !exists {
		return nil, &ServiceError{Code: errStateMachineDoesNotExist, Message: "State machine does not exist"}
	}

	return sm, nil
}

// ListStateMachines lists all state machines.
func (s *MemoryStorage) ListStateMachines(_ context.Context, maxResults int32, _ string) ([]*StateMachine, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	stateMachines := make([]*StateMachine, 0, len(s.stateMachines))
	for _, sm := range s.stateMachines {
		stateMachines = append(stateMachines, sm)
	}

	// Sort by creation date.
	sort.Slice(stateMachines, func(i, j int) bool {
		return stateMachines[i].CreationDate.Before(stateMachines[j].CreationDate)
	})

	if int32(len(stateMachines)) > maxResults { //nolint:gosec // slice length bounded by maxResults parameter
		stateMachines = stateMachines[:maxResults]
	}

	return stateMachines, "", nil
}

// StartExecution starts a new execution.
func (s *MemoryStorage) StartExecution(_ context.Context, stateMachineArn, name, input, traceHeader string) (*Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sm, exists := s.stateMachines[stateMachineArn]
	if !exists {
		return nil, &ServiceError{Code: errStateMachineDoesNotExist, Message: "State machine does not exist"}
	}

	execName := name
	if execName == "" {
		execName = uuid.New().String()
	}

	executionArn := fmt.Sprintf("arn:aws:states:%s:%s:execution:%s:%s", s.region, s.accountID, sm.Name, execName)

	if _, exists := s.executions[executionArn]; exists {
		return nil, &ServiceError{Code: errExecutionAlreadyExists, Message: "Execution already exists"}
	}

	now := time.Now()
	exec := s.createExecution(executionArn, stateMachineArn, execName, input, traceHeader, now)
	history := s.createExecutionHistory(sm.RoleArn, input, now)

	exec.Status = ExecutionStatusSucceeded
	exec.StopDate = &now
	exec.Output = input
	exec.OutputDetails = &CloudWatchEventsExecutionDataDetails{Included: true}

	s.executions[executionArn] = &executionData{execution: exec, history: history}

	return exec, nil
}

// createExecution creates a new execution object.
func (s *MemoryStorage) createExecution(arn, smArn, name, input, traceHeader string, now time.Time) *Execution {
	return &Execution{
		ExecutionArn:    arn,
		StateMachineArn: smArn,
		Name:            name,
		Status:          ExecutionStatusRunning,
		StartDate:       now,
		Input:           input,
		InputDetails:    &CloudWatchEventsExecutionDataDetails{Included: true},
		TraceHeader:     traceHeader,
	}
}

// createExecutionHistory creates execution history events for a pass-through execution.
func (s *MemoryStorage) createExecutionHistory(roleArn, input string, now time.Time) []*HistoryEvent {
	startID := atomic.AddInt64(&s.eventCounter, 1)
	endID := atomic.AddInt64(&s.eventCounter, 1)

	return []*HistoryEvent{
		{
			Timestamp: now, Type: HistoryEventTypeExecutionStarted, ID: startID, PreviousEventID: 0,
			ExecutionStartedEventDetails: &ExecutionStartedEventDetails{
				Input: input, InputDetails: &CloudWatchEventsExecutionDataDetails{Included: true}, RoleArn: roleArn,
			},
		},
		{
			Timestamp: now, Type: HistoryEventTypeExecutionSucceeded, ID: endID, PreviousEventID: startID,
			ExecutionSucceededEventDetails: &ExecutionSucceededEventDetails{
				Output: input, OutputDetails: &CloudWatchEventsExecutionDataDetails{Included: true},
			},
		},
	}
}

// StopExecution stops an execution.
func (s *MemoryStorage) StopExecution(_ context.Context, executionArn, errorCode, cause string) (*Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ed, exists := s.executions[executionArn]
	if !exists {
		return nil, &ServiceError{Code: errExecutionDoesNotExist, Message: "Execution does not exist"}
	}

	if ed.execution.Status != ExecutionStatusRunning {
		// Already stopped.
		return ed.execution, nil
	}

	now := time.Now()
	ed.execution.Status = ExecutionStatusAborted
	ed.execution.StopDate = &now
	ed.execution.Error = errorCode
	ed.execution.Cause = cause

	// Add abort event.
	eventID := atomic.AddInt64(&s.eventCounter, 1)
	abortEvent := &HistoryEvent{
		Timestamp:       now,
		Type:            HistoryEventTypeExecutionAborted,
		ID:              eventID,
		PreviousEventID: int64(len(ed.history)),
		ExecutionAbortedEventDetails: &ExecutionAbortedEventDetails{
			Error: errorCode,
			Cause: cause,
		},
	}

	ed.history = append(ed.history, abortEvent)

	return ed.execution, nil
}

// DescribeExecution describes an execution.
func (s *MemoryStorage) DescribeExecution(_ context.Context, executionArn string) (*Execution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ed, exists := s.executions[executionArn]
	if !exists {
		return nil, &ServiceError{Code: errExecutionDoesNotExist, Message: "Execution does not exist"}
	}

	return ed.execution, nil
}

// ListExecutions lists executions for a state machine.
func (s *MemoryStorage) ListExecutions(_ context.Context, stateMachineArn, statusFilter string, maxResults int32, _ string) ([]*Execution, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	var executions []*Execution

	for _, ed := range s.executions {
		if ed.execution.StateMachineArn != stateMachineArn {
			continue
		}

		if statusFilter != "" && string(ed.execution.Status) != statusFilter {
			continue
		}

		executions = append(executions, ed.execution)
	}

	// Sort by start date (most recent first).
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartDate.After(executions[j].StartDate)
	})

	if int32(len(executions)) > maxResults { //nolint:gosec // slice length bounded by maxResults parameter
		executions = executions[:maxResults]
	}

	return executions, "", nil
}

// GetExecutionHistory gets the history of an execution.
func (s *MemoryStorage) GetExecutionHistory(_ context.Context, executionArn string, maxResults int32, _ string, reverseOrder bool) ([]*HistoryEvent, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ed, exists := s.executions[executionArn]
	if !exists {
		return nil, "", &ServiceError{Code: errExecutionDoesNotExist, Message: "Execution does not exist"}
	}

	if maxResults <= 0 {
		maxResults = 100
	}

	// Copy events.
	events := make([]*HistoryEvent, len(ed.history))
	copy(events, ed.history)

	if reverseOrder {
		for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
			events[i], events[j] = events[j], events[i]
		}
	}

	if int32(len(events)) > maxResults { //nolint:gosec // slice length bounded by maxResults parameter
		events = events[:maxResults]
	}

	return events, "", nil
}

// DispatchAction checks if the action is valid.
func (s *MemoryStorage) DispatchAction(_ string) bool {
	return true
}
