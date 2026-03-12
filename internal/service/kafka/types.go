package kafka

// BrokerNodeGroupInfo represents the broker node group configuration.
type BrokerNodeGroupInfo struct {
	ClientSubnets  []string     `json:"clientSubnets"`
	InstanceType   string       `json:"instanceType"`
	SecurityGroups []string     `json:"securityGroups,omitempty"`
	StorageInfo    *StorageInfo `json:"storageInfo,omitempty"`
}

// StorageInfo represents storage configuration.
type StorageInfo struct {
	EBSStorageInfo *EBSStorageInfo `json:"ebsStorageInfo,omitempty"`
}

// EBSStorageInfo represents EBS storage configuration.
type EBSStorageInfo struct {
	VolumeSize int `json:"volumeSize,omitempty"`
}

// EncryptionInfo represents encryption configuration.
type EncryptionInfo struct {
	EncryptionAtRest    *EncryptionAtRest    `json:"encryptionAtRest,omitempty"`
	EncryptionInTransit *EncryptionInTransit `json:"encryptionInTransit,omitempty"`
}

// EncryptionAtRest represents encryption at rest configuration.
type EncryptionAtRest struct {
	DataVolumeKMSKeyID string `json:"dataVolumeKMSKeyId,omitempty"`
}

// EncryptionInTransit represents encryption in transit configuration.
type EncryptionInTransit struct {
	ClientBroker string `json:"clientBroker,omitempty"`
	InCluster    *bool  `json:"inCluster,omitempty"`
}

// BrokerSoftwareInfo represents the broker software information.
type BrokerSoftwareInfo struct {
	KafkaVersion          string `json:"kafkaVersion,omitempty"`
	ConfigurationArn      string `json:"configurationArn,omitempty"`
	ConfigurationRevision int64  `json:"configurationRevision,omitempty"`
}

// ClusterInfo represents an MSK cluster.
type ClusterInfo struct {
	ClusterArn                string               `json:"clusterArn"`
	ClusterName               string               `json:"clusterName"`
	CreationTime              string               `json:"creationTime,omitempty"`
	CurrentVersion            string               `json:"currentVersion"`
	State                     string               `json:"state"`
	CurrentBrokerSoftwareInfo *BrokerSoftwareInfo  `json:"currentBrokerSoftwareInfo,omitempty"`
	NumberOfBrokerNodes       int                  `json:"numberOfBrokerNodes"`
	BrokerNodeGroupInfo       *BrokerNodeGroupInfo `json:"brokerNodeGroupInfo,omitempty"`
	EncryptionInfo            *EncryptionInfo      `json:"encryptionInfo,omitempty"`
	Tags                      map[string]string    `json:"tags,omitempty"`
}

// CreateClusterRequest represents a CreateCluster request.
type CreateClusterRequest struct {
	ClusterName         string               `json:"clusterName"`
	KafkaVersion        string               `json:"kafkaVersion"`
	NumberOfBrokerNodes int                  `json:"numberOfBrokerNodes"`
	BrokerNodeGroupInfo *BrokerNodeGroupInfo `json:"brokerNodeGroupInfo"`
	EncryptionInfo      *EncryptionInfo      `json:"encryptionInfo,omitempty"`
	Tags                map[string]string    `json:"tags,omitempty"`
}

// CreateClusterResponse represents a CreateCluster response.
type CreateClusterResponse struct {
	ClusterArn  string `json:"clusterArn"`
	ClusterName string `json:"clusterName"`
	State       string `json:"state"`
}

// DescribeClusterResponse represents a DescribeCluster response.
type DescribeClusterResponse struct {
	ClusterInfo *ClusterInfo `json:"clusterInfo"`
}

// DeleteClusterResponse represents a DeleteCluster response.
type DeleteClusterResponse struct {
	ClusterArn string `json:"clusterArn"`
	State      string `json:"state"`
}

// ListClustersResponse represents a ListClusters response.
type ListClustersResponse struct {
	ClusterInfoList []ClusterInfo `json:"clusterInfoList"`
	NextToken       string        `json:"nextToken,omitempty"`
}

// GetBootstrapBrokersResponse represents a GetBootstrapBrokers response.
type GetBootstrapBrokersResponse struct {
	BootstrapBrokerString    string `json:"bootstrapBrokerString,omitempty"`
	BootstrapBrokerStringTLS string `json:"bootstrapBrokerStringTls,omitempty"`
}

// UpdateClusterConfigurationRequest represents an UpdateClusterConfiguration request.
type UpdateClusterConfigurationRequest struct {
	ConfigurationInfo *ConfigurationInfo `json:"configurationInfo"`
	CurrentVersion    string             `json:"currentVersion"`
}

// ConfigurationInfo represents configuration info for a cluster.
type ConfigurationInfo struct {
	Arn      string `json:"arn"`
	Revision int64  `json:"revision"`
}

// UpdateClusterConfigurationResponse represents an UpdateClusterConfiguration response.
type UpdateClusterConfigurationResponse struct {
	ClusterArn          string `json:"clusterArn"`
	ClusterOperationArn string `json:"clusterOperationArn"`
}

// Error represents an MSK error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
