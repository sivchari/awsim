//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func newSecretsManagerClient(t *testing.T) *secretsmanager.Client {
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

	return secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestSecretsManager_CreateAndDeleteSecret(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-create-delete"
	secretValue := "super-secret-value"

	// Create secret.
	createOutput, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretValue),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	if createOutput.ARN == nil {
		t.Fatal("secret ARN is nil")
	}

	t.Logf("Created secret: %s", *createOutput.ARN)

	// Delete secret.
	deleteOutput, err := client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to delete secret: %v", err)
	}

	if deleteOutput.ARN == nil {
		t.Fatal("deleted secret ARN is nil")
	}
}

func TestSecretsManager_GetSecretValue(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-get-value"
	secretValue := "my-secret-value-123"

	// Create secret.
	createOutput, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretValue),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Get secret value.
	getOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to get secret value: %v", err)
	}

	if *getOutput.SecretString != secretValue {
		t.Errorf("secret value mismatch: got %s, want %s", *getOutput.SecretString, secretValue)
	}

	if *getOutput.ARN != *createOutput.ARN {
		t.Errorf("ARN mismatch: got %s, want %s", *getOutput.ARN, *createOutput.ARN)
	}
}

func TestSecretsManager_PutSecretValue(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-put-value"
	initialValue := "initial-value"
	updatedValue := "updated-value"

	// Create secret.
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(initialValue),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Put new secret value.
	putOutput, err := client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(updatedValue),
	})
	if err != nil {
		t.Fatalf("failed to put secret value: %v", err)
	}

	if putOutput.VersionId == nil {
		t.Fatal("version ID is nil")
	}

	// Get and verify updated value.
	getOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to get secret value: %v", err)
	}

	if *getOutput.SecretString != updatedValue {
		t.Errorf("secret value mismatch: got %s, want %s", *getOutput.SecretString, updatedValue)
	}
}

func TestSecretsManager_ListSecrets(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-list"

	// Create secret.
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String("test-value"),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// List secrets.
	listOutput, err := client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
	if err != nil {
		t.Fatalf("failed to list secrets: %v", err)
	}

	found := false

	for _, secret := range listOutput.SecretList {
		if *secret.Name == secretName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("secret %s not found in list", secretName)
	}
}

func TestSecretsManager_DescribeSecret(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-describe"
	description := "This is a test secret"

	// Create secret with description.
	createOutput, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String("test-value"),
		Description:  aws.String(description),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Describe secret.
	describeOutput, err := client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to describe secret: %v", err)
	}

	if *describeOutput.Name != secretName {
		t.Errorf("name mismatch: got %s, want %s", *describeOutput.Name, secretName)
	}

	if *describeOutput.ARN != *createOutput.ARN {
		t.Errorf("ARN mismatch: got %s, want %s", *describeOutput.ARN, *createOutput.ARN)
	}

	if describeOutput.Description == nil || *describeOutput.Description != description {
		t.Errorf("description mismatch: got %v, want %s", describeOutput.Description, description)
	}
}

func TestSecretsManager_UpdateSecret(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-update"
	initialDescription := "Initial description"
	updatedDescription := "Updated description"
	updatedValue := "updated-secret-value"

	// Create secret.
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String("initial-value"),
		Description:  aws.String(initialDescription),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Update secret.
	updateOutput, err := client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretName),
		Description:  aws.String(updatedDescription),
		SecretString: aws.String(updatedValue),
	})
	if err != nil {
		t.Fatalf("failed to update secret: %v", err)
	}

	if updateOutput.ARN == nil {
		t.Fatal("updated secret ARN is nil")
	}

	// Verify description updated.
	describeOutput, err := client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to describe secret: %v", err)
	}

	if describeOutput.Description == nil || *describeOutput.Description != updatedDescription {
		t.Errorf("description mismatch: got %v, want %s", describeOutput.Description, updatedDescription)
	}

	// Verify value updated.
	getOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to get secret value: %v", err)
	}

	if *getOutput.SecretString != updatedValue {
		t.Errorf("secret value mismatch: got %s, want %s", *getOutput.SecretString, updatedValue)
	}
}

func TestSecretsManager_DeleteWithRecoveryWindow(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-recovery"

	// Create secret.
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String("test-value"),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	// Delete with recovery window.
	deleteOutput, err := client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:             aws.String(secretName),
		RecoveryWindowInDays: aws.Int64(7),
	})
	if err != nil {
		t.Fatalf("failed to delete secret: %v", err)
	}

	if deleteOutput.DeletionDate == nil {
		t.Fatal("deletion date is nil")
	}

	t.Logf("Secret scheduled for deletion at: %v", *deleteOutput.DeletionDate)

	// Verify secret is not accessible.
	_, err = client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err == nil {
		t.Fatal("expected error when getting deleted secret")
	}

	// Force delete for cleanup.
	_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
}

func TestSecretsManager_SecretWithBinary(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-binary"
	secretBinary := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// Create secret with binary value.
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretBinary: secretBinary,
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Get secret value.
	getOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to get secret value: %v", err)
	}

	if string(getOutput.SecretBinary) != string(secretBinary) {
		t.Errorf("secret binary mismatch: got %v, want %v", getOutput.SecretBinary, secretBinary)
	}
}

func TestSecretsManager_SecretWithJSON(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-json"

	secretData := map[string]string{
		"username": "admin",
		"password": "supersecret123",
		"host":     "db.example.com",
	}

	secretJSON, err := json.Marshal(secretData)
	if err != nil {
		t.Fatalf("failed to marshal secret data: %v", err)
	}

	// Create secret with JSON value.
	_, err = client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(string(secretJSON)),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Get secret value.
	getOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		t.Fatalf("failed to get secret value: %v", err)
	}

	// Parse and verify JSON.
	var retrievedData map[string]string
	if err := json.Unmarshal([]byte(*getOutput.SecretString), &retrievedData); err != nil {
		t.Fatalf("failed to unmarshal secret JSON: %v", err)
	}

	if retrievedData["username"] != secretData["username"] {
		t.Errorf("username mismatch: got %s, want %s", retrievedData["username"], secretData["username"])
	}

	if retrievedData["password"] != secretData["password"] {
		t.Errorf("password mismatch: got %s, want %s", retrievedData["password"], secretData["password"])
	}
}

func TestSecretsManager_VersionStages(t *testing.T) {
	client := newSecretsManagerClient(t)
	ctx := t.Context()
	secretName := "test-secret-versions"

	// Create secret with initial value.
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String("version-1"),
	})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	// Put new version.
	_, err = client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String("version-2"),
	})
	if err != nil {
		t.Fatalf("failed to put secret value: %v", err)
	}

	// Get current version (should be version-2).
	currentOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}

	if *currentOutput.SecretString != "version-2" {
		t.Errorf("current version mismatch: got %s, want version-2", *currentOutput.SecretString)
	}

	// Get previous version (should be version-1).
	previousOutput, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSPREVIOUS"),
	})
	if err != nil {
		t.Fatalf("failed to get previous version: %v", err)
	}

	if *previousOutput.SecretString != "version-1" {
		t.Errorf("previous version mismatch: got %s, want version-1", *previousOutput.SecretString)
	}
}
