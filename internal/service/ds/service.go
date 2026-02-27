package ds

import (
	"net/http"

	"github.com/sivchari/awsim/internal/service"
)

func init() {
	svc := NewService()
	service.Register(svc)
}

// Service implements the Directory Service.
type Service struct {
	storage Storage
}

// NewService creates a new Directory Service.
func NewService() *Service {
	return &Service{
		storage: NewMemoryStorage(),
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "ds"
}

// RegisterRoutes registers the Directory Service routes.
// Directory Service uses AWS JSON 1.1 protocol with X-Amz-Target header.
func (s *Service) RegisterRoutes(_ service.Router) {
	// Directory Service uses POST / with X-Amz-Target header
	// Routes are handled by the JSON protocol dispatcher.
}

// TargetPrefix returns the X-Amz-Target prefix for this service.
func (s *Service) TargetPrefix() string {
	return "DirectoryService_20150416"
}

// DispatchAction dispatches the action to the appropriate handler.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	action := r.Header.Get("X-Amz-Target")

	switch action {
	case "DirectoryService_20150416.CreateDirectory":
		s.CreateDirectory(w, r)
	case "DirectoryService_20150416.DescribeDirectories":
		s.DescribeDirectories(w, r)
	case "DirectoryService_20150416.DeleteDirectory":
		s.DeleteDirectory(w, r)
	case "DirectoryService_20150416.CreateSnapshot":
		s.CreateSnapshot(w, r)
	case "DirectoryService_20150416.DescribeSnapshots":
		s.DescribeSnapshots(w, r)
	case "DirectoryService_20150416.DeleteSnapshot":
		s.DeleteSnapshot(w, r)
	default:
		writeDSError(w, ErrUnsupportedOperation, "Unsupported operation: "+action, http.StatusBadRequest)
	}
}

// JSONProtocol is a marker method for JSON protocol services.
func (s *Service) JSONProtocol() {}
