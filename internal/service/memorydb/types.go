package memorydb

// Tag represents a key-value tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// Endpoint represents a cluster endpoint.
type Endpoint struct {
	Address string `json:"Address"`
	Port    int32  `json:"Port"`
}

// Cluster represents a MemoryDB cluster.
type Cluster struct {
	ACLName                 string                    `json:"ACLName,omitempty"`
	ARN                     string                    `json:"ARN,omitempty"`
	AutoMinorVersionUpgrade bool                      `json:"AutoMinorVersionUpgrade,omitempty"`
	AvailabilityMode        string                    `json:"AvailabilityMode,omitempty"`
	ClusterEndpoint         *Endpoint                 `json:"ClusterEndpoint,omitempty"`
	Description             string                    `json:"Description,omitempty"`
	Engine                  string                    `json:"Engine,omitempty"`
	EngineVersion           string                    `json:"EngineVersion,omitempty"`
	KmsKeyID                string                    `json:"KmsKeyId,omitempty"`
	MaintenanceWindow       string                    `json:"MaintenanceWindow,omitempty"`
	Name                    string                    `json:"Name,omitempty"`
	NodeType                string                    `json:"NodeType,omitempty"`
	NumberOfShards          int32                     `json:"NumberOfShards,omitempty"`
	ParameterGroupName      string                    `json:"ParameterGroupName,omitempty"`
	ParameterGroupStatus    string                    `json:"ParameterGroupStatus,omitempty"`
	SecurityGroups          []SecurityGroupMembership `json:"SecurityGroups,omitempty"`
	SnapshotRetentionLimit  int32                     `json:"SnapshotRetentionLimit,omitempty"`
	SnapshotWindow          string                    `json:"SnapshotWindow,omitempty"`
	SnsTopicArn             string                    `json:"SnsTopicArn,omitempty"`
	SnsTopicStatus          string                    `json:"SnsTopicStatus,omitempty"`
	Status                  string                    `json:"Status,omitempty"`
	SubnetGroupName         string                    `json:"SubnetGroupName,omitempty"`
	TLSEnabled              bool                      `json:"TLSEnabled,omitempty"`
}

// SecurityGroupMembership represents a security group association.
type SecurityGroupMembership struct {
	SecurityGroupID string `json:"SecurityGroupId,omitempty"`
	Status          string `json:"Status,omitempty"`
}

// Authentication represents user authentication info.
type Authentication struct {
	PasswordCount int32  `json:"PasswordCount,omitempty"`
	Type          string `json:"Type,omitempty"`
}

// User represents a MemoryDB user.
type User struct {
	ACLNames             []string        `json:"ACLNames,omitempty"`
	ARN                  string          `json:"ARN,omitempty"`
	AccessString         string          `json:"AccessString,omitempty"`
	Authentication       *Authentication `json:"Authentication,omitempty"`
	MinimumEngineVersion string          `json:"MinimumEngineVersion,omitempty"`
	Name                 string          `json:"Name,omitempty"`
	Status               string          `json:"Status,omitempty"`
}

// ACLPendingChanges represents pending ACL changes.
type ACLPendingChanges struct {
	UserNamesToAdd    []string `json:"UserNamesToAdd,omitempty"`
	UserNamesToRemove []string `json:"UserNamesToRemove,omitempty"`
}

// ACL represents a MemoryDB access control list.
type ACL struct {
	ARN                  string             `json:"ARN,omitempty"`
	Clusters             []string           `json:"Clusters,omitempty"`
	MinimumEngineVersion string             `json:"MinimumEngineVersion,omitempty"`
	Name                 string             `json:"Name,omitempty"`
	PendingChanges       *ACLPendingChanges `json:"PendingChanges,omitempty"`
	Status               string             `json:"Status,omitempty"`
	UserNames            []string           `json:"UserNames,omitempty"`
}

// Snapshot represents a MemoryDB snapshot.
type Snapshot struct {
	ARN         string `json:"ARN,omitempty"`
	ClusterName string `json:"ClusterConfiguration.Name,omitempty"`
	KmsKeyID    string `json:"KmsKeyId,omitempty"`
	Name        string `json:"Name,omitempty"`
	Source      string `json:"Source,omitempty"`
	Status      string `json:"Status,omitempty"`
}

// SubnetGroup represents a MemoryDB subnet group.
type SubnetGroup struct {
	ARN         string   `json:"ARN,omitempty"`
	Description string   `json:"Description,omitempty"`
	Name        string   `json:"Name,omitempty"`
	SubnetIDs   []string `json:"Subnets,omitempty"`
	VpcID       string   `json:"VpcId,omitempty"`
}

// ParameterGroup represents a MemoryDB parameter group.
type ParameterGroup struct {
	ARN         string `json:"ARN,omitempty"`
	Description string `json:"Description,omitempty"`
	Family      string `json:"Family,omitempty"`
	Name        string `json:"Name,omitempty"`
}

// CreateClusterRequest represents a CreateCluster request.
type CreateClusterRequest struct {
	ACLName                 string   `json:"ACLName"`
	ClusterName             string   `json:"ClusterName"`
	NodeType                string   `json:"NodeType"`
	AutoMinorVersionUpgrade *bool    `json:"AutoMinorVersionUpgrade,omitempty"`
	Description             string   `json:"Description,omitempty"`
	Engine                  string   `json:"Engine,omitempty"`
	EngineVersion           string   `json:"EngineVersion,omitempty"`
	KmsKeyID                string   `json:"KmsKeyId,omitempty"`
	MaintenanceWindow       string   `json:"MaintenanceWindow,omitempty"`
	NumReplicasPerShard     *int32   `json:"NumReplicasPerShard,omitempty"`
	NumShards               *int32   `json:"NumShards,omitempty"`
	ParameterGroupName      string   `json:"ParameterGroupName,omitempty"`
	Port                    *int32   `json:"Port,omitempty"`
	SecurityGroupIDs        []string `json:"SecurityGroupIds,omitempty"`
	SnapshotRetentionLimit  *int32   `json:"SnapshotRetentionLimit,omitempty"`
	SnapshotWindow          string   `json:"SnapshotWindow,omitempty"`
	SubnetGroupName         string   `json:"SubnetGroupName,omitempty"`
	TLSEnabled              *bool    `json:"TLSEnabled,omitempty"`
	Tags                    []Tag    `json:"Tags,omitempty"`
}

// CreateClusterResponse represents a CreateCluster response.
type CreateClusterResponse struct {
	Cluster *Cluster `json:"Cluster"`
}

// DeleteClusterRequest represents a DeleteCluster request.
type DeleteClusterRequest struct {
	ClusterName string `json:"ClusterName"`
}

// DeleteClusterResponse represents a DeleteCluster response.
type DeleteClusterResponse struct {
	Cluster *Cluster `json:"Cluster"`
}

// DescribeClustersRequest represents a DescribeClusters request.
type DescribeClustersRequest struct {
	ClusterName string `json:"ClusterName,omitempty"`
	MaxResults  int32  `json:"MaxResults,omitempty"`
	NextToken   string `json:"NextToken,omitempty"`
}

// DescribeClustersResponse represents a DescribeClusters response.
type DescribeClustersResponse struct {
	Clusters  []Cluster `json:"Clusters"`
	NextToken string    `json:"NextToken,omitempty"`
}

// UpdateClusterRequest represents an UpdateCluster request.
type UpdateClusterRequest struct {
	ClusterName            string   `json:"ClusterName"`
	ACLName                string   `json:"ACLName,omitempty"`
	Description            string   `json:"Description,omitempty"`
	EngineVersion          string   `json:"EngineVersion,omitempty"`
	MaintenanceWindow      string   `json:"MaintenanceWindow,omitempty"`
	NodeType               string   `json:"NodeType,omitempty"`
	ParameterGroupName     string   `json:"ParameterGroupName,omitempty"`
	SecurityGroupIDs       []string `json:"SecurityGroupIds,omitempty"`
	SnapshotRetentionLimit *int32   `json:"SnapshotRetentionLimit,omitempty"`
	SnapshotWindow         string   `json:"SnapshotWindow,omitempty"`
	SnsTopicArn            string   `json:"SnsTopicArn,omitempty"`
	SnsTopicStatus         string   `json:"SnsTopicStatus,omitempty"`
}

// UpdateClusterResponse represents an UpdateCluster response.
type UpdateClusterResponse struct {
	Cluster *Cluster `json:"Cluster"`
}

// AuthenticationMode represents user authentication mode in request.
type AuthenticationMode struct {
	Passwords []string `json:"Passwords,omitempty"`
	Type      string   `json:"Type,omitempty"`
}

// CreateUserRequest represents a CreateUser request.
type CreateUserRequest struct {
	AccessString       string              `json:"AccessString"`
	AuthenticationMode *AuthenticationMode `json:"AuthenticationMode"`
	UserName           string              `json:"UserName"`
	Tags               []Tag               `json:"Tags,omitempty"`
}

// CreateUserResponse represents a CreateUser response.
type CreateUserResponse struct {
	User *User `json:"User"`
}

// DeleteUserRequest represents a DeleteUser request.
type DeleteUserRequest struct {
	UserName string `json:"UserName"`
}

// DeleteUserResponse represents a DeleteUser response.
type DeleteUserResponse struct {
	User *User `json:"User"`
}

// DescribeUsersRequest represents a DescribeUsers request.
type DescribeUsersRequest struct {
	UserName   string `json:"UserName,omitempty"`
	MaxResults int32  `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// DescribeUsersResponse represents a DescribeUsers response.
type DescribeUsersResponse struct {
	Users     []User `json:"Users"`
	NextToken string `json:"NextToken,omitempty"`
}

// CreateACLRequest represents a CreateACL request.
type CreateACLRequest struct {
	ACLName   string   `json:"ACLName"`
	UserNames []string `json:"UserNames,omitempty"`
	Tags      []Tag    `json:"Tags,omitempty"`
}

// CreateACLResponse represents a CreateACL response.
type CreateACLResponse struct {
	ACL *ACL `json:"ACL"`
}

// DeleteACLRequest represents a DeleteACL request.
type DeleteACLRequest struct {
	ACLName string `json:"ACLName"`
}

// DeleteACLResponse represents a DeleteACL response.
type DeleteACLResponse struct {
	ACL *ACL `json:"ACL"`
}

// DescribeACLsRequest represents a DescribeACLs request.
type DescribeACLsRequest struct {
	ACLName    string `json:"ACLName,omitempty"`
	MaxResults int32  `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// DescribeACLsResponse represents a DescribeACLs response.
type DescribeACLsResponse struct {
	ACLs      []ACL  `json:"ACLs"`
	NextToken string `json:"NextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}
