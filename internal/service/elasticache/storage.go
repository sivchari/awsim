package elasticache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAccountID = "000000000000"
	defaultRegion    = "us-east-1"
)

// Storage defines the ElastiCache storage interface.
type Storage interface {
	CreateCacheCluster(ctx context.Context, input *CreateCacheClusterInput) (*CacheCluster, error)
	DeleteCacheCluster(ctx context.Context, clusterID string) (*CacheCluster, error)
	DescribeCacheClusters(ctx context.Context, clusterID string, showNodeInfo bool) ([]CacheCluster, error)
	ModifyCacheCluster(ctx context.Context, input *ModifyCacheClusterInput) (*CacheCluster, error)
	CreateReplicationGroup(ctx context.Context, input *CreateReplicationGroupInput) (*ReplicationGroup, error)
	DeleteReplicationGroup(ctx context.Context, groupID string) (*ReplicationGroup, error)
	DescribeReplicationGroups(ctx context.Context, groupID string) ([]ReplicationGroup, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu                sync.RWMutex
	cacheClusters     map[string]*CacheCluster
	replicationGroups map[string]*ReplicationGroup
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		cacheClusters:     make(map[string]*CacheCluster),
		replicationGroups: make(map[string]*ReplicationGroup),
	}
}

// CreateCacheCluster creates a new cache cluster.
func (m *MemoryStorage) CreateCacheCluster(_ context.Context, input *CreateCacheClusterInput) (*CacheCluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.cacheClusters[input.CacheClusterID]; exists {
		return nil, &Error{
			Code:    errCacheClusterAlreadyExists,
			Message: fmt.Sprintf("Cache cluster already exists: %s", input.CacheClusterID),
		}
	}

	cluster := m.buildCacheCluster(input)
	m.cacheClusters[input.CacheClusterID] = cluster

	return cluster, nil
}

func (m *MemoryStorage) buildCacheCluster(input *CreateCacheClusterInput) *CacheCluster {
	numNodes := input.NumCacheNodes
	if numNodes == 0 {
		numNodes = 1
	}

	az := input.PreferredAvailabilityZone
	if az == "" {
		az = defaultRegion + "a"
	}

	port := input.Port
	if port == 0 {
		port = m.getDefaultPort(input.Engine)
	}

	now := time.Now()

	cluster := &CacheCluster{
		CacheClusterID:             input.CacheClusterID,
		CacheClusterStatus:         CacheClusterStatusAvailable,
		CacheNodeType:              input.CacheNodeType,
		Engine:                     input.Engine,
		EngineVersion:              input.EngineVersion,
		NumCacheNodes:              numNodes,
		PreferredAvailabilityZone:  az,
		CacheClusterCreateTime:     now,
		PreferredMaintenanceWindow: input.PreferredMaintenanceWindow,
		CacheSubnetGroupName:       input.CacheSubnetGroupName,
		AutoMinorVersionUpgrade:    input.AutoMinorVersionUpgrade,
		SnapshotRetentionLimit:     input.SnapshotRetentionLimit,
		SnapshotWindow:             input.SnapshotWindow,
		ARN:                        m.cacheClusterArn(input.CacheClusterID),
		CacheNodes:                 m.buildCacheNodes(numNodes, az, port, now),
		SecurityGroups:             buildSecurityGroups(input.SecurityGroupIDs),
		ConfigurationEndpoint: &Endpoint{
			Address: fmt.Sprintf("%s.%s.cfg.%s.cache.amazonaws.com", input.CacheClusterID, generateID(), defaultRegion),
			Port:    port,
		},
	}

	return cluster
}

func (m *MemoryStorage) buildCacheNodes(numNodes int32, az string, port int32, createTime time.Time) []CacheNode {
	nodes := make([]CacheNode, 0, numNodes)

	for i := range numNodes {
		nodeID := fmt.Sprintf("%04d", i+1)
		nodes = append(nodes, CacheNode{
			CacheNodeID:              nodeID,
			CacheNodeStatus:          CacheClusterStatusAvailable,
			CacheNodeCreateTime:      createTime,
			CustomerAvailabilityZone: az,
			ParameterGroupStatus:     "in-sync",
			Endpoint: &Endpoint{
				Address: fmt.Sprintf("%s.%s.%s.cache.amazonaws.com", nodeID, generateID(), defaultRegion),
				Port:    port,
			},
		})
	}

	return nodes
}

// DeleteCacheCluster deletes a cache cluster.
func (m *MemoryStorage) DeleteCacheCluster(_ context.Context, clusterID string) (*CacheCluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, exists := m.cacheClusters[clusterID]
	if !exists {
		return nil, &Error{
			Code:    errCacheClusterNotFound,
			Message: fmt.Sprintf("Cache cluster not found: %s", clusterID),
		}
	}

	cluster.CacheClusterStatus = CacheClusterStatusDeleting

	delete(m.cacheClusters, clusterID)

	return cluster, nil
}

// DescribeCacheClusters describes cache clusters.
func (m *MemoryStorage) DescribeCacheClusters(_ context.Context, clusterID string, showNodeInfo bool) ([]CacheCluster, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clusterID != "" {
		cluster, exists := m.cacheClusters[clusterID]
		if !exists {
			return nil, &Error{
				Code:    errCacheClusterNotFound,
				Message: fmt.Sprintf("Cache cluster not found: %s", clusterID),
			}
		}

		result := *cluster
		if !showNodeInfo {
			result.CacheNodes = nil
		}

		return []CacheCluster{result}, nil
	}

	clusters := make([]CacheCluster, 0, len(m.cacheClusters))

	for _, cluster := range m.cacheClusters {
		result := *cluster
		if !showNodeInfo {
			result.CacheNodes = nil
		}

		clusters = append(clusters, result)
	}

	return clusters, nil
}

// ModifyCacheCluster modifies a cache cluster.
func (m *MemoryStorage) ModifyCacheCluster(_ context.Context, input *ModifyCacheClusterInput) (*CacheCluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, exists := m.cacheClusters[input.CacheClusterID]
	if !exists {
		return nil, &Error{
			Code:    errCacheClusterNotFound,
			Message: fmt.Sprintf("Cache cluster not found: %s", input.CacheClusterID),
		}
	}

	applyCacheClusterModifications(cluster, input)

	return cluster, nil
}

func applyCacheClusterModifications(cluster *CacheCluster, input *ModifyCacheClusterInput) {
	if input.CacheNodeType != "" {
		cluster.CacheNodeType = input.CacheNodeType
	}

	if input.EngineVersion != "" {
		cluster.EngineVersion = input.EngineVersion
	}

	if input.NumCacheNodes != nil {
		cluster.NumCacheNodes = *input.NumCacheNodes
	}

	if input.PreferredMaintenanceWindow != "" {
		cluster.PreferredMaintenanceWindow = input.PreferredMaintenanceWindow
	}

	if input.AutoMinorVersionUpgrade != nil {
		cluster.AutoMinorVersionUpgrade = *input.AutoMinorVersionUpgrade
	}

	if input.SnapshotRetentionLimit != nil {
		cluster.SnapshotRetentionLimit = *input.SnapshotRetentionLimit
	}

	if input.SnapshotWindow != "" {
		cluster.SnapshotWindow = input.SnapshotWindow
	}

	if len(input.SecurityGroupIDs) > 0 {
		cluster.SecurityGroups = buildSecurityGroups(input.SecurityGroupIDs)
	}
}

// CreateReplicationGroup creates a new replication group.
func (m *MemoryStorage) CreateReplicationGroup(_ context.Context, input *CreateReplicationGroupInput) (*ReplicationGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.replicationGroups[input.ReplicationGroupID]; exists {
		return nil, &Error{
			Code:    errReplicationGroupAlreadyExists,
			Message: fmt.Sprintf("Replication group already exists: %s", input.ReplicationGroupID),
		}
	}

	group := m.buildReplicationGroup(input)
	m.replicationGroups[input.ReplicationGroupID] = group

	return group, nil
}

func (m *MemoryStorage) buildReplicationGroup(input *CreateReplicationGroupInput) *ReplicationGroup {
	port := input.Port
	if port == 0 {
		port = m.getDefaultPort(input.Engine)
	}

	now := time.Now()

	group := &ReplicationGroup{
		ReplicationGroupID:         input.ReplicationGroupID,
		Description:                input.ReplicationGroupDescription,
		Status:                     ReplicationGroupStatusAvailable,
		CacheNodeType:              input.CacheNodeType,
		AutomaticFailover:          automaticFailoverStatus(input.AutomaticFailoverEnabled),
		MultiAZ:                    multiAZStatus(input.MultiAZEnabled),
		SnapshotRetentionLimit:     input.SnapshotRetentionLimit,
		SnapshotWindow:             input.SnapshotWindow,
		ClusterEnabled:             input.NumNodeGroups > 0,
		TransitEncryptionEnabled:   input.TransitEncryptionEnabled,
		AtRestEncryptionEnabled:    input.AtRestEncryptionEnabled,
		ARN:                        m.replicationGroupArn(input.ReplicationGroupID),
		ReplicationGroupCreateTime: now,
		AutoMinorVersionUpgrade:    input.AutoMinorVersionUpgrade,
		PreferredMaintenanceWindow: input.PreferredMaintenanceWindow,
		ConfigurationEndpoint: &Endpoint{
			Address: fmt.Sprintf("%s.%s.clustercfg.%s.cache.amazonaws.com", input.ReplicationGroupID, generateID(), defaultRegion),
			Port:    port,
		},
		NodeGroups: m.buildNodeGroups(input, port),
	}

	return group
}

func (m *MemoryStorage) buildNodeGroups(input *CreateReplicationGroupInput, port int32) []NodeGroup {
	numGroups := input.NumNodeGroups
	if numGroups == 0 {
		numGroups = 1
	}

	replicas := input.ReplicasPerNodeGroup

	groups := make([]NodeGroup, 0, numGroups)

	for i := range numGroups {
		groupID := fmt.Sprintf("%04d", i+1)
		nodeGroup := NodeGroup{
			NodeGroupID: groupID,
			Status:      ReplicationGroupStatusAvailable,
			PrimaryEndpoint: &Endpoint{
				Address: fmt.Sprintf("%s-%s.%s.%s.cache.amazonaws.com", input.ReplicationGroupID, groupID, generateID(), defaultRegion),
				Port:    port,
			},
			ReaderEndpoint: &Endpoint{
				Address: fmt.Sprintf("%s-%s-ro.%s.%s.cache.amazonaws.com", input.ReplicationGroupID, groupID, generateID(), defaultRegion),
				Port:    port,
			},
			NodeGroupMembers: m.buildNodeGroupMembers(input.ReplicationGroupID, groupID, replicas, port),
		}
		groups = append(groups, nodeGroup)
	}

	return groups
}

func (m *MemoryStorage) buildNodeGroupMembers(rgID, ngID string, replicas, port int32) []NodeGroupMember {
	// Primary + replicas
	totalNodes := 1 + replicas
	members := make([]NodeGroupMember, 0, totalNodes)

	for i := range totalNodes {
		role := "replica"
		if i == 0 {
			role = "primary"
		}

		nodeID := fmt.Sprintf("%04d", i+1)
		clusterID := fmt.Sprintf("%s-%s-%s", rgID, ngID, nodeID)

		members = append(members, NodeGroupMember{
			CacheClusterID:            clusterID,
			CacheNodeID:               nodeID,
			PreferredAvailabilityZone: defaultRegion + "a",
			CurrentRole:               role,
			ReadEndpoint: &Endpoint{
				Address: fmt.Sprintf("%s.%s.%s.cache.amazonaws.com", clusterID, generateID(), defaultRegion),
				Port:    port,
			},
		})
	}

	return members
}

// DeleteReplicationGroup deletes a replication group.
func (m *MemoryStorage) DeleteReplicationGroup(_ context.Context, groupID string) (*ReplicationGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, exists := m.replicationGroups[groupID]
	if !exists {
		return nil, &Error{
			Code:    errReplicationGroupNotFound,
			Message: fmt.Sprintf("Replication group not found: %s", groupID),
		}
	}

	group.Status = ReplicationGroupStatusDeleting

	delete(m.replicationGroups, groupID)

	return group, nil
}

// DescribeReplicationGroups describes replication groups.
func (m *MemoryStorage) DescribeReplicationGroups(_ context.Context, groupID string) ([]ReplicationGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupID != "" {
		group, exists := m.replicationGroups[groupID]
		if !exists {
			return nil, &Error{
				Code:    errReplicationGroupNotFound,
				Message: fmt.Sprintf("Replication group not found: %s", groupID),
			}
		}

		return []ReplicationGroup{*group}, nil
	}

	groups := make([]ReplicationGroup, 0, len(m.replicationGroups))
	for _, group := range m.replicationGroups {
		groups = append(groups, *group)
	}

	return groups, nil
}

// Helper functions.

func (m *MemoryStorage) cacheClusterArn(clusterID string) string {
	return fmt.Sprintf("arn:aws:elasticache:%s:%s:cluster:%s", defaultRegion, defaultAccountID, clusterID)
}

func (m *MemoryStorage) replicationGroupArn(groupID string) string {
	return fmt.Sprintf("arn:aws:elasticache:%s:%s:replicationgroup:%s", defaultRegion, defaultAccountID, groupID)
}

func (m *MemoryStorage) getDefaultPort(engine string) int32 {
	switch engine {
	case "redis", "valkey":
		return 6379
	case "memcached":
		return 11211
	default:
		return 6379
	}
}

func generateID() string {
	return uuid.New().String()[:8]
}

func buildSecurityGroups(sgIDs []string) []SecurityGroupMembership {
	if len(sgIDs) == 0 {
		return nil
	}

	groups := make([]SecurityGroupMembership, 0, len(sgIDs))
	for _, sgID := range sgIDs {
		groups = append(groups, SecurityGroupMembership{
			SecurityGroupID: sgID,
			Status:          "active",
		})
	}

	return groups
}

func automaticFailoverStatus(enabled bool) string {
	if enabled {
		return "enabled"
	}

	return "disabled"
}

func multiAZStatus(enabled bool) string {
	if enabled {
		return "enabled"
	}

	return "disabled"
}
