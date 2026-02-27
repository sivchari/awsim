package pipes

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the EventBridge Pipes service.
type Service struct {
	storage Storage
}

// New creates a new Pipes service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "pipes"
}

// RegisterRoutes registers the Pipes routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Pipe CRUD operations.
	r.Handle("POST", "/v1/pipes/{name}", s.CreatePipe)
	r.Handle("GET", "/v1/pipes/{name}", s.DescribePipe)
	r.Handle("PUT", "/v1/pipes/{name}", s.UpdatePipe)
	r.Handle("DELETE", "/v1/pipes/{name}", s.DeletePipe)

	// List pipes.
	r.Handle("GET", "/v1/pipes", s.ListPipes)

	// Pipe control operations.
	r.Handle("POST", "/v1/pipes/{name}/start", s.StartPipe)
	r.Handle("POST", "/v1/pipes/{name}/stop", s.StopPipe)

	// Tag operations.
	r.Handle("POST", "/tags/{arn...}", s.TagResource)
	r.Handle("DELETE", "/tags/{arn...}", s.UntagResource)
	r.Handle("GET", "/tags/{arn...}", s.ListTagsForResource)
}

var _ service.Service = (*Service)(nil)
