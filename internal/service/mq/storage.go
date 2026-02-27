package mq

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the MQ storage interface.
type Storage interface {
	CreateBroker(ctx context.Context, req *CreateBrokerRequest) (*Broker, error)
	DeleteBroker(ctx context.Context, brokerID string) error
	DescribeBroker(ctx context.Context, brokerID string) (*Broker, error)
	ListBrokers(ctx context.Context, maxResults int, nextToken string) ([]*Broker, string, error)
	UpdateBroker(ctx context.Context, brokerID string, req *UpdateBrokerRequest) (*Broker, error)
	CreateConfiguration(ctx context.Context, req *CreateConfigurationRequest) (*Configuration, error)
	GetConfiguration(ctx context.Context, configID string) (*Configuration, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu             sync.RWMutex
	brokers        map[string]*Broker
	configurations map[string]*Configuration
	region         string
	accountID      string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		brokers:        make(map[string]*Broker),
		configurations: make(map[string]*Configuration),
		region:         "us-east-1",
		accountID:      "123456789012",
	}
}

// CreateBroker creates a new broker.
func (s *MemoryStorage) CreateBroker(_ context.Context, req *CreateBrokerRequest) (*Broker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate broker name
	for _, b := range s.brokers {
		if b.BrokerName == req.BrokerName {
			return nil, &Error{
				Type:    ErrConflict,
				Message: fmt.Sprintf("Broker with name %s already exists", req.BrokerName),
			}
		}
	}

	brokerID := uuid.New().String()
	brokerArn := fmt.Sprintf("arn:aws:mq:%s:%s:broker:%s:%s", s.region, s.accountID, req.BrokerName, brokerID)

	users := make([]*User, len(req.Users))
	for i, u := range req.Users {
		users[i] = &User{
			Username: u.Username,
			Password: u.Password,
			Groups:   u.Groups,
		}
	}

	broker := &Broker{
		BrokerID:             brokerID,
		BrokerName:           req.BrokerName,
		BrokerArn:            brokerArn,
		BrokerState:          BrokerStateRunning, // Immediately running for emulator
		Created:              time.Now().UTC(),
		DeploymentMode:       req.DeploymentMode,
		EngineType:           req.EngineType,
		EngineVersion:        req.EngineVersion,
		HostInstanceType:     req.HostInstanceType,
		AutoMinorVersionUpgr: req.AutoMinorVersionUpgr,
		PubliclyAccessible:   req.PubliclyAccessible,
		Users:                users,
		Tags:                 req.Tags,
		Configuration:        req.Configuration,
	}

	s.brokers[brokerID] = broker

	return broker, nil
}

// DeleteBroker deletes a broker.
func (s *MemoryStorage) DeleteBroker(_ context.Context, brokerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.brokers[brokerID]; !exists {
		return &Error{
			Type:    ErrNotFound,
			Message: fmt.Sprintf("Broker %s not found", brokerID),
		}
	}

	delete(s.brokers, brokerID)

	return nil
}

// DescribeBroker retrieves a broker by ID.
func (s *MemoryStorage) DescribeBroker(_ context.Context, brokerID string) (*Broker, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	broker, exists := s.brokers[brokerID]
	if !exists {
		return nil, &Error{
			Type:    ErrNotFound,
			Message: fmt.Sprintf("Broker %s not found", brokerID),
		}
	}

	return broker, nil
}

// ListBrokers lists all brokers with pagination.
func (s *MemoryStorage) ListBrokers(_ context.Context, maxResults int, nextToken string) ([]*Broker, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 100
	}

	// Collect all brokers
	brokers := make([]*Broker, 0, len(s.brokers))
	for _, b := range s.brokers {
		brokers = append(brokers, b)
	}

	// Sort by broker name for consistent pagination
	sort.Slice(brokers, func(i, j int) bool {
		return brokers[i].BrokerName < brokers[j].BrokerName
	})

	// Handle pagination
	start := 0

	if nextToken != "" {
		for i, b := range brokers {
			if b.BrokerID == nextToken {
				start = i

				break
			}
		}
	}

	end := min(start+maxResults, len(brokers))

	result := brokers[start:end]
	newNextToken := ""

	if end < len(brokers) {
		newNextToken = brokers[end].BrokerID
	}

	return result, newNextToken, nil
}

// UpdateBroker updates a broker.
func (s *MemoryStorage) UpdateBroker(_ context.Context, brokerID string, req *UpdateBrokerRequest) (*Broker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	broker, exists := s.brokers[brokerID]
	if !exists {
		return nil, &Error{
			Type:    ErrNotFound,
			Message: fmt.Sprintf("Broker %s not found", brokerID),
		}
	}

	if req.EngineVersion != "" {
		broker.EngineVersion = req.EngineVersion
	}

	if req.HostInstanceType != "" {
		broker.HostInstanceType = req.HostInstanceType
	}

	if req.AutoMinorVersionUpgr != nil {
		broker.AutoMinorVersionUpgr = *req.AutoMinorVersionUpgr
	}

	if req.Configuration != nil {
		broker.Configuration = req.Configuration
	}

	return broker, nil
}

// CreateConfiguration creates a new configuration.
func (s *MemoryStorage) CreateConfiguration(_ context.Context, req *CreateConfigurationRequest) (*Configuration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate configuration name
	for _, c := range s.configurations {
		if c.Name == req.Name {
			return nil, &Error{
				Type:    ErrConflict,
				Message: fmt.Sprintf("Configuration with name %s already exists", req.Name),
			}
		}
	}

	configID := "c-" + uuid.New().String()[:8]
	configArn := fmt.Sprintf("arn:aws:mq:%s:%s:configuration:%s:%s", s.region, s.accountID, configID, req.Name)

	now := time.Now().UTC()
	revision := &ConfigurationRevision{
		Revision:    1,
		Created:     now,
		Description: "Initial revision",
	}

	config := &Configuration{
		ID:             configID,
		Arn:            configArn,
		Name:           req.Name,
		EngineType:     req.EngineType,
		EngineVersion:  req.EngineVersion,
		Created:        now,
		LatestRevision: revision,
		Tags:           req.Tags,
		Revisions:      []*ConfigurationRevision{revision},
	}

	s.configurations[configID] = config

	return config, nil
}

// GetConfiguration retrieves a configuration by ID.
func (s *MemoryStorage) GetConfiguration(_ context.Context, configID string) (*Configuration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configurations[configID]
	if !exists {
		return nil, &Error{
			Type:    ErrNotFound,
			Message: fmt.Sprintf("Configuration %s not found", configID),
		}
	}

	return config, nil
}
