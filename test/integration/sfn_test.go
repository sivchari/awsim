//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

func newSFNClient(t *testing.T) *sfn.Client {
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

	return sfn.NewFromConfig(cfg, func(o *sfn.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestSFN_CreateAndDescribeStateMachine(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-state-machine"
	definition := `{
		"Comment": "A simple state machine",
		"StartAt": "Pass",
		"States": {
			"Pass": {
				"Type": "Pass",
				"End": true
			}
		}
	}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create state machine.
	createOutput, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
	})
	if err != nil {
		t.Fatalf("failed to create state machine: %v", err)
	}

	if createOutput.StateMachineArn == nil {
		t.Fatal("state machine ARN is nil")
	}

	t.Logf("Created state machine: %s", *createOutput.StateMachineArn)

	// Describe state machine.
	describeOutput, err := client.DescribeStateMachine(ctx, &sfn.DescribeStateMachineInput{
		StateMachineArn: createOutput.StateMachineArn,
	})
	if err != nil {
		t.Fatalf("failed to describe state machine: %v", err)
	}

	if *describeOutput.Name != name {
		t.Errorf("name mismatch: got %s, want %s", *describeOutput.Name, name)
	}

	if *describeOutput.Definition != definition {
		t.Errorf("definition mismatch")
	}

	t.Logf("Described state machine: %s", *describeOutput.Name)
}

func TestSFN_ListStateMachines(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-list-state-machine"
	definition := `{"StartAt": "Pass", "States": {"Pass": {"Type": "Pass", "End": true}}}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create a state machine first.
	_, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
	})
	if err != nil {
		t.Fatalf("failed to create state machine: %v", err)
	}

	// List state machines.
	listOutput, err := client.ListStateMachines(ctx, &sfn.ListStateMachinesInput{
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		t.Fatalf("failed to list state machines: %v", err)
	}

	found := false

	for _, sm := range listOutput.StateMachines {
		if *sm.Name == name {
			found = true

			break
		}
	}

	if !found {
		t.Error("created state machine not found in list")
	}

	t.Logf("Listed %d state machines", len(listOutput.StateMachines))
}

func TestSFN_StartAndDescribeExecution(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-execution-state-machine"
	definition := `{"StartAt": "Pass", "States": {"Pass": {"Type": "Pass", "End": true}}}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create state machine.
	createOutput, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
	})
	if err != nil {
		t.Fatalf("failed to create state machine: %v", err)
	}

	// Start execution.
	execName := "test-execution"
	input := `{"key": "value"}`

	startOutput, err := client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: createOutput.StateMachineArn,
		Name:            aws.String(execName),
		Input:           aws.String(input),
	})
	if err != nil {
		t.Fatalf("failed to start execution: %v", err)
	}

	if startOutput.ExecutionArn == nil {
		t.Fatal("execution ARN is nil")
	}

	t.Logf("Started execution: %s", *startOutput.ExecutionArn)

	// Describe execution.
	describeOutput, err := client.DescribeExecution(ctx, &sfn.DescribeExecutionInput{
		ExecutionArn: startOutput.ExecutionArn,
	})
	if err != nil {
		t.Fatalf("failed to describe execution: %v", err)
	}

	if *describeOutput.Name != execName {
		t.Errorf("name mismatch: got %s, want %s", *describeOutput.Name, execName)
	}

	t.Logf("Described execution: %s, status: %s", *describeOutput.Name, describeOutput.Status)
}

func TestSFN_ListExecutions(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-list-execution-state-machine"
	definition := `{"StartAt": "Pass", "States": {"Pass": {"Type": "Pass", "End": true}}}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create state machine.
	createOutput, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
	})
	if err != nil {
		t.Fatalf("failed to create state machine: %v", err)
	}

	// Start an execution.
	_, err = client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: createOutput.StateMachineArn,
		Name:            aws.String("list-test-execution"),
	})
	if err != nil {
		t.Fatalf("failed to start execution: %v", err)
	}

	// List executions.
	listOutput, err := client.ListExecutions(ctx, &sfn.ListExecutionsInput{
		StateMachineArn: createOutput.StateMachineArn,
		MaxResults:      aws.Int32(100),
	})
	if err != nil {
		t.Fatalf("failed to list executions: %v", err)
	}

	if len(listOutput.Executions) < 1 {
		t.Error("expected at least one execution")
	}

	t.Logf("Listed %d executions", len(listOutput.Executions))
}

func TestSFN_GetExecutionHistory(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-history-state-machine"
	definition := `{"StartAt": "Pass", "States": {"Pass": {"Type": "Pass", "End": true}}}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create state machine.
	createOutput, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
	})
	if err != nil {
		t.Fatalf("failed to create state machine: %v", err)
	}

	// Start execution.
	startOutput, err := client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: createOutput.StateMachineArn,
		Name:            aws.String("history-test-execution"),
	})
	if err != nil {
		t.Fatalf("failed to start execution: %v", err)
	}

	// Get execution history.
	historyOutput, err := client.GetExecutionHistory(ctx, &sfn.GetExecutionHistoryInput{
		ExecutionArn: startOutput.ExecutionArn,
		MaxResults:   aws.Int32(100),
	})
	if err != nil {
		t.Fatalf("failed to get execution history: %v", err)
	}

	if len(historyOutput.Events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(historyOutput.Events))
	}

	t.Logf("Got %d history events", len(historyOutput.Events))
}

func TestSFN_DeleteStateMachine(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-delete-state-machine"
	definition := `{"StartAt": "Pass", "States": {"Pass": {"Type": "Pass", "End": true}}}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create state machine.
	createOutput, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
	})
	if err != nil {
		t.Fatalf("failed to create state machine: %v", err)
	}

	// Delete state machine.
	_, err = client.DeleteStateMachine(ctx, &sfn.DeleteStateMachineInput{
		StateMachineArn: createOutput.StateMachineArn,
	})
	if err != nil {
		t.Fatalf("failed to delete state machine: %v", err)
	}

	// Verify deletion.
	_, err = client.DescribeStateMachine(ctx, &sfn.DescribeStateMachineInput{
		StateMachineArn: createOutput.StateMachineArn,
	})
	if err == nil {
		t.Fatal("expected error for deleted state machine")
	}

	t.Log("Deleted state machine successfully")
}

func TestSFN_StateMachineNotFound(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	// Try to describe a non-existent state machine.
	_, err := client.DescribeStateMachine(ctx, &sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String("arn:aws:states:us-east-1:000000000000:stateMachine:nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent state machine")
	}

	t.Log("Got expected error for non-existent state machine")
}

func TestSFN_ExpressStateMachine(t *testing.T) {
	client := newSFNClient(t)
	ctx := t.Context()

	name := "test-express-state-machine"
	definition := `{"StartAt": "Pass", "States": {"Pass": {"Type": "Pass", "End": true}}}`
	roleArn := "arn:aws:iam::000000000000:role/test-role"

	// Create EXPRESS state machine.
	createOutput, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: aws.String(definition),
		RoleArn:    aws.String(roleArn),
		Type:       "EXPRESS",
	})
	if err != nil {
		t.Fatalf("failed to create express state machine: %v", err)
	}

	// Describe to verify type.
	describeOutput, err := client.DescribeStateMachine(ctx, &sfn.DescribeStateMachineInput{
		StateMachineArn: createOutput.StateMachineArn,
	})
	if err != nil {
		t.Fatalf("failed to describe state machine: %v", err)
	}

	if describeOutput.Type != "EXPRESS" {
		t.Errorf("type mismatch: got %s, want EXPRESS", describeOutput.Type)
	}

	t.Logf("Created EXPRESS state machine: %s", *describeOutput.Name)
}
