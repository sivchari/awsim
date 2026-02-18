package elasticache

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the ElastiCache service.
type Service struct {
	storage Storage
}

// New creates a new ElastiCache service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "elasticache"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers routes with the router.
// ElastiCache uses Query protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// ElastiCache uses Query protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for ElastiCache.
// ElastiCache uses Action parameter instead of X-Amz-Target header.
func (s *Service) TargetPrefix() string {
	return "AmazonElastiCacheV9"
}

// Actions returns the list of action names this service handles.
func (s *Service) Actions() []string {
	return []string{
		"CreateCacheCluster",
		"DeleteCacheCluster",
		"DescribeCacheClusters",
		"ModifyCacheCluster",
		"CreateReplicationGroup",
		"DeleteReplicationGroup",
		"DescribeReplicationGroups",
	}
}

// QueryProtocol is a marker method that indicates ElastiCache uses AWS Query protocol.
func (s *Service) QueryProtocol() {}
