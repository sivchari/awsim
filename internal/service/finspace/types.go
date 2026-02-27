// Package finspace provides an in-memory implementation of AWS FinSpace.
package finspace

// Error represents an error response.
type Error struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// KxEnvironment represents a managed kdb environment.
type KxEnvironment struct {
	AvailabilityZoneIDs       []string           `json:"availabilityZoneIds,omitempty"`
	AwsAccountID              string             `json:"awsAccountId,omitempty"`
	CertificateARN            string             `json:"certificateArn,omitempty"`
	CreationTimestamp         float64            `json:"creationTimestamp,omitempty"`
	CustomDNSConfiguration    []*CustomDNSServer `json:"customDNSConfiguration,omitempty"`
	DedicatedServiceAccountID string             `json:"dedicatedServiceAccountId,omitempty"`
	Description               string             `json:"description,omitempty"`
	DNSStatus                 string             `json:"dnsStatus,omitempty"`
	EnvironmentARN            string             `json:"environmentArn,omitempty"`
	EnvironmentID             string             `json:"environmentId,omitempty"`
	ErrorMessage              string             `json:"errorMessage,omitempty"`
	KmsKeyID                  string             `json:"kmsKeyId,omitempty"`
	Name                      string             `json:"name,omitempty"`
	Status                    string             `json:"status,omitempty"`
	TgwStatus                 string             `json:"tgwStatus,omitempty"`
	UpdateTimestamp           float64            `json:"updateTimestamp,omitempty"`
}

// CustomDNSServer represents a custom DNS server configuration.
type CustomDNSServer struct {
	CustomDNSServerIP   string `json:"customDNSServerIP,omitempty"`
	CustomDNSServerName string `json:"customDNSServerName,omitempty"`
}

// TransitGatewayConfiguration represents transit gateway configuration.
type TransitGatewayConfiguration struct {
	AttachmentNetworkACLConfiguration []*NetworkACLEntry `json:"attachmentNetworkAclConfiguration,omitempty"`
	RoutableCIDRSpace                 string             `json:"routableCIDRSpace,omitempty"`
	TransitGatewayID                  string             `json:"transitGatewayID,omitempty"`
}

// NetworkACLEntry represents a network ACL entry.
type NetworkACLEntry struct {
	CidrBlock  string        `json:"cidrBlock,omitempty"`
	ICMPType   *ICMPTypeCode `json:"icmpTypeCode,omitempty"`
	PortRange  *PortRange    `json:"portRange,omitempty"`
	Protocol   string        `json:"protocol,omitempty"`
	RuleAction string        `json:"ruleAction,omitempty"`
	RuleNumber int           `json:"ruleNumber,omitempty"`
}

// ICMPTypeCode represents ICMP type and code.
type ICMPTypeCode struct {
	Code int `json:"code,omitempty"`
	Type int `json:"type,omitempty"`
}

// PortRange represents a port range.
type PortRange struct {
	From int `json:"from,omitempty"`
	To   int `json:"to,omitempty"`
}

// KxDatabase represents a kdb database.
type KxDatabase struct {
	CreatedTimestamp         float64 `json:"createdTimestamp,omitempty"`
	DatabaseARN              string  `json:"databaseArn,omitempty"`
	DatabaseName             string  `json:"databaseName,omitempty"`
	Description              string  `json:"description,omitempty"`
	EnvironmentID            string  `json:"environmentId,omitempty"`
	LastCompletedChangesetID string  `json:"lastCompletedChangesetId,omitempty"`
	LastModifiedTimestamp    float64 `json:"lastModifiedTimestamp,omitempty"`
	NumBytes                 int64   `json:"numBytes,omitempty"`
	NumChangesets            int     `json:"numChangesets,omitempty"`
	NumFiles                 int     `json:"numFiles,omitempty"`
}

// KxUser represents a kdb user.
type KxUser struct {
	CreateTimestamp float64 `json:"createTimestamp,omitempty"`
	IamRole         string  `json:"iamRole,omitempty"`
	UpdateTimestamp float64 `json:"updateTimestamp,omitempty"`
	UserARN         string  `json:"userArn,omitempty"`
	UserName        string  `json:"userName,omitempty"`
}

// Tag represents a resource tag.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateKxEnvironmentRequest represents a CreateKxEnvironment request.
type CreateKxEnvironmentRequest struct {
	ClientToken string            `json:"clientToken,omitempty"`
	Description string            `json:"description,omitempty"`
	KmsKeyID    string            `json:"kmsKeyId,omitempty"`
	Name        string            `json:"name"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// CreateKxEnvironmentResponse represents a CreateKxEnvironment response.
type CreateKxEnvironmentResponse struct {
	CreationTimestamp float64 `json:"creationTimestamp,omitempty"`
	Description       string  `json:"description,omitempty"`
	EnvironmentARN    string  `json:"environmentArn,omitempty"`
	EnvironmentID     string  `json:"environmentId,omitempty"`
	KmsKeyID          string  `json:"kmsKeyId,omitempty"`
	Name              string  `json:"name,omitempty"`
	Status            string  `json:"status,omitempty"`
}

// GetKxEnvironmentRequest represents a GetKxEnvironment request.
type GetKxEnvironmentRequest struct {
	EnvironmentID string `json:"environmentId"`
}

// GetKxEnvironmentResponse represents a GetKxEnvironment response.
type GetKxEnvironmentResponse struct {
	AvailabilityZoneIDs         []string                     `json:"availabilityZoneIds,omitempty"`
	AwsAccountID                string                       `json:"awsAccountId,omitempty"`
	CertificateARN              string                       `json:"certificateArn,omitempty"`
	CreationTimestamp           float64                      `json:"creationTimestamp,omitempty"`
	CustomDNSConfiguration      []*CustomDNSServer           `json:"customDNSConfiguration,omitempty"`
	DedicatedServiceAccountID   string                       `json:"dedicatedServiceAccountId,omitempty"`
	Description                 string                       `json:"description,omitempty"`
	DNSStatus                   string                       `json:"dnsStatus,omitempty"`
	EnvironmentARN              string                       `json:"environmentArn,omitempty"`
	EnvironmentID               string                       `json:"environmentId,omitempty"`
	ErrorMessage                string                       `json:"errorMessage,omitempty"`
	KmsKeyID                    string                       `json:"kmsKeyId,omitempty"`
	Name                        string                       `json:"name,omitempty"`
	Status                      string                       `json:"status,omitempty"`
	TgwStatus                   string                       `json:"tgwStatus,omitempty"`
	TransitGatewayConfiguration *TransitGatewayConfiguration `json:"transitGatewayConfiguration,omitempty"`
	UpdateTimestamp             float64                      `json:"updateTimestamp,omitempty"`
}

// DeleteKxEnvironmentRequest represents a DeleteKxEnvironment request.
type DeleteKxEnvironmentRequest struct {
	ClientToken   string `json:"clientToken,omitempty"`
	EnvironmentID string `json:"environmentId"`
}

// DeleteKxEnvironmentResponse represents a DeleteKxEnvironment response.
type DeleteKxEnvironmentResponse struct{}

// ListKxEnvironmentsRequest represents a ListKxEnvironments request.
type ListKxEnvironmentsRequest struct {
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

// ListKxEnvironmentsResponse represents a ListKxEnvironments response.
type ListKxEnvironmentsResponse struct {
	Environments []*KxEnvironment `json:"environments,omitempty"`
	NextToken    string           `json:"nextToken,omitempty"`
}

// UpdateKxEnvironmentRequest represents an UpdateKxEnvironment request.
type UpdateKxEnvironmentRequest struct {
	ClientToken   string `json:"clientToken,omitempty"`
	Description   string `json:"description,omitempty"`
	EnvironmentID string `json:"environmentId"`
	Name          string `json:"name,omitempty"`
}

// UpdateKxEnvironmentResponse represents an UpdateKxEnvironment response.
type UpdateKxEnvironmentResponse struct {
	AvailabilityZoneIDs         []string                     `json:"availabilityZoneIds,omitempty"`
	AwsAccountID                string                       `json:"awsAccountId,omitempty"`
	CreationTimestamp           float64                      `json:"creationTimestamp,omitempty"`
	DedicatedServiceAccountID   string                       `json:"dedicatedServiceAccountId,omitempty"`
	Description                 string                       `json:"description,omitempty"`
	DNSStatus                   string                       `json:"dnsStatus,omitempty"`
	EnvironmentARN              string                       `json:"environmentArn,omitempty"`
	EnvironmentID               string                       `json:"environmentId,omitempty"`
	KmsKeyID                    string                       `json:"kmsKeyId,omitempty"`
	Name                        string                       `json:"name,omitempty"`
	Status                      string                       `json:"status,omitempty"`
	TgwStatus                   string                       `json:"tgwStatus,omitempty"`
	TransitGatewayConfiguration *TransitGatewayConfiguration `json:"transitGatewayConfiguration,omitempty"`
	UpdateTimestamp             float64                      `json:"updateTimestamp,omitempty"`
}

// CreateKxDatabaseRequest represents a CreateKxDatabase request.
type CreateKxDatabaseRequest struct {
	ClientToken   string            `json:"clientToken,omitempty"`
	DatabaseName  string            `json:"databaseName"`
	Description   string            `json:"description,omitempty"`
	EnvironmentID string            `json:"environmentId"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// CreateKxDatabaseResponse represents a CreateKxDatabase response.
type CreateKxDatabaseResponse struct {
	CreatedTimestamp float64 `json:"createdTimestamp,omitempty"`
	DatabaseARN      string  `json:"databaseArn,omitempty"`
	DatabaseName     string  `json:"databaseName,omitempty"`
	Description      string  `json:"description,omitempty"`
	EnvironmentID    string  `json:"environmentId,omitempty"`
}

// GetKxDatabaseRequest represents a GetKxDatabase request.
type GetKxDatabaseRequest struct {
	DatabaseName  string `json:"databaseName"`
	EnvironmentID string `json:"environmentId"`
}

// GetKxDatabaseResponse represents a GetKxDatabase response.
type GetKxDatabaseResponse struct {
	CreatedTimestamp         float64 `json:"createdTimestamp,omitempty"`
	DatabaseARN              string  `json:"databaseArn,omitempty"`
	DatabaseName             string  `json:"databaseName,omitempty"`
	Description              string  `json:"description,omitempty"`
	EnvironmentID            string  `json:"environmentId,omitempty"`
	LastCompletedChangesetID string  `json:"lastCompletedChangesetId,omitempty"`
	LastModifiedTimestamp    float64 `json:"lastModifiedTimestamp,omitempty"`
	NumBytes                 int64   `json:"numBytes,omitempty"`
	NumChangesets            int     `json:"numChangesets,omitempty"`
	NumFiles                 int     `json:"numFiles,omitempty"`
}

// DeleteKxDatabaseRequest represents a DeleteKxDatabase request.
type DeleteKxDatabaseRequest struct {
	ClientToken   string `json:"clientToken,omitempty"`
	DatabaseName  string `json:"databaseName"`
	EnvironmentID string `json:"environmentId"`
}

// DeleteKxDatabaseResponse represents a DeleteKxDatabase response.
type DeleteKxDatabaseResponse struct{}

// ListKxDatabasesRequest represents a ListKxDatabases request.
type ListKxDatabasesRequest struct {
	EnvironmentID string `json:"environmentId"`
	MaxResults    int    `json:"maxResults,omitempty"`
	NextToken     string `json:"nextToken,omitempty"`
}

// ListKxDatabasesResponse represents a ListKxDatabases response.
type ListKxDatabasesResponse struct {
	KxDatabases []*KxDatabase `json:"kxDatabases,omitempty"`
	NextToken   string        `json:"nextToken,omitempty"`
}

// UpdateKxDatabaseRequest represents an UpdateKxDatabase request.
type UpdateKxDatabaseRequest struct {
	ClientToken   string `json:"clientToken,omitempty"`
	DatabaseName  string `json:"databaseName"`
	Description   string `json:"description,omitempty"`
	EnvironmentID string `json:"environmentId"`
}

// UpdateKxDatabaseResponse represents an UpdateKxDatabase response.
type UpdateKxDatabaseResponse struct {
	DatabaseARN           string  `json:"databaseArn,omitempty"`
	DatabaseName          string  `json:"databaseName,omitempty"`
	Description           string  `json:"description,omitempty"`
	EnvironmentID         string  `json:"environmentId,omitempty"`
	LastModifiedTimestamp float64 `json:"lastModifiedTimestamp,omitempty"`
}

// CreateKxUserRequest represents a CreateKxUser request.
type CreateKxUserRequest struct {
	ClientToken   string            `json:"clientToken,omitempty"`
	EnvironmentID string            `json:"environmentId"`
	IamRole       string            `json:"iamRole"`
	Tags          map[string]string `json:"tags,omitempty"`
	UserName      string            `json:"userName"`
}

// CreateKxUserResponse represents a CreateKxUser response.
type CreateKxUserResponse struct {
	EnvironmentID string `json:"environmentId,omitempty"`
	IamRole       string `json:"iamRole,omitempty"`
	UserARN       string `json:"userArn,omitempty"`
	UserName      string `json:"userName,omitempty"`
}

// GetKxUserRequest represents a GetKxUser request.
type GetKxUserRequest struct {
	EnvironmentID string `json:"environmentId"`
	UserName      string `json:"userName"`
}

// GetKxUserResponse represents a GetKxUser response.
type GetKxUserResponse struct {
	EnvironmentID string `json:"environmentId,omitempty"`
	IamRole       string `json:"iamRole,omitempty"`
	UserARN       string `json:"userArn,omitempty"`
	UserName      string `json:"userName,omitempty"`
}

// DeleteKxUserRequest represents a DeleteKxUser request.
type DeleteKxUserRequest struct {
	ClientToken   string `json:"clientToken,omitempty"`
	EnvironmentID string `json:"environmentId"`
	UserName      string `json:"userName"`
}

// DeleteKxUserResponse represents a DeleteKxUser response.
type DeleteKxUserResponse struct{}

// ListKxUsersRequest represents a ListKxUsers request.
type ListKxUsersRequest struct {
	EnvironmentID string `json:"environmentId"`
	MaxResults    int    `json:"maxResults,omitempty"`
	NextToken     string `json:"nextToken,omitempty"`
}

// ListKxUsersResponse represents a ListKxUsers response.
type ListKxUsersResponse struct {
	NextToken string    `json:"nextToken,omitempty"`
	Users     []*KxUser `json:"users,omitempty"`
}

// UpdateKxUserRequest represents an UpdateKxUser request.
type UpdateKxUserRequest struct {
	ClientToken   string `json:"clientToken,omitempty"`
	EnvironmentID string `json:"environmentId"`
	IamRole       string `json:"iamRole"`
	UserName      string `json:"userName"`
}

// UpdateKxUserResponse represents an UpdateKxUser response.
type UpdateKxUserResponse struct {
	EnvironmentID string `json:"environmentId,omitempty"`
	IamRole       string `json:"iamRole,omitempty"`
	UserARN       string `json:"userArn,omitempty"`
	UserName      string `json:"userName,omitempty"`
}

// TagResourceRequest represents a TagResource request.
type TagResourceRequest struct {
	ResourceARN string            `json:"resourceArn"`
	Tags        map[string]string `json:"tags"`
}

// TagResourceResponse represents a TagResource response.
type TagResourceResponse struct{}

// UntagResourceRequest represents an UntagResource request.
type UntagResourceRequest struct {
	ResourceARN string   `json:"resourceArn"`
	TagKeys     []string `json:"tagKeys"`
}

// UntagResourceResponse represents an UntagResource response.
type UntagResourceResponse struct{}

// ListTagsForResourceRequest represents a ListTagsForResource request.
type ListTagsForResourceRequest struct {
	ResourceARN string `json:"resourceArn"`
}

// ListTagsForResourceResponse represents a ListTagsForResource response.
type ListTagsForResourceResponse struct {
	Tags map[string]string `json:"tags,omitempty"`
}
