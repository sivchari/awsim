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
// Note: Scheduler uses /scheduler prefix to avoid conflicts with S3 wildcard routes.
func (s *Service) Prefix() string {
	return "/scheduler"
}

// RegisterRoutes registers the service routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Schedule operations
	r.HandleFunc("POST", "/scheduler/schedules/{name}", s.CreateSchedule)
	r.HandleFunc("GET", "/scheduler/schedules/{name}", s.GetSchedule)
	r.HandleFunc("PUT", "/scheduler/schedules/{name}", s.UpdateSchedule)
	r.HandleFunc("DELETE", "/scheduler/schedules/{name}", s.DeleteSchedule)
	r.HandleFunc("GET", "/scheduler/schedules", s.ListSchedules)

	// Schedule group operations
	r.HandleFunc("POST", "/scheduler/schedule-groups/{name}", s.CreateScheduleGroup)
	r.HandleFunc("GET", "/scheduler/schedule-groups/{name}", s.GetScheduleGroup)
	r.HandleFunc("DELETE", "/scheduler/schedule-groups/{name}", s.DeleteScheduleGroup)
	r.HandleFunc("GET", "/scheduler/schedule-groups", s.ListScheduleGroups)
}

// Ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)
