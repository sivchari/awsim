package cli

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/cobra"
)

func newSQSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sqs",
		Short: "SQS commands",
	}

	cmd.AddCommand(newSQSCreateQueueCmd())

	return cmd
}

func newSQSCreateQueueCmd() *cobra.Command {
	var queueName, attrsJSON string

	cmd := &cobra.Command{
		Use:   "create-queue",
		Short: "Create an SQS queue",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &sqs.CreateQueueInput{
				QueueName: aws.String(queueName),
			}

			if attrsJSON != "" {
				var attrs map[string]string

				_ = json.Unmarshal([]byte(attrsJSON), &attrs)

				input.Attributes = attrs
			}

			out, err := client.CreateQueue(cmd.Context(), input)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&queueName, "queue-name", "", "Queue name")
	cmd.Flags().StringVar(&attrsJSON, "attributes", "", "Queue attributes (JSON)")
	cmd.Flags().String("region", "", "Region override (ignored)")

	return cmd
}
