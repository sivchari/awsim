package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	apigatewayTypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/spf13/cobra"
)

func newAPIGatewayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apigateway",
		Short: "API Gateway commands",
	}

	cmd.AddCommand(
		newAPIGatewayCreateRestAPICmd(),
		newAPIGatewayGetRestAPIsCmd(),
		newAPIGatewayGetRestAPICmd(),
		newAPIGatewayDeleteRestAPICmd(),
		newAPIGatewayCreateResourceCmd(),
		newAPIGatewayGetResourcesCmd(),
		newAPIGatewayGetResourceCmd(),
		newAPIGatewayDeleteResourceCmd(),
		newAPIGatewayPutMethodCmd(),
		newAPIGatewayGetMethodCmd(),
		newAPIGatewayPutIntegrationCmd(),
		newAPIGatewayGetIntegrationCmd(),
		newAPIGatewayCreateDeploymentCmd(),
		newAPIGatewayGetDeploymentsCmd(),
		newAPIGatewayGetDeploymentCmd(),
		newAPIGatewayDeleteDeploymentCmd(),
		newAPIGatewayCreateStageCmd(),
		newAPIGatewayGetStagesCmd(),
		newAPIGatewayGetStageCmd(),
		newAPIGatewayDeleteStageCmd(),
	)

	return cmd
}

func newAPIGatewayCreateRestAPICmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create-rest-api",
		Short: "Create a REST API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &apigateway.CreateRestApiInput{
				Name: aws.String(name),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			out, err := client.CreateRestApi(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-rest-api failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "REST API name")
	cmd.Flags().StringVar(&description, "description", "", "REST API description")

	return cmd
}

func newAPIGatewayGetRestAPIsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-rest-apis",
		Short: "List REST APIs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetRestApis(cmd.Context(), &apigateway.GetRestApisInput{})
			if err != nil {
				return fmt.Errorf("get-rest-apis failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}
}

func newAPIGatewayGetRestAPICmd() *cobra.Command {
	var restAPIID string

	cmd := &cobra.Command{
		Use:   "get-rest-api",
		Short: "Get a REST API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetRestApi(cmd.Context(), &apigateway.GetRestApiInput{
				RestApiId: aws.String(restAPIID),
			})
			if err != nil {
				return fmt.Errorf("get-rest-api failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")

	return cmd
}

func newAPIGatewayDeleteRestAPICmd() *cobra.Command {
	var restAPIID string

	cmd := &cobra.Command{
		Use:   "delete-rest-api",
		Short: "Delete a REST API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteRestApi(cmd.Context(), &apigateway.DeleteRestApiInput{
				RestApiId: aws.String(restAPIID),
			})
			if err != nil {
				return fmt.Errorf("delete-rest-api failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")

	return cmd
}

func newAPIGatewayCreateResourceCmd() *cobra.Command {
	var restAPIID, parentID, pathPart string

	cmd := &cobra.Command{
		Use:   "create-resource",
		Short: "Create a resource",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.CreateResource(cmd.Context(), &apigateway.CreateResourceInput{
				RestApiId: aws.String(restAPIID),
				ParentId:  aws.String(parentID),
				PathPart:  aws.String(pathPart),
			})
			if err != nil {
				return fmt.Errorf("create-resource failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&parentID, "parent-id", "", "Parent resource ID")
	cmd.Flags().StringVar(&pathPart, "path-part", "", "Path part")

	return cmd
}

func newAPIGatewayGetResourcesCmd() *cobra.Command {
	var restAPIID string

	cmd := &cobra.Command{
		Use:   "get-resources",
		Short: "List resources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetResources(cmd.Context(), &apigateway.GetResourcesInput{
				RestApiId: aws.String(restAPIID),
			})
			if err != nil {
				return fmt.Errorf("get-resources failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")

	return cmd
}

func newAPIGatewayGetResourceCmd() *cobra.Command {
	var restAPIID, resourceID string

	cmd := &cobra.Command{
		Use:   "get-resource",
		Short: "Get a resource",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetResource(cmd.Context(), &apigateway.GetResourceInput{
				RestApiId:  aws.String(restAPIID),
				ResourceId: aws.String(resourceID),
			})
			if err != nil {
				return fmt.Errorf("get-resource failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")

	return cmd
}

func newAPIGatewayDeleteResourceCmd() *cobra.Command {
	var restAPIID, resourceID string

	cmd := &cobra.Command{
		Use:   "delete-resource",
		Short: "Delete a resource",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteResource(cmd.Context(), &apigateway.DeleteResourceInput{
				RestApiId:  aws.String(restAPIID),
				ResourceId: aws.String(resourceID),
			})
			if err != nil {
				return fmt.Errorf("delete-resource failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")

	return cmd
}

func newAPIGatewayPutMethodCmd() *cobra.Command {
	var restAPIID, resourceID, httpMethod, authType string

	cmd := &cobra.Command{
		Use:   "put-method",
		Short: "Add a method to a resource",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.PutMethod(cmd.Context(), &apigateway.PutMethodInput{
				RestApiId:         aws.String(restAPIID),
				ResourceId:        aws.String(resourceID),
				HttpMethod:        aws.String(httpMethod),
				AuthorizationType: aws.String(authType),
			})
			if err != nil {
				return fmt.Errorf("put-method failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")
	cmd.Flags().StringVar(&httpMethod, "http-method", "", "HTTP method")
	cmd.Flags().StringVar(&authType, "authorization-type", "NONE", "Authorization type")

	return cmd
}

func newAPIGatewayGetMethodCmd() *cobra.Command {
	var restAPIID, resourceID, httpMethod string

	cmd := &cobra.Command{
		Use:   "get-method",
		Short: "Get a method",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetMethod(cmd.Context(), &apigateway.GetMethodInput{
				RestApiId:  aws.String(restAPIID),
				ResourceId: aws.String(resourceID),
				HttpMethod: aws.String(httpMethod),
			})
			if err != nil {
				return fmt.Errorf("get-method failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")
	cmd.Flags().StringVar(&httpMethod, "http-method", "", "HTTP method")

	return cmd
}

func newAPIGatewayPutIntegrationCmd() *cobra.Command {
	var restAPIID, resourceID, httpMethod, integrationType, uri string

	cmd := &cobra.Command{
		Use:   "put-integration",
		Short: "Add an integration to a method",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &apigateway.PutIntegrationInput{
				RestApiId:  aws.String(restAPIID),
				ResourceId: aws.String(resourceID),
				HttpMethod: aws.String(httpMethod),
				Type:       apigatewayIntegrationType(integrationType),
			}

			if uri != "" {
				input.Uri = aws.String(uri)
			}

			out, err := client.PutIntegration(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("put-integration failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")
	cmd.Flags().StringVar(&httpMethod, "http-method", "", "HTTP method")
	cmd.Flags().StringVar(&integrationType, "type", "", "Integration type (HTTP, AWS, MOCK, etc.)")
	cmd.Flags().StringVar(&uri, "uri", "", "Integration URI")

	return cmd
}

func newAPIGatewayGetIntegrationCmd() *cobra.Command {
	var restAPIID, resourceID, httpMethod string

	cmd := &cobra.Command{
		Use:   "get-integration",
		Short: "Get an integration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetIntegration(cmd.Context(), &apigateway.GetIntegrationInput{
				RestApiId:  aws.String(restAPIID),
				ResourceId: aws.String(resourceID),
				HttpMethod: aws.String(httpMethod),
			})
			if err != nil {
				return fmt.Errorf("get-integration failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")
	cmd.Flags().StringVar(&httpMethod, "http-method", "", "HTTP method")

	return cmd
}

func newAPIGatewayCreateDeploymentCmd() *cobra.Command {
	var restAPIID, description string

	cmd := &cobra.Command{
		Use:   "create-deployment",
		Short: "Create a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &apigateway.CreateDeploymentInput{
				RestApiId: aws.String(restAPIID),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			out, err := client.CreateDeployment(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-deployment failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&description, "description", "", "Deployment description")

	return cmd
}

func newAPIGatewayGetDeploymentsCmd() *cobra.Command {
	var restAPIID string

	cmd := &cobra.Command{
		Use:   "get-deployments",
		Short: "List deployments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetDeployments(cmd.Context(), &apigateway.GetDeploymentsInput{
				RestApiId: aws.String(restAPIID),
			})
			if err != nil {
				return fmt.Errorf("get-deployments failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")

	return cmd
}

func newAPIGatewayGetDeploymentCmd() *cobra.Command {
	var restAPIID, deploymentID string

	cmd := &cobra.Command{
		Use:   "get-deployment",
		Short: "Get a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetDeployment(cmd.Context(), &apigateway.GetDeploymentInput{
				RestApiId:    aws.String(restAPIID),
				DeploymentId: aws.String(deploymentID),
			})
			if err != nil {
				return fmt.Errorf("get-deployment failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&deploymentID, "deployment-id", "", "Deployment ID")

	return cmd
}

func newAPIGatewayDeleteDeploymentCmd() *cobra.Command {
	var restAPIID, deploymentID string

	cmd := &cobra.Command{
		Use:   "delete-deployment",
		Short: "Delete a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteDeployment(cmd.Context(), &apigateway.DeleteDeploymentInput{
				RestApiId:    aws.String(restAPIID),
				DeploymentId: aws.String(deploymentID),
			})
			if err != nil {
				return fmt.Errorf("delete-deployment failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&deploymentID, "deployment-id", "", "Deployment ID")

	return cmd
}

func newAPIGatewayCreateStageCmd() *cobra.Command {
	var restAPIID, stageName, deploymentID, description string

	cmd := &cobra.Command{
		Use:   "create-stage",
		Short: "Create a stage",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &apigateway.CreateStageInput{
				RestApiId:    aws.String(restAPIID),
				StageName:    aws.String(stageName),
				DeploymentId: aws.String(deploymentID),
			}

			if description != "" {
				input.Description = aws.String(description)
			}

			out, err := client.CreateStage(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("create-stage failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&stageName, "stage-name", "", "Stage name")
	cmd.Flags().StringVar(&deploymentID, "deployment-id", "", "Deployment ID")
	cmd.Flags().StringVar(&description, "description", "", "Stage description")

	return cmd
}

func newAPIGatewayGetStagesCmd() *cobra.Command {
	var restAPIID string

	cmd := &cobra.Command{
		Use:   "get-stages",
		Short: "List stages",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetStages(cmd.Context(), &apigateway.GetStagesInput{
				RestApiId: aws.String(restAPIID),
			})
			if err != nil {
				return fmt.Errorf("get-stages failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")

	return cmd
}

func newAPIGatewayGetStageCmd() *cobra.Command {
	var restAPIID, stageName string

	cmd := &cobra.Command{
		Use:   "get-stage",
		Short: "Get a stage",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetStage(cmd.Context(), &apigateway.GetStageInput{
				RestApiId: aws.String(restAPIID),
				StageName: aws.String(stageName),
			})
			if err != nil {
				return fmt.Errorf("get-stage failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&stageName, "stage-name", "", "Stage name")

	return cmd
}

func newAPIGatewayDeleteStageCmd() *cobra.Command {
	var restAPIID, stageName string

	cmd := &cobra.Command{
		Use:   "delete-stage",
		Short: "Delete a stage",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := apigateway.NewFromConfig(cfg, func(o *apigateway.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteStage(cmd.Context(), &apigateway.DeleteStageInput{
				RestApiId: aws.String(restAPIID),
				StageName: aws.String(stageName),
			})
			if err != nil {
				return fmt.Errorf("delete-stage failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&restAPIID, "rest-api-id", "", "REST API ID")
	cmd.Flags().StringVar(&stageName, "stage-name", "", "Stage name")

	return cmd
}

func apigatewayIntegrationType(s string) apigatewayTypes.IntegrationType {
	return apigatewayTypes.IntegrationType(s)
}
