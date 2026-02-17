package rds

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

// Storage defines the RDS storage interface.
type Storage interface {
	CreateDBInstance(ctx context.Context, input *CreateDBInstanceInput) (*DBInstance, error)
	DeleteDBInstance(ctx context.Context, identifier string, skipFinalSnapshot bool) (*DBInstance, error)
	DescribeDBInstances(ctx context.Context, identifier string) ([]DBInstance, error)
	ModifyDBInstance(ctx context.Context, input *ModifyDBInstanceInput) (*DBInstance, error)
	StartDBInstance(ctx context.Context, identifier string) (*DBInstance, error)
	StopDBInstance(ctx context.Context, identifier string) (*DBInstance, error)
	CreateDBCluster(ctx context.Context, input *CreateDBClusterInput) (*DBCluster, error)
	DeleteDBCluster(ctx context.Context, identifier string, skipFinalSnapshot bool) (*DBCluster, error)
	DescribeDBClusters(ctx context.Context, identifier string) ([]DBCluster, error)
	CreateDBSnapshot(ctx context.Context, input *CreateDBSnapshotInput) (*DBSnapshot, error)
	DeleteDBSnapshot(ctx context.Context, identifier string) (*DBSnapshot, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu        sync.RWMutex
	instances map[string]*DBInstance
	clusters  map[string]*DBCluster
	snapshots map[string]*DBSnapshot
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		instances: make(map[string]*DBInstance),
		clusters:  make(map[string]*DBCluster),
		snapshots: make(map[string]*DBSnapshot),
	}
}

// CreateDBInstance creates a new DB instance.
func (m *MemoryStorage) CreateDBInstance(_ context.Context, input *CreateDBInstanceInput) (*DBInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[input.DBInstanceIdentifier]; exists {
		return nil, &Error{
			Code:    errDBInstanceAlreadyExists,
			Message: fmt.Sprintf("DB instance already exists: %s", input.DBInstanceIdentifier),
		}
	}

	now := time.Now()
	instance := &DBInstance{
		DBInstanceIdentifier:       input.DBInstanceIdentifier,
		DBInstanceClass:            input.DBInstanceClass,
		Engine:                     input.Engine,
		EngineVersion:              input.EngineVersion,
		DBInstanceStatus:           DBInstanceStatusAvailable,
		MasterUsername:             input.MasterUsername,
		DBName:                     input.DBName,
		AllocatedStorage:           input.AllocatedStorage,
		InstanceCreateTime:         now,
		DBInstanceArn:              m.dbInstanceArn(input.DBInstanceIdentifier),
		StorageType:                input.StorageType,
		MultiAZ:                    input.MultiAZ,
		AvailabilityZone:           input.AvailabilityZone,
		BackupRetentionPeriod:      input.BackupRetentionPeriod,
		PreferredBackupWindow:      input.PreferredBackupWindow,
		PreferredMaintenanceWindow: input.PreferredMaintenanceWindow,
		PubliclyAccessible:         input.PubliclyAccessible,
		StorageEncrypted:           input.StorageEncrypted,
		DeletionProtection:         input.DeletionProtection,
		Tags:                       input.Tags,
		Endpoint: &Endpoint{
			Address: fmt.Sprintf("%s.%s.%s.rds.amazonaws.com", input.DBInstanceIdentifier, generateID(), defaultRegion),
			Port:    m.getDefaultPort(input.Engine),
		},
	}

	if input.StorageType == "" {
		instance.StorageType = "gp2"
	}

	if input.AllocatedStorage == 0 {
		instance.AllocatedStorage = 20
	}

	if input.AvailabilityZone == "" {
		instance.AvailabilityZone = defaultRegion + "a"
	}

	if len(input.VpcSecurityGroupIds) > 0 {
		for _, sgID := range input.VpcSecurityGroupIds {
			instance.VpcSecurityGroups = append(instance.VpcSecurityGroups, VpcSecurityGroupMembership{
				VpcSecurityGroupID: sgID,
				Status:             "active",
			})
		}
	}

	m.instances[input.DBInstanceIdentifier] = instance

	return instance, nil
}

// DeleteDBInstance deletes a DB instance.
func (m *MemoryStorage) DeleteDBInstance(_ context.Context, identifier string, _ bool) (*DBInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[identifier]
	if !exists {
		return nil, &Error{
			Code:    errDBInstanceNotFound,
			Message: fmt.Sprintf("DB instance not found: %s", identifier),
		}
	}

	instance.DBInstanceStatus = DBInstanceStatusDeleting
	delete(m.instances, identifier)

	return instance, nil
}

// DescribeDBInstances describes DB instances.
func (m *MemoryStorage) DescribeDBInstances(_ context.Context, identifier string) ([]DBInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if identifier != "" {
		instance, exists := m.instances[identifier]
		if !exists {
			return nil, &Error{
				Code:    errDBInstanceNotFound,
				Message: fmt.Sprintf("DB instance not found: %s", identifier),
			}
		}

		return []DBInstance{*instance}, nil
	}

	instances := make([]DBInstance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, *instance)
	}

	return instances, nil
}

// ModifyDBInstance modifies a DB instance.
func (m *MemoryStorage) ModifyDBInstance(_ context.Context, input *ModifyDBInstanceInput) (*DBInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[input.DBInstanceIdentifier]
	if !exists {
		return nil, &Error{
			Code:    errDBInstanceNotFound,
			Message: fmt.Sprintf("DB instance not found: %s", input.DBInstanceIdentifier),
		}
	}

	if input.DBInstanceClass != "" {
		instance.DBInstanceClass = input.DBInstanceClass
	}

	if input.AllocatedStorage > 0 {
		instance.AllocatedStorage = input.AllocatedStorage
	}

	if input.BackupRetentionPeriod != nil {
		instance.BackupRetentionPeriod = *input.BackupRetentionPeriod
	}

	if input.PreferredBackupWindow != "" {
		instance.PreferredBackupWindow = input.PreferredBackupWindow
	}

	if input.PreferredMaintenanceWindow != "" {
		instance.PreferredMaintenanceWindow = input.PreferredMaintenanceWindow
	}

	if input.MultiAZ != nil {
		instance.MultiAZ = *input.MultiAZ
	}

	if input.EngineVersion != "" {
		instance.EngineVersion = input.EngineVersion
	}

	if input.StorageType != "" {
		instance.StorageType = input.StorageType
	}

	if input.PubliclyAccessible != nil {
		instance.PubliclyAccessible = *input.PubliclyAccessible
	}

	if input.DeletionProtection != nil {
		instance.DeletionProtection = *input.DeletionProtection
	}

	if len(input.VpcSecurityGroupIds) > 0 {
		instance.VpcSecurityGroups = make([]VpcSecurityGroupMembership, 0, len(input.VpcSecurityGroupIds))
		for _, sgID := range input.VpcSecurityGroupIds {
			instance.VpcSecurityGroups = append(instance.VpcSecurityGroups, VpcSecurityGroupMembership{
				VpcSecurityGroupID: sgID,
				Status:             "active",
			})
		}
	}

	return instance, nil
}

// StartDBInstance starts a stopped DB instance.
func (m *MemoryStorage) StartDBInstance(_ context.Context, identifier string) (*DBInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[identifier]
	if !exists {
		return nil, &Error{
			Code:    errDBInstanceNotFound,
			Message: fmt.Sprintf("DB instance not found: %s", identifier),
		}
	}

	if instance.DBInstanceStatus != DBInstanceStatusStopped {
		return nil, &Error{
			Code:    errInvalidDBInstanceState,
			Message: fmt.Sprintf("DB instance is not in stopped state: %s", identifier),
		}
	}

	instance.DBInstanceStatus = DBInstanceStatusAvailable

	return instance, nil
}

// StopDBInstance stops a running DB instance.
func (m *MemoryStorage) StopDBInstance(_ context.Context, identifier string) (*DBInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[identifier]
	if !exists {
		return nil, &Error{
			Code:    errDBInstanceNotFound,
			Message: fmt.Sprintf("DB instance not found: %s", identifier),
		}
	}

	if instance.DBInstanceStatus != DBInstanceStatusAvailable {
		return nil, &Error{
			Code:    errInvalidDBInstanceState,
			Message: fmt.Sprintf("DB instance is not in available state: %s", identifier),
		}
	}

	instance.DBInstanceStatus = DBInstanceStatusStopped

	return instance, nil
}

// CreateDBCluster creates a new DB cluster.
func (m *MemoryStorage) CreateDBCluster(_ context.Context, input *CreateDBClusterInput) (*DBCluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clusters[input.DBClusterIdentifier]; exists {
		return nil, &Error{
			Code:    errDBClusterAlreadyExists,
			Message: fmt.Sprintf("DB cluster already exists: %s", input.DBClusterIdentifier),
		}
	}

	now := time.Now()
	port := input.Port
	if port == 0 {
		port = m.getDefaultPort(input.Engine)
	}

	cluster := &DBCluster{
		DBClusterIdentifier: input.DBClusterIdentifier,
		DBClusterArn:        m.dbClusterArn(input.DBClusterIdentifier),
		Engine:              input.Engine,
		EngineVersion:       input.EngineVersion,
		Status:              DBClusterStatusAvailable,
		MasterUsername:      input.MasterUsername,
		DatabaseName:        input.DatabaseName,
		Endpoint:            fmt.Sprintf("%s.cluster-%s.%s.rds.amazonaws.com", input.DBClusterIdentifier, generateID(), defaultRegion),
		ReaderEndpoint:      fmt.Sprintf("%s.cluster-ro-%s.%s.rds.amazonaws.com", input.DBClusterIdentifier, generateID(), defaultRegion),
		Port:                port,
		AllocatedStorage:    input.AllocatedStorage,
		ClusterCreateTime:   now,
		AvailabilityZones:   input.AvailabilityZones,
		StorageEncrypted:    input.StorageEncrypted,
		DeletionProtection:  input.DeletionProtection,
		Tags:                input.Tags,
	}

	if len(input.AvailabilityZones) == 0 {
		cluster.AvailabilityZones = []string{defaultRegion + "a", defaultRegion + "b", defaultRegion + "c"}
	}

	if len(input.VpcSecurityGroupIds) > 0 {
		for _, sgID := range input.VpcSecurityGroupIds {
			cluster.VpcSecurityGroups = append(cluster.VpcSecurityGroups, VpcSecurityGroupMembership{
				VpcSecurityGroupID: sgID,
				Status:             "active",
			})
		}
	}

	m.clusters[input.DBClusterIdentifier] = cluster

	return cluster, nil
}

// DeleteDBCluster deletes a DB cluster.
func (m *MemoryStorage) DeleteDBCluster(_ context.Context, identifier string, _ bool) (*DBCluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, exists := m.clusters[identifier]
	if !exists {
		return nil, &Error{
			Code:    errDBClusterNotFound,
			Message: fmt.Sprintf("DB cluster not found: %s", identifier),
		}
	}

	cluster.Status = DBClusterStatusDeleting
	delete(m.clusters, identifier)

	return cluster, nil
}

// DescribeDBClusters describes DB clusters.
func (m *MemoryStorage) DescribeDBClusters(_ context.Context, identifier string) ([]DBCluster, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if identifier != "" {
		cluster, exists := m.clusters[identifier]
		if !exists {
			return nil, &Error{
				Code:    errDBClusterNotFound,
				Message: fmt.Sprintf("DB cluster not found: %s", identifier),
			}
		}

		return []DBCluster{*cluster}, nil
	}

	clusters := make([]DBCluster, 0, len(m.clusters))
	for _, cluster := range m.clusters {
		clusters = append(clusters, *cluster)
	}

	return clusters, nil
}

// CreateDBSnapshot creates a DB snapshot.
func (m *MemoryStorage) CreateDBSnapshot(_ context.Context, input *CreateDBSnapshotInput) (*DBSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.snapshots[input.DBSnapshotIdentifier]; exists {
		return nil, &Error{
			Code:    errDBSnapshotAlreadyExists,
			Message: fmt.Sprintf("DB snapshot already exists: %s", input.DBSnapshotIdentifier),
		}
	}

	instance, exists := m.instances[input.DBInstanceIdentifier]
	if !exists {
		return nil, &Error{
			Code:    errDBInstanceNotFound,
			Message: fmt.Sprintf("DB instance not found: %s", input.DBInstanceIdentifier),
		}
	}

	now := time.Now()
	snapshot := &DBSnapshot{
		DBSnapshotIdentifier: input.DBSnapshotIdentifier,
		DBSnapshotArn:        m.dbSnapshotArn(input.DBSnapshotIdentifier),
		DBInstanceIdentifier: input.DBInstanceIdentifier,
		Engine:               instance.Engine,
		EngineVersion:        instance.EngineVersion,
		Status:               DBSnapshotStatusAvailable,
		SnapshotType:         "manual",
		SnapshotCreateTime:   now,
		AllocatedStorage:     instance.AllocatedStorage,
		Port:                 instance.Endpoint.Port,
		AvailabilityZone:     instance.AvailabilityZone,
		MasterUsername:       instance.MasterUsername,
		StorageType:          instance.StorageType,
		Encrypted:            instance.StorageEncrypted,
		Tags:                 input.Tags,
	}

	m.snapshots[input.DBSnapshotIdentifier] = snapshot

	return snapshot, nil
}

// DeleteDBSnapshot deletes a DB snapshot.
func (m *MemoryStorage) DeleteDBSnapshot(_ context.Context, identifier string) (*DBSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot, exists := m.snapshots[identifier]
	if !exists {
		return nil, &Error{
			Code:    errDBSnapshotNotFound,
			Message: fmt.Sprintf("DB snapshot not found: %s", identifier),
		}
	}

	delete(m.snapshots, identifier)

	return snapshot, nil
}

// Helper functions.

func (m *MemoryStorage) dbInstanceArn(identifier string) string {
	return fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", defaultRegion, defaultAccountID, identifier)
}

func (m *MemoryStorage) dbClusterArn(identifier string) string {
	return fmt.Sprintf("arn:aws:rds:%s:%s:cluster:%s", defaultRegion, defaultAccountID, identifier)
}

func (m *MemoryStorage) dbSnapshotArn(identifier string) string {
	return fmt.Sprintf("arn:aws:rds:%s:%s:snapshot:%s", defaultRegion, defaultAccountID, identifier)
}

func (m *MemoryStorage) getDefaultPort(engine string) int32 {
	switch engine {
	case "mysql", "mariadb", "aurora", "aurora-mysql":
		return 3306
	case "postgres", "aurora-postgresql":
		return 5432
	case "oracle-ee", "oracle-se", "oracle-se1", "oracle-se2":
		return 1521
	case "sqlserver-ee", "sqlserver-se", "sqlserver-ex", "sqlserver-web":
		return 1433
	default:
		return 3306
	}
}

func generateID() string {
	return uuid.New().String()[:8]
}
