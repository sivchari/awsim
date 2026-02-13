package glue

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Glue service.
type Service struct {
	storage Storage
}

// New creates a new Glue service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "glue"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the Glue routes.
// Glue uses AWS JSON 1.1 protocol with X-Amz-Target header.
func (s *Service) RegisterRoutes(r service.Router) {
	// Glue uses POST requests with X-Amz-Target header for all operations.
	r.HandleFunc("POST", "/", s.handleRequest)
}
