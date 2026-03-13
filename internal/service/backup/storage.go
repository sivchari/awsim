package backup

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for Backup storage operations.
type Storage interface {
	CreateVault(name string, input *CreateBackupVaultInput) (*Vault, error)
	DescribeVault(name string) (*Vault, error)
	ListVaults() []Vault
	DeleteVault(name string) error

	CreatePlan(input *CreateBackupPlanInput) (*Plan, error)
	GetPlan(planID string) (*Plan, error)
	ListPlans() []PlanListMember
	DeletePlan(planID string) error

	CreateSelection(planID string, input *CreateBackupSelectionInput) (*Selection, error)
	GetSelection(planID, selectionID string) (*Selection, error)
	ListSelections(planID string) []SelectionListMember
	DeleteSelection(planID, selectionID string) error
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu         sync.RWMutex
	vaults     map[string]*Vault
	plans      map[string]*Plan
	selections map[string]map[string]*Selection // planID -> selectionID -> selection
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		vaults:     make(map[string]*Vault),
		plans:      make(map[string]*Plan),
		selections: make(map[string]map[string]*Selection),
	}
}

func epochNow() float64 {
	return float64(time.Now().Unix())
}

// CreateVault creates a new backup vault.
func (m *MemoryStorage) CreateVault(name string, input *CreateBackupVaultInput) (*Vault, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.vaults[name]; exists {
		return nil, fmt.Errorf("AlreadyExistsException: backup vault %s already exists", name)
	}

	vault := &Vault{
		BackupVaultArn:  fmt.Sprintf("arn:aws:backup:us-east-1:000000000000:backup-vault:%s", name),
		BackupVaultName: name,
		CreationDate:    epochNow(),
	}

	if input != nil {
		vault.CreatorRequestID = input.CreatorRequestID
		vault.EncryptionKeyArn = input.EncryptionKeyArn
	}

	m.vaults[name] = vault

	return vault, nil
}

// DescribeVault returns a backup vault by name.
func (m *MemoryStorage) DescribeVault(name string) (*Vault, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vault, ok := m.vaults[name]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: backup vault %s not found", name)
	}

	return vault, nil
}

// ListVaults returns all backup vaults.
func (m *MemoryStorage) ListVaults() []Vault {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vaults := make([]Vault, 0, len(m.vaults))
	for _, v := range m.vaults {
		vaults = append(vaults, *v)
	}

	return vaults
}

// DeleteVault deletes a backup vault by name.
func (m *MemoryStorage) DeleteVault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.vaults[name]; !ok {
		return fmt.Errorf("ResourceNotFoundException: backup vault %s not found", name)
	}

	delete(m.vaults, name)

	return nil
}

// CreatePlan creates a new backup plan.
func (m *MemoryStorage) CreatePlan(input *CreateBackupPlanInput) (*Plan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	planID := uuid.New().String()
	versionID := uuid.New().String()
	now := epochNow()

	rules := make([]Rule, 0, len(input.BackupPlan.Rules))
	for _, r := range input.BackupPlan.Rules {
		rules = append(rules, Rule{
			RuleName:                r.RuleName,
			RuleID:                  uuid.New().String(),
			TargetBackupVaultName:   r.TargetBackupVaultName,
			ScheduleExpression:      r.ScheduleExpression,
			StartWindowMinutes:      r.StartWindowMinutes,
			CompletionWindowMinutes: r.CompletionWindowMinutes,
		})
	}

	plan := &Plan{
		BackupPlanArn: fmt.Sprintf("arn:aws:backup:us-east-1:000000000000:backup-plan:%s", planID),
		BackupPlanID:  planID,
		BackupPlan: &PlanData{
			BackupPlanName: input.BackupPlan.BackupPlanName,
			Rules:          rules,
		},
		CreationDate: now,
		VersionID:    versionID,
	}

	m.plans[planID] = plan

	return plan, nil
}

// GetPlan returns a backup plan by ID.
func (m *MemoryStorage) GetPlan(planID string) (*Plan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plan, ok := m.plans[planID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: backup plan %s not found", planID)
	}

	return plan, nil
}

// ListPlans returns all backup plans.
func (m *MemoryStorage) ListPlans() []PlanListMember {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plans := make([]PlanListMember, 0, len(m.plans))
	for _, p := range m.plans {
		plans = append(plans, PlanListMember{
			BackupPlanArn:  p.BackupPlanArn,
			BackupPlanID:   p.BackupPlanID,
			BackupPlanName: p.BackupPlan.BackupPlanName,
			CreationDate:   p.CreationDate,
			VersionID:      p.VersionID,
		})
	}

	return plans
}

// DeletePlan deletes a backup plan by ID.
func (m *MemoryStorage) DeletePlan(planID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.plans[planID]; !ok {
		return fmt.Errorf("ResourceNotFoundException: backup plan %s not found", planID)
	}

	delete(m.plans, planID)
	delete(m.selections, planID)

	return nil
}

// CreateSelection creates a new backup selection.
func (m *MemoryStorage) CreateSelection(planID string, input *CreateBackupSelectionInput) (*Selection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.plans[planID]; !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: backup plan %s not found", planID)
	}

	selectionID := uuid.New().String()

	selection := &Selection{
		BackupPlanID: planID,
		SelectionID:  selectionID,
		BackupSelection: &SelectionData{
			SelectionName: input.BackupSelection.SelectionName,
			IamRoleArn:    input.BackupSelection.IamRoleArn,
			Resources:     input.BackupSelection.Resources,
		},
		CreationDate: epochNow(),
	}

	if m.selections[planID] == nil {
		m.selections[planID] = make(map[string]*Selection)
	}

	m.selections[planID][selectionID] = selection

	return selection, nil
}

// GetSelection returns a backup selection by plan ID and selection ID.
func (m *MemoryStorage) GetSelection(planID, selectionID string) (*Selection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	planSelections, ok := m.selections[planID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: backup selection %s not found", selectionID)
	}

	selection, ok := planSelections[selectionID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: backup selection %s not found", selectionID)
	}

	return selection, nil
}

// ListSelections returns all backup selections for a plan.
func (m *MemoryStorage) ListSelections(planID string) []SelectionListMember {
	m.mu.RLock()
	defer m.mu.RUnlock()

	planSelections := m.selections[planID]
	selections := make([]SelectionListMember, 0, len(planSelections))

	for _, s := range planSelections {
		selections = append(selections, SelectionListMember{
			BackupPlanID:  s.BackupPlanID,
			CreationDate:  s.CreationDate,
			IamRoleArn:    s.BackupSelection.IamRoleArn,
			SelectionID:   s.SelectionID,
			SelectionName: s.BackupSelection.SelectionName,
		})
	}

	return selections
}

// DeleteSelection deletes a backup selection.
func (m *MemoryStorage) DeleteSelection(planID, selectionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	planSelections, ok := m.selections[planID]
	if !ok {
		return fmt.Errorf("ResourceNotFoundException: backup selection %s not found", selectionID)
	}

	if _, ok := planSelections[selectionID]; !ok {
		return fmt.Errorf("ResourceNotFoundException: backup selection %s not found", selectionID)
	}

	delete(planSelections, selectionID)

	return nil
}
