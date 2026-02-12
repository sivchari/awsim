package codeconnections

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes for CodeConnections.
const (
	errResourceNotFoundException = "ResourceNotFoundException"
	errConflictException         = "ConflictException"
)

// Storage defines the interface for CodeConnections storage.
type Storage interface {
	// Connection operations
	CreateConnection(ctx context.Context, name string, providerType string, hostArn string, tags []Tag) (*Connection, error)
	GetConnection(ctx context.Context, connectionArn string) (*Connection, error)
	DeleteConnection(ctx context.Context, connectionArn string) error
	ListConnections(ctx context.Context, providerTypeFilter, hostArnFilter, nextToken string, maxResults int32) ([]*Connection, string, error)

	// Host operations
	CreateHost(ctx context.Context, name, providerType, providerEndpoint string, vpcConfig *VpcConfiguration, tags []Tag) (*Host, error)
	GetHost(ctx context.Context, hostArn string) (*Host, error)
	DeleteHost(ctx context.Context, hostArn string) error
	ListHosts(ctx context.Context, nextToken string, maxResults int32) ([]*Host, string, error)
	UpdateHost(ctx context.Context, hostArn, providerEndpoint string, vpcConfig *VpcConfiguration) error

	// Repository link operations
	CreateRepositoryLink(ctx context.Context, connectionArn, ownerID, repositoryName, encryptionKeyArn string, tags []Tag) (*RepositoryLink, error)
	GetRepositoryLink(ctx context.Context, repositoryLinkID string) (*RepositoryLink, error)
	DeleteRepositoryLink(ctx context.Context, repositoryLinkID string) error
	ListRepositoryLinks(ctx context.Context, nextToken string, maxResults int32) ([]*RepositoryLink, string, error)
	UpdateRepositoryLink(ctx context.Context, repositoryLinkID, connectionArn, encryptionKeyArn string) (*RepositoryLink, error)

	// Tag operations
	ListTagsForResource(ctx context.Context, resourceArn string) ([]Tag, error)
	TagResource(ctx context.Context, resourceArn string, tags []Tag) error
	UntagResource(ctx context.Context, resourceArn string, tagKeys []string) error
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu              sync.RWMutex
	connections     map[string]*Connection
	hosts           map[string]*Host
	repositoryLinks map[string]*RepositoryLink
	accountID       string
	region          string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		connections:     make(map[string]*Connection),
		hosts:           make(map[string]*Host),
		repositoryLinks: make(map[string]*RepositoryLink),
		accountID:       "000000000000",
		region:          "us-east-1",
	}
}

// CreateConnection creates a new connection.
func (s *MemoryStorage) CreateConnection(_ context.Context, name, providerType, hostArn string, tags []Tag) (*Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	connectionID := uuid.New().String()
	connectionArn := fmt.Sprintf("arn:aws:codeconnections:%s:%s:connection/%s", s.region, s.accountID, connectionID)

	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}

	conn := &Connection{
		ConnectionArn:    connectionArn,
		ConnectionName:   name,
		ConnectionStatus: ConnectionStatusPending,
		OwnerAccountID:   s.accountID,
		ProviderType:     ProviderType(providerType),
		HostArn:          hostArn,
		CreatedAt:        time.Now(),
		Tags:             tagMap,
	}

	s.connections[connectionArn] = conn

	return conn, nil
}

// GetConnection retrieves a connection by ARN.
func (s *MemoryStorage) GetConnection(_ context.Context, connectionArn string) (*Connection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, ok := s.connections[connectionArn]
	if !ok {
		return nil, &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Connection %s not found.", connectionArn),
		}
	}

	return conn, nil
}

// DeleteConnection deletes a connection.
func (s *MemoryStorage) DeleteConnection(_ context.Context, connectionArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.connections[connectionArn]; !ok {
		return &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Connection %s not found.", connectionArn),
		}
	}

	delete(s.connections, connectionArn)

	return nil
}

// ListConnections lists connections with optional filters.
func (s *MemoryStorage) ListConnections(_ context.Context, providerTypeFilter, hostArnFilter, _ string, maxResults int32) ([]*Connection, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 50
	}

	connections := make([]*Connection, 0)

	for _, conn := range s.connections {
		if providerTypeFilter != "" && string(conn.ProviderType) != providerTypeFilter {
			continue
		}

		if hostArnFilter != "" && conn.HostArn != hostArnFilter {
			continue
		}

		connections = append(connections, conn)

		if len(connections) >= int(maxResults) {
			break
		}
	}

	return connections, "", nil
}

// CreateHost creates a new host.
func (s *MemoryStorage) CreateHost(_ context.Context, name, providerType, providerEndpoint string, vpcConfig *VpcConfiguration, tags []Tag) (*Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hostID := uuid.New().String()
	hostArn := fmt.Sprintf("arn:aws:codeconnections:%s:%s:host/%s", s.region, s.accountID, hostID)

	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}

	host := &Host{
		HostArn:          hostArn,
		Name:             name,
		Status:           "PENDING",
		ProviderType:     ProviderType(providerType),
		ProviderEndpoint: providerEndpoint,
		VpcConfiguration: vpcConfig,
		CreatedAt:        time.Now(),
		Tags:             tagMap,
	}

	s.hosts[hostArn] = host

	return host, nil
}

// GetHost retrieves a host by ARN.
func (s *MemoryStorage) GetHost(_ context.Context, hostArn string) (*Host, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	host, ok := s.hosts[hostArn]
	if !ok {
		return nil, &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Host %s not found.", hostArn),
		}
	}

	return host, nil
}

// DeleteHost deletes a host.
func (s *MemoryStorage) DeleteHost(_ context.Context, hostArn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.hosts[hostArn]; !ok {
		return &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Host %s not found.", hostArn),
		}
	}

	// Check if any connection uses this host
	for _, conn := range s.connections {
		if conn.HostArn == hostArn {
			return &ServiceError{
				Code:    errConflictException,
				Message: fmt.Sprintf("Host %s is in use by connection %s.", hostArn, conn.ConnectionArn),
			}
		}
	}

	delete(s.hosts, hostArn)

	return nil
}

// ListHosts lists all hosts.
func (s *MemoryStorage) ListHosts(_ context.Context, _ string, maxResults int32) ([]*Host, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 50
	}

	hosts := make([]*Host, 0)

	for _, host := range s.hosts {
		hosts = append(hosts, host)

		if len(hosts) >= int(maxResults) {
			break
		}
	}

	return hosts, "", nil
}

// UpdateHost updates a host.
func (s *MemoryStorage) UpdateHost(_ context.Context, hostArn, providerEndpoint string, vpcConfig *VpcConfiguration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	host, ok := s.hosts[hostArn]
	if !ok {
		return &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Host %s not found.", hostArn),
		}
	}

	if providerEndpoint != "" {
		host.ProviderEndpoint = providerEndpoint
	}

	if vpcConfig != nil {
		host.VpcConfiguration = vpcConfig
	}

	return nil
}

// CreateRepositoryLink creates a new repository link.
func (s *MemoryStorage) CreateRepositoryLink(_ context.Context, connectionArn, ownerID, repositoryName, encryptionKeyArn string, tags []Tag) (*RepositoryLink, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify connection exists
	conn, ok := s.connections[connectionArn]
	if !ok {
		return nil, &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Connection %s not found.", connectionArn),
		}
	}

	repositoryLinkID := uuid.New().String()
	repositoryLinkArn := fmt.Sprintf("arn:aws:codeconnections:%s:%s:repository-link/%s", s.region, s.accountID, repositoryLinkID)

	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}

	repoLink := &RepositoryLink{
		RepositoryLinkArn: repositoryLinkArn,
		RepositoryLinkID:  repositoryLinkID,
		ConnectionArn:     connectionArn,
		OwnerID:           ownerID,
		ProviderType:      conn.ProviderType,
		RepositoryName:    repositoryName,
		EncryptionKeyArn:  encryptionKeyArn,
		CreatedAt:         time.Now(),
		Tags:              tagMap,
	}

	s.repositoryLinks[repositoryLinkID] = repoLink

	return repoLink, nil
}

// GetRepositoryLink retrieves a repository link by ID.
func (s *MemoryStorage) GetRepositoryLink(_ context.Context, repositoryLinkID string) (*RepositoryLink, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoLink, ok := s.repositoryLinks[repositoryLinkID]
	if !ok {
		return nil, &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("RepositoryLink %s not found.", repositoryLinkID),
		}
	}

	return repoLink, nil
}

// DeleteRepositoryLink deletes a repository link.
func (s *MemoryStorage) DeleteRepositoryLink(_ context.Context, repositoryLinkID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.repositoryLinks[repositoryLinkID]; !ok {
		return &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("RepositoryLink %s not found.", repositoryLinkID),
		}
	}

	delete(s.repositoryLinks, repositoryLinkID)

	return nil
}

// ListRepositoryLinks lists all repository links.
func (s *MemoryStorage) ListRepositoryLinks(_ context.Context, _ string, maxResults int32) ([]*RepositoryLink, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 50
	}

	links := make([]*RepositoryLink, 0)

	for _, link := range s.repositoryLinks {
		links = append(links, link)

		if len(links) >= int(maxResults) {
			break
		}
	}

	return links, "", nil
}

// UpdateRepositoryLink updates a repository link.
func (s *MemoryStorage) UpdateRepositoryLink(_ context.Context, repositoryLinkID, connectionArn, encryptionKeyArn string) (*RepositoryLink, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoLink, ok := s.repositoryLinks[repositoryLinkID]
	if !ok {
		return nil, &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("RepositoryLink %s not found.", repositoryLinkID),
		}
	}

	if connectionArn != "" {
		// Verify new connection exists
		conn, ok := s.connections[connectionArn]
		if !ok {
			return nil, &ServiceError{
				Code:    errResourceNotFoundException,
				Message: fmt.Sprintf("Connection %s not found.", connectionArn),
			}
		}

		repoLink.ConnectionArn = connectionArn
		repoLink.ProviderType = conn.ProviderType
	}

	if encryptionKeyArn != "" {
		repoLink.EncryptionKeyArn = encryptionKeyArn
	}

	return repoLink, nil
}

// ListTagsForResource lists tags for a resource.
func (s *MemoryStorage) ListTagsForResource(_ context.Context, resourceArn string) ([]Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tagMap map[string]string

	// Check connections
	if conn, ok := s.connections[resourceArn]; ok {
		tagMap = conn.Tags
	} else if host, ok := s.hosts[resourceArn]; ok {
		tagMap = host.Tags
	} else {
		// Check repository links by ARN
		for _, link := range s.repositoryLinks {
			if link.RepositoryLinkArn == resourceArn {
				tagMap = link.Tags

				break
			}
		}
	}

	if tagMap == nil {
		return nil, &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Resource %s not found.", resourceArn),
		}
	}

	tags := make([]Tag, 0, len(tagMap))
	for k, v := range tagMap {
		tags = append(tags, Tag{Key: k, Value: v})
	}

	return tags, nil
}

// TagResource adds tags to a resource.
func (s *MemoryStorage) TagResource(_ context.Context, resourceArn string, tags []Tag) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var tagMap map[string]string

	// Check connections
	if conn, ok := s.connections[resourceArn]; ok {
		tagMap = conn.Tags
	} else if host, ok := s.hosts[resourceArn]; ok {
		tagMap = host.Tags
	} else {
		// Check repository links by ARN
		for _, link := range s.repositoryLinks {
			if link.RepositoryLinkArn == resourceArn {
				tagMap = link.Tags

				break
			}
		}
	}

	if tagMap == nil {
		return &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Resource %s not found.", resourceArn),
		}
	}

	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}

	return nil
}

// UntagResource removes tags from a resource.
func (s *MemoryStorage) UntagResource(_ context.Context, resourceArn string, tagKeys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var tagMap map[string]string

	// Check connections
	if conn, ok := s.connections[resourceArn]; ok {
		tagMap = conn.Tags
	} else if host, ok := s.hosts[resourceArn]; ok {
		tagMap = host.Tags
	} else {
		// Check repository links by ARN
		for _, link := range s.repositoryLinks {
			if link.RepositoryLinkArn == resourceArn {
				tagMap = link.Tags

				break
			}
		}
	}

	if tagMap == nil {
		return &ServiceError{
			Code:    errResourceNotFoundException,
			Message: fmt.Sprintf("Resource %s not found.", resourceArn),
		}
	}

	for _, key := range tagKeys {
		delete(tagMap, key)
	}

	return nil
}
