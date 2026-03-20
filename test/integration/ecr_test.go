//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/sivchari/golden"
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("RepositoryArn", "RepositoryUri", "RegistryId", "CreatedAt", "ResultMetadata")).Assert(t.Name()+"_create", createOutput)

	// Describe repositories.
	describeOutput, err := client.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("RepositoryArn", "RepositoryUri", "RegistryId", "CreatedAt", "ResultMetadata")).Assert(t.Name()+"_describe", describeOutput)
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ImageDigest", "RegistryId", "ResultMetadata")).Assert(t.Name()+"_put", putOutput)

	// List images.
	listOutput, err := client.ListImages(ctx, &ecr.ListImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ImageDigest", "ResultMetadata")).Assert(t.Name()+"_list", listOutput)
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

	_, err = client.PutImage(ctx, &ecr.PutImageInput{
		RepositoryName: aws.String(repoName),
		ImageManifest:  aws.String(manifest),
		ImageTag:       aws.String("v1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Batch get image.
	batchOutput, err := client.BatchGetImage(ctx, &ecr.BatchGetImageInput{
		RepositoryName: aws.String(repoName),
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String("v1")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ImageDigest", "RegistryId", "ResultMetadata")).Assert(t.Name(), batchOutput)
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ImageDigest", "ResultMetadata")).Assert(t.Name()+"_delete", deleteOutput)

	// Verify deletion.
	listOutput, err := client.ListImages(ctx, &ecr.ListImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_list_after_delete", listOutput)
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("RepositoryArn", "RepositoryUri", "RegistryId", "CreatedAt", "ResultMetadata")).Assert(t.Name()+"_delete", deleteOutput)
}

func TestECR_GetAuthorizationToken(t *testing.T) {
	client := newECRClient(t)
	ctx := t.Context()

	// Get authorization token.
	output, err := client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("AuthorizationToken", "ExpiresAt", "ResultMetadata")).Assert(t.Name(), output)
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
}
