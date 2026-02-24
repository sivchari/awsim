//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/stretchr/testify/require"
)

func newCloudTrailClient(t *testing.T) *cloudtrail.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return cloudtrail.NewFromConfig(cfg, func(o *cloudtrail.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCloudTrail_CreateAndDeleteTrail(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	trailName := "test-trail-create-delete"
	bucketName := "test-bucket"

	// Create trail.
	createOutput, err := client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:         aws.String(trailName),
		S3BucketName: aws.String(bucketName),
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.TrailARN)
	require.Equal(t, trailName, *createOutput.Name)

	t.Cleanup(func() {
		_, _ = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
			Name: aws.String(trailName),
		})
	})

	// Verify trail was created.
	descOutput, err := client.DescribeTrails(ctx, &cloudtrail.DescribeTrailsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(descOutput.TrailList), 1)

	found := false

	for _, trail := range descOutput.TrailList {
		if *trail.Name == trailName {
			found = true

			break
		}
	}

	require.True(t, found, "Trail not found in DescribeTrails response")

	// Delete trail.
	_, err = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
		Name: aws.String(trailName),
	})
	require.NoError(t, err)

	// Verify trail is deleted.
	descOutput, err = client.DescribeTrails(ctx, &cloudtrail.DescribeTrailsInput{})
	require.NoError(t, err)

	for _, trail := range descOutput.TrailList {
		require.NotEqual(t, trailName, *trail.Name)
	}
}

func TestCloudTrail_GetTrail(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	trailName := "test-trail-get"
	bucketName := "test-bucket"

	// Create trail.
	_, err := client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:         aws.String(trailName),
		S3BucketName: aws.String(bucketName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
			Name: aws.String(trailName),
		})
	})

	// Get trail.
	getOutput, err := client.GetTrail(ctx, &cloudtrail.GetTrailInput{
		Name: aws.String(trailName),
	})
	require.NoError(t, err)
	require.NotNil(t, getOutput.Trail)
	require.Equal(t, trailName, *getOutput.Trail.Name)
	require.Equal(t, bucketName, *getOutput.Trail.S3BucketName)
}

func TestCloudTrail_DescribeTrails(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	trailName := "test-trail-describe"
	bucketName := "test-bucket"

	// Create trail.
	_, err := client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:         aws.String(trailName),
		S3BucketName: aws.String(bucketName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
			Name: aws.String(trailName),
		})
	})

	// Describe all trails.
	descOutput, err := client.DescribeTrails(ctx, &cloudtrail.DescribeTrailsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(descOutput.TrailList), 1)

	// Describe specific trail.
	descOutput, err = client.DescribeTrails(ctx, &cloudtrail.DescribeTrailsInput{
		TrailNameList: []string{trailName},
	})
	require.NoError(t, err)
	require.Len(t, descOutput.TrailList, 1)
	require.Equal(t, trailName, *descOutput.TrailList[0].Name)
}

func TestCloudTrail_StartAndStopLogging(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	trailName := "test-trail-logging"
	bucketName := "test-bucket"

	// Create trail.
	_, err := client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:         aws.String(trailName),
		S3BucketName: aws.String(bucketName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
			Name: aws.String(trailName),
		})
	})

	// Start logging.
	_, err = client.StartLogging(ctx, &cloudtrail.StartLoggingInput{
		Name: aws.String(trailName),
	})
	require.NoError(t, err)

	// Verify logging started.
	statusOutput, err := client.GetTrailStatus(ctx, &cloudtrail.GetTrailStatusInput{
		Name: aws.String(trailName),
	})
	require.NoError(t, err)
	require.True(t, statusOutput.IsLogging)

	// Stop logging.
	_, err = client.StopLogging(ctx, &cloudtrail.StopLoggingInput{
		Name: aws.String(trailName),
	})
	require.NoError(t, err)

	// Verify logging stopped.
	statusOutput, err = client.GetTrailStatus(ctx, &cloudtrail.GetTrailStatusInput{
		Name: aws.String(trailName),
	})
	require.NoError(t, err)
	require.False(t, statusOutput.IsLogging)
}

func TestCloudTrail_LookupEvents(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	// LookupEvents returns empty list for MVP.
	output, err := client.LookupEvents(ctx, &cloudtrail.LookupEventsInput{})
	require.NoError(t, err)
	require.Empty(t, output.Events)
}

func TestCloudTrail_TrailNotFound(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	// Get non-existent trail.
	_, err := client.GetTrail(ctx, &cloudtrail.GetTrailInput{
		Name: aws.String("non-existent-trail"),
	})
	require.Error(t, err)

	// Delete non-existent trail.
	_, err = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
		Name: aws.String("non-existent-trail"),
	})
	require.Error(t, err)

	// Start logging on non-existent trail.
	_, err = client.StartLogging(ctx, &cloudtrail.StartLoggingInput{
		Name: aws.String("non-existent-trail"),
	})
	require.Error(t, err)

	// Stop logging on non-existent trail.
	_, err = client.StopLogging(ctx, &cloudtrail.StopLoggingInput{
		Name: aws.String("non-existent-trail"),
	})
	require.Error(t, err)
}

func TestCloudTrail_DuplicateTrail(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	trailName := "test-trail-duplicate"
	bucketName := "test-bucket"

	// Create trail.
	_, err := client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:         aws.String(trailName),
		S3BucketName: aws.String(bucketName),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
			Name: aws.String(trailName),
		})
	})

	// Try to create duplicate trail.
	_, err = client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:         aws.String(trailName),
		S3BucketName: aws.String(bucketName),
	})
	require.Error(t, err)
}

func TestCloudTrail_CreateTrailWithOptions(t *testing.T) {
	client := newCloudTrailClient(t)
	ctx := t.Context()

	trailName := "test-trail-options"
	bucketName := "test-bucket"
	s3Prefix := "logs/"

	// Create trail with options.
	createOutput, err := client.CreateTrail(ctx, &cloudtrail.CreateTrailInput{
		Name:                       aws.String(trailName),
		S3BucketName:               aws.String(bucketName),
		S3KeyPrefix:                aws.String(s3Prefix),
		IncludeGlobalServiceEvents: aws.Bool(false),
		IsMultiRegionTrail:         aws.Bool(true),
		EnableLogFileValidation:    aws.Bool(true),
		IsOrganizationTrail:        aws.Bool(false),
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.TrailARN)
	require.Equal(t, trailName, *createOutput.Name)
	require.Equal(t, s3Prefix, *createOutput.S3KeyPrefix)
	require.False(t, createOutput.IncludeGlobalServiceEvents)
	require.True(t, createOutput.IsMultiRegionTrail)
	require.True(t, createOutput.LogFileValidationEnabled)
	require.False(t, createOutput.IsOrganizationTrail)

	t.Cleanup(func() {
		_, _ = client.DeleteTrail(ctx, &cloudtrail.DeleteTrailInput{
			Name: aws.String(trailName),
		})
	})
}
