// Package scheduler provides EventBridge Scheduler service emulation for awsim.
package scheduler

import "time"

// Schedule state constants.
const (
	StateEnabled  = "ENABLED"
	StateDisabled = "DISABLED"
)

// FlexibleTimeWindowMode constants.
const (
	FlexibleTimeWindowOff      = "OFF"
	FlexibleTimeWindowFlexible = "FLEXIBLE"
)

// ActionAfterCompletion constants.
const (
	ActionNone   = "NONE"
	ActionDelete = "DELETE"
)

// ScheduleGroupState constants.
const (
	ScheduleGroupStateActive   = "ACTIVE"
	ScheduleGroupStateDeleting = "DELETING"
)

// Schedule represents an EventBridge Scheduler schedule.
type Schedule struct {
	Name                       string
	ARN                        string
	GroupName                  string
	Description                string
	ScheduleExpression         string
	ScheduleExpressionTimezone string
	StartDate                  *time.Time
	EndDate                    *time.Time
	State                      string
	FlexibleTimeWindow         *FlexibleTimeWindow
	Target                     *Target
	KmsKeyArn                  string
	ActionAfterCompletion      string
	CreationDate               time.Time
	LastModificationDate       time.Time
}

// ScheduleGroup represents an EventBridge Scheduler schedule group.
type ScheduleGroup struct {
	Name         string
	ARN          string
	State        string
	CreationDate time.Time
}

// FlexibleTimeWindow represents the flexible time window for a schedule.
type FlexibleTimeWindow struct {
	Mode                   string `json:"Mode"`
	MaximumWindowInMinutes *int32 `json:"MaximumWindowInMinutes,omitempty"`
}

// Target represents the target for a schedule.
type Target struct {
	Arn                         string                       `json:"Arn"`
	RoleArn                     string                       `json:"RoleArn"`
	Input                       string                       `json:"Input,omitempty"`
	DeadLetterConfig            *DeadLetterConfig            `json:"DeadLetterConfig,omitempty"`
	RetryPolicy                 *RetryPolicy                 `json:"RetryPolicy,omitempty"`
	EcsParameters               *EcsParameters               `json:"EcsParameters,omitempty"`
	EventBridgeParameters       *EventBridgeParameters       `json:"EventBridgeParameters,omitempty"`
	KinesisParameters           *KinesisParameters           `json:"KinesisParameters,omitempty"`
	SageMakerPipelineParameters *SageMakerPipelineParameters `json:"SageMakerPipelineParameters,omitempty"`
	SqsParameters               *SqsParameters               `json:"SqsParameters,omitempty"`
}

// DeadLetterConfig represents dead letter queue configuration.
type DeadLetterConfig struct {
	Arn string `json:"Arn,omitempty"`
}

// RetryPolicy represents retry policy configuration.
type RetryPolicy struct {
	MaximumEventAgeInSeconds *int32 `json:"MaximumEventAgeInSeconds,omitempty"`
	MaximumRetryAttempts     *int32 `json:"MaximumRetryAttempts,omitempty"`
}

// EcsParameters represents ECS target parameters.
type EcsParameters struct {
	TaskDefinitionArn string `json:"TaskDefinitionArn"`
	TaskCount         *int32 `json:"TaskCount,omitempty"`
	LaunchType        string `json:"LaunchType,omitempty"`
	PlatformVersion   string `json:"PlatformVersion,omitempty"`
}

// EventBridgeParameters represents EventBridge target parameters.
type EventBridgeParameters struct {
	DetailType string `json:"DetailType"`
	Source     string `json:"Source"`
}

// KinesisParameters represents Kinesis target parameters.
type KinesisParameters struct {
	PartitionKey string `json:"PartitionKey"`
}

// SageMakerPipelineParameters represents SageMaker Pipeline target parameters.
type SageMakerPipelineParameters struct {
	PipelineParameterList []PipelineParameter `json:"PipelineParameterList,omitempty"`
}

// PipelineParameter represents a SageMaker Pipeline parameter.
type PipelineParameter struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// SqsParameters represents SQS target parameters.
type SqsParameters struct {
	MessageGroupId string `json:"MessageGroupId,omitempty"`
}

// CreateScheduleRequest represents the CreateSchedule API request.
type CreateScheduleRequest struct {
	ActionAfterCompletion      string              `json:"ActionAfterCompletion,omitempty"`
	ClientToken                string              `json:"ClientToken,omitempty"`
	Description                string              `json:"Description,omitempty"`
	EndDate                    *string             `json:"EndDate,omitempty"`
	FlexibleTimeWindow         *FlexibleTimeWindow `json:"FlexibleTimeWindow"`
	GroupName                  string              `json:"GroupName,omitempty"`
	KmsKeyArn                  string              `json:"KmsKeyArn,omitempty"`
	ScheduleExpression         string              `json:"ScheduleExpression"`
	ScheduleExpressionTimezone string              `json:"ScheduleExpressionTimezone,omitempty"`
	StartDate                  *string             `json:"StartDate,omitempty"`
	State                      string              `json:"State,omitempty"`
	Target                     *Target             `json:"Target"`
}

// CreateScheduleResponse represents the CreateSchedule API response.
type CreateScheduleResponse struct {
	ScheduleArn string `json:"ScheduleArn"`
}

// GetScheduleResponse represents the GetSchedule API response.
type GetScheduleResponse struct {
	ActionAfterCompletion      string              `json:"ActionAfterCompletion,omitempty"`
	Arn                        string              `json:"Arn"`
	CreationDate               float64             `json:"CreationDate"`
	Description                string              `json:"Description,omitempty"`
	EndDate                    *string             `json:"EndDate,omitempty"`
	FlexibleTimeWindow         *FlexibleTimeWindow `json:"FlexibleTimeWindow"`
	GroupName                  string              `json:"GroupName"`
	KmsKeyArn                  string              `json:"KmsKeyArn,omitempty"`
	LastModificationDate       float64             `json:"LastModificationDate"`
	Name                       string              `json:"Name"`
	ScheduleExpression         string              `json:"ScheduleExpression"`
	ScheduleExpressionTimezone string              `json:"ScheduleExpressionTimezone"`
	StartDate                  *string             `json:"StartDate,omitempty"`
	State                      string              `json:"State"`
	Target                     *Target             `json:"Target"`
}

// UpdateScheduleRequest represents the UpdateSchedule API request.
type UpdateScheduleRequest struct {
	ActionAfterCompletion      string              `json:"ActionAfterCompletion,omitempty"`
	ClientToken                string              `json:"ClientToken,omitempty"`
	Description                string              `json:"Description,omitempty"`
	EndDate                    *string             `json:"EndDate,omitempty"`
	FlexibleTimeWindow         *FlexibleTimeWindow `json:"FlexibleTimeWindow"`
	GroupName                  string              `json:"GroupName,omitempty"`
	KmsKeyArn                  string              `json:"KmsKeyArn,omitempty"`
	ScheduleExpression         string              `json:"ScheduleExpression"`
	ScheduleExpressionTimezone string              `json:"ScheduleExpressionTimezone,omitempty"`
	StartDate                  *string             `json:"StartDate,omitempty"`
	State                      string              `json:"State,omitempty"`
	Target                     *Target             `json:"Target"`
}

// UpdateScheduleResponse represents the UpdateSchedule API response.
type UpdateScheduleResponse struct {
	ScheduleArn string `json:"ScheduleArn"`
}

// ListSchedulesResponse represents the ListSchedules API response.
type ListSchedulesResponse struct {
	NextToken string            `json:"NextToken,omitempty"`
	Schedules []ScheduleSummary `json:"Schedules"`
}

// ScheduleSummary represents a schedule summary for list operations.
type ScheduleSummary struct {
	Arn                  string         `json:"Arn"`
	CreationDate         float64        `json:"CreationDate"`
	GroupName            string         `json:"GroupName"`
	LastModificationDate float64        `json:"LastModificationDate"`
	Name                 string         `json:"Name"`
	State                string         `json:"State"`
	Target               *TargetSummary `json:"Target,omitempty"`
}

// TargetSummary represents a target summary for list operations.
type TargetSummary struct {
	Arn string `json:"Arn"`
}

// CreateScheduleGroupRequest represents the CreateScheduleGroup API request.
type CreateScheduleGroupRequest struct {
	ClientToken string `json:"ClientToken,omitempty"`
	Name        string `json:"Name"`
	Tags        []Tag  `json:"Tags,omitempty"`
}

// CreateScheduleGroupResponse represents the CreateScheduleGroup API response.
type CreateScheduleGroupResponse struct {
	ScheduleGroupArn string `json:"ScheduleGroupArn"`
}

// GetScheduleGroupResponse represents the GetScheduleGroup API response.
type GetScheduleGroupResponse struct {
	Arn                  string  `json:"Arn"`
	CreationDate         float64 `json:"CreationDate"`
	LastModificationDate float64 `json:"LastModificationDate"`
	Name                 string  `json:"Name"`
	State                string  `json:"State"`
}

// ListScheduleGroupsResponse represents the ListScheduleGroups API response.
type ListScheduleGroupsResponse struct {
	NextToken      string                 `json:"NextToken,omitempty"`
	ScheduleGroups []ScheduleGroupSummary `json:"ScheduleGroups"`
}

// ScheduleGroupSummary represents a schedule group summary for list operations.
type ScheduleGroupSummary struct {
	Arn                  string  `json:"Arn"`
	CreationDate         float64 `json:"CreationDate"`
	LastModificationDate float64 `json:"LastModificationDate"`
	Name                 string  `json:"Name"`
	State                string  `json:"State"`
}

// Tag represents a resource tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"Message"`
}

// Error represents a Scheduler service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
