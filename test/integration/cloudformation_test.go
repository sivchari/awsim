//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/stretchr/testify/require"
)

const (
	testTemplate = `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Description": "Test template for CloudFormation integration tests",
		"Resources": {
			"TestBucket": {
				"Type": "AWS::S3::Bucket",
				"Properties": {
					"BucketName": "test-bucket"
				}
			}
		}
	}`

	testTemplateWithParams = `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Description": "Test template with parameters",
		"Parameters": {
			"BucketName": {
				"Type": "String",
				"Default": "default-bucket",
				"Description": "Name of the S3 bucket"
			}
		},
		"Resources": {
			"TestBucket": {
				"Type": "AWS::S3::Bucket",
				"Properties": {
					"BucketName": {"Ref": "BucketName"}
				}
			}
		}
	}`
)

func newCloudFormationClient(t *testing.T) *cloudformation.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return cloudformation.NewFromConfig(cfg, func(o *cloudformation.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCloudFormation_CreateAndDeleteStack(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-create-delete"

	// Create stack.
	createOutput, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.StackId)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// Verify stack was created.
	descOutput, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err)
	require.Len(t, descOutput.Stacks, 1)
	require.Equal(t, stackName, *descOutput.Stacks[0].StackName)
	require.Equal(t, "CREATE_COMPLETE", string(descOutput.Stacks[0].StackStatus))

	// Delete stack.
	_, err = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err)

	// Verify stack is deleted (should return not found).
	_, err = client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	require.Error(t, err)
}

func TestCloudFormation_DescribeStacks(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-describe"

	// Create stack.
	_, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// Describe specific stack.
	descOutput, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err)
	require.Len(t, descOutput.Stacks, 1)
	require.Equal(t, stackName, *descOutput.Stacks[0].StackName)
	require.NotNil(t, descOutput.Stacks[0].CreationTime)

	// Describe all stacks (without filter).
	allOutput, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allOutput.Stacks), 1)
}

func TestCloudFormation_ListStacks(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-list"

	// Create stack.
	_, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// List all stacks.
	listOutput, err := client.ListStacks(ctx, &cloudformation.ListStacksInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.StackSummaries), 1)

	// Find our stack in the list.
	found := false

	for _, summary := range listOutput.StackSummaries {
		if *summary.StackName == stackName {
			found = true
			require.Equal(t, "CREATE_COMPLETE", string(summary.StackStatus))

			break
		}
	}

	require.True(t, found, "Stack not found in ListStacks response")
}

func TestCloudFormation_UpdateStack(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-update"

	// Create stack.
	_, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// Update stack with new template.
	updatedTemplate := `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Description": "Updated test template",
		"Resources": {
			"TestBucket": {
				"Type": "AWS::S3::Bucket",
				"Properties": {
					"BucketName": "updated-bucket"
				}
			},
			"TestBucket2": {
				"Type": "AWS::S3::Bucket",
				"Properties": {
					"BucketName": "new-bucket"
				}
			}
		}
	}`

	updateOutput, err := client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(updatedTemplate),
	})
	require.NoError(t, err)
	require.NotNil(t, updateOutput.StackId)

	// Verify stack was updated.
	descOutput, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err)
	require.Len(t, descOutput.Stacks, 1)
	require.Equal(t, "UPDATE_COMPLETE", string(descOutput.Stacks[0].StackStatus))
}

func TestCloudFormation_DescribeStackResources(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-resources"

	// Create stack.
	_, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// Describe stack resources.
	resourcesOutput, err := client.DescribeStackResources(ctx, &cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err)
	require.Len(t, resourcesOutput.StackResources, 1)
	require.Equal(t, "TestBucket", *resourcesOutput.StackResources[0].LogicalResourceId)
	require.Equal(t, "AWS::S3::Bucket", *resourcesOutput.StackResources[0].ResourceType)
	require.Equal(t, "CREATE_COMPLETE", string(resourcesOutput.StackResources[0].ResourceStatus))
}

func TestCloudFormation_GetTemplate(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-get-template"

	// Create stack.
	_, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// Get template.
	templateOutput, err := client.GetTemplate(ctx, &cloudformation.GetTemplateInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err)
	require.NotNil(t, templateOutput.TemplateBody)
	require.Contains(t, *templateOutput.TemplateBody, "TestBucket")
}

func TestCloudFormation_ValidateTemplate(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	// Validate template with parameters.
	validateOutput, err := client.ValidateTemplate(ctx, &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(testTemplateWithParams),
	})
	require.NoError(t, err)
	require.NotNil(t, validateOutput.Description)
	require.Equal(t, "Test template with parameters", *validateOutput.Description)
	require.Len(t, validateOutput.Parameters, 1)
	require.Equal(t, "BucketName", *validateOutput.Parameters[0].ParameterKey)
	require.Equal(t, "default-bucket", *validateOutput.Parameters[0].DefaultValue)

	// Validate invalid template.
	_, err = client.ValidateTemplate(ctx, &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String("invalid json"),
	})
	require.Error(t, err)
}

func TestCloudFormation_StackNotFound(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	// Describe non-existent stack.
	_, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String("non-existent-stack"),
	})
	require.Error(t, err)

	// Delete non-existent stack.
	_, err = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String("non-existent-stack"),
	})
	require.Error(t, err)

	// Update non-existent stack.
	_, err = client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    aws.String("non-existent-stack"),
		TemplateBody: aws.String(testTemplate),
	})
	require.Error(t, err)
}

func TestCloudFormation_DuplicateStack(t *testing.T) {
	client := newCloudFormationClient(t)
	ctx := t.Context()

	stackName := "test-stack-duplicate"

	// Create stack.
	_, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	// Try to create duplicate stack.
	_, err = client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(testTemplate),
	})
	require.Error(t, err)
}
