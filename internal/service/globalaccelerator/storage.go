package globalaccelerator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errNotFound           = "AcceleratorNotFoundException"
	errListenerNotFound   = "ListenerNotFoundException"
	errEndpointNotFound   = "EndpointGroupNotFoundException"
	errAcceleratorEnabled = "AcceleratorNotDisabledException"

	defaultAccountID = "000000000000"
)

// Storage defines the Global Accelerator storage interface.
type Storage interface {
	// Accelerator operations.
	CreateAccelerator(ctx context.Context, req *CreateAcceleratorRequest) (*Accelerator, error)
	GetAccelerator(ctx context.Context, arn string) (*Accelerator, error)
	ListAccelerators(ctx context.Context, maxResults int32, nextToken string) ([]*Accelerator, string, error)
	UpdateAccelerator(ctx context.Context, arn, name, ipAddressType string, enabled *bool) (*Accelerator, error)
	DeleteAccelerator(ctx context.Context, arn string) error

	// Listener operations.
	CreateListener(ctx context.Context, req *CreateListenerRequest) (*Listener, error)
	GetListener(ctx context.Context, arn string) (*Listener, error)
	ListListeners(ctx context.Context, acceleratorArn string, maxResults int32, nextToken string) ([]*Listener, string, error)
	UpdateListener(ctx context.Context, req *UpdateListenerRequest) (*Listener, error)
	DeleteListener(ctx context.Context, arn string) error

	// EndpointGroup operations.
	CreateEndpointGroup(ctx context.Context, req *CreateEndpointGroupRequest) (*EndpointGroup, error)
	GetEndpointGroup(ctx context.Context, arn string) (*EndpointGroup, error)
	ListEndpointGroups(ctx context.Context, listenerArn string, maxResults int32, nextToken string) ([]*EndpointGroup, string, error)
	UpdateEndpointGroup(ctx context.Context, req *UpdateEndpointGroupRequest) (*EndpointGroup, error)
	DeleteEndpointGroup(ctx context.Context, arn string) error
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu             sync.RWMutex
	accelerators   map[string]*Accelerator
	listeners      map[string]*Listener
	endpointGroups map[string]*EndpointGroup
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		accelerators:   make(map[string]*Accelerator),
		listeners:      make(map[string]*Listener),
		endpointGroups: make(map[string]*EndpointGroup),
	}
}

// CreateAccelerator creates a new accelerator.
func (s *MemoryStorage) CreateAccelerator(_ context.Context, req *CreateAcceleratorRequest) (*Accelerator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acceleratorID := uuid.New().String()
	arn := fmt.Sprintf("arn:aws:globalaccelerator::%s:accelerator/%s", defaultAccountID, acceleratorID)

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	ipAddressType := IpAddressTypeIPv4
	if req.IpAddressType != "" {
		ipAddressType = IpAddressType(req.IpAddressType)
	}

	// Generate static IP addresses.
	ipSets := []IpSet{
		{
			IpFamily:        "IPv4",
			IpAddresses:     []string{generateIPAddress(), generateIPAddress()},
			IpAddressFamily: "IPv4",
		},
	}

	if ipAddressType == IpAddressTypeDualStack {
		ipSets = append(ipSets, IpSet{
			IpFamily:        "IPv6",
			IpAddresses:     []string{generateIPv6Address()},
			IpAddressFamily: "IPv6",
		})
	}

	now := time.Now()
	accelerator := &Accelerator{
		AcceleratorArn: arn,
		Name:           req.Name,
		IpAddressType:  ipAddressType,
		Enabled:        enabled,
		IpSets:         ipSets,
		DNSName:        fmt.Sprintf("%s.awsglobalaccelerator.com", acceleratorID[:8]),
		Status:         AcceleratorStatusDeployed,
		CreatedTime:    now,
		LastModified:   now,
	}

	if ipAddressType == IpAddressTypeDualStack {
		accelerator.DualStackDNS = fmt.Sprintf("%s.dualstack.awsglobalaccelerator.com", acceleratorID[:8])
	}

	s.accelerators[arn] = accelerator

	return accelerator, nil
}

// GetAccelerator retrieves an accelerator by ARN.
func (s *MemoryStorage) GetAccelerator(_ context.Context, arn string) (*Accelerator, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accelerator, ok := s.accelerators[arn]
	if !ok {
		return nil, &ServiceError{Code: errNotFound, Message: "Accelerator not found"}
	}

	return accelerator, nil
}

// ListAccelerators lists all accelerators.
func (s *MemoryStorage) ListAccelerators(_ context.Context, maxResults int32, _ string) ([]*Accelerator, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	accelerators := make([]*Accelerator, 0, len(s.accelerators))
	for _, acc := range s.accelerators {
		accelerators = append(accelerators, acc)
		if len(accelerators) >= int(maxResults) {
			break
		}
	}

	return accelerators, "", nil
}

// UpdateAccelerator updates an accelerator.
func (s *MemoryStorage) UpdateAccelerator(_ context.Context, arn, name, ipAddressType string, enabled *bool) (*Accelerator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accelerator, ok := s.accelerators[arn]
	if !ok {
		return nil, &ServiceError{Code: errNotFound, Message: "Accelerator not found"}
	}

	if name != "" {
		accelerator.Name = name
	}

	if ipAddressType != "" {
		accelerator.IpAddressType = IpAddressType(ipAddressType)
	}

	if enabled != nil {
		accelerator.Enabled = *enabled
	}

	accelerator.LastModified = time.Now()

	return accelerator, nil
}

// DeleteAccelerator deletes an accelerator.
func (s *MemoryStorage) DeleteAccelerator(_ context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accelerator, ok := s.accelerators[arn]
	if !ok {
		return &ServiceError{Code: errNotFound, Message: "Accelerator not found"}
	}

	if accelerator.Enabled {
		return &ServiceError{Code: errAcceleratorEnabled, Message: "Accelerator must be disabled before deletion"}
	}

	// Delete associated listeners and endpoint groups.
	for listenerArn, listener := range s.listeners {
		if listener.AcceleratorArn == arn {
			// Delete endpoint groups for this listener.
			for egArn, eg := range s.endpointGroups {
				if eg.ListenerArn == listenerArn {
					delete(s.endpointGroups, egArn)
				}
			}

			delete(s.listeners, listenerArn)
		}
	}

	delete(s.accelerators, arn)

	return nil
}

// CreateListener creates a new listener.
func (s *MemoryStorage) CreateListener(_ context.Context, req *CreateListenerRequest) (*Listener, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify accelerator exists.
	if _, ok := s.accelerators[req.AcceleratorArn]; !ok {
		return nil, &ServiceError{Code: errNotFound, Message: "Accelerator not found"}
	}

	listenerID := uuid.New().String()
	arn := fmt.Sprintf("%s/listener/%s", req.AcceleratorArn, listenerID)

	portRanges := make([]PortRange, len(req.PortRanges))
	for i, pr := range req.PortRanges {
		portRanges[i] = PortRange{
			FromPort: pr.FromPort,
			ToPort:   pr.ToPort,
		}
	}

	clientAffinity := ClientAffinityNone
	if req.ClientAffinity != "" {
		clientAffinity = ClientAffinity(req.ClientAffinity)
	}

	listener := &Listener{
		ListenerArn:    arn,
		AcceleratorArn: req.AcceleratorArn,
		PortRanges:     portRanges,
		Protocol:       Protocol(req.Protocol),
		ClientAffinity: clientAffinity,
	}

	s.listeners[arn] = listener

	return listener, nil
}

// GetListener retrieves a listener by ARN.
func (s *MemoryStorage) GetListener(_ context.Context, arn string) (*Listener, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	listener, ok := s.listeners[arn]
	if !ok {
		return nil, &ServiceError{Code: errListenerNotFound, Message: "Listener not found"}
	}

	return listener, nil
}

// ListListeners lists listeners for an accelerator.
func (s *MemoryStorage) ListListeners(_ context.Context, acceleratorArn string, maxResults int32, _ string) ([]*Listener, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	listeners := make([]*Listener, 0)
	for _, listener := range s.listeners {
		if listener.AcceleratorArn == acceleratorArn {
			listeners = append(listeners, listener)
			if len(listeners) >= int(maxResults) {
				break
			}
		}
	}

	return listeners, "", nil
}

// UpdateListener updates a listener.
func (s *MemoryStorage) UpdateListener(_ context.Context, req *UpdateListenerRequest) (*Listener, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	listener, ok := s.listeners[req.ListenerArn]
	if !ok {
		return nil, &ServiceError{Code: errListenerNotFound, Message: "Listener not found"}
	}

	if len(req.PortRanges) > 0 {
		portRanges := make([]PortRange, len(req.PortRanges))
		for i, pr := range req.PortRanges {
			portRanges[i] = PortRange{
				FromPort: pr.FromPort,
				ToPort:   pr.ToPort,
			}
		}

		listener.PortRanges = portRanges
	}

	if req.Protocol != "" {
		listener.Protocol = Protocol(req.Protocol)
	}

	if req.ClientAffinity != "" {
		listener.ClientAffinity = ClientAffinity(req.ClientAffinity)
	}

	return listener, nil
}

// DeleteListener deletes a listener.
func (s *MemoryStorage) DeleteListener(_ context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.listeners[arn]; !ok {
		return &ServiceError{Code: errListenerNotFound, Message: "Listener not found"}
	}

	// Delete associated endpoint groups.
	for egArn, eg := range s.endpointGroups {
		if eg.ListenerArn == arn {
			delete(s.endpointGroups, egArn)
		}
	}

	delete(s.listeners, arn)

	return nil
}

// CreateEndpointGroup creates a new endpoint group.
func (s *MemoryStorage) CreateEndpointGroup(_ context.Context, req *CreateEndpointGroupRequest) (*EndpointGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify listener exists.
	if _, ok := s.listeners[req.ListenerArn]; !ok {
		return nil, &ServiceError{Code: errListenerNotFound, Message: "Listener not found"}
	}

	endpointGroupID := uuid.New().String()
	arn := fmt.Sprintf("%s/endpoint-group/%s", req.ListenerArn, endpointGroupID)

	trafficDialPercentage := 100.0
	if req.TrafficDialPercentage != nil {
		trafficDialPercentage = *req.TrafficDialPercentage
	}

	healthCheckProtocol := HealthCheckProtocolTCP
	if req.HealthCheckProtocol != "" {
		healthCheckProtocol = HealthCheckProtocol(req.HealthCheckProtocol)
	}

	healthCheckIntervalSeconds := int32(30)
	if req.HealthCheckIntervalSeconds != nil {
		healthCheckIntervalSeconds = *req.HealthCheckIntervalSeconds
	}

	thresholdCount := int32(3)
	if req.ThresholdCount != nil {
		thresholdCount = *req.ThresholdCount
	}

	endpoints := make([]EndpointDescription, len(req.EndpointConfigurations))
	for i, ec := range req.EndpointConfigurations {
		clientIPPreservation := false
		if ec.ClientIPPreservationEnabled != nil {
			clientIPPreservation = *ec.ClientIPPreservationEnabled
		}

		endpoints[i] = EndpointDescription{
			EndpointID:                  ec.EndpointID,
			Weight:                      ec.Weight,
			HealthState:                 HealthStateInitial,
			ClientIPPreservationEnabled: clientIPPreservation,
		}
	}

	portOverrides := make([]PortOverride, len(req.PortOverrides))
	for i, po := range req.PortOverrides {
		portOverrides[i] = PortOverride{
			ListenerPort: po.ListenerPort,
			EndpointPort: po.EndpointPort,
		}
	}

	endpointGroup := &EndpointGroup{
		EndpointGroupArn:           arn,
		ListenerArn:                req.ListenerArn,
		EndpointGroupRegion:        req.EndpointGroupRegion,
		EndpointDescriptions:       endpoints,
		TrafficDialPercentage:      trafficDialPercentage,
		HealthCheckPort:            req.HealthCheckPort,
		HealthCheckProtocol:        healthCheckProtocol,
		HealthCheckPath:            req.HealthCheckPath,
		HealthCheckIntervalSeconds: healthCheckIntervalSeconds,
		ThresholdCount:             thresholdCount,
		PortOverrides:              portOverrides,
	}

	s.endpointGroups[arn] = endpointGroup

	return endpointGroup, nil
}

// GetEndpointGroup retrieves an endpoint group by ARN.
func (s *MemoryStorage) GetEndpointGroup(_ context.Context, arn string) (*EndpointGroup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	endpointGroup, ok := s.endpointGroups[arn]
	if !ok {
		return nil, &ServiceError{Code: errEndpointNotFound, Message: "Endpoint group not found"}
	}

	return endpointGroup, nil
}

// ListEndpointGroups lists endpoint groups for a listener.
func (s *MemoryStorage) ListEndpointGroups(_ context.Context, listenerArn string, maxResults int32, _ string) ([]*EndpointGroup, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	endpointGroups := make([]*EndpointGroup, 0)
	for _, eg := range s.endpointGroups {
		if eg.ListenerArn == listenerArn {
			endpointGroups = append(endpointGroups, eg)
			if len(endpointGroups) >= int(maxResults) {
				break
			}
		}
	}

	return endpointGroups, "", nil
}

// UpdateEndpointGroup updates an endpoint group.
func (s *MemoryStorage) UpdateEndpointGroup(_ context.Context, req *UpdateEndpointGroupRequest) (*EndpointGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eg, ok := s.endpointGroups[req.EndpointGroupArn]
	if !ok {
		return nil, &ServiceError{Code: errEndpointNotFound, Message: "Endpoint group not found"}
	}

	if len(req.EndpointConfigurations) > 0 {
		endpoints := make([]EndpointDescription, len(req.EndpointConfigurations))
		for i, ec := range req.EndpointConfigurations {
			clientIPPreservation := false
			if ec.ClientIPPreservationEnabled != nil {
				clientIPPreservation = *ec.ClientIPPreservationEnabled
			}

			endpoints[i] = EndpointDescription{
				EndpointID:                  ec.EndpointID,
				Weight:                      ec.Weight,
				HealthState:                 HealthStateInitial,
				ClientIPPreservationEnabled: clientIPPreservation,
			}
		}

		eg.EndpointDescriptions = endpoints
	}

	if req.TrafficDialPercentage != nil {
		eg.TrafficDialPercentage = *req.TrafficDialPercentage
	}

	if req.HealthCheckPort != nil {
		eg.HealthCheckPort = req.HealthCheckPort
	}

	if req.HealthCheckProtocol != "" {
		eg.HealthCheckProtocol = HealthCheckProtocol(req.HealthCheckProtocol)
	}

	if req.HealthCheckPath != "" {
		eg.HealthCheckPath = req.HealthCheckPath
	}

	if req.HealthCheckIntervalSeconds != nil {
		eg.HealthCheckIntervalSeconds = *req.HealthCheckIntervalSeconds
	}

	if req.ThresholdCount != nil {
		eg.ThresholdCount = *req.ThresholdCount
	}

	if len(req.PortOverrides) > 0 {
		portOverrides := make([]PortOverride, len(req.PortOverrides))
		for i, po := range req.PortOverrides {
			portOverrides[i] = PortOverride{
				ListenerPort: po.ListenerPort,
				EndpointPort: po.EndpointPort,
			}
		}

		eg.PortOverrides = portOverrides
	}

	return eg, nil
}

// DeleteEndpointGroup deletes an endpoint group.
func (s *MemoryStorage) DeleteEndpointGroup(_ context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.endpointGroups[arn]; !ok {
		return &ServiceError{Code: errEndpointNotFound, Message: "Endpoint group not found"}
	}

	delete(s.endpointGroups, arn)

	return nil
}

// generateIPAddress generates a simulated static IP address.
func generateIPAddress() string {
	// Use 75.2.x.x range which is typical for Global Accelerator.
	return fmt.Sprintf("75.2.%d.%d", randomByte(), randomByte())
}

// generateIPv6Address generates a simulated IPv6 address.
func generateIPv6Address() string {
	return fmt.Sprintf("2600:9000:a%03x::%d", randomByte(), randomByte())
}

// randomByte generates a random byte value using UUID.
func randomByte() int {
	id := uuid.New()

	return int(id[0])
}
