package route53resolver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	errResourceNotFound = "ResourceNotFoundException"
	errResourceExists   = "ResourceExistsException"
	errResourceInUse    = "ResourceInUseException"
	errInvalidParameter = "InvalidParameterException"

	statusDeleting = "DELETING"
)

// Storage is the interface for Route 53 Resolver storage operations.
type Storage interface {
	// Resolver Endpoints
	CreateResolverEndpoint(ctx context.Context, req *CreateResolverEndpointRequest) (*ResolverEndpoint, error)
	GetResolverEndpoint(ctx context.Context, id string) (*ResolverEndpoint, error)
	DeleteResolverEndpoint(ctx context.Context, id string) (*ResolverEndpoint, error)
	ListResolverEndpoints(ctx context.Context, maxResults int, nextToken string) ([]*ResolverEndpoint, string, error)

	// Resolver Rules
	CreateResolverRule(ctx context.Context, req *CreateResolverRuleRequest) (*ResolverRule, error)
	GetResolverRule(ctx context.Context, id string) (*ResolverRule, error)
	DeleteResolverRule(ctx context.Context, id string) (*ResolverRule, error)
	ListResolverRules(ctx context.Context, maxResults int, nextToken string) ([]*ResolverRule, string, error)

	// Resolver Rule Associations
	AssociateResolverRule(ctx context.Context, req *AssociateResolverRuleRequest) (*ResolverRuleAssociation, error)
	DisassociateResolverRule(ctx context.Context, ruleID, vpcID string) (*ResolverRuleAssociation, error)
	ListResolverRuleAssociations(ctx context.Context, maxResults int, nextToken string) ([]*ResolverRuleAssociation, string, error)
}

// MemoryStorage implements in-memory storage for Route 53 Resolver.
type MemoryStorage struct {
	mu           sync.RWMutex
	endpoints    map[string]*ResolverEndpoint
	rules        map[string]*ResolverRule
	associations map[string]*ResolverRuleAssociation
	accountID    string
	region       string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		endpoints:    make(map[string]*ResolverEndpoint),
		rules:        make(map[string]*ResolverRule),
		associations: make(map[string]*ResolverRuleAssociation),
		accountID:    "123456789012",
		region:       "us-east-1",
	}
}

// CreateResolverEndpoint creates a new resolver endpoint.
func (s *MemoryStorage) CreateResolverEndpoint(_ context.Context, req *CreateResolverEndpointRequest) (*ResolverEndpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := "rslvr-" + req.Direction[:2] + "-" + uuid.New().String()[:8]
	now := time.Now().UTC().Format(time.RFC3339)

	// Extract VPC ID from first subnet (simplified)
	vpcID := "vpc-" + uuid.New().String()[:8]

	ipAddresses := make([]*IPAddressResponse, 0, len(req.IPAddresses))

	for _, ipReq := range req.IPAddresses {
		ipAddr := &IPAddressResponse{
			IPAddressID:      "rslvr-ip-" + uuid.New().String()[:8],
			SubnetID:         ipReq.SubnetID,
			IP:               ipReq.IP,
			IPv6:             ipReq.IPv6,
			Status:           "ATTACHED",
			CreationTime:     now,
			ModificationTime: now,
		}

		if ipAddr.IP == "" {
			ipAddr.IP = fmt.Sprintf("10.0.%d.%d", len(ipAddresses), len(ipAddresses)+1)
		}

		ipAddresses = append(ipAddresses, ipAddr)
	}

	endpoint := &ResolverEndpoint{
		ID:                    id,
		CreatorRequestID:      req.CreatorRequestID,
		ARN:                   fmt.Sprintf("arn:aws:route53resolver:%s:%s:resolver-endpoint/%s", s.region, s.accountID, id),
		Name:                  req.Name,
		SecurityGroupIDs:      req.SecurityGroupIDs,
		Direction:             req.Direction,
		IPAddressCount:        len(ipAddresses),
		HostVPCID:             vpcID,
		Status:                "OPERATIONAL",
		CreationTime:          now,
		ModificationTime:      now,
		OutpostArn:            req.OutpostArn,
		PreferredInstanceType: req.PreferredInstanceType,
		ResolverEndpointType:  req.ResolverEndpointType,
		Protocols:             req.Protocols,
		IPAddresses:           ipAddresses,
	}

	if endpoint.ResolverEndpointType == "" {
		endpoint.ResolverEndpointType = "IPV4"
	}

	if len(endpoint.Protocols) == 0 {
		endpoint.Protocols = []string{"Do53"}
	}

	s.endpoints[id] = endpoint

	return endpoint, nil
}

// GetResolverEndpoint retrieves a resolver endpoint by ID.
func (s *MemoryStorage) GetResolverEndpoint(_ context.Context, id string) (*ResolverEndpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	endpoint, exists := s.endpoints[id]
	if !exists {
		return nil, &ResolverError{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Resolver endpoint with ID '%s' does not exist", id),
		}
	}

	return endpoint, nil
}

// DeleteResolverEndpoint deletes a resolver endpoint.
func (s *MemoryStorage) DeleteResolverEndpoint(_ context.Context, id string) (*ResolverEndpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	endpoint, exists := s.endpoints[id]
	if !exists {
		return nil, &ResolverError{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Resolver endpoint with ID '%s' does not exist", id),
		}
	}

	endpoint.Status = statusDeleting

	delete(s.endpoints, id)

	return endpoint, nil
}

// ListResolverEndpoints lists resolver endpoints.
func (s *MemoryStorage) ListResolverEndpoints(_ context.Context, maxResults int, _ string) ([]*ResolverEndpoint, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	endpoints := make([]*ResolverEndpoint, 0, len(s.endpoints))
	for _, endpoint := range s.endpoints {
		endpoints = append(endpoints, endpoint)
		if len(endpoints) >= maxResults {
			break
		}
	}

	return endpoints, "", nil
}

// CreateResolverRule creates a new resolver rule.
func (s *MemoryStorage) CreateResolverRule(_ context.Context, req *CreateResolverRuleRequest) (*ResolverRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := "rslvr-rr-" + uuid.New().String()[:8]
	now := time.Now().UTC().Format(time.RFC3339)

	targetIPs := make([]*TargetAddress, 0, len(req.TargetIPs))
	for _, t := range req.TargetIPs {
		targetIPs = append(targetIPs, &TargetAddress{
			IP:       t.IP,
			Port:     t.Port,
			IPv6:     t.IPv6,
			Protocol: t.Protocol,
		})
	}

	rule := &ResolverRule{
		ID:                 id,
		CreatorRequestID:   req.CreatorRequestID,
		ARN:                fmt.Sprintf("arn:aws:route53resolver:%s:%s:resolver-rule/%s", s.region, s.accountID, id),
		DomainName:         req.DomainName,
		Status:             "COMPLETE",
		RuleType:           req.RuleType,
		Name:               req.Name,
		TargetIPs:          targetIPs,
		ResolverEndpointID: req.ResolverEndpointID,
		OwnerID:            s.accountID,
		ShareStatus:        "NOT_SHARED",
		CreationTime:       now,
		ModificationTime:   now,
	}

	s.rules[id] = rule

	return rule, nil
}

// GetResolverRule retrieves a resolver rule by ID.
func (s *MemoryStorage) GetResolverRule(_ context.Context, id string) (*ResolverRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, exists := s.rules[id]
	if !exists {
		return nil, &ResolverError{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Resolver rule with ID '%s' does not exist", id),
		}
	}

	return rule, nil
}

// DeleteResolverRule deletes a resolver rule.
func (s *MemoryStorage) DeleteResolverRule(_ context.Context, id string) (*ResolverRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[id]
	if !exists {
		return nil, &ResolverError{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Resolver rule with ID '%s' does not exist", id),
		}
	}

	// Check if rule is associated with any VPC
	for _, assoc := range s.associations {
		if assoc.ResolverRuleID == id {
			return nil, &ResolverError{
				Code:    errResourceInUse,
				Message: fmt.Sprintf("Resolver rule '%s' is still associated with VPCs", id),
			}
		}
	}

	rule.Status = statusDeleting

	delete(s.rules, id)

	return rule, nil
}

// ListResolverRules lists resolver rules.
func (s *MemoryStorage) ListResolverRules(_ context.Context, maxResults int, _ string) ([]*ResolverRule, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	rules := make([]*ResolverRule, 0, len(s.rules))
	for _, rule := range s.rules {
		rules = append(rules, rule)
		if len(rules) >= maxResults {
			break
		}
	}

	return rules, "", nil
}

// AssociateResolverRule associates a resolver rule with a VPC.
func (s *MemoryStorage) AssociateResolverRule(_ context.Context, req *AssociateResolverRuleRequest) (*ResolverRuleAssociation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if rule exists
	if _, exists := s.rules[req.ResolverRuleID]; !exists {
		return nil, &ResolverError{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Resolver rule with ID '%s' does not exist", req.ResolverRuleID),
		}
	}

	// Check if association already exists
	for _, assoc := range s.associations {
		if assoc.ResolverRuleID == req.ResolverRuleID && assoc.VPCID == req.VPCID {
			return nil, &ResolverError{
				Code:    errResourceExists,
				Message: fmt.Sprintf("Resolver rule '%s' is already associated with VPC '%s'", req.ResolverRuleID, req.VPCID),
			}
		}
	}

	id := "rslvr-rrassoc-" + uuid.New().String()[:8]

	assoc := &ResolverRuleAssociation{
		ID:             id,
		ResolverRuleID: req.ResolverRuleID,
		Name:           req.Name,
		VPCID:          req.VPCID,
		Status:         "COMPLETE",
	}

	s.associations[id] = assoc

	return assoc, nil
}

// DisassociateResolverRule disassociates a resolver rule from a VPC.
func (s *MemoryStorage) DisassociateResolverRule(_ context.Context, ruleID, vpcID string) (*ResolverRuleAssociation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, assoc := range s.associations {
		if assoc.ResolverRuleID == ruleID && assoc.VPCID == vpcID {
			assoc.Status = statusDeleting

			delete(s.associations, id)

			return assoc, nil
		}
	}

	return nil, &ResolverError{
		Code:    errResourceNotFound,
		Message: fmt.Sprintf("Association between resolver rule '%s' and VPC '%s' does not exist", ruleID, vpcID),
	}
}

// ListResolverRuleAssociations lists resolver rule associations.
func (s *MemoryStorage) ListResolverRuleAssociations(_ context.Context, maxResults int, _ string) ([]*ResolverRuleAssociation, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 10
	}

	associations := make([]*ResolverRuleAssociation, 0, len(s.associations))
	for _, assoc := range s.associations {
		associations = append(associations, assoc)
		if len(associations) >= maxResults {
			break
		}
	}

	return associations, "", nil
}
