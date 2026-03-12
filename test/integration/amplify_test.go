//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
)

func newAmplifyClient(t *testing.T) *amplify.Client {
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

	return amplify.NewFromConfig(cfg, func(o *amplify.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestAmplify_CreateApp(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	result, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("test-app"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	if result.App == nil {
		t.Fatal("expected App to be set")
	}

	if result.App.AppId == nil || *result.App.AppId == "" {
		t.Error("expected AppId to be set")
	}

	if *result.App.Name != "test-app" {
		t.Errorf("expected name 'test-app', got %s", *result.App.Name)
	}
}

func TestAmplify_GetApp(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	createResult, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("get-app-test"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	getResult, err := client.GetApp(ctx, &amplify.GetAppInput{
		AppId: createResult.App.AppId,
	})
	if err != nil {
		t.Fatalf("failed to get app: %v", err)
	}

	if *getResult.App.Name != "get-app-test" {
		t.Errorf("expected name 'get-app-test', got %s", *getResult.App.Name)
	}
}

func TestAmplify_ListApps(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	_, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("list-app-test"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	listResult, err := client.ListApps(ctx, &amplify.ListAppsInput{})
	if err != nil {
		t.Fatalf("failed to list apps: %v", err)
	}

	if len(listResult.Apps) == 0 {
		t.Error("expected at least one app")
	}
}

func TestAmplify_DeleteApp(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	createResult, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("delete-app-test"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	_, err = client.DeleteApp(ctx, &amplify.DeleteAppInput{
		AppId: createResult.App.AppId,
	})
	if err != nil {
		t.Fatalf("failed to delete app: %v", err)
	}

	_, err = client.GetApp(ctx, &amplify.GetAppInput{
		AppId: createResult.App.AppId,
	})
	if err == nil {
		t.Fatal("expected error for deleted app")
	}
}

func TestAmplify_UpdateApp(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	createResult, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("update-app-test"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	updateResult, err := client.UpdateApp(ctx, &amplify.UpdateAppInput{
		AppId:       createResult.App.AppId,
		Description: aws.String("updated description"),
	})
	if err != nil {
		t.Fatalf("failed to update app: %v", err)
	}

	if *updateResult.App.Description != "updated description" {
		t.Errorf("expected description 'updated description', got %s", *updateResult.App.Description)
	}
}

func TestAmplify_CreateBranch(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	appResult, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("branch-test-app"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	branchResult, err := client.CreateBranch(ctx, &amplify.CreateBranchInput{
		AppId:      appResult.App.AppId,
		BranchName: aws.String("main"),
	})
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	if branchResult.Branch == nil {
		t.Fatal("expected Branch to be set")
	}

	if *branchResult.Branch.BranchName != "main" {
		t.Errorf("expected branch name 'main', got %s", *branchResult.Branch.BranchName)
	}
}

func TestAmplify_ListBranches(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	appResult, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("list-branches-app"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	_, err = client.CreateBranch(ctx, &amplify.CreateBranchInput{
		AppId:      appResult.App.AppId,
		BranchName: aws.String("develop"),
	})
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	listResult, err := client.ListBranches(ctx, &amplify.ListBranchesInput{
		AppId: appResult.App.AppId,
	})
	if err != nil {
		t.Fatalf("failed to list branches: %v", err)
	}

	if len(listResult.Branches) != 1 {
		t.Errorf("expected 1 branch, got %d", len(listResult.Branches))
	}
}

func TestAmplify_DeleteBranch(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	appResult, err := client.CreateApp(ctx, &amplify.CreateAppInput{
		Name: aws.String("delete-branch-app"),
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	_, err = client.CreateBranch(ctx, &amplify.CreateBranchInput{
		AppId:      appResult.App.AppId,
		BranchName: aws.String("feature"),
	})
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	_, err = client.DeleteBranch(ctx, &amplify.DeleteBranchInput{
		AppId:      appResult.App.AppId,
		BranchName: aws.String("feature"),
	})
	if err != nil {
		t.Fatalf("failed to delete branch: %v", err)
	}

	_, err = client.GetBranch(ctx, &amplify.GetBranchInput{
		AppId:      appResult.App.AppId,
		BranchName: aws.String("feature"),
	})
	if err == nil {
		t.Fatal("expected error for deleted branch")
	}
}

func TestAmplify_AppNotFound(t *testing.T) {
	client := newAmplifyClient(t)
	ctx := t.Context()

	_, err := client.GetApp(ctx, &amplify.GetAppInput{
		AppId: aws.String("nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent app")
	}
}
