package cli

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

func newS3APICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "s3api",
		Short: "S3 API commands",
	}

	cmd.AddCommand(
		newS3APIPutObjectCmd(),
		newS3APIPutBucketNotificationConfigurationCmd(),
		newS3APIPutBucketCorsCmd(),
	)

	return cmd
}

func newS3APIPutObjectCmd() *cobra.Command {
	var bucket, key string

	cmd := &cobra.Command{
		Use:   "put-object",
		Short: "Put an object into S3",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := s3.NewFromConfig(cfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
				o.UsePathStyle = true
			})

			_, err = client.PutObject(cmd.Context(), &s3.PutObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
				Body:   strings.NewReader(""),
			})

			return err
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "Bucket name")
	cmd.Flags().StringVar(&key, "key", "", "Object key")

	return cmd
}

func newS3APIPutBucketNotificationConfigurationCmd() *cobra.Command {
	var bucket, notifConfig string

	cmd := &cobra.Command{
		Use:   "put-bucket-notification-configuration",
		Short: "Set bucket notification configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := s3.NewFromConfig(cfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
				o.UsePathStyle = true
			})

			var nc struct {
				EventBridgeConfiguration *struct{} `json:"EventBridgeConfiguration"`
			}

			_ = json.Unmarshal([]byte(notifConfig), &nc)

			input := &s3.PutBucketNotificationConfigurationInput{
				Bucket:                    aws.String(bucket),
				NotificationConfiguration: &s3Types.NotificationConfiguration{},
			}

			if nc.EventBridgeConfiguration != nil {
				input.NotificationConfiguration.EventBridgeConfiguration = &s3Types.EventBridgeConfiguration{}
			}

			_, err = client.PutBucketNotificationConfiguration(cmd.Context(), input)

			return err
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "Bucket name")
	cmd.Flags().StringVar(&notifConfig, "notification-configuration", "", "Notification configuration (JSON)")

	return cmd
}

func newS3APIPutBucketCorsCmd() *cobra.Command {
	var bucket, corsConfig string

	cmd := &cobra.Command{
		Use:   "put-bucket-cors",
		Short: "Set bucket CORS configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := s3.NewFromConfig(cfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
				o.UsePathStyle = true
			})

			var input struct {
				CORSRules []struct {
					AllowedHeaders []string `json:"AllowedHeaders"`
					AllowedMethods []string `json:"AllowedMethods"`
					AllowedOrigins []string `json:"AllowedOrigins"`
				} `json:"CORSRules"`
			}

			_ = json.Unmarshal([]byte(corsConfig), &input)

			var rules []s3Types.CORSRule
			for _, r := range input.CORSRules {
				rules = append(rules, s3Types.CORSRule{
					AllowedHeaders: r.AllowedHeaders,
					AllowedMethods: r.AllowedMethods,
					AllowedOrigins: r.AllowedOrigins,
				})
			}

			_, err = client.PutBucketCors(cmd.Context(), &s3.PutBucketCorsInput{
				Bucket:            aws.String(bucket),
				CORSConfiguration: &s3Types.CORSConfiguration{CORSRules: rules},
			})

			return err
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "Bucket name")
	cmd.Flags().StringVar(&corsConfig, "cors-configuration", "", "CORS configuration (JSON)")

	return cmd
}
