//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

func newKMSClient(t *testing.T) *kms.Client {
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

	return kms.NewFromConfig(cfg, func(o *kms.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestKMS_CreateAndDescribeKey(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test key"),
		KeyUsage:    types.KeyUsageTypeEncryptDecrypt,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	if createOutput.KeyMetadata == nil || createOutput.KeyMetadata.KeyId == nil {
		t.Fatal("key metadata is nil")
	}

	keyID := *createOutput.KeyMetadata.KeyId
	t.Logf("Created key: %s", keyID)

	// Describe key.
	describeOutput, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to describe key: %v", err)
	}

	if *describeOutput.KeyMetadata.KeyId != keyID {
		t.Errorf("key ID mismatch: got %s, want %s", *describeOutput.KeyMetadata.KeyId, keyID)
	}

	if *describeOutput.KeyMetadata.Description != "Test key" {
		t.Errorf("description mismatch: got %s, want Test key", *describeOutput.KeyMetadata.Description)
	}

	t.Logf("Described key: %s", keyID)
}

func TestKMS_ListKeys(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create a key first.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test list key"),
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId

	// List keys.
	listOutput, err := client.ListKeys(ctx, &kms.ListKeysInput{
		Limit: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}

	if len(listOutput.Keys) == 0 {
		t.Fatal("no keys returned")
	}

	// Find our key.
	found := false
	for _, key := range listOutput.Keys {
		if *key.KeyId == keyID {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("created key %s not found in list", keyID)
	}

	t.Logf("Listed %d keys", len(listOutput.Keys))
}

func TestKMS_EnableDisableKey(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test enable/disable key"),
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId

	// Disable key.
	_, err = client.DisableKey(ctx, &kms.DisableKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to disable key: %v", err)
	}

	// Verify disabled.
	describeOutput, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to describe key: %v", err)
	}

	if describeOutput.KeyMetadata.KeyState != types.KeyStateDisabled {
		t.Errorf("key state should be Disabled, got %s", describeOutput.KeyMetadata.KeyState)
	}

	// Enable key.
	_, err = client.EnableKey(ctx, &kms.EnableKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to enable key: %v", err)
	}

	// Verify enabled.
	describeOutput, err = client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to describe key: %v", err)
	}

	if describeOutput.KeyMetadata.KeyState != types.KeyStateEnabled {
		t.Errorf("key state should be Enabled, got %s", describeOutput.KeyMetadata.KeyState)
	}

	t.Logf("Enable/disable key test passed: %s", keyID)
}

func TestKMS_ScheduleKeyDeletion(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test deletion key"),
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId

	// Schedule deletion.
	deleteOutput, err := client.ScheduleKeyDeletion(ctx, &kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(keyID),
		PendingWindowInDays: aws.Int32(7),
	})
	if err != nil {
		t.Fatalf("failed to schedule key deletion: %v", err)
	}

	if *deleteOutput.KeyId != keyID {
		t.Errorf("key ID mismatch: got %s, want %s", *deleteOutput.KeyId, keyID)
	}

	// Verify pending deletion.
	describeOutput, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to describe key: %v", err)
	}

	if describeOutput.KeyMetadata.KeyState != types.KeyStatePendingDeletion {
		t.Errorf("key state should be PendingDeletion, got %s", describeOutput.KeyMetadata.KeyState)
	}

	t.Logf("Scheduled key deletion: %s", keyID)
}

func TestKMS_EncryptDecrypt(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test encrypt/decrypt key"),
		KeyUsage:    types.KeyUsageTypeEncryptDecrypt,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId
	plaintext := []byte("Hello, KMS!")

	// Encrypt.
	encryptOutput, err := client.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     aws.String(keyID),
		Plaintext: plaintext,
	})
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if len(encryptOutput.CiphertextBlob) == 0 {
		t.Fatal("ciphertext is empty")
	}

	// Decrypt.
	decryptOutput, err := client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: encryptOutput.CiphertextBlob,
	})
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if string(decryptOutput.Plaintext) != string(plaintext) {
		t.Errorf("plaintext mismatch: got %s, want %s", decryptOutput.Plaintext, plaintext)
	}

	t.Logf("Encrypt/decrypt test passed")
}

func TestKMS_GenerateDataKey(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test generate data key"),
		KeyUsage:    types.KeyUsageTypeEncryptDecrypt,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId

	// Generate data key.
	dataKeyOutput, err := client.GenerateDataKey(ctx, &kms.GenerateDataKeyInput{
		KeyId:   aws.String(keyID),
		KeySpec: types.DataKeySpecAes256,
	})
	if err != nil {
		t.Fatalf("failed to generate data key: %v", err)
	}

	if len(dataKeyOutput.Plaintext) == 0 {
		t.Fatal("plaintext data key is empty")
	}

	if len(dataKeyOutput.CiphertextBlob) == 0 {
		t.Fatal("ciphertext data key is empty")
	}

	if len(dataKeyOutput.Plaintext) != 32 {
		t.Errorf("plaintext data key should be 32 bytes, got %d", len(dataKeyOutput.Plaintext))
	}

	// Verify we can decrypt the ciphertext blob.
	decryptOutput, err := client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: dataKeyOutput.CiphertextBlob,
	})
	if err != nil {
		t.Fatalf("failed to decrypt data key: %v", err)
	}

	if string(decryptOutput.Plaintext) != string(dataKeyOutput.Plaintext) {
		t.Error("decrypted data key does not match original plaintext")
	}

	t.Logf("Generated data key of length %d bytes", len(dataKeyOutput.Plaintext))
}

func TestKMS_CreateAndDeleteAlias(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test alias key"),
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId
	aliasName := "alias/test-alias"

	// Create alias.
	_, err = client.CreateAlias(ctx, &kms.CreateAliasInput{
		AliasName:   aws.String(aliasName),
		TargetKeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	t.Logf("Created alias: %s", aliasName)

	// List aliases.
	listOutput, err := client.ListAliases(ctx, &kms.ListAliasesInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to list aliases: %v", err)
	}

	found := false
	for _, alias := range listOutput.Aliases {
		if *alias.AliasName == aliasName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("alias %s not found in list", aliasName)
	}

	// Delete alias.
	_, err = client.DeleteAlias(ctx, &kms.DeleteAliasInput{
		AliasName: aws.String(aliasName),
	})
	if err != nil {
		t.Fatalf("failed to delete alias: %v", err)
	}

	t.Logf("Deleted alias: %s", aliasName)
}

func TestKMS_KeyNotFound(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Try to describe a non-existent key.
	_, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String("non-existent-key-id"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
}

func TestKMS_EncryptWithAlias(t *testing.T) {
	client := newKMSClient(t)
	ctx := t.Context()

	// Create key.
	createOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("Test encrypt with alias"),
		KeyUsage:    types.KeyUsageTypeEncryptDecrypt,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}

	keyID := *createOutput.KeyMetadata.KeyId
	aliasName := "alias/test-encrypt-alias"

	// Create alias.
	_, err = client.CreateAlias(ctx, &kms.CreateAliasInput{
		AliasName:   aws.String(aliasName),
		TargetKeyId: aws.String(keyID),
	})
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteAlias(ctx, &kms.DeleteAliasInput{
			AliasName: aws.String(aliasName),
		})
	})

	plaintext := []byte("Hello via alias!")

	// Encrypt using alias.
	encryptOutput, err := client.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     aws.String(aliasName),
		Plaintext: plaintext,
	})
	if err != nil {
		t.Fatalf("failed to encrypt with alias: %v", err)
	}

	// Decrypt.
	decryptOutput, err := client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: encryptOutput.CiphertextBlob,
	})
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if string(decryptOutput.Plaintext) != string(plaintext) {
		t.Errorf("plaintext mismatch: got %s, want %s", decryptOutput.Plaintext, plaintext)
	}

	t.Logf("Encrypt with alias test passed")
}
