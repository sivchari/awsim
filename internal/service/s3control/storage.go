package s3control

import (
	"context"
	"fmt"
	"sync"
)

// Storage is the interface for S3 Control storage operations.
type Storage interface {
	// Public Access Block operations
	GetPublicAccessBlock(ctx context.Context, accountID string) (*PublicAccessBlockConfiguration, error)
	PutPublicAccessBlock(ctx context.Context, accountID string, config *PublicAccessBlockConfiguration) error
	DeletePublicAccessBlock(ctx context.Context, accountID string) error

	// Access Point operations
	CreateAccessPoint(ctx context.Context, accountID string, ap *AccessPoint) (*AccessPoint, error)
	GetAccessPoint(ctx context.Context, accountID, name string) (*AccessPoint, error)
	DeleteAccessPoint(ctx context.Context, accountID, name string) error
	ListAccessPoints(ctx context.Context, accountID, bucket string, maxResults int, nextToken string) ([]*AccessPoint, string, error)
}

// MemoryStorage implements in-memory storage for S3 Control.
type MemoryStorage struct {
	mu                 sync.RWMutex
	publicAccessBlocks map[string]*PublicAccessBlockConfiguration // key: accountID
	accessPoints       map[string]map[string]*AccessPoint         // key: accountID -> name -> AccessPoint
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		publicAccessBlocks: make(map[string]*PublicAccessBlockConfiguration),
		accessPoints:       make(map[string]map[string]*AccessPoint),
	}
}

// GetPublicAccessBlock retrieves the public access block configuration for an account.
func (s *MemoryStorage) GetPublicAccessBlock(_ context.Context, accountID string) (*PublicAccessBlockConfiguration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.publicAccessBlocks[accountID]
	if !exists {
		return nil, &Error{
			Code:    ErrNoSuchPublicAccessBlockConfiguration,
			Message: "The public access block configuration was not found",
		}
	}

	return config, nil
}

// PutPublicAccessBlock sets the public access block configuration for an account.
func (s *MemoryStorage) PutPublicAccessBlock(_ context.Context, accountID string, config *PublicAccessBlockConfiguration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.publicAccessBlocks[accountID] = config

	return nil
}

// DeletePublicAccessBlock removes the public access block configuration for an account.
func (s *MemoryStorage) DeletePublicAccessBlock(_ context.Context, accountID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.publicAccessBlocks, accountID)

	return nil
}

// CreateAccessPoint creates a new access point.
func (s *MemoryStorage) CreateAccessPoint(_ context.Context, accountID string, ap *AccessPoint) (*AccessPoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accessPoints[accountID]; !exists {
		s.accessPoints[accountID] = make(map[string]*AccessPoint)
	}

	if _, exists := s.accessPoints[accountID][ap.Name]; exists {
		return nil, &Error{
			Code:    ErrAccessPointAlreadyOwnedByYou,
			Message: fmt.Sprintf("Access point %s already exists", ap.Name),
		}
	}

	// Generate ARN and alias
	ap.AccountID = accountID
	ap.AccessPointArn = fmt.Sprintf("arn:aws:s3:%s:%s:accesspoint/%s", "us-east-1", accountID, ap.Name)
	ap.Alias = fmt.Sprintf("%s-%s-s3alias", ap.Name, accountID[:12])

	if ap.VpcConfiguration != nil {
		ap.NetworkOrigin = "VPC"
	} else {
		ap.NetworkOrigin = "Internet"
	}

	ap.Endpoints = map[string]string{
		"https": fmt.Sprintf("https://%s-%s.s3-accesspoint.us-east-1.amazonaws.com", ap.Name, accountID),
	}

	s.accessPoints[accountID][ap.Name] = ap

	return ap, nil
}

// GetAccessPoint retrieves an access point.
func (s *MemoryStorage) GetAccessPoint(_ context.Context, accountID, name string) (*AccessPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accountAPs, exists := s.accessPoints[accountID]
	if !exists {
		return nil, &Error{
			Code:    ErrNoSuchAccessPoint,
			Message: fmt.Sprintf("Access point %s does not exist", name),
		}
	}

	ap, exists := accountAPs[name]
	if !exists {
		return nil, &Error{
			Code:    ErrNoSuchAccessPoint,
			Message: fmt.Sprintf("Access point %s does not exist", name),
		}
	}

	return ap, nil
}

// DeleteAccessPoint deletes an access point.
func (s *MemoryStorage) DeleteAccessPoint(_ context.Context, accountID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountAPs, exists := s.accessPoints[accountID]
	if !exists {
		return &Error{
			Code:    ErrNoSuchAccessPoint,
			Message: fmt.Sprintf("Access point %s does not exist", name),
		}
	}

	if _, exists := accountAPs[name]; !exists {
		return &Error{
			Code:    ErrNoSuchAccessPoint,
			Message: fmt.Sprintf("Access point %s does not exist", name),
		}
	}

	delete(accountAPs, name)

	return nil
}

// ListAccessPoints lists access points for an account.
func (s *MemoryStorage) ListAccessPoints(_ context.Context, accountID, bucket string, maxResults int, nextToken string) ([]*AccessPoint, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 1000
	}

	accountAPs, exists := s.accessPoints[accountID]
	if !exists {
		return []*AccessPoint{}, "", nil
	}

	var accessPoints []*AccessPoint

	for _, ap := range accountAPs {
		if bucket != "" && ap.Bucket != bucket {
			continue
		}

		accessPoints = append(accessPoints, ap)
	}

	// Simple pagination (no sorting for simplicity)
	start := 0

	if nextToken != "" {
		for i, ap := range accessPoints {
			if ap.Name == nextToken {
				start = i

				break
			}
		}
	}

	end := min(start+maxResults, len(accessPoints))

	result := accessPoints[start:end]
	newNextToken := ""

	if end < len(accessPoints) {
		newNextToken = accessPoints[end].Name
	}

	return result, newNextToken, nil
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}
