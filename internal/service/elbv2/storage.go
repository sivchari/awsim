package elbv2

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"
)

// Storage defines the storage interface for ELB v2 service.
type Storage interface {
	CreateLoadBalancer(ctx context.Context, req *CreateLoadBalancerRequest) (*LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, loadBalancerArn string) error
	DescribeLoadBalancers(ctx context.Context, arns, names []string) ([]*LoadBalancer, error)

	CreateTargetGroup(ctx context.Context, req *CreateTargetGroupRequest) (*TargetGroup, error)
	DeleteTargetGroup(ctx context.Context, targetGroupArn string) error
	DescribeTargetGroups(ctx context.Context, arns, names []string, lbArn string) ([]*TargetGroup, error)

	RegisterTargets(ctx context.Context, targetGroupArn string, targets []Target) error
	DeregisterTargets(ctx context.Context, targetGroupArn string, targets []Target) error

	CreateListener(ctx context.Context, req *CreateListenerRequest) (*Listener, error)
	DeleteListener(ctx context.Context, listenerArn string) error
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu            sync.RWMutex
	loadBalancers map[string]*LoadBalancer // keyed by ARN
	targetGroups  map[string]*TargetGroup  // keyed by ARN
	listeners     map[string]*Listener     // keyed by ARN
	targets       map[string][]Target      // keyed by targetGroupArn
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		loadBalancers: make(map[string]*LoadBalancer),
		targetGroups:  make(map[string]*TargetGroup),
		listeners:     make(map[string]*Listener),
		targets:       make(map[string][]Target),
	}
}

// CreateLoadBalancer creates a new load balancer.
func (m *MemoryStorage) CreateLoadBalancer(_ context.Context, req *CreateLoadBalancerRequest) (*LoadBalancer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name.
	for _, lb := range m.loadBalancers {
		if lb.LoadBalancerName == req.Name {
			return nil, &Error{
				Code:    "DuplicateLoadBalancerName",
				Message: fmt.Sprintf("A load balancer with the name '%s' already exists", req.Name),
			}
		}
	}

	lbType := req.Type
	if lbType == "" {
		lbType = "application"
	}

	scheme := req.Scheme
	if scheme == "" {
		scheme = "internet-facing"
	}

	ipAddressType := req.IPAddressType
	if ipAddressType == "" {
		ipAddressType = "ipv4"
	}

	lbID := uuid.New().String()[:17]
	arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s/%s/%s",
		defaultRegion, defaultAccountID, lbType[:3], req.Name, lbID)

	dnsName := fmt.Sprintf("%s-%s.%s.elb.amazonaws.com", req.Name, lbID[:8], defaultRegion)

	azs := make([]AvailabilityZone, 0, len(req.Subnets))
	for i, subnet := range req.Subnets {
		azs = append(azs, AvailabilityZone{
			ZoneName: fmt.Sprintf("%s%c", defaultRegion, 'a'+byte(i%3)),
			SubnetID: subnet,
		})
	}

	lb := &LoadBalancer{
		LoadBalancerArn:       arn,
		DNSName:               dnsName,
		CanonicalHostedZoneID: "Z35SXDOTRQ7X7K",
		CreatedTime:           time.Now(),
		LoadBalancerName:      req.Name,
		Scheme:                scheme,
		VpcID:                 "vpc-" + uuid.New().String()[:8],
		State: LoadBalancerState{
			Code: "active",
		},
		Type:              lbType,
		AvailabilityZones: azs,
		SecurityGroups:    req.SecurityGroups,
		IPAddressType:     ipAddressType,
	}

	m.loadBalancers[arn] = lb

	return lb, nil
}

// DeleteLoadBalancer deletes a load balancer.
func (m *MemoryStorage) DeleteLoadBalancer(_ context.Context, loadBalancerArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.loadBalancers[loadBalancerArn]; !ok {
		return &Error{
			Code:    "LoadBalancerNotFound",
			Message: fmt.Sprintf("Load balancer '%s' not found", loadBalancerArn),
		}
	}

	// Delete associated listeners.
	for arn, listener := range m.listeners {
		if listener.LoadBalancerArn == loadBalancerArn {
			delete(m.listeners, arn)
		}
	}

	delete(m.loadBalancers, loadBalancerArn)

	return nil
}

// DescribeLoadBalancers describes load balancers.
func (m *MemoryStorage) DescribeLoadBalancers(_ context.Context, arns, names []string) ([]*LoadBalancer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*LoadBalancer, 0)

	if len(arns) == 0 && len(names) == 0 {
		// Return all load balancers.
		for _, lb := range m.loadBalancers {
			result = append(result, lb)
		}

		return result, nil
	}

	// Filter by ARNs.
	arnSet := make(map[string]bool)
	for _, arn := range arns {
		arnSet[arn] = true
	}

	// Filter by names.
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	for _, lb := range m.loadBalancers {
		if len(arns) > 0 && arnSet[lb.LoadBalancerArn] {
			result = append(result, lb)

			continue
		}

		if len(names) > 0 && nameSet[lb.LoadBalancerName] {
			result = append(result, lb)
		}
	}

	return result, nil
}

// CreateTargetGroup creates a new target group.
func (m *MemoryStorage) CreateTargetGroup(_ context.Context, req *CreateTargetGroupRequest) (*TargetGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate name.
	for _, tg := range m.targetGroups {
		if tg.TargetGroupName == req.Name {
			return nil, &Error{
				Code:    "DuplicateTargetGroupName",
				Message: fmt.Sprintf("A target group with the name '%s' already exists", req.Name),
			}
		}
	}

	targetType := req.TargetType
	if targetType == "" {
		targetType = "instance"
	}

	healthCheckPort := req.HealthCheckPort
	if healthCheckPort == "" {
		healthCheckPort = "traffic-port"
	}

	healthCheckProtocol := req.HealthCheckProtocol
	if healthCheckProtocol == "" {
		healthCheckProtocol = req.Protocol
		if healthCheckProtocol == "" {
			healthCheckProtocol = "HTTP"
		}
	}

	healthCheckPath := req.HealthCheckPath
	if healthCheckPath == "" && (healthCheckProtocol == "HTTP" || healthCheckProtocol == "HTTPS") {
		healthCheckPath = "/"
	}

	healthCheckInterval := req.HealthCheckIntervalSeconds
	if healthCheckInterval == 0 {
		healthCheckInterval = 30
	}

	healthCheckTimeout := req.HealthCheckTimeoutSeconds
	if healthCheckTimeout == 0 {
		healthCheckTimeout = 5
	}

	healthyThreshold := req.HealthyThresholdCount
	if healthyThreshold == 0 {
		healthyThreshold = 5
	}

	unhealthyThreshold := req.UnhealthyThresholdCount
	if unhealthyThreshold == 0 {
		unhealthyThreshold = 2
	}

	tgID := uuid.New().String()[:17]
	arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:targetgroup/%s/%s",
		defaultRegion, defaultAccountID, req.Name, tgID)

	tg := &TargetGroup{
		TargetGroupArn:             arn,
		TargetGroupName:            req.Name,
		Protocol:                   req.Protocol,
		Port:                       req.Port,
		VpcID:                      req.VpcID,
		HealthCheckEnabled:         true,
		HealthCheckIntervalSeconds: healthCheckInterval,
		HealthCheckPath:            healthCheckPath,
		HealthCheckPort:            healthCheckPort,
		HealthCheckProtocol:        healthCheckProtocol,
		HealthCheckTimeoutSeconds:  healthCheckTimeout,
		HealthyThresholdCount:      healthyThreshold,
		UnhealthyThresholdCount:    unhealthyThreshold,
		TargetType:                 targetType,
		LoadBalancerArns:           []string{},
	}

	m.targetGroups[arn] = tg
	m.targets[arn] = []Target{}

	return tg, nil
}

// DeleteTargetGroup deletes a target group.
func (m *MemoryStorage) DeleteTargetGroup(_ context.Context, targetGroupArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.targetGroups[targetGroupArn]; !ok {
		return &Error{
			Code:    "TargetGroupNotFound",
			Message: fmt.Sprintf("Target group '%s' not found", targetGroupArn),
		}
	}

	delete(m.targetGroups, targetGroupArn)
	delete(m.targets, targetGroupArn)

	return nil
}

// DescribeTargetGroups describes target groups.
func (m *MemoryStorage) DescribeTargetGroups(_ context.Context, arns, names []string, lbArn string) ([]*TargetGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*TargetGroup, 0)

	if len(arns) == 0 && len(names) == 0 && lbArn == "" {
		// Return all target groups.
		for _, tg := range m.targetGroups {
			result = append(result, tg)
		}

		return result, nil
	}

	// Filter by ARNs.
	arnSet := make(map[string]bool)
	for _, arn := range arns {
		arnSet[arn] = true
	}

	// Filter by names.
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	for _, tg := range m.targetGroups {
		if len(arns) > 0 && arnSet[tg.TargetGroupArn] {
			result = append(result, tg)

			continue
		}

		if len(names) > 0 && nameSet[tg.TargetGroupName] {
			result = append(result, tg)

			continue
		}

		if lbArn != "" && slices.Contains(tg.LoadBalancerArns, lbArn) {
			result = append(result, tg)
		}
	}

	return result, nil
}

// RegisterTargets registers targets with a target group.
func (m *MemoryStorage) RegisterTargets(_ context.Context, targetGroupArn string, targets []Target) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.targetGroups[targetGroupArn]; !ok {
		return &Error{
			Code:    "TargetGroupNotFound",
			Message: fmt.Sprintf("Target group '%s' not found", targetGroupArn),
		}
	}

	existingTargets := m.targets[targetGroupArn]
	existingSet := make(map[string]bool)

	for _, t := range existingTargets {
		existingSet[t.ID] = true
	}

	for _, t := range targets {
		if !existingSet[t.ID] {
			existingTargets = append(existingTargets, t)
		}
	}

	m.targets[targetGroupArn] = existingTargets

	return nil
}

// DeregisterTargets deregisters targets from a target group.
func (m *MemoryStorage) DeregisterTargets(_ context.Context, targetGroupArn string, targets []Target) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.targetGroups[targetGroupArn]; !ok {
		return &Error{
			Code:    "TargetGroupNotFound",
			Message: fmt.Sprintf("Target group '%s' not found", targetGroupArn),
		}
	}

	removeSet := make(map[string]bool)
	for _, t := range targets {
		removeSet[t.ID] = true
	}

	existingTargets := m.targets[targetGroupArn]
	newTargets := make([]Target, 0, len(existingTargets))

	for _, t := range existingTargets {
		if !removeSet[t.ID] {
			newTargets = append(newTargets, t)
		}
	}

	m.targets[targetGroupArn] = newTargets

	return nil
}

// CreateListener creates a new listener.
func (m *MemoryStorage) CreateListener(_ context.Context, req *CreateListenerRequest) (*Listener, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lb, ok := m.loadBalancers[req.LoadBalancerArn]
	if !ok {
		return nil, &Error{
			Code:    "LoadBalancerNotFound",
			Message: fmt.Sprintf("Load balancer '%s' not found", req.LoadBalancerArn),
		}
	}

	listenerID := uuid.New().String()[:17]

	// Parse load balancer ID from ARN for listener ARN.
	lbIDStart := len(req.LoadBalancerArn) - 17
	lbID := req.LoadBalancerArn[lbIDStart:]

	// Get load balancer type from the ARN.
	lbType := lb.Type[:3]

	arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:listener/%s/%s/%s/%s",
		defaultRegion, defaultAccountID, lbType, lb.LoadBalancerName, lbID, listenerID)

	listener := &Listener{
		ListenerArn:     arn,
		LoadBalancerArn: req.LoadBalancerArn,
		Port:            req.Port,
		Protocol:        req.Protocol,
		DefaultActions:  req.DefaultActions,
	}

	m.listeners[arn] = listener

	// Update target group's load balancer ARNs.
	for _, action := range req.DefaultActions {
		if action.TargetGroupArn != "" {
			if tg, exists := m.targetGroups[action.TargetGroupArn]; exists {
				if !slices.Contains(tg.LoadBalancerArns, req.LoadBalancerArn) {
					tg.LoadBalancerArns = append(tg.LoadBalancerArns, req.LoadBalancerArn)
				}
			}
		}
	}

	return listener, nil
}

// DeleteListener deletes a listener.
func (m *MemoryStorage) DeleteListener(_ context.Context, listenerArn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.listeners[listenerArn]; !ok {
		return &Error{
			Code:    "ListenerNotFound",
			Message: fmt.Sprintf("Listener '%s' not found", listenerArn),
		}
	}

	delete(m.listeners, listenerArn)

	return nil
}
