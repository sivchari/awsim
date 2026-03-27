package iam

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sivchari/kumo/internal/storage"
)

const (
	defaultPath     = "/"
	defaultMaxItems = 100
	accessKeyActive = "Active"
)

// Error codes.
const (
	errEntityAlreadyExists = "EntityAlreadyExists"
	errNoSuchEntity        = "NoSuchEntity"
	errDeleteConflict      = "DeleteConflict"
	errLimitExceeded       = "LimitExceeded"
)

// Storage defines the IAM storage interface.
type Storage interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, userName string) error
	GetUser(ctx context.Context, userName string) (*User, error)
	ListUsers(ctx context.Context, pathPrefix string, maxItems int) ([]User, error)

	CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, roleName string) error
	GetRole(ctx context.Context, roleName string) (*Role, error)
	ListRoles(ctx context.Context, pathPrefix string, maxItems int) ([]Role, error)

	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*Policy, error)
	DeletePolicy(ctx context.Context, policyArn string) error
	GetPolicy(ctx context.Context, policyArn string) (*Policy, error)
	ListPolicies(ctx context.Context, pathPrefix string, maxItems int, onlyAttached bool) ([]Policy, error)

	AttachUserPolicy(ctx context.Context, userName, policyArn string) error
	DetachUserPolicy(ctx context.Context, userName, policyArn string) error
	AttachRolePolicy(ctx context.Context, roleName, policyArn string) error
	DetachRolePolicy(ctx context.Context, roleName, policyArn string) error

	CreateAccessKey(ctx context.Context, userName string) (*AccessKey, error)
	DeleteAccessKey(ctx context.Context, userName, accessKeyID string) error
	ListAccessKeys(ctx context.Context, userName string, maxItems int) ([]AccessKeyMetadata, error)
}

// Option is a configuration option for MemoryStorage.
type Option func(*MemoryStorage)

// WithDataDir enables persistent storage in the specified directory.
func WithDataDir(dir string) Option {
	return func(s *MemoryStorage) {
		s.dataDir = dir
	}
}

// Compile-time interface checks.
var (
	_ json.Marshaler   = (*MemoryStorage)(nil)
	_ json.Unmarshaler = (*MemoryStorage)(nil)
)

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu         sync.RWMutex                     `json:"-"`
	Users      map[string]*User                 `json:"users"`
	Roles      map[string]*Role                 `json:"roles"`
	Policies   map[string]*Policy               `json:"policies"`   // key is ARN
	AccessKeys map[string]map[string]*AccessKey `json:"accessKeys"` // userName -> accessKeyID -> AccessKey
	accountID  string
	dataDir    string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage(opts ...Option) *MemoryStorage {
	s := &MemoryStorage{
		Users:      make(map[string]*User),
		Roles:      make(map[string]*Role),
		Policies:   make(map[string]*Policy),
		AccessKeys: make(map[string]map[string]*AccessKey),
		accountID:  "123456789012",
	}
	for _, o := range opts {
		o(s)
	}

	if s.dataDir != "" {
		_ = storage.Load(s.dataDir, "iam", s)
	}

	return s
}

// MarshalJSON serializes the storage state to JSON.
func (s *MemoryStorage) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type Alias MemoryStorage

	data, err := json.Marshal(&struct{ *Alias }{Alias: (*Alias)(s)})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	return data, nil
}

// UnmarshalJSON restores the storage state from JSON.
func (s *MemoryStorage) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	type Alias MemoryStorage

	aux := &struct{ *Alias }{Alias: (*Alias)(s)}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	if s.Users == nil {
		s.Users = make(map[string]*User)
	}

	if s.Roles == nil {
		s.Roles = make(map[string]*Role)
	}

	if s.Policies == nil {
		s.Policies = make(map[string]*Policy)
	}

	if s.AccessKeys == nil {
		s.AccessKeys = make(map[string]map[string]*AccessKey)
	}

	return nil
}

// Close saves the storage state to disk if persistence is enabled.
func (s *MemoryStorage) Close() error {
	if s.dataDir == "" {
		return nil
	}

	if err := storage.Save(s.dataDir, "iam", s); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// CreateUser creates a new IAM user.
func (s *MemoryStorage) CreateUser(_ context.Context, req *CreateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Users[req.UserName]; exists {
		return nil, &Error{
			Code:    errEntityAlreadyExists,
			Message: fmt.Sprintf("User with name %s already exists.", req.UserName),
		}
	}

	path := req.Path
	if path == "" {
		path = defaultPath
	}

	user := &User{
		UserName:         req.UserName,
		UserID:           generateID("AIDA"),
		Arn:              fmt.Sprintf("arn:aws:iam::%s:user%s%s", s.accountID, path, req.UserName),
		Path:             path,
		CreateDate:       time.Now().UTC(),
		Tags:             req.Tags,
		AttachedPolicies: []AttachedPolicy{},
	}

	s.Users[req.UserName] = user
	s.AccessKeys[req.UserName] = make(map[string]*AccessKey)

	return user, nil
}

// DeleteUser deletes an IAM user.
func (s *MemoryStorage) DeleteUser(_ context.Context, userName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.Users[userName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	if len(user.AttachedPolicies) > 0 {
		return &Error{
			Code:    errDeleteConflict,
			Message: "Cannot delete entity, must detach all policies first.",
		}
	}

	if keys, ok := s.AccessKeys[userName]; ok && len(keys) > 0 {
		return &Error{
			Code:    errDeleteConflict,
			Message: "Cannot delete entity, must delete access keys first.",
		}
	}

	delete(s.Users, userName)
	delete(s.AccessKeys, userName)

	return nil
}

// GetUser gets an IAM user.
func (s *MemoryStorage) GetUser(_ context.Context, userName string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.Users[userName]
	if !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	return user, nil
}

// ListUsers lists IAM users.
func (s *MemoryStorage) ListUsers(_ context.Context, pathPrefix string, maxItems int) ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	if pathPrefix == "" {
		pathPrefix = defaultPath
	}

	users := make([]User, 0)

	for _, user := range s.Users {
		if strings.HasPrefix(user.Path, pathPrefix) {
			users = append(users, *user)
			if len(users) >= maxItems {
				break
			}
		}
	}

	return users, nil
}

// CreateRole creates a new IAM role.
func (s *MemoryStorage) CreateRole(_ context.Context, req *CreateRoleRequest) (*Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Roles[req.RoleName]; exists {
		return nil, &Error{
			Code:    errEntityAlreadyExists,
			Message: fmt.Sprintf("Role with name %s already exists.", req.RoleName),
		}
	}

	path := req.Path
	if path == "" {
		path = defaultPath
	}

	maxSessionDuration := req.MaxSessionDuration
	if maxSessionDuration == 0 {
		maxSessionDuration = 3600 // 1 hour default
	}

	role := &Role{
		RoleName:                 req.RoleName,
		RoleID:                   generateID("AROA"),
		Arn:                      fmt.Sprintf("arn:aws:iam::%s:role%s%s", s.accountID, path, req.RoleName),
		Path:                     path,
		CreateDate:               time.Now().UTC(),
		AssumeRolePolicyDocument: req.AssumeRolePolicyDocument,
		Description:              req.Description,
		MaxSessionDuration:       maxSessionDuration,
		Tags:                     req.Tags,
		AttachedPolicies:         []AttachedPolicy{},
	}

	s.Roles[req.RoleName] = role

	return role, nil
}

// DeleteRole deletes an IAM role.
func (s *MemoryStorage) DeleteRole(_ context.Context, roleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.Roles[roleName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The role with name %s cannot be found.", roleName),
		}
	}

	if len(role.AttachedPolicies) > 0 {
		return &Error{
			Code:    errDeleteConflict,
			Message: "Cannot delete entity, must detach all policies first.",
		}
	}

	delete(s.Roles, roleName)

	return nil
}

// GetRole gets an IAM role.
func (s *MemoryStorage) GetRole(_ context.Context, roleName string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.Roles[roleName]
	if !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The role with name %s cannot be found.", roleName),
		}
	}

	return role, nil
}

// ListRoles lists IAM roles.
func (s *MemoryStorage) ListRoles(_ context.Context, pathPrefix string, maxItems int) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	if pathPrefix == "" {
		pathPrefix = defaultPath
	}

	roles := make([]Role, 0)

	for _, role := range s.Roles {
		if strings.HasPrefix(role.Path, pathPrefix) {
			roles = append(roles, *role)
			if len(roles) >= maxItems {
				break
			}
		}
	}

	return roles, nil
}

// CreatePolicy creates a new IAM policy.
func (s *MemoryStorage) CreatePolicy(_ context.Context, req *CreatePolicyRequest) (*Policy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := req.Path
	if path == "" {
		path = defaultPath
	}

	arn := fmt.Sprintf("arn:aws:iam::%s:policy%s%s", s.accountID, path, req.PolicyName)

	if _, exists := s.Policies[arn]; exists {
		return nil, &Error{
			Code:    errEntityAlreadyExists,
			Message: fmt.Sprintf("A policy called %s already exists. Duplicate names are not allowed.", req.PolicyName),
		}
	}

	now := time.Now().UTC()

	policy := &Policy{
		PolicyName:       req.PolicyName,
		PolicyID:         generateID("ANPA"),
		Arn:              arn,
		Path:             path,
		DefaultVersionID: "v1",
		AttachmentCount:  0,
		IsAttachable:     true,
		CreateDate:       now,
		UpdateDate:       now,
		Description:      req.Description,
		Tags:             req.Tags,
		PolicyDocument:   req.PolicyDocument,
	}

	s.Policies[arn] = policy

	return policy, nil
}

// DeletePolicy deletes an IAM policy.
func (s *MemoryStorage) DeletePolicy(_ context.Context, policyArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.Policies[policyArn]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s does not exist.", policyArn),
		}
	}

	if policy.AttachmentCount > 0 {
		return &Error{
			Code:    errDeleteConflict,
			Message: "Cannot delete a policy attached to entities.",
		}
	}

	delete(s.Policies, policyArn)

	return nil
}

// GetPolicy gets an IAM policy.
func (s *MemoryStorage) GetPolicy(_ context.Context, policyArn string) (*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.Policies[policyArn]
	if !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s does not exist.", policyArn),
		}
	}

	return policy, nil
}

// ListPolicies lists IAM policies.
func (s *MemoryStorage) ListPolicies(_ context.Context, pathPrefix string, maxItems int, onlyAttached bool) ([]Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	if pathPrefix == "" {
		pathPrefix = defaultPath
	}

	policies := make([]Policy, 0)

	for _, policy := range s.Policies {
		if !strings.HasPrefix(policy.Path, pathPrefix) {
			continue
		}

		if onlyAttached && policy.AttachmentCount == 0 {
			continue
		}

		policies = append(policies, *policy)

		if len(policies) >= maxItems {
			break
		}
	}

	return policies, nil
}

// AttachUserPolicy attaches a policy to a user.
func (s *MemoryStorage) AttachUserPolicy(_ context.Context, userName, policyArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.Users[userName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	policy, exists := s.Policies[policyArn]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s does not exist.", policyArn),
		}
	}

	for _, ap := range user.AttachedPolicies {
		if ap.PolicyArn == policyArn {
			return nil // Already attached
		}
	}

	user.AttachedPolicies = append(user.AttachedPolicies, AttachedPolicy{
		PolicyName: policy.PolicyName,
		PolicyArn:  policyArn,
	})
	policy.AttachmentCount++

	return nil
}

// DetachUserPolicy detaches a policy from a user.
func (s *MemoryStorage) DetachUserPolicy(_ context.Context, userName, policyArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.Users[userName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	policy, exists := s.Policies[policyArn]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s does not exist.", policyArn),
		}
	}

	found := false

	for i, ap := range user.AttachedPolicies {
		if ap.PolicyArn == policyArn {
			user.AttachedPolicies = append(user.AttachedPolicies[:i], user.AttachedPolicies[i+1:]...)
			found = true

			break
		}
	}

	if !found {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s is not attached to user %s.", policyArn, userName),
		}
	}

	policy.AttachmentCount--

	return nil
}

// AttachRolePolicy attaches a policy to a role.
func (s *MemoryStorage) AttachRolePolicy(_ context.Context, roleName, policyArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.Roles[roleName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The role with name %s cannot be found.", roleName),
		}
	}

	policy, exists := s.Policies[policyArn]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s does not exist.", policyArn),
		}
	}

	for _, ap := range role.AttachedPolicies {
		if ap.PolicyArn == policyArn {
			return nil // Already attached
		}
	}

	role.AttachedPolicies = append(role.AttachedPolicies, AttachedPolicy{
		PolicyName: policy.PolicyName,
		PolicyArn:  policyArn,
	})
	policy.AttachmentCount++

	return nil
}

// DetachRolePolicy detaches a policy from a role.
func (s *MemoryStorage) DetachRolePolicy(_ context.Context, roleName, policyArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.Roles[roleName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The role with name %s cannot be found.", roleName),
		}
	}

	policy, exists := s.Policies[policyArn]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s does not exist.", policyArn),
		}
	}

	found := false

	for i, ap := range role.AttachedPolicies {
		if ap.PolicyArn == policyArn {
			role.AttachedPolicies = append(role.AttachedPolicies[:i], role.AttachedPolicies[i+1:]...)
			found = true

			break
		}
	}

	if !found {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("Policy %s is not attached to role %s.", policyArn, roleName),
		}
	}

	policy.AttachmentCount--

	return nil
}

// CreateAccessKey creates a new access key for a user.
func (s *MemoryStorage) CreateAccessKey(_ context.Context, userName string) (*AccessKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Users[userName]; !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	keys := s.AccessKeys[userName]
	if len(keys) >= 2 {
		return nil, &Error{
			Code:    errLimitExceeded,
			Message: "Cannot exceed quota for AccessKeysPerUser: 2",
		}
	}

	accessKey := &AccessKey{
		AccessKeyID:     generateAccessKeyID(),
		SecretAccessKey: generateSecretAccessKey(),
		Status:          accessKeyActive,
		UserName:        userName,
		CreateDate:      time.Now().UTC(),
	}

	s.AccessKeys[userName][accessKey.AccessKeyID] = accessKey

	return accessKey, nil
}

// DeleteAccessKey deletes an access key.
func (s *MemoryStorage) DeleteAccessKey(_ context.Context, userName, accessKeyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Users[userName]; !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	keys, exists := s.AccessKeys[userName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The Access Key with id %s cannot be found.", accessKeyID),
		}
	}

	if _, exists := keys[accessKeyID]; !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The Access Key with id %s cannot be found.", accessKeyID),
		}
	}

	delete(s.AccessKeys[userName], accessKeyID)

	return nil
}

// ListAccessKeys lists access keys for a user.
func (s *MemoryStorage) ListAccessKeys(_ context.Context, userName string, maxItems int) ([]AccessKeyMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.Users[userName]; !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	keys := make([]AccessKeyMetadata, 0)

	for _, key := range s.AccessKeys[userName] {
		keys = append(keys, AccessKeyMetadata{
			AccessKeyID: key.AccessKeyID,
			Status:      key.Status,
			UserName:    key.UserName,
			CreateDate:  key.CreateDate,
		})

		if len(keys) >= maxItems {
			break
		}
	}

	return keys, nil
}

// generateID generates a unique ID with a prefix.
func generateID(prefix string) string {
	return prefix + strings.ToUpper(uuid.New().String()[:17])
}

// generateAccessKeyID generates an AWS-style access key ID.
func generateAccessKeyID() string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)

	return "AKIA" + strings.ToUpper(hex.EncodeToString(b))[:16]
}

// generateSecretAccessKey generates an AWS-style secret access key.
func generateSecretAccessKey() string {
	b := make([]byte, 30)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)[:40]
}
