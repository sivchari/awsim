package codeguruprofiler

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for CodeGuru Profiler storage operations.
type Storage interface {
	CreateProfilingGroup(input *CreateProfilingGroupInput) *ProfilingGroup
	DescribeProfilingGroup(name string) (*ProfilingGroup, error)
	UpdateProfilingGroup(name string, input *UpdateProfilingGroupInput) (*ProfilingGroup, error)
	DeleteProfilingGroup(name string) error
	ListProfilingGroups() []ProfilingGroup
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu     sync.RWMutex
	groups map[string]*ProfilingGroup
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		groups: make(map[string]*ProfilingGroup),
	}
}

// CreateProfilingGroup creates a new profiling group.
func (m *MemoryStorage) CreateProfilingGroup(input *CreateProfilingGroupInput) *ProfilingGroup {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	id := uuid.New().String()

	computePlatform := input.ComputePlatform
	if computePlatform == "" {
		computePlatform = "Default"
	}

	group := &ProfilingGroup{
		AgentOrchestrationConfig: input.AgentOrchestrationConfig,
		Arn:                      fmt.Sprintf("arn:aws:codeguru-profiler:us-east-1:000000000000:profilingGroup/%s/%s", input.ProfilingGroupName, id),
		ComputePlatform:          computePlatform,
		CreatedAt:                now,
		Name:                     input.ProfilingGroupName,
		ProfilingStatus:          &ProfilingStatus{},
		Tags:                     input.Tags,
		UpdatedAt:                now,
	}

	if group.AgentOrchestrationConfig == nil {
		group.AgentOrchestrationConfig = &AgentOrchestrationConfig{
			ProfilingEnabled: true,
		}
	}

	m.groups[input.ProfilingGroupName] = group

	return group
}

// DescribeProfilingGroup returns a profiling group by name.
func (m *MemoryStorage) DescribeProfilingGroup(name string) (*ProfilingGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, ok := m.groups[name]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: profiling group %s not found", name)
	}

	return group, nil
}

// UpdateProfilingGroup updates a profiling group.
func (m *MemoryStorage) UpdateProfilingGroup(name string, input *UpdateProfilingGroupInput) (*ProfilingGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, ok := m.groups[name]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: profiling group %s not found", name)
	}

	if input.AgentOrchestrationConfig != nil {
		group.AgentOrchestrationConfig = input.AgentOrchestrationConfig
	}

	group.UpdatedAt = time.Now().UTC()

	return group, nil
}

// DeleteProfilingGroup deletes a profiling group by name.
func (m *MemoryStorage) DeleteProfilingGroup(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.groups[name]; !ok {
		return fmt.Errorf("ResourceNotFoundException: profiling group %s not found", name)
	}

	delete(m.groups, name)

	return nil
}

// ListProfilingGroups returns all profiling groups.
func (m *MemoryStorage) ListProfilingGroups() []ProfilingGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ProfilingGroup, 0, len(m.groups))
	for _, group := range m.groups {
		result = append(result, *group)
	}

	return result
}
