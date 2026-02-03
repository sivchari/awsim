package sqs

import (
	"net/http"

	"github.com/sivchari/awsim/internal/service"
)

const defaultBaseURL = "http://localhost:4566"

func init() {
	service.Register(New(NewMemoryStorage(defaultBaseURL), defaultBaseURL))
}

// Service implements the SQS service.
type Service struct {
	storage Storage
	baseURL string
}

// New creates a new SQS service.
func New(storage Storage, baseURL string) *Service {
	return &Service{
		storage: storage,
		baseURL: baseURL,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "sqs"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the SQS routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// SQS uses POST with Action parameter for all operations.
	// The root endpoint handles all actions.
	r.HandleFunc("POST", "/", s.dispatchAction)

	// Also handle queue-specific paths for SDK compatibility.
	r.HandleFunc("POST", "/{accountID}/{queueName}", s.handleQueueAction)
}

// handleQueueAction handles actions on a specific queue URL path.
func (s *Service) handleQueueAction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSQSError(w, "InvalidParameterValue", "Failed to parse form", http.StatusBadRequest)

		return
	}

	// Construct queue URL from path.
	queueURL := s.baseURL + r.URL.Path
	r.Form.Set("QueueUrl", queueURL)

	s.dispatchAction(w, r)
}
