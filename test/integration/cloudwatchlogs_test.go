//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/sivchari/golden"
)

func newCloudWatchLogsClient(t *testing.T) *cloudwatchlogs.Client {
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

	return cloudwatchlogs.NewFromConfig(cfg, func(o *cloudwatchlogs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCloudWatchLogs_CreateAndDeleteLogGroup(t *testing.T) {
	client := newCloudWatchLogsClient(t)
	ctx := t.Context()
	logGroupName := "test-create-delete-log-group"

	// Create log group
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify log group exists
	descResult, err := client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, lg := range descResult.LogGroups {
		if *lg.LogGroupName == logGroupName {
			found = true

			break
		}
	}

	if !found {
		t.Error("created log group not found in describe result")
	}

	// Delete log group
	_, err = client.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCloudWatchLogs_CreateAndDeleteLogStream(t *testing.T) {
	client := newCloudWatchLogsClient(t)
	ctx := t.Context()
	logGroupName := "test-log-stream-group"
	logStreamName := "test-log-stream"

	// Create log group
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteLogStream(context.Background(), &cloudwatchlogs.DeleteLogStreamInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: aws.String(logStreamName),
		})
		_, _ = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(logGroupName),
		})
	})

	// Create log stream
	_, err = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify log stream exists
	descResult, err := client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroupName),
		LogStreamNamePrefix: aws.String(logStreamName),
	})
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, ls := range descResult.LogStreams {
		if *ls.LogStreamName == logStreamName {
			found = true

			break
		}
	}

	if !found {
		t.Error("created log stream not found in describe result")
	}

	// Delete log stream
	_, err = client.DeleteLogStream(ctx, &cloudwatchlogs.DeleteLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCloudWatchLogs_PutAndGetLogEvents(t *testing.T) {
	client := newCloudWatchLogsClient(t)
	ctx := t.Context()
	logGroupName := "test-put-get-events-group"
	logStreamName := "test-put-get-events-stream"

	// Create log group
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteLogStream(context.Background(), &cloudwatchlogs.DeleteLogStreamInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: aws.String(logStreamName),
		})
		_, _ = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(logGroupName),
		})
	})

	// Create log stream
	_, err = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Put log events
	now := time.Now().UnixMilli()
	_, err = client.PutLogEvents(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		LogEvents: []types.InputLogEvent{
			{
				Timestamp: aws.Int64(now),
				Message:   aws.String("Test message 1"),
			},
			{
				Timestamp: aws.Int64(now + 1000),
				Message:   aws.String("Test message 2"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get log events
	getResult, err := client.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		StartFromHead: aws.Bool(true),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("Timestamp", "IngestionTime", "NextBackwardToken", "NextForwardToken", "ResultMetadata")).Assert(t.Name(), getResult)
}

func TestCloudWatchLogs_FilterLogEvents(t *testing.T) {
	client := newCloudWatchLogsClient(t)
	ctx := t.Context()
	logGroupName := "test-filter-events-group"
	logStreamName := "test-filter-events-stream"

	// Create log group
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteLogStream(context.Background(), &cloudwatchlogs.DeleteLogStreamInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: aws.String(logStreamName),
		})
		_, _ = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(logGroupName),
		})
	})

	// Create log stream
	_, err = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Put log events
	now := time.Now().UnixMilli()
	_, err = client.PutLogEvents(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		LogEvents: []types.InputLogEvent{
			{
				Timestamp: aws.Int64(now),
				Message:   aws.String("ERROR: Something went wrong"),
			},
			{
				Timestamp: aws.Int64(now + 1000),
				Message:   aws.String("INFO: All is well"),
			},
			{
				Timestamp: aws.Int64(now + 2000),
				Message:   aws.String("ERROR: Another error occurred"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Filter log events for ERROR
	filterResult, err := client.FilterLogEvents(ctx, &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		FilterPattern: aws.String("ERROR"),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("Timestamp", "IngestionTime", "EventId", "ResultMetadata")).Assert(t.Name(), filterResult)
}

func TestCloudWatchLogs_DescribeLogGroups(t *testing.T) {
	client := newCloudWatchLogsClient(t)
	ctx := t.Context()
	logGroupPrefix := "test-describe-groups-"
	logGroupName1 := logGroupPrefix + "alpha"
	logGroupName2 := logGroupPrefix + "beta"

	// Create log groups
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName1),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName2),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(logGroupName1),
		})
		_, _ = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(logGroupName2),
		})
	})

	// Describe log groups with prefix
	descResult, err := client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroupPrefix),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("Arn", "LogGroupArn", "CreationTime", "ResultMetadata")).Assert(t.Name(), descResult)
}

func TestCloudWatchLogs_DescribeLogStreams(t *testing.T) {
	client := newCloudWatchLogsClient(t)
	ctx := t.Context()
	logGroupName := "test-describe-streams-group"
	logStreamPrefix := "test-stream-"
	logStreamName1 := logStreamPrefix + "alpha"
	logStreamName2 := logStreamPrefix + "beta"

	// Create log group
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteLogStream(context.Background(), &cloudwatchlogs.DeleteLogStreamInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: aws.String(logStreamName1),
		})
		_, _ = client.DeleteLogStream(context.Background(), &cloudwatchlogs.DeleteLogStreamInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: aws.String(logStreamName2),
		})
		_, _ = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(logGroupName),
		})
	})

	// Create log streams
	_, err = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName1),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName2),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Describe log streams with prefix
	descResult, err := client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroupName),
		LogStreamNamePrefix: aws.String(logStreamPrefix),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("Arn", "CreationTime", "FirstEventTimestamp", "LastEventTimestamp", "LastIngestionTime", "UploadSequenceToken", "ResultMetadata")).Assert(t.Name(), descResult)
}
