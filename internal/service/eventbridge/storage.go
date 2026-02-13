package eventbridge

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Default event bus name.
const defaultEventBusName = "default"

// Error codes.
const (
	errEventBusNotFound      = "ResourceNotFoundException"
	errEventBusAlreadyExists = "ResourceAlreadyExistsException"
	errRuleNotFound          = "ResourceNotFoundException"
	errInvalidParameter      = "ValidationException"
)

// Storage defines the EventBridge storage interface.
type Storage interface {
	// Event Bus operations.
	CreateEventBus(ctx context.Context, req *CreateEventBusRequest) (*EventBus, error)
	DeleteEventBus(ctx context.Context, name string) error
	DescribeEventBus(ctx context.Context, name string) (*EventBus, error)
	ListEventBuses(ctx context.Context, namePrefix string, limit int32, nextToken string) ([]*EventBus, string, error)

	// Rule operations.
	PutRule(ctx context.Context, req *PutRuleRequest) (*Rule, error)
	DeleteRule(ctx context.Context, eventBusName, ruleName string, force bool) error
	DescribeRule(ctx context.Context, eventBusName, ruleName string) (*Rule, error)
	ListRules(ctx context.Context, eventBusName, namePrefix string, limit int32, nextToken string) ([]*Rule, string, error)

	// Target operations.
	PutTargets(ctx context.Context, eventBusName, ruleName string, targets []TargetInput) ([]PutTargetsResultEntry, error)
	RemoveTargets(ctx context.Context, eventBusName, ruleName string, ids []string, force bool) ([]RemoveTargetsResultEntry, error)
	ListTargetsByRule(ctx context.Context, eventBusName, ruleName string, limit int32, nextToken string) ([]*Target, string, error)

	// Event operations.
	PutEvents(ctx context.Context, entries []PutEventsRequestEntry) ([]PutEventsResultEntry, error)

	// DispatchAction dispatches the request to the appropriate handler.
	DispatchAction(action string) bool
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu         sync.RWMutex
	eventBuses map[string]*EventBus
	rules      map[string]map[string]*Rule     // eventBusName -> ruleName -> Rule
	targets    map[string]map[string][]*Target // eventBusName:ruleName -> targets
	region     string
	accountID  string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		eventBuses: make(map[string]*EventBus),
		rules:      make(map[string]map[string]*Rule),
		targets:    make(map[string]map[string][]*Target),
		region:     "us-east-1",
		accountID:  "000000000000",
	}

	// Create default event bus.
	now := time.Now()
	storage.eventBuses[defaultEventBusName] = &EventBus{
		Name:         defaultEventBusName,
		Arn:          fmt.Sprintf("arn:aws:events:%s:%s:event-bus/%s", storage.region, storage.accountID, defaultEventBusName),
		CreationTime: now,
		LastModified: now,
	}

	storage.rules[defaultEventBusName] = make(map[string]*Rule)

	return storage
}

// CreateEventBus creates a new event bus.
func (s *MemoryStorage) CreateEventBus(_ context.Context, req *CreateEventBusRequest) (*EventBus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.eventBuses[req.Name]; exists {
		return nil, &ServiceError{Code: errEventBusAlreadyExists, Message: "Event bus already exists"}
	}

	now := time.Now()
	eventBus := &EventBus{
		Name:         req.Name,
		Arn:          fmt.Sprintf("arn:aws:events:%s:%s:event-bus/%s", s.region, s.accountID, req.Name),
		Description:  req.Description,
		CreationTime: now,
		LastModified: now,
	}

	s.eventBuses[req.Name] = eventBus
	s.rules[req.Name] = make(map[string]*Rule)

	return eventBus, nil
}

// DeleteEventBus deletes an event bus.
func (s *MemoryStorage) DeleteEventBus(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == defaultEventBusName {
		return &ServiceError{Code: errInvalidParameter, Message: "Cannot delete the default event bus"}
	}

	if _, exists := s.eventBuses[name]; !exists {
		return &ServiceError{Code: errEventBusNotFound, Message: "Event bus not found"}
	}

	delete(s.eventBuses, name)
	delete(s.rules, name)

	// Delete all targets for rules on this event bus.
	for key := range s.targets {
		if strings.HasPrefix(key, name+":") {
			delete(s.targets, key)
		}
	}

	return nil
}

// DescribeEventBus describes an event bus.
func (s *MemoryStorage) DescribeEventBus(_ context.Context, name string) (*EventBus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if name == "" {
		name = defaultEventBusName
	}

	eventBus, exists := s.eventBuses[name]
	if !exists {
		return nil, &ServiceError{Code: errEventBusNotFound, Message: "Event bus not found"}
	}

	return eventBus, nil
}

// ListEventBuses lists event buses.
func (s *MemoryStorage) ListEventBuses(_ context.Context, namePrefix string, limit int32, _ string) ([]*EventBus, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	var eventBuses []*EventBus

	for _, eb := range s.eventBuses {
		if namePrefix == "" || strings.HasPrefix(eb.Name, namePrefix) {
			eventBuses = append(eventBuses, eb)
		}

		if int32(len(eventBuses)) >= limit { //nolint:gosec // slice length bounded by limit parameter
			break
		}
	}

	return eventBuses, "", nil
}

// PutRule creates or updates a rule.
func (s *MemoryStorage) PutRule(_ context.Context, req *PutRuleRequest) (*Rule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventBusName := req.EventBusName
	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	if _, exists := s.eventBuses[eventBusName]; !exists {
		return nil, &ServiceError{Code: errEventBusNotFound, Message: "Event bus not found"}
	}

	now := time.Now()
	state := RuleStateEnabled

	if req.State == "DISABLED" {
		state = RuleStateDisabled
	}

	rule := &Rule{
		Name:               req.Name,
		Arn:                fmt.Sprintf("arn:aws:events:%s:%s:rule/%s/%s", s.region, s.accountID, eventBusName, req.Name),
		EventBusName:       eventBusName,
		EventPattern:       req.EventPattern,
		ScheduleExpression: req.ScheduleExpression,
		State:              state,
		Description:        req.Description,
		RoleArn:            req.RoleArn,
		CreationTime:       now,
		LastModified:       now,
	}

	if s.rules[eventBusName] == nil {
		s.rules[eventBusName] = make(map[string]*Rule)
	}

	s.rules[eventBusName][req.Name] = rule

	return rule, nil
}

// DeleteRule deletes a rule.
func (s *MemoryStorage) DeleteRule(_ context.Context, eventBusName, ruleName string, _ bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	rules, exists := s.rules[eventBusName]
	if !exists {
		return &ServiceError{Code: errRuleNotFound, Message: "Rule not found"}
	}

	if _, exists := rules[ruleName]; !exists {
		return &ServiceError{Code: errRuleNotFound, Message: "Rule not found"}
	}

	delete(rules, ruleName)

	// Delete targets for this rule.
	targetKey := eventBusName + ":" + ruleName
	delete(s.targets, targetKey)

	return nil
}

// DescribeRule describes a rule.
func (s *MemoryStorage) DescribeRule(_ context.Context, eventBusName, ruleName string) (*Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	rules, exists := s.rules[eventBusName]
	if !exists {
		return nil, &ServiceError{Code: errRuleNotFound, Message: "Rule not found"}
	}

	rule, exists := rules[ruleName]
	if !exists {
		return nil, &ServiceError{Code: errRuleNotFound, Message: "Rule not found"}
	}

	return rule, nil
}

// ListRules lists rules for an event bus.
func (s *MemoryStorage) ListRules(_ context.Context, eventBusName, namePrefix string, limit int32, _ string) ([]*Rule, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	if limit <= 0 {
		limit = 10
	}

	rules, exists := s.rules[eventBusName]
	if !exists {
		return nil, "", nil
	}

	var result []*Rule

	for _, rule := range rules {
		if namePrefix == "" || strings.HasPrefix(rule.Name, namePrefix) {
			result = append(result, rule)
		}

		if int32(len(result)) >= limit { //nolint:gosec // slice length bounded by limit parameter
			break
		}
	}

	return result, "", nil
}

// PutTargets adds targets to a rule.
func (s *MemoryStorage) PutTargets(_ context.Context, eventBusName, ruleName string, targets []TargetInput) ([]PutTargetsResultEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	rules, exists := s.rules[eventBusName]
	if !exists {
		return nil, &ServiceError{Code: errRuleNotFound, Message: "Rule not found"}
	}

	if _, exists := rules[ruleName]; !exists {
		return nil, &ServiceError{Code: errRuleNotFound, Message: "Rule not found"}
	}

	targetKey := eventBusName + ":" + ruleName

	if s.targets[targetKey] == nil {
		s.targets[targetKey] = make(map[string][]*Target)
	}

	var failedEntries []PutTargetsResultEntry

	for _, t := range targets {
		target := &Target{
			ID:        t.ID,
			Arn:       t.Arn,
			RoleArn:   t.RoleArn,
			Input:     t.Input,
			InputPath: t.InputPath,
		}

		// Find and update existing target or add new one.
		found := false
		existingTargets := s.targets[targetKey][ruleName]

		for i, existing := range existingTargets {
			if existing.ID == t.ID {
				existingTargets[i] = target
				found = true

				break
			}
		}

		if !found {
			s.targets[targetKey][ruleName] = append(s.targets[targetKey][ruleName], target)
		}
	}

	return failedEntries, nil
}

// RemoveTargets removes targets from a rule.
func (s *MemoryStorage) RemoveTargets(_ context.Context, eventBusName, ruleName string, ids []string, _ bool) ([]RemoveTargetsResultEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	targetKey := eventBusName + ":" + ruleName

	var failedEntries []RemoveTargetsResultEntry

	if s.targets[targetKey] == nil {
		return failedEntries, nil
	}

	existingTargets := s.targets[targetKey][ruleName]

	var newTargets []*Target

	idsToRemove := make(map[string]bool)
	for _, id := range ids {
		idsToRemove[id] = true
	}

	for _, target := range existingTargets {
		if !idsToRemove[target.ID] {
			newTargets = append(newTargets, target)
		}
	}

	s.targets[targetKey][ruleName] = newTargets

	return failedEntries, nil
}

// ListTargetsByRule lists targets for a rule.
func (s *MemoryStorage) ListTargetsByRule(_ context.Context, eventBusName, ruleName string, limit int32, _ string) ([]*Target, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if eventBusName == "" {
		eventBusName = defaultEventBusName
	}

	if limit <= 0 {
		limit = 100
	}

	targetKey := eventBusName + ":" + ruleName

	if s.targets[targetKey] == nil {
		return nil, "", nil
	}

	targets := s.targets[targetKey][ruleName]

	if int32(len(targets)) > limit { //nolint:gosec // slice length bounded by limit parameter
		targets = targets[:limit]
	}

	return targets, "", nil
}

// PutEvents puts events to the event bus.
func (s *MemoryStorage) PutEvents(_ context.Context, entries []PutEventsRequestEntry) ([]PutEventsResultEntry, error) {
	results := make([]PutEventsResultEntry, len(entries))

	for i := range entries {
		// Generate event ID for successful entries.
		results[i] = PutEventsResultEntry{
			EventID: uuid.New().String(),
		}
	}

	return results, nil
}

// DispatchAction checks if the action is valid.
func (s *MemoryStorage) DispatchAction(_ string) bool {
	return true
}
