//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

func newKinesisClient(t *testing.T) *kinesis.Client {
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

	return kinesis.NewFromConfig(cfg, func(o *kinesis.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestKinesis_CreateAndDescribeStream(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	streamName := "test-stream"

	// Create stream.
	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	t.Logf("Created stream: %s", streamName)

	// Describe stream.
	describeOutput, err := client.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		t.Fatalf("failed to describe stream: %v", err)
	}

	if *describeOutput.StreamDescription.StreamName != streamName {
		t.Errorf("stream name mismatch: got %s, want %s", *describeOutput.StreamDescription.StreamName, streamName)
	}

	if describeOutput.StreamDescription.StreamStatus != types.StreamStatusActive {
		t.Errorf("stream status mismatch: got %s, want ACTIVE", describeOutput.StreamDescription.StreamStatus)
	}

	t.Logf("Described stream: %s, status: %s", *describeOutput.StreamDescription.StreamName, describeOutput.StreamDescription.StreamStatus)
}

func TestKinesis_ListStreams(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	// Create a stream first.
	streamName := "test-list-stream"

	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	// List streams.
	listOutput, err := client.ListStreams(ctx, &kinesis.ListStreamsInput{
		Limit: aws.Int32(100),
	})
	if err != nil {
		t.Fatalf("failed to list streams: %v", err)
	}

	found := false

	for _, name := range listOutput.StreamNames {
		if name == streamName {
			found = true

			break
		}
	}

	if !found {
		t.Error("created stream not found in list")
	}

	t.Logf("Listed %d streams", len(listOutput.StreamNames))
}

func TestKinesis_ListShards(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	streamName := "test-shards-stream"

	// Create a stream with multiple shards.
	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(2),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	// List shards.
	listOutput, err := client.ListShards(ctx, &kinesis.ListShardsInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		t.Fatalf("failed to list shards: %v", err)
	}

	if len(listOutput.Shards) != 2 {
		t.Errorf("shard count mismatch: got %d, want 2", len(listOutput.Shards))
	}

	t.Logf("Listed %d shards", len(listOutput.Shards))
}

func TestKinesis_PutAndGetRecords(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	streamName := "test-records-stream"

	// Create stream.
	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	// Put record.
	putOutput, err := client.PutRecord(ctx, &kinesis.PutRecordInput{
		StreamName:   aws.String(streamName),
		Data:         []byte("test data"),
		PartitionKey: aws.String("partition-1"),
	})
	if err != nil {
		t.Fatalf("failed to put record: %v", err)
	}

	if putOutput.ShardId == nil {
		t.Error("shard ID is nil")
	}

	if putOutput.SequenceNumber == nil {
		t.Error("sequence number is nil")
	}

	t.Logf("Put record to shard %s with sequence number %s", *putOutput.ShardId, *putOutput.SequenceNumber)

	// Get shard iterator.
	iteratorOutput, err := client.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
		StreamName:        aws.String(streamName),
		ShardId:           putOutput.ShardId,
		ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
	})
	if err != nil {
		t.Fatalf("failed to get shard iterator: %v", err)
	}

	if iteratorOutput.ShardIterator == nil {
		t.Error("shard iterator is nil")
	}

	// Get records.
	getOutput, err := client.GetRecords(ctx, &kinesis.GetRecordsInput{
		ShardIterator: iteratorOutput.ShardIterator,
		Limit:         aws.Int32(100),
	})
	if err != nil {
		t.Fatalf("failed to get records: %v", err)
	}

	if len(getOutput.Records) != 1 {
		t.Errorf("record count mismatch: got %d, want 1", len(getOutput.Records))
	}

	if string(getOutput.Records[0].Data) != "test data" {
		t.Errorf("record data mismatch: got %s, want 'test data'", string(getOutput.Records[0].Data))
	}

	t.Logf("Got %d records", len(getOutput.Records))
}

func TestKinesis_PutRecords(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	streamName := "test-put-records-stream"

	// Create stream.
	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	// Put multiple records.
	putOutput, err := client.PutRecords(ctx, &kinesis.PutRecordsInput{
		StreamName: aws.String(streamName),
		Records: []types.PutRecordsRequestEntry{
			{
				Data:         []byte("record 1"),
				PartitionKey: aws.String("partition-1"),
			},
			{
				Data:         []byte("record 2"),
				PartitionKey: aws.String("partition-2"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to put records: %v", err)
	}

	if putOutput.FailedRecordCount != nil && *putOutput.FailedRecordCount != 0 {
		t.Errorf("failed record count: got %d, want 0", *putOutput.FailedRecordCount)
	}

	if len(putOutput.Records) != 2 {
		t.Errorf("record count mismatch: got %d, want 2", len(putOutput.Records))
	}

	t.Logf("Put %d records", len(putOutput.Records))
}

func TestKinesis_DeleteStream(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	streamName := "test-delete-stream"

	// Create stream.
	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	// Delete stream.
	_, err = client.DeleteStream(ctx, &kinesis.DeleteStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		t.Fatalf("failed to delete stream: %v", err)
	}

	// Verify deletion.
	_, err = client.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err == nil {
		t.Fatal("expected error for deleted stream")
	}

	t.Log("Deleted stream successfully")
}

func TestKinesis_StreamNotFound(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	// Try to describe a non-existent stream.
	_, err := client.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String("nonexistent-stream"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent stream")
	}

	t.Log("Got expected error for non-existent stream")
}

func TestKinesis_GetShardIteratorTypes(t *testing.T) {
	client := newKinesisClient(t)
	ctx := t.Context()

	streamName := "test-iterator-types-stream"

	// Create stream.
	_, err := client.CreateStream(ctx, &kinesis.CreateStreamInput{
		StreamName: aws.String(streamName),
		ShardCount: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	// Put a record to get shard ID.
	putOutput, err := client.PutRecord(ctx, &kinesis.PutRecordInput{
		StreamName:   aws.String(streamName),
		Data:         []byte("test"),
		PartitionKey: aws.String("key"),
	})
	if err != nil {
		t.Fatalf("failed to put record: %v", err)
	}

	tests := []struct {
		name         string
		iteratorType types.ShardIteratorType
	}{
		{"TRIM_HORIZON", types.ShardIteratorTypeTrimHorizon},
		{"LATEST", types.ShardIteratorTypeLatest},
		{"AT_SEQUENCE_NUMBER", types.ShardIteratorTypeAtSequenceNumber},
		{"AFTER_SEQUENCE_NUMBER", types.ShardIteratorTypeAfterSequenceNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &kinesis.GetShardIteratorInput{
				StreamName:        aws.String(streamName),
				ShardId:           putOutput.ShardId,
				ShardIteratorType: tt.iteratorType,
			}

			if tt.iteratorType == types.ShardIteratorTypeAtSequenceNumber ||
				tt.iteratorType == types.ShardIteratorTypeAfterSequenceNumber {
				input.StartingSequenceNumber = putOutput.SequenceNumber
			}

			output, err := client.GetShardIterator(ctx, input)
			if err != nil {
				t.Fatalf("failed to get shard iterator: %v", err)
			}

			if output.ShardIterator == nil {
				t.Error("shard iterator is nil")
			}

			t.Logf("Got shard iterator for type %s", tt.name)
		})
	}
}
