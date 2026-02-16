package batch_test

import (
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sivchari/awsim/internal/server"
	_ "github.com/sivchari/awsim/internal/service/batch"
)

func TestBatchSDK_CreateComputeEnvironment(t *testing.T) {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	awsCfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	client := batch.NewFromConfig(awsCfg, func(o *batch.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
	})

	ceName := "test-compute-env"

	// Create compute environment
	result, err := client.CreateComputeEnvironment(t.Context(), &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
		State:                  types.CEStateEnabled,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ceName, *result.ComputeEnvironmentName)
	assert.NotEmpty(t, *result.ComputeEnvironmentArn)

	// Clean up
	_, err = client.DeleteComputeEnvironment(t.Context(), &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatchSDK_DescribeComputeEnvironments(t *testing.T) {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	awsCfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	client := batch.NewFromConfig(awsCfg, func(o *batch.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
	})

	ceName := "describe-test-ce"

	// Create compute environment
	_, err = client.CreateComputeEnvironment(t.Context(), &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Describe compute environments
	result, err := client.DescribeComputeEnvironments(t.Context(), &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []string{ceName},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.ComputeEnvironments, 1)
	assert.Equal(t, ceName, *result.ComputeEnvironments[0].ComputeEnvironmentName)

	// Clean up
	_, err = client.DeleteComputeEnvironment(t.Context(), &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}

func TestBatchSDK_CreateJobQueue(t *testing.T) {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	awsCfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	client := batch.NewFromConfig(awsCfg, func(o *batch.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
	})

	ceName := "jq-test-ce"
	jqName := "test-job-queue"

	// Create compute environment first
	ceResult, err := client.CreateComputeEnvironment(t.Context(), &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(ceName),
		Type:                   types.CETypeManaged,
	})
	require.NoError(t, err)

	// Create job queue
	jqResult, err := client.CreateJobQueue(t.Context(), &batch.CreateJobQueueInput{
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

	// Clean up
	_, err = client.DeleteJobQueue(t.Context(), &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jqName),
	})
	require.NoError(t, err)

	_, err = client.DeleteComputeEnvironment(t.Context(), &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(ceName),
	})
	require.NoError(t, err)
}
