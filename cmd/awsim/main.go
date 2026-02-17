// Package main is the entry point for the awsim CLI.
package main

import (
	"log"

	"github.com/sivchari/awsim/internal/server"
	// Register services via init().
	_ "github.com/sivchari/awsim/internal/service/acm"
	_ "github.com/sivchari/awsim/internal/service/apigateway"
	_ "github.com/sivchari/awsim/internal/service/appsync"
	_ "github.com/sivchari/awsim/internal/service/athena"
	_ "github.com/sivchari/awsim/internal/service/batch"
	_ "github.com/sivchari/awsim/internal/service/cloudfront"
	_ "github.com/sivchari/awsim/internal/service/cloudwatch"
	_ "github.com/sivchari/awsim/internal/service/cloudwatchlogs"
	_ "github.com/sivchari/awsim/internal/service/codeconnections"
	_ "github.com/sivchari/awsim/internal/service/cognito"
	_ "github.com/sivchari/awsim/internal/service/dynamodb"
	_ "github.com/sivchari/awsim/internal/service/ec2"
	_ "github.com/sivchari/awsim/internal/service/ecr"
	_ "github.com/sivchari/awsim/internal/service/ecs"
	_ "github.com/sivchari/awsim/internal/service/eks"
	_ "github.com/sivchari/awsim/internal/service/eventbridge"
	_ "github.com/sivchari/awsim/internal/service/firehose"
	_ "github.com/sivchari/awsim/internal/service/globalaccelerator"
	_ "github.com/sivchari/awsim/internal/service/glue"
	_ "github.com/sivchari/awsim/internal/service/iam"
	_ "github.com/sivchari/awsim/internal/service/kinesis"
	_ "github.com/sivchari/awsim/internal/service/kms"
	_ "github.com/sivchari/awsim/internal/service/lambda"
	_ "github.com/sivchari/awsim/internal/service/rds"
	_ "github.com/sivchari/awsim/internal/service/s3"
	_ "github.com/sivchari/awsim/internal/service/s3tables"
	_ "github.com/sivchari/awsim/internal/service/secretsmanager"
	_ "github.com/sivchari/awsim/internal/service/sesv2"
	_ "github.com/sivchari/awsim/internal/service/sfn"
	_ "github.com/sivchari/awsim/internal/service/sns"
	_ "github.com/sivchari/awsim/internal/service/sqs"
	_ "github.com/sivchari/awsim/internal/service/ssm"
	_ "github.com/sivchari/awsim/internal/service/xray"
)

func main() {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
