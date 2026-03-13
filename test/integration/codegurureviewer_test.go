//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer/types"
)

func newCodeGuruReviewerClient(t *testing.T) *codegurureviewer.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return codegurureviewer.NewFromConfig(cfg, func(o *codegurureviewer.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCodeGuruReviewer_AssociateRepository(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	result, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("my-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	if result.RepositoryAssociation == nil {
		t.Fatal("expected RepositoryAssociation to be set")
	}

	if result.RepositoryAssociation.AssociationArn == nil || *result.RepositoryAssociation.AssociationArn == "" {
		t.Error("expected AssociationArn to be set")
	}

	if *result.RepositoryAssociation.Name != "my-repo" {
		t.Errorf("expected name 'my-repo', got %s", *result.RepositoryAssociation.Name)
	}

	if result.RepositoryAssociation.State != types.RepositoryAssociationStateAssociated {
		t.Errorf("expected state Associated, got %s", result.RepositoryAssociation.State)
	}
}

func TestCodeGuruReviewer_DescribeRepositoryAssociation(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("describe-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	result, err := client.DescribeRepositoryAssociation(ctx, &codegurureviewer.DescribeRepositoryAssociationInput{
		AssociationArn: assocResult.RepositoryAssociation.AssociationArn,
	})
	if err != nil {
		t.Fatalf("failed to describe repository association: %v", err)
	}

	if *result.RepositoryAssociation.Name != "describe-repo" {
		t.Errorf("expected name 'describe-repo', got %s", *result.RepositoryAssociation.Name)
	}
}

func TestCodeGuruReviewer_DisassociateRepository(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("disassociate-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	result, err := client.DisassociateRepository(ctx, &codegurureviewer.DisassociateRepositoryInput{
		AssociationArn: assocResult.RepositoryAssociation.AssociationArn,
	})
	if err != nil {
		t.Fatalf("failed to disassociate repository: %v", err)
	}

	if result.RepositoryAssociation.State != types.RepositoryAssociationStateDisassociated {
		t.Errorf("expected state Disassociated, got %s", result.RepositoryAssociation.State)
	}
}

func TestCodeGuruReviewer_ListRepositoryAssociations(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	_, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("list-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	result, err := client.ListRepositoryAssociations(ctx, &codegurureviewer.ListRepositoryAssociationsInput{})
	if err != nil {
		t.Fatalf("failed to list repository associations: %v", err)
	}

	if len(result.RepositoryAssociationSummaries) == 0 {
		t.Error("expected at least one repository association")
	}
}

func TestCodeGuruReviewer_CreateCodeReview(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("review-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	result, err := client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("test-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	if result.CodeReview == nil {
		t.Fatal("expected CodeReview to be set")
	}

	if result.CodeReview.CodeReviewArn == nil || *result.CodeReview.CodeReviewArn == "" {
		t.Error("expected CodeReviewArn to be set")
	}

	if *result.CodeReview.Name != "test-review" {
		t.Errorf("expected name 'test-review', got %s", *result.CodeReview.Name)
	}
}

func TestCodeGuruReviewer_DescribeCodeReview(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("describe-review-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	createResult, err := client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("describe-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	result, err := client.DescribeCodeReview(ctx, &codegurureviewer.DescribeCodeReviewInput{
		CodeReviewArn: createResult.CodeReview.CodeReviewArn,
	})
	if err != nil {
		t.Fatalf("failed to describe code review: %v", err)
	}

	if *result.CodeReview.Name != "describe-review" {
		t.Errorf("expected name 'describe-review', got %s", *result.CodeReview.Name)
	}
}

func TestCodeGuruReviewer_ListCodeReviews(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("list-review-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	_, err = client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("list-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	result, err := client.ListCodeReviews(ctx, &codegurureviewer.ListCodeReviewsInput{
		Type: types.TypeRepositoryAnalysis,
	})
	if err != nil {
		t.Fatalf("failed to list code reviews: %v", err)
	}

	if len(result.CodeReviewSummaries) == 0 {
		t.Error("expected at least one code review")
	}
}

func TestCodeGuruReviewer_ListRecommendations(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("recommendations-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	createResult, err := client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("recommendations-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	result, err := client.ListRecommendations(ctx, &codegurureviewer.ListRecommendationsInput{
		CodeReviewArn: createResult.CodeReview.CodeReviewArn,
	})
	if err != nil {
		t.Fatalf("failed to list recommendations: %v", err)
	}

	if result.RecommendationSummaries == nil {
		t.Error("expected RecommendationSummaries to be non-nil")
	}
}

func TestCodeGuruReviewer_PutRecommendationFeedback(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("feedback-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	createResult, err := client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("feedback-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	_, err = client.PutRecommendationFeedback(ctx, &codegurureviewer.PutRecommendationFeedbackInput{
		CodeReviewArn:    createResult.CodeReview.CodeReviewArn,
		RecommendationId: aws.String("rec-1"),
		Reactions:        []types.Reaction{types.ReactionThumbsUp},
	})
	if err != nil {
		t.Fatalf("failed to put recommendation feedback: %v", err)
	}
}

func TestCodeGuruReviewer_DescribeRecommendationFeedback(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("describe-feedback-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	createResult, err := client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("describe-feedback-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	_, err = client.PutRecommendationFeedback(ctx, &codegurureviewer.PutRecommendationFeedbackInput{
		CodeReviewArn:    createResult.CodeReview.CodeReviewArn,
		RecommendationId: aws.String("rec-2"),
		Reactions:        []types.Reaction{types.ReactionThumbsDown},
	})
	if err != nil {
		t.Fatalf("failed to put recommendation feedback: %v", err)
	}

	result, err := client.DescribeRecommendationFeedback(ctx, &codegurureviewer.DescribeRecommendationFeedbackInput{
		CodeReviewArn:    createResult.CodeReview.CodeReviewArn,
		RecommendationId: aws.String("rec-2"),
	})
	if err != nil {
		t.Fatalf("failed to describe recommendation feedback: %v", err)
	}

	if result.RecommendationFeedback == nil {
		t.Fatal("expected RecommendationFeedback to be set")
	}

	if *result.RecommendationFeedback.RecommendationId != "rec-2" {
		t.Errorf("expected RecommendationId 'rec-2', got %s", *result.RecommendationFeedback.RecommendationId)
	}
}

func TestCodeGuruReviewer_ListRecommendationFeedback(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	assocResult, err := client.AssociateRepository(ctx, &codegurureviewer.AssociateRepositoryInput{
		Repository: &types.Repository{
			CodeCommit: &types.CodeCommitRepository{
				Name: aws.String("list-feedback-repo"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to associate repository: %v", err)
	}

	createResult, err := client.CreateCodeReview(ctx, &codegurureviewer.CreateCodeReviewInput{
		Name:                     aws.String("list-feedback-review"),
		RepositoryAssociationArn: assocResult.RepositoryAssociation.AssociationArn,
		Type: &types.CodeReviewType{
			RepositoryAnalysis: &types.RepositoryAnalysis{
				RepositoryHead: &types.RepositoryHeadSourceCodeType{
					BranchName: aws.String("main"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create code review: %v", err)
	}

	_, err = client.PutRecommendationFeedback(ctx, &codegurureviewer.PutRecommendationFeedbackInput{
		CodeReviewArn:    createResult.CodeReview.CodeReviewArn,
		RecommendationId: aws.String("rec-3"),
		Reactions:        []types.Reaction{types.ReactionThumbsUp},
	})
	if err != nil {
		t.Fatalf("failed to put recommendation feedback: %v", err)
	}

	result, err := client.ListRecommendationFeedback(ctx, &codegurureviewer.ListRecommendationFeedbackInput{
		CodeReviewArn: createResult.CodeReview.CodeReviewArn,
	})
	if err != nil {
		t.Fatalf("failed to list recommendation feedback: %v", err)
	}

	if len(result.RecommendationFeedbackSummaries) == 0 {
		t.Error("expected at least one recommendation feedback")
	}
}

func TestCodeGuruReviewer_AssociationNotFound(t *testing.T) {
	client := newCodeGuruReviewerClient(t)
	ctx := t.Context()

	_, err := client.DescribeRepositoryAssociation(ctx, &codegurureviewer.DescribeRepositoryAssociationInput{
		AssociationArn: aws.String("arn:aws:codeguru-reviewer:us-east-1:000000000000:association:nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent repository association")
	}
}
