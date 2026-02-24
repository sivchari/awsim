package cloudtrail

import (
	"context"
	"sync"
	"time"
)

// Error codes.
const (
	errTrailNotFound      = "TrailNotFoundException"
	errTrailAlreadyExists = "TrailAlreadyExistsException"
	errValidationError    = "ValidationException"
)

// Default values.
const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "123456789012"
)

// Storage defines the CloudTrail storage interface.
type Storage interface {
	CreateTrail(ctx context.Context, req *CreateTrailRequest) (*Trail, error)
	DeleteTrail(ctx context.Context, name string) error
	GetTrail(ctx context.Context, name string) (*Trail, error)
	DescribeTrails(ctx context.Context, names []string) ([]*Trail, error)
	StartLogging(ctx context.Context, name string) error
	StopLogging(ctx context.Context, name string) error
	LookupEvents(ctx context.Context, req *LookupEventsRequest) ([]*Event, string, error)
	GetTrailStatus(ctx context.Context, name string) (*Trail, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu        sync.RWMutex
	trails    map[string]*Trail
	region    string
	accountID string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		trails:    make(map[string]*Trail),
		region:    defaultRegion,
		accountID: defaultAccountID,
	}
}

// CreateTrail creates a new trail.
func (m *MemoryStorage) CreateTrail(_ context.Context, req *CreateTrailRequest) (*Trail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Name == "" {
		return nil, &Error{Code: errValidationError, Message: "Trail name is required"}
	}

	if req.S3BucketName == "" {
		return nil, &Error{Code: errValidationError, Message: "S3 bucket name is required"}
	}

	if _, exists := m.trails[req.Name]; exists {
		return nil, &Error{Code: errTrailAlreadyExists, Message: "Trail already exists"}
	}

	trail := &Trail{
		Name:                       req.Name,
		TrailARN:                   generateTrailARN(m.region, m.accountID, req.Name),
		S3BucketName:               req.S3BucketName,
		S3KeyPrefix:                req.S3KeyPrefix,
		IncludeGlobalServiceEvents: true,
		IsMultiRegionTrail:         false,
		HomeRegion:                 m.region,
		IsLogging:                  false,
		LogFileValidationEnabled:   false,
		CloudWatchLogsLogGroupArn:  req.CloudWatchLogsLogGroupArn,
		CloudWatchLogsRoleArn:      req.CloudWatchLogsRoleArn,
		KMSKeyID:                   req.KMSKeyID,
		HasCustomEventSelectors:    false,
		HasInsightSelectors:        false,
		IsOrganizationTrail:        false,
		CreationTime:               time.Now(),
	}

	if req.IncludeGlobalServiceEvents != nil {
		trail.IncludeGlobalServiceEvents = *req.IncludeGlobalServiceEvents
	}

	if req.IsMultiRegionTrail != nil {
		trail.IsMultiRegionTrail = *req.IsMultiRegionTrail
	}

	if req.EnableLogFileValidation != nil {
		trail.LogFileValidationEnabled = *req.EnableLogFileValidation
	}

	if req.IsOrganizationTrail != nil {
		trail.IsOrganizationTrail = *req.IsOrganizationTrail
	}

	m.trails[req.Name] = trail

	return trail, nil
}

// DeleteTrail deletes a trail.
func (m *MemoryStorage) DeleteTrail(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.trails[name]; !exists {
		return &Error{Code: errTrailNotFound, Message: "Trail not found"}
	}

	delete(m.trails, name)

	return nil
}

// GetTrail gets a trail by name.
func (m *MemoryStorage) GetTrail(_ context.Context, name string) (*Trail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	trail, exists := m.trails[name]
	if !exists {
		return nil, &Error{Code: errTrailNotFound, Message: "Trail not found"}
	}

	return trail, nil
}

// DescribeTrails describes trails.
func (m *MemoryStorage) DescribeTrails(_ context.Context, names []string) ([]*Trail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(names) == 0 {
		// Return all trails.
		result := make([]*Trail, 0, len(m.trails))
		for _, trail := range m.trails {
			result = append(result, trail)
		}

		return result, nil
	}

	// Return specified trails.
	result := make([]*Trail, 0, len(names))

	for _, name := range names {
		if trail, exists := m.trails[name]; exists {
			result = append(result, trail)
		}
	}

	return result, nil
}

// StartLogging starts logging for a trail.
func (m *MemoryStorage) StartLogging(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	trail, exists := m.trails[name]
	if !exists {
		return &Error{Code: errTrailNotFound, Message: "Trail not found"}
	}

	trail.IsLogging = true

	return nil
}

// StopLogging stops logging for a trail.
func (m *MemoryStorage) StopLogging(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	trail, exists := m.trails[name]
	if !exists {
		return &Error{Code: errTrailNotFound, Message: "Trail not found"}
	}

	trail.IsLogging = false

	return nil
}

// LookupEvents looks up events.
// For MVP, this returns an empty list as we don't capture actual events.
func (m *MemoryStorage) LookupEvents(_ context.Context, _ *LookupEventsRequest) ([]*Event, string, error) {
	// Return empty events list for MVP.
	return []*Event{}, "", nil
}

// GetTrailStatus gets the status of a trail.
func (m *MemoryStorage) GetTrailStatus(_ context.Context, name string) (*Trail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	trail, exists := m.trails[name]
	if !exists {
		return nil, &Error{Code: errTrailNotFound, Message: "Trail not found"}
	}

	return trail, nil
}

// Helper functions.

func generateTrailARN(region, accountID, trailName string) string {
	return "arn:aws:cloudtrail:" + region + ":" + accountID + ":trail/" + trailName
}
