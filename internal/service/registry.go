package service

import (
	"sync"
)

// globalRegistry is the default registry for auto-registration via init().
var globalRegistry = NewRegistry()

// Register adds a service to the global registry.
// This is typically called from init() in each service package.
func Register(svc Service) {
	globalRegistry.Register(svc)
}

// Services returns all services from the global registry.
func Services() []Service {
	return globalRegistry.All()
}

// Registry manages service registration and discovery.
type Registry struct {
	mu       sync.RWMutex
	services map[string]Service
}

// NewRegistry creates a new service registry.
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]Service),
	}
}

// Register adds a service to the registry.
func (r *Registry) Register(svc Service) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[svc.Name()] = svc
}

// Get returns a service by name.
func (r *Registry) Get(name string) (Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, ok := r.services[name]

	return svc, ok
}

// All returns all registered services.
func (r *Registry) All() []Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]Service, 0, len(r.services))

	for _, svc := range r.services {
		services = append(services, svc)
	}

	return services
}

// Names returns the names of all registered services.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))

	for name := range r.services {
		names = append(names, name)
	}

	return names
}
