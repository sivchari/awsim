package sesv2

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errNotFound         = "NotFoundException"
	errAlreadyExists    = "AlreadyExistsException"
	errInvalidParameter = "ValidationException"
	errBadRequest       = "BadRequestException"
)

// Storage defines the interface for SES v2 storage operations.
type Storage interface {
	// Email Identity operations.
	CreateEmailIdentity(ctx context.Context, req *CreateEmailIdentityRequest) (*EmailIdentity, error)
	GetEmailIdentity(ctx context.Context, emailIdentity string) (*EmailIdentity, error)
	ListEmailIdentities(ctx context.Context, nextToken string, pageSize int32) ([]*EmailIdentity, string, error)
	DeleteEmailIdentity(ctx context.Context, emailIdentity string) error

	// Configuration Set operations.
	CreateConfigurationSet(ctx context.Context, req *CreateConfigurationSetRequest) (*ConfigurationSet, error)
	GetConfigurationSet(ctx context.Context, name string) (*ConfigurationSet, error)
	ListConfigurationSets(ctx context.Context, nextToken string, pageSize int32) ([]string, string, error)
	DeleteConfigurationSet(ctx context.Context, name string) error

	// Send Email.
	SendEmail(ctx context.Context, req *SendEmailRequest) (string, error)

	// Get sent emails (for testing purposes).
	GetSentEmails(ctx context.Context) ([]*SentEmail, error)
}

// MemoryStorage implements Storage with in-memory data structures.
type MemoryStorage struct {
	mu                sync.RWMutex
	emailIdentities   map[string]*EmailIdentity
	configurationSets map[string]*ConfigurationSet
	sentEmails        []*SentEmail
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		emailIdentities:   make(map[string]*EmailIdentity),
		configurationSets: make(map[string]*ConfigurationSet),
		sentEmails:        make([]*SentEmail, 0),
	}
}

// CreateEmailIdentity creates a new email identity.
func (s *MemoryStorage) CreateEmailIdentity(_ context.Context, req *CreateEmailIdentityRequest) (*EmailIdentity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.EmailIdentity == "" {
		return nil, &IdentityError{
			Code:    errInvalidParameter,
			Message: "EmailIdentity is required",
		}
	}

	if _, exists := s.emailIdentities[req.EmailIdentity]; exists {
		return nil, &IdentityError{
			Code:    errAlreadyExists,
			Message: "The email identity already exists",
		}
	}

	identityType := "EMAIL_ADDRESS"
	if !strings.Contains(req.EmailIdentity, "@") {
		identityType = "DOMAIN"
	}

	identity := &EmailIdentity{
		IdentityName:             req.EmailIdentity,
		IdentityType:             identityType,
		VerifiedForSendingStatus: true, // Auto-verify for testing.
		DkimAttributes: &DkimAttributes{
			SigningEnabled:          true,
			Status:                  "SUCCESS",
			SigningAttributesOrigin: "AWS_SES",
			Tokens:                  []string{uuid.New().String()[:20]},
		},
		CreatedAt: time.Now(),
	}

	s.emailIdentities[req.EmailIdentity] = identity

	return identity, nil
}

// GetEmailIdentity retrieves an email identity.
func (s *MemoryStorage) GetEmailIdentity(_ context.Context, emailIdentity string) (*EmailIdentity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identity, exists := s.emailIdentities[emailIdentity]
	if !exists {
		return nil, &IdentityError{
			Code:    errNotFound,
			Message: "The email identity does not exist",
		}
	}

	return identity, nil
}

// ListEmailIdentities lists all email identities.
func (s *MemoryStorage) ListEmailIdentities(_ context.Context, _ string, pageSize int32) ([]*EmailIdentity, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if pageSize <= 0 {
		pageSize = 100
	}

	identities := make([]*EmailIdentity, 0, len(s.emailIdentities))
	for _, identity := range s.emailIdentities {
		identities = append(identities, identity)
	}

	// Simple pagination (no actual cursor).
	if len(identities) > int(pageSize) {
		identities = identities[:pageSize]
	}

	return identities, "", nil
}

// DeleteEmailIdentity deletes an email identity.
func (s *MemoryStorage) DeleteEmailIdentity(_ context.Context, emailIdentity string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.emailIdentities[emailIdentity]; !exists {
		return &IdentityError{
			Code:    errNotFound,
			Message: "The email identity does not exist",
		}
	}

	delete(s.emailIdentities, emailIdentity)

	return nil
}

// CreateConfigurationSet creates a new configuration set.
func (s *MemoryStorage) CreateConfigurationSet(_ context.Context, req *CreateConfigurationSetRequest) (*ConfigurationSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.ConfigurationSetName == "" {
		return nil, &IdentityError{
			Code:    errInvalidParameter,
			Message: "ConfigurationSetName is required",
		}
	}

	if _, exists := s.configurationSets[req.ConfigurationSetName]; exists {
		return nil, &IdentityError{
			Code:    errAlreadyExists,
			Message: "The configuration set already exists",
		}
	}

	configSet := &ConfigurationSet{
		Name:              req.ConfigurationSetName,
		DeliveryOptions:   req.DeliveryOptions,
		ReputationOptions: req.ReputationOptions,
		SendingOptions:    req.SendingOptions,
		TrackingOptions:   req.TrackingOptions,
		Tags:              req.Tags,
	}

	// Set defaults if not provided.
	if configSet.SendingOptions == nil {
		configSet.SendingOptions = &SendingOptions{SendingEnabled: true}
	}

	s.configurationSets[req.ConfigurationSetName] = configSet

	return configSet, nil
}

// GetConfigurationSet retrieves a configuration set.
func (s *MemoryStorage) GetConfigurationSet(_ context.Context, name string) (*ConfigurationSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configSet, exists := s.configurationSets[name]
	if !exists {
		return nil, &IdentityError{
			Code:    errNotFound,
			Message: "The configuration set does not exist",
		}
	}

	return configSet, nil
}

// ListConfigurationSets lists all configuration sets.
func (s *MemoryStorage) ListConfigurationSets(_ context.Context, _ string, pageSize int32) ([]string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if pageSize <= 0 {
		pageSize = 100
	}

	names := make([]string, 0, len(s.configurationSets))
	for name := range s.configurationSets {
		names = append(names, name)
	}

	if len(names) > int(pageSize) {
		names = names[:pageSize]
	}

	return names, "", nil
}

// DeleteConfigurationSet deletes a configuration set.
func (s *MemoryStorage) DeleteConfigurationSet(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configurationSets[name]; !exists {
		return &IdentityError{
			Code:    errNotFound,
			Message: "The configuration set does not exist",
		}
	}

	delete(s.configurationSets, name)

	return nil
}

// SendEmail sends an email (stores it for testing).
func (s *MemoryStorage) SendEmail(_ context.Context, req *SendEmailRequest) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Basic validation.
	if req.Destination == nil ||
		(len(req.Destination.ToAddresses) == 0 &&
			len(req.Destination.CcAddresses) == 0 &&
			len(req.Destination.BccAddresses) == 0) {
		return "", &IdentityError{
			Code:    errBadRequest,
			Message: "Destination is required",
		}
	}

	if req.Content == nil {
		return "", &IdentityError{
			Code:    errBadRequest,
			Message: "Content is required",
		}
	}

	// Generate message ID.
	messageID := uuid.New().String()

	// Extract subject and body from simple email content.
	subject, body := extractSimpleEmailContent(req.Content.Simple)

	// Store the sent email.
	sentEmail := &SentEmail{
		MessageID:            messageID,
		FromEmailAddress:     req.FromEmailAddress,
		Destination:          req.Destination,
		Subject:              subject,
		Body:                 body,
		ConfigurationSetName: req.ConfigurationSetName,
		SentAt:               time.Now(),
	}

	s.sentEmails = append(s.sentEmails, sentEmail)

	return messageID, nil
}

// GetSentEmails returns all sent emails (for testing).
func (s *MemoryStorage) GetSentEmails(_ context.Context) ([]*SentEmail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sentEmails, nil
}

// extractSimpleEmailContent extracts subject and body from a SimpleEmail.
func extractSimpleEmailContent(simple *SimpleEmail) (subject, body string) {
	if simple == nil {
		return "", ""
	}

	if simple.Subject != nil {
		subject = simple.Subject.Data
	}

	if simple.Body == nil {
		return subject, ""
	}

	if simple.Body.Text != nil {
		return subject, simple.Body.Text.Data
	}

	if simple.Body.HTML != nil {
		return subject, simple.Body.HTML.Data
	}

	return subject, ""
}
