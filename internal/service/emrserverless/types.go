package emrserverless

import (
	"encoding/json"
	"fmt"
	"time"
)

// Application states.
const (
	ApplicationStateCreating   = "CREATING"
	ApplicationStateCreated    = "CREATED"
	ApplicationStateStarting   = "STARTING"
	ApplicationStateStarted    = "STARTED"
	ApplicationStateStopping   = "STOPPING"
	ApplicationStateStopped    = "STOPPED"
	ApplicationStateTerminated = "TERMINATED"
)

// JobRun states.
const (
	JobRunStateSubmitted  = "SUBMITTED"
	JobRunStatePending    = "PENDING"
	JobRunStateScheduled  = "SCHEDULED"
	JobRunStateRunning    = "RUNNING"
	JobRunStateSuccess    = "SUCCESS"
	JobRunStateFailed     = "FAILED"
	JobRunStateCancelling = "CANCELLING"
	JobRunStateCancelled  = "CANCELLED"
	JobRunStateQueued     = "QUEUED"
)

// JobRun modes.
const (
	JobRunModeBatch     = "BATCH"
	JobRunModeStreaming = "STREAMING"
)

// Architecture types.
const (
	ArchitectureARM64 = "ARM64"
	ArchitectureX8664 = "X86_64"
)

// Error codes.
const (
	errValidationException  = "ValidationException"
	errResourceNotFound     = "ResourceNotFoundException"
	errConflictException    = "ConflictException"
	errInternalServerError  = "InternalServerException"
	errServiceQuotaExceeded = "ServiceQuotaExceededException"
)

// Pagination defaults.
const (
	defaultPageLimit = 20
	maxPageLimit     = 50
)

// AWSTimestamp wraps time.Time for AWS-style JSON serialization (Unix epoch float64).
type AWSTimestamp struct {
	time.Time
}

// MarshalJSON serializes time to Unix epoch float64.
func (t AWSTimestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		data, err := json.Marshal(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal nil timestamp: %w", err)
		}

		return data, nil
	}

	data, err := json.Marshal(float64(t.Unix()) + float64(t.Nanosecond())/1e9)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal timestamp: %w", err)
	}

	return data, nil
}

// UnmarshalJSON deserializes Unix epoch float64 to time.Time.
func (t *AWSTimestamp) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("failed to unmarshal timestamp: %w", err)
	}

	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	t.Time = time.Unix(sec, nsec)

	return nil
}

// Application represents an EMR Serverless application.
type Application struct {
	ApplicationID           string                      `json:"applicationId"`
	Arn                     string                      `json:"arn"`
	Name                    string                      `json:"name,omitempty"`
	Type                    string                      `json:"type"`
	ReleaseLabel            string                      `json:"releaseLabel"`
	State                   string                      `json:"state"`
	StateDetails            string                      `json:"stateDetails,omitempty"`
	Architecture            string                      `json:"architecture,omitempty"`
	InitialCapacity         map[string]*InitialCapacity `json:"initialCapacity,omitempty"`
	MaximumCapacity         *MaximumCapacity            `json:"maximumCapacity,omitempty"`
	AutoStartConfiguration  *AutoStartConfiguration     `json:"autoStartConfiguration,omitempty"`
	AutoStopConfiguration   *AutoStopConfiguration      `json:"autoStopConfiguration,omitempty"`
	NetworkConfiguration    *NetworkConfiguration       `json:"networkConfiguration,omitempty"`
	MonitoringConfiguration *MonitoringConfiguration    `json:"monitoringConfiguration,omitempty"`
	Tags                    map[string]string           `json:"tags,omitempty"`
	CreatedAt               AWSTimestamp                `json:"createdAt"`
	UpdatedAt               AWSTimestamp                `json:"updatedAt"`
}

// InitialCapacity represents initial capacity configuration.
type InitialCapacity struct {
	WorkerCount         int64                `json:"workerCount"`
	WorkerConfiguration *WorkerConfiguration `json:"workerConfiguration,omitempty"`
}

// WorkerConfiguration represents worker configuration.
type WorkerConfiguration struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Disk   string `json:"disk,omitempty"`
}

// MaximumCapacity represents maximum capacity configuration.
type MaximumCapacity struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Disk   string `json:"disk,omitempty"`
}

// AutoStartConfiguration represents auto start configuration.
type AutoStartConfiguration struct {
	Enabled bool `json:"enabled"`
}

// AutoStopConfiguration represents auto stop configuration.
type AutoStopConfiguration struct {
	Enabled            bool  `json:"enabled"`
	IdleTimeoutMinutes int32 `json:"idleTimeoutMinutes,omitempty"`
}

// NetworkConfiguration represents network configuration.
type NetworkConfiguration struct {
	SubnetIDs        []string `json:"subnetIds,omitempty"`
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`
}

// MonitoringConfiguration represents monitoring configuration.
type MonitoringConfiguration struct {
	S3MonitoringConfiguration                 *S3MonitoringConfiguration                 `json:"s3MonitoringConfiguration,omitempty"`
	ManagedPersistenceMonitoringConfiguration *ManagedPersistenceMonitoringConfiguration `json:"managedPersistenceMonitoringConfiguration,omitempty"`
	CloudWatchLoggingConfiguration            *CloudWatchLoggingConfiguration            `json:"cloudWatchLoggingConfiguration,omitempty"`
}

// S3MonitoringConfiguration represents S3 monitoring configuration.
type S3MonitoringConfiguration struct {
	LogURI           string `json:"logUri,omitempty"`
	EncryptionKeyArn string `json:"encryptionKeyArn,omitempty"`
}

// ManagedPersistenceMonitoringConfiguration represents managed persistence monitoring.
type ManagedPersistenceMonitoringConfiguration struct {
	Enabled          bool   `json:"enabled"`
	EncryptionKeyArn string `json:"encryptionKeyArn,omitempty"`
}

// CloudWatchLoggingConfiguration represents CloudWatch logging configuration.
type CloudWatchLoggingConfiguration struct {
	Enabled             bool                `json:"enabled"`
	LogGroupName        string              `json:"logGroupName,omitempty"`
	LogStreamNamePrefix string              `json:"logStreamNamePrefix,omitempty"`
	EncryptionKeyArn    string              `json:"encryptionKeyArn,omitempty"`
	LogTypes            map[string][]string `json:"logTypes,omitempty"`
}

// ApplicationSummary represents a summary of an EMR Serverless application.
type ApplicationSummary struct {
	ApplicationID string       `json:"id"`
	Arn           string       `json:"arn"`
	Name          string       `json:"name,omitempty"`
	Type          string       `json:"type"`
	ReleaseLabel  string       `json:"releaseLabel"`
	State         string       `json:"state"`
	StateDetails  string       `json:"stateDetails,omitempty"`
	Architecture  string       `json:"architecture,omitempty"`
	CreatedAt     AWSTimestamp `json:"createdAt"`
	UpdatedAt     AWSTimestamp `json:"updatedAt"`
}

// JobRun represents an EMR Serverless job run.
type JobRun struct {
	ApplicationID                 string                    `json:"applicationId"`
	JobRunID                      string                    `json:"jobRunId"`
	Arn                           string                    `json:"arn"`
	Name                          string                    `json:"name,omitempty"`
	State                         string                    `json:"state"`
	StateDetails                  string                    `json:"stateDetails,omitempty"`
	Mode                          string                    `json:"mode,omitempty"`
	ReleaseLabel                  string                    `json:"releaseLabel"`
	ExecutionRole                 string                    `json:"executionRole"`
	JobDriver                     *JobDriver                `json:"jobDriver"`
	ConfigurationOverrides        *ConfigurationOverrides   `json:"configurationOverrides,omitempty"`
	Tags                          map[string]string         `json:"tags,omitempty"`
	TotalResourceUtilization      *TotalResourceUtilization `json:"totalResourceUtilization,omitempty"`
	TotalExecutionDurationSeconds int64                     `json:"totalExecutionDurationSeconds,omitempty"`
	ExecutionTimeoutMinutes       int64                     `json:"executionTimeoutMinutes,omitempty"`
	CreatedAt                     AWSTimestamp              `json:"createdAt"`
	UpdatedAt                     AWSTimestamp              `json:"updatedAt"`
	CreatedBy                     string                    `json:"createdBy"`
}

// JobDriver represents the job driver configuration.
type JobDriver struct {
	SparkSubmit *SparkSubmit `json:"sparkSubmit,omitempty"`
	Hive        *Hive        `json:"hive,omitempty"`
}

// SparkSubmit represents Spark submit configuration.
type SparkSubmit struct {
	EntryPoint            string   `json:"entryPoint"`
	EntryPointArguments   []string `json:"entryPointArguments,omitempty"`
	SparkSubmitParameters string   `json:"sparkSubmitParameters,omitempty"`
}

// Hive represents Hive job configuration.
type Hive struct {
	Query         string `json:"query"`
	InitQueryFile string `json:"initQueryFile,omitempty"`
	Parameters    string `json:"parameters,omitempty"`
}

// ConfigurationOverrides represents configuration overrides.
type ConfigurationOverrides struct {
	ApplicationConfiguration []*Configuration         `json:"applicationConfiguration,omitempty"`
	MonitoringConfiguration  *MonitoringConfiguration `json:"monitoringConfiguration,omitempty"`
}

// Configuration represents a configuration entry.
type Configuration struct {
	Classification string            `json:"classification"`
	Properties     map[string]string `json:"properties,omitempty"`
	Configurations []*Configuration  `json:"configurations,omitempty"`
}

// TotalResourceUtilization represents resource utilization.
type TotalResourceUtilization struct {
	VCPUHour      float64 `json:"vCPUHour,omitempty"`
	MemoryGBHour  float64 `json:"memoryGBHour,omitempty"`
	StorageGBHour float64 `json:"storageGBHour,omitempty"`
}

// JobRunSummary represents a summary of a job run.
type JobRunSummary struct {
	ApplicationID string       `json:"applicationId"`
	JobRunID      string       `json:"id"`
	Arn           string       `json:"arn"`
	Name          string       `json:"name,omitempty"`
	State         string       `json:"state"`
	StateDetails  string       `json:"stateDetails,omitempty"`
	Mode          string       `json:"mode,omitempty"`
	ReleaseLabel  string       `json:"releaseLabel"`
	Type          string       `json:"type,omitempty"`
	CreatedAt     AWSTimestamp `json:"createdAt"`
	UpdatedAt     AWSTimestamp `json:"updatedAt"`
	CreatedBy     string       `json:"createdBy"`
}

// CreateApplicationInput represents the input for CreateApplication.
type CreateApplicationInput struct {
	Name                    string                      `json:"name,omitempty"`
	Type                    string                      `json:"type"`
	ReleaseLabel            string                      `json:"releaseLabel"`
	Architecture            string                      `json:"architecture,omitempty"`
	ClientToken             string                      `json:"clientToken,omitempty"`
	InitialCapacity         map[string]*InitialCapacity `json:"initialCapacity,omitempty"`
	MaximumCapacity         *MaximumCapacity            `json:"maximumCapacity,omitempty"`
	AutoStartConfiguration  *AutoStartConfiguration     `json:"autoStartConfiguration,omitempty"`
	AutoStopConfiguration   *AutoStopConfiguration      `json:"autoStopConfiguration,omitempty"`
	NetworkConfiguration    *NetworkConfiguration       `json:"networkConfiguration,omitempty"`
	MonitoringConfiguration *MonitoringConfiguration    `json:"monitoringConfiguration,omitempty"`
	Tags                    map[string]string           `json:"tags,omitempty"`
}

// CreateApplicationOutput represents the output for CreateApplication.
type CreateApplicationOutput struct {
	ApplicationID string `json:"applicationId"`
	Arn           string `json:"arn"`
	Name          string `json:"name,omitempty"`
}

// GetApplicationOutput represents the output for GetApplication.
type GetApplicationOutput struct {
	Application *Application `json:"application"`
}

// ListApplicationsInput represents the input for ListApplications.
type ListApplicationsInput struct {
	MaxResults int32    `json:"maxResults,omitempty"`
	NextToken  string   `json:"nextToken,omitempty"`
	States     []string `json:"states,omitempty"`
}

// ListApplicationsOutput represents the output for ListApplications.
type ListApplicationsOutput struct {
	Applications []*ApplicationSummary `json:"applications"`
	NextToken    string                `json:"nextToken,omitempty"`
}

// UpdateApplicationInput represents the input for UpdateApplication.
type UpdateApplicationInput struct {
	ApplicationID           string                      `json:"-"`
	Architecture            string                      `json:"architecture,omitempty"`
	AutoStartConfiguration  *AutoStartConfiguration     `json:"autoStartConfiguration,omitempty"`
	AutoStopConfiguration   *AutoStopConfiguration      `json:"autoStopConfiguration,omitempty"`
	InitialCapacity         map[string]*InitialCapacity `json:"initialCapacity,omitempty"`
	MaximumCapacity         *MaximumCapacity            `json:"maximumCapacity,omitempty"`
	NetworkConfiguration    *NetworkConfiguration       `json:"networkConfiguration,omitempty"`
	MonitoringConfiguration *MonitoringConfiguration    `json:"monitoringConfiguration,omitempty"`
	ReleaseLabel            string                      `json:"releaseLabel,omitempty"`
}

// UpdateApplicationOutput represents the output for UpdateApplication.
type UpdateApplicationOutput struct {
	Application *Application `json:"application"`
}

// StartJobRunInput represents the input for StartJobRun.
type StartJobRunInput struct {
	ApplicationID           string                  `json:"-"`
	ClientToken             string                  `json:"clientToken,omitempty"`
	Name                    string                  `json:"name,omitempty"`
	ExecutionRoleArn        string                  `json:"executionRoleArn"`
	JobDriver               *JobDriver              `json:"jobDriver"`
	ConfigurationOverrides  *ConfigurationOverrides `json:"configurationOverrides,omitempty"`
	Tags                    map[string]string       `json:"tags,omitempty"`
	ExecutionTimeoutMinutes int64                   `json:"executionTimeoutMinutes,omitempty"`
	Mode                    string                  `json:"mode,omitempty"`
}

// StartJobRunOutput represents the output for StartJobRun.
type StartJobRunOutput struct {
	ApplicationID string `json:"applicationId"`
	JobRunID      string `json:"jobRunId"`
	Arn           string `json:"arn"`
}

// GetJobRunOutput represents the output for GetJobRun.
type GetJobRunOutput struct {
	JobRun *JobRun `json:"jobRun"`
}

// ListJobRunsInput represents the input for ListJobRuns.
type ListJobRunsInput struct {
	ApplicationID   string   `json:"-"`
	MaxResults      int32    `json:"maxResults,omitempty"`
	NextToken       string   `json:"nextToken,omitempty"`
	States          []string `json:"states,omitempty"`
	Mode            string   `json:"mode,omitempty"`
	CreatedAtBefore string   `json:"createdAtBefore,omitempty"`
	CreatedAtAfter  string   `json:"createdAtAfter,omitempty"`
}

// ListJobRunsOutput represents the output for ListJobRuns.
type ListJobRunsOutput struct {
	JobRuns   []*JobRunSummary `json:"jobRuns"`
	NextToken string           `json:"nextToken,omitempty"`
}

// CancelJobRunOutput represents the output for CancelJobRun.
type CancelJobRunOutput struct {
	ApplicationID string `json:"applicationId"`
	JobRunID      string `json:"jobRunId"`
}

// ErrorResponse represents an AWS error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}
