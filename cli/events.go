package cli

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebTypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/spf13/cobra"
)

func newEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "EventBridge commands",
	}

	cmd.AddCommand(
		newEventsCreateEventBusCmd(),
		newEventsCreateConnectionCmd(),
		newEventsDescribeConnectionCmd(),
		newEventsCreateAPIDestinationCmd(),
		newEventsPutRuleCmd(),
		newEventsPutTargetsCmd(),
	)

	return cmd
}

func newEventsCreateEventBusCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create-event-bus",
		Short: "Create an EventBridge event bus",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.CreateEventBus(cmd.Context(), &eventbridge.CreateEventBusInput{
				Name: aws.String(name),
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Event bus name")

	return cmd
}

func newEventsCreateConnectionCmd() *cobra.Command {
	var name, authType, authParamsJSON string

	cmd := &cobra.Command{
		Use:   "create-connection",
		Short: "Create an EventBridge connection",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &eventbridge.CreateConnectionInput{
				Name:              aws.String(name),
				AuthorizationType: ebTypes.ConnectionAuthorizationType(authType),
			}

			if authParamsJSON != "" {
				var params struct {
					APIKeyAuthParameters *struct {
						ApiKeyName  string `json:"ApiKeyName"`
						ApiKeyValue string `json:"ApiKeyValue"`
					} `json:"ApiKeyAuthParameters"`
				}

				_ = json.Unmarshal([]byte(authParamsJSON), &params)

				if params.APIKeyAuthParameters != nil {
					input.AuthParameters = &ebTypes.CreateConnectionAuthRequestParameters{
						ApiKeyAuthParameters: &ebTypes.CreateConnectionApiKeyAuthRequestParameters{
							ApiKeyName:  aws.String(params.APIKeyAuthParameters.ApiKeyName),
							ApiKeyValue: aws.String(params.APIKeyAuthParameters.ApiKeyValue),
						},
					}
				}
			}

			out, err := client.CreateConnection(cmd.Context(), input)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Connection name")
	cmd.Flags().StringVar(&authType, "authorization-type", "", "Authorization type")
	cmd.Flags().StringVar(&authParamsJSON, "auth-parameters", "", "Auth parameters (JSON)")

	return cmd
}

func newEventsDescribeConnectionCmd() *cobra.Command {
	var name, query, output string

	cmd := &cobra.Command{
		Use:   "describe-connection",
		Short: "Describe an EventBridge connection",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.DescribeConnection(cmd.Context(), &eventbridge.DescribeConnectionInput{
				Name: aws.String(name),
			})
			if err != nil {
				return err
			}

			// Handle --query and --output for AWS CLI compatibility.
			if query == "ConnectionArn" && output == "text" {
				if out.ConnectionArn != nil {
					_, _ = os.Stdout.WriteString(*out.ConnectionArn + "\n")
				}

				return nil
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Connection name")
	cmd.Flags().StringVar(&query, "query", "", "JMESPath query (limited support)")
	cmd.Flags().StringVar(&output, "output", "", "Output format")

	return cmd
}

func newEventsCreateAPIDestinationCmd() *cobra.Command {
	var name, connArn, endpoint, method string
	var rateLimit int32

	cmd := &cobra.Command{
		Use:   "create-api-destination",
		Short: "Create an EventBridge API destination",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.CreateApiDestination(cmd.Context(), &eventbridge.CreateApiDestinationInput{
				Name:                         aws.String(name),
				ConnectionArn:                aws.String(connArn),
				InvocationEndpoint:           aws.String(endpoint),
				HttpMethod:                   ebTypes.ApiDestinationHttpMethod(method),
				InvocationRateLimitPerSecond: aws.Int32(rateLimit),
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "API destination name")
	cmd.Flags().StringVar(&connArn, "connection-arn", "", "Connection ARN")
	cmd.Flags().StringVar(&endpoint, "invocation-endpoint", "", "Invocation endpoint URL")
	cmd.Flags().StringVar(&method, "http-method", "POST", "HTTP method")
	cmd.Flags().Int32Var(&rateLimit, "invocation-rate-limit-per-second", 300, "Rate limit")

	return cmd
}

func newEventsPutRuleCmd() *cobra.Command {
	var name, eventBusName, eventPattern, state string

	cmd := &cobra.Command{
		Use:   "put-rule",
		Short: "Create or update an EventBridge rule",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &eventbridge.PutRuleInput{Name: aws.String(name)}

			if eventBusName != "" {
				input.EventBusName = aws.String(eventBusName)
			}

			if eventPattern != "" {
				input.EventPattern = aws.String(eventPattern)
			}

			if state != "" {
				input.State = ebTypes.RuleState(state)
			}

			out, err := client.PutRule(cmd.Context(), input)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&eventBusName, "event-bus-name", "", "Event bus name")
	cmd.Flags().StringVar(&eventPattern, "event-pattern", "", "Event pattern (JSON)")
	cmd.Flags().StringVar(&state, "state", "", "Rule state (ENABLED/DISABLED)")

	return cmd
}

func newEventsPutTargetsCmd() *cobra.Command {
	var rule, eventBusName, targetsJSON string

	cmd := &cobra.Command{
		Use:   "put-targets",
		Short: "Add targets to an EventBridge rule",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := eventbridge.NewFromConfig(cfg, func(o *eventbridge.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &eventbridge.PutTargetsInput{Rule: aws.String(rule)}

			if eventBusName != "" {
				input.EventBusName = aws.String(eventBusName)
			}

			if targetsJSON != "" {
				input.Targets = parseTargets(targetsJSON)
			}

			out, err := client.PutTargets(cmd.Context(), input)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}

	cmd.Flags().StringVar(&rule, "rule", "", "Rule name")
	cmd.Flags().StringVar(&eventBusName, "event-bus-name", "", "Event bus name")
	cmd.Flags().StringVar(&targetsJSON, "targets", "", "Targets (JSON array)")

	return cmd
}

func parseTargets(s string) []ebTypes.Target {
	var raw []struct {
		ID             string `json:"Id"`
		Arn            string `json:"Arn"`
		InputPath      string `json:"InputPath,omitempty"`
		HTTPParameters *struct {
			PathParameterValues []string `json:"PathParameterValues,omitempty"`
		} `json:"HttpParameters,omitempty"`
	}

	_ = json.Unmarshal([]byte(s), &raw)

	var targets []ebTypes.Target

	for _, t := range raw {
		target := ebTypes.Target{
			Id:  aws.String(t.ID),
			Arn: aws.String(t.Arn),
		}

		if t.InputPath != "" {
			target.InputPath = aws.String(t.InputPath)
		}

		if t.HTTPParameters != nil {
			target.HttpParameters = &ebTypes.HttpParameters{
				PathParameterValues: t.HTTPParameters.PathParameterValues,
			}
		}

		targets = append(targets, target)
	}

	return targets
}
