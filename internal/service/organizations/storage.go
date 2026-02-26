package organizations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errAWSOrganizationsNotInUseException    = "AWSOrganizationsNotInUseException"
	errAccountNotFoundException             = "AccountNotFoundException"
	errAlreadyInOrganizationException       = "AlreadyInOrganizationException"
	errOrganizationNotEmptyException        = "OrganizationNotEmptyException"
	errParentNotFoundException              = "ParentNotFoundException"
	errDuplicateOrganizationalUnitException = "DuplicateOrganizationalUnitException"
	errPolicyNotFoundException              = "PolicyNotFoundException"
	errPolicyNotAttachedException           = "PolicyNotAttachedException"
	errDuplicatePolicyAttachmentException   = "DuplicatePolicyAttachmentException"
	errInvalidInputException                = "InvalidInputException"
)

// Default values.
const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "123456789012"
)

// Feature set values.
const (
	featureSetAll = "ALL"
)

// Account status values.
const (
	accountStatusActive = "ACTIVE"
)

// Account state values.
const (
	accountStateActive = "ACTIVE"
)

// Joined method values.
const (
	joinedMethodCreated = "CREATED"
)

// Create account status values.
const (
	createAccountStatusSucceeded = "SUCCEEDED"
)

// Policy type values.
const (
	policyTypeServiceControlPolicy = "SERVICE_CONTROL_POLICY"
)

// Policy status values.
const (
	policyStatusEnabled = "ENABLED"
)

// Storage defines the Organizations storage interface.
type Storage interface {
	// Organization operations
	CreateOrganization(ctx context.Context, featureSet string) (*Organization, error)
	DeleteOrganization(ctx context.Context) error
	DescribeOrganization(ctx context.Context) (*Organization, error)

	// Account operations
	CreateAccount(ctx context.Context, req *CreateAccountInput) (*CreateAccountStatus, error)
	DescribeAccount(ctx context.Context, accountID string) (*Account, error)
	ListAccounts(ctx context.Context, maxResults int32, nextToken string) ([]*Account, string, error)

	// Organizational unit operations
	CreateOrganizationalUnit(ctx context.Context, name, parentID string) (*OrganizationalUnit, error)
	ListOrganizationalUnitsForParent(ctx context.Context, parentID string, maxResults int32, nextToken string) ([]*OrganizationalUnit, string, error)

	// Policy operations
	AttachPolicy(ctx context.Context, policyID, targetID string) error
	DetachPolicy(ctx context.Context, policyID, targetID string) error

	// Root operations
	ListRoots(ctx context.Context, maxResults int32, nextToken string) ([]*Root, string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu                  sync.RWMutex
	organization        *Organization
	root                *Root
	accounts            map[string]*Account            // accountID -> account
	organizationalUnits map[string]*OrganizationalUnit // ouID -> OU
	ouParents           map[string]string              // ouID -> parentID
	policies            map[string]*Policy             // policyID -> policy
	policyAttachments   map[string]map[string]bool     // targetID -> policyID -> attached
	region              string
	accountID           string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		accounts:            make(map[string]*Account),
		organizationalUnits: make(map[string]*OrganizationalUnit),
		ouParents:           make(map[string]string),
		policies:            make(map[string]*Policy),
		policyAttachments:   make(map[string]map[string]bool),
		region:              defaultRegion,
		accountID:           defaultAccountID,
	}

	storage.initializeDefaultPolicy()

	return storage
}

func (m *MemoryStorage) initializeDefaultPolicy() {
	// Create a default full access SCP.
	defaultSCPID := "p-FullAWSAccess"
	m.policies[defaultSCPID] = &Policy{
		Content: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`,
		PolicySummary: &PolicySummary{
			ARN:         fmt.Sprintf("arn:aws:organizations::%s:policy/o-example/service_control_policy/%s", defaultAccountID, defaultSCPID),
			AWSManaged:  true,
			Description: "Allows access to every operation",
			ID:          defaultSCPID,
			Name:        "FullAWSAccess",
			Type:        policyTypeServiceControlPolicy,
		},
	}
}

// CreateOrganization creates a new organization.
func (m *MemoryStorage) CreateOrganization(_ context.Context, featureSet string) (*Organization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.organization != nil {
		return nil, &Error{Code: errAlreadyInOrganizationException, Message: "You are already a member of an organization"}
	}

	if featureSet == "" {
		featureSet = featureSetAll
	}

	orgID := "o-" + generateShortID()
	rootID := "r-" + generateShortID()[:4]

	m.organization = &Organization{
		ARN:                fmt.Sprintf("arn:aws:organizations::%s:organization/%s", m.accountID, orgID),
		FeatureSet:         featureSet,
		ID:                 orgID,
		MasterAccountARN:   fmt.Sprintf("arn:aws:organizations::%s:account/%s/%s", m.accountID, orgID, m.accountID),
		MasterAccountEmail: "master@example.com",
		MasterAccountID:    m.accountID,
	}

	if featureSet == featureSetAll {
		m.organization.AvailablePolicyTypes = []PolicyTypeSummary{
			{Status: policyStatusEnabled, Type: policyTypeServiceControlPolicy},
		}
	}

	// Create the root.
	m.root = &Root{
		ARN:  fmt.Sprintf("arn:aws:organizations::%s:root/%s/%s", m.accountID, orgID, rootID),
		ID:   rootID,
		Name: "Root",
	}

	if featureSet == featureSetAll {
		m.root.PolicyTypes = []PolicyTypeSummary{
			{Status: policyStatusEnabled, Type: policyTypeServiceControlPolicy},
		}
	}

	// Add the management account.
	m.accounts[m.accountID] = &Account{
		ARN:             m.organization.MasterAccountARN,
		Email:           m.organization.MasterAccountEmail,
		ID:              m.accountID,
		JoinedMethod:    joinedMethodCreated,
		JoinedTimestamp: ToAWSTimestamp(time.Now()),
		Name:            "Management Account",
		State:           accountStateActive,
		Status:          accountStatusActive,
	}

	// Attach the default SCP to the root.
	m.policyAttachments[rootID] = map[string]bool{
		"p-FullAWSAccess": true,
	}

	return m.organization, nil
}

// DeleteOrganization deletes the organization.
func (m *MemoryStorage) DeleteOrganization(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.organization == nil {
		return &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	// Check if there are any member accounts (besides the management account).
	if len(m.accounts) > 1 {
		return &Error{Code: errOrganizationNotEmptyException, Message: "Organization still has member accounts"}
	}

	// Check if there are any OUs.
	if len(m.organizationalUnits) > 0 {
		return &Error{Code: errOrganizationNotEmptyException, Message: "Organization still has organizational units"}
	}

	// Delete the organization.
	m.organization = nil
	m.root = nil
	m.accounts = make(map[string]*Account)
	m.policyAttachments = make(map[string]map[string]bool)

	return nil
}

// DescribeOrganization returns the organization.
func (m *MemoryStorage) DescribeOrganization(_ context.Context) (*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.organization == nil {
		return nil, &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	return m.organization, nil
}

// CreateAccount creates a new account.
func (m *MemoryStorage) CreateAccount(_ context.Context, req *CreateAccountInput) (*CreateAccountStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.organization == nil {
		return nil, &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	if req.AccountName == "" || req.Email == "" {
		return nil, &Error{Code: errInvalidInputException, Message: "AccountName and Email are required"}
	}

	// Generate a new account ID.
	accountID := generateAccountID()
	requestID := uuid.New().String()
	now := time.Now()

	// Create the account.
	account := &Account{
		ARN:             fmt.Sprintf("arn:aws:organizations::%s:account/%s/%s", m.accountID, m.organization.ID, accountID),
		Email:           req.Email,
		ID:              accountID,
		JoinedMethod:    joinedMethodCreated,
		JoinedTimestamp: ToAWSTimestamp(now),
		Name:            req.AccountName,
		State:           accountStateActive,
		Status:          accountStatusActive,
	}

	m.accounts[accountID] = account

	// Attach the default SCP to the new account.
	m.policyAttachments[accountID] = map[string]bool{
		"p-FullAWSAccess": true,
	}

	// Create the status.
	status := &CreateAccountStatus{
		AccountID:          accountID,
		AccountName:        req.AccountName,
		CompletedTimestamp: ToAWSTimestamp(now),
		ID:                 requestID,
		RequestedTimestamp: ToAWSTimestamp(now),
		State:              createAccountStatusSucceeded,
	}

	return status, nil
}

// DescribeAccount returns an account.
func (m *MemoryStorage) DescribeAccount(_ context.Context, accountID string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.organization == nil {
		return nil, &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	account, exists := m.accounts[accountID]
	if !exists {
		return nil, &Error{Code: errAccountNotFoundException, Message: "Account not found: " + accountID}
	}

	return account, nil
}

// ListAccounts lists all accounts.
func (m *MemoryStorage) ListAccounts(_ context.Context, maxResults int32, _ string) ([]*Account, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.organization == nil {
		return nil, "", &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	result := make([]*Account, 0, len(m.accounts))

	for _, account := range m.accounts {
		result = append(result, account)

		//nolint:gosec // len(result) is bounded by the number of accounts which is limited.
		if maxResults > 0 && int32(len(result)) >= maxResults {
			break
		}
	}

	return result, "", nil
}

// CreateOrganizationalUnit creates a new organizational unit.
func (m *MemoryStorage) CreateOrganizationalUnit(_ context.Context, name, parentID string) (*OrganizationalUnit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.organization == nil {
		return nil, &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	// Validate parent ID.
	if !m.isValidParentID(parentID) {
		return nil, &Error{Code: errParentNotFoundException, Message: "Parent not found: " + parentID}
	}

	// Check for duplicate OU name under the same parent.
	for ouID, ou := range m.organizationalUnits {
		if m.ouParents[ouID] == parentID && ou.Name == name {
			return nil, &Error{Code: errDuplicateOrganizationalUnitException, Message: "OU with this name already exists under the parent"}
		}
	}

	// Generate OU ID.
	ouID := fmt.Sprintf("ou-%s-%s", generateShortID()[:4], generateShortID()[:8])

	ou := &OrganizationalUnit{
		ARN:  fmt.Sprintf("arn:aws:organizations::%s:ou/%s/%s", m.accountID, m.organization.ID, ouID),
		ID:   ouID,
		Name: name,
	}

	m.organizationalUnits[ouID] = ou
	m.ouParents[ouID] = parentID

	// Attach the default SCP to the new OU.
	m.policyAttachments[ouID] = map[string]bool{
		"p-FullAWSAccess": true,
	}

	return ou, nil
}

// ListOrganizationalUnitsForParent lists OUs under a parent.
func (m *MemoryStorage) ListOrganizationalUnitsForParent(_ context.Context, parentID string, maxResults int32, _ string) ([]*OrganizationalUnit, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.organization == nil {
		return nil, "", &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	// Validate parent ID.
	if !m.isValidParentID(parentID) {
		return nil, "", &Error{Code: errParentNotFoundException, Message: "Parent not found: " + parentID}
	}

	result := make([]*OrganizationalUnit, 0)

	for ouID, ou := range m.organizationalUnits {
		if m.ouParents[ouID] == parentID {
			result = append(result, ou)

			//nolint:gosec // len(result) is bounded by the number of OUs which is limited.
			if maxResults > 0 && int32(len(result)) >= maxResults {
				break
			}
		}
	}

	return result, "", nil
}

// AttachPolicy attaches a policy to a target.
func (m *MemoryStorage) AttachPolicy(_ context.Context, policyID, targetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.organization == nil {
		return &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	// Validate policy ID.
	if _, exists := m.policies[policyID]; !exists {
		return &Error{Code: errPolicyNotFoundException, Message: "Policy not found: " + policyID}
	}

	// Validate target ID.
	if !m.isValidTargetID(targetID) {
		return &Error{Code: errInvalidInputException, Message: "Invalid target ID: " + targetID}
	}

	// Check if already attached.
	if attachments, exists := m.policyAttachments[targetID]; exists {
		if attachments[policyID] {
			return &Error{Code: errDuplicatePolicyAttachmentException, Message: "Policy is already attached to the target"}
		}
	}

	// Attach the policy.
	if m.policyAttachments[targetID] == nil {
		m.policyAttachments[targetID] = make(map[string]bool)
	}

	m.policyAttachments[targetID][policyID] = true

	return nil
}

// DetachPolicy detaches a policy from a target.
func (m *MemoryStorage) DetachPolicy(_ context.Context, policyID, targetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.organization == nil {
		return &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	// Validate policy ID.
	if _, exists := m.policies[policyID]; !exists {
		return &Error{Code: errPolicyNotFoundException, Message: "Policy not found: " + policyID}
	}

	// Validate target ID.
	if !m.isValidTargetID(targetID) {
		return &Error{Code: errInvalidInputException, Message: "Invalid target ID: " + targetID}
	}

	// Check if attached.
	attachments, exists := m.policyAttachments[targetID]
	if !exists || !attachments[policyID] {
		return &Error{Code: errPolicyNotAttachedException, Message: "Policy is not attached to the target"}
	}

	// Detach the policy.
	delete(m.policyAttachments[targetID], policyID)

	return nil
}

// ListRoots lists the roots.
func (m *MemoryStorage) ListRoots(_ context.Context, _ int32, _ string) ([]*Root, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.organization == nil {
		return nil, "", &Error{Code: errAWSOrganizationsNotInUseException, Message: "Your account is not a member of an organization"}
	}

	if m.root == nil {
		return []*Root{}, "", nil
	}

	return []*Root{m.root}, "", nil
}

// isValidParentID checks if a parent ID is valid (root or OU).
func (m *MemoryStorage) isValidParentID(parentID string) bool {
	// Check if it's the root.
	if m.root != nil && m.root.ID == parentID {
		return true
	}

	// Check if it's an OU.
	_, exists := m.organizationalUnits[parentID]

	return exists
}

// isValidTargetID checks if a target ID is valid (root, OU, or account).
func (m *MemoryStorage) isValidTargetID(targetID string) bool {
	// Check if it's the root.
	if m.root != nil && m.root.ID == targetID {
		return true
	}

	// Check if it's an OU.
	if _, exists := m.organizationalUnits[targetID]; exists {
		return true
	}

	// Check if it's an account.
	_, exists := m.accounts[targetID]

	return exists
}

// Helper functions.

func generateShortID() string {
	return uuid.New().String()[:12]
}

func generateAccountID() string {
	// Generate a 12-digit account ID.
	id := uint64(uuid.New().ID())

	return fmt.Sprintf("%012d", id%1000000000000)
}
