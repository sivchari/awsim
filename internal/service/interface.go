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

// JSONProtocolService is an optional interface for services using AWS JSON 1.0 protocol.
// Services implementing this interface will have their handlers dispatched via
// a unified POST / endpoint based on the X-Amz-Target header.
type JSONProtocolService interface {
	// TargetPrefix returns the X-Amz-Target prefix for this service.
	// e.g., "AmazonSQS" for SQS, "DynamoDB_20120810" for DynamoDB
	TargetPrefix() string

	// DispatchAction handles the JSON protocol request after routing.
	DispatchAction(w http.ResponseWriter, r *http.Request)

	// JSONProtocol is a marker method to distinguish from QueryProtocolService.
	JSONProtocol()
}

// QueryProtocolService is an optional interface for services using AWS Query protocol.
// Services implementing this interface will have their handlers dispatched via
// a unified POST / endpoint, with form data converted to JSON before dispatch.
type QueryProtocolService interface {
	// TargetPrefix returns the target prefix for this service.
	// This is used to set the X-Amz-Target header after converting
	// the Query request to JSON format.
	// e.g., "AmazonSimpleNotificationService" for SNS
	TargetPrefix() string

	// DispatchAction handles the request after Query-to-JSON conversion.
	DispatchAction(w http.ResponseWriter, r *http.Request)

	// Actions returns the list of action names this service handles.
	// This is used by the dispatcher to route requests to the correct service.
	Actions() []string

	// QueryProtocol is a marker method to distinguish from JSONProtocolService.
	QueryProtocol()
}
