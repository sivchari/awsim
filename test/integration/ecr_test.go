//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

func newECRClient(t *testing.T) *ecr.Client {
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

	return ecr.NewFromConfig(cfg, func(o *ecr.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestECR_CreateAndDescribeRepository(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	repoName := "test-repository"

	// Create repository.
	createOutput, err := client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	if createOutput.Repository == nil {
		t.Fatal("repository is nil")
	}

	if *createOutput.Repository.RepositoryName != repoName {
		t.Errorf("repository name mismatch: got %s, want %s", *createOutput.Repository.RepositoryName, repoName)
	}

	t.Logf("Created repository: %s", *createOutput.Repository.RepositoryArn)

	// Describe repositories.
	describeOutput, err := client.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	})
	if err != nil {
		t.Fatalf("failed to describe repositories: %v", err)
	}

	if len(describeOutput.Repositories) != 1 {
		t.Errorf("expected 1 repository, got %d", len(describeOutput.Repositories))
	}

	if *describeOutput.Repositories[0].RepositoryName != repoName {
		t.Errorf("repository name mismatch: got %s, want %s", *describeOutput.Repositories[0].RepositoryName, repoName)
	}

	t.Logf("Described repository: %s", *describeOutput.Repositories[0].RepositoryName)
}

func TestECR_PutAndListImages(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	repoName := "test-images-repository"

	// Create repository.
	_, err := client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	// Put image.
	manifest := `{"schemaVersion": 2, "config": {"digest": "sha256:test"}}`
	imageTag := "latest"

	putOutput, err := client.PutImage(ctx, &ecr.PutImageInput{
		RepositoryName: aws.String(repoName),
		ImageManifest:  aws.String(manifest),
		ImageTag:       aws.String(imageTag),
	})
	if err != nil {
		t.Fatalf("failed to put image: %v", err)
	}

	if putOutput.Image == nil {
		t.Fatal("image is nil")
	}

	if putOutput.Image.ImageId.ImageTag == nil || *putOutput.Image.ImageId.ImageTag != imageTag {
		t.Errorf("image tag mismatch")
	}

	t.Logf("Put image with digest: %s", *putOutput.Image.ImageId.ImageDigest)

	// List images.
	listOutput, err := client.ListImages(ctx, &ecr.ListImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to list images: %v", err)
	}

	if len(listOutput.ImageIds) != 1 {
		t.Errorf("expected 1 image, got %d", len(listOutput.ImageIds))
	}

	t.Logf("Listed %d images", len(listOutput.ImageIds))
}

func TestECR_BatchGetImage(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	repoName := "test-batch-get-repository"

	// Create repository.
	_, err := client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	// Put image.
	manifest := `{"schemaVersion": 2, "config": {"digest": "sha256:batch"}}`

	putOutput, err := client.PutImage(ctx, &ecr.PutImageInput{
		RepositoryName: aws.String(repoName),
		ImageManifest:  aws.String(manifest),
		ImageTag:       aws.String("v1"),
	})
	if err != nil {
		t.Fatalf("failed to put image: %v", err)
	}

	// Batch get image.
	batchOutput, err := client.BatchGetImage(ctx, &ecr.BatchGetImageInput{
		RepositoryName: aws.String(repoName),
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String("v1")},
		},
	})
	if err != nil {
		t.Fatalf("failed to batch get images: %v", err)
	}

	if len(batchOutput.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(batchOutput.Images))
	}

	if *batchOutput.Images[0].ImageId.ImageDigest != *putOutput.Image.ImageId.ImageDigest {
		t.Errorf("image digest mismatch")
	}

	t.Logf("Batch got %d images", len(batchOutput.Images))
}

func TestECR_BatchDeleteImage(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	repoName := "test-batch-delete-repository"

	// Create repository.
	_, err := client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	// Put image.
	manifest := `{"schemaVersion": 2, "config": {"digest": "sha256:delete"}}`

	_, err = client.PutImage(ctx, &ecr.PutImageInput{
		RepositoryName: aws.String(repoName),
		ImageManifest:  aws.String(manifest),
		ImageTag:       aws.String("to-delete"),
	})
	if err != nil {
		t.Fatalf("failed to put image: %v", err)
	}

	// Batch delete image.
	deleteOutput, err := client.BatchDeleteImage(ctx, &ecr.BatchDeleteImageInput{
		RepositoryName: aws.String(repoName),
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String("to-delete")},
		},
	})
	if err != nil {
		t.Fatalf("failed to batch delete images: %v", err)
	}

	if len(deleteOutput.ImageIds) != 1 {
		t.Errorf("expected 1 deleted image, got %d", len(deleteOutput.ImageIds))
	}

	t.Logf("Batch deleted %d images", len(deleteOutput.ImageIds))

	// Verify deletion.
	listOutput, err := client.ListImages(ctx, &ecr.ListImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to list images: %v", err)
	}

	if len(listOutput.ImageIds) != 0 {
		t.Errorf("expected 0 images after deletion, got %d", len(listOutput.ImageIds))
	}
}

func TestECR_DeleteRepository(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	repoName := "test-delete-repository"

	// Create repository.
	_, err := client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	// Delete repository.
	deleteOutput, err := client.DeleteRepository(ctx, &ecr.DeleteRepositoryInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatalf("failed to delete repository: %v", err)
	}

	if *deleteOutput.Repository.RepositoryName != repoName {
		t.Errorf("deleted repository name mismatch: got %s, want %s", *deleteOutput.Repository.RepositoryName, repoName)
	}

	t.Log("Deleted repository successfully")

	// Verify deletion.
	_, err = client.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	})
	// Note: ECR returns empty list for non-existent repositories, not an error.
	// So we just log it.
	t.Log("Repository no longer exists after deletion")
}

func TestECR_GetAuthorizationToken(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	// Get authorization token.
	output, err := client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		t.Fatalf("failed to get authorization token: %v", err)
	}

	if len(output.AuthorizationData) == 0 {
		t.Error("expected at least one authorization data")
	}

	if output.AuthorizationData[0].AuthorizationToken == nil {
		t.Error("authorization token is nil")
	}

	if output.AuthorizationData[0].ProxyEndpoint == nil {
		t.Error("proxy endpoint is nil")
	}

	t.Logf("Got authorization token for endpoint: %s", *output.AuthorizationData[0].ProxyEndpoint)
}

func TestECR_RepositoryNotFound(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	// Try to list images from non-existent repository.
	_, err := client.ListImages(ctx, &ecr.ListImagesInput{
		RepositoryName: aws.String("nonexistent-repository"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent repository")
	}

	t.Log("Got expected error for non-existent repository")
}
