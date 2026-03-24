// Package cloudformation provides CloudFormation service emulation for kumo.
package cloudformation

import (
	"encoding/xml"
	"time"
)

const cfnXMLNS = "http://cloudformation.amazonaws.com/doc/2010-05-15/"

// Stack status constants.
const (
	StackStatusCreateInProgress = "CREATE_IN_PROGRESS"
	StackStatusCreateComplete   = "CREATE_COMPLETE"
	StackStatusCreateFailed     = "CREATE_FAILED"
	StackStatusUpdateInProgress = "UPDATE_IN_PROGRESS"
	StackStatusUpdateComplete   = "UPDATE_COMPLETE"
	StackStatusUpdateFailed     = "UPDATE_FAILED"
	StackStatusDeleteInProgress = "DELETE_IN_PROGRESS"
	StackStatusDeleteComplete   = "DELETE_COMPLETE"
	StackStatusDeleteFailed     = "DELETE_FAILED"
)

// Resource status constants.
const (
	ResourceStatusCreateComplete = "CREATE_COMPLETE"
	ResourceStatusDeleteComplete = "DELETE_COMPLETE"
)

// Stack represents a CloudFormation stack.
type Stack struct {
	StackID           string
	StackName         string
	TemplateBody      string
	Parameters        map[string]string
	StackStatus       string
	StackStatusReason string
	CreationTime      time.Time
	LastUpdatedTime   time.Time
	DeletionTime      time.Time
	Resources         []StackResource
}

// StackResource represents a resource in a stack.
type StackResource struct {
	LogicalResourceID  string
	PhysicalResourceID string
	ResourceType       string
	ResourceStatus     string
	Timestamp          time.Time
	StackID            string
	StackName          string
}

// TemplateValidationResult represents the result of template validation.
type TemplateValidationResult struct {
	Parameters   []TemplateParameter
	Description  string
	Capabilities []string
}

// TemplateParameter represents a parameter defined in a template.
type TemplateParameter struct {
	ParameterKey   string
	DefaultValue   string
	NoEcho         bool
	Description    string
	ParameterType  string
	AllowedValues  []string
	AllowedPattern string
}

// Request types.

// CreateStackRequest represents a CreateStack request.
type CreateStackRequest struct {
	StackName    string            `json:"StackName"`
	TemplateBody string            `json:"TemplateBody,omitempty"`
	TemplateURL  string            `json:"TemplateURL,omitempty"`
	Parameters   map[string]string `json:"Parameters,omitempty"`
}

// DeleteStackRequest represents a DeleteStack request.
type DeleteStackRequest struct {
	StackName string `json:"StackName"`
}

// DescribeStacksRequest represents a DescribeStacks request.
type DescribeStacksRequest struct {
	StackName string `json:"StackName,omitempty"`
}

// ListStacksRequest represents a ListStacks request.
type ListStacksRequest struct {
	StackStatusFilter []string `json:"StackStatusFilter,omitempty"`
}

// UpdateStackRequest represents an UpdateStack request.
type UpdateStackRequest struct {
	StackName    string            `json:"StackName"`
	TemplateBody string            `json:"TemplateBody,omitempty"`
	Parameters   map[string]string `json:"Parameters,omitempty"`
}

// DescribeStackResourcesRequest represents a DescribeStackResources request.
type DescribeStackResourcesRequest struct {
	StackName         string `json:"StackName,omitempty"`
	LogicalResourceID string `json:"LogicalResourceId,omitempty"`
}

// GetTemplateRequest represents a GetTemplate request.
type GetTemplateRequest struct {
	StackName string `json:"StackName"`
}

// ValidateTemplateRequest represents a ValidateTemplate request.
type ValidateTemplateRequest struct {
	TemplateBody string `json:"TemplateBody,omitempty"`
	TemplateURL  string `json:"TemplateURL,omitempty"`
}

// XML Response types.

// XMLCreateStackResponse is the XML response for CreateStack.
type XMLCreateStackResponse struct {
	XMLName          xml.Name             `xml:"CreateStackResponse"`
	Xmlns            string               `xml:"xmlns,attr"`
	Result           XMLCreateStackResult `xml:"CreateStackResult"`
	ResponseMetadata XMLResponseMetadata  `xml:"ResponseMetadata"`
}

// XMLCreateStackResult contains the result of CreateStack.
type XMLCreateStackResult struct {
	StackID string `xml:"StackId"`
}

// XMLDeleteStackResponse is the XML response for DeleteStack.
type XMLDeleteStackResponse struct {
	XMLName          xml.Name             `xml:"DeleteStackResponse"`
	Xmlns            string               `xml:"xmlns,attr"`
	Result           XMLDeleteStackResult `xml:"DeleteStackResult"`
	ResponseMetadata XMLResponseMetadata  `xml:"ResponseMetadata"`
}

// XMLDeleteStackResult is an empty result for DeleteStack.
type XMLDeleteStackResult struct{}

// XMLDescribeStacksResponse is the XML response for DescribeStacks.
type XMLDescribeStacksResponse struct {
	XMLName          xml.Name                `xml:"DescribeStacksResponse"`
	Xmlns            string                  `xml:"xmlns,attr"`
	Result           XMLDescribeStacksResult `xml:"DescribeStacksResult"`
	ResponseMetadata XMLResponseMetadata     `xml:"ResponseMetadata"`
}

// XMLDescribeStacksResult contains the result of DescribeStacks.
type XMLDescribeStacksResult struct {
	Stacks XMLStacks `xml:"Stacks"`
}

// XMLStacks contains a list of stacks.
type XMLStacks struct {
	Members []XMLStack `xml:"member"`
}

// XMLStack represents a stack in XML format.
type XMLStack struct {
	StackID           string        `xml:"StackId"`
	StackName         string        `xml:"StackName"`
	StackStatus       string        `xml:"StackStatus"`
	StackStatusReason string        `xml:"StackStatusReason,omitempty"`
	CreationTime      string        `xml:"CreationTime"`
	LastUpdatedTime   string        `xml:"LastUpdatedTime,omitempty"`
	DeletionTime      string        `xml:"DeletionTime,omitempty"`
	Parameters        XMLParameters `xml:"Parameters,omitempty"`
}

// XMLParameters contains a list of parameters.
type XMLParameters struct {
	Members []XMLParameter `xml:"member"`
}

// XMLParameter represents a parameter in XML format.
type XMLParameter struct {
	ParameterKey   string `xml:"ParameterKey"`
	ParameterValue string `xml:"ParameterValue"`
}

// XMLListStacksResponse is the XML response for ListStacks.
type XMLListStacksResponse struct {
	XMLName          xml.Name            `xml:"ListStacksResponse"`
	Xmlns            string              `xml:"xmlns,attr"`
	Result           XMLListStacksResult `xml:"ListStacksResult"`
	ResponseMetadata XMLResponseMetadata `xml:"ResponseMetadata"`
}

// XMLListStacksResult contains the result of ListStacks.
type XMLListStacksResult struct {
	StackSummaries XMLStackSummaries `xml:"StackSummaries"`
}

// XMLStackSummaries contains a list of stack summaries.
type XMLStackSummaries struct {
	Members []XMLStackSummary `xml:"member"`
}

// XMLStackSummary represents a stack summary in XML format.
type XMLStackSummary struct {
	StackID           string `xml:"StackId"`
	StackName         string `xml:"StackName"`
	StackStatus       string `xml:"StackStatus"`
	StackStatusReason string `xml:"StackStatusReason,omitempty"`
	CreationTime      string `xml:"CreationTime"`
	LastUpdatedTime   string `xml:"LastUpdatedTime,omitempty"`
	DeletionTime      string `xml:"DeletionTime,omitempty"`
}

// XMLUpdateStackResponse is the XML response for UpdateStack.
type XMLUpdateStackResponse struct {
	XMLName          xml.Name             `xml:"UpdateStackResponse"`
	Xmlns            string               `xml:"xmlns,attr"`
	Result           XMLUpdateStackResult `xml:"UpdateStackResult"`
	ResponseMetadata XMLResponseMetadata  `xml:"ResponseMetadata"`
}

// XMLUpdateStackResult contains the result of UpdateStack.
type XMLUpdateStackResult struct {
	StackID string `xml:"StackId"`
}

// XMLDescribeStackResourcesResponse is the XML response for DescribeStackResources.
type XMLDescribeStackResourcesResponse struct {
	XMLName          xml.Name                        `xml:"DescribeStackResourcesResponse"`
	Xmlns            string                          `xml:"xmlns,attr"`
	Result           XMLDescribeStackResourcesResult `xml:"DescribeStackResourcesResult"`
	ResponseMetadata XMLResponseMetadata             `xml:"ResponseMetadata"`
}

// XMLDescribeStackResourcesResult contains the result of DescribeStackResources.
type XMLDescribeStackResourcesResult struct {
	StackResources XMLStackResources `xml:"StackResources"`
}

// XMLStackResources contains a list of stack resources.
type XMLStackResources struct {
	Members []XMLStackResource `xml:"member"`
}

// XMLStackResource represents a stack resource in XML format.
type XMLStackResource struct {
	LogicalResourceID  string `xml:"LogicalResourceId"`
	PhysicalResourceID string `xml:"PhysicalResourceId"`
	ResourceType       string `xml:"ResourceType"`
	ResourceStatus     string `xml:"ResourceStatus"`
	Timestamp          string `xml:"Timestamp"`
	StackID            string `xml:"StackId"`
	StackName          string `xml:"StackName"`
}

// XMLGetTemplateResponse is the XML response for GetTemplate.
type XMLGetTemplateResponse struct {
	XMLName          xml.Name             `xml:"GetTemplateResponse"`
	Xmlns            string               `xml:"xmlns,attr"`
	Result           XMLGetTemplateResult `xml:"GetTemplateResult"`
	ResponseMetadata XMLResponseMetadata  `xml:"ResponseMetadata"`
}

// XMLGetTemplateResult contains the result of GetTemplate.
type XMLGetTemplateResult struct {
	TemplateBody string `xml:"TemplateBody"`
}

// XMLValidateTemplateResponse is the XML response for ValidateTemplate.
type XMLValidateTemplateResponse struct {
	XMLName          xml.Name                  `xml:"ValidateTemplateResponse"`
	Xmlns            string                    `xml:"xmlns,attr"`
	Result           XMLValidateTemplateResult `xml:"ValidateTemplateResult"`
	ResponseMetadata XMLResponseMetadata       `xml:"ResponseMetadata"`
}

// XMLValidateTemplateResult contains the result of ValidateTemplate.
type XMLValidateTemplateResult struct {
	Parameters   XMLTemplateParameters `xml:"Parameters,omitempty"`
	Description  string                `xml:"Description,omitempty"`
	Capabilities XMLCapabilities       `xml:"Capabilities,omitempty"`
}

// XMLTemplateParameters contains a list of template parameters.
type XMLTemplateParameters struct {
	Members []XMLTemplateParameter `xml:"member"`
}

// XMLTemplateParameter represents a template parameter in XML format.
type XMLTemplateParameter struct {
	ParameterKey  string `xml:"ParameterKey"`
	DefaultValue  string `xml:"DefaultValue,omitempty"`
	NoEcho        bool   `xml:"NoEcho"`
	Description   string `xml:"Description,omitempty"`
	ParameterType string `xml:"ParameterType,omitempty"`
}

// XMLCapabilities contains a list of capabilities.
type XMLCapabilities struct {
	Members []string `xml:"member"`
}

// XMLResponseMetadata contains response metadata.
type XMLResponseMetadata struct {
	RequestID string `xml:"RequestId"`
}

// XMLErrorResponse is the XML error response.
type XMLErrorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	Error     XMLError `xml:"Error"`
	RequestID string   `xml:"RequestId"`
}

// XMLError represents an error in XML format.
type XMLError struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

// Error represents a CloudFormation error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
