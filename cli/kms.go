package cli

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/spf13/cobra"
)

func newKMSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kms",
		Short: "KMS commands",
	}

	cmd.AddCommand(
		newKMSCreateKeyCmd(),
		newKMSCreateAliasCmd(),
	)

	return cmd
}

func newKMSCreateKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-key",
		Short: "Create a KMS key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := kms.NewFromConfig(cfg, func(o *kms.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.CreateKey(cmd.Context(), &kms.CreateKeyInput{})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}
}

func newKMSCreateAliasCmd() *cobra.Command {
	var targetKeyID, aliasName string

	cmd := &cobra.Command{
		Use:   "create-alias",
		Short: "Create a KMS alias",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := kms.NewFromConfig(cfg, func(o *kms.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.CreateAlias(cmd.Context(), &kms.CreateAliasInput{
				TargetKeyId: aws.String(targetKeyID),
				AliasName:   aws.String(aliasName),
			})

			return err
		},
	}

	cmd.Flags().StringVar(&targetKeyID, "target-key-id", "", "Target key ID")
	cmd.Flags().StringVar(&aliasName, "alias-name", "", "Alias name")

	return cmd
}
