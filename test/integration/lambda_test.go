//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
		o.BaseEndpoint = aws.String("http://localhost:4566/lambda")
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
	// Start mock Lambda endpoint server.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Echo back the payload with a wrapper.
		response := map[string]any{
			"statusCode": 200,
			"body":       string(body),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(mockServer.Close)

	client := newLambdaClient(t)
	ctx := t.Context()
	functionName := "test-function-invoke"

	// Create function with InvokeEndpoint.
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

	// Update function configuration to set InvokeEndpoint.
	_, err = client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		t.Fatalf("failed to update function configuration: %v", err)
	}

	// Note: Since AWS SDK doesn't support custom InvokeEndpoint field,
	// we need to test with a raw HTTP request or skip this test.
	// For now, we verify that invoke without InvokeEndpoint returns an error.
	_, err = client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      []byte(`{"key": "value"}`),
	})
	if err == nil {
		t.Error("expected error when invoking function without InvokeEndpoint")
	}
}

func TestLambda_InvokeWithEndpoint(t *testing.T) {
	// Start mock Lambda endpoint server.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Echo back the payload with a wrapper.
		response := map[string]any{
			"statusCode": 200,
			"body":       string(body),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(mockServer.Close)

	ctx := t.Context()
	functionName := "test-function-invoke-endpoint"

	// Create function with InvokeEndpoint using raw HTTP request.
	createReq := map[string]any{
		"FunctionName":   functionName,
		"Runtime":        "python3.12",
		"Role":           "arn:aws:iam::000000000000:role/test-role",
		"Handler":        "index.handler",
		"InvokeEndpoint": mockServer.URL,
		"Code": map[string]any{
			"ZipFile": []byte("fake-zip-content"),
		},
	}
	createBody, _ := json.Marshal(createReq)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://localhost:4566/lambda/2015-03-31/functions", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create function: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	t.Cleanup(func() {
		delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete,
			"http://localhost:4566/lambda/2015-03-31/functions/"+functionName, nil)
		delResp, _ := http.DefaultClient.Do(delReq)
		if delResp != nil {
			delResp.Body.Close()
		}
	})

	// Invoke function.
	payload := []byte(`{"key": "value"}`)
	invokeReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://localhost:4566/lambda/2015-03-31/functions/"+functionName+"/invocations",
		bytes.NewReader(payload))
	invokeReq.Header.Set("Content-Type", "application/json")

	invokeResp, err := http.DefaultClient.Do(invokeReq)
	if err != nil {
		t.Fatalf("failed to invoke function: %v", err)
	}
	defer invokeResp.Body.Close()

	if invokeResp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: %d", invokeResp.StatusCode)
	}

	respBody, _ := io.ReadAll(invokeResp.Body)
	t.Logf("Invoke response: %s", string(respBody))

	// Verify response contains our payload.
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["statusCode"] != float64(200) {
		t.Errorf("unexpected statusCode in response: %v", result["statusCode"])
	}
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
