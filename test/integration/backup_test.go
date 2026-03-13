//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
)

func newBackupClient(t *testing.T) *backup.Client {
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

	return backup.NewFromConfig(cfg, func(o *backup.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestBackup_CreateBackupVault(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	result, err := client.CreateBackupVault(ctx, &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String("test-vault"),
	})
	if err != nil {
		t.Fatalf("failed to create backup vault: %v", err)
	}

	if result.BackupVaultName == nil || *result.BackupVaultName != "test-vault" {
		t.Errorf("expected vault name 'test-vault', got %v", result.BackupVaultName)
	}

	if result.BackupVaultArn == nil || *result.BackupVaultArn == "" {
		t.Error("expected BackupVaultArn to be set")
	}

	if result.CreationDate == nil {
		t.Error("expected CreationDate to be set")
	}
}

func TestBackup_DescribeBackupVault(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	_, err := client.CreateBackupVault(ctx, &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String("describe-vault"),
	})
	if err != nil {
		t.Fatalf("failed to create backup vault: %v", err)
	}

	result, err := client.DescribeBackupVault(ctx, &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String("describe-vault"),
	})
	if err != nil {
		t.Fatalf("failed to describe backup vault: %v", err)
	}

	if *result.BackupVaultName != "describe-vault" {
		t.Errorf("expected vault name 'describe-vault', got %s", *result.BackupVaultName)
	}
}

func TestBackup_ListBackupVaults(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	_, err := client.CreateBackupVault(ctx, &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String("list-vault"),
	})
	if err != nil {
		t.Fatalf("failed to create backup vault: %v", err)
	}

	result, err := client.ListBackupVaults(ctx, &backup.ListBackupVaultsInput{})
	if err != nil {
		t.Fatalf("failed to list backup vaults: %v", err)
	}

	if len(result.BackupVaultList) == 0 {
		t.Error("expected at least one backup vault")
	}
}

func TestBackup_DeleteBackupVault(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	_, err := client.CreateBackupVault(ctx, &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String("delete-vault"),
	})
	if err != nil {
		t.Fatalf("failed to create backup vault: %v", err)
	}

	_, err = client.DeleteBackupVault(ctx, &backup.DeleteBackupVaultInput{
		BackupVaultName: aws.String("delete-vault"),
	})
	if err != nil {
		t.Fatalf("failed to delete backup vault: %v", err)
	}

	_, err = client.DescribeBackupVault(ctx, &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String("delete-vault"),
	})
	if err == nil {
		t.Fatal("expected error for deleted vault")
	}
}

func TestBackup_CreateBackupPlan(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	result, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("test-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("daily-rule"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	if result.BackupPlanId == nil || *result.BackupPlanId == "" {
		t.Error("expected BackupPlanId to be set")
	}

	if result.BackupPlanArn == nil || *result.BackupPlanArn == "" {
		t.Error("expected BackupPlanArn to be set")
	}

	if result.CreationDate == nil {
		t.Error("expected CreationDate to be set")
	}
}

func TestBackup_GetBackupPlan(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	createResult, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("get-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	result, err := client.GetBackupPlan(ctx, &backup.GetBackupPlanInput{
		BackupPlanId: createResult.BackupPlanId,
	})
	if err != nil {
		t.Fatalf("failed to get backup plan: %v", err)
	}

	if result.BackupPlan == nil {
		t.Fatal("expected BackupPlan to be set")
	}

	if *result.BackupPlan.BackupPlanName != "get-plan" {
		t.Errorf("expected plan name 'get-plan', got %s", *result.BackupPlan.BackupPlanName)
	}

	if len(result.BackupPlan.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result.BackupPlan.Rules))
	}

	if *result.BackupPlan.Rules[0].RuleName != "rule-1" {
		t.Errorf("expected rule name 'rule-1', got %s", *result.BackupPlan.Rules[0].RuleName)
	}
}

func TestBackup_ListBackupPlans(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	_, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("list-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	result, err := client.ListBackupPlans(ctx, &backup.ListBackupPlansInput{})
	if err != nil {
		t.Fatalf("failed to list backup plans: %v", err)
	}

	if len(result.BackupPlansList) == 0 {
		t.Error("expected at least one backup plan")
	}
}

func TestBackup_DeleteBackupPlan(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	createResult, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("delete-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	_, err = client.DeleteBackupPlan(ctx, &backup.DeleteBackupPlanInput{
		BackupPlanId: createResult.BackupPlanId,
	})
	if err != nil {
		t.Fatalf("failed to delete backup plan: %v", err)
	}

	_, err = client.GetBackupPlan(ctx, &backup.GetBackupPlanInput{
		BackupPlanId: createResult.BackupPlanId,
	})
	if err == nil {
		t.Fatal("expected error for deleted plan")
	}
}

func TestBackup_CreateBackupSelection(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	planResult, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("selection-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	result, err := client.CreateBackupSelection(ctx, &backup.CreateBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		BackupSelection: &types.BackupSelection{
			SelectionName: aws.String("test-selection"),
			IamRoleArn:    aws.String("arn:aws:iam::000000000000:role/test-role"),
			Resources:     []string{"arn:aws:ec2:us-east-1:000000000000:volume/*"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup selection: %v", err)
	}

	if result.SelectionId == nil || *result.SelectionId == "" {
		t.Error("expected SelectionId to be set")
	}

	if result.BackupPlanId == nil || *result.BackupPlanId == "" {
		t.Error("expected BackupPlanId to be set")
	}
}

func TestBackup_GetBackupSelection(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	planResult, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("get-selection-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	selResult, err := client.CreateBackupSelection(ctx, &backup.CreateBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		BackupSelection: &types.BackupSelection{
			SelectionName: aws.String("get-selection"),
			IamRoleArn:    aws.String("arn:aws:iam::000000000000:role/test-role"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup selection: %v", err)
	}

	result, err := client.GetBackupSelection(ctx, &backup.GetBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		SelectionId:  selResult.SelectionId,
	})
	if err != nil {
		t.Fatalf("failed to get backup selection: %v", err)
	}

	if result.BackupSelection == nil {
		t.Fatal("expected BackupSelection to be set")
	}

	if *result.BackupSelection.SelectionName != "get-selection" {
		t.Errorf("expected selection name 'get-selection', got %s", *result.BackupSelection.SelectionName)
	}
}

func TestBackup_ListBackupSelections(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	planResult, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("list-selection-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	_, err = client.CreateBackupSelection(ctx, &backup.CreateBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		BackupSelection: &types.BackupSelection{
			SelectionName: aws.String("list-selection"),
			IamRoleArn:    aws.String("arn:aws:iam::000000000000:role/test-role"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup selection: %v", err)
	}

	result, err := client.ListBackupSelections(ctx, &backup.ListBackupSelectionsInput{
		BackupPlanId: planResult.BackupPlanId,
	})
	if err != nil {
		t.Fatalf("failed to list backup selections: %v", err)
	}

	if len(result.BackupSelectionsList) != 1 {
		t.Errorf("expected 1 selection, got %d", len(result.BackupSelectionsList))
	}
}

func TestBackup_DeleteBackupSelection(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	planResult, err := client.CreateBackupPlan(ctx, &backup.CreateBackupPlanInput{
		BackupPlan: &types.BackupPlanInput{
			BackupPlanName: aws.String("delete-selection-plan"),
			Rules: []types.BackupRuleInput{
				{
					RuleName:              aws.String("rule-1"),
					TargetBackupVaultName: aws.String("test-vault"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup plan: %v", err)
	}

	selResult, err := client.CreateBackupSelection(ctx, &backup.CreateBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		BackupSelection: &types.BackupSelection{
			SelectionName: aws.String("delete-selection"),
			IamRoleArn:    aws.String("arn:aws:iam::000000000000:role/test-role"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create backup selection: %v", err)
	}

	_, err = client.DeleteBackupSelection(ctx, &backup.DeleteBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		SelectionId:  selResult.SelectionId,
	})
	if err != nil {
		t.Fatalf("failed to delete backup selection: %v", err)
	}

	_, err = client.GetBackupSelection(ctx, &backup.GetBackupSelectionInput{
		BackupPlanId: planResult.BackupPlanId,
		SelectionId:  selResult.SelectionId,
	})
	if err == nil {
		t.Fatal("expected error for deleted selection")
	}
}

func TestBackup_VaultNotFound(t *testing.T) {
	client := newBackupClient(t)
	ctx := t.Context()

	_, err := client.DescribeBackupVault(ctx, &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String("nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent vault")
	}
}
