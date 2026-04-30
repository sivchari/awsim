package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	amplifyTypes "github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/spf13/cobra"
)

func newAmplifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "amplify",
		Short: "Amplify commands",
	}

	cmd.AddCommand(
		newAmplifyCreateAppCmd(),
		newAmplifyGetAppCmd(),
		newAmplifyListAppsCmd(),
		newAmplifyUpdateAppCmd(),
		newAmplifyDeleteAppCmd(),
		newAmplifyCreateBranchCmd(),
		newAmplifyGetBranchCmd(),
		newAmplifyListBranchesCmd(),
		newAmplifyDeleteBranchCmd(),
	)

	return cmd
}

func newAmplifyCreateAppCmd() *cobra.Command {
	var name, description, repository, platform string

	cmd := &cobra.Command{
		Use:   "create-app",
		Short: "Create an Amplify app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &amplify.CreateAppInput{
				Name: aws.String(name),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			if repository != "" {
				input.Repository = aws.String(repository)
			}

			if platform != "" {
				input.Platform = amplifyPlatform(platform)
			}

			out, err := client.CreateApp(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-app failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "App name")
	cmd.Flags().StringVar(&description, "description", "", "App description")
	cmd.Flags().StringVar(&repository, "repository", "", "Repository URL")
	cmd.Flags().StringVar(&platform, "platform", "", "Platform (WEB, WEB_DYNAMIC, WEB_COMPUTE)")

	return cmd
}

func newAmplifyGetAppCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "get-app",
		Short: "Get an Amplify app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetApp(cmd.Context(), &amplify.GetAppInput{
				AppId: aws.String(appID),
			})
			if err != nil {
				return fmt.Errorf("get-app failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")

	return cmd
}

func newAmplifyListAppsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-apps",
		Short: "List Amplify apps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListApps(cmd.Context(), &amplify.ListAppsInput{})
			if err != nil {
				return fmt.Errorf("list-apps failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}
}

func newAmplifyUpdateAppCmd() *cobra.Command {
	var appID, name, description, platform string

	cmd := &cobra.Command{
		Use:   "update-app",
		Short: "Update an Amplify app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &amplify.UpdateAppInput{
				AppId: aws.String(appID),
			}

			if name != "" {
				input.Name = aws.String(name)
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			if platform != "" {
				input.Platform = amplifyPlatform(platform)
			}

			out, err := client.UpdateApp(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("update-app failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")
	cmd.Flags().StringVar(&name, "name", "", "App name")
	cmd.Flags().StringVar(&description, "description", "", "App description")
	cmd.Flags().StringVar(&platform, "platform", "", "Platform")

	return cmd
}

func newAmplifyDeleteAppCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "delete-app",
		Short: "Delete an Amplify app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteApp(cmd.Context(), &amplify.DeleteAppInput{
				AppId: aws.String(appID),
			})
			if err != nil {
				return fmt.Errorf("delete-app failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")

	return cmd
}

func newAmplifyCreateBranchCmd() *cobra.Command {
	var appID, branchName, description, framework, stage string

	cmd := &cobra.Command{
		Use:   "create-branch",
		Short: "Create an Amplify branch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &amplify.CreateBranchInput{
				AppId:      aws.String(appID),
				BranchName: aws.String(branchName),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			if framework != "" {
				input.Framework = aws.String(framework)
			}

			if stage != "" {
				input.Stage = amplifyStage(stage)
			}

			out, err := client.CreateBranch(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-branch failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")
	cmd.Flags().StringVar(&branchName, "branch-name", "", "Branch name")
	cmd.Flags().StringVar(&description, "description", "", "Branch description")
	cmd.Flags().StringVar(&framework, "framework", "", "Framework")
	cmd.Flags().StringVar(&stage, "stage", "", "Stage (PRODUCTION, BETA, DEVELOPMENT, etc.)")

	return cmd
}

func newAmplifyGetBranchCmd() *cobra.Command {
	var appID, branchName string

	cmd := &cobra.Command{
		Use:   "get-branch",
		Short: "Get an Amplify branch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetBranch(cmd.Context(), &amplify.GetBranchInput{
				AppId:      aws.String(appID),
				BranchName: aws.String(branchName),
			})
			if err != nil {
				return fmt.Errorf("get-branch failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")
	cmd.Flags().StringVar(&branchName, "branch-name", "", "Branch name")

	return cmd
}

func newAmplifyListBranchesCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "list-branches",
		Short: "List Amplify branches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListBranches(cmd.Context(), &amplify.ListBranchesInput{
				AppId: aws.String(appID),
			})
			if err != nil {
				return fmt.Errorf("list-branches failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")

	return cmd
}

func newAmplifyDeleteBranchCmd() *cobra.Command {
	var appID, branchName string

	cmd := &cobra.Command{
		Use:   "delete-branch",
		Short: "Delete an Amplify branch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := amplify.NewFromConfig(cfg, func(o *amplify.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteBranch(cmd.Context(), &amplify.DeleteBranchInput{
				AppId:      aws.String(appID),
				BranchName: aws.String(branchName),
			})
			if err != nil {
				return fmt.Errorf("delete-branch failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "App ID")
	cmd.Flags().StringVar(&branchName, "branch-name", "", "Branch name")

	return cmd
}

func amplifyPlatform(s string) amplifyTypes.Platform {
	return amplifyTypes.Platform(s)
}

func amplifyStage(s string) amplifyTypes.Stage {
	return amplifyTypes.Stage(s)
}
