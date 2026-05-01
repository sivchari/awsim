package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	athenaTypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/spf13/cobra"
)

func newAthenaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "athena",
		Short: "Athena commands",
	}

	cmd.AddCommand(
		newAthenaCreateWorkGroupCmd(),
		newAthenaDeleteWorkGroupCmd(),
		newAthenaStartQueryExecutionCmd(),
		newAthenaStopQueryExecutionCmd(),
		newAthenaGetQueryExecutionCmd(),
		newAthenaGetQueryResultsCmd(),
		newAthenaListQueryExecutionsCmd(),
	)

	return cmd
}

func newAthenaCreateWorkGroupCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create-work-group",
		Short: "Create a workgroup",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &athena.CreateWorkGroupInput{
				Name: aws.String(name),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			out, err := client.CreateWorkGroup(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-work-group failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Work group name")
	cmd.Flags().StringVar(&description, "description", "", "Work group description")

	return cmd
}

func newAthenaDeleteWorkGroupCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "delete-work-group",
		Short: "Delete a workgroup",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteWorkGroup(cmd.Context(), &athena.DeleteWorkGroupInput{
				WorkGroup: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("delete-work-group failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "work-group", "", "Work group name")

	return cmd
}

func newAthenaStartQueryExecutionCmd() *cobra.Command {
	var queryString, database, workGroup, outputLocation string

	cmd := &cobra.Command{
		Use:   "start-query-execution",
		Short: "Start a query execution",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &athena.StartQueryExecutionInput{
				QueryString: aws.String(queryString),
			}

			if database != "" {
				input.QueryExecutionContext = &athenaTypes.QueryExecutionContext{
					Database: aws.String(database),
				}
			}

			if workGroup != "" {
				input.WorkGroup = aws.String(workGroup)
			}

			if outputLocation != "" {
				input.ResultConfiguration = &athenaTypes.ResultConfiguration{
					OutputLocation: aws.String(outputLocation),
				}
			}

			out, err := client.StartQueryExecution(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("start-query-execution failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&queryString, "query-string", "", "SQL query string")
	cmd.Flags().StringVar(&database, "database", "", "Database name")
	cmd.Flags().StringVar(&workGroup, "work-group", "", "Work group name")
	cmd.Flags().StringVar(&outputLocation, "output-location", "", "S3 output location")

	return cmd
}

func newAthenaStopQueryExecutionCmd() *cobra.Command {
	var queryExecutionID string

	cmd := &cobra.Command{
		Use:   "stop-query-execution",
		Short: "Stop a query execution",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.StopQueryExecution(cmd.Context(), &athena.StopQueryExecutionInput{
				QueryExecutionId: aws.String(queryExecutionID),
			})
			if err != nil {
				return fmt.Errorf("stop-query-execution failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&queryExecutionID, "query-execution-id", "", "Query execution ID")

	return cmd
}

func newAthenaGetQueryExecutionCmd() *cobra.Command {
	var queryExecutionID string

	cmd := &cobra.Command{
		Use:   "get-query-execution",
		Short: "Get a query execution",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetQueryExecution(cmd.Context(), &athena.GetQueryExecutionInput{
				QueryExecutionId: aws.String(queryExecutionID),
			})
			if err != nil {
				return fmt.Errorf("get-query-execution failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&queryExecutionID, "query-execution-id", "", "Query execution ID")

	return cmd
}

func newAthenaGetQueryResultsCmd() *cobra.Command {
	var queryExecutionID string

	cmd := &cobra.Command{
		Use:   "get-query-results",
		Short: "Get query results",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetQueryResults(cmd.Context(), &athena.GetQueryResultsInput{
				QueryExecutionId: aws.String(queryExecutionID),
			})
			if err != nil {
				return fmt.Errorf("get-query-results failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&queryExecutionID, "query-execution-id", "", "Query execution ID")

	return cmd
}

func newAthenaListQueryExecutionsCmd() *cobra.Command {
	var workGroup string

	cmd := &cobra.Command{
		Use:   "list-query-executions",
		Short: "List query executions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := athena.NewFromConfig(cfg, func(o *athena.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &athena.ListQueryExecutionsInput{}

			if workGroup != "" {
				input.WorkGroup = aws.String(workGroup)
			}

			out, err := client.ListQueryExecutions(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("list-query-executions failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&workGroup, "work-group", "", "Work group name")

	return cmd
}
