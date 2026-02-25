//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/stretchr/testify/require"
)

func newSageMakerClient(t *testing.T) *sagemaker.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return sagemaker.NewFromConfig(cfg, func(o *sagemaker.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestSageMaker_CreateAndDeleteNotebookInstance(t *testing.T) {
	client := newSageMakerClient(t)
	ctx := t.Context()

	instanceName := "test-notebook"

	// Create notebook instance.
	createOutput, err := client.CreateNotebookInstance(ctx, &sagemaker.CreateNotebookInstanceInput{
		NotebookInstanceName: aws.String(instanceName),
		InstanceType:         types.InstanceTypeMlT2Medium,
		RoleArn:              aws.String("arn:aws:iam::123456789012:role/sagemaker-role"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.NotebookInstanceArn)

	t.Cleanup(func() {
		_, _ = client.DeleteNotebookInstance(ctx, &sagemaker.DeleteNotebookInstanceInput{
			NotebookInstanceName: aws.String(instanceName),
		})
	})

	// Describe notebook instance.
	descOutput, err := client.DescribeNotebookInstance(ctx, &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(instanceName),
	})
	require.NoError(t, err)
	require.Equal(t, instanceName, *descOutput.NotebookInstanceName)
	require.Equal(t, types.InstanceTypeMlT2Medium, descOutput.InstanceType)

	// Delete notebook instance.
	_, err = client.DeleteNotebookInstance(ctx, &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(instanceName),
	})
	require.NoError(t, err)

	// Verify notebook instance is deleted.
	_, err = client.DescribeNotebookInstance(ctx, &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(instanceName),
	})
	require.Error(t, err)
}

func TestSageMaker_ListNotebookInstances(t *testing.T) {
	client := newSageMakerClient(t)
	ctx := t.Context()

	// Create multiple notebook instances.
	instanceNames := []string{"test-notebook-1", "test-notebook-2"}

	for _, name := range instanceNames {
		_, err := client.CreateNotebookInstance(ctx, &sagemaker.CreateNotebookInstanceInput{
			NotebookInstanceName: aws.String(name),
			InstanceType:         types.InstanceTypeMlT2Medium,
			RoleArn:              aws.String("arn:aws:iam::123456789012:role/sagemaker-role"),
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, name := range instanceNames {
			_, _ = client.DeleteNotebookInstance(ctx, &sagemaker.DeleteNotebookInstanceInput{
				NotebookInstanceName: aws.String(name),
			})
		}
	})

	// List notebook instances.
	listOutput, err := client.ListNotebookInstances(ctx, &sagemaker.ListNotebookInstancesInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.NotebookInstances), 2)
}

func TestSageMaker_CreateAndDescribeTrainingJob(t *testing.T) {
	client := newSageMakerClient(t)
	ctx := t.Context()

	jobName := "test-training-job"

	// Create training job.
	createOutput, err := client.CreateTrainingJob(ctx, &sagemaker.CreateTrainingJobInput{
		TrainingJobName: aws.String(jobName),
		AlgorithmSpecification: &types.AlgorithmSpecification{
			TrainingImage:     aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/my-image:latest"),
			TrainingInputMode: types.TrainingInputModeFile,
		},
		RoleArn: aws.String("arn:aws:iam::123456789012:role/sagemaker-role"),
		OutputDataConfig: &types.OutputDataConfig{
			S3OutputPath: aws.String("s3://my-bucket/output"),
		},
		ResourceConfig: &types.ResourceConfig{
			InstanceType:   types.TrainingInstanceTypeMlM4Xlarge,
			InstanceCount:  aws.Int32(1),
			VolumeSizeInGB: aws.Int32(50),
		},
		StoppingCondition: &types.StoppingCondition{
			MaxRuntimeInSeconds: aws.Int32(3600),
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.TrainingJobArn)

	// Describe training job.
	descOutput, err := client.DescribeTrainingJob(ctx, &sagemaker.DescribeTrainingJobInput{
		TrainingJobName: aws.String(jobName),
	})
	require.NoError(t, err)
	require.Equal(t, jobName, *descOutput.TrainingJobName)
	require.Equal(t, types.TrainingJobStatusCompleted, descOutput.TrainingJobStatus)
}

func TestSageMaker_CreateAndDeleteModel(t *testing.T) {
	client := newSageMakerClient(t)
	ctx := t.Context()

	modelName := "test-model"

	// Create model.
	createOutput, err := client.CreateModel(ctx, &sagemaker.CreateModelInput{
		ModelName: aws.String(modelName),
		PrimaryContainer: &types.ContainerDefinition{
			Image: aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/my-image:latest"),
		},
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/sagemaker-role"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.ModelArn)

	t.Cleanup(func() {
		_, _ = client.DeleteModel(ctx, &sagemaker.DeleteModelInput{
			ModelName: aws.String(modelName),
		})
	})

	// Delete model.
	_, err = client.DeleteModel(ctx, &sagemaker.DeleteModelInput{
		ModelName: aws.String(modelName),
	})
	require.NoError(t, err)

	// Try to delete again (should fail).
	_, err = client.DeleteModel(ctx, &sagemaker.DeleteModelInput{
		ModelName: aws.String(modelName),
	})
	require.Error(t, err)
}

func TestSageMaker_CreateAndDeleteEndpoint(t *testing.T) {
	client := newSageMakerClient(t)
	ctx := t.Context()

	endpointName := "test-endpoint"
	endpointConfigName := "test-endpoint-config"

	// Create endpoint.
	createOutput, err := client.CreateEndpoint(ctx, &sagemaker.CreateEndpointInput{
		EndpointName:       aws.String(endpointName),
		EndpointConfigName: aws.String(endpointConfigName),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.EndpointArn)

	t.Cleanup(func() {
		_, _ = client.DeleteEndpoint(ctx, &sagemaker.DeleteEndpointInput{
			EndpointName: aws.String(endpointName),
		})
	})

	// Describe endpoint.
	descOutput, err := client.DescribeEndpoint(ctx, &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(endpointName),
	})
	require.NoError(t, err)
	require.Equal(t, endpointName, *descOutput.EndpointName)
	require.Equal(t, endpointConfigName, *descOutput.EndpointConfigName)
	require.Equal(t, types.EndpointStatusInService, descOutput.EndpointStatus)

	// Delete endpoint.
	_, err = client.DeleteEndpoint(ctx, &sagemaker.DeleteEndpointInput{
		EndpointName: aws.String(endpointName),
	})
	require.NoError(t, err)

	// Verify endpoint is deleted.
	_, err = client.DescribeEndpoint(ctx, &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(endpointName),
	})
	require.Error(t, err)
}

func TestSageMaker_NotebookInstanceNotFound(t *testing.T) {
	client := newSageMakerClient(t)
	ctx := t.Context()

	// Describe non-existent notebook instance.
	_, err := client.DescribeNotebookInstance(ctx, &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String("non-existent-notebook"),
	})
	require.Error(t, err)

	// Delete non-existent notebook instance.
	_, err = client.DeleteNotebookInstance(ctx, &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String("non-existent-notebook"),
	})
	require.Error(t, err)
}
