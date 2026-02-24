package configservice

import "time"

// ConfigurationRecorder represents an AWS Config configuration recorder.
type ConfigurationRecorder struct {
	Name           string
	RoleARN        string
	RecordingGroup *RecordingGroup
}

// RecordingGroup specifies what resources to record.
type RecordingGroup struct {
	AllSupported               bool
	IncludeGlobalResourceTypes bool
	ResourceTypes              []string
}

// ConfigurationRecorderStatus represents recorder status.
type ConfigurationRecorderStatus struct {
	Name          string
	Recording     bool
	LastStatus    string
	LastStartTime *time.Time
	LastStopTime  *time.Time
}

// ConfigRule represents an AWS Config rule.
type ConfigRule struct {
	ConfigRuleName  string
	ConfigRuleARN   string
	ConfigRuleID    string
	Description     string
	Source          *Source
	Scope           *Scope
	ConfigRuleState string
}

// Source specifies the rule source.
type Source struct {
	Owner            string
	SourceIdentifier string
}

// Scope specifies rule scope.
type Scope struct {
	ComplianceResourceTypes []string
}

// EvaluationResult represents a single evaluation result.
type EvaluationResult struct {
	EvaluationResultIdentifier *EvaluationResultIdentifier
	ComplianceType             string
	ResultRecordedTime         time.Time
}

// EvaluationResultIdentifier identifies the evaluation.
type EvaluationResultIdentifier struct {
	EvaluationResultQualifier *EvaluationResultQualifier
}

// EvaluationResultQualifier contains the qualifier details.
type EvaluationResultQualifier struct {
	ConfigRuleName string
	ResourceType   string
	ResourceID     string
}

// Error represents a Config service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// PutConfigurationRecorderRequest represents the PutConfigurationRecorder API request.
type PutConfigurationRecorderRequest struct {
	ConfigurationRecorder *ConfigurationRecorderInput `json:"ConfigurationRecorder"`
}

// ConfigurationRecorderInput represents the input for configuration recorder.
type ConfigurationRecorderInput struct {
	Name           string               `json:"name"`
	RoleARN        string               `json:"roleARN"`
	RecordingGroup *RecordingGroupInput `json:"recordingGroup,omitempty"`
}

// RecordingGroupInput represents the input for recording group.
type RecordingGroupInput struct {
	AllSupported               *bool    `json:"allSupported,omitempty"`
	IncludeGlobalResourceTypes *bool    `json:"includeGlobalResourceTypes,omitempty"`
	ResourceTypes              []string `json:"resourceTypes,omitempty"`
}

// PutConfigurationRecorderResponse represents the PutConfigurationRecorder API response.
type PutConfigurationRecorderResponse struct{}

// DeleteConfigurationRecorderRequest represents the DeleteConfigurationRecorder API request.
type DeleteConfigurationRecorderRequest struct {
	ConfigurationRecorderName string `json:"ConfigurationRecorderName"`
}

// DeleteConfigurationRecorderResponse represents the DeleteConfigurationRecorder API response.
type DeleteConfigurationRecorderResponse struct{}

// DescribeConfigurationRecordersRequest represents the DescribeConfigurationRecorders API request.
type DescribeConfigurationRecordersRequest struct {
	ConfigurationRecorderNames []string `json:"ConfigurationRecorderNames,omitempty"`
}

// DescribeConfigurationRecordersResponse represents the DescribeConfigurationRecorders API response.
type DescribeConfigurationRecordersResponse struct {
	ConfigurationRecorders []ConfigurationRecorderOutput `json:"ConfigurationRecorders"`
}

// ConfigurationRecorderOutput represents the output format of a configuration recorder.
type ConfigurationRecorderOutput struct {
	Name           string                `json:"name"`
	RoleARN        string                `json:"roleARN"`
	RecordingGroup *RecordingGroupOutput `json:"recordingGroup,omitempty"`
}

// RecordingGroupOutput represents the output format of a recording group.
type RecordingGroupOutput struct {
	AllSupported               bool     `json:"allSupported"`
	IncludeGlobalResourceTypes bool     `json:"includeGlobalResourceTypes"`
	ResourceTypes              []string `json:"resourceTypes,omitempty"`
}

// StartConfigurationRecorderRequest represents the StartConfigurationRecorder API request.
type StartConfigurationRecorderRequest struct {
	ConfigurationRecorderName string `json:"ConfigurationRecorderName"`
}

// StartConfigurationRecorderResponse represents the StartConfigurationRecorder API response.
type StartConfigurationRecorderResponse struct{}

// StopConfigurationRecorderRequest represents the StopConfigurationRecorder API request.
type StopConfigurationRecorderRequest struct {
	ConfigurationRecorderName string `json:"ConfigurationRecorderName"`
}

// StopConfigurationRecorderResponse represents the StopConfigurationRecorder API response.
type StopConfigurationRecorderResponse struct{}

// PutConfigRuleRequest represents the PutConfigRule API request.
type PutConfigRuleRequest struct {
	ConfigRule *ConfigRuleInput `json:"ConfigRule"`
}

// ConfigRuleInput represents the input for config rule.
type ConfigRuleInput struct {
	ConfigRuleName string       `json:"ConfigRuleName"`
	Description    string       `json:"Description,omitempty"`
	Source         *SourceInput `json:"Source"`
	Scope          *ScopeInput  `json:"Scope,omitempty"`
}

// SourceInput represents the input for rule source.
type SourceInput struct {
	Owner            string `json:"Owner"`
	SourceIdentifier string `json:"SourceIdentifier"`
}

// ScopeInput represents the input for rule scope.
type ScopeInput struct {
	ComplianceResourceTypes []string `json:"ComplianceResourceTypes,omitempty"`
}

// PutConfigRuleResponse represents the PutConfigRule API response.
type PutConfigRuleResponse struct{}

// DeleteConfigRuleRequest represents the DeleteConfigRule API request.
type DeleteConfigRuleRequest struct {
	ConfigRuleName string `json:"ConfigRuleName"`
}

// DeleteConfigRuleResponse represents the DeleteConfigRule API response.
type DeleteConfigRuleResponse struct{}

// DescribeConfigRulesRequest represents the DescribeConfigRules API request.
type DescribeConfigRulesRequest struct {
	ConfigRuleNames []string `json:"ConfigRuleNames,omitempty"`
	NextToken       string   `json:"NextToken,omitempty"`
}

// DescribeConfigRulesResponse represents the DescribeConfigRules API response.
type DescribeConfigRulesResponse struct {
	ConfigRules []ConfigRuleOutput `json:"ConfigRules"`
	NextToken   string             `json:"NextToken,omitempty"`
}

// ConfigRuleOutput represents the output format of a config rule.
type ConfigRuleOutput struct {
	ConfigRuleName  string        `json:"ConfigRuleName"`
	ConfigRuleARN   string        `json:"ConfigRuleArn"`
	ConfigRuleID    string        `json:"ConfigRuleId"`
	Description     string        `json:"Description,omitempty"`
	Source          *SourceOutput `json:"Source"`
	Scope           *ScopeOutput  `json:"Scope,omitempty"`
	ConfigRuleState string        `json:"ConfigRuleState"`
}

// SourceOutput represents the output format of rule source.
type SourceOutput struct {
	Owner            string `json:"Owner"`
	SourceIdentifier string `json:"SourceIdentifier"`
}

// ScopeOutput represents the output format of rule scope.
type ScopeOutput struct {
	ComplianceResourceTypes []string `json:"ComplianceResourceTypes,omitempty"`
}

// GetComplianceDetailsByConfigRuleRequest represents the GetComplianceDetailsByConfigRule API request.
type GetComplianceDetailsByConfigRuleRequest struct {
	ConfigRuleName  string   `json:"ConfigRuleName"`
	ComplianceTypes []string `json:"ComplianceTypes,omitempty"`
	Limit           *int32   `json:"Limit,omitempty"`
	NextToken       string   `json:"NextToken,omitempty"`
}

// GetComplianceDetailsByConfigRuleResponse represents the GetComplianceDetailsByConfigRule API response.
type GetComplianceDetailsByConfigRuleResponse struct {
	EvaluationResults []EvaluationResultOutput `json:"EvaluationResults"`
	NextToken         string                   `json:"NextToken,omitempty"`
}

// EvaluationResultOutput represents the output format of an evaluation result.
type EvaluationResultOutput struct {
	EvaluationResultIdentifier *EvaluationResultIdentifierOutput `json:"EvaluationResultIdentifier,omitempty"`
	ComplianceType             string                            `json:"ComplianceType"`
	ResultRecordedTime         float64                           `json:"ResultRecordedTime,omitempty"`
}

// EvaluationResultIdentifierOutput represents the output format of evaluation result identifier.
type EvaluationResultIdentifierOutput struct {
	EvaluationResultQualifier *EvaluationResultQualifierOutput `json:"EvaluationResultQualifier,omitempty"`
}

// EvaluationResultQualifierOutput represents the output format of evaluation result qualifier.
type EvaluationResultQualifierOutput struct {
	ConfigRuleName string `json:"ConfigRuleName,omitempty"`
	ResourceType   string `json:"ResourceType,omitempty"`
	ResourceID     string `json:"ResourceId,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}
