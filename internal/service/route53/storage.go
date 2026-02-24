package route53

import (
	"errors"
	"sync"
)

var (
	// ErrHostedZoneNotFound is returned when a hosted zone is not found.
	ErrHostedZoneNotFound = errors.New("hosted zone not found")
	// ErrHostedZoneAlreadyExists is returned when a hosted zone already exists.
	ErrHostedZoneAlreadyExists = errors.New("hosted zone already exists")
	// ErrHostedZoneNotEmpty is returned when trying to delete a non-empty hosted zone.
	ErrHostedZoneNotEmpty = errors.New("hosted zone is not empty")
	// ErrRecordSetNotFound is returned when a record set is not found.
	ErrRecordSetNotFound = errors.New("record set not found")
	// ErrRecordSetAlreadyExists is returned when a record set already exists.
	ErrRecordSetAlreadyExists = errors.New("record set already exists")
	// ErrInvalidInput is returned when input is invalid.
	ErrInvalidInput = errors.New("invalid input")
)

// Storage defines the interface for Route 53 storage operations.
type Storage interface {
	CreateHostedZone(zone *HostedZone) error
	GetHostedZone(id string) (*HostedZone, error)
	ListHostedZones() ([]*HostedZone, error)
	DeleteHostedZone(id string) error
	GetRecordSets(hostedZoneID string) ([]ResourceRecordSet, error)
	ChangeRecordSets(hostedZoneID string, changes []Change) error
}

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu          sync.RWMutex
	hostedZones map[string]*HostedZone
	recordSets  map[string][]ResourceRecordSet // key: hostedZoneID
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		hostedZones: make(map[string]*HostedZone),
		recordSets:  make(map[string][]ResourceRecordSet),
	}
}

// CreateHostedZone creates a new hosted zone.
func (s *MemoryStorage) CreateHostedZone(zone *HostedZone) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hostedZones[zone.ID]; exists {
		return ErrHostedZoneAlreadyExists
	}

	s.hostedZones[zone.ID] = zone
	s.recordSets[zone.ID] = []ResourceRecordSet{}

	return nil
}

// GetHostedZone retrieves a hosted zone by ID.
func (s *MemoryStorage) GetHostedZone(id string) (*HostedZone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zone, exists := s.hostedZones[id]
	if !exists {
		return nil, ErrHostedZoneNotFound
	}

	return zone, nil
}

// ListHostedZones lists all hosted zones.
func (s *MemoryStorage) ListHostedZones() ([]*HostedZone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zones := make([]*HostedZone, 0, len(s.hostedZones))
	for _, zone := range s.hostedZones {
		zones = append(zones, zone)
	}

	return zones, nil
}

// DeleteHostedZone deletes a hosted zone.
func (s *MemoryStorage) DeleteHostedZone(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hostedZones[id]; !exists {
		return ErrHostedZoneNotFound
	}

	// Check if hosted zone has record sets (other than NS and SOA)
	if records, ok := s.recordSets[id]; ok {
		for _, r := range records {
			if r.Type != "NS" && r.Type != "SOA" {
				return ErrHostedZoneNotEmpty
			}
		}
	}

	delete(s.hostedZones, id)
	delete(s.recordSets, id)

	return nil
}

// GetRecordSets retrieves all record sets for a hosted zone.
func (s *MemoryStorage) GetRecordSets(hostedZoneID string) ([]ResourceRecordSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.hostedZones[hostedZoneID]; !exists {
		return nil, ErrHostedZoneNotFound
	}

	records, ok := s.recordSets[hostedZoneID]
	if !ok {
		return []ResourceRecordSet{}, nil
	}

	return records, nil
}

// ChangeRecordSets applies changes to record sets.
func (s *MemoryStorage) ChangeRecordSets(hostedZoneID string, changes []Change) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hostedZones[hostedZoneID]; !exists {
		return ErrHostedZoneNotFound
	}

	records := s.recordSets[hostedZoneID]

	for _, change := range changes {
		switch change.Action {
		case "CREATE":
			if s.findRecordIndex(records, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type) >= 0 {
				return ErrRecordSetAlreadyExists
			}
			records = append(records, change.ResourceRecordSet)
		case "DELETE":
			idx := s.findRecordIndex(records, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type)
			if idx < 0 {
				return ErrRecordSetNotFound
			}
			records = append(records[:idx], records[idx+1:]...)
		case "UPSERT":
			idx := s.findRecordIndex(records, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type)
			if idx >= 0 {
				records[idx] = change.ResourceRecordSet
			} else {
				records = append(records, change.ResourceRecordSet)
			}
		default:
			return ErrInvalidInput
		}
	}

	s.recordSets[hostedZoneID] = records

	// Update record count
	if zone, exists := s.hostedZones[hostedZoneID]; exists {
		zone.ResourceRecordSetCount = int64(len(records))
	}

	return nil
}

// findRecordIndex finds the index of a record set by name and type.
func (s *MemoryStorage) findRecordIndex(records []ResourceRecordSet, name, recordType string) int {
	for i, r := range records {
		if r.Name == name && r.Type == recordType {
			return i
		}
	}
	return -1
}
