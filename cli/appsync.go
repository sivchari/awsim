package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	appsyncTypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/spf13/cobra"
)

func newAppSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appsync",
		Short: "AppSync commands",
	}

	cmd.AddCommand(
		newAppSyncCreateGraphqlAPICmd(),
		newAppSyncGetGraphqlAPICmd(),
		newAppSyncListGraphqlAPIsCmd(),
		newAppSyncDeleteGraphqlAPICmd(),
		newAppSyncCreateDataSourceCmd(),
		newAppSyncCreateResolverCmd(),
		newAppSyncStartSchemaCreationCmd(),
	)

	return cmd
}

func newAppSyncCreateGraphqlAPICmd() *cobra.Command {
	var name, authType string

	cmd := &cobra.Command{
		Use:   "create-graphql-api",
		Short: "Create a GraphQL API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &appsync.CreateGraphqlApiInput{
				Name:               aws.String(name),
				AuthenticationType: appsyncTypes.AuthenticationType(authType),
			}

			out, err := client.CreateGraphqlApi(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-graphql-api failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "API name")
	cmd.Flags().StringVar(&authType, "authentication-type", "", "Authentication type (API_KEY, AWS_IAM, etc.)")

	return cmd
}

func newAppSyncGetGraphqlAPICmd() *cobra.Command {
	var apiID string

	cmd := &cobra.Command{
		Use:   "get-graphql-api",
		Short: "Get a GraphQL API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetGraphqlApi(cmd.Context(), &appsync.GetGraphqlApiInput{
				ApiId: aws.String(apiID),
			})
			if err != nil {
				return fmt.Errorf("get-graphql-api failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&apiID, "api-id", "", "API ID")

	return cmd
}

func newAppSyncListGraphqlAPIsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-graphql-apis",
		Short: "List GraphQL APIs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListGraphqlApis(cmd.Context(), &appsync.ListGraphqlApisInput{})
			if err != nil {
				return fmt.Errorf("list-graphql-apis failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}
}

func newAppSyncDeleteGraphqlAPICmd() *cobra.Command {
	var apiID string

	cmd := &cobra.Command{
		Use:   "delete-graphql-api",
		Short: "Delete a GraphQL API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteGraphqlApi(cmd.Context(), &appsync.DeleteGraphqlApiInput{
				ApiId: aws.String(apiID),
			})
			if err != nil {
				return fmt.Errorf("delete-graphql-api failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&apiID, "api-id", "", "API ID")

	return cmd
}

func newAppSyncCreateDataSourceCmd() *cobra.Command {
	var apiID, name, dsType, description, serviceRoleArn string

	cmd := &cobra.Command{
		Use:   "create-data-source",
		Short: "Create a data source",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &appsync.CreateDataSourceInput{
				ApiId: aws.String(apiID),
				Name:  aws.String(name),
				Type:  appsyncTypes.DataSourceType(dsType),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			if serviceRoleArn != "" {
				input.ServiceRoleArn = aws.String(serviceRoleArn)
			}

			out, err := client.CreateDataSource(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-data-source failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&apiID, "api-id", "", "API ID")
	cmd.Flags().StringVar(&name, "name", "", "Data source name")
	cmd.Flags().StringVar(&dsType, "type", "", "Data source type (NONE, AMAZON_DYNAMODB, AWS_LAMBDA, etc.)")
	cmd.Flags().StringVar(&description, "description", "", "Data source description")
	cmd.Flags().StringVar(&serviceRoleArn, "service-role-arn", "", "Service role ARN")

	return cmd
}

func newAppSyncCreateResolverCmd() *cobra.Command {
	var apiID, typeName, fieldName, dataSourceName, requestTemplate, responseTemplate string

	cmd := &cobra.Command{
		Use:   "create-resolver",
		Short: "Create a resolver",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &appsync.CreateResolverInput{
				ApiId:     aws.String(apiID),
				TypeName:  aws.String(typeName),
				FieldName: aws.String(fieldName),
			}

			if dataSourceName != "" {
				input.DataSourceName = aws.String(dataSourceName)
			}

			if requestTemplate != "" {
				input.RequestMappingTemplate = aws.String(requestTemplate)
			}

			if responseTemplate != "" {
				input.ResponseMappingTemplate = aws.String(responseTemplate)
			}

			out, err := client.CreateResolver(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-resolver failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&apiID, "api-id", "", "API ID")
	cmd.Flags().StringVar(&typeName, "type-name", "", "Type name")
	cmd.Flags().StringVar(&fieldName, "field-name", "", "Field name")
	cmd.Flags().StringVar(&dataSourceName, "data-source-name", "", "Data source name")
	cmd.Flags().StringVar(&requestTemplate, "request-mapping-template", "", "Request mapping template")
	cmd.Flags().StringVar(&responseTemplate, "response-mapping-template", "", "Response mapping template")

	return cmd
}

func newAppSyncStartSchemaCreationCmd() *cobra.Command {
	var apiID, definition string

	cmd := &cobra.Command{
		Use:   "start-schema-creation",
		Short: "Start schema creation",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := appsync.NewFromConfig(cfg, func(o *appsync.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.StartSchemaCreation(cmd.Context(), &appsync.StartSchemaCreationInput{
				ApiId:      aws.String(apiID),
				Definition: []byte(definition),
			})
			if err != nil {
				return fmt.Errorf("start-schema-creation failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&apiID, "api-id", "", "API ID")
	cmd.Flags().StringVar(&definition, "definition", "", "Schema definition (SDL)")

	return cmd
}
