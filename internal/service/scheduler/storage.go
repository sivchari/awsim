package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Error codes.
const (
	errResourceNotFound    = "ResourceNotFoundException"
	errConflictException   = "ConflictException"
	errValidationException = "ValidationException"
)

// Default values.
const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "123456789012"
	defaultGroupName = "default"
	defaultTimezone  = "UTC"
)

// Storage defines the Scheduler service storage interface.
type Storage interface {
	// Schedule operations
	CreateSchedule(ctx context.Context, name string, req *CreateScheduleRequest) (*Schedule, error)
	GetSchedule(ctx context.Context, name, groupName string) (*Schedule, error)
	UpdateSchedule(ctx context.Context, name string, req *UpdateScheduleRequest) (*Schedule, error)
	DeleteSchedule(ctx context.Context, name, groupName string) error
	ListSchedules(ctx context.Context, groupName string, limit int32) ([]*Schedule, error)

	// Schedule group operations
	CreateScheduleGroup(ctx context.Context, name string, req *CreateScheduleGroupRequest) (*ScheduleGroup, error)
	GetScheduleGroup(ctx context.Context, name string) (*ScheduleGroup, error)
	DeleteScheduleGroup(ctx context.Context, name string) error
	ListScheduleGroups(ctx context.Context, limit int32) ([]*ScheduleGroup, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu             sync.RWMutex
	schedules      map[string]*Schedule      // key: groupName/scheduleName
	scheduleGroups map[string]*ScheduleGroup // key: groupName
	region         string
	accountID      string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	ms := &MemoryStorage{
		schedules:      make(map[string]*Schedule),
		scheduleGroups: make(map[string]*ScheduleGroup),
		region:         defaultRegion,
		accountID:      defaultAccountID,
	}

	// Create default schedule group.
	ms.scheduleGroups[defaultGroupName] = &ScheduleGroup{
		Name:         defaultGroupName,
		ARN:          generateScheduleGroupARN(defaultRegion, defaultAccountID, defaultGroupName),
		State:        ScheduleGroupStateActive,
		CreationDate: time.Now(),
	}

	return ms
}

// CreateSchedule creates a new schedule.
func (m *MemoryStorage) CreateSchedule(_ context.Context, name string, req *CreateScheduleRequest) (*Schedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	groupName := defaultString(req.GroupName, defaultGroupName)
	key := scheduleKey(groupName, name)

	if _, exists := m.schedules[key]; exists {
		return nil, &Error{Code: errConflictException, Message: "Schedule already exists: " + name}
	}

	// Check if group exists.
	if _, exists := m.scheduleGroups[groupName]; !exists {
		return nil, &Error{Code: errResourceNotFound, Message: "ScheduleGroup not found: " + groupName}
	}

	scheduleARN := generateScheduleARN(m.region, m.accountID, groupName, name)
	now := time.Now()

	schedule := &Schedule{
		Name:                       name,
		ARN:                        scheduleARN,
		GroupName:                  groupName,
		Description:                req.Description,
		ScheduleExpression:         req.ScheduleExpression,
		ScheduleExpressionTimezone: defaultString(req.ScheduleExpressionTimezone, defaultTimezone),
		State:                      defaultString(req.State, StateEnabled),
		FlexibleTimeWindow:         req.FlexibleTimeWindow,
		Target:                     req.Target,
		KmsKeyArn:                  req.KmsKeyArn,
		ActionAfterCompletion:      defaultString(req.ActionAfterCompletion, ActionNone),
		CreationDate:               now,
		LastModificationDate:       now,
	}

	if req.StartDate != nil {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err == nil {
			schedule.StartDate = &t
		}
	}

	if req.EndDate != nil {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err == nil {
			schedule.EndDate = &t
		}
	}

	m.schedules[key] = schedule

	return schedule, nil
}

// GetSchedule gets a schedule.
func (m *MemoryStorage) GetSchedule(_ context.Context, name, groupName string) (*Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupName = defaultString(groupName, defaultGroupName)
	key := scheduleKey(groupName, name)

	schedule, exists := m.schedules[key]
	if !exists {
		return nil, &Error{Code: errResourceNotFound, Message: "Schedule not found: " + name}
	}

	return schedule, nil
}

// UpdateSchedule updates a schedule.
func (m *MemoryStorage) UpdateSchedule(_ context.Context, name string, req *UpdateScheduleRequest) (*Schedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	groupName := defaultString(req.GroupName, defaultGroupName)
	key := scheduleKey(groupName, name)

	schedule, exists := m.schedules[key]
	if !exists {
		return nil, &Error{Code: errResourceNotFound, Message: "Schedule not found: " + name}
	}

	// Update fields.
	schedule.Description = req.Description
	schedule.ScheduleExpression = req.ScheduleExpression
	schedule.ScheduleExpressionTimezone = defaultString(req.ScheduleExpressionTimezone, defaultTimezone)
	schedule.State = defaultString(req.State, schedule.State)
	schedule.FlexibleTimeWindow = req.FlexibleTimeWindow
	schedule.Target = req.Target
	schedule.KmsKeyArn = req.KmsKeyArn
	schedule.ActionAfterCompletion = defaultString(req.ActionAfterCompletion, schedule.ActionAfterCompletion)
	schedule.LastModificationDate = time.Now()

	if req.StartDate != nil {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err == nil {
			schedule.StartDate = &t
		}
	} else {
		schedule.StartDate = nil
	}

	if req.EndDate != nil {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err == nil {
			schedule.EndDate = &t
		}
	} else {
		schedule.EndDate = nil
	}

	return schedule, nil
}

// DeleteSchedule deletes a schedule.
func (m *MemoryStorage) DeleteSchedule(_ context.Context, name, groupName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	groupName = defaultString(groupName, defaultGroupName)
	key := scheduleKey(groupName, name)

	if _, exists := m.schedules[key]; !exists {
		return &Error{Code: errResourceNotFound, Message: "Schedule not found: " + name}
	}

	delete(m.schedules, key)

	return nil
}

// ListSchedules lists schedules.
func (m *MemoryStorage) ListSchedules(_ context.Context, groupName string, limit int32) ([]*Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Schedule, 0, len(m.schedules))

	for _, schedule := range m.schedules {
		if groupName != "" && schedule.GroupName != groupName {
			continue
		}

		result = append(result, schedule)

		if limit > 0 && int32(len(result)) >= limit {
			break
		}
	}

	return result, nil
}

// CreateScheduleGroup creates a new schedule group.
func (m *MemoryStorage) CreateScheduleGroup(_ context.Context, name string, _ *CreateScheduleGroupRequest) (*ScheduleGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.scheduleGroups[name]; exists {
		return nil, &Error{Code: errConflictException, Message: "ScheduleGroup already exists: " + name}
	}

	groupARN := generateScheduleGroupARN(m.region, m.accountID, name)
	now := time.Now()

	group := &ScheduleGroup{
		Name:         name,
		ARN:          groupARN,
		State:        ScheduleGroupStateActive,
		CreationDate: now,
	}

	m.scheduleGroups[name] = group

	return group, nil
}

// GetScheduleGroup gets a schedule group.
func (m *MemoryStorage) GetScheduleGroup(_ context.Context, name string) (*ScheduleGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, exists := m.scheduleGroups[name]
	if !exists {
		return nil, &Error{Code: errResourceNotFound, Message: "ScheduleGroup not found: " + name}
	}

	return group, nil
}

// DeleteScheduleGroup deletes a schedule group.
func (m *MemoryStorage) DeleteScheduleGroup(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == defaultGroupName {
		return &Error{Code: errValidationException, Message: "Cannot delete default schedule group"}
	}

	if _, exists := m.scheduleGroups[name]; !exists {
		return &Error{Code: errResourceNotFound, Message: "ScheduleGroup not found: " + name}
	}

	// Check if any schedules use this group.
	for _, schedule := range m.schedules {
		if schedule.GroupName == name {
			return &Error{Code: errConflictException, Message: "ScheduleGroup has schedules: " + name}
		}
	}

	delete(m.scheduleGroups, name)

	return nil
}

// ListScheduleGroups lists schedule groups.
func (m *MemoryStorage) ListScheduleGroups(_ context.Context, limit int32) ([]*ScheduleGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ScheduleGroup, 0, len(m.scheduleGroups))

	for _, group := range m.scheduleGroups {
		result = append(result, group)

		if limit > 0 && int32(len(result)) >= limit {
			break
		}
	}

	return result, nil
}

// Helper functions.

func generateScheduleARN(region, accountID, groupName, scheduleName string) string {
	return fmt.Sprintf("arn:aws:scheduler:%s:%s:schedule/%s/%s", region, accountID, groupName, scheduleName)
}

func generateScheduleGroupARN(region, accountID, groupName string) string {
	return fmt.Sprintf("arn:aws:scheduler:%s:%s:schedule-group/%s", region, accountID, groupName)
}

func scheduleKey(groupName, scheduleName string) string {
	return groupName + "/" + scheduleName
}

func defaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

// Ensure MemoryStorage implements Storage.
var _ Storage = (*MemoryStorage)(nil)
