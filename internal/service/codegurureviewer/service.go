package codegurureviewer

import (
	"github.com/sivchari/kumo/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the AWS CodeGuru Reviewer service.
type Service struct {
	storage Storage
}

// New creates a new CodeGuru Reviewer service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "codeguru-reviewer"
}

// Prefix returns the URL prefix for CodeGuru Reviewer.
func (s *Service) Prefix() string {
	return "/associations"
}

// RegisterRoutes registers routes with the router.
func (s *Service) RegisterRoutes(r service.Router) {
	// Repository association operations.
	r.Handle("POST", "/associations", s.AssociateRepository)
	r.Handle("GET", "/associations", s.ListRepositoryAssociations)
	r.Handle("GET", "/associations/{AssociationArn}", s.DescribeRepositoryAssociation)
	r.Handle("DELETE", "/associations/{AssociationArn}", s.DisassociateRepository)

	// Code review operations.
	r.Handle("POST", "/codereviews", s.CreateCodeReview)
	r.Handle("GET", "/codereviews", s.ListCodeReviews)
	r.Handle("GET", "/codereviews/{CodeReviewArn}", s.DescribeCodeReview)
	r.Handle("GET", "/codereviews/{CodeReviewArn}/Recommendations", s.ListRecommendations)

	// Recommendation feedback operations.
	r.Handle("PUT", "/feedback", s.PutRecommendationFeedback)
	r.Handle("GET", "/feedback/{CodeReviewArn}", s.DescribeRecommendationFeedback)
	r.Handle("GET", "/feedback/{CodeReviewArn}/RecommendationFeedback", s.ListRecommendationFeedback)
}
