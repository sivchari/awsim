// Package main is the entry point for the awsim CLI.
package main

import (
	"log"

	"github.com/sivchari/awsim/internal/server"
)

func main() {
	cfg := server.DefaultConfig()
	srv := server.New(cfg)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
