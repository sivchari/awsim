//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
)

func newAthenaClient(t *testing.T) *athena.Client {
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

	return athena.NewFromConfig(cfg, func(o *athena.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestAthena_StartQueryExecution(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Start query execution.
	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT 1"),
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: aws.String("default"),
		},
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: aws.String("s3://test-bucket/results/"),
		},
	})
	if err != nil {
		t.Fatalf("failed to start query execution: %v", err)
	}

	if startOutput.QueryExecutionId == nil || *startOutput.QueryExecutionId == "" {
		t.Fatal("query execution ID is empty")
	}

	t.Logf("Started query execution: %s", *startOutput.QueryExecutionId)
}

func TestAthena_GetQueryExecution(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Start query execution.
	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT * FROM test"),
	})
	if err != nil {
		t.Fatalf("failed to start query execution: %v", err)
	}

	queryExecutionID := *startOutput.QueryExecutionId

	// Get query execution.
	getOutput, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
		QueryExecutionId: aws.String(queryExecutionID),
	})
	if err != nil {
		t.Fatalf("failed to get query execution: %v", err)
	}

	if getOutput.QueryExecution == nil {
		t.Fatal("query execution is nil")
	}

	if *getOutput.QueryExecution.QueryExecutionId != queryExecutionID {
		t.Errorf("query execution ID mismatch: got %s, want %s",
			*getOutput.QueryExecution.QueryExecutionId, queryExecutionID)
	}

	if getOutput.QueryExecution.Status == nil {
		t.Fatal("query execution status is nil")
	}

	t.Logf("Query execution state: %s", getOutput.QueryExecution.Status.State)
}

func TestAthena_GetQueryResults(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Start query execution.
	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT column1, column2 FROM test"),
	})
	if err != nil {
		t.Fatalf("failed to start query execution: %v", err)
	}

	queryExecutionID := *startOutput.QueryExecutionId

	// Get query results.
	resultsOutput, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionID),
	})
	if err != nil {
		t.Fatalf("failed to get query results: %v", err)
	}

	if resultsOutput.ResultSet == nil {
		t.Fatal("result set is nil")
	}

	t.Logf("Got %d rows in result set", len(resultsOutput.ResultSet.Rows))
}

func TestAthena_ListQueryExecutions(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Start a few query executions.
	for i := 0; i < 3; i++ {
		_, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
			QueryString: aws.String("SELECT 1"),
		})
		if err != nil {
			t.Fatalf("failed to start query execution %d: %v", i, err)
		}
	}

	// List query executions.
	listOutput, err := client.ListQueryExecutions(ctx, &athena.ListQueryExecutionsInput{
		MaxResults: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to list query executions: %v", err)
	}

	if len(listOutput.QueryExecutionIds) == 0 {
		t.Fatal("no query execution IDs returned")
	}

	t.Logf("Listed %d query executions", len(listOutput.QueryExecutionIds))
}

func TestAthena_CreateAndDeleteWorkGroup(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()
	workGroupName := "test-workgroup-create-delete"

	// Create workgroup.
	_, err := client.CreateWorkGroup(ctx, &athena.CreateWorkGroupInput{
		Name:        aws.String(workGroupName),
		Description: aws.String("Test workgroup for integration tests"),
	})
	if err != nil {
		t.Fatalf("failed to create workgroup: %v", err)
	}

	t.Logf("Created workgroup: %s", workGroupName)

	// Delete workgroup.
	_, err = client.DeleteWorkGroup(ctx, &athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(workGroupName),
	})
	if err != nil {
		t.Fatalf("failed to delete workgroup: %v", err)
	}

	t.Logf("Deleted workgroup: %s", workGroupName)
}

func TestAthena_StopQueryExecution(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Start query execution.
	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT * FROM large_table"),
	})
	if err != nil {
		t.Fatalf("failed to start query execution: %v", err)
	}

	queryExecutionID := *startOutput.QueryExecutionId

	// Stop query execution.
	_, err = client.StopQueryExecution(ctx, &athena.StopQueryExecutionInput{
		QueryExecutionId: aws.String(queryExecutionID),
	})
	if err != nil {
		t.Fatalf("failed to stop query execution: %v", err)
	}

	t.Logf("Stopped query execution: %s", queryExecutionID)
}

func TestAthena_QueryExecutionWithWorkGroup(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()
	workGroupName := "test-workgroup-query"

	// Create workgroup.
	_, err := client.CreateWorkGroup(ctx, &athena.CreateWorkGroupInput{
		Name: aws.String(workGroupName),
	})
	if err != nil {
		t.Fatalf("failed to create workgroup: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteWorkGroup(ctx, &athena.DeleteWorkGroupInput{
			WorkGroup:             aws.String(workGroupName),
			RecursiveDeleteOption: aws.Bool(true),
		})
	})

	// Start query execution in the workgroup.
	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT 1"),
		WorkGroup:   aws.String(workGroupName),
	})
	if err != nil {
		t.Fatalf("failed to start query execution: %v", err)
	}

	// Get query execution and verify workgroup.
	getOutput, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
		QueryExecutionId: startOutput.QueryExecutionId,
	})
	if err != nil {
		t.Fatalf("failed to get query execution: %v", err)
	}

	if getOutput.QueryExecution.WorkGroup == nil || *getOutput.QueryExecution.WorkGroup != workGroupName {
		t.Errorf("workgroup mismatch: got %v, want %s", getOutput.QueryExecution.WorkGroup, workGroupName)
	}
}

func TestAthena_QueryExecutionNotFound(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Try to get a non-existent query execution.
	_, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
		QueryExecutionId: aws.String("non-existent-id"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent query execution")
	}
}

func TestAthena_WorkGroupNotFound(t *testing.T) {
	client := newAthenaClient(t)
	ctx := t.Context()

	// Try to start query in a non-existent workgroup.
	_, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT 1"),
		WorkGroup:   aws.String("non-existent-workgroup"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent workgroup")
	}
}
