//nolint:tagliatelle // AWS DS API uses PascalCase for JSON tags
package ds

import "time"

// Directory states.
const (
	DirectoryStateCreating   = "Creating"
	DirectoryStateCreated    = "Created"
	DirectoryStateActive     = "Active"
	DirectoryStateDeleting   = "Deleting"
	DirectoryStateDeleted    = "Deleted"
	DirectoryStateFailed     = "Failed"
	DirectoryStateRestoring  = "Restoring"
	DirectoryStateImpaired   = "Impaired"
	DirectoryStateInoperable = "Inoperable"
	DirectoryStateRequested  = "Requested"
)

// Directory sizes.
const (
	DirectorySizeSmall = "Small"
	DirectorySizeLarge = "Large"
)

// Directory types.
const (
	DirectoryTypeSimpleAD    = "SimpleAD"
	DirectoryTypeMicrosoftAD = "MicrosoftAD"
	DirectoryTypeADConnector = "ADConnector"
)

// Snapshot states.
const (
	SnapshotStateCreating  = "Creating"
	SnapshotStateCompleted = "Completed"
	SnapshotStateFailed    = "Failed"
)

// Snapshot types.
const (
	SnapshotTypeAuto   = "Auto"
	SnapshotTypeManual = "Manual"
)

// Directory represents a directory in the service.
type Directory struct {
	DirectoryID        string
	Name               string
	ShortName          string
	Password           string
	Description        string
	Size               string
	Type               string
	Stage              string
	StageReason        string
	DNSIPAddrs         []string
	AccessURL          string
	Alias              string
	SSOEnabled         bool
	DesiredNumberOfDCs int
	VPCSettings        *DirectoryVPCSettings
	ConnectSettings    *DirectoryConnectSettings
	RadiusSettings     *RadiusSettings
	LaunchTime         time.Time
	StageLastUpdatedAt time.Time
	Edition            string
	RegionsInfo        *RegionsInfo
}

// DirectoryVPCSettings represents VPC settings for a directory.
type DirectoryVPCSettings struct {
	VPCID             string
	SubnetIDs         []string
	SecurityGroupID   string
	AvailabilityZones []string
}

// DirectoryConnectSettings represents connection settings for AD Connector.
type DirectoryConnectSettings struct {
	VPCID             string
	SubnetIDs         []string
	CustomerDNSIPs    []string
	CustomerUserName  string
	SecurityGroupID   string
	AvailabilityZones []string
	ConnectIPs        []string
}

// RadiusSettings represents RADIUS authentication settings.
type RadiusSettings struct {
	RadiusServers          []string
	RadiusPort             int
	RadiusTimeout          int
	RadiusRetries          int
	SharedSecret           string
	AuthenticationProtocol string
	DisplayLabel           string
	UseSameUsername        bool
}

// RegionsInfo represents multi-region information.
type RegionsInfo struct {
	PrimaryRegion     string
	AdditionalRegions []string
}

// Snapshot represents a directory snapshot.
type Snapshot struct {
	SnapshotID  string
	DirectoryID string
	Name        string
	Type        string
	Status      string
	StartTime   time.Time
}

// CreateDirectoryRequest is the request for CreateDirectory.
type CreateDirectoryRequest struct {
	Name        string                   `json:"Name"`
	ShortName   string                   `json:"ShortName,omitempty"`
	Password    string                   `json:"Password"`
	Description string                   `json:"Description,omitempty"`
	Size        string                   `json:"Size"`
	VPCSettings *DirectoryVPCSettingsReq `json:"VpcSettings,omitempty"`
	Tags        []Tag                    `json:"Tags,omitempty"`
}

// DirectoryVPCSettingsReq is the VPC settings in request.
type DirectoryVPCSettingsReq struct {
	VPCID     string   `json:"VpcId"`
	SubnetIDs []string `json:"SubnetIds"`
}

// Tag represents a resource tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// CreateDirectoryResponse is the response for CreateDirectory.
type CreateDirectoryResponse struct {
	DirectoryID string `json:"DirectoryId"`
}

// DescribeDirectoriesRequest is the request for DescribeDirectories.
type DescribeDirectoriesRequest struct {
	DirectoryIDs []string `json:"DirectoryIds,omitempty"`
	NextToken    string   `json:"NextToken,omitempty"`
	Limit        int      `json:"Limit,omitempty"`
}

// DescribeDirectoriesResponse is the response for DescribeDirectories.
type DescribeDirectoriesResponse struct {
	DirectoryDescriptions []*DirectoryDescription `json:"DirectoryDescriptions"`
	NextToken             string                  `json:"NextToken,omitempty"`
}

// DirectoryDescription represents a directory description.
type DirectoryDescription struct {
	DirectoryID        string                    `json:"DirectoryId"`
	Name               string                    `json:"Name"`
	ShortName          string                    `json:"ShortName,omitempty"`
	Size               string                    `json:"Size,omitempty"`
	Edition            string                    `json:"Edition,omitempty"`
	Alias              string                    `json:"Alias,omitempty"`
	AccessURL          string                    `json:"AccessUrl,omitempty"`
	Description        string                    `json:"Description,omitempty"`
	DNSIPAddrs         []string                  `json:"DnsIpAddrs,omitempty"`
	Stage              string                    `json:"Stage"`
	StageReason        string                    `json:"StageReason,omitempty"`
	LaunchTime         float64                   `json:"LaunchTime,omitempty"`
	StageLastUpdatedAt float64                   `json:"StageLastUpdatedDateTime,omitempty"`
	Type               string                    `json:"Type"`
	VPCSettings        *DirectoryVPCSettingsResp `json:"VpcSettings,omitempty"`
	ConnectSettings    *DirectoryConnectResp     `json:"ConnectSettings,omitempty"`
	SSOEnabled         bool                      `json:"SsoEnabled"`
	DesiredNumberOfDCs int                       `json:"DesiredNumberOfDomainControllers,omitempty"`
	RegionsInfo        *RegionsInfoResp          `json:"RegionsInfo,omitempty"`
}

// DirectoryVPCSettingsResp is the VPC settings in response.
type DirectoryVPCSettingsResp struct {
	VPCID             string   `json:"VpcId"`
	SubnetIDs         []string `json:"SubnetIds"`
	SecurityGroupID   string   `json:"SecurityGroupId,omitempty"`
	AvailabilityZones []string `json:"AvailabilityZones,omitempty"`
}

// DirectoryConnectResp is the connect settings in response.
type DirectoryConnectResp struct {
	VPCID             string   `json:"VpcId"`
	SubnetIDs         []string `json:"SubnetIds"`
	CustomerDNSIPs    []string `json:"CustomerDnsIps,omitempty"`
	CustomerUserName  string   `json:"CustomerUserName,omitempty"`
	SecurityGroupID   string   `json:"SecurityGroupId,omitempty"`
	AvailabilityZones []string `json:"AvailabilityZones,omitempty"`
	ConnectIPs        []string `json:"ConnectIps,omitempty"`
}

// RegionsInfoResp is the regions info in response.
type RegionsInfoResp struct {
	PrimaryRegion     string   `json:"PrimaryRegion,omitempty"`
	AdditionalRegions []string `json:"AdditionalRegions,omitempty"`
}

// DeleteDirectoryRequest is the request for DeleteDirectory.
type DeleteDirectoryRequest struct {
	DirectoryID string `json:"DirectoryId"`
}

// DeleteDirectoryResponse is the response for DeleteDirectory.
type DeleteDirectoryResponse struct {
	DirectoryID string `json:"DirectoryId"`
}

// CreateSnapshotRequest is the request for CreateSnapshot.
type CreateSnapshotRequest struct {
	DirectoryID string `json:"DirectoryId"`
	Name        string `json:"Name,omitempty"`
}

// CreateSnapshotResponse is the response for CreateSnapshot.
type CreateSnapshotResponse struct {
	SnapshotID string `json:"SnapshotId"`
}

// DescribeSnapshotsRequest is the request for DescribeSnapshots.
type DescribeSnapshotsRequest struct {
	DirectoryID string   `json:"DirectoryId,omitempty"`
	SnapshotIDs []string `json:"SnapshotIds,omitempty"`
	NextToken   string   `json:"NextToken,omitempty"`
	Limit       int      `json:"Limit,omitempty"`
}

// DescribeSnapshotsResponse is the response for DescribeSnapshots.
type DescribeSnapshotsResponse struct {
	Snapshots []*SnapshotDescription `json:"Snapshots"`
	NextToken string                 `json:"NextToken,omitempty"`
}

// SnapshotDescription represents a snapshot description.
type SnapshotDescription struct {
	SnapshotID  string  `json:"SnapshotId"`
	DirectoryID string  `json:"DirectoryId"`
	Name        string  `json:"Name,omitempty"`
	Type        string  `json:"Type"`
	Status      string  `json:"Status"`
	StartTime   float64 `json:"StartTime,omitempty"`
}

// DeleteSnapshotRequest is the request for DeleteSnapshot.
type DeleteSnapshotRequest struct {
	SnapshotID string `json:"SnapshotId"`
}

// DeleteSnapshotResponse is the response for DeleteSnapshot.
type DeleteSnapshotResponse struct {
	SnapshotID string `json:"SnapshotId"`
}

// Error represents a Directory Service error.
type Error struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Error codes for Directory Service.
const (
	ErrEntityDoesNotExist     = "EntityDoesNotExistException"
	ErrEntityAlreadyExists    = "EntityAlreadyExistsException"
	ErrInvalidParameter       = "InvalidParameterException"
	ErrDirectoryLimitExceeded = "DirectoryLimitExceededException"
	ErrSnapshotLimitExceeded  = "SnapshotLimitExceededException"
	ErrServiceException       = "ServiceException"
	ErrClientException        = "ClientException"
	ErrUnsupportedOperation   = "UnsupportedOperationException"
)
