//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
)

func newElasticBeanstalkClient(t *testing.T) *elasticbeanstalk.Client {
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

	return elasticbeanstalk.NewFromConfig(cfg, func(o *elasticbeanstalk.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestElasticBeanstalk_CreateAndDeleteApplication(t *testing.T) {
	client := newElasticBeanstalkClient(t)
	ctx := t.Context()
	appName := "test-app"

	createResult, err := client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
		Description:     aws.String("test application"),
	})
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	if *createResult.Application.ApplicationName != appName {
		t.Errorf("expected app name %s, got %s", appName, *createResult.Application.ApplicationName)
	}

	if createResult.Application.ApplicationArn == nil || *createResult.Application.ApplicationArn == "" {
		t.Error("expected application ARN to be set")
	}

	// Delete
	_, err = client.DeleteApplication(ctx, &elasticbeanstalk.DeleteApplicationInput{
		ApplicationName: aws.String(appName),
	})
	if err != nil {
		t.Fatalf("failed to delete application: %v", err)
	}
}

func TestElasticBeanstalk_DescribeApplications(t *testing.T) {
	client := newElasticBeanstalkClient(t)
	ctx := t.Context()
	appName := "test-describe-app"

	_, err := client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
		Description:     aws.String("describe test"),
	})
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(t.Context(), &elasticbeanstalk.DeleteApplicationInput{
			ApplicationName: aws.String(appName),
		})
	})

	describeResult, err := client.DescribeApplications(ctx, &elasticbeanstalk.DescribeApplicationsInput{
		ApplicationNames: []string{appName},
	})
	if err != nil {
		t.Fatalf("failed to describe applications: %v", err)
	}

	if len(describeResult.Applications) != 1 {
		t.Fatalf("expected 1 application, got %d", len(describeResult.Applications))
	}

	if *describeResult.Applications[0].ApplicationName != appName {
		t.Errorf("expected app name %s, got %s", appName, *describeResult.Applications[0].ApplicationName)
	}
}

func TestElasticBeanstalk_UpdateApplication(t *testing.T) {
	client := newElasticBeanstalkClient(t)
	ctx := t.Context()
	appName := "test-update-app"

	_, err := client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
	})
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(t.Context(), &elasticbeanstalk.DeleteApplicationInput{
			ApplicationName: aws.String(appName),
		})
	})

	updateResult, err := client.UpdateApplication(ctx, &elasticbeanstalk.UpdateApplicationInput{
		ApplicationName: aws.String(appName),
		Description:     aws.String("updated description"),
	})
	if err != nil {
		t.Fatalf("failed to update application: %v", err)
	}

	if *updateResult.Application.Description != "updated description" {
		t.Errorf("expected description 'updated description', got %s", *updateResult.Application.Description)
	}
}

func TestElasticBeanstalk_CreateAndTerminateEnvironment(t *testing.T) {
	client := newElasticBeanstalkClient(t)
	ctx := t.Context()
	appName := "test-env-app"
	envName := "test-env"

	_, err := client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
	})
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(t.Context(), &elasticbeanstalk.DeleteApplicationInput{
			ApplicationName: aws.String(appName),
		})
	})

	createResult, err := client.CreateEnvironment(ctx, &elasticbeanstalk.CreateEnvironmentInput{
		ApplicationName: aws.String(appName),
		EnvironmentName: aws.String(envName),
	})
	if err != nil {
		t.Fatalf("failed to create environment: %v", err)
	}

	if *createResult.EnvironmentName != envName {
		t.Errorf("expected env name %s, got %s", envName, *createResult.EnvironmentName)
	}

	if createResult.EnvironmentId == nil || *createResult.EnvironmentId == "" {
		t.Error("expected environment ID to be set")
	}

	// Terminate
	terminateResult, err := client.TerminateEnvironment(ctx, &elasticbeanstalk.TerminateEnvironmentInput{
		EnvironmentName: aws.String(envName),
	})
	if err != nil {
		t.Fatalf("failed to terminate environment: %v", err)
	}

	if *terminateResult.EnvironmentName != envName {
		t.Errorf("expected env name %s, got %s", envName, *terminateResult.EnvironmentName)
	}
}

func TestElasticBeanstalk_DescribeEnvironments(t *testing.T) {
	client := newElasticBeanstalkClient(t)
	ctx := t.Context()
	appName := "test-desc-env-app"
	envName := "test-desc-env"

	_, err := client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
	})
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.TerminateEnvironment(t.Context(), &elasticbeanstalk.TerminateEnvironmentInput{
			EnvironmentName: aws.String(envName),
		})
		_, _ = client.DeleteApplication(t.Context(), &elasticbeanstalk.DeleteApplicationInput{
			ApplicationName: aws.String(appName),
		})
	})

	_, err = client.CreateEnvironment(ctx, &elasticbeanstalk.CreateEnvironmentInput{
		ApplicationName: aws.String(appName),
		EnvironmentName: aws.String(envName),
	})
	if err != nil {
		t.Fatalf("failed to create environment: %v", err)
	}

	describeResult, err := client.DescribeEnvironments(ctx, &elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentNames: []string{envName},
	})
	if err != nil {
		t.Fatalf("failed to describe environments: %v", err)
	}

	if len(describeResult.Environments) != 1 {
		t.Fatalf("expected 1 environment, got %d", len(describeResult.Environments))
	}

	if *describeResult.Environments[0].EnvironmentName != envName {
		t.Errorf("expected env name %s, got %s", envName, *describeResult.Environments[0].EnvironmentName)
	}
}

func TestElasticBeanstalk_DuplicateApplication(t *testing.T) {
	client := newElasticBeanstalkClient(t)
	ctx := t.Context()
	appName := "test-dup-app"

	_, err := client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
	})
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(t.Context(), &elasticbeanstalk.DeleteApplicationInput{
			ApplicationName: aws.String(appName),
		})
	})

	_, err = client.CreateApplication(ctx, &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(appName),
	})
	if err == nil {
		t.Fatal("expected error for duplicate application")
	}
}
