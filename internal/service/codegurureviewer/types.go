package codegurureviewer

// RepositoryAssociation represents a CodeGuru Reviewer repository association.
type RepositoryAssociation struct {
	AssociationArn       string            `json:"AssociationArn"`
	AssociationID        string            `json:"AssociationId"`
	ConnectionArn        string            `json:"ConnectionArn,omitempty"`
	CreatedTimeStamp     float64           `json:"CreatedTimeStamp"`
	LastUpdatedTimeStamp float64           `json:"LastUpdatedTimeStamp"`
	Name                 string            `json:"Name"`
	Owner                string            `json:"Owner"`
	ProviderType         string            `json:"ProviderType"`
	State                string            `json:"State"`
	StateReason          string            `json:"StateReason,omitempty"`
	Tags                 map[string]string `json:"Tags,omitempty"`
}

// RepositoryAssociationSummary represents a summary of a repository association.
type RepositoryAssociationSummary struct {
	AssociationArn       string  `json:"AssociationArn"`
	AssociationID        string  `json:"AssociationId"`
	ConnectionArn        string  `json:"ConnectionArn,omitempty"`
	LastUpdatedTimeStamp float64 `json:"LastUpdatedTimeStamp"`
	Name                 string  `json:"Name"`
	Owner                string  `json:"Owner"`
	ProviderType         string  `json:"ProviderType"`
	State                string  `json:"State"`
}

// CodeReview represents a CodeGuru Reviewer code review.
type CodeReview struct {
	AssociationArn       string  `json:"AssociationArn"`
	CodeReviewArn        string  `json:"CodeReviewArn"`
	CreatedTimeStamp     float64 `json:"CreatedTimeStamp"`
	LastUpdatedTimeStamp float64 `json:"LastUpdatedTimeStamp"`
	Name                 string  `json:"Name"`
	Owner                string  `json:"Owner"`
	ProviderType         string  `json:"ProviderType"`
	RepositoryName       string  `json:"RepositoryName"`
	State                string  `json:"State"`
	StateReason          string  `json:"StateReason,omitempty"`
	Type                 string  `json:"Type"`
}

// RecommendationSummary represents a recommendation from a code review.
type RecommendationSummary struct {
	Description            string `json:"Description"`
	EndLine                int    `json:"EndLine"`
	FilePath               string `json:"FilePath"`
	RecommendationID       string `json:"RecommendationId"`
	Severity               string `json:"Severity,omitempty"`
	StartLine              int    `json:"StartLine"`
	RecommendationCategory string `json:"RecommendationCategory,omitempty"`
}

// RecommendationFeedback represents feedback on a recommendation.
type RecommendationFeedback struct {
	CodeReviewArn        string   `json:"CodeReviewArn"`
	CreatedTimeStamp     float64  `json:"CreatedTimeStamp"`
	LastUpdatedTimeStamp float64  `json:"LastUpdatedTimeStamp"`
	Reactions            []string `json:"Reactions"`
	RecommendationID     string   `json:"RecommendationId"`
	UserID               string   `json:"UserId"`
}

// AssociateRepositoryInput represents the request body for AssociateRepository.
type AssociateRepositoryInput struct {
	Repository *RepositoryInput  `json:"Repository"`
	Tags       map[string]string `json:"Tags,omitempty"`
}

// RepositoryInput represents a repository in an associate request.
type RepositoryInput struct {
	CodeCommit             *CodeCommitRepository       `json:"CodeCommit,omitempty"`
	Bitbucket              *ThirdPartySourceRepository `json:"Bitbucket,omitempty"`
	GitHubEnterpriseServer *ThirdPartySourceRepository `json:"GitHubEnterpriseServer,omitempty"`
	S3Bucket               *S3Repository               `json:"S3Bucket,omitempty"`
}

// CodeCommitRepository represents a CodeCommit repository.
type CodeCommitRepository struct {
	Name string `json:"Name"`
}

// ThirdPartySourceRepository represents a third-party repository.
type ThirdPartySourceRepository struct {
	ConnectionArn string `json:"ConnectionArn"`
	Name          string `json:"Name"`
	Owner         string `json:"Owner"`
}

// S3Repository represents an S3 repository.
type S3Repository struct {
	BucketName string `json:"BucketName"`
	Name       string `json:"Name"`
}

// CreateCodeReviewInput represents the request body for CreateCodeReview.
type CreateCodeReviewInput struct {
	Name                     string `json:"Name"`
	RepositoryAssociationArn string `json:"RepositoryAssociationArn"`
}

// PutRecommendationFeedbackInput represents the request body for PutRecommendationFeedback.
type PutRecommendationFeedbackInput struct {
	CodeReviewArn    string   `json:"CodeReviewArn"`
	Reactions        []string `json:"Reactions"`
	RecommendationID string   `json:"RecommendationId"`
}

// AssociateRepositoryResponse represents the response for AssociateRepository.
type AssociateRepositoryResponse struct {
	RepositoryAssociation *RepositoryAssociation `json:"RepositoryAssociation"`
	Tags                  map[string]string      `json:"Tags,omitempty"`
}

// CodeReviewResponse represents the response for CreateCodeReview/DescribeCodeReview.
type CodeReviewResponse struct {
	CodeReview *CodeReview `json:"CodeReview"`
}

// ListRepositoryAssociationsResponse represents the response for ListRepositoryAssociations.
type ListRepositoryAssociationsResponse struct {
	RepositoryAssociationSummaries []RepositoryAssociationSummary `json:"RepositoryAssociationSummaries"`
	NextToken                      string                         `json:"NextToken,omitempty"`
}

// ListCodeReviewsResponse represents the response for ListCodeReviews.
type ListCodeReviewsResponse struct {
	CodeReviewSummaries []CodeReview `json:"CodeReviewSummaries"`
	NextToken           string       `json:"NextToken,omitempty"`
}

// ListRecommendationsResponse represents the response for ListRecommendations.
type ListRecommendationsResponse struct {
	RecommendationSummaries []RecommendationSummary `json:"RecommendationSummaries"`
	NextToken               string                  `json:"NextToken,omitempty"`
}

// RecommendationFeedbackResponse represents the response for DescribeRecommendationFeedback.
type RecommendationFeedbackResponse struct {
	RecommendationFeedback *RecommendationFeedback `json:"RecommendationFeedback"`
}

// ListRecommendationFeedbackResponse represents the response for ListRecommendationFeedback.
type ListRecommendationFeedbackResponse struct {
	RecommendationFeedbackSummaries []RecommendationFeedback `json:"RecommendationFeedbackSummaries"`
	NextToken                       string                   `json:"NextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"Message"`
	Type    string `json:"Type"`
}
