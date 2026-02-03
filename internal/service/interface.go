// Package service provides interfaces and utilities for AWS service implementations.
package service

import (
	"net/http"
)

// Service is the common interface for all AWS service implementations.
type Service interface {
	// Name returns the service name (e.g., "s3", "sqs", "dynamodb").
	Name() string

	// Prefix returns the URL prefix for path-based routing (e.g., "/s3").
	// Returns empty string for host-based routing.
	Prefix() string

	// RegisterRoutes registers the service's routes with the router.
	RegisterRoutes(r Router)
}

// Router is the interface for registering HTTP routes.
type Router interface {
	// Handle registers a handler for the given method and pattern.
	Handle(method, pattern string, handler http.HandlerFunc)

	// HandleFunc is an alias for Handle for compatibility.
	HandleFunc(method, pattern string, handler http.HandlerFunc)
}

// Handler is the interface for operation handlers.
type Handler interface {
	// ServeHTTP handles the HTTP request.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
