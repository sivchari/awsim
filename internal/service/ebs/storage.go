package ebs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAccountID = "000000000000"
	defaultBlockSize = 524288 // 512 KiB

	errSnapshotNotFound = "ResourceNotFoundException"
)

// ServiceError represents an EBS service error.
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Storage defines the EBS storage interface.
type Storage interface {
	StartSnapshot(ctx context.Context, req *StartSnapshotRequest) (*Snapshot, error)
	CompleteSnapshot(ctx context.Context, snapshotID string) (*Snapshot, error)
	ListSnapshotBlocks(ctx context.Context, snapshotID string) (*ListSnapshotBlocksResponse, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		snapshots: make(map[string]*Snapshot),
	}
}

// StartSnapshot starts a new snapshot.
func (m *MemoryStorage) StartSnapshot(_ context.Context, req *StartSnapshotRequest) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshotID := "snap-" + uuid.New().String()[:12]
	snapshot := &Snapshot{
		BlockSize:        defaultBlockSize,
		Description:      req.Description,
		KmsKeyArn:        req.KmsKeyArn,
		OwnerID:          defaultAccountID,
		ParentSnapshotID: req.ParentSnapshotID,
		SnapshotID:       snapshotID,
		StartTime:        time.Now().Unix(),
		Status:           "pending",
		Tags:             req.Tags,
		VolumeSize:       req.VolumeSize,
	}

	m.snapshots[snapshotID] = snapshot

	return snapshot, nil
}

// CompleteSnapshot completes a snapshot.
func (m *MemoryStorage) CompleteSnapshot(_ context.Context, snapshotID string) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot, exists := m.snapshots[snapshotID]
	if !exists {
		return nil, &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Snapshot %s does not exist.", snapshotID),
		}
	}

	snapshot.Status = "completed"

	return snapshot, nil
}

// ListSnapshotBlocks lists blocks in a snapshot.
func (m *MemoryStorage) ListSnapshotBlocks(_ context.Context, snapshotID string) (*ListSnapshotBlocksResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot, exists := m.snapshots[snapshotID]
	if !exists {
		return nil, &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Snapshot %s does not exist.", snapshotID),
		}
	}

	return &ListSnapshotBlocksResponse{
		Blocks:     []Block{},
		BlockSize:  snapshot.BlockSize,
		VolumeSize: snapshot.VolumeSize,
	}, nil
}
