//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func newSTSClient(t *testing.T) *sts.Client {
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

	return sts.NewFromConfig(cfg, func(o *sts.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestSTS_GetCallerIdentity(t *testing.T) {
	client := newSTSClient(t)
	ctx := t.Context()

	result, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		t.Fatalf("failed to get caller identity: %v", err)
	}

	if result.Account == nil || *result.Account == "" {
		t.Error("expected Account to be set")
	}

	if result.Arn == nil || *result.Arn == "" {
		t.Error("expected Arn to be set")
	}

	if result.UserId == nil || *result.UserId == "" {
		t.Error("expected UserId to be set")
	}
}

func TestSTS_AssumeRole(t *testing.T) {
	client := newSTSClient(t)
	ctx := t.Context()

	result, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		RoleSessionName: aws.String("test-session"),
	})
	if err != nil {
		t.Fatalf("failed to assume role: %v", err)
	}

	if result.Credentials == nil {
		t.Fatal("expected Credentials to be set")
	}

	if result.Credentials.AccessKeyId == nil || *result.Credentials.AccessKeyId == "" {
		t.Error("expected AccessKeyId to be set")
	}

	if result.Credentials.SecretAccessKey == nil || *result.Credentials.SecretAccessKey == "" {
		t.Error("expected SecretAccessKey to be set")
	}

	if result.Credentials.SessionToken == nil || *result.Credentials.SessionToken == "" {
		t.Error("expected SessionToken to be set")
	}

	if result.AssumedRoleUser == nil {
		t.Fatal("expected AssumedRoleUser to be set")
	}

	if result.AssumedRoleUser.Arn == nil || *result.AssumedRoleUser.Arn == "" {
		t.Error("expected AssumedRoleUser.Arn to be set")
	}

	if result.AssumedRoleUser.AssumedRoleId == nil || *result.AssumedRoleUser.AssumedRoleId == "" {
		t.Error("expected AssumedRoleUser.AssumedRoleId to be set")
	}
}

func TestSTS_AssumeRoleWithWebIdentity(t *testing.T) {
	client := newSTSClient(t)
	ctx := t.Context()

	result, err := client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String("arn:aws:iam::000000000000:role/web-identity-role"),
		RoleSessionName:  aws.String("web-session"),
		WebIdentityToken: aws.String("mock-token"),
	})
	if err != nil {
		t.Fatalf("failed to assume role with web identity: %v", err)
	}

	if result.Credentials == nil {
		t.Fatal("expected Credentials to be set")
	}

	if result.AssumedRoleUser == nil {
		t.Fatal("expected AssumedRoleUser to be set")
	}
}

func TestSTS_GetSessionToken(t *testing.T) {
	client := newSTSClient(t)
	ctx := t.Context()

	result, err := client.GetSessionToken(ctx, &sts.GetSessionTokenInput{})
	if err != nil {
		t.Fatalf("failed to get session token: %v", err)
	}

	if result.Credentials == nil {
		t.Fatal("expected Credentials to be set")
	}

	if result.Credentials.AccessKeyId == nil || *result.Credentials.AccessKeyId == "" {
		t.Error("expected AccessKeyId to be set")
	}

	if result.Credentials.SecretAccessKey == nil || *result.Credentials.SecretAccessKey == "" {
		t.Error("expected SecretAccessKey to be set")
	}

	if result.Credentials.SessionToken == nil || *result.Credentials.SessionToken == "" {
		t.Error("expected SessionToken to be set")
	}
}

func TestSTS_GetFederationToken(t *testing.T) {
	client := newSTSClient(t)
	ctx := t.Context()

	result, err := client.GetFederationToken(ctx, &sts.GetFederationTokenInput{
		Name: aws.String("test-federated-user"),
	})
	if err != nil {
		t.Fatalf("failed to get federation token: %v", err)
	}

	if result.Credentials == nil {
		t.Fatal("expected Credentials to be set")
	}

	if result.FederatedUser == nil {
		t.Fatal("expected FederatedUser to be set")
	}

	if result.FederatedUser.Arn == nil || *result.FederatedUser.Arn == "" {
		t.Error("expected FederatedUser.Arn to be set")
	}

	if result.FederatedUser.FederatedUserId == nil || *result.FederatedUser.FederatedUserId == "" {
		t.Error("expected FederatedUser.FederatedUserId to be set")
	}
}

func TestSTS_AssumeRole_MissingRoleArn(t *testing.T) {
	client := newSTSClient(t)
	ctx := t.Context()

	_, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(""),
		RoleSessionName: aws.String("test-session"),
	})
	if err == nil {
		t.Fatal("expected error for missing RoleArn")
	}
}
