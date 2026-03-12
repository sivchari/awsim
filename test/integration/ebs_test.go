//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ebs"
)

func newEBSClient(t *testing.T) *ebs.Client {
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

	return ebs.NewFromConfig(cfg, func(o *ebs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestEBS_StartSnapshot(t *testing.T) {
	client := newEBSClient(t)
	ctx := t.Context()

	result, err := client.StartSnapshot(ctx, &ebs.StartSnapshotInput{
		VolumeSize:  aws.Int64(100),
		Description: aws.String("test snapshot"),
	})
	if err != nil {
		t.Fatalf("failed to start snapshot: %v", err)
	}

	if result.SnapshotId == nil || *result.SnapshotId == "" {
		t.Error("expected snapshot ID to be set")
	}

	if result.Status != "pending" {
		t.Errorf("expected status pending, got %s", string(result.Status))
	}

	if *result.VolumeSize != 100 {
		t.Errorf("expected volume size 100, got %d", *result.VolumeSize)
	}

	if *result.Description != "test snapshot" {
		t.Errorf("expected description 'test snapshot', got %s", *result.Description)
	}
}

func TestEBS_CompleteSnapshot(t *testing.T) {
	client := newEBSClient(t)
	ctx := t.Context()

	startResult, err := client.StartSnapshot(ctx, &ebs.StartSnapshotInput{
		VolumeSize: aws.Int64(50),
	})
	if err != nil {
		t.Fatalf("failed to start snapshot: %v", err)
	}

	completeResult, err := client.CompleteSnapshot(ctx, &ebs.CompleteSnapshotInput{
		SnapshotId:         startResult.SnapshotId,
		ChangedBlocksCount: aws.Int32(0),
	})
	if err != nil {
		t.Fatalf("failed to complete snapshot: %v", err)
	}

	if completeResult.Status != "completed" {
		t.Errorf("expected status completed, got %s", string(completeResult.Status))
	}
}

func TestEBS_ListSnapshotBlocks(t *testing.T) {
	client := newEBSClient(t)
	ctx := t.Context()

	startResult, err := client.StartSnapshot(ctx, &ebs.StartSnapshotInput{
		VolumeSize: aws.Int64(10),
	})
	if err != nil {
		t.Fatalf("failed to start snapshot: %v", err)
	}

	_, err = client.CompleteSnapshot(ctx, &ebs.CompleteSnapshotInput{
		SnapshotId:         startResult.SnapshotId,
		ChangedBlocksCount: aws.Int32(0),
	})
	if err != nil {
		t.Fatalf("failed to complete snapshot: %v", err)
	}

	listResult, err := client.ListSnapshotBlocks(ctx, &ebs.ListSnapshotBlocksInput{
		SnapshotId: startResult.SnapshotId,
	})
	if err != nil {
		t.Fatalf("failed to list snapshot blocks: %v", err)
	}

	if listResult.Blocks == nil {
		t.Error("expected blocks to be non-nil")
	}

	if *listResult.VolumeSize != 10 {
		t.Errorf("expected volume size 10, got %d", *listResult.VolumeSize)
	}
}

func TestEBS_SnapshotNotFound(t *testing.T) {
	client := newEBSClient(t)
	ctx := t.Context()

	_, err := client.ListSnapshotBlocks(ctx, &ebs.ListSnapshotBlocksInput{
		SnapshotId: aws.String("snap-nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent snapshot")
	}
}
