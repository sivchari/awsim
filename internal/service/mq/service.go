package mq

import (
	"net/http"
	"strings"

	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Amazon MQ service.
type Service struct {
	storage Storage
}

// New creates a new MQ service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "mq"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return "/mq"
}

// RegisterRoutes registers the MQ routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Broker operations
	r.Handle("POST", "/mq/v1/brokers", s.handleCreateBroker)
	r.Handle("GET", "/mq/v1/brokers", s.handleListBrokers)
	r.Handle("GET", "/mq/v1/brokers/{broker-id}", s.handleDescribeBroker)
	r.Handle("DELETE", "/mq/v1/brokers/{broker-id}", s.handleDeleteBroker)
	r.Handle("PUT", "/mq/v1/brokers/{broker-id}", s.handleUpdateBroker)

	// Configuration operations
	r.Handle("POST", "/mq/v1/configurations", s.handleCreateConfiguration)
}

// handleCreateBroker handles POST /v1/brokers.
func (s *Service) handleCreateBroker(w http.ResponseWriter, r *http.Request) {
	s.CreateBroker(w, r)
}

// handleListBrokers handles GET /v1/brokers.
func (s *Service) handleListBrokers(w http.ResponseWriter, r *http.Request) {
	s.ListBrokers(w, r)
}

// handleDescribeBroker handles GET /v1/brokers/{broker-id}.
func (s *Service) handleDescribeBroker(w http.ResponseWriter, r *http.Request) {
	brokerID := extractBrokerIDFromPath(r.URL.Path)
	s.DescribeBroker(w, r, brokerID)
}

// handleDeleteBroker handles DELETE /v1/brokers/{broker-id}.
func (s *Service) handleDeleteBroker(w http.ResponseWriter, r *http.Request) {
	brokerID := extractBrokerIDFromPath(r.URL.Path)
	s.DeleteBroker(w, r, brokerID)
}

// handleUpdateBroker handles PUT /v1/brokers/{broker-id}.
func (s *Service) handleUpdateBroker(w http.ResponseWriter, r *http.Request) {
	brokerID := extractBrokerIDFromPath(r.URL.Path)
	s.UpdateBroker(w, r, brokerID)
}

// handleCreateConfiguration handles POST /v1/configurations.
func (s *Service) handleCreateConfiguration(w http.ResponseWriter, r *http.Request) {
	s.CreateConfiguration(w, r)
}

// extractBrokerIDFromPath extracts broker ID from URL path.
// Path format: /mq/v1/brokers/{broker-id}
func extractBrokerIDFromPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "mq" && parts[1] == "v1" && parts[2] == "brokers" {
		return parts[3]
	}

	return ""
}
