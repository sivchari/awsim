package lambda

import (
	"fmt"
	"io"
	"os"

	"github.com/sivchari/kumo/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

// Compile-time check that Service implements io.Closer.
var _ io.Closer = (*Service)(nil)

func init() {
	var opts []Option
	if dir := os.Getenv("KUMO_DATA_DIR"); dir != "" {
		opts = append(opts, WithDataDir(dir))
	}

	service.Register(New(NewMemoryStorage(defaultBaseURL, opts...), defaultBaseURL))
}

// Service implements the Lambda service.
type Service struct {
	storage Storage
	baseURL string
}

// New creates a new Lambda service.
func New(storage Storage, baseURL string) *Service {
	return &Service{
		storage: storage,
		baseURL: baseURL,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "lambda"
}

// RegisterRoutes registers the Lambda routes.
// Note: Routes use /lambda prefix to avoid conflicts with S3 wildcard routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// CreateFunction: POST /lambda/2015-03-31/functions
	r.Handle("POST", "/lambda/2015-03-31/functions", s.CreateFunction)

	// ListFunctions: GET /lambda/2015-03-31/functions
	r.Handle("GET", "/lambda/2015-03-31/functions", s.ListFunctions)

	// GetFunction: GET /lambda/2015-03-31/functions/{FunctionName}
	r.Handle("GET", "/lambda/2015-03-31/functions/{functionName}", s.GetFunction)

	// DeleteFunction: DELETE /lambda/2015-03-31/functions/{FunctionName}
	r.Handle("DELETE", "/lambda/2015-03-31/functions/{functionName}", s.DeleteFunction)

	// UpdateFunctionCode: PUT /lambda/2015-03-31/functions/{FunctionName}/code
	r.Handle("PUT", "/lambda/2015-03-31/functions/{functionName}/code", s.UpdateFunctionCode)

	// UpdateFunctionConfiguration: PUT /lambda/2015-03-31/functions/{FunctionName}/configuration
	r.Handle("PUT", "/lambda/2015-03-31/functions/{functionName}/configuration", s.UpdateFunctionConfiguration)

	// Invoke: POST /lambda/2015-03-31/functions/{FunctionName}/invocations
	r.Handle("POST", "/lambda/2015-03-31/functions/{functionName}/invocations", s.Invoke)

	// EventSourceMapping operations
	// CreateEventSourceMapping: POST /lambda/2015-03-31/event-source-mappings
	r.Handle("POST", "/lambda/2015-03-31/event-source-mappings", s.CreateEventSourceMapping)

	// ListEventSourceMappings: GET /lambda/2015-03-31/event-source-mappings
	r.Handle("GET", "/lambda/2015-03-31/event-source-mappings", s.ListEventSourceMappings)

	// GetEventSourceMapping: GET /lambda/2015-03-31/event-source-mappings/{UUID}
	r.Handle("GET", "/lambda/2015-03-31/event-source-mappings/{uuid}", s.GetEventSourceMapping)

	// UpdateEventSourceMapping: PUT /lambda/2015-03-31/event-source-mappings/{UUID}
	r.Handle("PUT", "/lambda/2015-03-31/event-source-mappings/{uuid}", s.UpdateEventSourceMapping)

	// DeleteEventSourceMapping: DELETE /lambda/2015-03-31/event-source-mappings/{UUID}
	r.Handle("DELETE", "/lambda/2015-03-31/event-source-mappings/{uuid}", s.DeleteEventSourceMapping)
}

// Close saves the storage state if persistence is enabled.
func (s *Service) Close() error {
	if c, ok := s.storage.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close storage: %w", err)
		}
	}

	return nil
}
