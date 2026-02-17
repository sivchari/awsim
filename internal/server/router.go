package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Route represents a registered HTTP route.
type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// Router is the HTTP router for awsim.
type Router struct {
	mux           *http.ServeMux
	routes        []Route
	prefixRouters map[string]*http.ServeMux // Separate routers for services with prefixes
	logger        *slog.Logger
}

// NewRouter creates a new router.
func NewRouter(logger *slog.Logger) *Router {
	r := &Router{
		mux:           http.NewServeMux(),
		routes:        make([]Route, 0),
		prefixRouters: make(map[string]*http.ServeMux),
		logger:        logger,
	}

	return r
}

// Handle registers a handler for the given method and pattern.
func (r *Router) Handle(method, pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	})

	// Check if this is a prefixed route (e.g., /lambda/...)
	// Routes with specific prefixes are registered in separate ServeMux instances
	// to avoid conflicts with wildcard routes like /{bucket}/{key...}
	prefix := extractRoutePrefix(pattern)
	if prefix != "" {
		if _, ok := r.prefixRouters[prefix]; !ok {
			r.prefixRouters[prefix] = http.NewServeMux()
		}

		fullPattern := method + " " + pattern
		r.prefixRouters[prefix].HandleFunc(fullPattern, r.wrapHandler(method, pattern, handler))
		r.logger.Debug("registered prefixed route", "method", method, "pattern", pattern, "prefix", prefix)

		return
	}

	// Use Go 1.22+ method pattern
	fullPattern := method + " " + pattern
	r.mux.HandleFunc(fullPattern, r.wrapHandler(method, pattern, handler))
	r.logger.Debug("registered route", "method", method, "pattern", pattern)
}

// extractRoutePrefix extracts service prefixes like "/lambda" from patterns.
// Returns empty string for patterns without service prefixes.
func extractRoutePrefix(pattern string) string {
	// Known service prefixes that need isolation from wildcard routes
	// S3 Tables uses /buckets, /namespaces, /tables, /get-table paths
	// CloudFront uses /2020-05-31 versioned paths
	// /service is for RPC v2 CBOR protocol
	prefixes := []string{"/lambda", "/eks", "/iam", "/buckets", "/namespaces", "/tables", "/get-table", "/apigateway", "/ses", "/2020-05-31", "/service"}

	for _, prefix := range prefixes {
		if len(pattern) >= len(prefix) && pattern[:len(prefix)] == prefix {
			return prefix
		}
	}

	return ""
}

// HandleFunc is an alias for Handle for compatibility with service.Router interface.
func (r *Router) HandleFunc(method, pattern string, handler http.HandlerFunc) {
	r.Handle(method, pattern, handler)
}

// wrapHandler wraps a handler with logging and request ID injection.
func (r *Router) wrapHandler(method, pattern string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Add AWS-style headers
		w.Header().Set("x-amz-request-id", requestID)
		w.Header().Set("x-amzn-RequestId", requestID)

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the actual handler
		handler(wrapped, req)

		// Log the request
		r.logger.Info("request",
			"method", method,
			"path", req.URL.Path,
			"pattern", pattern,
			"status", wrapped.statusCode,
			"duration", time.Since(start),
			"request_id", requestID,
		)
	}
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Handle health endpoint before ServeMux to avoid route conflicts.
	if req.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))

		return
	}

	// Check if the request matches a prefix router first
	for prefix, mux := range r.prefixRouters {
		if len(req.URL.Path) >= len(prefix) && req.URL.Path[:len(prefix)] == prefix {
			mux.ServeHTTP(w, req)

			return
		}
	}

	r.mux.ServeHTTP(w, req)
}

// Routes returns all registered routes.
func (r *Router) Routes() []Route {
	return r.routes
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
