// Package main is the entry point for the awsim CLI.
package main

import (
	"log"

	"github.com/sivchari/awsim/internal/server"
	// Register services via init().
	_ "github.com/sivchari/awsim/internal/service/dynamodb"
	_ "github.com/sivchari/awsim/internal/service/s3"
	_ "github.com/sivchari/awsim/internal/service/secretsmanager"
	_ "github.com/sivchari/awsim/internal/service/sqs"
)

func main() {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
