package macie2

import "time"

// --- Domain Types ---

// AllowList represents an Amazon Macie2 allow list resource.
type AllowList struct {
	ID          string
	Name        string
	ARN         string
	Description string
	Criteria    AllowListCriteria
	Tags        map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AllowListCriteria represents the criteria for an allow list.
type AllowListCriteria struct {
	Regex       string
	S3WordsList *S3WordsList
}

// S3WordsList represents an S3 object reference for a words list.
type S3WordsList struct {
	BucketName string
	ObjectKey  string
}

// ClassificationJob represents an Amazon Macie2 classification job.
type ClassificationJob struct {
	JobID           string
	Name            string
	Description     string
	JobType         string
	JobStatus       string
	S3JobDefinition S3JobDefinition
	Tags            map[string]string
	CreatedAt       time.Time
}

// S3JobDefinition represents the S3 job definition for a classification job.
type S3JobDefinition struct {
	BucketDefinitions []BucketDefinition
}

// BucketDefinition represents a bucket definition for an S3 job.
type BucketDefinition struct {
	AccountID string
	Buckets   []string
}

// CustomDataIdentifier represents an Amazon Macie2 custom data identifier.
type CustomDataIdentifier struct {
	ID          string
	Name        string
	ARN         string
	Description string
	Regex       string
	Keywords    []string
	Tags        map[string]string
	CreatedAt   time.Time
}

// FindingsFilter represents an Amazon Macie2 findings filter.
type FindingsFilter struct {
	ID              string
	Name            string
	ARN             string
	Description     string
	Action          string
	FindingCriteria FindingCriteria
	Tags            map[string]string
	Position        int32
}

// FindingCriteria represents the criteria for a findings filter.
type FindingCriteria struct {
	Criterion map[string]CriterionValues
}

// CriterionValues represents the values for a criterion.
type CriterionValues struct {
	Eq  []string
	Neq []string
	Gt  *int64
	Gte *int64
	Lt  *int64
	Lte *int64
}

// MacieSession represents the Macie session state.
type MacieSession struct {
	FindingPublishingFrequency string
	ServiceRole                string
	Status                     string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

// Finding represents a Macie2 finding.
type Finding struct {
	ID          string
	Type        string
	Description string
	Severity    FindingSeverity
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FindingSeverity represents the severity of a finding.
type FindingSeverity struct {
	Score       int64
	Description string
}

// Error represents a Macie2 service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// --- Macie Session ---

// EnableMacieRequest represents the EnableMacie API request.
type EnableMacieRequest struct {
	FindingPublishingFrequency string `json:"findingPublishingFrequency,omitempty"`
	Status                     string `json:"status,omitempty"`
}

// EnableMacieResponse represents the EnableMacie API response.
type EnableMacieResponse struct{}

// GetMacieSessionResponse represents the GetMacieSession API response.
type GetMacieSessionResponse struct {
	CreatedAt                  time.Time `json:"createdAt"`
	FindingPublishingFrequency string    `json:"findingPublishingFrequency"`
	ServiceRole                string    `json:"serviceRole"`
	Status                     string    `json:"status"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// UpdateMacieSessionRequest represents the UpdateMacieSession API request.
type UpdateMacieSessionRequest struct {
	FindingPublishingFrequency string `json:"findingPublishingFrequency,omitempty"`
	Status                     string `json:"status,omitempty"`
}

// UpdateMacieSessionResponse represents the UpdateMacieSession API response.
type UpdateMacieSessionResponse struct{}

// DisableMacieResponse represents the DisableMacie API response.
type DisableMacieResponse struct{}

// --- Allow Lists ---

// AllowListCriteriaInput represents the criteria input for an allow list.
type AllowListCriteriaInput struct {
	Regex       string            `json:"regex,omitempty"`
	S3WordsList *S3WordsListInput `json:"s3WordsList,omitempty"`
}

// S3WordsListInput represents the S3 words list input.
type S3WordsListInput struct {
	BucketName string `json:"bucketName"`
	ObjectKey  string `json:"objectKey"`
}

// CreateAllowListRequest represents the CreateAllowList API request.
type CreateAllowListRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Criteria    AllowListCriteriaInput `json:"criteria"`
	Tags        map[string]string      `json:"tags,omitempty"`
}

// CreateAllowListResponse represents the CreateAllowList API response.
type CreateAllowListResponse struct {
	ID  string `json:"id"`
	ARN string `json:"arn"`
}

// GetAllowListResponse represents the GetAllowList API response.
type GetAllowListResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	ARN         string                     `json:"arn"`
	Description string                     `json:"description,omitempty"`
	Criteria    GetAllowListCriteriaOutput `json:"criteria"`
	Tags        map[string]string          `json:"tags,omitempty"`
	CreatedAt   time.Time                  `json:"createdAt"`
	UpdatedAt   time.Time                  `json:"updatedAt"`
}

// GetAllowListCriteriaOutput represents the criteria output for a get allow list response.
type GetAllowListCriteriaOutput struct {
	Regex       string             `json:"regex,omitempty"`
	S3WordsList *S3WordsListOutput `json:"s3WordsList,omitempty"`
}

// S3WordsListOutput represents the S3 words list output.
type S3WordsListOutput struct {
	BucketName string `json:"bucketName"`
	ObjectKey  string `json:"objectKey"`
}

// UpdateAllowListRequest represents the UpdateAllowList API request.
type UpdateAllowListRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Criteria    AllowListCriteriaInput `json:"criteria"`
}

// UpdateAllowListResponse represents the UpdateAllowList API response.
type UpdateAllowListResponse struct {
	ID  string `json:"id"`
	ARN string `json:"arn"`
}

// ListAllowListsResponse represents the ListAllowLists API response.
type ListAllowListsResponse struct {
	AllowLists []AllowListSummary `json:"allowLists"`
	NextToken  string             `json:"nextToken,omitempty"`
}

// AllowListSummary represents a summary entry in the ListAllowLists response.
type AllowListSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ARN         string    `json:"arn"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// --- Classification Jobs ---

// S3JobDefinitionInput represents the S3 job definition input.
type S3JobDefinitionInput struct {
	BucketDefinitions []BucketDefinitionInput `json:"bucketDefinitions,omitempty"`
}

// BucketDefinitionInput represents a bucket definition input.
type BucketDefinitionInput struct {
	AccountID string   `json:"accountId"`
	Buckets   []string `json:"buckets"`
}

// CreateClassificationJobRequest represents the CreateClassificationJob API request.
type CreateClassificationJobRequest struct {
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	JobType         string               `json:"jobType"`
	S3JobDefinition S3JobDefinitionInput `json:"s3JobDefinition"`
	Tags            map[string]string    `json:"tags,omitempty"`
}

// CreateClassificationJobResponse represents the CreateClassificationJob API response.
type CreateClassificationJobResponse struct {
	JobID  string `json:"jobId"`
	JobArn string `json:"jobArn"`
}

// DescribeClassificationJobResponse represents the DescribeClassificationJob API response.
type DescribeClassificationJobResponse struct {
	JobID           string                `json:"jobId"`
	Name            string                `json:"name"`
	Description     string                `json:"description,omitempty"`
	JobType         string                `json:"jobType"`
	JobStatus       string                `json:"jobStatus"`
	S3JobDefinition S3JobDefinitionOutput `json:"s3JobDefinition"`
	Tags            map[string]string     `json:"tags,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
}

// S3JobDefinitionOutput represents the S3 job definition output.
type S3JobDefinitionOutput struct {
	BucketDefinitions []BucketDefinitionOutput `json:"bucketDefinitions,omitempty"`
}

// BucketDefinitionOutput represents a bucket definition output.
type BucketDefinitionOutput struct {
	AccountID string   `json:"accountId"`
	Buckets   []string `json:"buckets"`
}

// ListClassificationJobsRequest represents the ListClassificationJobs API request.
type ListClassificationJobsRequest struct {
	MaxResults *int32 `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

// ListClassificationJobsResponse represents the ListClassificationJobs API response.
type ListClassificationJobsResponse struct {
	Items     []ClassificationJobSummary `json:"items"`
	NextToken string                     `json:"nextToken,omitempty"`
}

// ClassificationJobSummary represents a summary entry in the ListClassificationJobs response.
type ClassificationJobSummary struct {
	JobID     string    `json:"jobId"`
	Name      string    `json:"name"`
	JobType   string    `json:"jobType"`
	JobStatus string    `json:"jobStatus"`
	CreatedAt time.Time `json:"createdAt"`
}

// UpdateClassificationJobRequest represents the UpdateClassificationJob API request.
type UpdateClassificationJobRequest struct {
	JobStatus string `json:"jobStatus"`
}

// UpdateClassificationJobResponse represents the UpdateClassificationJob API response.
type UpdateClassificationJobResponse struct{}

// --- Custom Data Identifiers ---

// CreateCustomDataIdentifierRequest represents the CreateCustomDataIdentifier API request.
type CreateCustomDataIdentifierRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Regex       string            `json:"regex"`
	Keywords    []string          `json:"keywords,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// CreateCustomDataIdentifierResponse represents the CreateCustomDataIdentifier API response.
type CreateCustomDataIdentifierResponse struct {
	CustomDataIdentifierID string `json:"customDataIdentifierId"`
}

// GetCustomDataIdentifierResponse represents the GetCustomDataIdentifier API response.
type GetCustomDataIdentifierResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ARN         string            `json:"arn"`
	Description string            `json:"description,omitempty"`
	Regex       string            `json:"regex"`
	Keywords    []string          `json:"keywords,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// ListCustomDataIdentifiersRequest represents the ListCustomDataIdentifiers API request.
type ListCustomDataIdentifiersRequest struct {
	MaxResults *int32 `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

// ListCustomDataIdentifiersResponse represents the ListCustomDataIdentifiers API response.
type ListCustomDataIdentifiersResponse struct {
	Items     []CustomDataIdentifierSummary `json:"items"`
	NextToken string                        `json:"nextToken,omitempty"`
}

// CustomDataIdentifierSummary represents a summary entry in the ListCustomDataIdentifiers response.
type CustomDataIdentifierSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ARN         string    `json:"arn"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// --- Findings Filters ---

// FindingCriteriaInput represents the finding criteria input.
type FindingCriteriaInput struct {
	Criterion map[string]CriterionValuesInput `json:"criterion,omitempty"`
}

// CriterionValuesInput represents the criterion values input.
type CriterionValuesInput struct {
	Eq  []string `json:"eq,omitempty"`
	Neq []string `json:"neq,omitempty"`
	Gt  *int64   `json:"gt,omitempty"`
	Gte *int64   `json:"gte,omitempty"`
	Lt  *int64   `json:"lt,omitempty"`
	Lte *int64   `json:"lte,omitempty"`
}

// CreateFindingsFilterRequest represents the CreateFindingsFilter API request.
type CreateFindingsFilterRequest struct {
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	Action          string               `json:"action"`
	FindingCriteria FindingCriteriaInput `json:"findingCriteria"`
	Position        *int32               `json:"position,omitempty"`
	Tags            map[string]string    `json:"tags,omitempty"`
}

// CreateFindingsFilterResponse represents the CreateFindingsFilter API response.
type CreateFindingsFilterResponse struct {
	ID  string `json:"id"`
	ARN string `json:"arn"`
}

// GetFindingsFilterResponse represents the GetFindingsFilter API response.
type GetFindingsFilterResponse struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	ARN             string                `json:"arn"`
	Description     string                `json:"description,omitempty"`
	Action          string                `json:"action"`
	FindingCriteria FindingCriteriaOutput `json:"findingCriteria"`
	Tags            map[string]string     `json:"tags,omitempty"`
	Position        int32                 `json:"position"`
}

// FindingCriteriaOutput represents the finding criteria output.
type FindingCriteriaOutput struct {
	Criterion map[string]CriterionValuesOutput `json:"criterion,omitempty"`
}

// CriterionValuesOutput represents the criterion values output.
type CriterionValuesOutput struct {
	Eq  []string `json:"eq,omitempty"`
	Neq []string `json:"neq,omitempty"`
	Gt  *int64   `json:"gt,omitempty"`
	Gte *int64   `json:"gte,omitempty"`
	Lt  *int64   `json:"lt,omitempty"`
	Lte *int64   `json:"lte,omitempty"`
}

// UpdateFindingsFilterRequest represents the UpdateFindingsFilter API request.
type UpdateFindingsFilterRequest struct {
	Name            string                `json:"name,omitempty"`
	Description     string                `json:"description,omitempty"`
	Action          string                `json:"action,omitempty"`
	FindingCriteria *FindingCriteriaInput `json:"findingCriteria,omitempty"`
	Position        *int32                `json:"position,omitempty"`
}

// UpdateFindingsFilterResponse represents the UpdateFindingsFilter API response.
type UpdateFindingsFilterResponse struct {
	ID  string `json:"id"`
	ARN string `json:"arn"`
}

// ListFindingsFiltersResponse represents the ListFindingsFilters API response.
type ListFindingsFiltersResponse struct {
	FindingsFilterListItems []FindingsFilterSummary `json:"findingsFilterListItems"`
	NextToken               string                  `json:"nextToken,omitempty"`
}

// FindingsFilterSummary represents a summary entry in the ListFindingsFilters response.
type FindingsFilterSummary struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	ARN    string            `json:"arn"`
	Action string            `json:"action"`
	Tags   map[string]string `json:"tags,omitempty"`
}

// --- Findings ---

// GetFindingsRequest represents the GetFindings API request.
type GetFindingsRequest struct {
	FindingIDs []string `json:"findingIds"`
}

// GetFindingsResponse represents the GetFindings API response.
type GetFindingsResponse struct {
	Findings []FindingDetail `json:"findings"`
}

// FindingDetail represents a finding detail in the GetFindings response.
type FindingDetail struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"`
	Description string                `json:"description,omitempty"`
	Severity    FindingSeverityOutput `json:"severity"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

// FindingSeverityOutput represents the severity output.
type FindingSeverityOutput struct {
	Score       int64  `json:"score"`
	Description string `json:"description"`
}

// ListFindingsRequest represents the ListFindings API request.
type ListFindingsRequest struct {
	FindingCriteria *FindingCriteriaInput `json:"findingCriteria,omitempty"`
	MaxResults      *int32                `json:"maxResults,omitempty"`
	NextToken       string                `json:"nextToken,omitempty"`
}

// ListFindingsResponse represents the ListFindings API response.
type ListFindingsResponse struct {
	FindingIDs []string `json:"findingIds"`
	NextToken  string   `json:"nextToken,omitempty"`
}

// --- Error Response ---

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}
