package eks

import "time"

// EpochTime is a custom type for serializing time.Time to Unix epoch seconds.
type EpochTime float64

// NewEpochTime creates an EpochTime from time.Time.
func NewEpochTime(t time.Time) EpochTime {
	return EpochTime(float64(t.UnixNano()) / 1e9)
}

// Cluster represents an EKS cluster.
type Cluster struct {
	Name                       string                     `json:"name"`
	Arn                        string                     `json:"arn"`
	CreatedAt                  *EpochTime                 `json:"createdAt,omitempty"`
	Version                    string                     `json:"version,omitempty"`
	Endpoint                   string                     `json:"endpoint,omitempty"`
	RoleArn                    string                     `json:"roleArn,omitempty"`
	ResourcesVpcConfig         *VpcConfigResponse         `json:"resourcesVpcConfig,omitempty"`
	KubernetesNetworkConfig    *KubernetesNetworkConfig   `json:"kubernetesNetworkConfig,omitempty"`
	Logging                    *Logging                   `json:"logging,omitempty"`
	Identity                   *Identity                  `json:"identity,omitempty"`
	Status                     string                     `json:"status"`
	CertificateAuthority       *Certificate               `json:"certificateAuthority,omitempty"`
	PlatformVersion            string                     `json:"platformVersion,omitempty"`
	Tags                       map[string]string          `json:"tags,omitempty"`
	EncryptionConfig           []EncryptionConfig         `json:"encryptionConfig,omitempty"`
	ConnectorConfig            *ConnectorConfig           `json:"connectorConfig,omitempty"`
	Health                     *ClusterHealth             `json:"health,omitempty"`
	OutpostConfig              *OutpostConfig             `json:"outpostConfig,omitempty"`
	AccessConfig               *AccessConfigResponse      `json:"accessConfig,omitempty"`
	UpgradePolicyResponse      *UpgradePolicyResponse     `json:"upgradePolicy,omitempty"`
	ZonalShiftConfig           *ZonalShiftConfigResponse  `json:"zonalShiftConfig,omitempty"`
	RemoteNetworkConfig        *RemoteNetworkConfigResult `json:"remoteNetworkConfig,omitempty"`
	ComputeConfig              *ComputeConfigResponse     `json:"computeConfig,omitempty"`
	StorageConfig              *StorageConfigResponse     `json:"storageConfig,omitempty"`
	ClientRequestToken         string                     `json:"-"`
	BootstrapSelfManagedAddons bool                       `json:"-"`
}

// VpcConfigRequest represents the VPC configuration for a cluster request.
type VpcConfigRequest struct {
	SubnetIDs             []string `json:"subnetIds,omitempty"`
	SecurityGroupIDs      []string `json:"securityGroupIds,omitempty"`
	EndpointPublicAccess  *bool    `json:"endpointPublicAccess,omitempty"`
	EndpointPrivateAccess *bool    `json:"endpointPrivateAccess,omitempty"`
	PublicAccessCidrs     []string `json:"publicAccessCidrs,omitempty"`
}

// VpcConfigResponse represents the VPC configuration for a cluster response.
type VpcConfigResponse struct {
	SubnetIDs              []string `json:"subnetIds,omitempty"`
	SecurityGroupIDs       []string `json:"securityGroupIds,omitempty"`
	ClusterSecurityGroupID string   `json:"clusterSecurityGroupId,omitempty"`
	VpcID                  string   `json:"vpcId,omitempty"`
	EndpointPublicAccess   bool     `json:"endpointPublicAccess"`
	EndpointPrivateAccess  bool     `json:"endpointPrivateAccess"`
	PublicAccessCidrs      []string `json:"publicAccessCidrs,omitempty"`
}

// KubernetesNetworkConfig represents the Kubernetes network configuration.
type KubernetesNetworkConfig struct {
	ServiceIpv4Cidr string `json:"serviceIpv4Cidr,omitempty"`
	ServiceIpv6Cidr string `json:"serviceIpv6Cidr,omitempty"`
	IPFamily        string `json:"ipFamily,omitempty"`
}

// Logging represents the cluster logging configuration.
type Logging struct {
	ClusterLogging []LogSetup `json:"clusterLogging,omitempty"`
}

// LogSetup represents a log setup.
type LogSetup struct {
	Types   []string `json:"types,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
}

// Identity represents the identity provider information.
type Identity struct {
	Oidc *OIDC `json:"oidc,omitempty"`
}

// OIDC represents the OIDC identity provider.
type OIDC struct {
	Issuer string `json:"issuer,omitempty"`
}

// Certificate represents the certificate authority data.
type Certificate struct {
	Data string `json:"data,omitempty"`
}

// EncryptionConfig represents the encryption configuration.
type EncryptionConfig struct {
	Resources []string  `json:"resources,omitempty"`
	Provider  *Provider `json:"provider,omitempty"`
}

// Provider represents the encryption provider.
type Provider struct {
	KeyArn string `json:"keyArn,omitempty"`
}

// ConnectorConfig represents the connector configuration.
type ConnectorConfig struct {
	ActivationID     string `json:"activationId,omitempty"`
	ActivationCode   string `json:"activationCode,omitempty"`
	ActivationExpiry string `json:"activationExpiry,omitempty"`
	Provider         string `json:"provider,omitempty"`
	RoleArn          string `json:"roleArn,omitempty"`
}

// ClusterHealth represents the cluster health status.
type ClusterHealth struct {
	Issues []ClusterIssue `json:"issues,omitempty"`
}

// ClusterIssue represents a cluster issue.
type ClusterIssue struct {
	Code        string   `json:"code,omitempty"`
	Message     string   `json:"message,omitempty"`
	ResourceIDs []string `json:"resourceIds,omitempty"`
}

// OutpostConfig represents the outpost configuration.
type OutpostConfig struct {
	OutpostArns              []string `json:"outpostArns,omitempty"`
	ControlPlaneInstanceType string   `json:"controlPlaneInstanceType,omitempty"`
	ControlPlanePlacement    *string  `json:"controlPlanePlacement,omitempty"`
}

// AccessConfigResponse represents the access configuration response.
type AccessConfigResponse struct {
	BootstrapClusterCreatorAdminPermissions *bool  `json:"bootstrapClusterCreatorAdminPermissions,omitempty"`
	AuthenticationMode                      string `json:"authenticationMode,omitempty"`
}

// UpgradePolicyResponse represents the upgrade policy response.
type UpgradePolicyResponse struct {
	SupportType string `json:"supportType,omitempty"`
}

// ZonalShiftConfigResponse represents the zonal shift configuration response.
type ZonalShiftConfigResponse struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// RemoteNetworkConfigResult represents the remote network configuration result.
type RemoteNetworkConfigResult struct {
	RemoteNodeNetworks []RemoteNodeNetwork `json:"remoteNodeNetworks,omitempty"`
	RemotePodNetworks  []RemotePodNetwork  `json:"remotePodNetworks,omitempty"`
}

// RemoteNodeNetwork represents a remote node network.
type RemoteNodeNetwork struct {
	Cidrs []string `json:"cidrs,omitempty"`
}

// RemotePodNetwork represents a remote pod network.
type RemotePodNetwork struct {
	Cidrs []string `json:"cidrs,omitempty"`
}

// ComputeConfigResponse represents the compute configuration response.
type ComputeConfigResponse struct {
	Enabled     *bool    `json:"enabled,omitempty"`
	NodePools   []string `json:"nodePools,omitempty"`
	NodeRoleArn string   `json:"nodeRoleArn,omitempty"`
}

// StorageConfigResponse represents the storage configuration response.
type StorageConfigResponse struct {
	BlockStorage *BlockStorage `json:"blockStorage,omitempty"`
}

// BlockStorage represents the block storage configuration.
type BlockStorage struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// Nodegroup represents an EKS node group.
type Nodegroup struct {
	NodegroupName  string              `json:"nodegroupName"`
	NodegroupArn   string              `json:"nodegroupArn"`
	ClusterName    string              `json:"clusterName"`
	Version        string              `json:"version,omitempty"`
	ReleaseVersion string              `json:"releaseVersion,omitempty"`
	CreatedAt      *EpochTime          `json:"createdAt,omitempty"`
	ModifiedAt     *EpochTime          `json:"modifiedAt,omitempty"`
	Status         string              `json:"status"`
	CapacityType   string              `json:"capacityType,omitempty"`
	ScalingConfig  *NodegroupScaling   `json:"scalingConfig,omitempty"`
	InstanceTypes  []string            `json:"instanceTypes,omitempty"`
	Subnets        []string            `json:"subnets,omitempty"`
	RemoteAccess   *RemoteAccessConfig `json:"remoteAccess,omitempty"`
	AmiType        string              `json:"amiType,omitempty"`
	NodeRole       string              `json:"nodeRole,omitempty"`
	Labels         map[string]string   `json:"labels,omitempty"`
	Taints         []Taint             `json:"taints,omitempty"`
	Resources      *NodegroupResources `json:"resources,omitempty"`
	DiskSize       *int                `json:"diskSize,omitempty"`
	Health         *NodegroupHealth    `json:"health,omitempty"`
	UpdateConfig   *NodegroupUpdate    `json:"updateConfig,omitempty"`
	LaunchTemplate *LaunchTemplate     `json:"launchTemplate,omitempty"`
	Tags           map[string]string   `json:"tags,omitempty"`
}

// NodegroupScaling represents the scaling configuration for a node group.
type NodegroupScaling struct {
	MinSize     *int `json:"minSize,omitempty"`
	MaxSize     *int `json:"maxSize,omitempty"`
	DesiredSize *int `json:"desiredSize,omitempty"`
}

// RemoteAccessConfig represents the remote access configuration.
type RemoteAccessConfig struct {
	Ec2SshKey            string   `json:"ec2SshKey,omitempty"`
	SourceSecurityGroups []string `json:"sourceSecurityGroups,omitempty"`
}

// Taint represents a Kubernetes taint.
type Taint struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
}

// NodegroupResources represents the resources associated with a node group.
type NodegroupResources struct {
	AutoScalingGroups         []AutoScalingGroup `json:"autoScalingGroups,omitempty"`
	RemoteAccessSecurityGroup string             `json:"remoteAccessSecurityGroup,omitempty"`
}

// AutoScalingGroup represents an auto scaling group.
type AutoScalingGroup struct {
	Name string `json:"name,omitempty"`
}

// NodegroupHealth represents the health status of a node group.
type NodegroupHealth struct {
	Issues []Issue `json:"issues,omitempty"`
}

// Issue represents a node group issue.
type Issue struct {
	Code        string   `json:"code,omitempty"`
	Message     string   `json:"message,omitempty"`
	ResourceIDs []string `json:"resourceIds,omitempty"`
}

// NodegroupUpdate represents the update configuration for a node group.
type NodegroupUpdate struct {
	MaxUnavailable           *int `json:"maxUnavailable,omitempty"`
	MaxUnavailablePercentage *int `json:"maxUnavailablePercentage,omitempty"`
}

// LaunchTemplate represents a launch template specification.
type LaunchTemplate struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	ID      string `json:"id,omitempty"`
}

// Request types.

// CreateClusterRequest represents a CreateCluster request.
type CreateClusterRequest struct {
	Name                       string                   `json:"name"`
	Version                    string                   `json:"version,omitempty"`
	RoleArn                    string                   `json:"roleArn"`
	ResourcesVpcConfig         *VpcConfigRequest        `json:"resourcesVpcConfig"`
	KubernetesNetworkConfig    *KubernetesNetworkConfig `json:"kubernetesNetworkConfig,omitempty"`
	Logging                    *Logging                 `json:"logging,omitempty"`
	ClientRequestToken         string                   `json:"clientRequestToken,omitempty"`
	Tags                       map[string]string        `json:"tags,omitempty"`
	EncryptionConfig           []EncryptionConfig       `json:"encryptionConfig,omitempty"`
	OutpostConfig              *OutpostConfig           `json:"outpostConfig,omitempty"`
	AccessConfig               *AccessConfigRequest     `json:"accessConfig,omitempty"`
	BootstrapSelfManagedAddons *bool                    `json:"bootstrapSelfManagedAddons,omitempty"`
	UpgradePolicy              *UpgradePolicyRequest    `json:"upgradePolicy,omitempty"`
	ZonalShiftConfig           *ZonalShiftConfigRequest `json:"zonalShiftConfig,omitempty"`
	RemoteNetworkConfig        *RemoteNetworkConfig     `json:"remoteNetworkConfig,omitempty"`
	ComputeConfig              *ComputeConfigRequest    `json:"computeConfig,omitempty"`
	StorageConfig              *StorageConfigRequest    `json:"storageConfig,omitempty"`
}

// AccessConfigRequest represents the access configuration request.
type AccessConfigRequest struct {
	BootstrapClusterCreatorAdminPermissions *bool  `json:"bootstrapClusterCreatorAdminPermissions,omitempty"`
	AuthenticationMode                      string `json:"authenticationMode,omitempty"`
}

// UpgradePolicyRequest represents the upgrade policy request.
type UpgradePolicyRequest struct {
	SupportType string `json:"supportType,omitempty"`
}

// ZonalShiftConfigRequest represents the zonal shift configuration request.
type ZonalShiftConfigRequest struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// RemoteNetworkConfig represents the remote network configuration.
type RemoteNetworkConfig struct {
	RemoteNodeNetworks []RemoteNodeNetwork `json:"remoteNodeNetworks,omitempty"`
	RemotePodNetworks  []RemotePodNetwork  `json:"remotePodNetworks,omitempty"`
}

// ComputeConfigRequest represents the compute configuration request.
type ComputeConfigRequest struct {
	Enabled     *bool    `json:"enabled,omitempty"`
	NodePools   []string `json:"nodePools,omitempty"`
	NodeRoleArn string   `json:"nodeRoleArn,omitempty"`
}

// StorageConfigRequest represents the storage configuration request.
type StorageConfigRequest struct {
	BlockStorage *BlockStorage `json:"blockStorage,omitempty"`
}

// CreateNodegroupRequest represents a CreateNodegroup request.
type CreateNodegroupRequest struct {
	ClusterName        string              `json:"-"`
	NodegroupName      string              `json:"nodegroupName"`
	ScalingConfig      *NodegroupScaling   `json:"scalingConfig,omitempty"`
	DiskSize           *int                `json:"diskSize,omitempty"`
	Subnets            []string            `json:"subnets"`
	InstanceTypes      []string            `json:"instanceTypes,omitempty"`
	AmiType            string              `json:"amiType,omitempty"`
	RemoteAccess       *RemoteAccessConfig `json:"remoteAccess,omitempty"`
	NodeRole           string              `json:"nodeRole"`
	Labels             map[string]string   `json:"labels,omitempty"`
	Taints             []Taint             `json:"taints,omitempty"`
	Tags               map[string]string   `json:"tags,omitempty"`
	ClientRequestToken string              `json:"clientRequestToken,omitempty"`
	LaunchTemplate     *LaunchTemplate     `json:"launchTemplate,omitempty"`
	UpdateConfig       *NodegroupUpdate    `json:"updateConfig,omitempty"`
	CapacityType       string              `json:"capacityType,omitempty"`
	Version            string              `json:"version,omitempty"`
	ReleaseVersion     string              `json:"releaseVersion,omitempty"`
}

// Response types.

// CreateClusterResponse represents a CreateCluster response.
type CreateClusterResponse struct {
	Cluster *Cluster `json:"cluster"`
}

// DeleteClusterResponse represents a DeleteCluster response.
type DeleteClusterResponse struct {
	Cluster *Cluster `json:"cluster"`
}

// DescribeClusterResponse represents a DescribeCluster response.
type DescribeClusterResponse struct {
	Cluster *Cluster `json:"cluster"`
}

// ListClustersResponse represents a ListClusters response.
type ListClustersResponse struct {
	Clusters  []string `json:"clusters"`
	NextToken string   `json:"nextToken,omitempty"`
}

// CreateNodegroupResponse represents a CreateNodegroup response.
type CreateNodegroupResponse struct {
	Nodegroup *Nodegroup `json:"nodegroup"`
}

// DeleteNodegroupResponse represents a DeleteNodegroup response.
type DeleteNodegroupResponse struct {
	Nodegroup *Nodegroup `json:"nodegroup"`
}

// DescribeNodegroupResponse represents a DescribeNodegroup response.
type DescribeNodegroupResponse struct {
	Nodegroup *Nodegroup `json:"nodegroup"`
}

// ListNodegroupsResponse represents a ListNodegroups response.
type ListNodegroupsResponse struct {
	Nodegroups []string `json:"nodegroups"`
	NextToken  string   `json:"nextToken,omitempty"`
}

// Error types.

// Error represents an EKS error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
