package eks

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	statusCreating = "CREATING"
	statusActive   = "ACTIVE"
	statusDeleting = "DELETING"

	defaultKubernetesVersion = "1.29"
	defaultPlatformVersion   = "eks.1"
)

// Storage defines the EKS storage interface.
type Storage interface {
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*Cluster, error)
	DeleteCluster(ctx context.Context, name string) (*Cluster, error)
	DescribeCluster(ctx context.Context, name string) (*Cluster, error)
	ListClusters(ctx context.Context, maxResults int, nextToken string) ([]string, string, error)
	CreateNodegroup(ctx context.Context, req *CreateNodegroupRequest) (*Nodegroup, error)
	DeleteNodegroup(ctx context.Context, clusterName, nodegroupName string) (*Nodegroup, error)
	DescribeNodegroup(ctx context.Context, clusterName, nodegroupName string) (*Nodegroup, error)
	ListNodegroups(ctx context.Context, clusterName string, maxResults int, nextToken string) ([]string, string, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu         sync.RWMutex
	clusters   map[string]*Cluster
	nodegroups map[string]map[string]*Nodegroup // clusterName -> nodegroupName -> Nodegroup
	region     string
	accountID  string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		clusters:   make(map[string]*Cluster),
		nodegroups: make(map[string]map[string]*Nodegroup),
		region:     "us-east-1",
		accountID:  "123456789012",
	}
}

// CreateCluster creates a new EKS cluster.
func (s *MemoryStorage) CreateCluster(_ context.Context, req *CreateClusterRequest) (*Cluster, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[req.Name]; exists {
		return nil, &Error{
			Code:    "ResourceInUseException",
			Message: fmt.Sprintf("Cluster already exists with name: %s", req.Name),
		}
	}

	now := time.Now()
	version := req.Version
	if version == "" {
		version = defaultKubernetesVersion
	}

	clusterArn := fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", s.region, s.accountID, req.Name)
	endpoint := fmt.Sprintf("https://%s.gr7.%s.eks.amazonaws.com", uuid.New().String()[:8], s.region)
	oidcIssuer := fmt.Sprintf("https://oidc.eks.%s.amazonaws.com/id/%s", s.region, uuid.New().String()[:32])

	// Generate fake certificate authority data
	caData := base64.StdEncoding.EncodeToString([]byte("fake-certificate-authority-data"))

	var vpcConfig *VpcConfigResponse
	if req.ResourcesVpcConfig != nil {
		endpointPublicAccess := true
		endpointPrivateAccess := false
		if req.ResourcesVpcConfig.EndpointPublicAccess != nil {
			endpointPublicAccess = *req.ResourcesVpcConfig.EndpointPublicAccess
		}
		if req.ResourcesVpcConfig.EndpointPrivateAccess != nil {
			endpointPrivateAccess = *req.ResourcesVpcConfig.EndpointPrivateAccess
		}

		vpcConfig = &VpcConfigResponse{
			SubnetIDs:              req.ResourcesVpcConfig.SubnetIDs,
			SecurityGroupIDs:       req.ResourcesVpcConfig.SecurityGroupIDs,
			ClusterSecurityGroupID: fmt.Sprintf("sg-%s", uuid.New().String()[:17]),
			VpcID:                  fmt.Sprintf("vpc-%s", uuid.New().String()[:17]),
			EndpointPublicAccess:   endpointPublicAccess,
			EndpointPrivateAccess:  endpointPrivateAccess,
			PublicAccessCidrs:      req.ResourcesVpcConfig.PublicAccessCidrs,
		}
		if len(vpcConfig.PublicAccessCidrs) == 0 {
			vpcConfig.PublicAccessCidrs = []string{"0.0.0.0/0"}
		}
	}

	cluster := &Cluster{
		Name:               req.Name,
		Arn:                clusterArn,
		CreatedAt:          &now,
		Version:            version,
		Endpoint:           endpoint,
		RoleArn:            req.RoleArn,
		ResourcesVpcConfig: vpcConfig,
		KubernetesNetworkConfig: &KubernetesNetworkConfig{
			ServiceIpv4Cidr: "10.100.0.0/16",
			IpFamily:        "ipv4",
		},
		Identity: &Identity{
			Oidc: &OIDC{
				Issuer: oidcIssuer,
			},
		},
		Status: statusCreating,
		CertificateAuthority: &Certificate{
			Data: caData,
		},
		PlatformVersion:  defaultPlatformVersion,
		Tags:             req.Tags,
		EncryptionConfig: req.EncryptionConfig,
		Health: &ClusterHealth{
			Issues: []ClusterIssue{},
		},
	}

	if req.Logging != nil {
		cluster.Logging = req.Logging
	}

	if req.AccessConfig != nil {
		cluster.AccessConfig = &AccessConfigResponse{
			BootstrapClusterCreatorAdminPermissions: req.AccessConfig.BootstrapClusterCreatorAdminPermissions,
			AuthenticationMode:                      req.AccessConfig.AuthenticationMode,
		}
	}

	// Set status to ACTIVE immediately for testing purposes.
	cluster.Status = statusActive

	s.clusters[req.Name] = cluster
	s.nodegroups[req.Name] = make(map[string]*Nodegroup)

	return cluster, nil
}

// DeleteCluster deletes an EKS cluster.
func (s *MemoryStorage) DeleteCluster(_ context.Context, name string) (*Cluster, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cluster, exists := s.clusters[name]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No cluster found for name: %s", name),
		}
	}

	// Check if there are any nodegroups
	if nodegroups, ok := s.nodegroups[name]; ok && len(nodegroups) > 0 {
		return nil, &Error{
			Code:    "ResourceInUseException",
			Message: "Cluster has nodegroups attached",
		}
	}

	cluster.Status = statusDeleting
	delete(s.clusters, name)
	delete(s.nodegroups, name)

	return cluster, nil
}

// DescribeCluster describes an EKS cluster.
func (s *MemoryStorage) DescribeCluster(_ context.Context, name string) (*Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cluster, exists := s.clusters[name]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No cluster found for name: %s", name),
		}
	}

	return cluster, nil
}

// ListClusters lists all EKS clusters.
func (s *MemoryStorage) ListClusters(_ context.Context, _ int, _ string) ([]string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.clusters))
	for name := range s.clusters {
		names = append(names, name)
	}

	return names, "", nil
}

// CreateNodegroup creates a new EKS node group.
func (s *MemoryStorage) CreateNodegroup(_ context.Context, req *CreateNodegroupRequest) (*Nodegroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cluster, exists := s.clusters[req.ClusterName]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No cluster found for name: %s", req.ClusterName),
		}
	}

	if _, exists := s.nodegroups[req.ClusterName][req.NodegroupName]; exists {
		return nil, &Error{
			Code:    "ResourceInUseException",
			Message: fmt.Sprintf("Nodegroup already exists with name: %s", req.NodegroupName),
		}
	}

	now := time.Now()
	nodegroupArn := fmt.Sprintf("arn:aws:eks:%s:%s:nodegroup/%s/%s/%s",
		s.region, s.accountID, req.ClusterName, req.NodegroupName, uuid.New().String()[:8])

	instanceTypes := req.InstanceTypes
	if len(instanceTypes) == 0 {
		instanceTypes = []string{"t3.medium"}
	}

	capacityType := req.CapacityType
	if capacityType == "" {
		capacityType = "ON_DEMAND"
	}

	amiType := req.AmiType
	if amiType == "" {
		amiType = "AL2_x86_64"
	}

	scalingConfig := req.ScalingConfig
	if scalingConfig == nil {
		minSize := 1
		maxSize := 2
		desiredSize := 1
		scalingConfig = &NodegroupScaling{
			MinSize:     &minSize,
			MaxSize:     &maxSize,
			DesiredSize: &desiredSize,
		}
	}

	nodegroup := &Nodegroup{
		NodegroupName:  req.NodegroupName,
		NodegroupArn:   nodegroupArn,
		ClusterName:    req.ClusterName,
		Version:        cluster.Version,
		ReleaseVersion: fmt.Sprintf("%s-20231116", cluster.Version),
		CreatedAt:      &now,
		ModifiedAt:     &now,
		Status:         statusCreating,
		CapacityType:   capacityType,
		ScalingConfig:  scalingConfig,
		InstanceTypes:  instanceTypes,
		Subnets:        req.Subnets,
		RemoteAccess:   req.RemoteAccess,
		AmiType:        amiType,
		NodeRole:       req.NodeRole,
		Labels:         req.Labels,
		Taints:         req.Taints,
		DiskSize:       req.DiskSize,
		UpdateConfig:   req.UpdateConfig,
		LaunchTemplate: req.LaunchTemplate,
		Tags:           req.Tags,
		Resources: &NodegroupResources{
			AutoScalingGroups: []AutoScalingGroup{
				{Name: fmt.Sprintf("eks-%s-%s", req.NodegroupName, uuid.New().String()[:8])},
			},
		},
		Health: &NodegroupHealth{
			Issues: []Issue{},
		},
	}

	// Set status to ACTIVE immediately for testing purposes.
	nodegroup.Status = statusActive

	s.nodegroups[req.ClusterName][req.NodegroupName] = nodegroup

	return nodegroup, nil
}

// DeleteNodegroup deletes an EKS node group.
func (s *MemoryStorage) DeleteNodegroup(_ context.Context, clusterName, nodegroupName string) (*Nodegroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[clusterName]; !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No cluster found for name: %s", clusterName),
		}
	}

	nodegroups, exists := s.nodegroups[clusterName]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No nodegroup found for name: %s", nodegroupName),
		}
	}

	nodegroup, exists := nodegroups[nodegroupName]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No nodegroup found for name: %s", nodegroupName),
		}
	}

	nodegroup.Status = statusDeleting
	delete(s.nodegroups[clusterName], nodegroupName)

	return nodegroup, nil
}

// DescribeNodegroup describes an EKS node group.
func (s *MemoryStorage) DescribeNodegroup(_ context.Context, clusterName, nodegroupName string) (*Nodegroup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.clusters[clusterName]; !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No cluster found for name: %s", clusterName),
		}
	}

	nodegroups, exists := s.nodegroups[clusterName]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No nodegroup found for name: %s", nodegroupName),
		}
	}

	nodegroup, exists := nodegroups[nodegroupName]
	if !exists {
		return nil, &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No nodegroup found for name: %s", nodegroupName),
		}
	}

	return nodegroup, nil
}

// ListNodegroups lists all EKS node groups for a cluster.
func (s *MemoryStorage) ListNodegroups(_ context.Context, clusterName string, _ int, _ string) ([]string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.clusters[clusterName]; !exists {
		return nil, "", &Error{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("No cluster found for name: %s", clusterName),
		}
	}

	nodegroups, exists := s.nodegroups[clusterName]
	if !exists {
		return []string{}, "", nil
	}

	names := make([]string, 0, len(nodegroups))
	for name := range nodegroups {
		names = append(names, name)
	}

	return names, "", nil
}
