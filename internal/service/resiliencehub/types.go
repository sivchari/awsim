// Package resiliencehub provides an in-memory implementation of AWS Resilience Hub.
package resiliencehub

// Error represents an error response.
type Error struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// App represents a Resilience Hub application.
type App struct {
	AppARN                      string               `json:"appArn,omitempty"`
	AssessmentSchedule          string               `json:"assessmentSchedule,omitempty"`
	ComplianceStatus            string               `json:"complianceStatus,omitempty"`
	CreationTime                float64              `json:"creationTime,omitempty"`
	Description                 string               `json:"description,omitempty"`
	DriftStatus                 string               `json:"driftStatus,omitempty"`
	EventSubscriptions          []*EventSubscription `json:"eventSubscriptions,omitempty"`
	LastAppComplianceEvalTime   float64              `json:"lastAppComplianceEvaluationTime,omitempty"`
	LastDriftEvalTime           float64              `json:"lastDriftEvaluationTime,omitempty"`
	LastResiliencyScoreEvalTime float64              `json:"lastResiliencyScoreEvaluationTime,omitempty"`
	Name                        string               `json:"name,omitempty"`
	PermissionModel             *PermissionModel     `json:"permissionModel,omitempty"`
	PolicyARN                   string               `json:"policyArn,omitempty"`
	ResiliencyScore             float64              `json:"resiliencyScore,omitempty"`
	RpoInSecs                   int                  `json:"rpoInSecs,omitempty"`
	RtoInSecs                   int                  `json:"rtoInSecs,omitempty"`
	Status                      string               `json:"status,omitempty"`
	Tags                        map[string]string    `json:"tags,omitempty"`
}

// EventSubscription represents an event subscription.
type EventSubscription struct {
	EventType   string `json:"eventType,omitempty"`
	Name        string `json:"name,omitempty"`
	SnsTopicARN string `json:"snsTopicArn,omitempty"`
}

// PermissionModel represents the permission model.
type PermissionModel struct {
	CrossAccountRoleARNs []string `json:"crossAccountRoleArns,omitempty"`
	InvokerRoleName      string   `json:"invokerRoleName,omitempty"`
	Type                 string   `json:"type,omitempty"`
}

// AppSummary represents a summary of an application.
type AppSummary struct {
	AppARN                    string  `json:"appArn,omitempty"`
	AssessmentSchedule        string  `json:"assessmentSchedule,omitempty"`
	ComplianceStatus          string  `json:"complianceStatus,omitempty"`
	CreationTime              float64 `json:"creationTime,omitempty"`
	Description               string  `json:"description,omitempty"`
	DriftStatus               string  `json:"driftStatus,omitempty"`
	LastAppComplianceEvalTime float64 `json:"lastAppComplianceEvaluationTime,omitempty"`
	Name                      string  `json:"name,omitempty"`
	ResiliencyScore           float64 `json:"resiliencyScore,omitempty"`
	RpoInSecs                 int     `json:"rpoInSecs,omitempty"`
	RtoInSecs                 int     `json:"rtoInSecs,omitempty"`
	Status                    string  `json:"status,omitempty"`
}

// ResiliencyPolicy represents a resiliency policy.
type ResiliencyPolicy struct {
	CreationTime           float64                   `json:"creationTime,omitempty"`
	DataLocationConstraint string                    `json:"dataLocationConstraint,omitempty"`
	EstimatedCostTier      string                    `json:"estimatedCostTier,omitempty"`
	Policy                 map[string]*FailurePolicy `json:"policy,omitempty"`
	PolicyARN              string                    `json:"policyArn,omitempty"`
	PolicyDescription      string                    `json:"policyDescription,omitempty"`
	PolicyName             string                    `json:"policyName,omitempty"`
	Tags                   map[string]string         `json:"tags,omitempty"`
	Tier                   string                    `json:"tier,omitempty"`
}

// FailurePolicy represents failure policy for a disruption type.
type FailurePolicy struct {
	RpoInSecs int `json:"rpoInSecs,omitempty"`
	RtoInSecs int `json:"rtoInSecs,omitempty"`
}

// AppAssessment represents an application assessment.
type AppAssessment struct {
	AppARN                string                           `json:"appArn,omitempty"`
	AppVersion            string                           `json:"appVersion,omitempty"`
	AssessmentARN         string                           `json:"assessmentArn,omitempty"`
	AssessmentName        string                           `json:"assessmentName,omitempty"`
	AssessmentStatus      string                           `json:"assessmentStatus,omitempty"`
	Compliance            map[string]*DisruptionCompliance `json:"compliance,omitempty"`
	ComplianceStatus      string                           `json:"complianceStatus,omitempty"`
	Cost                  *Cost                            `json:"cost,omitempty"`
	DriftStatus           string                           `json:"driftStatus,omitempty"`
	EndTime               float64                          `json:"endTime,omitempty"`
	Invoker               string                           `json:"invoker,omitempty"`
	Message               string                           `json:"message,omitempty"`
	Policy                *ResiliencyPolicy                `json:"policy,omitempty"`
	ResiliencyScore       *ResiliencyScore                 `json:"resiliencyScore,omitempty"`
	ResourceErrorsDetails *ResourceErrorsDetails           `json:"resourceErrorsDetails,omitempty"`
	StartTime             float64                          `json:"startTime,omitempty"`
	Tags                  map[string]string                `json:"tags,omitempty"`
	VersionName           string                           `json:"versionName,omitempty"`
}

// DisruptionCompliance represents compliance for a disruption type.
type DisruptionCompliance struct {
	AchievableRpoInSecs int    `json:"achievableRpoInSecs,omitempty"`
	AchievableRtoInSecs int    `json:"achievableRtoInSecs,omitempty"`
	ComplianceStatus    string `json:"complianceStatus,omitempty"`
	CurrentRpoInSecs    int    `json:"currentRpoInSecs,omitempty"`
	CurrentRtoInSecs    int    `json:"currentRtoInSecs,omitempty"`
	Message             string `json:"message,omitempty"`
	RpoDescription      string `json:"rpoDescription,omitempty"`
	RpoReferenceID      string `json:"rpoReferenceId,omitempty"`
	RtoDescription      string `json:"rtoDescription,omitempty"`
	RtoReferenceID      string `json:"rtoReferenceId,omitempty"`
}

// Cost represents a cost estimate.
type Cost struct {
	Amount    float64 `json:"amount,omitempty"`
	Currency  string  `json:"currency,omitempty"`
	Frequency string  `json:"frequency,omitempty"`
}

// ResiliencyScore represents a resiliency score.
type ResiliencyScore struct {
	ComponentScore  map[string]*ScoringComponentResiliencyScore `json:"componentScore,omitempty"`
	DisruptionScore map[string]float64                          `json:"disruptionScore,omitempty"`
	Score           float64                                     `json:"score,omitempty"`
}

// ScoringComponentResiliencyScore represents component resiliency score.
type ScoringComponentResiliencyScore struct {
	ExcludedCount    int64   `json:"excludedCount,omitempty"`
	OutstandingCount int64   `json:"outstandingCount,omitempty"`
	PossibleScore    float64 `json:"possibleScore,omitempty"`
	Score            float64 `json:"score,omitempty"`
}

// ResourceErrorsDetails represents resource errors details.
type ResourceErrorsDetails struct {
	HasMoreErrors  bool             `json:"hasMoreErrors,omitempty"`
	ResourceErrors []*ResourceError `json:"resourceErrors,omitempty"`
}

// ResourceError represents a resource error.
type ResourceError struct {
	LogicalResourceID  string `json:"logicalResourceId,omitempty"`
	PhysicalResourceID string `json:"physicalResourceId,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

// AppAssessmentSummary represents a summary of an assessment.
type AppAssessmentSummary struct {
	AppARN           string  `json:"appArn,omitempty"`
	AppVersion       string  `json:"appVersion,omitempty"`
	AssessmentARN    string  `json:"assessmentArn,omitempty"`
	AssessmentName   string  `json:"assessmentName,omitempty"`
	AssessmentStatus string  `json:"assessmentStatus,omitempty"`
	ComplianceStatus string  `json:"complianceStatus,omitempty"`
	Cost             *Cost   `json:"cost,omitempty"`
	DriftStatus      string  `json:"driftStatus,omitempty"`
	EndTime          float64 `json:"endTime,omitempty"`
	Invoker          string  `json:"invoker,omitempty"`
	Message          string  `json:"message,omitempty"`
	ResiliencyScore  float64 `json:"resiliencyScore,omitempty"`
	StartTime        float64 `json:"startTime,omitempty"`
	VersionName      string  `json:"versionName,omitempty"`
}

// CreateAppRequest represents a CreateApp request.
type CreateAppRequest struct {
	AssessmentSchedule string               `json:"assessmentSchedule,omitempty"`
	ClientToken        string               `json:"clientToken,omitempty"`
	Description        string               `json:"description,omitempty"`
	EventSubscriptions []*EventSubscription `json:"eventSubscriptions,omitempty"`
	Name               string               `json:"name"`
	PermissionModel    *PermissionModel     `json:"permissionModel,omitempty"`
	PolicyARN          string               `json:"policyArn,omitempty"`
	Tags               map[string]string    `json:"tags,omitempty"`
}

// CreateAppResponse represents a CreateApp response.
type CreateAppResponse struct {
	App *App `json:"app,omitempty"`
}

// DescribeAppRequest represents a DescribeApp request.
type DescribeAppRequest struct {
	AppARN string `json:"appArn"`
}

// DescribeAppResponse represents a DescribeApp response.
type DescribeAppResponse struct {
	App *App `json:"app,omitempty"`
}

// UpdateAppRequest represents an UpdateApp request.
type UpdateAppRequest struct {
	AppARN                   string               `json:"appArn"`
	AssessmentSchedule       string               `json:"assessmentSchedule,omitempty"`
	ClearResiliencyPolicyARN bool                 `json:"clearResiliencyPolicyArn,omitempty"`
	Description              string               `json:"description,omitempty"`
	EventSubscriptions       []*EventSubscription `json:"eventSubscriptions,omitempty"`
	PermissionModel          *PermissionModel     `json:"permissionModel,omitempty"`
	PolicyARN                string               `json:"policyArn,omitempty"`
}

// UpdateAppResponse represents an UpdateApp response.
type UpdateAppResponse struct {
	App *App `json:"app,omitempty"`
}

// DeleteAppRequest represents a DeleteApp request.
type DeleteAppRequest struct {
	AppARN      string `json:"appArn"`
	ClientToken string `json:"clientToken,omitempty"`
	ForceDelete bool   `json:"forceDelete,omitempty"`
}

// DeleteAppResponse represents a DeleteApp response.
type DeleteAppResponse struct {
	AppARN string `json:"appArn,omitempty"`
}

// ListAppsRequest represents a ListApps request.
type ListAppsRequest struct {
	AppARN                 string  `json:"appArn,omitempty"`
	FromLastAssessmentTime float64 `json:"fromLastAssessmentTime,omitempty"`
	MaxResults             int     `json:"maxResults,omitempty"`
	Name                   string  `json:"name,omitempty"`
	NextToken              string  `json:"nextToken,omitempty"`
	ReverseOrder           bool    `json:"reverseOrder,omitempty"`
	ToLastAssessmentTime   float64 `json:"toLastAssessmentTime,omitempty"`
}

// ListAppsResponse represents a ListApps response.
type ListAppsResponse struct {
	AppSummaries []*AppSummary `json:"appSummaries,omitempty"`
	NextToken    string        `json:"nextToken,omitempty"`
}

// CreateResiliencyPolicyRequest represents a CreateResiliencyPolicy request.
type CreateResiliencyPolicyRequest struct {
	ClientToken            string                    `json:"clientToken,omitempty"`
	DataLocationConstraint string                    `json:"dataLocationConstraint,omitempty"`
	Policy                 map[string]*FailurePolicy `json:"policy"`
	PolicyDescription      string                    `json:"policyDescription,omitempty"`
	PolicyName             string                    `json:"policyName"`
	Tags                   map[string]string         `json:"tags,omitempty"`
	Tier                   string                    `json:"tier"`
}

// CreateResiliencyPolicyResponse represents a CreateResiliencyPolicy response.
type CreateResiliencyPolicyResponse struct {
	Policy *ResiliencyPolicy `json:"policy,omitempty"`
}

// DescribeResiliencyPolicyRequest represents a DescribeResiliencyPolicy request.
type DescribeResiliencyPolicyRequest struct {
	PolicyARN string `json:"policyArn"`
}

// DescribeResiliencyPolicyResponse represents a DescribeResiliencyPolicy response.
type DescribeResiliencyPolicyResponse struct {
	Policy *ResiliencyPolicy `json:"policy,omitempty"`
}

// UpdateResiliencyPolicyRequest represents an UpdateResiliencyPolicy request.
type UpdateResiliencyPolicyRequest struct {
	DataLocationConstraint string                    `json:"dataLocationConstraint,omitempty"`
	Policy                 map[string]*FailurePolicy `json:"policy,omitempty"`
	PolicyARN              string                    `json:"policyArn"`
	PolicyDescription      string                    `json:"policyDescription,omitempty"`
	PolicyName             string                    `json:"policyName,omitempty"`
	Tier                   string                    `json:"tier,omitempty"`
}

// UpdateResiliencyPolicyResponse represents an UpdateResiliencyPolicy response.
type UpdateResiliencyPolicyResponse struct {
	Policy *ResiliencyPolicy `json:"policy,omitempty"`
}

// DeleteResiliencyPolicyRequest represents a DeleteResiliencyPolicy request.
type DeleteResiliencyPolicyRequest struct {
	ClientToken string `json:"clientToken,omitempty"`
	PolicyARN   string `json:"policyArn"`
}

// DeleteResiliencyPolicyResponse represents a DeleteResiliencyPolicy response.
type DeleteResiliencyPolicyResponse struct {
	PolicyARN string `json:"policyArn,omitempty"`
}

// ListResiliencyPoliciesRequest represents a ListResiliencyPolicies request.
type ListResiliencyPoliciesRequest struct {
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
	PolicyName string `json:"policyName,omitempty"`
}

// ListResiliencyPoliciesResponse represents a ListResiliencyPolicies response.
type ListResiliencyPoliciesResponse struct {
	NextToken          string              `json:"nextToken,omitempty"`
	ResiliencyPolicies []*ResiliencyPolicy `json:"resiliencyPolicies,omitempty"`
}

// StartAppAssessmentRequest represents a StartAppAssessment request.
type StartAppAssessmentRequest struct {
	AppARN         string            `json:"appArn"`
	AppVersion     string            `json:"appVersion"`
	AssessmentName string            `json:"assessmentName"`
	ClientToken    string            `json:"clientToken,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
}

// StartAppAssessmentResponse represents a StartAppAssessment response.
type StartAppAssessmentResponse struct {
	Assessment *AppAssessment `json:"assessment,omitempty"`
}

// DescribeAppAssessmentRequest represents a DescribeAppAssessment request.
type DescribeAppAssessmentRequest struct {
	AssessmentARN string `json:"assessmentArn"`
}

// DescribeAppAssessmentResponse represents a DescribeAppAssessment response.
type DescribeAppAssessmentResponse struct {
	Assessment *AppAssessment `json:"assessment,omitempty"`
}

// DeleteAppAssessmentRequest represents a DeleteAppAssessment request.
type DeleteAppAssessmentRequest struct {
	AssessmentARN string `json:"assessmentArn"`
	ClientToken   string `json:"clientToken,omitempty"`
}

// DeleteAppAssessmentResponse represents a DeleteAppAssessment response.
type DeleteAppAssessmentResponse struct {
	AssessmentARN    string `json:"assessmentArn,omitempty"`
	AssessmentStatus string `json:"assessmentStatus,omitempty"`
}

// ListAppAssessmentsRequest represents a ListAppAssessments request.
type ListAppAssessmentsRequest struct {
	AppARN           string `json:"appArn,omitempty"`
	AssessmentName   string `json:"assessmentName,omitempty"`
	AssessmentStatus string `json:"assessmentStatus,omitempty"`
	ComplianceStatus string `json:"complianceStatus,omitempty"`
	Invoker          string `json:"invoker,omitempty"`
	MaxResults       int    `json:"maxResults,omitempty"`
	NextToken        string `json:"nextToken,omitempty"`
	ReverseOrder     bool   `json:"reverseOrder,omitempty"`
}

// ListAppAssessmentsResponse represents a ListAppAssessments response.
type ListAppAssessmentsResponse struct {
	AssessmentSummaries []*AppAssessmentSummary `json:"assessmentSummaries,omitempty"`
	NextToken           string                  `json:"nextToken,omitempty"`
}

// TagResourceRequest represents a TagResource request.
type TagResourceRequest struct {
	ResourceARN string            `json:"resourceArn"`
	Tags        map[string]string `json:"tags"`
}

// TagResourceResponse represents a TagResource response.
type TagResourceResponse struct{}

// UntagResourceRequest represents an UntagResource request.
type UntagResourceRequest struct {
	ResourceARN string   `json:"resourceArn"`
	TagKeys     []string `json:"tagKeys"`
}

// UntagResourceResponse represents an UntagResource response.
type UntagResourceResponse struct{}

// ListTagsForResourceRequest represents a ListTagsForResource request.
type ListTagsForResourceRequest struct {
	ResourceARN string `json:"resourceArn"`
}

// ListTagsForResourceResponse represents a ListTagsForResource response.
type ListTagsForResourceResponse struct {
	Tags map[string]string `json:"tags,omitempty"`
}
