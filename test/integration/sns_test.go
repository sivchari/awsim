//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func newSNSClient(t *testing.T) *sns.Client {
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

	return sns.NewFromConfig(cfg, func(o *sns.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestSNS_CreateAndDeleteTopic(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-create-delete"

	// Create topic.
	createOutput, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	if createOutput.TopicArn == nil {
		t.Fatal("topic ARN is nil")
	}

	t.Logf("Created topic: %s", *createOutput.TopicArn)

	// Delete topic.
	_, err = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
		TopicArn: createOutput.TopicArn,
	})
	if err != nil {
		t.Fatalf("failed to delete topic: %v", err)
	}
}

func TestSNS_ListTopics(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-list"

	// Create topic.
	createOutput, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: createOutput.TopicArn,
		})
	})

	// List topics.
	listOutput, err := client.ListTopics(ctx, &sns.ListTopicsInput{})
	if err != nil {
		t.Fatalf("failed to list topics: %v", err)
	}

	found := false

	for _, topic := range listOutput.Topics {
		if topic.TopicArn != nil && *topic.TopicArn == *createOutput.TopicArn {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("topic %s not found in list", *createOutput.TopicArn)
	}
}

func TestSNS_SubscribeAndUnsubscribe(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-subscribe"

	// Create topic.
	createOutput, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: createOutput.TopicArn,
		})
	})

	// Subscribe.
	subscribeOutput, err := client.Subscribe(ctx, &sns.SubscribeInput{
		TopicArn: createOutput.TopicArn,
		Protocol: aws.String("sqs"),
		Endpoint: aws.String("arn:aws:sqs:us-east-1:000000000000:test-queue"),
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	if subscribeOutput.SubscriptionArn == nil {
		t.Fatal("subscription ARN is nil")
	}

	t.Logf("Subscribed: %s", *subscribeOutput.SubscriptionArn)

	// Unsubscribe.
	_, err = client.Unsubscribe(ctx, &sns.UnsubscribeInput{
		SubscriptionArn: subscribeOutput.SubscriptionArn,
	})
	if err != nil {
		t.Fatalf("failed to unsubscribe: %v", err)
	}
}

func TestSNS_ListSubscriptions(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-list-subs"

	// Create topic.
	createOutput, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: createOutput.TopicArn,
		})
	})

	// Subscribe.
	subscribeOutput, err := client.Subscribe(ctx, &sns.SubscribeInput{
		TopicArn: createOutput.TopicArn,
		Protocol: aws.String("sqs"),
		Endpoint: aws.String("arn:aws:sqs:us-east-1:000000000000:test-queue"),
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.Unsubscribe(ctx, &sns.UnsubscribeInput{
			SubscriptionArn: subscribeOutput.SubscriptionArn,
		})
	})

	// List subscriptions.
	listOutput, err := client.ListSubscriptions(ctx, &sns.ListSubscriptionsInput{})
	if err != nil {
		t.Fatalf("failed to list subscriptions: %v", err)
	}

	found := false

	for _, sub := range listOutput.Subscriptions {
		if sub.SubscriptionArn != nil && *sub.SubscriptionArn == *subscribeOutput.SubscriptionArn {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("subscription %s not found in list", *subscribeOutput.SubscriptionArn)
	}
}

func TestSNS_ListSubscriptionsByTopic(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-list-subs-by-topic"

	// Create topic.
	createOutput, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: createOutput.TopicArn,
		})
	})

	// Subscribe.
	subscribeOutput, err := client.Subscribe(ctx, &sns.SubscribeInput{
		TopicArn: createOutput.TopicArn,
		Protocol: aws.String("sqs"),
		Endpoint: aws.String("arn:aws:sqs:us-east-1:000000000000:test-queue"),
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.Unsubscribe(ctx, &sns.UnsubscribeInput{
			SubscriptionArn: subscribeOutput.SubscriptionArn,
		})
	})

	// List subscriptions by topic.
	listOutput, err := client.ListSubscriptionsByTopic(ctx, &sns.ListSubscriptionsByTopicInput{
		TopicArn: createOutput.TopicArn,
	})
	if err != nil {
		t.Fatalf("failed to list subscriptions by topic: %v", err)
	}

	if len(listOutput.Subscriptions) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(listOutput.Subscriptions))
	}

	if *listOutput.Subscriptions[0].SubscriptionArn != *subscribeOutput.SubscriptionArn {
		t.Errorf("subscription ARN mismatch: got %s, want %s",
			*listOutput.Subscriptions[0].SubscriptionArn, *subscribeOutput.SubscriptionArn)
	}
}

func TestSNS_Publish(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-publish"
	message := "Hello, SNS!"

	// Create topic.
	createOutput, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: createOutput.TopicArn,
		})
	})

	// Publish message.
	publishOutput, err := client.Publish(ctx, &sns.PublishInput{
		TopicArn: createOutput.TopicArn,
		Message:  aws.String(message),
		Subject:  aws.String("Test Subject"),
	})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	if publishOutput.MessageId == nil {
		t.Fatal("message ID is nil")
	}

	t.Logf("Published message: %s", *publishOutput.MessageId)
}

func TestSNS_CreateTopicIdempotent(t *testing.T) {
	client := newSNSClient(t)
	ctx := t.Context()
	topicName := "test-topic-idempotent"

	// Create topic first time.
	createOutput1, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: createOutput1.TopicArn,
		})
	})

	// Create topic second time (should return the same ARN).
	createOutput2, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		t.Fatalf("failed to create topic second time: %v", err)
	}

	if *createOutput1.TopicArn != *createOutput2.TopicArn {
		t.Errorf("topic ARN mismatch: first %s, second %s",
			*createOutput1.TopicArn, *createOutput2.TopicArn)
	}
}
