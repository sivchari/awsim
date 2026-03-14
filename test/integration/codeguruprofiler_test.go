//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler"
	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
)

func newCodeGuruProfilerClient(t *testing.T) *codeguruprofiler.Client {
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

	return codeguruprofiler.NewFromConfig(cfg, func(o *codeguruprofiler.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCodeGuruProfiler_CreateProfilingGroup(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	result, err := client.CreateProfilingGroup(ctx, &codeguruprofiler.CreateProfilingGroupInput{
		ProfilingGroupName: aws.String("test-group"),
	})
	if err != nil {
		t.Fatalf("failed to create profiling group: %v", err)
	}

	if result.ProfilingGroup == nil {
		t.Fatal("expected ProfilingGroup to be set")
	}

	if result.ProfilingGroup.Arn == nil || *result.ProfilingGroup.Arn == "" {
		t.Error("expected Arn to be set")
	}

	if *result.ProfilingGroup.Name != "test-group" {
		t.Errorf("expected name 'test-group', got %s", *result.ProfilingGroup.Name)
	}

	if result.ProfilingGroup.ComputePlatform != types.ComputePlatformDefault {
		t.Errorf("expected compute platform Default, got %s", result.ProfilingGroup.ComputePlatform)
	}
}

func TestCodeGuruProfiler_CreateProfilingGroupWithConfig(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	result, err := client.CreateProfilingGroup(ctx, &codeguruprofiler.CreateProfilingGroupInput{
		ProfilingGroupName: aws.String("config-group"),
		ComputePlatform:    types.ComputePlatformAwslambda,
		AgentOrchestrationConfig: &types.AgentOrchestrationConfig{
			ProfilingEnabled: aws.Bool(false),
		},
	})
	if err != nil {
		t.Fatalf("failed to create profiling group: %v", err)
	}

	if result.ProfilingGroup.ComputePlatform != types.ComputePlatformAwslambda {
		t.Errorf("expected compute platform AWSLambda, got %s", result.ProfilingGroup.ComputePlatform)
	}

	if result.ProfilingGroup.AgentOrchestrationConfig == nil {
		t.Fatal("expected AgentOrchestrationConfig to be set")
	}

	if *result.ProfilingGroup.AgentOrchestrationConfig.ProfilingEnabled {
		t.Error("expected ProfilingEnabled to be false")
	}
}

func TestCodeGuruProfiler_DescribeProfilingGroup(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	_, err := client.CreateProfilingGroup(ctx, &codeguruprofiler.CreateProfilingGroupInput{
		ProfilingGroupName: aws.String("describe-group"),
	})
	if err != nil {
		t.Fatalf("failed to create profiling group: %v", err)
	}

	result, err := client.DescribeProfilingGroup(ctx, &codeguruprofiler.DescribeProfilingGroupInput{
		ProfilingGroupName: aws.String("describe-group"),
	})
	if err != nil {
		t.Fatalf("failed to describe profiling group: %v", err)
	}

	if *result.ProfilingGroup.Name != "describe-group" {
		t.Errorf("expected name 'describe-group', got %s", *result.ProfilingGroup.Name)
	}
}

func TestCodeGuruProfiler_UpdateProfilingGroup(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	_, err := client.CreateProfilingGroup(ctx, &codeguruprofiler.CreateProfilingGroupInput{
		ProfilingGroupName: aws.String("update-group"),
	})
	if err != nil {
		t.Fatalf("failed to create profiling group: %v", err)
	}

	result, err := client.UpdateProfilingGroup(ctx, &codeguruprofiler.UpdateProfilingGroupInput{
		ProfilingGroupName: aws.String("update-group"),
		AgentOrchestrationConfig: &types.AgentOrchestrationConfig{
			ProfilingEnabled: aws.Bool(false),
		},
	})
	if err != nil {
		t.Fatalf("failed to update profiling group: %v", err)
	}

	if result.ProfilingGroup.AgentOrchestrationConfig == nil {
		t.Fatal("expected AgentOrchestrationConfig to be set")
	}

	if *result.ProfilingGroup.AgentOrchestrationConfig.ProfilingEnabled {
		t.Error("expected ProfilingEnabled to be false after update")
	}
}

func TestCodeGuruProfiler_DeleteProfilingGroup(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	_, err := client.CreateProfilingGroup(ctx, &codeguruprofiler.CreateProfilingGroupInput{
		ProfilingGroupName: aws.String("delete-group"),
	})
	if err != nil {
		t.Fatalf("failed to create profiling group: %v", err)
	}

	_, err = client.DeleteProfilingGroup(ctx, &codeguruprofiler.DeleteProfilingGroupInput{
		ProfilingGroupName: aws.String("delete-group"),
	})
	if err != nil {
		t.Fatalf("failed to delete profiling group: %v", err)
	}

	_, err = client.DescribeProfilingGroup(ctx, &codeguruprofiler.DescribeProfilingGroupInput{
		ProfilingGroupName: aws.String("delete-group"),
	})
	if err == nil {
		t.Fatal("expected error for deleted profiling group")
	}
}

func TestCodeGuruProfiler_ListProfilingGroups(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	_, err := client.CreateProfilingGroup(ctx, &codeguruprofiler.CreateProfilingGroupInput{
		ProfilingGroupName: aws.String("list-group"),
	})
	if err != nil {
		t.Fatalf("failed to create profiling group: %v", err)
	}

	result, err := client.ListProfilingGroups(ctx, &codeguruprofiler.ListProfilingGroupsInput{})
	if err != nil {
		t.Fatalf("failed to list profiling groups: %v", err)
	}

	if len(result.ProfilingGroups) == 0 {
		t.Error("expected at least one profiling group")
	}

	if len(result.ProfilingGroupNames) == 0 {
		t.Error("expected at least one profiling group name")
	}
}

func TestCodeGuruProfiler_ProfilingGroupNotFound(t *testing.T) {
	client := newCodeGuruProfilerClient(t)
	ctx := t.Context()

	_, err := client.DescribeProfilingGroup(ctx, &codeguruprofiler.DescribeProfilingGroupInput{
		ProfilingGroupName: aws.String("nonexistent-group"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent profiling group")
	}
}
