package ec2

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

const defaultAccountID = "000000000000"

// Instance state codes.
const (
	InstanceStatePending      = 0
	InstanceStateRunning      = 16
	InstanceStateShuttingDown = 32
	InstanceStateTerminated   = 48
	InstanceStateStopping     = 64
	InstanceStateStopped      = 80
)

// Instance state names.
const (
	InstanceStateNamePending      = "pending"
	InstanceStateNameRunning      = "running"
	InstanceStateNameShuttingDown = "shutting-down"
	InstanceStateNameTerminated   = "terminated"
	InstanceStateNameStopping     = "stopping"
	InstanceStateNameStopped      = "stopped"
)

// Storage defines the EC2 storage interface.
type Storage interface {
	RunInstances(ctx context.Context, req *RunInstancesRequest) ([]*Instance, string, error)
	TerminateInstances(ctx context.Context, instanceIDs []string) ([]InstanceStateChange, error)
	DescribeInstances(ctx context.Context, instanceIDs []string) ([]*Reservation, error)
	StartInstances(ctx context.Context, instanceIDs []string) ([]InstanceStateChange, error)
	StopInstances(ctx context.Context, instanceIDs []string) ([]InstanceStateChange, error)
	CreateSecurityGroup(ctx context.Context, req *CreateSecurityGroupRequest) (*SecurityGroup, error)
	DeleteSecurityGroup(ctx context.Context, groupID, groupName string) error
	AuthorizeSecurityGroupIngress(ctx context.Context, groupID, groupName string, permissions []IPPermission) error
	AuthorizeSecurityGroupEgress(ctx context.Context, groupID string, permissions []IPPermission) error
	CreateKeyPair(ctx context.Context, keyName, keyType string) (*KeyPair, error)
	DeleteKeyPair(ctx context.Context, keyName, keyPairID string) error
	DescribeKeyPairs(ctx context.Context, keyNames, keyPairIDs []string) ([]*KeyPair, error)
}

// InstanceStateChange represents an instance state change.
type InstanceStateChange struct {
	InstanceID    string
	CurrentState  InstanceState
	PreviousState InstanceState
}

// Reservation represents a group of instances launched together.
type Reservation struct {
	ReservationID string
	OwnerID       string
	Instances     []*Instance
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu             sync.RWMutex
	instances      map[string]*Instance
	reservations   map[string]*Reservation
	securityGroups map[string]*SecurityGroup
	keyPairs       map[string]*KeyPair
}

// NewMemoryStorage creates a new in-memory EC2 storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		instances:      make(map[string]*Instance),
		reservations:   make(map[string]*Reservation),
		securityGroups: make(map[string]*SecurityGroup),
		keyPairs:       make(map[string]*KeyPair),
	}
}

// RunInstances creates new EC2 instances.
func (m *MemoryStorage) RunInstances(_ context.Context, req *RunInstancesRequest) ([]*Instance, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := max(req.MaxCount, req.MinCount)

	if count <= 0 {
		count = 1
	}

	reservationID := "r-" + generateID()
	instances := make([]*Instance, 0, count)

	for i := 0; i < count; i++ {
		instance := &Instance{
			InstanceID:       "i-" + generateID(),
			ImageID:          req.ImageID,
			InstanceType:     req.InstanceType,
			State:            InstanceState{Code: InstanceStateRunning, Name: InstanceStateNameRunning},
			PrivateIPAddress: generatePrivateIP(),
			KeyName:          req.KeyName,
			LaunchTime:       time.Now(),
			SecurityGroups:   m.resolveSecurityGroups(req.SecurityGroupIDs, req.SecurityGroups),
		}
		instances = append(instances, instance)
		m.instances[instance.InstanceID] = instance
	}

	m.reservations[reservationID] = &Reservation{
		ReservationID: reservationID,
		OwnerID:       defaultAccountID,
		Instances:     instances,
	}

	return instances, reservationID, nil
}

// TerminateInstances terminates EC2 instances.
func (m *MemoryStorage) TerminateInstances(_ context.Context, instanceIDs []string) ([]InstanceStateChange, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var changes []InstanceStateChange

	for _, id := range instanceIDs {
		instance, exists := m.instances[id]
		if !exists {
			return nil, &Error{
				Code:    "InvalidInstanceID.NotFound",
				Message: fmt.Sprintf("The instance ID '%s' does not exist", id),
			}
		}

		prevState := instance.State
		instance.State = InstanceState{Code: InstanceStateTerminated, Name: InstanceStateNameTerminated}

		changes = append(changes, InstanceStateChange{
			InstanceID:    id,
			CurrentState:  instance.State,
			PreviousState: prevState,
		})
	}

	return changes, nil
}

// DescribeInstances describes EC2 instances.
func (m *MemoryStorage) DescribeInstances(_ context.Context, instanceIDs []string) ([]*Reservation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(instanceIDs) == 0 {
		reservations := make([]*Reservation, 0, len(m.reservations))
		for _, r := range m.reservations {
			reservations = append(reservations, r)
		}

		return reservations, nil
	}

	reservationMap := make(map[string]*Reservation)

	for _, id := range instanceIDs {
		instance, exists := m.instances[id]
		if !exists {
			return nil, &Error{
				Code:    "InvalidInstanceID.NotFound",
				Message: fmt.Sprintf("The instance ID '%s' does not exist", id),
			}
		}

		for resID, res := range m.reservations {
			for _, inst := range res.Instances {
				if inst.InstanceID == id {
					if _, ok := reservationMap[resID]; !ok {
						reservationMap[resID] = &Reservation{
							ReservationID: res.ReservationID,
							OwnerID:       res.OwnerID,
							Instances:     []*Instance{},
						}
					}

					reservationMap[resID].Instances = append(reservationMap[resID].Instances, instance)
				}
			}
		}
	}

	reservations := make([]*Reservation, 0, len(reservationMap))
	for _, r := range reservationMap {
		reservations = append(reservations, r)
	}

	return reservations, nil
}

// StartInstances starts stopped EC2 instances.
func (m *MemoryStorage) StartInstances(_ context.Context, instanceIDs []string) ([]InstanceStateChange, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var changes []InstanceStateChange

	for _, id := range instanceIDs {
		instance, exists := m.instances[id]
		if !exists {
			return nil, &Error{
				Code:    "InvalidInstanceID.NotFound",
				Message: fmt.Sprintf("The instance ID '%s' does not exist", id),
			}
		}

		prevState := instance.State
		instance.State = InstanceState{Code: InstanceStateRunning, Name: InstanceStateNameRunning}

		changes = append(changes, InstanceStateChange{
			InstanceID:    id,
			CurrentState:  instance.State,
			PreviousState: prevState,
		})
	}

	return changes, nil
}

// StopInstances stops running EC2 instances.
func (m *MemoryStorage) StopInstances(_ context.Context, instanceIDs []string) ([]InstanceStateChange, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var changes []InstanceStateChange

	for _, id := range instanceIDs {
		instance, exists := m.instances[id]
		if !exists {
			return nil, &Error{
				Code:    "InvalidInstanceID.NotFound",
				Message: fmt.Sprintf("The instance ID '%s' does not exist", id),
			}
		}

		prevState := instance.State
		instance.State = InstanceState{Code: InstanceStateStopped, Name: InstanceStateNameStopped}

		changes = append(changes, InstanceStateChange{
			InstanceID:    id,
			CurrentState:  instance.State,
			PreviousState: prevState,
		})
	}

	return changes, nil
}

// CreateSecurityGroup creates a new security group.
func (m *MemoryStorage) CreateSecurityGroup(_ context.Context, req *CreateSecurityGroupRequest) (*SecurityGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sg := range m.securityGroups {
		if sg.GroupName == req.GroupName {
			return nil, &Error{
				Code:    "InvalidGroup.Duplicate",
				Message: fmt.Sprintf("The security group '%s' already exists", req.GroupName),
			}
		}
	}

	sg := &SecurityGroup{
		GroupID:      "sg-" + generateID(),
		GroupName:    req.GroupName,
		Description:  req.GroupDescription,
		VpcID:        req.VpcID,
		IngressRules: []IPPermission{},
		EgressRules:  []IPPermission{},
	}

	m.securityGroups[sg.GroupID] = sg

	return sg, nil
}

// DeleteSecurityGroup deletes a security group.
func (m *MemoryStorage) DeleteSecurityGroup(_ context.Context, groupID, groupName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if groupID != "" {
		if _, exists := m.securityGroups[groupID]; !exists {
			return &Error{
				Code:    "InvalidGroup.NotFound",
				Message: fmt.Sprintf("The security group '%s' does not exist", groupID),
			}
		}

		delete(m.securityGroups, groupID)

		return nil
	}

	for id, sg := range m.securityGroups {
		if sg.GroupName == groupName {
			delete(m.securityGroups, id)

			return nil
		}
	}

	return &Error{
		Code:    "InvalidGroup.NotFound",
		Message: fmt.Sprintf("The security group '%s' does not exist", groupName),
	}
}

// AuthorizeSecurityGroupIngress adds ingress rules to a security group.
func (m *MemoryStorage) AuthorizeSecurityGroupIngress(_ context.Context, groupID, groupName string, permissions []IPPermission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg := m.findSecurityGroup(groupID, groupName)
	if sg == nil {
		return &Error{
			Code:    "InvalidGroup.NotFound",
			Message: "The security group does not exist",
		}
	}

	sg.IngressRules = append(sg.IngressRules, permissions...)

	return nil
}

// AuthorizeSecurityGroupEgress adds egress rules to a security group.
func (m *MemoryStorage) AuthorizeSecurityGroupEgress(_ context.Context, groupID string, permissions []IPPermission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg, exists := m.securityGroups[groupID]
	if !exists {
		return &Error{
			Code:    "InvalidGroup.NotFound",
			Message: fmt.Sprintf("The security group '%s' does not exist", groupID),
		}
	}

	sg.EgressRules = append(sg.EgressRules, permissions...)

	return nil
}

// CreateKeyPair creates a new key pair.
func (m *MemoryStorage) CreateKeyPair(_ context.Context, keyName, _ string) (*KeyPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, kp := range m.keyPairs {
		if kp.KeyName == keyName {
			return nil, &Error{
				Code:    "InvalidKeyPair.Duplicate",
				Message: fmt.Sprintf("The keypair '%s' already exists", keyName),
			}
		}
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, &Error{
			Code:    "InternalError",
			Message: "Failed to generate key pair",
		}
	}

	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, &Error{
			Code:    "InternalError",
			Message: "Failed to generate public key",
		}
	}

	fingerprint := generateFingerprint(pubKey.Marshal())

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	kp := &KeyPair{
		KeyName:        keyName,
		KeyFingerprint: fingerprint,
		KeyPairID:      "key-" + generateID(),
		KeyMaterial:    string(privateKeyPEM),
		CreateTime:     time.Now(),
	}

	m.keyPairs[kp.KeyPairID] = kp

	return kp, nil
}

// DeleteKeyPair deletes a key pair.
func (m *MemoryStorage) DeleteKeyPair(_ context.Context, keyName, keyPairID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if keyPairID != "" {
		if _, exists := m.keyPairs[keyPairID]; !exists {
			return &Error{
				Code:    "InvalidKeyPair.NotFound",
				Message: fmt.Sprintf("The key pair '%s' does not exist", keyPairID),
			}
		}

		delete(m.keyPairs, keyPairID)

		return nil
	}

	for id, kp := range m.keyPairs {
		if kp.KeyName == keyName {
			delete(m.keyPairs, id)

			return nil
		}
	}

	return &Error{
		Code:    "InvalidKeyPair.NotFound",
		Message: fmt.Sprintf("The key pair '%s' does not exist", keyName),
	}
}

// DescribeKeyPairs describes key pairs.
func (m *MemoryStorage) DescribeKeyPairs(_ context.Context, keyNames, keyPairIDs []string) ([]*KeyPair, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(keyNames) == 0 && len(keyPairIDs) == 0 {
		return m.listAllKeyPairs(), nil
	}

	keyPairSet, err := m.collectKeyPairs(keyNames, keyPairIDs)
	if err != nil {
		return nil, err
	}

	return m.keyPairSetToSlice(keyPairSet), nil
}

// listAllKeyPairs returns all key pairs without KeyMaterial.
func (m *MemoryStorage) listAllKeyPairs() []*KeyPair {
	keyPairs := make([]*KeyPair, 0, len(m.keyPairs))

	for _, kp := range m.keyPairs {
		keyPairs = append(keyPairs, copyKeyPairInfo(kp))
	}

	return keyPairs
}

// collectKeyPairs collects key pairs by names and IDs.
func (m *MemoryStorage) collectKeyPairs(keyNames, keyPairIDs []string) (map[string]*KeyPair, error) {
	keyPairSet := make(map[string]*KeyPair)

	for _, name := range keyNames {
		kp, err := m.findKeyPairByName(name)
		if err != nil {
			return nil, err
		}

		keyPairSet[kp.KeyPairID] = kp
	}

	for _, id := range keyPairIDs {
		kp, exists := m.keyPairs[id]
		if !exists {
			return nil, &Error{
				Code:    "InvalidKeyPair.NotFound",
				Message: fmt.Sprintf("The key pair '%s' does not exist", id),
			}
		}

		keyPairSet[kp.KeyPairID] = kp
	}

	return keyPairSet, nil
}

// findKeyPairByName finds a key pair by name.
func (m *MemoryStorage) findKeyPairByName(name string) (*KeyPair, error) {
	for _, kp := range m.keyPairs {
		if kp.KeyName == name {
			return kp, nil
		}
	}

	return nil, &Error{
		Code:    "InvalidKeyPair.NotFound",
		Message: fmt.Sprintf("The key pair '%s' does not exist", name),
	}
}

// keyPairSetToSlice converts a key pair set to a slice without KeyMaterial.
func (m *MemoryStorage) keyPairSetToSlice(keyPairSet map[string]*KeyPair) []*KeyPair {
	keyPairs := make([]*KeyPair, 0, len(keyPairSet))

	for _, kp := range keyPairSet {
		keyPairs = append(keyPairs, copyKeyPairInfo(kp))
	}

	return keyPairs
}

// copyKeyPairInfo copies key pair info without KeyMaterial.
func copyKeyPairInfo(kp *KeyPair) *KeyPair {
	return &KeyPair{
		KeyName:        kp.KeyName,
		KeyFingerprint: kp.KeyFingerprint,
		KeyPairID:      kp.KeyPairID,
		CreateTime:     kp.CreateTime,
	}
}

// findSecurityGroup finds a security group by ID or name.
func (m *MemoryStorage) findSecurityGroup(groupID, groupName string) *SecurityGroup {
	if groupID != "" {
		return m.securityGroups[groupID]
	}

	for _, sg := range m.securityGroups {
		if sg.GroupName == groupName {
			return sg
		}
	}

	return nil
}

// resolveSecurityGroups resolves security group IDs and names to GroupIdentifiers.
func (m *MemoryStorage) resolveSecurityGroups(groupIDs, groupNames []string) []GroupIdentifier {
	var result []GroupIdentifier

	for _, id := range groupIDs {
		if sg, exists := m.securityGroups[id]; exists {
			result = append(result, GroupIdentifier{
				GroupID:   sg.GroupID,
				GroupName: sg.GroupName,
			})
		}
	}

	for _, name := range groupNames {
		for _, sg := range m.securityGroups {
			if sg.GroupName == name {
				result = append(result, GroupIdentifier{
					GroupID:   sg.GroupID,
					GroupName: sg.GroupName,
				})

				break
			}
		}
	}

	return result
}

// generateID generates a random ID.
func generateID() string {
	return uuid.New().String()[:17]
}

// generatePrivateIP generates a random private IP address.
func generatePrivateIP() string {
	return fmt.Sprintf("10.0.%d.%d", randByte(), randByte())
}

// randByte returns a random byte value.
func randByte() int {
	b := make([]byte, 1)
	_, _ = rand.Read(b)

	return int(b[0])
}

// generateFingerprint generates a fingerprint from a public key.
func generateFingerprint(pubKey []byte) string {
	hash := sha256.Sum256(pubKey)

	return hex.EncodeToString(hash[:])
}
