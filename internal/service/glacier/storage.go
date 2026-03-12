package glacier

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"

	errVaultNotFound = "ResourceNotFoundException"
)

// ServiceError represents a Glacier service error.
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Storage defines the Glacier storage interface.
type Storage interface {
	CreateVault(ctx context.Context, vaultName string) (*Vault, error)
	DescribeVault(ctx context.Context, vaultName string) (*Vault, error)
	DeleteVault(ctx context.Context, vaultName string) error
	ListVaults(ctx context.Context) ([]Vault, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu     sync.RWMutex
	vaults map[string]*Vault
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		vaults: make(map[string]*Vault),
	}
}

// CreateVault creates a new vault.
func (m *MemoryStorage) CreateVault(_ context.Context, vaultName string) (*Vault, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// CreateVault is idempotent - no error if vault already exists
	if v, exists := m.vaults[vaultName]; exists {
		return v, nil
	}

	vault := &Vault{
		CreationDate:     time.Now().UTC().Format(time.RFC3339),
		NumberOfArchives: 0,
		SizeInBytes:      0,
		VaultARN:         fmt.Sprintf("arn:aws:glacier:%s:%s:vaults/%s", defaultRegion, defaultAccountID, vaultName),
		VaultName:        vaultName,
	}

	m.vaults[vaultName] = vault

	return vault, nil
}

// DescribeVault returns a vault.
func (m *MemoryStorage) DescribeVault(_ context.Context, vaultName string) (*Vault, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vault, exists := m.vaults[vaultName]
	if !exists {
		return nil, &ServiceError{
			Code:    errVaultNotFound,
			Message: fmt.Sprintf("Vault not found: arn:aws:glacier:%s:%s:vaults/%s", defaultRegion, defaultAccountID, vaultName),
		}
	}

	return vault, nil
}

// DeleteVault deletes a vault.
func (m *MemoryStorage) DeleteVault(_ context.Context, vaultName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.vaults[vaultName]; !exists {
		return &ServiceError{
			Code:    errVaultNotFound,
			Message: fmt.Sprintf("Vault not found: arn:aws:glacier:%s:%s:vaults/%s", defaultRegion, defaultAccountID, vaultName),
		}
	}

	delete(m.vaults, vaultName)

	return nil
}

// ListVaults returns all vaults.
func (m *MemoryStorage) ListVaults(_ context.Context) ([]Vault, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vaults := make([]Vault, 0, len(m.vaults))
	for _, v := range m.vaults {
		vaults = append(vaults, *v)
	}

	return vaults, nil
}
