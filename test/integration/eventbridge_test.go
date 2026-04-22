//go:build integration

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/sivchari/golden"
)

func newEventBridgeClient(t *testing.T) *eventbridge.Client {
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

	return eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestEventBridge_CreateAndDescribeEventBus(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create event bus.
	createOutput, err := client.CreateEventBus(ctx, &eventbridge.CreateEventBusInput{
		Name: aws.String("test-event-bus"),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("EventBusArn", "ResultMetadata")).Assert(t.Name()+"_create", createOutput)

	// Describe event bus.
	describeOutput, err := client.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
		Name: aws.String("test-event-bus"),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("Arn", "ResultMetadata")).Assert(t.Name()+"_describe", describeOutput)
}

func TestEventBridge_ListEventBuses(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create an event bus first.
	_, err := client.CreateEventBus(ctx, &eventbridge.CreateEventBusInput{
		Name: aws.String("test-list-event-bus"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// List event buses.
	listOutput, err := client.ListEventBuses(ctx, &eventbridge.ListEventBusesInput{
		Limit: aws.Int32(10),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Default event bus should always be present.
	foundDefault := false

	for _, eb := range listOutput.EventBuses {
		if *eb.Name == "default" {
			foundDefault = true

			break
		}
	}

	if !foundDefault {
		t.Error("default event bus not found in list")
	}
}

func TestEventBridge_PutAndDescribeRule(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Put rule on default event bus.
	putOutput, err := client.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String("test-rule"),
		EventPattern: aws.String(`{"source": ["test.source"]}`),
		State:        types.RuleStateEnabled,
		Description:  aws.String("Test rule"),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("RuleArn", "ResultMetadata")).Assert(t.Name()+"_put", putOutput)

	// Describe rule.
	describeOutput, err := client.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
		Name: aws.String("test-rule"),
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("Arn", "ResultMetadata")).Assert(t.Name()+"_describe", describeOutput)
}

func TestEventBridge_ListRules(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create a rule first.
	_, err := client.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String("test-list-rule"),
		EventPattern: aws.String(`{"source": ["test.source"]}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// List rules.
	listOutput, err := client.ListRules(ctx, &eventbridge.ListRulesInput{
		Limit: aws.Int32(10),
	})
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, rule := range listOutput.Rules {
		if *rule.Name == "test-list-rule" {
			found = true

			break
		}
	}

	if !found {
		t.Error("created rule not found in list")
	}
}

func TestEventBridge_PutAndListTargets(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create a rule first.
	_, err := client.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String("test-targets-rule"),
		EventPattern: aws.String(`{"source": ["test.source"]}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Put targets.
	putTargetsOutput, err := client.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: aws.String("test-targets-rule"),
		Targets: []types.Target{
			{
				Id:  aws.String("target-1"),
				Arn: aws.String("arn:aws:lambda:us-east-1:000000000000:function:test-function"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_put_targets", putTargetsOutput)

	// List targets.
	listTargetsOutput, err := client.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
		Rule: aws.String("test-targets-rule"),
	})
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, target := range listTargetsOutput.Targets {
		if *target.Id == "target-1" {
			found = true

			break
		}
	}

	if !found {
		t.Error("created target not found in list")
	}
}

func TestEventBridge_RemoveTargets(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create a rule and add a target.
	_, err := client.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String("test-remove-targets-rule"),
		EventPattern: aws.String(`{"source": ["test.source"]}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: aws.String("test-remove-targets-rule"),
		Targets: []types.Target{
			{
				Id:  aws.String("target-to-remove"),
				Arn: aws.String("arn:aws:lambda:us-east-1:000000000000:function:test-function"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Remove targets.
	removeOutput, err := client.RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
		Rule: aws.String("test-remove-targets-rule"),
		Ids:  []string{"target-to-remove"},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name(), removeOutput)
}

func TestEventBridge_PutEvents(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Put events.
	putEventsOutput, err := client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:     aws.String("test.source"),
				DetailType: aws.String("test.detail.type"),
				Detail:     aws.String(`{"key": "value"}`),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	golden.New(t, golden.WithIgnoreFields("EventId", "ResultMetadata")).Assert(t.Name(), putEventsOutput)
}

func TestEventBridge_DeleteRule(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create a rule.
	_, err := client.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String("test-delete-rule"),
		EventPattern: aws.String(`{"source": ["test.source"]}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Delete rule.
	_, err = client.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
		Name: aws.String("test-delete-rule"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify deletion.
	_, err = client.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
		Name: aws.String("test-delete-rule"),
	})
	if err == nil {
		t.Fatal("expected error for deleted rule")
	}
}

func TestEventBridge_DeleteEventBus(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create an event bus.
	_, err := client.CreateEventBus(ctx, &eventbridge.CreateEventBusInput{
		Name: aws.String("test-delete-event-bus"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Delete event bus.
	_, err = client.DeleteEventBus(ctx, &eventbridge.DeleteEventBusInput{
		Name: aws.String("test-delete-event-bus"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify deletion.
	_, err = client.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
		Name: aws.String("test-delete-event-bus"),
	})
	if err == nil {
		t.Fatal("expected error for deleted event bus")
	}
}

func TestEventBridge_PutEvents_Delivery(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create rule with event pattern.
	_, err := client.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String("delivery-test-rule"),
		EventPattern: aws.String(`{"source": ["order.service"], "detail-type": ["OrderCreated"]}`),
		State:        types.RuleStateEnabled,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Add target to rule.
	_, err = client.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: aws.String("delivery-test-rule"),
		Targets: []types.Target{
			{
				Id:  aws.String("sqs-target"),
				Arn: aws.String("arn:aws:sqs:us-east-1:000000000000:order-queue"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Put matching event.
	_, err = client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:     aws.String("order.service"),
				DetailType: aws.String("OrderCreated"),
				Detail:     aws.String(`{"orderId": "123"}`),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Put non-matching event.
	_, err = client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:     aws.String("other.service"),
				DetailType: aws.String("SomethingElse"),
				Detail:     aws.String(`{"data": "ignored"}`),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check delivered events via kumo endpoint.
	resp, err := http.Get("http://localhost:4566/kumo/eventbridge/delivered-events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var delivered []struct {
		Source     string `json:"Source"`
		DetailType string `json:"DetailType"`
		RuleName   string `json:"RuleName"`
		TargetID   string `json:"TargetId"`
		TargetArn  string `json:"TargetArn"`
	}

	if err := json.Unmarshal(body, &delivered); err != nil {
		t.Fatal(err)
	}

	// Find our delivery.
	found := false

	for _, d := range delivered {
		if d.Source == "order.service" && d.RuleName == "delivery-test-rule" && d.TargetID == "sqs-target" {
			found = true

			break
		}
	}

	if !found {
		t.Fatalf("expected matching event to be delivered to sqs-target, got: %s", string(body))
	}
}

func TestEventBridge_EventBusNotFound(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Try to describe a non-existent event bus.
	_, err := client.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
		Name: aws.String("nonexistent-event-bus"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent event bus")
	}
}
