//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func newLambdaClient(t *testing.T) *lambda.Client {
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

	return lambda.NewFromConfig(cfg, func(o *lambda.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestLambda_CreateAndDeleteFunction(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-create-delete"

	// Create function.
	createOutput, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	if createOutput.FunctionArn == nil {
		t.Fatal("function ARN is nil")
	}

	t.Logf("Created function: %s", *createOutput.FunctionArn)

	// Delete function.
	_, err = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		t.Fatalf("failed to delete function: %v", err)
	}
}

func TestLambda_GetFunction(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-get"

	// Create function.
	_, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})
	})

	// Get function.
	getOutput, err := client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		t.Fatalf("failed to get function: %v", err)
	}

	if getOutput.Configuration == nil {
		t.Fatal("function configuration is nil")
	}

	if *getOutput.Configuration.FunctionName != functionName {
		t.Errorf("function name mismatch: got %s, want %s",
			*getOutput.Configuration.FunctionName, functionName)
	}
}

func TestLambda_ListFunctions(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-list"

	// Create function.
	createOutput, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})
	})

	// List functions.
	listOutput, err := client.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		t.Fatalf("failed to list functions: %v", err)
	}

	found := false

	for _, fn := range listOutput.Functions {
		if fn.FunctionArn != nil && *fn.FunctionArn == *createOutput.FunctionArn {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("function %s not found in list", *createOutput.FunctionArn)
	}
}

func TestLambda_Invoke(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-invoke"

	// Create function.
	_, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})
	})

	// Invoke function.
	payload := []byte(`{"key": "value"}`)
	invokeOutput, err := client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	})
	if err != nil {
		t.Fatalf("failed to invoke function: %v", err)
	}

	if invokeOutput.StatusCode == nil || *invokeOutput.StatusCode != 200 {
		t.Errorf("unexpected status code: %v", invokeOutput.StatusCode)
	}

	t.Logf("Invoke response: %s", string(invokeOutput.Payload))
}

func TestLambda_UpdateFunctionCode(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-update-code"

	// Create function.
	createOutput, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})
	})

	originalCodeSha := *createOutput.CodeSha256

	// Update function code.
	updateOutput, err := client.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		ZipFile:      []byte("new-fake-zip-content"),
	})
	if err != nil {
		t.Fatalf("failed to update function code: %v", err)
	}

	if *updateOutput.CodeSha256 == originalCodeSha {
		t.Error("code SHA256 should have changed after update")
	}
}

func TestLambda_UpdateFunctionConfiguration(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-update-config"

	// Create function.
	_, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
		MemorySize: aws.Int32(128),
		Timeout:    aws.Int32(3),
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})
	})

	// Update function configuration.
	updateOutput, err := client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		MemorySize:   aws.Int32(256),
		Timeout:      aws.Int32(30),
		Description:  aws.String("Updated description"),
	})
	if err != nil {
		t.Fatalf("failed to update function configuration: %v", err)
	}

	if *updateOutput.MemorySize != 256 {
		t.Errorf("memory size mismatch: got %d, want 256", *updateOutput.MemorySize)
	}

	if *updateOutput.Timeout != 30 {
		t.Errorf("timeout mismatch: got %d, want 30", *updateOutput.Timeout)
	}

	if *updateOutput.Description != "Updated description" {
		t.Errorf("description mismatch: got %s, want 'Updated description'", *updateOutput.Description)
	}
}

func TestLambda_CreateFunctionIdempotent(t *testing.T) {
	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-idempotent"

	// Create function first time.
	_, err := client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})
	})

	// Create function second time (should fail with conflict).
	_, err = client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      types.RuntimePython312,
		Role:         aws.String("arn:aws:iam::000000000000:role/test-role"),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
	})
	if err == nil {
		t.Error("expected error when creating duplicate function")
	}
}
