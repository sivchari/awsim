package lambda

import (
	"github.com/sivchari/awsim/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

func init() {
	service.Register(New(NewMemoryStorage(defaultBaseURL), defaultBaseURL))
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

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Lambda routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// CreateFunction: POST /2015-03-31/functions
	r.Handle("POST", "/2015-03-31/functions", s.CreateFunction)

	// ListFunctions: GET /2015-03-31/functions
	r.Handle("GET", "/2015-03-31/functions", s.ListFunctions)

	// GetFunction: GET /2015-03-31/functions/{FunctionName}
	r.Handle("GET", "/2015-03-31/functions/{functionName}", s.GetFunction)

	// DeleteFunction: DELETE /2015-03-31/functions/{FunctionName}
	r.Handle("DELETE", "/2015-03-31/functions/{functionName}", s.DeleteFunction)

	// UpdateFunctionCode: PUT /2015-03-31/functions/{FunctionName}/code
	r.Handle("PUT", "/2015-03-31/functions/{functionName}/code", s.UpdateFunctionCode)

	// UpdateFunctionConfiguration: PUT /2015-03-31/functions/{FunctionName}/configuration
	r.Handle("PUT", "/2015-03-31/functions/{functionName}/configuration", s.UpdateFunctionConfiguration)

	// Invoke: POST /2015-03-31/functions/{FunctionName}/invocations
	r.Handle("POST", "/2015-03-31/functions/{functionName}/invocations", s.Invoke)
}
