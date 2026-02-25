// Package scheduler provides EventBridge Scheduler service emulation for awsim.
package scheduler

import (
	"github.com/sivchari/awsim/internal/service"
)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the EventBridge Scheduler service.
type Service struct {
	storage Storage
}

// New creates a new Scheduler service.
func New(storage Storage) *Service {
	return &Service{storage: storage}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "scheduler"
}

// Prefix returns the URL prefix for the service.
func (s *Service) Prefix() string {
	return ""
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Schedule operations
	r.HandleFunc("POST", "/schedules/{name}", s.CreateSchedule)
	r.HandleFunc("GET", "/schedules/{name}", s.GetSchedule)
	r.HandleFunc("PUT", "/schedules/{name}", s.UpdateSchedule)
	r.HandleFunc("DELETE", "/schedules/{name}", s.DeleteSchedule)
	r.HandleFunc("GET", "/schedules", s.ListSchedules)

	// Schedule group operations
	r.HandleFunc("POST", "/schedule-groups/{name}", s.CreateScheduleGroup)
	r.HandleFunc("GET", "/schedule-groups/{name}", s.GetScheduleGroup)
	r.HandleFunc("DELETE", "/schedule-groups/{name}", s.DeleteScheduleGroup)
	r.HandleFunc("GET", "/schedule-groups", s.ListScheduleGroups)
}

// Ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)
