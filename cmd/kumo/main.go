// Package main is the entry point for the kumo CLI.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	kumocli "github.com/sivchari/kumo/cli"
	"github.com/sivchari/kumo/internal/server"
	_ "github.com/sivchari/kumo/internal/service/acm" // Register services via init().
	_ "github.com/sivchari/kumo/internal/service/amplify"
	_ "github.com/sivchari/kumo/internal/service/apigateway"
	_ "github.com/sivchari/kumo/internal/service/appmesh"
	_ "github.com/sivchari/kumo/internal/service/appsync"
	_ "github.com/sivchari/kumo/internal/service/athena"
	_ "github.com/sivchari/kumo/internal/service/backup"
	_ "github.com/sivchari/kumo/internal/service/batch"
	_ "github.com/sivchari/kumo/internal/service/ce"
	_ "github.com/sivchari/kumo/internal/service/cloudformation"
	_ "github.com/sivchari/kumo/internal/service/cloudfront"
	_ "github.com/sivchari/kumo/internal/service/cloudtrail"
	_ "github.com/sivchari/kumo/internal/service/cloudwatch"
	_ "github.com/sivchari/kumo/internal/service/cloudwatchlogs"
	_ "github.com/sivchari/kumo/internal/service/codeconnections"
	_ "github.com/sivchari/kumo/internal/service/codeguruprofiler"
	_ "github.com/sivchari/kumo/internal/service/codegurureviewer"
	_ "github.com/sivchari/kumo/internal/service/cognito"
	_ "github.com/sivchari/kumo/internal/service/comprehend"
	_ "github.com/sivchari/kumo/internal/service/configservice"
	_ "github.com/sivchari/kumo/internal/service/dataexchange"
	_ "github.com/sivchari/kumo/internal/service/dlm"
	_ "github.com/sivchari/kumo/internal/service/documentdb"
	_ "github.com/sivchari/kumo/internal/service/ds"
	_ "github.com/sivchari/kumo/internal/service/dynamodb"
	_ "github.com/sivchari/kumo/internal/service/ebs"
	_ "github.com/sivchari/kumo/internal/service/ec2"
	_ "github.com/sivchari/kumo/internal/service/ecr"
	_ "github.com/sivchari/kumo/internal/service/ecs"
	_ "github.com/sivchari/kumo/internal/service/eks"
	_ "github.com/sivchari/kumo/internal/service/elasticache"
	_ "github.com/sivchari/kumo/internal/service/elasticbeanstalk"
	_ "github.com/sivchari/kumo/internal/service/elbv2"
	_ "github.com/sivchari/kumo/internal/service/emrserverless"
	_ "github.com/sivchari/kumo/internal/service/entityresolution"
	_ "github.com/sivchari/kumo/internal/service/eventbridge"
	_ "github.com/sivchari/kumo/internal/service/finspace"
	_ "github.com/sivchari/kumo/internal/service/firehose"
	_ "github.com/sivchari/kumo/internal/service/forecast"
	_ "github.com/sivchari/kumo/internal/service/gamelift"
	_ "github.com/sivchari/kumo/internal/service/glacier"
	_ "github.com/sivchari/kumo/internal/service/globalaccelerator"
	_ "github.com/sivchari/kumo/internal/service/glue"
	_ "github.com/sivchari/kumo/internal/service/iam"
	_ "github.com/sivchari/kumo/internal/service/kafka"
	_ "github.com/sivchari/kumo/internal/service/kinesis"
	_ "github.com/sivchari/kumo/internal/service/kms"
	_ "github.com/sivchari/kumo/internal/service/lambda"
	_ "github.com/sivchari/kumo/internal/service/location"
	_ "github.com/sivchari/kumo/internal/service/macie2"
	_ "github.com/sivchari/kumo/internal/service/memorydb"
	_ "github.com/sivchari/kumo/internal/service/mq"
	_ "github.com/sivchari/kumo/internal/service/neptune"
	_ "github.com/sivchari/kumo/internal/service/organizations"
	_ "github.com/sivchari/kumo/internal/service/pinpointsmsvoicev2"
	_ "github.com/sivchari/kumo/internal/service/pipes"
	_ "github.com/sivchari/kumo/internal/service/rds"
	_ "github.com/sivchari/kumo/internal/service/redshift"
	_ "github.com/sivchari/kumo/internal/service/rekognition"
	_ "github.com/sivchari/kumo/internal/service/resiliencehub"
	_ "github.com/sivchari/kumo/internal/service/route53"
	_ "github.com/sivchari/kumo/internal/service/route53resolver"
	_ "github.com/sivchari/kumo/internal/service/s3"
	_ "github.com/sivchari/kumo/internal/service/s3control"
	_ "github.com/sivchari/kumo/internal/service/s3tables"
	_ "github.com/sivchari/kumo/internal/service/sagemaker"
	_ "github.com/sivchari/kumo/internal/service/scheduler"
	_ "github.com/sivchari/kumo/internal/service/secretsmanager"
	_ "github.com/sivchari/kumo/internal/service/securitylake"
	_ "github.com/sivchari/kumo/internal/service/servicequotas"
	_ "github.com/sivchari/kumo/internal/service/sesv2"
	_ "github.com/sivchari/kumo/internal/service/sfn"
	_ "github.com/sivchari/kumo/internal/service/sns"
	_ "github.com/sivchari/kumo/internal/service/sqs"
	_ "github.com/sivchari/kumo/internal/service/ssm"
	_ "github.com/sivchari/kumo/internal/service/sts"
	_ "github.com/sivchari/kumo/internal/service/xray"
)

func main() {
	root := kumocli.NewRootCmd()

	// Add serve command for backward compatibility.
	// Running `kumo` without subcommands also starts the server.
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the kumo server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServer()
		},
	}

	root.AddCommand(serveCmd)

	// If no subcommand is given, start the server (backward compatible).
	if len(os.Args) == 1 {
		if err := runServer(); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer() error {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	if err := srv.Run(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
