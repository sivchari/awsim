package ebs

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
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
	PutSnapshotBlock(ctx context.Context, snapshotID string, blockIndex int32, data []byte, checksum string) error
	GetSnapshotBlock(ctx context.Context, snapshotID string, blockIndex int32) ([]byte, string, error)
	ListChangedBlocks(ctx context.Context, firstSnapshotID, secondSnapshotID string) (*ListChangedBlocksResponse, error)
}

// blockData stores the raw data and checksum for a single snapshot block.
type blockData struct {
	data     []byte
	checksum string
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
	blocks    map[string]map[int32]*blockData
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		snapshots: make(map[string]*Snapshot),
		blocks:    make(map[string]map[int32]*blockData),
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

	snapshotBlocks := m.blocks[snapshotID]

	indices := make([]int, 0, len(snapshotBlocks))

	for idx := range snapshotBlocks {
		indices = append(indices, int(idx))
	}

	sort.Ints(indices)

	blocks := make([]Block, 0, len(indices))

	for _, idx := range indices {
		token := base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))
		blocks = append(blocks, Block{
			BlockIndex: int32(idx), //nolint:gosec // idx is always a non-negative block index
			BlockToken: token,
		})
	}

	return &ListSnapshotBlocksResponse{
		Blocks:     blocks,
		BlockSize:  snapshot.BlockSize,
		VolumeSize: snapshot.VolumeSize,
	}, nil
}

// PutSnapshotBlock stores a block in the specified snapshot.
func (m *MemoryStorage) PutSnapshotBlock(_ context.Context, snapshotID string, blockIndex int32, data []byte, checksum string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot, exists := m.snapshots[snapshotID]
	if !exists {
		return &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Snapshot %s does not exist.", snapshotID),
		}
	}

	if snapshot.Status != "pending" {
		return &ServiceError{
			Code:    errValidation,
			Message: fmt.Sprintf("Snapshot %s is not in pending state.", snapshotID),
		}
	}

	if m.blocks[snapshotID] == nil {
		m.blocks[snapshotID] = make(map[int32]*blockData)
	}

	m.blocks[snapshotID][blockIndex] = &blockData{
		data:     data,
		checksum: checksum,
	}

	return nil
}

// GetSnapshotBlock retrieves a block from the specified snapshot.
func (m *MemoryStorage) GetSnapshotBlock(_ context.Context, snapshotID string, blockIndex int32) ([]byte, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.snapshots[snapshotID]
	if !exists {
		return nil, "", &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Snapshot %s does not exist.", snapshotID),
		}
	}

	snapshotBlocks, ok := m.blocks[snapshotID]
	if !ok {
		return nil, "", &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Block index %d does not exist in snapshot %s.", blockIndex, snapshotID),
		}
	}

	block, ok := snapshotBlocks[blockIndex]
	if !ok {
		return nil, "", &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Block index %d does not exist in snapshot %s.", blockIndex, snapshotID),
		}
	}

	return block.data, block.checksum, nil
}

// collectSortedIndices merges block indices from two block maps and returns them sorted.
func collectSortedIndices(first, second map[int32]*blockData) []int {
	indexSet := make(map[int32]struct{})

	for idx := range first {
		indexSet[idx] = struct{}{}
	}

	for idx := range second {
		indexSet[idx] = struct{}{}
	}

	indices := make([]int, 0, len(indexSet))

	for idx := range indexSet {
		indices = append(indices, int(idx))
	}

	sort.Ints(indices)

	return indices
}

// buildChangedBlocks builds a list of ChangedBlock entries by comparing two block maps.
func buildChangedBlocks(indices []int, first, second map[int32]*blockData) []ChangedBlock {
	changedBlocks := make([]ChangedBlock, 0, len(indices))

	for _, idx := range indices {
		i := int32(idx) //nolint:gosec // idx is always a non-negative block index
		_, inFirst := first[i]
		_, inSecond := second[i]

		if !inFirst && !inSecond {
			continue
		}

		cb := ChangedBlock{BlockIndex: i}

		if inFirst {
			cb.FirstBlockToken = base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))
		}

		if inSecond {
			cb.SecondBlockToken = base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))
		}

		changedBlocks = append(changedBlocks, cb)
	}

	return changedBlocks
}

// ListChangedBlocks compares two snapshots and returns changed blocks.
func (m *MemoryStorage) ListChangedBlocks(_ context.Context, firstSnapshotID, secondSnapshotID string) (*ListChangedBlocksResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secondSnapshot, exists := m.snapshots[secondSnapshotID]
	if !exists {
		return nil, &ServiceError{
			Code:    errSnapshotNotFound,
			Message: fmt.Sprintf("Snapshot %s does not exist.", secondSnapshotID),
		}
	}

	var firstBlocks map[int32]*blockData

	if firstSnapshotID != "" {
		if _, ok := m.snapshots[firstSnapshotID]; !ok {
			return nil, &ServiceError{
				Code:    errSnapshotNotFound,
				Message: fmt.Sprintf("Snapshot %s does not exist.", firstSnapshotID),
			}
		}

		firstBlocks = m.blocks[firstSnapshotID]
	}

	secondBlocks := m.blocks[secondSnapshotID]
	indices := collectSortedIndices(firstBlocks, secondBlocks)
	changedBlocks := buildChangedBlocks(indices, firstBlocks, secondBlocks)

	return &ListChangedBlocksResponse{
		BlockSize:     secondSnapshot.BlockSize,
		ChangedBlocks: changedBlocks,
		VolumeSize:    secondSnapshot.VolumeSize,
	}, nil
}
