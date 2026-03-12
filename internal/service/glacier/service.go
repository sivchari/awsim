package glacier

import (
	"github.com/sivchari/awsim/internal/service"
)

// Service implements the AWS Glacier service.
type Service struct {
	storage Storage
}

// New creates a new Glacier service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "glacier"
}

// RegisterRoutes registers the Glacier routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Vault operations.
	r.Handle("PUT", "/-/vaults/{vaultName}", s.CreateVault)
	r.Handle("GET", "/-/vaults/{vaultName}", s.DescribeVault)
	r.Handle("DELETE", "/-/vaults/{vaultName}", s.DeleteVault)
	r.Handle("GET", "/-/vaults", s.ListVaults)
}

func init() {
	service.Register(New(NewMemoryStorage()))
}
