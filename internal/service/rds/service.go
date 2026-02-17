package rds

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service implements the RDS service.
type Service struct {
	storage Storage
}

// New creates a new RDS service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "rds"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers routes with the router.
// RDS uses Query protocol, so routes are registered via DispatchAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// RDS uses Query protocol, routing is handled by DispatchAction.
}

// TargetPrefix returns the X-Amz-Target header prefix for RDS.
// RDS uses Action parameter instead of X-Amz-Target header.
func (s *Service) TargetPrefix() string {
	return "AmazonRDSv19"
}

// Actions returns the list of action names this service handles.
func (s *Service) Actions() []string {
	return []string{
		"CreateDBInstance",
		"DeleteDBInstance",
		"DescribeDBInstances",
		"ModifyDBInstance",
		"StartDBInstance",
		"StopDBInstance",
		"CreateDBCluster",
		"DeleteDBCluster",
		"DescribeDBClusters",
		"CreateDBSnapshot",
		"DeleteDBSnapshot",
	}
}

// QueryProtocol is a marker method that indicates RDS uses AWS Query protocol.
func (s *Service) QueryProtocol() {}
