//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/entityresolution"
	"github.com/aws/aws-sdk-go-v2/service/entityresolution/types"
)

func newEntityResolutionClient(t *testing.T) *entityresolution.Client {
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

	return entityresolution.NewFromConfig(cfg, func(o *entityresolution.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestEntityResolution_CreateAndDeleteSchemaMapping(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	schemaName := "test-schema"

	createResult, err := client.CreateSchemaMapping(ctx, &entityresolution.CreateSchemaMappingInput{
		SchemaName: aws.String(schemaName),
		MappedInputFields: []types.SchemaInputAttribute{
			{
				FieldName: aws.String("email"),
				Type:      types.SchemaAttributeTypeEmailAddress,
			},
			{
				FieldName: aws.String("name"),
				Type:      types.SchemaAttributeTypeName,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create schema mapping: %v", err)
	}

	if *createResult.SchemaName != schemaName {
		t.Errorf("expected schema name %s, got %s", schemaName, *createResult.SchemaName)
	}

	if createResult.SchemaArn == nil || *createResult.SchemaArn == "" {
		t.Error("expected schema ARN to be set")
	}

	// Delete
	_, err = client.DeleteSchemaMapping(ctx, &entityresolution.DeleteSchemaMappingInput{
		SchemaName: aws.String(schemaName),
	})
	if err != nil {
		t.Fatalf("failed to delete schema mapping: %v", err)
	}
}

func TestEntityResolution_GetSchemaMapping(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	schemaName := "test-get-schema"

	_, err := client.CreateSchemaMapping(ctx, &entityresolution.CreateSchemaMappingInput{
		SchemaName:  aws.String(schemaName),
		Description: aws.String("test description"),
		MappedInputFields: []types.SchemaInputAttribute{
			{
				FieldName: aws.String("phone"),
				Type:      types.SchemaAttributeTypePhoneNumber,
			},
			{
				FieldName: aws.String("address"),
				Type:      types.SchemaAttributeTypeAddress,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create schema mapping: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSchemaMapping(context.Background(), &entityresolution.DeleteSchemaMappingInput{
			SchemaName: aws.String(schemaName),
		})
	})

	getResult, err := client.GetSchemaMapping(ctx, &entityresolution.GetSchemaMappingInput{
		SchemaName: aws.String(schemaName),
	})
	if err != nil {
		t.Fatalf("failed to get schema mapping: %v", err)
	}

	if *getResult.SchemaName != schemaName {
		t.Errorf("expected schema name %s, got %s", schemaName, *getResult.SchemaName)
	}

	if *getResult.Description != "test description" {
		t.Errorf("expected description 'test description', got %s", *getResult.Description)
	}

	if len(getResult.MappedInputFields) != 2 {
		t.Errorf("expected 2 mapped input fields, got %d", len(getResult.MappedInputFields))
	}
}

func TestEntityResolution_ListSchemaMappings(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	schemaName := "test-list-schema"

	_, err := client.CreateSchemaMapping(ctx, &entityresolution.CreateSchemaMappingInput{
		SchemaName: aws.String(schemaName),
		MappedInputFields: []types.SchemaInputAttribute{
			{
				FieldName: aws.String("id"),
				Type:      types.SchemaAttributeTypeUniqueId,
			},
			{
				FieldName: aws.String("name"),
				Type:      types.SchemaAttributeTypeName,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create schema mapping: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSchemaMapping(context.Background(), &entityresolution.DeleteSchemaMappingInput{
			SchemaName: aws.String(schemaName),
		})
	})

	listResult, err := client.ListSchemaMappings(ctx, &entityresolution.ListSchemaMappingsInput{})
	if err != nil {
		t.Fatalf("failed to list schema mappings: %v", err)
	}

	found := false
	for _, s := range listResult.SchemaList {
		if *s.SchemaName == schemaName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected to find schema %s in list", schemaName)
	}
}

func TestEntityResolution_CreateAndDeleteMatchingWorkflow(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	workflowName := "test-matching-workflow"

	createResult, err := client.CreateMatchingWorkflow(ctx, &entityresolution.CreateMatchingWorkflowInput{
		WorkflowName: aws.String(workflowName),
		InputSourceConfig: []types.InputSource{
			{
				InputSourceARN: aws.String("arn:aws:glue:us-east-1:000000000000:table/db/table1"),
				SchemaName:     aws.String("test-schema"),
			},
		},
		OutputSourceConfig: []types.OutputSource{
			{
				OutputS3Path: aws.String("s3://bucket/output/"),
				Output: []types.OutputAttribute{
					{
						Name: aws.String("id"),
					},
				},
			},
		},
		ResolutionTechniques: &types.ResolutionTechniques{
			ResolutionType: types.ResolutionTypeRuleMatching,
		},
		RoleArn: aws.String("arn:aws:iam::000000000000:role/test-role"),
	})
	if err != nil {
		t.Fatalf("failed to create matching workflow: %v", err)
	}

	if *createResult.WorkflowName != workflowName {
		t.Errorf("expected workflow name %s, got %s", workflowName, *createResult.WorkflowName)
	}

	if createResult.WorkflowArn == nil || *createResult.WorkflowArn == "" {
		t.Error("expected workflow ARN to be set")
	}

	// Delete
	_, err = client.DeleteMatchingWorkflow(ctx, &entityresolution.DeleteMatchingWorkflowInput{
		WorkflowName: aws.String(workflowName),
	})
	if err != nil {
		t.Fatalf("failed to delete matching workflow: %v", err)
	}
}

func TestEntityResolution_GetMatchingWorkflow(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	workflowName := "test-get-matching-workflow"

	_, err := client.CreateMatchingWorkflow(ctx, &entityresolution.CreateMatchingWorkflowInput{
		WorkflowName: aws.String(workflowName),
		InputSourceConfig: []types.InputSource{
			{
				InputSourceARN: aws.String("arn:aws:glue:us-east-1:000000000000:table/db/table1"),
			},
		},
		OutputSourceConfig: []types.OutputSource{
			{
				OutputS3Path: aws.String("s3://bucket/output/"),
				Output: []types.OutputAttribute{
					{Name: aws.String("id")},
				},
			},
		},
		ResolutionTechniques: &types.ResolutionTechniques{
			ResolutionType: types.ResolutionTypeRuleMatching,
		},
		RoleArn: aws.String("arn:aws:iam::000000000000:role/test-role"),
	})
	if err != nil {
		t.Fatalf("failed to create matching workflow: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteMatchingWorkflow(context.Background(), &entityresolution.DeleteMatchingWorkflowInput{
			WorkflowName: aws.String(workflowName),
		})
	})

	getResult, err := client.GetMatchingWorkflow(ctx, &entityresolution.GetMatchingWorkflowInput{
		WorkflowName: aws.String(workflowName),
	})
	if err != nil {
		t.Fatalf("failed to get matching workflow: %v", err)
	}

	if *getResult.WorkflowName != workflowName {
		t.Errorf("expected workflow name %s, got %s", workflowName, *getResult.WorkflowName)
	}
}

func TestEntityResolution_CreateAndDeleteIdMappingWorkflow(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	workflowName := "test-idmapping-workflow"

	createResult, err := client.CreateIdMappingWorkflow(ctx, &entityresolution.CreateIdMappingWorkflowInput{
		WorkflowName: aws.String(workflowName),
		InputSourceConfig: []types.IdMappingWorkflowInputSource{
			{
				InputSourceARN: aws.String("arn:aws:glue:us-east-1:000000000000:table/db/table1"),
			},
		},
		IdMappingTechniques: &types.IdMappingTechniques{
			IdMappingType: types.IdMappingTypeRuleBased,
		},
	})
	if err != nil {
		t.Fatalf("failed to create ID mapping workflow: %v", err)
	}

	if *createResult.WorkflowName != workflowName {
		t.Errorf("expected workflow name %s, got %s", workflowName, *createResult.WorkflowName)
	}

	if createResult.WorkflowArn == nil || *createResult.WorkflowArn == "" {
		t.Error("expected workflow ARN to be set")
	}

	// Delete
	_, err = client.DeleteIdMappingWorkflow(ctx, &entityresolution.DeleteIdMappingWorkflowInput{
		WorkflowName: aws.String(workflowName),
	})
	if err != nil {
		t.Fatalf("failed to delete ID mapping workflow: %v", err)
	}
}

func TestEntityResolution_SchemaNotFound(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()

	_, err := client.GetSchemaMapping(ctx, &entityresolution.GetSchemaMappingInput{
		SchemaName: aws.String("non-existent-schema"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent schema")
	}
}

func TestEntityResolution_DuplicateSchema(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()
	schemaName := "test-duplicate-schema"

	_, err := client.CreateSchemaMapping(ctx, &entityresolution.CreateSchemaMappingInput{
		SchemaName: aws.String(schemaName),
		MappedInputFields: []types.SchemaInputAttribute{
			{
				FieldName: aws.String("id"),
				Type:      types.SchemaAttributeTypeUniqueId,
			},
			{
				FieldName: aws.String("name"),
				Type:      types.SchemaAttributeTypeName,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create schema mapping: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSchemaMapping(context.Background(), &entityresolution.DeleteSchemaMappingInput{
			SchemaName: aws.String(schemaName),
		})
	})

	_, err = client.CreateSchemaMapping(ctx, &entityresolution.CreateSchemaMappingInput{
		SchemaName: aws.String(schemaName),
		MappedInputFields: []types.SchemaInputAttribute{
			{
				FieldName: aws.String("id"),
				Type:      types.SchemaAttributeTypeUniqueId,
			},
			{
				FieldName: aws.String("name"),
				Type:      types.SchemaAttributeTypeName,
			},
		},
	})
	if err == nil {
		t.Fatal("expected error for duplicate schema")
	}
}

func TestEntityResolution_ListProviderServices(t *testing.T) {
	client := newEntityResolutionClient(t)
	ctx := t.Context()

	listResult, err := client.ListProviderServices(ctx, &entityresolution.ListProviderServicesInput{})
	if err != nil {
		t.Fatalf("failed to list provider services: %v", err)
	}

	if listResult.ProviderServiceSummaries == nil {
		t.Error("expected non-nil provider service summaries")
	}
}
