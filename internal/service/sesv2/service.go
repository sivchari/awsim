package sesv2

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the SES v2 service.
type Service struct {
	storage Storage
}

// New creates a new SES v2 service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "sesv2"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return "/ses"
}

// RegisterRoutes registers the SES v2 routes.
// SES v2 uses REST API with path-based routing.
func (s *Service) RegisterRoutes(r service.Router) {
	// Email Identity routes.
	r.HandleFunc("POST", "/ses/v2/email/identities", s.CreateEmailIdentity)
	r.HandleFunc("GET", "/ses/v2/email/identities", s.ListEmailIdentities)
	r.HandleFunc("GET", "/ses/v2/email/identities/{emailIdentity}", s.GetEmailIdentity)
	r.HandleFunc("DELETE", "/ses/v2/email/identities/{emailIdentity}", s.DeleteEmailIdentity)

	// Configuration Set routes.
	r.HandleFunc("POST", "/ses/v2/email/configuration-sets", s.CreateConfigurationSet)
	r.HandleFunc("GET", "/ses/v2/email/configuration-sets", s.ListConfigurationSets)
	r.HandleFunc("GET", "/ses/v2/email/configuration-sets/{configurationSetName}", s.GetConfigurationSet)
	r.HandleFunc("DELETE", "/ses/v2/email/configuration-sets/{configurationSetName}", s.DeleteConfigurationSet)

	// Send Email route.
	r.HandleFunc("POST", "/ses/v2/email/outbound-emails", s.SendEmail)
}
