package organizations

import (
	"encoding/json"
	"time"
)

// AWSTimestamp is a time.Time that marshals to Unix timestamp (float64).
// AWS APIs use Unix timestamps in JSON responses.
type AWSTimestamp struct {
	time.Time
}

// MarshalJSON implements json.Marshaler.
func (t AWSTimestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return json.Marshal(nil) //nolint:wrapcheck // MarshalJSON interface requirement
	}

	return json.Marshal(float64(t.Unix()) + float64(t.Nanosecond())/1e9) //nolint:wrapcheck // MarshalJSON interface requirement
}

// ToAWSTimestamp converts time.Time to AWSTimestamp.
func ToAWSTimestamp(t time.Time) AWSTimestamp {
	return AWSTimestamp{Time: t}
}

// Domain types.

// Organization represents an AWS Organization.
type Organization struct {
	ARN                  string              `json:"Arn,omitempty"`
	AvailablePolicyTypes []PolicyTypeSummary `json:"AvailablePolicyTypes,omitempty"`
	FeatureSet           string              `json:"FeatureSet,omitempty"`
	ID                   string              `json:"Id,omitempty"`
	MasterAccountARN     string              `json:"MasterAccountArn,omitempty"`
	MasterAccountEmail   string              `json:"MasterAccountEmail,omitempty"`
	MasterAccountID      string              `json:"MasterAccountId,omitempty"`
}

// PolicyTypeSummary represents a policy type summary.
type PolicyTypeSummary struct {
	Status string `json:"Status,omitempty"`
	Type   string `json:"Type,omitempty"`
}

// Account represents an AWS account in an organization.
type Account struct {
	ARN             string       `json:"Arn,omitempty"`
	Email           string       `json:"Email,omitempty"`
	ID              string       `json:"Id,omitempty"`
	JoinedMethod    string       `json:"JoinedMethod,omitempty"`
	JoinedTimestamp AWSTimestamp `json:"JoinedTimestamp,omitempty"`
	Name            string       `json:"Name,omitempty"`
	State           string       `json:"State,omitempty"`
	Status          string       `json:"Status,omitempty"`
}

// OrganizationalUnit represents an organizational unit.
type OrganizationalUnit struct {
	ARN  string `json:"Arn,omitempty"`
	ID   string `json:"Id,omitempty"`
	Name string `json:"Name,omitempty"`
}

// Root represents the root of an organization.
type Root struct {
	ARN         string              `json:"Arn,omitempty"`
	ID          string              `json:"Id,omitempty"`
	Name        string              `json:"Name,omitempty"`
	PolicyTypes []PolicyTypeSummary `json:"PolicyTypes,omitempty"`
}

// Policy represents a policy.
type Policy struct {
	Content       string         `json:"Content,omitempty"`
	PolicySummary *PolicySummary `json:"PolicySummary,omitempty"`
}

// PolicySummary represents a policy summary.
type PolicySummary struct {
	ARN         string `json:"Arn,omitempty"`
	AWSManaged  bool   `json:"AwsManaged,omitempty"`
	Description string `json:"Description,omitempty"`
	ID          string `json:"Id,omitempty"`
	Name        string `json:"Name,omitempty"`
	Type        string `json:"Type,omitempty"`
}

// CreateAccountStatus represents the status of a CreateAccount request.
type CreateAccountStatus struct {
	AccountID          string       `json:"AccountId,omitempty"`
	AccountName        string       `json:"AccountName,omitempty"`
	CompletedTimestamp AWSTimestamp `json:"CompletedTimestamp,omitempty"`
	FailureReason      string       `json:"FailureReason,omitempty"`
	GovCloudAccountID  string       `json:"GovCloudAccountId,omitempty"`
	ID                 string       `json:"Id,omitempty"`
	RequestedTimestamp AWSTimestamp `json:"RequestedTimestamp,omitempty"`
	State              string       `json:"State,omitempty"`
}

// Tag represents a tag.
type Tag struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

// PolicyTargetSummary represents a policy attachment target.
type PolicyTargetSummary struct {
	ARN      string `json:"Arn,omitempty"`
	Name     string `json:"Name,omitempty"`
	TargetID string `json:"TargetId,omitempty"`
	Type     string `json:"Type,omitempty"`
}

// Request/Response types.

// CreateOrganizationInput represents the input for CreateOrganization.
type CreateOrganizationInput struct {
	FeatureSet string `json:"FeatureSet,omitempty"`
}

// CreateOrganizationOutput represents the output for CreateOrganization.
type CreateOrganizationOutput struct {
	Organization *Organization `json:"Organization,omitempty"`
}

// DeleteOrganizationInput represents the input for DeleteOrganization.
type DeleteOrganizationInput struct{}

// DescribeOrganizationInput represents the input for DescribeOrganization.
type DescribeOrganizationInput struct{}

// DescribeOrganizationOutput represents the output for DescribeOrganization.
type DescribeOrganizationOutput struct {
	Organization *Organization `json:"Organization,omitempty"`
}

// CreateAccountInput represents the input for CreateAccount.
type CreateAccountInput struct {
	AccountName            string `json:"AccountName"`
	Email                  string `json:"Email"`
	IamUserAccessToBilling string `json:"IamUserAccessToBilling,omitempty"`
	RoleName               string `json:"RoleName,omitempty"`
	Tags                   []Tag  `json:"Tags,omitempty"`
}

// CreateAccountOutput represents the output for CreateAccount.
type CreateAccountOutput struct {
	CreateAccountStatus *CreateAccountStatus `json:"CreateAccountStatus,omitempty"`
}

// DescribeAccountInput represents the input for DescribeAccount.
type DescribeAccountInput struct {
	AccountID string `json:"AccountId"`
}

// DescribeAccountOutput represents the output for DescribeAccount.
type DescribeAccountOutput struct {
	Account *Account `json:"Account,omitempty"`
}

// ListAccountsInput represents the input for ListAccounts.
type ListAccountsInput struct {
	MaxResults int32  `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListAccountsOutput represents the output for ListAccounts.
type ListAccountsOutput struct {
	Accounts  []*Account `json:"Accounts,omitempty"`
	NextToken string     `json:"NextToken,omitempty"`
}

// CreateOrganizationalUnitInput represents the input for CreateOrganizationalUnit.
type CreateOrganizationalUnitInput struct {
	Name     string `json:"Name"`
	ParentID string `json:"ParentId"`
	Tags     []Tag  `json:"Tags,omitempty"`
}

// CreateOrganizationalUnitOutput represents the output for CreateOrganizationalUnit.
type CreateOrganizationalUnitOutput struct {
	OrganizationalUnit *OrganizationalUnit `json:"OrganizationalUnit,omitempty"`
}

// ListOrganizationalUnitsForParentInput represents the input for ListOrganizationalUnitsForParent.
type ListOrganizationalUnitsForParentInput struct {
	MaxResults int32  `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
	ParentID   string `json:"ParentId"`
}

// ListOrganizationalUnitsForParentOutput represents the output for ListOrganizationalUnitsForParent.
type ListOrganizationalUnitsForParentOutput struct {
	NextToken           string                `json:"NextToken,omitempty"`
	OrganizationalUnits []*OrganizationalUnit `json:"OrganizationalUnits,omitempty"`
}

// AttachPolicyInput represents the input for AttachPolicy.
type AttachPolicyInput struct {
	PolicyID string `json:"PolicyId"`
	TargetID string `json:"TargetId"`
}

// DetachPolicyInput represents the input for DetachPolicy.
type DetachPolicyInput struct {
	PolicyID string `json:"PolicyId"`
	TargetID string `json:"TargetId"`
}

// ListRootsInput represents the input for ListRoots.
type ListRootsInput struct {
	MaxResults int32  `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListRootsOutput represents the output for ListRoots.
type ListRootsOutput struct {
	NextToken string  `json:"NextToken,omitempty"`
	Roots     []*Root `json:"Roots,omitempty"`
}

// Error represents an Organizations error.
type Error struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}
