package cli

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

func newS3Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "S3 commands",
	}

	cmd.AddCommand(newS3MBCmd())

	return cmd
}

func newS3MBCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mb s3://bucket-name",
		Short: "Create an S3 bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bucket := strings.TrimPrefix(args[0], "s3://")

			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := s3.NewFromConfig(cfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
				o.UsePathStyle = true
			})

			_, err = client.CreateBucket(cmd.Context(), &s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				return fmt.Errorf("make_bucket failed: s3://%s %w", bucket, err)
			}

			fmt.Printf("make_bucket: %s\n", bucket)

			return nil
		},
	}
}
