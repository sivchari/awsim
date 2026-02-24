package route53

import "github.com/sivchari/awsim/internal/service"

func init() {
	storage := NewMemoryStorage()
	service.Register(New(storage))
}

// Service is the Route 53 service.
type Service struct {
	storage Storage
}

// New creates a new Route 53 service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "route53"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Hosted Zones
	r.Handle("POST", "/2013-04-01/hostedzone", s.CreateHostedZone)
	r.Handle("GET", "/2013-04-01/hostedzone", s.ListHostedZones)
	r.Handle("GET", "/2013-04-01/hostedzone/{id}", s.GetHostedZone)
	r.Handle("DELETE", "/2013-04-01/hostedzone/{id}", s.DeleteHostedZone)

	// Resource Record Sets
	r.Handle("POST", "/2013-04-01/hostedzone/{id}/rrset", s.ChangeResourceRecordSets)
	r.Handle("GET", "/2013-04-01/hostedzone/{id}/rrset", s.ListResourceRecordSets)
}
