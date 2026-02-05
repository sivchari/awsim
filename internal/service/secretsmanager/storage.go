package secretsmanager

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultRegion         = "us-east-1"
	defaultAccountID      = "000000000000"
	defaultRecoveryWindow = 30
	stageCurrent          = "AWSCURRENT"
	stagePrevious         = "AWSPREVIOUS"
)

// Storage defines the Secrets Manager storage interface.
type Storage interface {
	CreateSecret(ctx context.Context, req *CreateSecretRequest) (*Secret, error)
	GetSecretValue(ctx context.Context, secretID, versionID, versionStage string) (*Secret, *SecretVersion, error)
	PutSecretValue(ctx context.Context, secretID, clientToken, secretString string, secretBinary []byte, versionStages []string) (*Secret, *SecretVersion, error)
	DeleteSecret(ctx context.Context, secretID string, recoveryWindow int64, forceDelete bool) (*Secret, error)
	ListSecrets(ctx context.Context, maxResults int, nextToken string, includePlannedDeletion bool) ([]*Secret, string, error)
	DescribeSecret(ctx context.Context, secretID string) (*Secret, error)
	UpdateSecret(ctx context.Context, req *UpdateSecretRequest) (*Secret, *SecretVersion, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	secrets map[string]*Secret // keyed by name
	baseURL string
}

// NewMemoryStorage creates a new in-memory Secrets Manager storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		secrets: make(map[string]*Secret),
		baseURL: baseURL,
	}
}

// CreateSecret creates a new secret.
func (m *MemoryStorage) CreateSecret(_ context.Context, req *CreateSecretRequest) (*Secret, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.secrets[req.Name]; exists {
		return nil, &SecretError{
			Code:    "ResourceExistsException",
			Message: fmt.Sprintf("The operation failed because the secret %s already exists.", req.Name),
		}
	}

	now := time.Now()
	versionID := uuid.New().String()

	secret := &Secret{
		ARN:             m.buildARN(req.Name),
		Name:            req.Name,
		Description:     req.Description,
		KmsKeyID:        req.KmsKeyID,
		VersionID:       versionID,
		SecretString:    req.SecretString,
		SecretBinary:    req.SecretBinary,
		CreatedDate:     now,
		LastChangedDate: now,
		Tags:            req.Tags,
		VersionIDs:      make(map[string]*SecretVersion),
	}

	// Create initial version if secret value is provided.
	if req.SecretString != "" || len(req.SecretBinary) > 0 {
		version := &SecretVersion{
			VersionID:     versionID,
			SecretString:  req.SecretString,
			SecretBinary:  req.SecretBinary,
			VersionStages: []string{stageCurrent},
			CreatedDate:   now,
			KmsKeyID:      req.KmsKeyID,
		}
		secret.VersionIDs[versionID] = version
	}

	m.secrets[req.Name] = secret

	return secret, nil
}

// GetSecretValue retrieves a secret value.
func (m *MemoryStorage) GetSecretValue(_ context.Context, secretID, versionID, versionStage string) (*Secret, *SecretVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secret := m.findSecret(secretID)
	if secret == nil {
		return nil, nil, &SecretError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Secrets Manager can't find the specified secret: %s", secretID),
		}
	}

	if secret.DeletedDate != nil {
		return nil, nil, &SecretError{
			Code:    "InvalidRequestException",
			Message: "You can't perform this operation on a secret that's scheduled for deletion.",
		}
	}

	// Default to AWSCURRENT if no version specified.
	if versionStage == "" && versionID == "" {
		versionStage = stageCurrent
	}

	var version *SecretVersion

	if versionID != "" {
		version = secret.VersionIDs[versionID]
	} else {
		// Find version by stage.
		for _, v := range secret.VersionIDs {
			if slices.Contains(v.VersionStages, versionStage) {
				version = v

				break
			}
		}
	}

	if version == nil {
		return nil, nil, &SecretError{
			Code:    "ResourceNotFoundException",
			Message: "Secrets Manager can't find the specified secret version.",
		}
	}

	// Update last accessed date.
	now := time.Now()
	secret.LastAccessedDate = &now

	return secret, version, nil
}

// PutSecretValue puts a new value into an existing secret.
func (m *MemoryStorage) PutSecretValue(_ context.Context, secretID, clientToken, secretString string, secretBinary []byte, versionStages []string) (*Secret, *SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	secret := m.findSecret(secretID)
	if secret == nil {
		return nil, nil, &SecretError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Secrets Manager can't find the specified secret: %s", secretID),
		}
	}

	if secret.DeletedDate != nil {
		return nil, nil, &SecretError{
			Code:    "InvalidRequestException",
			Message: "You can't perform this operation on a secret that's scheduled for deletion.",
		}
	}

	now := time.Now()
	versionID := clientToken

	if versionID == "" {
		versionID = uuid.New().String()
	}

	// Default stages.
	if len(versionStages) == 0 {
		versionStages = []string{stageCurrent}
	}

	// Remove AWSCURRENT stage from previous version.
	for _, v := range secret.VersionIDs {
		newStages := make([]string, 0)

		for _, stage := range v.VersionStages {
			if stage != stageCurrent {
				newStages = append(newStages, stage)
			} else {
				// Add AWSPREVIOUS to the old current version.
				newStages = append(newStages, stagePrevious)
			}
		}

		v.VersionStages = newStages
	}

	version := &SecretVersion{
		VersionID:     versionID,
		SecretString:  secretString,
		SecretBinary:  secretBinary,
		VersionStages: versionStages,
		CreatedDate:   now,
		KmsKeyID:      secret.KmsKeyID,
	}

	secret.VersionIDs[versionID] = version
	secret.VersionID = versionID
	secret.SecretString = secretString
	secret.SecretBinary = secretBinary
	secret.LastChangedDate = now

	return secret, version, nil
}

// DeleteSecret marks a secret for deletion.
func (m *MemoryStorage) DeleteSecret(_ context.Context, secretID string, recoveryWindow int64, forceDelete bool) (*Secret, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	secret := m.findSecret(secretID)
	if secret == nil {
		return nil, &SecretError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Secrets Manager can't find the specified secret: %s", secretID),
		}
	}

	now := time.Now()

	if forceDelete {
		// Immediately delete.
		delete(m.secrets, secret.Name)
		secret.DeletedDate = &now

		return secret, nil
	}

	if recoveryWindow == 0 {
		recoveryWindow = defaultRecoveryWindow
	}

	deletionDate := now.AddDate(0, 0, int(recoveryWindow))
	secret.DeletedDate = &now
	secret.ScheduledDeletionDate = &deletionDate
	secret.RecoveryWindowInDays = recoveryWindow

	return secret, nil
}

// ListSecrets returns all secrets.
func (m *MemoryStorage) ListSecrets(_ context.Context, maxResults int, nextToken string, includePlannedDeletion bool) ([]*Secret, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	// Collect all secrets.
	allSecrets := make([]*Secret, 0, len(m.secrets))

	for _, secret := range m.secrets {
		// Skip deleted secrets unless requested.
		if secret.DeletedDate != nil && !includePlannedDeletion {
			continue
		}

		allSecrets = append(allSecrets, secret)
	}

	// Sort by name for consistent ordering.
	sort.Slice(allSecrets, func(i, j int) bool {
		return allSecrets[i].Name < allSecrets[j].Name
	})

	// Handle pagination.
	startIdx := 0

	if nextToken != "" {
		for i, s := range allSecrets {
			if s.Name == nextToken {
				startIdx = i
				break
			}
		}
	}

	endIdx := min(startIdx+maxResults, len(allSecrets))
	result := allSecrets[startIdx:endIdx]
	var newNextToken string

	if endIdx < len(allSecrets) {
		newNextToken = allSecrets[endIdx].Name
	}

	return result, newNextToken, nil
}

// DescribeSecret returns secret metadata.
func (m *MemoryStorage) DescribeSecret(_ context.Context, secretID string) (*Secret, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secret := m.findSecret(secretID)
	if secret == nil {
		return nil, &SecretError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Secrets Manager can't find the specified secret: %s", secretID),
		}
	}

	return secret, nil
}

// UpdateSecret updates a secret.
func (m *MemoryStorage) UpdateSecret(_ context.Context, req *UpdateSecretRequest) (*Secret, *SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	secret := m.findSecret(req.SecretID)
	if secret == nil {
		return nil, nil, &SecretError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Secrets Manager can't find the specified secret: %s", req.SecretID),
		}
	}

	if secret.DeletedDate != nil {
		return nil, nil, &SecretError{
			Code:    "InvalidRequestException",
			Message: "You can't perform this operation on a secret that's scheduled for deletion.",
		}
	}

	now := time.Now()

	if req.Description != "" {
		secret.Description = req.Description
	}

	if req.KmsKeyID != "" {
		secret.KmsKeyID = req.KmsKeyID
	}

	var version *SecretVersion

	// Create new version if secret value is provided.
	if req.SecretString != "" || len(req.SecretBinary) > 0 {
		versionID := req.ClientRequestToken
		if versionID == "" {
			versionID = uuid.New().String()
		}

		// Remove AWSCURRENT stage from previous version.
		for _, v := range secret.VersionIDs {
			newStages := make([]string, 0)

			for _, stage := range v.VersionStages {
				if stage != stageCurrent {
					newStages = append(newStages, stage)
				} else {
					newStages = append(newStages, stagePrevious)
				}
			}

			v.VersionStages = newStages
		}

		version = &SecretVersion{
			VersionID:     versionID,
			SecretString:  req.SecretString,
			SecretBinary:  req.SecretBinary,
			VersionStages: []string{stageCurrent},
			CreatedDate:   now,
			KmsKeyID:      secret.KmsKeyID,
		}

		secret.VersionIDs[versionID] = version
		secret.VersionID = versionID
		secret.SecretString = req.SecretString
		secret.SecretBinary = req.SecretBinary
	}

	secret.LastChangedDate = now

	return secret, version, nil
}

// findSecret finds a secret by name or ARN.
func (m *MemoryStorage) findSecret(secretID string) *Secret {
	// Try by name first.
	if secret, exists := m.secrets[secretID]; exists {
		return secret
	}

	// Try by ARN.
	for _, secret := range m.secrets {
		if secret.ARN == secretID {
			return secret
		}
	}

	return nil
}

// buildARN builds an ARN for a secret.
func (m *MemoryStorage) buildARN(name string) string {
	// Extract region from baseURL if possible, otherwise use default.
	region := defaultRegion

	return fmt.Sprintf("arn:aws:secretsmanager:%s:%s:secret:%s-%s",
		region, defaultAccountID, name, strings.ReplaceAll(uuid.New().String()[:6], "-", ""))
}
