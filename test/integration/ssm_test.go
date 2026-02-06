//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func newSSMClient(t *testing.T) *ssm.Client {
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

	return ssm.NewFromConfig(cfg, func(o *ssm.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestSSM_PutAndGetParameter(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()
	paramName := "/test/param1"
	paramValue := "test-value"

	// Put parameter.
	putOutput, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(paramName),
		Value: aws.String(paramValue),
		Type:  types.ParameterTypeString,
	})
	if err != nil {
		t.Fatalf("failed to put parameter: %v", err)
	}

	if putOutput.Version != 1 {
		t.Errorf("expected version 1, got %d", putOutput.Version)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
			Name: aws.String(paramName),
		})
	})

	// Get parameter.
	getOutput, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(paramName),
	})
	if err != nil {
		t.Fatalf("failed to get parameter: %v", err)
	}

	if *getOutput.Parameter.Name != paramName {
		t.Errorf("expected name %s, got %s", paramName, *getOutput.Parameter.Name)
	}

	if *getOutput.Parameter.Value != paramValue {
		t.Errorf("expected value %s, got %s", paramValue, *getOutput.Parameter.Value)
	}
}

func TestSSM_PutParameter_Update(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()
	paramName := "/test/param-update"

	// Put initial parameter.
	_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(paramName),
		Value: aws.String("initial-value"),
		Type:  types.ParameterTypeString,
	})
	if err != nil {
		t.Fatalf("failed to put parameter: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
			Name: aws.String(paramName),
		})
	})

	// Update parameter without Overwrite - should fail.
	_, err = client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(paramName),
		Value: aws.String("new-value"),
		Type:  types.ParameterTypeString,
	})
	if err == nil {
		t.Fatal("expected error when updating without Overwrite")
	}

	// Update parameter with Overwrite - should succeed.
	putOutput, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Value:     aws.String("updated-value"),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to update parameter: %v", err)
	}

	if putOutput.Version != 2 {
		t.Errorf("expected version 2, got %d", putOutput.Version)
	}
}

func TestSSM_GetParameters(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()

	params := []struct {
		name  string
		value string
	}{
		{"/test/multi/param1", "value1"},
		{"/test/multi/param2", "value2"},
		{"/test/multi/param3", "value3"},
	}

	// Put parameters.
	for _, p := range params {
		_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
			Name:  aws.String(p.name),
			Value: aws.String(p.value),
			Type:  types.ParameterTypeString,
		})
		if err != nil {
			t.Fatalf("failed to put parameter %s: %v", p.name, err)
		}
	}

	t.Cleanup(func() {
		names := make([]string, len(params))
		for i, p := range params {
			names[i] = p.name
		}

		_, _ = client.DeleteParameters(ctx, &ssm.DeleteParametersInput{
			Names: names,
		})
	})

	// Get parameters including one invalid.
	names := []string{"/test/multi/param1", "/test/multi/param2", "/test/multi/nonexistent"}
	getOutput, err := client.GetParameters(ctx, &ssm.GetParametersInput{
		Names: names,
	})
	if err != nil {
		t.Fatalf("failed to get parameters: %v", err)
	}

	if len(getOutput.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(getOutput.Parameters))
	}

	if len(getOutput.InvalidParameters) != 1 {
		t.Errorf("expected 1 invalid parameter, got %d", len(getOutput.InvalidParameters))
	}

	if getOutput.InvalidParameters[0] != "/test/multi/nonexistent" {
		t.Errorf("expected invalid parameter /test/multi/nonexistent, got %s", getOutput.InvalidParameters[0])
	}
}

func TestSSM_GetParametersByPath(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()

	params := []struct {
		name  string
		value string
	}{
		{"/myapp/config/param1", "value1"},
		{"/myapp/config/param2", "value2"},
		{"/myapp/config/nested/param3", "value3"},
	}

	// Put parameters.
	for _, p := range params {
		_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
			Name:  aws.String(p.name),
			Value: aws.String(p.value),
			Type:  types.ParameterTypeString,
		})
		if err != nil {
			t.Fatalf("failed to put parameter %s: %v", p.name, err)
		}
	}

	t.Cleanup(func() {
		names := make([]string, len(params))
		for i, p := range params {
			names[i] = p.name
		}

		_, _ = client.DeleteParameters(ctx, &ssm.DeleteParametersInput{
			Names: names,
		})
	})

	// Get by path non-recursive.
	getOutput, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path: aws.String("/myapp/config"),
	})
	if err != nil {
		t.Fatalf("failed to get parameters by path: %v", err)
	}

	if len(getOutput.Parameters) != 2 {
		t.Errorf("expected 2 parameters (non-recursive), got %d", len(getOutput.Parameters))
	}

	// Get by path recursive.
	getOutput, err = client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:      aws.String("/myapp/config"),
		Recursive: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to get parameters by path: %v", err)
	}

	if len(getOutput.Parameters) != 3 {
		t.Errorf("expected 3 parameters (recursive), got %d", len(getOutput.Parameters))
	}
}

func TestSSM_DeleteParameter(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()
	paramName := "/test/param-delete"

	// Put parameter.
	_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(paramName),
		Value: aws.String("test-value"),
		Type:  types.ParameterTypeString,
	})
	if err != nil {
		t.Fatalf("failed to put parameter: %v", err)
	}

	// Delete parameter.
	_, err = client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(paramName),
	})
	if err != nil {
		t.Fatalf("failed to delete parameter: %v", err)
	}

	// Get parameter - should fail.
	_, err = client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(paramName),
	})
	if err == nil {
		t.Fatal("expected error when getting deleted parameter")
	}
}

func TestSSM_DeleteParameters(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()

	params := []string{"/test/delete/param1", "/test/delete/param2"}

	// Put parameters.
	for _, name := range params {
		_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
			Name:  aws.String(name),
			Value: aws.String("test-value"),
			Type:  types.ParameterTypeString,
		})
		if err != nil {
			t.Fatalf("failed to put parameter %s: %v", name, err)
		}
	}

	// Delete parameters including one that doesn't exist.
	names := append(params, "/test/delete/nonexistent")
	deleteOutput, err := client.DeleteParameters(ctx, &ssm.DeleteParametersInput{
		Names: names,
	})
	if err != nil {
		t.Fatalf("failed to delete parameters: %v", err)
	}

	if len(deleteOutput.DeletedParameters) != 2 {
		t.Errorf("expected 2 deleted parameters, got %d", len(deleteOutput.DeletedParameters))
	}

	if len(deleteOutput.InvalidParameters) != 1 {
		t.Errorf("expected 1 invalid parameter, got %d", len(deleteOutput.InvalidParameters))
	}
}

func TestSSM_DescribeParameters(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()

	params := []struct {
		name        string
		value       string
		description string
	}{
		{"/test/describe/param1", "value1", "Test parameter 1"},
		{"/test/describe/param2", "value2", "Test parameter 2"},
	}

	// Put parameters.
	for _, p := range params {
		_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
			Name:        aws.String(p.name),
			Value:       aws.String(p.value),
			Type:        types.ParameterTypeString,
			Description: aws.String(p.description),
		})
		if err != nil {
			t.Fatalf("failed to put parameter %s: %v", p.name, err)
		}
	}

	t.Cleanup(func() {
		names := make([]string, len(params))
		for i, p := range params {
			names[i] = p.name
		}

		_, _ = client.DeleteParameters(ctx, &ssm.DeleteParametersInput{
			Names: names,
		})
	})

	// Describe parameters.
	descOutput, err := client.DescribeParameters(ctx, &ssm.DescribeParametersInput{})
	if err != nil {
		t.Fatalf("failed to describe parameters: %v", err)
	}

	if len(descOutput.Parameters) < 2 {
		t.Errorf("expected at least 2 parameters, got %d", len(descOutput.Parameters))
	}

	// Check if our parameters are in the list.
	found := 0

	for _, p := range descOutput.Parameters {
		for _, expected := range params {
			if *p.Name == expected.name {
				found++

				if p.Description != nil && *p.Description != expected.description {
					t.Errorf("expected description %s, got %s", expected.description, *p.Description)
				}

				break
			}
		}
	}

	if found != 2 {
		t.Errorf("expected to find 2 parameters, found %d", found)
	}
}

func TestSSM_GetParameter_NotFound(t *testing.T) {
	client := newSSMClient(t)
	ctx := t.Context()

	_, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String("/nonexistent/param"),
	})
	if err == nil {
		t.Fatal("expected error when getting nonexistent parameter")
	}
}
