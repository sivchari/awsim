package gamelift

import "time"

// Build represents a GameLift build.
type Build struct {
	BuildID         string
	BuildARN        string
	Name            string
	Version         string
	Status          string
	SizeOnDisk      int64
	OperatingSystem string
	CreationTime    time.Time
}

// Fleet represents a GameLift fleet.
type Fleet struct {
	FleetID                        string
	FleetARN                       string
	Name                           string
	Description                    string
	BuildID                        string
	Status                         string
	FleetType                      string
	InstanceType                   string
	ServerLaunchPath               string
	NewGameSessionProtectionPolicy string
	CreationTime                   time.Time
}

// GameSession represents a game session.
type GameSession struct {
	GameSessionID             string
	GameSessionARN            string
	FleetID                   string
	FleetARN                  string
	Name                      string
	Status                    string
	CurrentPlayerSessionCount int
	MaximumPlayerSessionCount int
	IPAddress                 string
	Port                      int
	CreationTime              time.Time
}

// PlayerSession represents a player session.
type PlayerSession struct {
	PlayerSessionID string
	GameSessionID   string
	FleetID         string
	FleetARN        string
	PlayerID        string
	Status          string
	IPAddress       string
	Port            int
	CreationTime    time.Time
}

// Error represents a GameLift service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// CreateBuildRequest represents the CreateBuild API request.
type CreateBuildRequest struct {
	Name             string           `json:"Name,omitempty"`
	Version          string           `json:"Version,omitempty"`
	OperatingSystem  string           `json:"OperatingSystem,omitempty"`
	Tags             []TagInput       `json:"Tags,omitempty"`
	StorageLocation  *StorageLocation `json:"StorageLocation,omitempty"`
	ServerSdkVersion string           `json:"ServerSdkVersion,omitempty"`
}

// TagInput represents a tag input.
type TagInput struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// StorageLocation represents an S3 storage location.
type StorageLocation struct {
	Bucket        string `json:"Bucket,omitempty"`
	Key           string `json:"Key,omitempty"`
	RoleArn       string `json:"RoleArn,omitempty"`
	ObjectVersion string `json:"ObjectVersion,omitempty"`
}

// CreateBuildResponse represents the CreateBuild API response.
type CreateBuildResponse struct {
	Build             *BuildOutput       `json:"Build,omitempty"`
	UploadCredentials *UploadCredentials `json:"UploadCredentials,omitempty"`
	StorageLocation   *StorageLocation   `json:"StorageLocation,omitempty"`
}

// BuildOutput represents the output format of a build.
type BuildOutput struct {
	BuildID          string  `json:"BuildId,omitempty"`
	BuildARN         string  `json:"BuildArn,omitempty"`
	Name             string  `json:"Name,omitempty"`
	Version          string  `json:"Version,omitempty"`
	Status           string  `json:"Status,omitempty"`
	SizeOnDisk       int64   `json:"SizeOnDisk,omitempty"`
	OperatingSystem  string  `json:"OperatingSystem,omitempty"`
	CreationTime     float64 `json:"CreationTime,omitempty"`
	ServerSdkVersion string  `json:"ServerSdkVersion,omitempty"`
}

// UploadCredentials represents temporary upload credentials.
type UploadCredentials struct {
	AccessKeyId     string `json:"AccessKeyId,omitempty"`
	SecretAccessKey string `json:"SecretAccessKey,omitempty"`
	SessionToken    string `json:"SessionToken,omitempty"`
}

// DescribeBuildRequest represents the DescribeBuild API request.
type DescribeBuildRequest struct {
	BuildId string `json:"BuildId"`
}

// DescribeBuildResponse represents the DescribeBuild API response.
type DescribeBuildResponse struct {
	Build *BuildOutput `json:"Build,omitempty"`
}

// ListBuildsRequest represents the ListBuilds API request.
type ListBuildsRequest struct {
	Status    string `json:"Status,omitempty"`
	Limit     *int32 `json:"Limit,omitempty"`
	NextToken string `json:"NextToken,omitempty"`
}

// ListBuildsResponse represents the ListBuilds API response.
type ListBuildsResponse struct {
	Builds    []BuildOutput `json:"Builds,omitempty"`
	NextToken string        `json:"NextToken,omitempty"`
}

// DeleteBuildRequest represents the DeleteBuild API request.
type DeleteBuildRequest struct {
	BuildId string `json:"BuildId"`
}

// DeleteBuildResponse represents the DeleteBuild API response.
type DeleteBuildResponse struct{}

// CreateFleetRequest represents the CreateFleet API request.
type CreateFleetRequest struct {
	Name                           string                `json:"Name"`
	BuildId                        string                `json:"BuildId,omitempty"`
	ScriptId                       string                `json:"ScriptId,omitempty"`
	Description                    string                `json:"Description,omitempty"`
	EC2InstanceType                string                `json:"EC2InstanceType,omitempty"`
	FleetType                      string                `json:"FleetType,omitempty"`
	ServerLaunchPath               string                `json:"ServerLaunchPath,omitempty"`
	ServerLaunchParameters         string                `json:"ServerLaunchParameters,omitempty"`
	NewGameSessionProtectionPolicy string                `json:"NewGameSessionProtectionPolicy,omitempty"`
	RuntimeConfiguration           *RuntimeConfiguration `json:"RuntimeConfiguration,omitempty"`
	EC2InboundPermissions          []IpPermission        `json:"EC2InboundPermissions,omitempty"`
	Tags                           []TagInput            `json:"Tags,omitempty"`
}

// RuntimeConfiguration represents runtime configuration.
type RuntimeConfiguration struct {
	ServerProcesses                     []ServerProcess `json:"ServerProcesses,omitempty"`
	MaxConcurrentGameSessionActivations *int32          `json:"MaxConcurrentGameSessionActivations,omitempty"`
	GameSessionActivationTimeoutSeconds *int32          `json:"GameSessionActivationTimeoutSeconds,omitempty"`
}

// ServerProcess represents a server process.
type ServerProcess struct {
	LaunchPath           string `json:"LaunchPath"`
	Parameters           string `json:"Parameters,omitempty"`
	ConcurrentExecutions int32  `json:"ConcurrentExecutions"`
}

// IpPermission represents an IP permission.
type IpPermission struct {
	FromPort int32  `json:"FromPort"`
	ToPort   int32  `json:"ToPort"`
	IpRange  string `json:"IpRange"`
	Protocol string `json:"Protocol"`
}

// CreateFleetResponse represents the CreateFleet API response.
type CreateFleetResponse struct {
	FleetAttributes *FleetAttributesOutput `json:"FleetAttributes,omitempty"`
}

// FleetAttributesOutput represents the output format of fleet attributes.
type FleetAttributesOutput struct {
	FleetId                        string  `json:"FleetId,omitempty"`
	FleetArn                       string  `json:"FleetArn,omitempty"`
	FleetType                      string  `json:"FleetType,omitempty"`
	InstanceType                   string  `json:"InstanceType,omitempty"`
	Description                    string  `json:"Description,omitempty"`
	Name                           string  `json:"Name,omitempty"`
	CreationTime                   float64 `json:"CreationTime,omitempty"`
	Status                         string  `json:"Status,omitempty"`
	BuildId                        string  `json:"BuildId,omitempty"`
	BuildArn                       string  `json:"BuildArn,omitempty"`
	ServerLaunchPath               string  `json:"ServerLaunchPath,omitempty"`
	ServerLaunchParameters         string  `json:"ServerLaunchParameters,omitempty"`
	NewGameSessionProtectionPolicy string  `json:"NewGameSessionProtectionPolicy,omitempty"`
	OperatingSystem                string  `json:"OperatingSystem,omitempty"`
}

// DescribeFleetAttributesRequest represents the DescribeFleetAttributes API request.
type DescribeFleetAttributesRequest struct {
	FleetIds  []string `json:"FleetIds,omitempty"`
	Limit     *int32   `json:"Limit,omitempty"`
	NextToken string   `json:"NextToken,omitempty"`
}

// DescribeFleetAttributesResponse represents the DescribeFleetAttributes API response.
type DescribeFleetAttributesResponse struct {
	FleetAttributes []FleetAttributesOutput `json:"FleetAttributes,omitempty"`
	NextToken       string                  `json:"NextToken,omitempty"`
}

// ListFleetsRequest represents the ListFleets API request.
type ListFleetsRequest struct {
	BuildId   string `json:"BuildId,omitempty"`
	ScriptId  string `json:"ScriptId,omitempty"`
	Limit     *int32 `json:"Limit,omitempty"`
	NextToken string `json:"NextToken,omitempty"`
}

// ListFleetsResponse represents the ListFleets API response.
type ListFleetsResponse struct {
	FleetIds  []string `json:"FleetIds,omitempty"`
	NextToken string   `json:"NextToken,omitempty"`
}

// DeleteFleetRequest represents the DeleteFleet API request.
type DeleteFleetRequest struct {
	FleetId string `json:"FleetId"`
}

// DeleteFleetResponse represents the DeleteFleet API response.
type DeleteFleetResponse struct{}

// CreateGameSessionRequest represents the CreateGameSession API request.
type CreateGameSessionRequest struct {
	FleetId                   string         `json:"FleetId,omitempty"`
	AliasId                   string         `json:"AliasId,omitempty"`
	MaximumPlayerSessionCount int32          `json:"MaximumPlayerSessionCount"`
	Name                      string         `json:"Name,omitempty"`
	GameProperties            []GameProperty `json:"GameProperties,omitempty"`
	GameSessionId             string         `json:"GameSessionId,omitempty"`
	IdempotencyToken          string         `json:"IdempotencyToken,omitempty"`
	GameSessionData           string         `json:"GameSessionData,omitempty"`
	Location                  string         `json:"Location,omitempty"`
}

// GameProperty represents a game property.
type GameProperty struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// CreateGameSessionResponse represents the CreateGameSession API response.
type CreateGameSessionResponse struct {
	GameSession *GameSessionOutput `json:"GameSession,omitempty"`
}

// GameSessionOutput represents the output format of a game session.
type GameSessionOutput struct {
	GameSessionId               string         `json:"GameSessionId,omitempty"`
	Name                        string         `json:"Name,omitempty"`
	FleetId                     string         `json:"FleetId,omitempty"`
	FleetArn                    string         `json:"FleetArn,omitempty"`
	CreationTime                float64        `json:"CreationTime,omitempty"`
	TerminationTime             float64        `json:"TerminationTime,omitempty"`
	CurrentPlayerSessionCount   int32          `json:"CurrentPlayerSessionCount,omitempty"`
	MaximumPlayerSessionCount   int32          `json:"MaximumPlayerSessionCount,omitempty"`
	Status                      string         `json:"Status,omitempty"`
	StatusReason                string         `json:"StatusReason,omitempty"`
	GameProperties              []GameProperty `json:"GameProperties,omitempty"`
	IpAddress                   string         `json:"IpAddress,omitempty"`
	DnsName                     string         `json:"DnsName,omitempty"`
	Port                        int32          `json:"Port,omitempty"`
	PlayerSessionCreationPolicy string         `json:"PlayerSessionCreationPolicy,omitempty"`
	CreatorId                   string         `json:"CreatorId,omitempty"`
	GameSessionData             string         `json:"GameSessionData,omitempty"`
	MatchmakerData              string         `json:"MatchmakerData,omitempty"`
	Location                    string         `json:"Location,omitempty"`
}

// DescribeGameSessionsRequest represents the DescribeGameSessions API request.
type DescribeGameSessionsRequest struct {
	FleetId       string `json:"FleetId,omitempty"`
	GameSessionId string `json:"GameSessionId,omitempty"`
	AliasId       string `json:"AliasId,omitempty"`
	Location      string `json:"Location,omitempty"`
	StatusFilter  string `json:"StatusFilter,omitempty"`
	Limit         *int32 `json:"Limit,omitempty"`
	NextToken     string `json:"NextToken,omitempty"`
}

// DescribeGameSessionsResponse represents the DescribeGameSessions API response.
type DescribeGameSessionsResponse struct {
	GameSessions []GameSessionOutput `json:"GameSessions,omitempty"`
	NextToken    string              `json:"NextToken,omitempty"`
}

// UpdateGameSessionRequest represents the UpdateGameSession API request.
type UpdateGameSessionRequest struct {
	GameSessionId               string         `json:"GameSessionId"`
	MaximumPlayerSessionCount   *int32         `json:"MaximumPlayerSessionCount,omitempty"`
	Name                        string         `json:"Name,omitempty"`
	PlayerSessionCreationPolicy string         `json:"PlayerSessionCreationPolicy,omitempty"`
	ProtectionPolicy            string         `json:"ProtectionPolicy,omitempty"`
	GameProperties              []GameProperty `json:"GameProperties,omitempty"`
}

// UpdateGameSessionResponse represents the UpdateGameSession API response.
type UpdateGameSessionResponse struct {
	GameSession *GameSessionOutput `json:"GameSession,omitempty"`
}

// CreatePlayerSessionRequest represents the CreatePlayerSession API request.
type CreatePlayerSessionRequest struct {
	GameSessionId string `json:"GameSessionId"`
	PlayerId      string `json:"PlayerId"`
	PlayerData    string `json:"PlayerData,omitempty"`
}

// CreatePlayerSessionResponse represents the CreatePlayerSession API response.
type CreatePlayerSessionResponse struct {
	PlayerSession *PlayerSessionOutput `json:"PlayerSession,omitempty"`
}

// PlayerSessionOutput represents the output format of a player session.
type PlayerSessionOutput struct {
	PlayerSessionId string  `json:"PlayerSessionId,omitempty"`
	PlayerId        string  `json:"PlayerId,omitempty"`
	GameSessionId   string  `json:"GameSessionId,omitempty"`
	FleetId         string  `json:"FleetId,omitempty"`
	FleetArn        string  `json:"FleetArn,omitempty"`
	CreationTime    float64 `json:"CreationTime,omitempty"`
	TerminationTime float64 `json:"TerminationTime,omitempty"`
	Status          string  `json:"Status,omitempty"`
	IpAddress       string  `json:"IpAddress,omitempty"`
	DnsName         string  `json:"DnsName,omitempty"`
	Port            int32   `json:"Port,omitempty"`
	PlayerData      string  `json:"PlayerData,omitempty"`
}

// CreatePlayerSessionsRequest represents the CreatePlayerSessions API request.
type CreatePlayerSessionsRequest struct {
	GameSessionId string            `json:"GameSessionId"`
	PlayerIds     []string          `json:"PlayerIds"`
	PlayerDataMap map[string]string `json:"PlayerDataMap,omitempty"`
}

// CreatePlayerSessionsResponse represents the CreatePlayerSessions API response.
type CreatePlayerSessionsResponse struct {
	PlayerSessions []PlayerSessionOutput `json:"PlayerSessions,omitempty"`
}

// DescribePlayerSessionsRequest represents the DescribePlayerSessions API request.
type DescribePlayerSessionsRequest struct {
	GameSessionId             string `json:"GameSessionId,omitempty"`
	PlayerId                  string `json:"PlayerId,omitempty"`
	PlayerSessionId           string `json:"PlayerSessionId,omitempty"`
	PlayerSessionStatusFilter string `json:"PlayerSessionStatusFilter,omitempty"`
	Limit                     *int32 `json:"Limit,omitempty"`
	NextToken                 string `json:"NextToken,omitempty"`
}

// DescribePlayerSessionsResponse represents the DescribePlayerSessions API response.
type DescribePlayerSessionsResponse struct {
	PlayerSessions []PlayerSessionOutput `json:"PlayerSessions,omitempty"`
	NextToken      string                `json:"NextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}
