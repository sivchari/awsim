package codegurureviewer

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

func epochNow() float64 {
	return float64(time.Now().Unix())
}

// Storage defines the interface for CodeGuru Reviewer storage operations.
type Storage interface {
	AssociateRepository(input *AssociateRepositoryInput) *RepositoryAssociation
	DescribeRepositoryAssociation(arn string) (*RepositoryAssociation, error)
	DisassociateRepository(arn string) (*RepositoryAssociation, error)
	ListRepositoryAssociations() []RepositoryAssociationSummary

	CreateCodeReview(input *CreateCodeReviewInput) (*CodeReview, error)
	DescribeCodeReview(arn string) (*CodeReview, error)
	ListCodeReviews() []CodeReview

	ListRecommendations(codeReviewArn string) []RecommendationSummary

	PutRecommendationFeedback(input *PutRecommendationFeedbackInput) error
	DescribeRecommendationFeedback(codeReviewArn, recommendationID string) (*RecommendationFeedback, error)
	ListRecommendationFeedback(codeReviewArn string) []RecommendationFeedback
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu           sync.RWMutex
	associations map[string]*RepositoryAssociation
	codeReviews  map[string]*CodeReview
	feedback     map[string]map[string]*RecommendationFeedback // codeReviewArn -> recommendationID -> feedback
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		associations: make(map[string]*RepositoryAssociation),
		codeReviews:  make(map[string]*CodeReview),
		feedback:     make(map[string]map[string]*RecommendationFeedback),
	}
}

// AssociateRepository creates a new repository association.
func (m *MemoryStorage) AssociateRepository(input *AssociateRepositoryInput) *RepositoryAssociation {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := epochNow()
	arn := fmt.Sprintf("arn:aws:codeguru-reviewer:us-east-1:000000000000:association:%s", id)

	name, owner, providerType, connectionArn := extractRepoDetails(input.Repository)

	assoc := &RepositoryAssociation{
		AssociationArn:       arn,
		AssociationID:        id,
		ConnectionArn:        connectionArn,
		CreatedTimeStamp:     now,
		LastUpdatedTimeStamp: now,
		Name:                 name,
		Owner:                owner,
		ProviderType:         providerType,
		State:                "Associated",
		Tags:                 input.Tags,
	}

	m.associations[arn] = assoc

	return assoc
}

func extractRepoDetails(repo *RepositoryInput) (name, owner, providerType, connectionArn string) {
	if repo == nil {
		return "", "", "", ""
	}

	if repo.CodeCommit != nil {
		return repo.CodeCommit.Name, "", "CodeCommit", ""
	}

	if repo.Bitbucket != nil {
		return repo.Bitbucket.Name, repo.Bitbucket.Owner, "Bitbucket", repo.Bitbucket.ConnectionArn
	}

	if repo.GitHubEnterpriseServer != nil {
		return repo.GitHubEnterpriseServer.Name, repo.GitHubEnterpriseServer.Owner, "GitHubEnterpriseServer", repo.GitHubEnterpriseServer.ConnectionArn
	}

	if repo.S3Bucket != nil {
		return repo.S3Bucket.Name, "", "S3Bucket", ""
	}

	return "", "", "", ""
}

// DescribeRepositoryAssociation returns a repository association by ARN.
func (m *MemoryStorage) DescribeRepositoryAssociation(arn string) (*RepositoryAssociation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	assoc, ok := m.associations[arn]
	if !ok {
		return nil, fmt.Errorf("NotFoundException: repository association %s not found", arn)
	}

	return assoc, nil
}

// DisassociateRepository removes a repository association.
func (m *MemoryStorage) DisassociateRepository(arn string) (*RepositoryAssociation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	assoc, ok := m.associations[arn]
	if !ok {
		return nil, fmt.Errorf("NotFoundException: repository association %s not found", arn)
	}

	assoc.State = "Disassociated"
	assoc.LastUpdatedTimeStamp = epochNow()

	delete(m.associations, arn)

	return assoc, nil
}

// ListRepositoryAssociations returns all repository associations.
func (m *MemoryStorage) ListRepositoryAssociations() []RepositoryAssociationSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]RepositoryAssociationSummary, 0, len(m.associations))
	for _, assoc := range m.associations {
		result = append(result, RepositoryAssociationSummary{
			AssociationArn:       assoc.AssociationArn,
			AssociationID:        assoc.AssociationID,
			ConnectionArn:        assoc.ConnectionArn,
			LastUpdatedTimeStamp: assoc.LastUpdatedTimeStamp,
			Name:                 assoc.Name,
			Owner:                assoc.Owner,
			ProviderType:         assoc.ProviderType,
			State:                assoc.State,
		})
	}

	return result
}

// CreateCodeReview creates a new code review.
func (m *MemoryStorage) CreateCodeReview(input *CreateCodeReviewInput) (*CodeReview, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the association.
	assoc, ok := m.associations[input.RepositoryAssociationArn]
	if !ok {
		return nil, fmt.Errorf("NotFoundException: repository association %s not found", input.RepositoryAssociationArn)
	}

	id := uuid.New().String()
	now := epochNow()
	arn := fmt.Sprintf("arn:aws:codeguru-reviewer:us-east-1:000000000000:association:%s:codereview:%s", assoc.AssociationID, id)

	review := &CodeReview{
		AssociationArn:       assoc.AssociationArn,
		CodeReviewArn:        arn,
		CreatedTimeStamp:     now,
		LastUpdatedTimeStamp: now,
		Name:                 input.Name,
		Owner:                assoc.Owner,
		ProviderType:         assoc.ProviderType,
		RepositoryName:       assoc.Name,
		State:                "Completed",
		Type:                 "RepositoryAnalysis",
	}

	m.codeReviews[arn] = review

	return review, nil
}

// DescribeCodeReview returns a code review by ARN.
func (m *MemoryStorage) DescribeCodeReview(arn string) (*CodeReview, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	review, ok := m.codeReviews[arn]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: code review %s not found", arn)
	}

	return review, nil
}

// ListCodeReviews returns all code reviews.
func (m *MemoryStorage) ListCodeReviews() []CodeReview {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]CodeReview, 0, len(m.codeReviews))
	for _, review := range m.codeReviews {
		result = append(result, *review)
	}

	return result
}

// ListRecommendations returns recommendations for a code review.
func (m *MemoryStorage) ListRecommendations(_ string) []RecommendationSummary {
	return []RecommendationSummary{}
}

// PutRecommendationFeedback stores feedback for a recommendation.
func (m *MemoryStorage) PutRecommendationFeedback(input *PutRecommendationFeedbackInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.codeReviews[input.CodeReviewArn]; !ok {
		return fmt.Errorf("ResourceNotFoundException: code review %s not found", input.CodeReviewArn)
	}

	now := epochNow()

	if m.feedback[input.CodeReviewArn] == nil {
		m.feedback[input.CodeReviewArn] = make(map[string]*RecommendationFeedback)
	}

	fb := &RecommendationFeedback{
		CodeReviewArn:        input.CodeReviewArn,
		CreatedTimeStamp:     now,
		LastUpdatedTimeStamp: now,
		Reactions:            input.Reactions,
		RecommendationID:     input.RecommendationID,
		UserID:               "test-user",
	}

	m.feedback[input.CodeReviewArn][input.RecommendationID] = fb

	return nil
}

// DescribeRecommendationFeedback returns feedback for a recommendation.
func (m *MemoryStorage) DescribeRecommendationFeedback(codeReviewArn, recommendationID string) (*RecommendationFeedback, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fbMap, ok := m.feedback[codeReviewArn]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: feedback not found")
	}

	fb, ok := fbMap[recommendationID]
	if !ok {
		return nil, fmt.Errorf("ResourceNotFoundException: feedback not found")
	}

	return fb, nil
}

// ListRecommendationFeedback returns all feedback for a code review.
func (m *MemoryStorage) ListRecommendationFeedback(codeReviewArn string) []RecommendationFeedback {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fbMap := m.feedback[codeReviewArn]
	result := make([]RecommendationFeedback, 0, len(fbMap))

	for _, fb := range fbMap {
		result = append(result, *fb)
	}

	return result
}
