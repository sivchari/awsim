// Package iam provides IAM service emulation for awsim.
package iam

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the IAM service.
type Service struct {
	storage Storage
}

// New creates a new IAM service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "iam"
}

// Prefix returns the URL prefix for this service.
func (s *Service) Prefix() string {
	return "/iam"
}

// RegisterRoutes registers the IAM routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// IAM uses a single endpoint with Action parameter.
	r.HandleFunc("POST", "/iam/", s.DispatchAction)
	r.HandleFunc("GET", "/iam/", s.DispatchAction)
}
