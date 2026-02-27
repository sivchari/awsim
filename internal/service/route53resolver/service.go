package route53resolver

import "github.com/sivchari/awsim/internal/service"

// Compile-time check to ensure Service implements JSONProtocolService.
var _ service.JSONProtocolService = (*Service)(nil)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Route 53 Resolver service.
type Service struct {
	storage Storage
}

// New creates a new Route 53 Resolver service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "route53resolver"
}

// RegisterRoutes registers the Route 53 Resolver routes.
// Route 53 Resolver uses AWS JSON 1.1 protocol via the JSONProtocolService interface.
func (s *Service) RegisterRoutes(_ service.Router) {
	// No routes to register - Route 53 Resolver uses JSON protocol dispatcher
}

// TargetPrefix returns the X-Amz-Target header prefix for Route 53 Resolver.
func (s *Service) TargetPrefix() string {
	return "Route53Resolver"
}

// JSONProtocol is a marker method that indicates Route 53 Resolver uses AWS JSON 1.1 protocol.
func (s *Service) JSONProtocol() {}
