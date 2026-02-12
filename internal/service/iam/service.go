// Package iam provides IAM service emulation for awsim.
package iam

import (
	"net/http"

	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the IAM service.
type Service struct {
	storage        Storage
	actionHandlers map[string]http.HandlerFunc
}

// New creates a new IAM service.
func New(storage Storage) *Service {
	s := &Service{
		storage: storage,
	}
	s.initActionHandlers()

	return s
}

// initActionHandlers initializes the action handlers map.
func (s *Service) initActionHandlers() {
	s.actionHandlers = map[string]http.HandlerFunc{
		// User management
		"CreateUser": s.CreateUser,
		"DeleteUser": s.DeleteUser,
		"GetUser":    s.GetUser,
		"ListUsers":  s.ListUsers,
		// Role management
		"CreateRole": s.CreateRole,
		"DeleteRole": s.DeleteRole,
		"GetRole":    s.GetRole,
		"ListRoles":  s.ListRoles,
		// Policy management
		"CreatePolicy": s.CreatePolicy,
		"DeletePolicy": s.DeletePolicy,
		"GetPolicy":    s.GetPolicy,
		"ListPolicies": s.ListPolicies,
		// Policy attachments
		"AttachUserPolicy": s.AttachUserPolicy,
		"DetachUserPolicy": s.DetachUserPolicy,
		"AttachRolePolicy": s.AttachRolePolicy,
		"DetachRolePolicy": s.DetachRolePolicy,
		// Access keys
		"CreateAccessKey": s.CreateAccessKey,
		"DeleteAccessKey": s.DeleteAccessKey,
		"ListAccessKeys":  s.ListAccessKeys,
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
	// Register both with and without trailing slash for SDK compatibility.
	r.HandleFunc("POST", "/iam/", s.DispatchAction)
	r.HandleFunc("GET", "/iam/", s.DispatchAction)
	r.HandleFunc("POST", "/iam", s.DispatchAction)
	r.HandleFunc("GET", "/iam", s.DispatchAction)
}
