package eventbridge

import (
	"time"
)

// EventBusState represents the state of an event bus.
type EventBusState string

// Event bus states.
const (
	EventBusStateActive  EventBusState = "ACTIVE"
	EventBusStateDeleted EventBusState = "DELETED"
)

// RuleState represents the state of a rule.
type RuleState string

// Rule states.
const (
	RuleStateEnabled  RuleState = "ENABLED"
	RuleStateDisabled RuleState = "DISABLED"
)

// EventBus represents an event bus.
type EventBus struct {
	Name         string
	Arn          string
	Description  string
	CreationTime time.Time
	LastModified time.Time
}

// Rule represents an EventBridge rule.
type Rule struct {
	Name               string
	Arn                string
	EventBusName       string
	EventPattern       string
	ScheduleExpression string
	State              RuleState
	Description        string
	RoleArn            string
	CreationTime       time.Time
	LastModified       time.Time
}

// Target represents a rule target.
type Target struct {
	ID        string
	Arn       string
	RoleArn   string
	Input     string
	InputPath string
}

// PutEventsRequestEntry represents an entry in PutEvents request.
type PutEventsRequestEntry struct {
	Source       string     `json:"Source,omitempty"`
	DetailType   string     `json:"DetailType,omitempty"`
	Detail       string     `json:"Detail,omitempty"`
	EventBusName string     `json:"EventBusName,omitempty"`
	Time         *time.Time `json:"Time,omitempty"`
	Resources    []string   `json:"Resources,omitempty"`
}

// PutEventsResultEntry represents an entry in PutEvents response.
type PutEventsResultEntry struct {
	EventID      string `json:"EventId,omitempty"`
	ErrorCode    string `json:"ErrorCode,omitempty"`
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

// CreateEventBusRequest is the request for CreateEventBus.
type CreateEventBusRequest struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
}

// CreateEventBusResponse is the response for CreateEventBus.
type CreateEventBusResponse struct {
	EventBusArn string `json:"EventBusArn,omitempty"`
}

// DeleteEventBusRequest is the request for DeleteEventBus.
type DeleteEventBusRequest struct {
	Name string `json:"Name"`
}

// DeleteEventBusResponse is the response for DeleteEventBus.
type DeleteEventBusResponse struct{}

// DescribeEventBusRequest is the request for DescribeEventBus.
type DescribeEventBusRequest struct {
	Name string `json:"Name,omitempty"`
}

// DescribeEventBusResponse is the response for DescribeEventBus.
type DescribeEventBusResponse struct {
	Name        string `json:"Name,omitempty"`
	Arn         string `json:"Arn,omitempty"`
	Description string `json:"Description,omitempty"`
}

// ListEventBusesRequest is the request for ListEventBuses.
type ListEventBusesRequest struct {
	NamePrefix string `json:"NamePrefix,omitempty"`
	Limit      int32  `json:"Limit,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// EventBusOutput represents an event bus in API responses.
type EventBusOutput struct {
	Name        string `json:"Name,omitempty"`
	Arn         string `json:"Arn,omitempty"`
	Description string `json:"Description,omitempty"`
}

// ListEventBusesResponse is the response for ListEventBuses.
type ListEventBusesResponse struct {
	EventBuses []EventBusOutput `json:"EventBuses,omitempty"`
	NextToken  string           `json:"NextToken,omitempty"`
}

// PutRuleRequest is the request for PutRule.
type PutRuleRequest struct {
	Name               string `json:"Name"`
	EventBusName       string `json:"EventBusName,omitempty"`
	EventPattern       string `json:"EventPattern,omitempty"`
	ScheduleExpression string `json:"ScheduleExpression,omitempty"`
	State              string `json:"State,omitempty"`
	Description        string `json:"Description,omitempty"`
	RoleArn            string `json:"RoleArn,omitempty"`
}

// PutRuleResponse is the response for PutRule.
type PutRuleResponse struct {
	RuleArn string `json:"RuleArn,omitempty"`
}

// DeleteRuleRequest is the request for DeleteRule.
type DeleteRuleRequest struct {
	Name         string `json:"Name"`
	EventBusName string `json:"EventBusName,omitempty"`
	Force        bool   `json:"Force,omitempty"`
}

// DeleteRuleResponse is the response for DeleteRule.
type DeleteRuleResponse struct{}

// DescribeRuleRequest is the request for DescribeRule.
type DescribeRuleRequest struct {
	Name         string `json:"Name"`
	EventBusName string `json:"EventBusName,omitempty"`
}

// DescribeRuleResponse is the response for DescribeRule.
type DescribeRuleResponse struct {
	Name               string `json:"Name,omitempty"`
	Arn                string `json:"Arn,omitempty"`
	EventBusName       string `json:"EventBusName,omitempty"`
	EventPattern       string `json:"EventPattern,omitempty"`
	ScheduleExpression string `json:"ScheduleExpression,omitempty"`
	State              string `json:"State,omitempty"`
	Description        string `json:"Description,omitempty"`
	RoleArn            string `json:"RoleArn,omitempty"`
}

// ListRulesRequest is the request for ListRules.
type ListRulesRequest struct {
	EventBusName string `json:"EventBusName,omitempty"`
	NamePrefix   string `json:"NamePrefix,omitempty"`
	Limit        int32  `json:"Limit,omitempty"`
	NextToken    string `json:"NextToken,omitempty"`
}

// RuleOutput represents a rule in API responses.
type RuleOutput struct {
	Name               string `json:"Name,omitempty"`
	Arn                string `json:"Arn,omitempty"`
	EventBusName       string `json:"EventBusName,omitempty"`
	EventPattern       string `json:"EventPattern,omitempty"`
	ScheduleExpression string `json:"ScheduleExpression,omitempty"`
	State              string `json:"State,omitempty"`
	Description        string `json:"Description,omitempty"`
	RoleArn            string `json:"RoleArn,omitempty"`
}

// ListRulesResponse is the response for ListRules.
type ListRulesResponse struct {
	Rules     []RuleOutput `json:"Rules,omitempty"`
	NextToken string       `json:"NextToken,omitempty"`
}

// TargetInput represents a target in API requests.
type TargetInput struct {
	ID        string `json:"Id"`
	Arn       string `json:"Arn"`
	RoleArn   string `json:"RoleArn,omitempty"`
	Input     string `json:"Input,omitempty"`
	InputPath string `json:"InputPath,omitempty"`
}

// PutTargetsRequest is the request for PutTargets.
type PutTargetsRequest struct {
	Rule         string        `json:"Rule"`
	EventBusName string        `json:"EventBusName,omitempty"`
	Targets      []TargetInput `json:"Targets"`
}

// PutTargetsResultEntry represents an entry in PutTargets response.
type PutTargetsResultEntry struct {
	TargetID     string `json:"TargetId,omitempty"`
	ErrorCode    string `json:"ErrorCode,omitempty"`
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

// PutTargetsResponse is the response for PutTargets.
type PutTargetsResponse struct {
	FailedEntryCount int32                   `json:"FailedEntryCount"`
	FailedEntries    []PutTargetsResultEntry `json:"FailedEntries,omitempty"`
}

// RemoveTargetsRequest is the request for RemoveTargets.
type RemoveTargetsRequest struct {
	Rule         string   `json:"Rule"`
	EventBusName string   `json:"EventBusName,omitempty"`
	IDs          []string `json:"Ids"`
	Force        bool     `json:"Force,omitempty"`
}

// RemoveTargetsResultEntry represents an entry in RemoveTargets response.
type RemoveTargetsResultEntry struct {
	TargetID     string `json:"TargetId,omitempty"`
	ErrorCode    string `json:"ErrorCode,omitempty"`
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

// RemoveTargetsResponse is the response for RemoveTargets.
type RemoveTargetsResponse struct {
	FailedEntryCount int32                      `json:"FailedEntryCount"`
	FailedEntries    []RemoveTargetsResultEntry `json:"FailedEntries,omitempty"`
}

// PutEventsRequest is the request for PutEvents.
type PutEventsRequest struct {
	Entries []PutEventsRequestEntry `json:"Entries"`
}

// PutEventsResponse is the response for PutEvents.
type PutEventsResponse struct {
	FailedEntryCount int32                  `json:"FailedEntryCount"`
	Entries          []PutEventsResultEntry `json:"Entries,omitempty"`
}

// ListTargetsByRuleRequest is the request for ListTargetsByRule.
type ListTargetsByRuleRequest struct {
	Rule         string `json:"Rule"`
	EventBusName string `json:"EventBusName,omitempty"`
	Limit        int32  `json:"Limit,omitempty"`
	NextToken    string `json:"NextToken,omitempty"`
}

// TargetOutput represents a target in API responses.
type TargetOutput struct {
	ID        string `json:"Id,omitempty"`
	Arn       string `json:"Arn,omitempty"`
	RoleArn   string `json:"RoleArn,omitempty"`
	Input     string `json:"Input,omitempty"`
	InputPath string `json:"InputPath,omitempty"`
}

// ListTargetsByRuleResponse is the response for ListTargetsByRule.
type ListTargetsByRuleResponse struct {
	Targets   []TargetOutput `json:"Targets,omitempty"`
	NextToken string         `json:"NextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}

// ServiceError represents a service-level error.
type ServiceError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *ServiceError) Error() string {
	return e.Message
}
