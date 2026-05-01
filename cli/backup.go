package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backupTypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "AWS Backup commands",
	}

	cmd.AddCommand(
		newBackupCreateBackupVaultCmd(),
		newBackupDescribeBackupVaultCmd(),
		newBackupListBackupVaultsCmd(),
		newBackupDeleteBackupVaultCmd(),
		newBackupCreateBackupPlanCmd(),
		newBackupGetBackupPlanCmd(),
		newBackupListBackupPlansCmd(),
		newBackupDeleteBackupPlanCmd(),
		newBackupCreateBackupSelectionCmd(),
		newBackupGetBackupSelectionCmd(),
		newBackupListBackupSelectionsCmd(),
		newBackupDeleteBackupSelectionCmd(),
	)

	return cmd
}

func newBackupCreateBackupVaultCmd() *cobra.Command {
	var vaultName string

	cmd := &cobra.Command{
		Use:   "create-backup-vault",
		Short: "Create a backup vault",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.CreateBackupVault(cmd.Context(), &backup.CreateBackupVaultInput{
				BackupVaultName: aws.String(vaultName),
			})
			if err != nil {
				return fmt.Errorf("create-backup-vault failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&vaultName, "backup-vault-name", "", "Backup vault name")

	return cmd
}

func newBackupDescribeBackupVaultCmd() *cobra.Command {
	var vaultName string

	cmd := &cobra.Command{
		Use:   "describe-backup-vault",
		Short: "Describe a backup vault",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.DescribeBackupVault(cmd.Context(), &backup.DescribeBackupVaultInput{
				BackupVaultName: aws.String(vaultName),
			})
			if err != nil {
				return fmt.Errorf("describe-backup-vault failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&vaultName, "backup-vault-name", "", "Backup vault name")

	return cmd
}

func newBackupListBackupVaultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-backup-vaults",
		Short: "List backup vaults",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListBackupVaults(cmd.Context(), &backup.ListBackupVaultsInput{})
			if err != nil {
				return fmt.Errorf("list-backup-vaults failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}
}

func newBackupDeleteBackupVaultCmd() *cobra.Command {
	var vaultName string

	cmd := &cobra.Command{
		Use:   "delete-backup-vault",
		Short: "Delete a backup vault",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteBackupVault(cmd.Context(), &backup.DeleteBackupVaultInput{
				BackupVaultName: aws.String(vaultName),
			})
			if err != nil {
				return fmt.Errorf("delete-backup-vault failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&vaultName, "backup-vault-name", "", "Backup vault name")

	return cmd
}

func newBackupCreateBackupPlanCmd() *cobra.Command {
	var planJSON string

	cmd := &cobra.Command{
		Use:   "create-backup-plan",
		Short: "Create a backup plan",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			var plan backupTypes.BackupPlanInput

			_ = json.Unmarshal([]byte(planJSON), &plan)

			out, err := client.CreateBackupPlan(cmd.Context(), &backup.CreateBackupPlanInput{
				BackupPlan: &plan,
			})
			if err != nil {
				return fmt.Errorf("create-backup-plan failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planJSON, "backup-plan", "", "Backup plan (JSON)")

	return cmd
}

func newBackupGetBackupPlanCmd() *cobra.Command {
	var planID string

	cmd := &cobra.Command{
		Use:   "get-backup-plan",
		Short: "Get a backup plan",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetBackupPlan(cmd.Context(), &backup.GetBackupPlanInput{
				BackupPlanId: aws.String(planID),
			})
			if err != nil {
				return fmt.Errorf("get-backup-plan failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "backup-plan-id", "", "Backup plan ID")

	return cmd
}

func newBackupListBackupPlansCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-backup-plans",
		Short: "List backup plans",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListBackupPlans(cmd.Context(), &backup.ListBackupPlansInput{})
			if err != nil {
				return fmt.Errorf("list-backup-plans failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}
}

func newBackupDeleteBackupPlanCmd() *cobra.Command {
	var planID string

	cmd := &cobra.Command{
		Use:   "delete-backup-plan",
		Short: "Delete a backup plan",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteBackupPlan(cmd.Context(), &backup.DeleteBackupPlanInput{
				BackupPlanId: aws.String(planID),
			})
			if err != nil {
				return fmt.Errorf("delete-backup-plan failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "backup-plan-id", "", "Backup plan ID")

	return cmd
}

func newBackupCreateBackupSelectionCmd() *cobra.Command {
	var planID, selectionJSON string

	cmd := &cobra.Command{
		Use:   "create-backup-selection",
		Short: "Create a backup selection",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			var selection backupTypes.BackupSelection

			_ = json.Unmarshal([]byte(selectionJSON), &selection)

			out, err := client.CreateBackupSelection(cmd.Context(), &backup.CreateBackupSelectionInput{
				BackupPlanId:    aws.String(planID),
				BackupSelection: &selection,
			})
			if err != nil {
				return fmt.Errorf("create-backup-selection failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "backup-plan-id", "", "Backup plan ID")
	cmd.Flags().StringVar(&selectionJSON, "backup-selection", "", "Backup selection (JSON)")

	return cmd
}

func newBackupGetBackupSelectionCmd() *cobra.Command {
	var planID, selectionID string

	cmd := &cobra.Command{
		Use:   "get-backup-selection",
		Short: "Get a backup selection",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetBackupSelection(cmd.Context(), &backup.GetBackupSelectionInput{
				BackupPlanId: aws.String(planID),
				SelectionId:  aws.String(selectionID),
			})
			if err != nil {
				return fmt.Errorf("get-backup-selection failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "backup-plan-id", "", "Backup plan ID")
	cmd.Flags().StringVar(&selectionID, "selection-id", "", "Selection ID")

	return cmd
}

func newBackupListBackupSelectionsCmd() *cobra.Command {
	var planID string

	cmd := &cobra.Command{
		Use:   "list-backup-selections",
		Short: "List backup selections",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListBackupSelections(cmd.Context(), &backup.ListBackupSelectionsInput{
				BackupPlanId: aws.String(planID),
			})
			if err != nil {
				return fmt.Errorf("list-backup-selections failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "backup-plan-id", "", "Backup plan ID")

	return cmd
}

func newBackupDeleteBackupSelectionCmd() *cobra.Command {
	var planID, selectionID string

	cmd := &cobra.Command{
		Use:   "delete-backup-selection",
		Short: "Delete a backup selection",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := backup.NewFromConfig(cfg, func(o *backup.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteBackupSelection(cmd.Context(), &backup.DeleteBackupSelectionInput{
				BackupPlanId: aws.String(planID),
				SelectionId:  aws.String(selectionID),
			})
			if err != nil {
				return fmt.Errorf("delete-backup-selection failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "backup-plan-id", "", "Backup plan ID")
	cmd.Flags().StringVar(&selectionID, "selection-id", "", "Selection ID")

	return cmd
}
