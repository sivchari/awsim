//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
)

func newGlacierClient(t *testing.T) *glacier.Client {
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

	return glacier.NewFromConfig(cfg, func(o *glacier.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestGlacier_CreateAndDeleteVault(t *testing.T) {
	client := newGlacierClient(t)
	ctx := t.Context()
	vaultName := "test-vault"

	_, err := client.CreateVault(ctx, &glacier.CreateVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	// Delete
	_, err = client.DeleteVault(ctx, &glacier.DeleteVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("failed to delete vault: %v", err)
	}
}

func TestGlacier_DescribeVault(t *testing.T) {
	client := newGlacierClient(t)
	ctx := t.Context()
	vaultName := "test-describe-vault"

	_, err := client.CreateVault(ctx, &glacier.CreateVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteVault(t.Context(), &glacier.DeleteVaultInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(vaultName),
		})
	})

	describeResult, err := client.DescribeVault(ctx, &glacier.DescribeVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("failed to describe vault: %v", err)
	}

	if *describeResult.VaultName != vaultName {
		t.Errorf("expected vault name %s, got %s", vaultName, *describeResult.VaultName)
	}

	if describeResult.VaultARN == nil || *describeResult.VaultARN == "" {
		t.Error("expected vault ARN to be set")
	}

	if describeResult.NumberOfArchives != 0 {
		t.Errorf("expected 0 archives, got %d", describeResult.NumberOfArchives)
	}
}

func TestGlacier_ListVaults(t *testing.T) {
	client := newGlacierClient(t)
	ctx := t.Context()
	vaultName := "test-list-vault"

	_, err := client.CreateVault(ctx, &glacier.CreateVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteVault(t.Context(), &glacier.DeleteVaultInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(vaultName),
		})
	})

	listResult, err := client.ListVaults(ctx, &glacier.ListVaultsInput{
		AccountId: aws.String("-"),
	})
	if err != nil {
		t.Fatalf("failed to list vaults: %v", err)
	}

	found := false

	for _, v := range listResult.VaultList {
		if *v.VaultName == vaultName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected to find vault %s in list", vaultName)
	}
}

func TestGlacier_VaultNotFound(t *testing.T) {
	client := newGlacierClient(t)
	ctx := t.Context()

	_, err := client.DescribeVault(ctx, &glacier.DescribeVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String("non-existent-vault"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent vault")
	}
}

func TestGlacier_CreateVaultIdempotent(t *testing.T) {
	client := newGlacierClient(t)
	ctx := t.Context()
	vaultName := "test-idempotent-vault"

	_, err := client.CreateVault(ctx, &glacier.CreateVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteVault(t.Context(), &glacier.DeleteVaultInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(vaultName),
		})
	})

	// Create same vault again - should succeed (idempotent)
	_, err = client.CreateVault(ctx, &glacier.CreateVaultInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
	})
	if err != nil {
		t.Fatalf("expected idempotent create to succeed, got error: %v", err)
	}
}
