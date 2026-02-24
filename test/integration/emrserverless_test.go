//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"
	"github.com/stretchr/testify/require"
)

func newEMRServerlessClient(t *testing.T) *emrserverless.Client {
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

	return emrserverless.NewFromConfig(cfg, func(o *emrserverless.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestEMRServerless_CreateAndGetApplication(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Name:         aws.String("test-spark-app"),
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.ApplicationId)
	require.NotEmpty(t, createOutput.Arn)

	applicationID := createOutput.ApplicationId

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: applicationID,
		})
	})

	// Get application.
	getOutput, err := client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: applicationID,
	})
	require.NoError(t, err)
	require.Equal(t, *applicationID, *getOutput.Application.ApplicationId)
	require.Equal(t, "test-spark-app", *getOutput.Application.Name)
	require.Equal(t, "Spark", *getOutput.Application.Type)
	require.Equal(t, "emr-6.9.0", *getOutput.Application.ReleaseLabel)
	require.Equal(t, types.ApplicationStateCreated, getOutput.Application.State)
}

func TestEMRServerless_CreateHiveApplication(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create Hive application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Name:         aws.String("test-hive-app"),
		Type:         aws.String("Hive"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createOutput.ApplicationId)

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Verify application type.
	getOutput, err := client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)
	require.Equal(t, "Hive", *getOutput.Application.Type)
}

func TestEMRServerless_ListApplications(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create multiple applications.
	appIDs := make([]*string, 0)

	for i := 0; i < 3; i++ {
		output, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
			Type:         aws.String("Spark"),
			ReleaseLabel: aws.String("emr-6.9.0"),
		})
		require.NoError(t, err)
		appIDs = append(appIDs, output.ApplicationId)
	}

	t.Cleanup(func() {
		for _, id := range appIDs {
			_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
				ApplicationId: id,
			})
		}
	})

	// List applications.
	listOutput, err := client.ListApplications(ctx, &emrserverless.ListApplicationsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Applications), 3)
}

func TestEMRServerless_ListApplicationsWithStateFilter(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// List with CREATED state filter.
	listOutput, err := client.ListApplications(ctx, &emrserverless.ListApplicationsInput{
		States: []types.ApplicationState{types.ApplicationStateCreated},
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Applications), 1)

	// List with STARTED state filter (should not include new app).
	listOutput, err = client.ListApplications(ctx, &emrserverless.ListApplicationsInput{
		States: []types.ApplicationState{types.ApplicationStateStarted},
	})
	require.NoError(t, err)
	// May be empty or contain previously started apps.
}

func TestEMRServerless_UpdateApplication(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Update application.
	_, err = client.UpdateApplication(ctx, &emrserverless.UpdateApplicationInput{
		ApplicationId: createOutput.ApplicationId,
		ReleaseLabel:  aws.String("emr-7.0.0"),
	})
	require.NoError(t, err)

	// Verify update.
	getOutput, err := client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)
	require.Equal(t, "emr-7.0.0", *getOutput.Application.ReleaseLabel)
}

func TestEMRServerless_DeleteApplication(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	// Delete application.
	_, err = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Verify deletion.
	_, err = client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.Error(t, err)
}

func TestEMRServerless_StartAndStopApplication(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// Stop before delete if needed.
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Start application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Verify application is started.
	getOutput, err := client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)
	require.Equal(t, types.ApplicationStateStarted, getOutput.Application.State)

	// Stop application.
	_, err = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Verify application is stopped.
	getOutput, err = client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)
	require.Equal(t, types.ApplicationStateStopped, getOutput.Application.State)
}

func TestEMRServerless_StartJobRun(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Start application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Start job run.
	jobOutput, err := client.StartJobRun(ctx, &emrserverless.StartJobRunInput{
		ApplicationId:    createOutput.ApplicationId,
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/test-execution-role"),
		JobDriver: &types.JobDriverMemberSparkSubmit{
			Value: types.SparkSubmit{
				EntryPoint: aws.String("s3://bucket/script.py"),
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, jobOutput.JobRunId)
	require.NotEmpty(t, jobOutput.Arn)
	require.Equal(t, *createOutput.ApplicationId, *jobOutput.ApplicationId)
}

func TestEMRServerless_GetJobRun(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Start application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Start job run.
	jobOutput, err := client.StartJobRun(ctx, &emrserverless.StartJobRunInput{
		ApplicationId:    createOutput.ApplicationId,
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/test-execution-role"),
		JobDriver: &types.JobDriverMemberSparkSubmit{
			Value: types.SparkSubmit{
				EntryPoint: aws.String("s3://bucket/script.py"),
			},
		},
		Name: aws.String("test-job"),
	})
	require.NoError(t, err)

	// Get job run.
	getOutput, err := client.GetJobRun(ctx, &emrserverless.GetJobRunInput{
		ApplicationId: createOutput.ApplicationId,
		JobRunId:      jobOutput.JobRunId,
	})
	require.NoError(t, err)
	require.Equal(t, *jobOutput.JobRunId, *getOutput.JobRun.JobRunId)
	require.Equal(t, "test-job", *getOutput.JobRun.Name)
	require.Equal(t, types.JobRunStateRunning, getOutput.JobRun.State)
}

func TestEMRServerless_ListJobRuns(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Start application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Start multiple job runs.
	for i := 0; i < 3; i++ {
		_, err = client.StartJobRun(ctx, &emrserverless.StartJobRunInput{
			ApplicationId:    createOutput.ApplicationId,
			ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/test-execution-role"),
			JobDriver: &types.JobDriverMemberSparkSubmit{
				Value: types.SparkSubmit{
					EntryPoint: aws.String("s3://bucket/script.py"),
				},
			},
		})
		require.NoError(t, err)
	}

	// List job runs.
	listOutput, err := client.ListJobRuns(ctx, &emrserverless.ListJobRunsInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.JobRuns), 3)
}

func TestEMRServerless_CancelJobRun(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Start application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Start job run.
	jobOutput, err := client.StartJobRun(ctx, &emrserverless.StartJobRunInput{
		ApplicationId:    createOutput.ApplicationId,
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/test-execution-role"),
		JobDriver: &types.JobDriverMemberSparkSubmit{
			Value: types.SparkSubmit{
				EntryPoint: aws.String("s3://bucket/script.py"),
			},
		},
	})
	require.NoError(t, err)

	// Cancel job run.
	cancelOutput, err := client.CancelJobRun(ctx, &emrserverless.CancelJobRunInput{
		ApplicationId: createOutput.ApplicationId,
		JobRunId:      jobOutput.JobRunId,
	})
	require.NoError(t, err)
	require.Equal(t, *jobOutput.JobRunId, *cancelOutput.JobRunId)

	// Verify job is cancelled.
	getOutput, err := client.GetJobRun(ctx, &emrserverless.GetJobRunInput{
		ApplicationId: createOutput.ApplicationId,
		JobRunId:      jobOutput.JobRunId,
	})
	require.NoError(t, err)
	require.Equal(t, types.JobRunStateCancelled, getOutput.JobRun.State)
}

func TestEMRServerless_NotFoundErrors(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Get non-existent application.
	_, err := client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: aws.String("nonexistent"),
	})
	require.Error(t, err)

	// Delete non-existent application.
	_, err = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
		ApplicationId: aws.String("nonexistent"),
	})
	require.Error(t, err)

	// Start non-existent application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: aws.String("nonexistent"),
	})
	require.Error(t, err)

	// Stop non-existent application.
	_, err = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
		ApplicationId: aws.String("nonexistent"),
	})
	require.Error(t, err)
}

func TestEMRServerless_ConflictErrors(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application.
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Try to stop a CREATED application (should fail).
	_, err = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.Error(t, err)

	// Start application.
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)

	// Try to start a STARTED application (should fail).
	_, err = client.StartApplication(ctx, &emrserverless.StartApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.Error(t, err)

	// Try to delete a STARTED application (should fail).
	_, err = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.Error(t, err)
}

func TestEMRServerless_AutoStartApplication(t *testing.T) {
	client := newEMRServerlessClient(t)
	ctx := t.Context()

	// Create application with auto-start enabled (default).
	createOutput, err := client.CreateApplication(ctx, &emrserverless.CreateApplicationInput{
		Type:         aws.String("Spark"),
		ReleaseLabel: aws.String("emr-6.9.0"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
		_, _ = client.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
			ApplicationId: createOutput.ApplicationId,
		})
	})

	// Start job run on CREATED application - should auto-start.
	_, err = client.StartJobRun(ctx, &emrserverless.StartJobRunInput{
		ApplicationId:    createOutput.ApplicationId,
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/test-execution-role"),
		JobDriver: &types.JobDriverMemberSparkSubmit{
			Value: types.SparkSubmit{
				EntryPoint: aws.String("s3://bucket/script.py"),
			},
		},
	})
	require.NoError(t, err)

	// Verify application is now STARTED.
	getOutput, err := client.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: createOutput.ApplicationId,
	})
	require.NoError(t, err)
	require.Equal(t, types.ApplicationStateStarted, getOutput.Application.State)
}
