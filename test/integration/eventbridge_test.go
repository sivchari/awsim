//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
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
		t.Fatalf("failed to create event bus: %v", err)
	}

	if createOutput.EventBusArn == nil {
		t.Fatal("event bus ARN is nil")
	}

	t.Logf("Created event bus: %s", *createOutput.EventBusArn)

	// Describe event bus.
	describeOutput, err := client.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
		Name: aws.String("test-event-bus"),
	})
	if err != nil {
		t.Fatalf("failed to describe event bus: %v", err)
	}

	if *describeOutput.Name != "test-event-bus" {
		t.Errorf("name mismatch: got %s, want test-event-bus", *describeOutput.Name)
	}

	t.Logf("Described event bus: %s", *describeOutput.Name)
}

func TestEventBridge_ListEventBuses(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create an event bus first.
	_, err := client.CreateEventBus(ctx, &eventbridge.CreateEventBusInput{
		Name: aws.String("test-list-event-bus"),
	})
	if err != nil {
		t.Fatalf("failed to create event bus: %v", err)
	}

	// List event buses.
	listOutput, err := client.ListEventBuses(ctx, &eventbridge.ListEventBusesInput{
		Limit: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to list event buses: %v", err)
	}

	if len(listOutput.EventBuses) == 0 {
		t.Fatal("no event buses returned")
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

	t.Logf("Listed %d event buses", len(listOutput.EventBuses))
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
		t.Fatalf("failed to put rule: %v", err)
	}

	if putOutput.RuleArn == nil {
		t.Fatal("rule ARN is nil")
	}

	t.Logf("Created rule: %s", *putOutput.RuleArn)

	// Describe rule.
	describeOutput, err := client.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
		Name: aws.String("test-rule"),
	})
	if err != nil {
		t.Fatalf("failed to describe rule: %v", err)
	}

	if *describeOutput.Name != "test-rule" {
		t.Errorf("name mismatch: got %s, want test-rule", *describeOutput.Name)
	}

	if describeOutput.State != types.RuleStateEnabled {
		t.Errorf("state mismatch: got %s, want ENABLED", describeOutput.State)
	}

	t.Logf("Described rule: %s", *describeOutput.Name)
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
		t.Fatalf("failed to put rule: %v", err)
	}

	// List rules.
	listOutput, err := client.ListRules(ctx, &eventbridge.ListRulesInput{
		Limit: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to list rules: %v", err)
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

	t.Logf("Listed %d rules", len(listOutput.Rules))
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
		t.Fatalf("failed to put rule: %v", err)
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
		t.Fatalf("failed to put targets: %v", err)
	}

	if putTargetsOutput.FailedEntryCount != 0 {
		t.Errorf("failed entry count: got %d, want 0", putTargetsOutput.FailedEntryCount)
	}

	t.Log("Added target to rule")

	// List targets.
	listTargetsOutput, err := client.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
		Rule: aws.String("test-targets-rule"),
	})
	if err != nil {
		t.Fatalf("failed to list targets: %v", err)
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

	t.Logf("Listed %d targets", len(listTargetsOutput.Targets))
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
		t.Fatalf("failed to put rule: %v", err)
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
		t.Fatalf("failed to put targets: %v", err)
	}

	// Remove targets.
	removeOutput, err := client.RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
		Rule: aws.String("test-remove-targets-rule"),
		Ids:  []string{"target-to-remove"},
	})
	if err != nil {
		t.Fatalf("failed to remove targets: %v", err)
	}

	if removeOutput.FailedEntryCount != 0 {
		t.Errorf("failed entry count: got %d, want 0", removeOutput.FailedEntryCount)
	}

	t.Log("Removed target from rule")
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
		t.Fatalf("failed to put events: %v", err)
	}

	if putEventsOutput.FailedEntryCount != 0 {
		t.Errorf("failed entry count: got %d, want 0", putEventsOutput.FailedEntryCount)
	}

	if len(putEventsOutput.Entries) != 1 {
		t.Errorf("entries count: got %d, want 1", len(putEventsOutput.Entries))
	}

	if putEventsOutput.Entries[0].EventId == nil {
		t.Error("event ID is nil")
	}

	t.Logf("Put event with ID: %s", *putEventsOutput.Entries[0].EventId)
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
		t.Fatalf("failed to put rule: %v", err)
	}

	// Delete rule.
	_, err = client.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
		Name: aws.String("test-delete-rule"),
	})
	if err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	// Verify deletion.
	_, err = client.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
		Name: aws.String("test-delete-rule"),
	})
	if err == nil {
		t.Fatal("expected error for deleted rule")
	}

	t.Log("Deleted rule successfully")
}

func TestEventBridge_DeleteEventBus(t *testing.T) {
	client := newEventBridgeClient(t)
	ctx := t.Context()

	// Create an event bus.
	_, err := client.CreateEventBus(ctx, &eventbridge.CreateEventBusInput{
		Name: aws.String("test-delete-event-bus"),
	})
	if err != nil {
		t.Fatalf("failed to create event bus: %v", err)
	}

	// Delete event bus.
	_, err = client.DeleteEventBus(ctx, &eventbridge.DeleteEventBusInput{
		Name: aws.String("test-delete-event-bus"),
	})
	if err != nil {
		t.Fatalf("failed to delete event bus: %v", err)
	}

	// Verify deletion.
	_, err = client.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
		Name: aws.String("test-delete-event-bus"),
	})
	if err == nil {
		t.Fatal("expected error for deleted event bus")
	}

	t.Log("Deleted event bus successfully")
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

	t.Log("Got expected error for non-existent event bus")
}
