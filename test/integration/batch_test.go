//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatch_CreateComputeEnvironment(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "test-compute-env"

	// Create compute environment.
	result, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
		State:                  types.CEStateEnabled,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ceName, *result.ComputeEnvironmentName)
	assert.NotEmpty(t, *result.ComputeEnvironmentArn)

	// Clean up.
	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatch_DescribeComputeEnvironments(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "describe-test-ce"

	// Create compute environment.
	_, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Describe compute environments.
	result, err := client.DescribeComputeEnvironments(ctx, &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []string{ceName},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.ComputeEnvironments, 1)
	assert.Equal(t, ceName, *result.ComputeEnvironments[0].ComputeEnvironmentName)

	// Clean up.
	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatch_CreateJobQueue(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "jq-test-ce"
	jqName := "test-job-queue"

	// Create compute environment first.
	ceResult, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Create job queue.
	jqResult, err := client.CreateJobQueue(ctx, &batch.CreateJobQueueInput{
		JobQueueName: aws.String(jqName),
		Priority:     aws.Int32(1),
		ComputeEnvironmentOrder: []types.ComputeEnvironmentOrder{
			{
				ComputeEnvironment: ceResult.ComputeEnvironmentArn,
				Order:              aws.Int32(1),
			},
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, jqResult)
	assert.Equal(t, jqName, *jqResult.JobQueueName)
	assert.NotEmpty(t, *jqResult.JobQueueArn)

	// Clean up.
	_, err = client.DeleteJobQueue(ctx, &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jqName),
	})
	require.NoError(t, err)

	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatch_DescribeJobQueues(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "describe-jq-test-ce"
	jqName := "describe-test-jq"

	// Create compute environment first.
	ceResult, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Create job queue.
	_, err = client.CreateJobQueue(ctx, &batch.CreateJobQueueInput{
		JobQueueName: aws.String(jqName),
		Priority:     aws.Int32(1),
		ComputeEnvironmentOrder: []types.ComputeEnvironmentOrder{
			{
				ComputeEnvironment: ceResult.ComputeEnvironmentArn,
				Order:              aws.Int32(1),
			},
		},
	})
	require.NoError(t, err)

	// Describe job queues.
	result, err := client.DescribeJobQueues(ctx, &batch.DescribeJobQueuesInput{
		JobQueues: []string{jqName},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.JobQueues, 1)
	assert.Equal(t, jqName, *result.JobQueues[0].JobQueueName)

	// Clean up.
	_, err = client.DeleteJobQueue(ctx, &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jqName),
	})
	require.NoError(t, err)

	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatch_RegisterJobDefinition(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	jdName := "test-job-definition"

	// Register job definition.
	result, err := client.RegisterJobDefinition(ctx, &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(jdName),
		Type:              types.JobDefinitionTypeContainer,
		ContainerProperties: &types.ContainerProperties{
			Image:  aws.String("busybox"),
			Vcpus:  aws.Int32(1),
			Memory: aws.Int32(512),
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, jdName, *result.JobDefinitionName)
	assert.NotEmpty(t, *result.JobDefinitionArn)
	assert.Equal(t, int32(1), *result.Revision)
}

func TestBatch_SubmitJob(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "submit-test-ce"
	jqName := "submit-test-jq"
	jdName := "submit-test-jd"

	// Create compute environment.
	ceResult, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Create job queue.
	jqResult, err := client.CreateJobQueue(ctx, &batch.CreateJobQueueInput{
		JobQueueName: aws.String(jqName),
		Priority:     aws.Int32(1),
		ComputeEnvironmentOrder: []types.ComputeEnvironmentOrder{
			{
				ComputeEnvironment: ceResult.ComputeEnvironmentArn,
				Order:              aws.Int32(1),
			},
		},
	})
	require.NoError(t, err)

	// Register job definition.
	jdResult, err := client.RegisterJobDefinition(ctx, &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(jdName),
		Type:              types.JobDefinitionTypeContainer,
		ContainerProperties: &types.ContainerProperties{
			Image:  aws.String("busybox"),
			Vcpus:  aws.Int32(1),
			Memory: aws.Int32(512),
		},
	})
	require.NoError(t, err)

	// Submit job.
	jobResult, err := client.SubmitJob(ctx, &batch.SubmitJobInput{
		JobName:       aws.String("test-job"),
		JobQueue:      jqResult.JobQueueArn,
		JobDefinition: jdResult.JobDefinitionArn,
	})
	require.NoError(t, err)
	assert.NotNil(t, jobResult)
	assert.Equal(t, "test-job", *jobResult.JobName)
	assert.NotEmpty(t, *jobResult.JobId)
	assert.NotEmpty(t, *jobResult.JobArn)

	// Clean up.
	_, err = client.DeleteJobQueue(ctx, &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jqName),
	})
	require.NoError(t, err)

	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatch_DescribeJobs(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "describe-job-test-ce"
	jqName := "describe-job-test-jq"
	jdName := "describe-job-test-jd"

	// Create compute environment.
	ceResult, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Create job queue.
	jqResult, err := client.CreateJobQueue(ctx, &batch.CreateJobQueueInput{
		JobQueueName: aws.String(jqName),
		Priority:     aws.Int32(1),
		ComputeEnvironmentOrder: []types.ComputeEnvironmentOrder{
			{
				ComputeEnvironment: ceResult.ComputeEnvironmentArn,
				Order:              aws.Int32(1),
			},
		},
	})
	require.NoError(t, err)

	// Register job definition.
	jdResult, err := client.RegisterJobDefinition(ctx, &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(jdName),
		Type:              types.JobDefinitionTypeContainer,
		ContainerProperties: &types.ContainerProperties{
			Image:  aws.String("busybox"),
			Vcpus:  aws.Int32(1),
			Memory: aws.Int32(512),
		},
	})
	require.NoError(t, err)

	// Submit job.
	jobResult, err := client.SubmitJob(ctx, &batch.SubmitJobInput{
		JobName:       aws.String("describe-test-job"),
		JobQueue:      jqResult.JobQueueArn,
		JobDefinition: jdResult.JobDefinitionArn,
	})
	require.NoError(t, err)

	// Describe job.
	describeResult, err := client.DescribeJobs(ctx, &batch.DescribeJobsInput{
		Jobs: []string{*jobResult.JobId},
	})
	require.NoError(t, err)
	assert.NotNil(t, describeResult)
	assert.Len(t, describeResult.Jobs, 1)
	assert.Equal(t, *jobResult.JobId, *describeResult.Jobs[0].JobId)
	assert.Equal(t, "describe-test-job", *describeResult.Jobs[0].JobName)

	// Clean up.
	_, err = client.DeleteJobQueue(ctx, &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jqName),
	})
	require.NoError(t, err)

	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatch_TerminateJob(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createBatchClient(t)

	ceName := "terminate-job-test-ce"
	jqName := "terminate-job-test-jq"
	jdName := "terminate-job-test-jd"

	// Create compute environment.
	ceResult, err := client.CreateComputeEnvironment(ctx, &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Create job queue.
	jqResult, err := client.CreateJobQueue(ctx, &batch.CreateJobQueueInput{
		JobQueueName: aws.String(jqName),
		Priority:     aws.Int32(1),
		ComputeEnvironmentOrder: []types.ComputeEnvironmentOrder{
			{
				ComputeEnvironment: ceResult.ComputeEnvironmentArn,
				Order:              aws.Int32(1),
			},
		},
	})
	require.NoError(t, err)

	// Register job definition.
	jdResult, err := client.RegisterJobDefinition(ctx, &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(jdName),
		Type:              types.JobDefinitionTypeContainer,
		ContainerProperties: &types.ContainerProperties{
			Image:  aws.String("busybox"),
			Vcpus:  aws.Int32(1),
			Memory: aws.Int32(512),
		},
	})
	require.NoError(t, err)

	// Submit job.
	jobResult, err := client.SubmitJob(ctx, &batch.SubmitJobInput{
		JobName:       aws.String("terminate-test-job"),
		JobQueue:      jqResult.JobQueueArn,
		JobDefinition: jdResult.JobDefinitionArn,
	})
	require.NoError(t, err)

	// Terminate job.
	_, err = client.TerminateJob(ctx, &batch.TerminateJobInput{
		JobId:  jobResult.JobId,
		Reason: aws.String("Test termination"),
	})
	require.NoError(t, err)

	// Verify job is terminated.
	describeResult, err := client.DescribeJobs(ctx, &batch.DescribeJobsInput{
		Jobs: []string{*jobResult.JobId},
	})
	require.NoError(t, err)
	assert.Len(t, describeResult.Jobs, 1)
	assert.Equal(t, types.JobStatusFailed, describeResult.Jobs[0].Status)

	// Clean up.
	_, err = client.DeleteJobQueue(ctx, &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jqName),
	})
	require.NoError(t, err)

	_, err = client.DeleteComputeEnvironment(ctx, &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func createBatchClient(t *testing.T) *batch.Client {
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

	return batch.NewFromConfig(cfg, func(o *batch.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}
