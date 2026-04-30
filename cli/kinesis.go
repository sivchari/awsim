package cli

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/spf13/cobra"
)

func newKinesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kinesis",
		Short: "Kinesis commands",
	}

	cmd.AddCommand(newKinesisCreateStreamCmd())

	return cmd
}

func newKinesisCreateStreamCmd() *cobra.Command {
	var streamName string

	var shardCount int32

	cmd := &cobra.Command{
		Use:   "create-stream",
		Short: "Create a Kinesis stream",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := kinesis.NewFromConfig(cfg, func(o *kinesis.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.CreateStream(cmd.Context(), &kinesis.CreateStreamInput{
				StreamName: aws.String(streamName),
				ShardCount: aws.Int32(shardCount),
			})
			if err != nil {
				return fmt.Errorf("create-stream failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&streamName, "stream-name", "", "Stream name")
	cmd.Flags().Int32Var(&shardCount, "shard-count", 1, "Number of shards")
	cmd.Flags().String("region", "", "Region override (ignored)")

	return cmd
}
