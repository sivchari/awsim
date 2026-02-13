//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
)

func newAPIGatewayClient(t *testing.T) *apigateway.Client {
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

	return apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/apigateway")
	})
}

func TestAPIGateway_CreateAndGetRestApi(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	apiName := "test-rest-api"

	// Create REST API.
	createOutput, err := client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name:        aws.String(apiName),
		Description: aws.String("Test REST API"),
	})
	if err != nil {
		t.Fatalf("failed to create REST API: %v", err)
	}

	if createOutput.Id == nil {
		t.Fatal("REST API ID is nil")
	}

	if *createOutput.Name != apiName {
		t.Errorf("REST API name mismatch: got %s, want %s", *createOutput.Name, apiName)
	}

	t.Logf("Created REST API: %s", *createOutput.Id)

	// Get REST API.
	getOutput, err := client.GetRestApi(ctx, &apigateway.GetRestApiInput{
		RestApiId: createOutput.Id,
	})
	if err != nil {
		t.Fatalf("failed to get REST API: %v", err)
	}

	if *getOutput.Name != apiName {
		t.Errorf("REST API name mismatch: got %s, want %s", *getOutput.Name, apiName)
	}

	t.Logf("Got REST API: %s", *getOutput.Name)
}

func TestAPIGateway_GetRestApis(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	// Create a REST API first.
	createOutput, err := client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name: aws.String("test-list-api"),
	})
	if err != nil {
		t.Fatalf("failed to create REST API: %v", err)
	}

	// Get REST APIs.
	listOutput, err := client.GetRestApis(ctx, &apigateway.GetRestApisInput{})
	if err != nil {
		t.Fatalf("failed to get REST APIs: %v", err)
	}

	found := false

	for _, api := range listOutput.Items {
		if api.Id != nil && *api.Id == *createOutput.Id {
			found = true

			break
		}
	}

	if !found {
		t.Error("created REST API not found in list")
	}

	t.Logf("Listed %d REST APIs", len(listOutput.Items))
}

func TestAPIGateway_CreateAndGetResource(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	// Create REST API.
	apiOutput, err := client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name: aws.String("test-resource-api"),
	})
	if err != nil {
		t.Fatalf("failed to create REST API: %v", err)
	}

	// Get resources to find root resource.
	resourcesOutput, err := client.GetResources(ctx, &apigateway.GetResourcesInput{
		RestApiId: apiOutput.Id,
	})
	if err != nil {
		t.Fatalf("failed to get resources: %v", err)
	}

	var rootResourceID string

	for _, res := range resourcesOutput.Items {
		if res.Path != nil && *res.Path == "/" {
			rootResourceID = *res.Id

			break
		}
	}

	if rootResourceID == "" {
		t.Fatal("root resource not found")
	}

	// Create resource.
	createOutput, err := client.CreateResource(ctx, &apigateway.CreateResourceInput{
		RestApiId: apiOutput.Id,
		ParentId:  aws.String(rootResourceID),
		PathPart:  aws.String("users"),
	})
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	if *createOutput.PathPart != "users" {
		t.Errorf("resource path part mismatch: got %s, want users", *createOutput.PathPart)
	}

	if *createOutput.Path != "/users" {
		t.Errorf("resource path mismatch: got %s, want /users", *createOutput.Path)
	}

	t.Logf("Created resource: %s", *createOutput.Id)

	// Get resource.
	getOutput, err := client.GetResource(ctx, &apigateway.GetResourceInput{
		RestApiId:  apiOutput.Id,
		ResourceId: createOutput.Id,
	})
	if err != nil {
		t.Fatalf("failed to get resource: %v", err)
	}

	if *getOutput.Path != "/users" {
		t.Errorf("resource path mismatch: got %s, want /users", *getOutput.Path)
	}

	t.Logf("Got resource: %s", *getOutput.Path)
}

func TestAPIGateway_PutMethodAndIntegration(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	// Create REST API.
	apiOutput, err := client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name: aws.String("test-method-api"),
	})
	if err != nil {
		t.Fatalf("failed to create REST API: %v", err)
	}

	// Get root resource.
	resourcesOutput, err := client.GetResources(ctx, &apigateway.GetResourcesInput{
		RestApiId: apiOutput.Id,
	})
	if err != nil {
		t.Fatalf("failed to get resources: %v", err)
	}

	var rootResourceID string

	for _, res := range resourcesOutput.Items {
		if res.Path != nil && *res.Path == "/" {
			rootResourceID = *res.Id

			break
		}
	}

	// Put method.
	methodOutput, err := client.PutMethod(ctx, &apigateway.PutMethodInput{
		RestApiId:         apiOutput.Id,
		ResourceId:        aws.String(rootResourceID),
		HttpMethod:        aws.String("GET"),
		AuthorizationType: aws.String("NONE"),
	})
	if err != nil {
		t.Fatalf("failed to put method: %v", err)
	}

	if *methodOutput.HttpMethod != "GET" {
		t.Errorf("method HTTP method mismatch: got %s, want GET", *methodOutput.HttpMethod)
	}

	t.Logf("Put method: %s", *methodOutput.HttpMethod)

	// Put integration.
	integrationOutput, err := client.PutIntegration(ctx, &apigateway.PutIntegrationInput{
		RestApiId:  apiOutput.Id,
		ResourceId: aws.String(rootResourceID),
		HttpMethod: aws.String("GET"),
		Type:       types.IntegrationTypeMock,
	})
	if err != nil {
		t.Fatalf("failed to put integration: %v", err)
	}

	if integrationOutput.Type != types.IntegrationTypeMock {
		t.Errorf("integration type mismatch: got %s, want MOCK", integrationOutput.Type)
	}

	t.Logf("Put integration: %s", integrationOutput.Type)
}

func TestAPIGateway_CreateDeploymentAndStage(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	// Create REST API.
	apiOutput, err := client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name: aws.String("test-deployment-api"),
	})
	if err != nil {
		t.Fatalf("failed to create REST API: %v", err)
	}

	// Create deployment.
	deploymentOutput, err := client.CreateDeployment(ctx, &apigateway.CreateDeploymentInput{
		RestApiId:   apiOutput.Id,
		Description: aws.String("Test deployment"),
	})
	if err != nil {
		t.Fatalf("failed to create deployment: %v", err)
	}

	if deploymentOutput.Id == nil {
		t.Fatal("deployment ID is nil")
	}

	t.Logf("Created deployment: %s", *deploymentOutput.Id)

	// Create stage.
	stageOutput, err := client.CreateStage(ctx, &apigateway.CreateStageInput{
		RestApiId:    apiOutput.Id,
		StageName:    aws.String("prod"),
		DeploymentId: deploymentOutput.Id,
	})
	if err != nil {
		t.Fatalf("failed to create stage: %v", err)
	}

	if *stageOutput.StageName != "prod" {
		t.Errorf("stage name mismatch: got %s, want prod", *stageOutput.StageName)
	}

	t.Logf("Created stage: %s", *stageOutput.StageName)

	// Get stage.
	getStageOutput, err := client.GetStage(ctx, &apigateway.GetStageInput{
		RestApiId: apiOutput.Id,
		StageName: aws.String("prod"),
	})
	if err != nil {
		t.Fatalf("failed to get stage: %v", err)
	}

	if *getStageOutput.StageName != "prod" {
		t.Errorf("stage name mismatch: got %s, want prod", *getStageOutput.StageName)
	}

	t.Logf("Got stage: %s", *getStageOutput.StageName)
}

func TestAPIGateway_DeleteRestApi(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	// Create REST API.
	createOutput, err := client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name: aws.String("test-delete-api"),
	})
	if err != nil {
		t.Fatalf("failed to create REST API: %v", err)
	}

	// Delete REST API.
	_, err = client.DeleteRestApi(ctx, &apigateway.DeleteRestApiInput{
		RestApiId: createOutput.Id,
	})
	if err != nil {
		t.Fatalf("failed to delete REST API: %v", err)
	}

	t.Log("Deleted REST API successfully")

	// Verify deletion.
	_, err = client.GetRestApi(ctx, &apigateway.GetRestApiInput{
		RestApiId: createOutput.Id,
	})
	if err == nil {
		t.Error("expected error for deleted REST API")
	}

	t.Log("Verified REST API deletion")
}

func TestAPIGateway_RestApiNotFound(t *testing.T) {
	client := newAPIGatewayClient(t)
	ctx := t.Context()

	// Try to get non-existent REST API.
	_, err := client.GetRestApi(ctx, &apigateway.GetRestApiInput{
		RestApiId: aws.String("nonexistent-api"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent REST API")
	}

	t.Log("Got expected error for non-existent REST API")
}
