package dataexchange

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for Data Exchange storage operations.
type Storage interface {
	CreateDataSet(input *CreateDataSetInput) *DataSet
	GetDataSet(id string) (*DataSet, error)
	ListDataSets() []DataSet
	UpdateDataSet(id string, input *UpdateDataSetInput) (*DataSet, error)
	DeleteDataSet(id string) error

	CreateRevision(dataSetID string, input *CreateRevisionInput) (*Revision, error)
	GetRevision(dataSetID, revisionID string) (*Revision, error)
	ListRevisions(dataSetID string) ([]Revision, error)
	UpdateRevision(dataSetID, revisionID string, input *UpdateRevisionInput) (*Revision, error)
	DeleteRevision(dataSetID, revisionID string) error

	CreateJob(input *CreateJobInput) *Job
	GetJob(id string) (*Job, error)
	ListJobs() []Job
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu        sync.RWMutex
	dataSets  map[string]*DataSet
	revisions map[string]map[string]*Revision // dataSetID -> revisionID -> revision
	jobs      map[string]*Job
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		dataSets:  make(map[string]*DataSet),
		revisions: make(map[string]map[string]*Revision),
		jobs:      make(map[string]*Job),
	}
}

// CreateDataSet creates a new data set.
func (m *MemoryStorage) CreateDataSet(input *CreateDataSetInput) *DataSet {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := time.Now().UTC()

	ds := &DataSet{
		Arn:         fmt.Sprintf("arn:aws:dataexchange:us-east-1:000000000000:data-sets/%s", id),
		AssetType:   input.AssetType,
		CreatedAt:   now,
		Description: input.Description,
		ID:          id,
		Name:        input.Name,
		Origin:      "OWNED",
		UpdatedAt:   now,
		Tags:        input.Tags,
	}

	m.dataSets[id] = ds

	return ds
}

// GetDataSet returns a data set by ID.
func (m *MemoryStorage) GetDataSet(id string) (*DataSet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ds, ok := m.dataSets[id]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: data set %s not found", id)
	}

	return ds, nil
}

// ListDataSets returns all data sets.
func (m *MemoryStorage) ListDataSets() []DataSet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]DataSet, 0, len(m.dataSets))
	for _, ds := range m.dataSets {
		result = append(result, *ds)
	}

	return result
}

// UpdateDataSet updates a data set.
func (m *MemoryStorage) UpdateDataSet(id string, input *UpdateDataSetInput) (*DataSet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ds, ok := m.dataSets[id]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: data set %s not found", id)
	}

	if input.Name != "" {
		ds.Name = input.Name
	}

	if input.Description != "" {
		ds.Description = input.Description
	}

	ds.UpdatedAt = time.Now().UTC()

	return ds, nil
}

// DeleteDataSet deletes a data set by ID.
func (m *MemoryStorage) DeleteDataSet(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.dataSets[id]; !ok {
		return fmt.Errorf("ResourceNotFoundException: data set %s not found", id)
	}

	delete(m.dataSets, id)
	delete(m.revisions, id)

	return nil
}

// CreateRevision creates a new revision for a data set.
func (m *MemoryStorage) CreateRevision(dataSetID string, input *CreateRevisionInput) (*Revision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.dataSets[dataSetID]; !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: data set %s not found", dataSetID)
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	rev := &Revision{
		Arn:       fmt.Sprintf("arn:aws:dataexchange:us-east-1:000000000000:data-sets/%s/revisions/%s", dataSetID, id),
		Comment:   input.Comment,
		CreatedAt: now,
		DataSetID: dataSetID,
		ID:        id,
		UpdatedAt: now,
	}

	if m.revisions[dataSetID] == nil {
		m.revisions[dataSetID] = make(map[string]*Revision)
	}

	m.revisions[dataSetID][id] = rev

	return rev, nil
}

// GetRevision returns a revision by data set ID and revision ID.
func (m *MemoryStorage) GetRevision(dataSetID, revisionID string) (*Revision, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dsRevisions, ok := m.revisions[dataSetID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: revision %s not found", revisionID)
	}

	rev, ok := dsRevisions[revisionID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: revision %s not found", revisionID)
	}

	return rev, nil
}

// ListRevisions returns all revisions for a data set.
func (m *MemoryStorage) ListRevisions(dataSetID string) ([]Revision, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.dataSets[dataSetID]; !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: data set %s not found", dataSetID)
	}

	dsRevisions := m.revisions[dataSetID]
	result := make([]Revision, 0, len(dsRevisions))

	for _, rev := range dsRevisions {
		result = append(result, *rev)
	}

	return result, nil
}

// UpdateRevision updates a revision.
func (m *MemoryStorage) UpdateRevision(dataSetID, revisionID string, input *UpdateRevisionInput) (*Revision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dsRevisions, ok := m.revisions[dataSetID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: revision %s not found", revisionID)
	}

	rev, ok := dsRevisions[revisionID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: revision %s not found", revisionID)
	}

	if input.Comment != "" {
		rev.Comment = input.Comment
	}

	if input.Finalized != nil {
		rev.Finalized = *input.Finalized
	}

	rev.UpdatedAt = time.Now().UTC()

	return rev, nil
}

// DeleteRevision deletes a revision.
func (m *MemoryStorage) DeleteRevision(dataSetID, revisionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dsRevisions, ok := m.revisions[dataSetID]
	if !ok {
		return fmt.Errorf("ResourceNotFoundException: revision %s not found", revisionID)
	}

	if _, ok := dsRevisions[revisionID]; !ok {
		return fmt.Errorf("ResourceNotFoundException: revision %s not found", revisionID)
	}

	delete(dsRevisions, revisionID)

	return nil
}

// CreateJob creates a new job.
func (m *MemoryStorage) CreateJob(input *CreateJobInput) *Job {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := time.Now().UTC()

	job := &Job{
		Arn:       fmt.Sprintf("arn:aws:dataexchange:us-east-1:000000000000:jobs/%s", id),
		CreatedAt: now,
		ID:        id,
		State:     "WAITING",
		Type:      input.Type,
		UpdatedAt: now,
	}

	m.jobs[id] = job

	return job
}

// GetJob returns a job by ID.
func (m *MemoryStorage) GetJob(id string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[id]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: job %s not found", id)
	}

	return job, nil
}

// ListJobs returns all jobs.
func (m *MemoryStorage) ListJobs() []Job {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		result = append(result, *job)
	}

	return result
}
