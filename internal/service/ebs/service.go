package ebs

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the AWS EBS direct API service.
type Service struct {
	storage Storage
}

// New creates a new EBS service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "ebs"
}

// RegisterRoutes registers the EBS routes.
func (s *Service) RegisterRoutes(r service.Router) {
	r.Handle("POST", "/snapshots", s.StartSnapshotHandler)
	r.Handle("POST", "/snapshots/completion/{snapshotId}", s.CompleteSnapshotHandler)
	r.Handle("GET", "/snapshots/{snapshotId}/blocks", s.ListSnapshotBlocksHandler)
	r.Handle("PUT", "/snapshots/{snapshotId}/blocks/{blockIndex}", s.PutSnapshotBlockHandler)
	r.Handle("GET", "/snapshots/{snapshotId}/blocks/{blockIndex}", s.GetSnapshotBlockHandler)
	r.Handle("GET", "/snapshots/{secondSnapshotId}/changedblocks", s.ListChangedBlocksHandler)
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
