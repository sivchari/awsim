package iam

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu         sync.RWMutex
	users      map[string]*User
	roles      map[string]*Role
	policies   map[string]*Policy               // key is ARN
	accessKeys map[string]map[string]*AccessKey // userName -> accessKeyID -> AccessKey
	accountID  string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:      make(map[string]*User),
		roles:      make(map[string]*Role),
		policies:   make(map[string]*Policy),
		accessKeys: make(map[string]map[string]*AccessKey),
		accountID:  "123456789012",
	}
}

// CreateUser creates a new IAM user.
func (s *MemoryStorage) CreateUser(_ context.Context, req *CreateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[req.UserName]; exists {
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

	s.users[req.UserName] = user
	s.accessKeys[req.UserName] = make(map[string]*AccessKey)

	return user, nil
}

// DeleteUser deletes an IAM user.
func (s *MemoryStorage) DeleteUser(_ context.Context, userName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userName]
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

	if keys, ok := s.accessKeys[userName]; ok && len(keys) > 0 {
		return &Error{
			Code:    errDeleteConflict,
			Message: "Cannot delete entity, must delete access keys first.",
		}
	}

	delete(s.users, userName)
	delete(s.accessKeys, userName)

	return nil
}

// GetUser gets an IAM user.
func (s *MemoryStorage) GetUser(_ context.Context, userName string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userName]
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

	for _, user := range s.users {
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

	if _, exists := s.roles[req.RoleName]; exists {
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

	s.roles[req.RoleName] = role

	return role, nil
}

// DeleteRole deletes an IAM role.
func (s *MemoryStorage) DeleteRole(_ context.Context, roleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[roleName]
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

	delete(s.roles, roleName)

	return nil
}

// GetRole gets an IAM role.
func (s *MemoryStorage) GetRole(_ context.Context, roleName string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[roleName]
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

	for _, role := range s.roles {
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

	if _, exists := s.policies[arn]; exists {
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

	s.policies[arn] = policy

	return policy, nil
}

// DeletePolicy deletes an IAM policy.
func (s *MemoryStorage) DeletePolicy(_ context.Context, policyArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[policyArn]
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

	delete(s.policies, policyArn)

	return nil
}

// GetPolicy gets an IAM policy.
func (s *MemoryStorage) GetPolicy(_ context.Context, policyArn string) (*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[policyArn]
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

	for _, policy := range s.policies {
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

	user, exists := s.users[userName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	policy, exists := s.policies[policyArn]
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

	user, exists := s.users[userName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	policy, exists := s.policies[policyArn]
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

	role, exists := s.roles[roleName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The role with name %s cannot be found.", roleName),
		}
	}

	policy, exists := s.policies[policyArn]
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

	role, exists := s.roles[roleName]
	if !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The role with name %s cannot be found.", roleName),
		}
	}

	policy, exists := s.policies[policyArn]
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

	if _, exists := s.users[userName]; !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	keys := s.accessKeys[userName]
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

	s.accessKeys[userName][accessKey.AccessKeyID] = accessKey

	return accessKey, nil
}

// DeleteAccessKey deletes an access key.
func (s *MemoryStorage) DeleteAccessKey(_ context.Context, userName, accessKeyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[userName]; !exists {
		return &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	keys, exists := s.accessKeys[userName]
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

	delete(s.accessKeys[userName], accessKeyID)

	return nil
}

// ListAccessKeys lists access keys for a user.
func (s *MemoryStorage) ListAccessKeys(_ context.Context, userName string, maxItems int) ([]AccessKeyMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.users[userName]; !exists {
		return nil, &Error{
			Code:    errNoSuchEntity,
			Message: fmt.Sprintf("The user with name %s cannot be found.", userName),
		}
	}

	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	keys := make([]AccessKeyMetadata, 0)

	for _, key := range s.accessKeys[userName] {
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
