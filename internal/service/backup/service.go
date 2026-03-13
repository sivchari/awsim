package backup

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the AWS Backup service.
type Service struct {
	storage Storage
}

// New creates a new Backup service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "backup"
}

// Prefix returns the URL prefix for Backup.
func (s *Service) Prefix() string {
	return "/backup-vaults"
}

// RegisterRoutes registers routes with the router.
func (s *Service) RegisterRoutes(r service.Router) {
	// Backup vault operations.
	r.Handle("PUT", "/backup-vaults/{backupVaultName}", s.CreateBackupVault)
	r.Handle("GET", "/backup-vaults/{backupVaultName}", s.DescribeBackupVault)
	r.Handle("GET", "/backup-vaults", s.ListBackupVaults)
	r.Handle("DELETE", "/backup-vaults/{backupVaultName}", s.DeleteBackupVault)

	// Backup plan operations.
	r.Handle("PUT", "/backup/plans", s.CreateBackupPlan)
	r.Handle("GET", "/backup/plans/{backupPlanId}", s.GetBackupPlan)
	r.Handle("GET", "/backup/plans", s.ListBackupPlans)
	r.Handle("DELETE", "/backup/plans/{backupPlanId}", s.DeleteBackupPlan)

	// Backup selection operations.
	r.Handle("PUT", "/backup/plans/{backupPlanId}/selections", s.CreateBackupSelection)
	r.Handle("GET", "/backup/plans/{backupPlanId}/selections/{selectionId}", s.GetBackupSelection)
	r.Handle("GET", "/backup/plans/{backupPlanId}/selections", s.ListBackupSelections)
	r.Handle("DELETE", "/backup/plans/{backupPlanId}/selections/{selectionId}", s.DeleteBackupSelection)
}
