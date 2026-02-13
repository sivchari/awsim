//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirehose_CreateDeliveryStream(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createFirehoseClient(t, ctx)

	streamName := "test-delivery-stream"

	// Create delivery stream.
	result, err := client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, *result.DeliveryStreamARN)

	// Clean up.
	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
}

func TestFirehose_DescribeDeliveryStream(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createFirehoseClient(t, ctx)

	streamName := "describe-test-stream"

	// Create delivery stream.
	_, err := client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)

	// Describe delivery stream.
	result, err := client.DescribeDeliveryStream(ctx, &firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, streamName, *result.DeliveryStreamDescription.DeliveryStreamName)
	assert.Equal(t, types.DeliveryStreamStatusActive, result.DeliveryStreamDescription.DeliveryStreamStatus)

	// Clean up.
	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
}

func TestFirehose_ListDeliveryStreams(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createFirehoseClient(t, ctx)

	streamName1 := "list-test-stream-1"
	streamName2 := "list-test-stream-2"

	// Create delivery streams.
	_, err := client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName1),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)

	_, err = client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName2),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)

	// List delivery streams.
	result, err := client.ListDeliveryStreams(ctx, &firehose.ListDeliveryStreamsInput{
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.DeliveryStreamNames), 2)

	// Clean up.
	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName1),
	})
	require.NoError(t, err)

	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName2),
	})
	require.NoError(t, err)
}

func TestFirehose_PutRecord(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createFirehoseClient(t, ctx)

	streamName := "put-record-test-stream"

	// Create delivery stream.
	_, err := client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)

	// Put record.
	result, err := client.PutRecord(ctx, &firehose.PutRecordInput{
		DeliveryStreamName: aws.String(streamName),
		Record: &types.Record{
			Data: []byte("test data"),
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, *result.RecordId)

	// Clean up.
	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
}

func TestFirehose_PutRecordBatch(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createFirehoseClient(t, ctx)

	streamName := "put-record-batch-test-stream"

	// Create delivery stream.
	_, err := client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)

	// Put record batch.
	result, err := client.PutRecordBatch(ctx, &firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(streamName),
		Records: []types.Record{
			{Data: []byte("test data 1")},
			{Data: []byte("test data 2")},
			{Data: []byte("test data 3")},
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(0), result.FailedPutCount)
	assert.Len(t, result.RequestResponses, 3)

	for _, resp := range result.RequestResponses {
		assert.NotEmpty(t, *resp.RecordId)
	}

	// Clean up.
	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
}

func TestFirehose_UpdateDestination(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createFirehoseClient(t, ctx)

	streamName := "update-dest-test-stream"

	// Create delivery stream.
	_, err := client.CreateDeliveryStream(ctx, &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	})
	require.NoError(t, err)

	// Get stream description to get destination ID.
	descResult, err := client.DescribeDeliveryStream(ctx, &firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
	require.Len(t, descResult.DeliveryStreamDescription.Destinations, 1)

	destID := descResult.DeliveryStreamDescription.Destinations[0].DestinationId
	versionID := descResult.DeliveryStreamDescription.VersionId

	// Update destination.
	_, err = client.UpdateDestination(ctx, &firehose.UpdateDestinationInput{
		DeliveryStreamName:             aws.String(streamName),
		CurrentDeliveryStreamVersionId: versionID,
		DestinationId:                  destID,
		S3DestinationUpdate: &types.S3DestinationUpdate{
			BucketARN: aws.String("arn:aws:s3:::test-bucket"),
			RoleARN:   aws.String("arn:aws:iam::000000000000:role/test-role"),
		},
	})
	require.NoError(t, err)

	// Clean up.
	_, err = client.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	})
	require.NoError(t, err)
}

func createFirehoseClient(t *testing.T, ctx context.Context) *firehose.Client {
	t.Helper()

	cfg := loadAWSConfig(t, ctx)

	return firehose.NewFromConfig(cfg, func(o *firehose.Options) {
		o.BaseEndpoint = aws.String(testEndpoint)
	})
}
