// Package mq provides Amazon MQ service emulation for awsim.
package mq

import (
	"time"
)

// Broker deployment modes.
const (
	DeploymentModeSingleInstance = "SINGLE_INSTANCE"
	DeploymentModeActiveStandby  = "ACTIVE_STANDBY_MULTI_AZ"
	DeploymentModeClusterMultiAZ = "CLUSTER_MULTI_AZ"
)

// Broker engine types.
const (
	EngineTypeActiveMQ = "ACTIVEMQ"
	EngineTypeRabbitMQ = "RABBITMQ"
)

// Broker states.
const (
	BrokerStateCreating      = "CREATING"
	BrokerStateRunning       = "RUNNING"
	BrokerStateRebooting     = "REBOOTING"
	BrokerStateDeletionInPrg = "DELETION_IN_PROGRESS"
)

// Configuration revision states.
const (
	ConfigRevisionActive = "ACTIVE"
)

// Broker represents an Amazon MQ broker.
type Broker struct {
	BrokerID             string
	BrokerName           string
	BrokerArn            string
	BrokerState          string
	Created              time.Time
	DeploymentMode       string
	EngineType           string
	EngineVersion        string
	HostInstanceType     string
	AutoMinorVersionUpgr bool
	PubliclyAccessible   bool
	Users                []*User
	Tags                 map[string]string
	Configuration        *ConfigurationID
}

// User represents a broker user.
type User struct {
	Username string
	Password string
	Groups   []string
}

// ConfigurationID references a configuration.
type ConfigurationID struct {
	ID       string
	Revision int
}

// Configuration represents a broker configuration.
type Configuration struct {
	ID             string
	Arn            string
	Name           string
	EngineType     string
	EngineVersion  string
	Description    string
	Created        time.Time
	LatestRevision *ConfigurationRevision
	Tags           map[string]string
	Revisions      []*ConfigurationRevision
}

// ConfigurationRevision represents a configuration revision.
type ConfigurationRevision struct {
	Revision    int
	Created     time.Time
	Description string
	Data        string
}

// CreateBrokerRequest is the request for CreateBroker.
type CreateBrokerRequest struct {
	BrokerName           string            `json:"brokerName"`
	EngineType           string            `json:"engineType"`
	EngineVersion        string            `json:"engineVersion"`
	HostInstanceType     string            `json:"hostInstanceType"`
	DeploymentMode       string            `json:"deploymentMode"`
	Users                []*UserInput      `json:"users"`
	AutoMinorVersionUpgr bool              `json:"autoMinorVersionUpgrade"`
	PubliclyAccessible   bool              `json:"publiclyAccessible"`
	Tags                 map[string]string `json:"tags,omitempty"`
	Configuration        *ConfigurationID  `json:"configuration,omitempty"`
}

// UserInput is the input for creating a user.
type UserInput struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Groups   []string `json:"groups,omitempty"`
}

// CreateBrokerResponse is the response for CreateBroker.
type CreateBrokerResponse struct {
	BrokerArn string `json:"brokerArn,omitempty"`
	BrokerID  string `json:"brokerId,omitempty"`
}

// DeleteBrokerRequest is the request for DeleteBroker.
type DeleteBrokerRequest struct {
	BrokerID string `json:"brokerId"`
}

// DeleteBrokerResponse is the response for DeleteBroker.
type DeleteBrokerResponse struct {
	BrokerID string `json:"brokerId,omitempty"`
}

// DescribeBrokerRequest is the request for DescribeBroker.
type DescribeBrokerRequest struct {
	BrokerID string `json:"brokerId"`
}

// DescribeBrokerResponse is the response for DescribeBroker.
type DescribeBrokerResponse struct {
	BrokerArn            string                  `json:"brokerArn,omitempty"`
	BrokerID             string                  `json:"brokerId,omitempty"`
	BrokerName           string                  `json:"brokerName,omitempty"`
	BrokerState          string                  `json:"brokerState,omitempty"`
	Created              string                  `json:"created,omitempty"`
	DeploymentMode       string                  `json:"deploymentMode,omitempty"`
	EngineType           string                  `json:"engineType,omitempty"`
	EngineVersion        string                  `json:"engineVersion,omitempty"`
	HostInstanceType     string                  `json:"hostInstanceType,omitempty"`
	AutoMinorVersionUpgr bool                    `json:"autoMinorVersionUpgrade"`
	PubliclyAccessible   bool                    `json:"publiclyAccessible"`
	Users                []*UserSummary          `json:"users,omitempty"`
	Tags                 map[string]string       `json:"tags,omitempty"`
	Configurations       *ConfigurationsResponse `json:"configurations,omitempty"`
	BrokerInstances      []*BrokerInstance       `json:"brokerInstances,omitempty"`
}

// UserSummary is a summary of a user.
type UserSummary struct {
	Username string `json:"username"`
}

// ConfigurationsResponse contains configuration info for a broker.
type ConfigurationsResponse struct {
	Current *ConfigurationIDResponse `json:"current,omitempty"`
}

// ConfigurationIDResponse is a configuration reference in responses.
type ConfigurationIDResponse struct {
	ID       string `json:"id,omitempty"`
	Revision int    `json:"revision,omitempty"`
}

// BrokerInstance represents a broker instance.
type BrokerInstance struct {
	ConsoleURL string   `json:"consoleUrl,omitempty"`
	Endpoints  []string `json:"endpoints,omitempty"`
}

// ListBrokersRequest is the request for ListBrokers.
type ListBrokersRequest struct {
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

// ListBrokersResponse is the response for ListBrokers.
type ListBrokersResponse struct {
	BrokerSummaries []*BrokerSummary `json:"brokerSummaries,omitempty"`
	NextToken       string           `json:"nextToken,omitempty"`
}

// BrokerSummary is a summary of a broker.
type BrokerSummary struct {
	BrokerArn        string `json:"brokerArn,omitempty"`
	BrokerID         string `json:"brokerId,omitempty"`
	BrokerName       string `json:"brokerName,omitempty"`
	BrokerState      string `json:"brokerState,omitempty"`
	Created          string `json:"created,omitempty"`
	DeploymentMode   string `json:"deploymentMode,omitempty"`
	EngineType       string `json:"engineType,omitempty"`
	HostInstanceType string `json:"hostInstanceType,omitempty"`
}

// CreateConfigurationRequest is the request for CreateConfiguration.
type CreateConfigurationRequest struct {
	Name          string            `json:"name"`
	EngineType    string            `json:"engineType"`
	EngineVersion string            `json:"engineVersion"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// CreateConfigurationResponse is the response for CreateConfiguration.
type CreateConfigurationResponse struct {
	Arn            string                     `json:"arn,omitempty"`
	Created        string                     `json:"created,omitempty"`
	ID             string                     `json:"id,omitempty"`
	Name           string                     `json:"name,omitempty"`
	LatestRevision *ConfigurationRevisionResp `json:"latestRevision,omitempty"`
}

// ConfigurationRevisionResp is a revision in responses.
type ConfigurationRevisionResp struct {
	Created     string `json:"created,omitempty"`
	Description string `json:"description,omitempty"`
	Revision    int    `json:"revision,omitempty"`
}

// UpdateBrokerRequest is the request for UpdateBroker.
type UpdateBrokerRequest struct {
	BrokerID             string           `json:"brokerId"`
	EngineVersion        string           `json:"engineVersion,omitempty"`
	HostInstanceType     string           `json:"hostInstanceType,omitempty"`
	AutoMinorVersionUpgr *bool            `json:"autoMinorVersionUpgrade,omitempty"`
	Configuration        *ConfigurationID `json:"configuration,omitempty"`
}

// UpdateBrokerResponse is the response for UpdateBroker.
type UpdateBrokerResponse struct {
	BrokerID             string                   `json:"brokerId,omitempty"`
	EngineVersion        string                   `json:"engineVersion,omitempty"`
	HostInstanceType     string                   `json:"hostInstanceType,omitempty"`
	AutoMinorVersionUpgr bool                     `json:"autoMinorVersionUpgrade"`
	Configuration        *ConfigurationIDResponse `json:"configuration,omitempty"`
}

// Error represents an MQ error.
type Error struct {
	Type    string `json:"__type"` //nolint:tagliatelle // AWS MQ API uses __type
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Error codes for MQ.
const (
	ErrBadRequest     = "BadRequestException"
	ErrNotFound       = "NotFoundException"
	ErrConflict       = "ConflictException"
	ErrInternalServer = "InternalServerErrorException"
	ErrForbidden      = "ForbiddenException"
)
