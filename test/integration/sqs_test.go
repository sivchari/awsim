//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func newSQSClient(t *testing.T) *sqs.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(awsimEndpoint)
	})
}

func TestSQS_CreateAndDeleteQueue(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-create-delete"

	// Create queue.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	if createOutput.QueueUrl == nil {
		t.Fatal("queue URL is nil")
	}

	t.Logf("Created queue: %s", *createOutput.QueueUrl)

	// Delete queue.
	_, err = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: createOutput.QueueUrl,
	})
	if err != nil {
		t.Fatalf("failed to delete queue: %v", err)
	}
}

func TestSQS_ListQueues(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-list"

	// Create queue.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	defer func() {
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createOutput.QueueUrl,
		})
	}()

	// List queues.
	listOutput, err := client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		t.Fatalf("failed to list queues: %v", err)
	}

	found := false

	for _, url := range listOutput.QueueUrls {
		if url == *createOutput.QueueUrl {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("queue %s not found in list", *createOutput.QueueUrl)
	}
}

func TestSQS_GetQueueUrl(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-get-url"

	// Create queue.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	defer func() {
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createOutput.QueueUrl,
		})
	}()

	// Get queue URL.
	getOutput, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to get queue URL: %v", err)
	}

	if *getOutput.QueueUrl != *createOutput.QueueUrl {
		t.Errorf("queue URL mismatch: got %s, want %s", *getOutput.QueueUrl, *createOutput.QueueUrl)
	}
}

func TestSQS_SendAndReceiveMessage(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-send-receive"
	messageBody := "Hello, SQS!"

	// Create queue.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	defer func() {
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createOutput.QueueUrl,
		})
	}()

	// Send message.
	sendOutput, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    createOutput.QueueUrl,
		MessageBody: aws.String(messageBody),
	})
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	if sendOutput.MessageId == nil {
		t.Fatal("message ID is nil")
	}

	t.Logf("Sent message: %s", *sendOutput.MessageId)

	// Receive message.
	receiveOutput, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            createOutput.QueueUrl,
		MaxNumberOfMessages: 1,
	})
	if err != nil {
		t.Fatalf("failed to receive message: %v", err)
	}

	if len(receiveOutput.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(receiveOutput.Messages))
	}

	if *receiveOutput.Messages[0].Body != messageBody {
		t.Errorf("message body mismatch: got %s, want %s", *receiveOutput.Messages[0].Body, messageBody)
	}

	// Delete message.
	_, err = client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      createOutput.QueueUrl,
		ReceiptHandle: receiveOutput.Messages[0].ReceiptHandle,
	})
	if err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}
}

func TestSQS_PurgeQueue(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-purge"

	// Create queue.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	defer func() {
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createOutput.QueueUrl,
		})
	}()

	// Send multiple messages.
	for i := 0; i < 3; i++ {
		_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    createOutput.QueueUrl,
			MessageBody: aws.String("test message"),
		})
		if err != nil {
			t.Fatalf("failed to send message: %v", err)
		}
	}

	// Purge queue.
	_, err = client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
		QueueUrl: createOutput.QueueUrl,
	})
	if err != nil {
		t.Fatalf("failed to purge queue: %v", err)
	}

	// Verify queue is empty.
	receiveOutput, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            createOutput.QueueUrl,
		MaxNumberOfMessages: 10,
	})
	if err != nil {
		t.Fatalf("failed to receive message: %v", err)
	}

	if len(receiveOutput.Messages) != 0 {
		t.Errorf("expected 0 messages after purge, got %d", len(receiveOutput.Messages))
	}
}

func TestSQS_GetQueueAttributes(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-attributes"

	// Create queue with custom attributes.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"VisibilityTimeout": "60",
		},
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	defer func() {
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createOutput.QueueUrl,
		})
	}()

	// Get queue attributes.
	getOutput, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: createOutput.QueueUrl,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
	})
	if err != nil {
		t.Fatalf("failed to get queue attributes: %v", err)
	}

	if _, ok := getOutput.Attributes["QueueArn"]; !ok {
		t.Error("QueueArn attribute not found")
	}

	if vt, ok := getOutput.Attributes["VisibilityTimeout"]; !ok || vt != "60" {
		t.Errorf("VisibilityTimeout mismatch: got %s, want 60", vt)
	}
}

func TestSQS_SetQueueAttributes(t *testing.T) {
	client := newSQSClient(t)
	ctx := context.Background()
	queueName := "test-queue-set-attributes"

	// Create queue.
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	defer func() {
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createOutput.QueueUrl,
		})
	}()

	// Set queue attributes.
	_, err = client.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl: createOutput.QueueUrl,
		Attributes: map[string]string{
			"VisibilityTimeout": "120",
		},
	})
	if err != nil {
		t.Fatalf("failed to set queue attributes: %v", err)
	}

	// Verify attributes.
	getOutput, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: createOutput.QueueUrl,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameVisibilityTimeout,
		},
	})
	if err != nil {
		t.Fatalf("failed to get queue attributes: %v", err)
	}

	if vt, ok := getOutput.Attributes["VisibilityTimeout"]; !ok || vt != "120" {
		t.Errorf("VisibilityTimeout mismatch: got %s, want 120", vt)
	}
}
