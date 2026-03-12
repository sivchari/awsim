package entityresolution

import "time"

// SchemaInputAttribute represents a mapped input field in a schema.
type SchemaInputAttribute struct {
	FieldName string `json:"fieldName"`
	Type      string `json:"type"`
	MatchKey  string `json:"matchKey,omitempty"`
	GroupName string `json:"groupName,omitempty"`
	SubType   string `json:"subType,omitempty"`
	Hashed    bool   `json:"hashed,omitempty"`
}

// SchemaMapping represents a schema mapping.
type SchemaMapping struct {
	SchemaName        string                 `json:"schemaName"`
	SchemaArn         string                 `json:"schemaArn"`
	Description       string                 `json:"description,omitempty"`
	MappedInputFields []SchemaInputAttribute `json:"mappedInputFields"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
	Tags              map[string]string      `json:"tags,omitempty"`
}

// InputSource represents an input source for a workflow.
type InputSource struct {
	InputSourceARN     string `json:"inputSourceARN"`
	SchemaName         string `json:"schemaName,omitempty"`
	ApplyNormalization *bool  `json:"applyNormalization,omitempty"`
}

// OutputAttribute represents an output attribute.
type OutputAttribute struct {
	Name   string `json:"name"`
	Hashed bool   `json:"hashed,omitempty"`
}

// OutputSource represents an output source for a workflow.
type OutputSource struct {
	OutputS3Path       string            `json:"outputS3Path"`
	Output             []OutputAttribute `json:"output"`
	KMSArn             string            `json:"KMSArn,omitempty"`
	ApplyNormalization *bool             `json:"applyNormalization,omitempty"`
}

// ResolutionTechniques represents resolution techniques configuration.
type ResolutionTechniques struct {
	ResolutionType string `json:"resolutionType"`
}

// MatchingWorkflow represents a matching workflow.
type MatchingWorkflow struct {
	WorkflowName         string                `json:"workflowName"`
	WorkflowArn          string                `json:"workflowArn"`
	Description          string                `json:"description,omitempty"`
	InputSourceConfig    []InputSource         `json:"inputSourceConfig"`
	OutputSourceConfig   []OutputSource        `json:"outputSourceConfig"`
	ResolutionTechniques *ResolutionTechniques `json:"resolutionTechniques"`
	RoleArn              string                `json:"roleArn"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
	Tags                 map[string]string     `json:"tags,omitempty"`
}

// IdMappingTechniques represents ID mapping techniques configuration.
type IdMappingTechniques struct {
	IdMappingType string `json:"idMappingType"`
}

// IdMappingInputSource represents an input source for an ID mapping workflow.
type IdMappingInputSource struct {
	InputSourceARN string `json:"inputSourceARN"`
	SchemaName     string `json:"schemaName,omitempty"`
	Type           string `json:"type,omitempty"`
}

// IdMappingOutputSource represents an output source for an ID mapping workflow.
type IdMappingOutputSource struct {
	OutputS3Path string `json:"outputS3Path"`
	KMSArn       string `json:"KMSArn,omitempty"`
}

// IdMappingWorkflow represents an ID mapping workflow.
type IdMappingWorkflow struct {
	WorkflowName        string                  `json:"workflowName"`
	WorkflowArn         string                  `json:"workflowArn"`
	Description         string                  `json:"description,omitempty"`
	InputSourceConfig   []IdMappingInputSource  `json:"inputSourceConfig"`
	OutputSourceConfig  []IdMappingOutputSource `json:"outputSourceConfig,omitempty"`
	IdMappingTechniques *IdMappingTechniques    `json:"idMappingTechniques"`
	RoleArn             string                  `json:"roleArn,omitempty"`
	CreatedAt           time.Time               `json:"createdAt"`
	UpdatedAt           time.Time               `json:"updatedAt"`
	Tags                map[string]string       `json:"tags,omitempty"`
}

// ProviderService represents a provider service.
type ProviderService struct {
	ProviderName        string `json:"providerName"`
	ProviderServiceName string `json:"providerServiceName"`
	ProviderServiceArn  string `json:"providerServiceArn"`
	ProviderServiceType string `json:"providerServiceType"`
}

// Request/Response types.

// CreateSchemaMappingRequest represents the request for CreateSchemaMapping.
type CreateSchemaMappingRequest struct {
	SchemaName        string                 `json:"schemaName"`
	Description       string                 `json:"description,omitempty"`
	MappedInputFields []SchemaInputAttribute `json:"mappedInputFields"`
	Tags              map[string]string      `json:"tags,omitempty"`
}

// CreateMatchingWorkflowRequest represents the request for CreateMatchingWorkflow.
type CreateMatchingWorkflowRequest struct {
	WorkflowName         string                `json:"workflowName"`
	Description          string                `json:"description,omitempty"`
	InputSourceConfig    []InputSource         `json:"inputSourceConfig"`
	OutputSourceConfig   []OutputSource        `json:"outputSourceConfig"`
	ResolutionTechniques *ResolutionTechniques `json:"resolutionTechniques"`
	RoleArn              string                `json:"roleArn"`
	Tags                 map[string]string     `json:"tags,omitempty"`
}

// CreateIdMappingWorkflowRequest represents the request for CreateIdMappingWorkflow.
type CreateIdMappingWorkflowRequest struct {
	WorkflowName        string                  `json:"workflowName"`
	Description         string                  `json:"description,omitempty"`
	InputSourceConfig   []IdMappingInputSource  `json:"inputSourceConfig"`
	OutputSourceConfig  []IdMappingOutputSource `json:"outputSourceConfig,omitempty"`
	IdMappingTechniques *IdMappingTechniques    `json:"idMappingTechniques"`
	RoleArn             string                  `json:"roleArn,omitempty"`
	Tags                map[string]string       `json:"tags,omitempty"`
}

// SchemaMappingSummary represents a summary of a schema mapping for list responses.
type SchemaMappingSummary struct {
	SchemaName string    `json:"schemaName"`
	SchemaArn  string    `json:"schemaArn"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// MatchingWorkflowSummary represents a summary of a matching workflow for list responses.
type MatchingWorkflowSummary struct {
	WorkflowName string    `json:"workflowName"`
	WorkflowArn  string    `json:"workflowArn"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// IdMappingWorkflowSummary represents a summary of an ID mapping workflow for list responses.
type IdMappingWorkflowSummary struct {
	WorkflowName string    `json:"workflowName"`
	WorkflowArn  string    `json:"workflowArn"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ListSchemaMappingsResponse represents the response for ListSchemaMappings.
type ListSchemaMappingsResponse struct {
	SchemaList []SchemaMappingSummary `json:"schemaList"`
	NextToken  string                 `json:"nextToken,omitempty"`
}

// ListMatchingWorkflowsResponse represents the response for ListMatchingWorkflows.
type ListMatchingWorkflowsResponse struct {
	WorkflowSummaries []MatchingWorkflowSummary `json:"workflowSummaries"`
	NextToken         string                    `json:"nextToken,omitempty"`
}

// ListIdMappingWorkflowsResponse represents the response for ListIdMappingWorkflows.
type ListIdMappingWorkflowsResponse struct {
	WorkflowSummaries []IdMappingWorkflowSummary `json:"workflowSummaries"`
	NextToken         string                     `json:"nextToken,omitempty"`
}

// ListProviderServicesResponse represents the response for ListProviderServices.
type ListProviderServicesResponse struct {
	ProviderServiceSummaries []ProviderService `json:"providerServiceSummaries"`
	NextToken                string            `json:"nextToken,omitempty"`
}

// Error represents an Entity Resolution error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// Error codes.
const (
	errValidation    = "ValidationException"
	errNotFound      = "ResourceNotFoundException"
	errConflict      = "ConflictException"
	errInternalError = "InternalServerException"
	errAccessDenied  = "AccessDeniedException"
)
