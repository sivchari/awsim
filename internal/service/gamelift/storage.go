package gamelift

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errNotFoundException        = "NotFoundException"
	errInvalidRequestException  = "InvalidRequestException"
	errConflictException        = "ConflictException"
	errInternalServiceException = "InternalServiceException"
	errLimitExceededException   = "LimitExceededException"
)

// Default values.
const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "123456789012"
	defaultIPAddress = "10.0.0.1"
	defaultPort      = int32(7777)
)

// Build status values.
const (
	buildStatusInitialized = "INITIALIZED"
	buildStatusReady       = "READY"
)

// Fleet status values.
const (
	fleetStatusNew    = "NEW"
	fleetStatusActive = "ACTIVE"
)

// Game session status values.
const (
	gameSessionStatusActive     = "ACTIVE"
	gameSessionStatusActivating = "ACTIVATING"
)

// Player session status values.
const (
	playerSessionStatusReserved = "RESERVED"
	playerSessionStatusActive   = "ACTIVE"
)

// Storage defines the GameLift service storage interface.
type Storage interface {
	// Build operations
	CreateBuild(ctx context.Context, req *CreateBuildRequest) (*Build, error)
	DescribeBuild(ctx context.Context, buildID string) (*Build, error)
	ListBuilds(ctx context.Context, status string, limit int32) ([]*Build, error)
	DeleteBuild(ctx context.Context, buildID string) error

	// Fleet operations
	CreateFleet(ctx context.Context, req *CreateFleetRequest) (*Fleet, error)
	DescribeFleetAttributes(ctx context.Context, fleetIDs []string) ([]*Fleet, error)
	ListFleets(ctx context.Context, buildID string, limit int32) ([]string, error)
	DeleteFleet(ctx context.Context, fleetID string) error

	// Game session operations
	CreateGameSession(ctx context.Context, req *CreateGameSessionRequest) (*GameSession, error)
	DescribeGameSessions(ctx context.Context, fleetID, gameSessionID string) ([]*GameSession, error)
	UpdateGameSession(ctx context.Context, req *UpdateGameSessionRequest) (*GameSession, error)

	// Player session operations
	CreatePlayerSession(ctx context.Context, gameSessionID, playerID string) (*PlayerSession, error)
	CreatePlayerSessions(ctx context.Context, gameSessionID string, playerIDs []string) ([]*PlayerSession, error)
	DescribePlayerSessions(ctx context.Context, gameSessionID, playerSessionID, playerID string) ([]*PlayerSession, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu             sync.RWMutex
	builds         map[string]*Build
	fleets         map[string]*Fleet
	gameSessions   map[string]*GameSession
	playerSessions map[string]*PlayerSession
	region         string
	accountID      string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		builds:         make(map[string]*Build),
		fleets:         make(map[string]*Fleet),
		gameSessions:   make(map[string]*GameSession),
		playerSessions: make(map[string]*PlayerSession),
		region:         defaultRegion,
		accountID:      defaultAccountID,
	}
}

// CreateBuild creates a new build.
func (m *MemoryStorage) CreateBuild(_ context.Context, req *CreateBuildRequest) (*Build, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	buildID := "build-" + uuid.New().String()[:8]
	buildARN := generateBuildARN(m.region, m.accountID, buildID)

	build := &Build{
		BuildID:         buildID,
		BuildARN:        buildARN,
		Name:            req.Name,
		Version:         req.Version,
		Status:          buildStatusInitialized,
		SizeOnDisk:      0,
		OperatingSystem: defaultString(req.OperatingSystem, "AMAZON_LINUX_2"),
		CreationTime:    time.Now(),
	}

	m.builds[buildID] = build

	return build, nil
}

// DescribeBuild describes a build.
func (m *MemoryStorage) DescribeBuild(_ context.Context, buildID string) (*Build, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	build, exists := m.builds[buildID]
	if !exists {
		return nil, &Error{Code: errNotFoundException, Message: "Build not found: " + buildID}
	}

	return build, nil
}

// ListBuilds lists builds.
func (m *MemoryStorage) ListBuilds(_ context.Context, status string, limit int32) ([]*Build, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Build, 0, len(m.builds))

	for _, build := range m.builds {
		if status != "" && build.Status != status {
			continue
		}

		result = append(result, build)

		//nolint:gosec // len(result) is bounded by the number of builds which is limited.
		if limit > 0 && int32(len(result)) >= limit {
			break
		}
	}

	return result, nil
}

// DeleteBuild deletes a build.
func (m *MemoryStorage) DeleteBuild(_ context.Context, buildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.builds[buildID]; !exists {
		return &Error{Code: errNotFoundException, Message: "Build not found: " + buildID}
	}

	delete(m.builds, buildID)

	return nil
}

// CreateFleet creates a new fleet.
func (m *MemoryStorage) CreateFleet(_ context.Context, req *CreateFleetRequest) (*Fleet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Name == "" {
		return nil, &Error{Code: errInvalidRequestException, Message: "Name is required"}
	}

	fleetID := "fleet-" + uuid.New().String()[:8]
	fleetARN := generateFleetARN(m.region, m.accountID, fleetID)

	fleet := &Fleet{
		FleetID:                        fleetID,
		FleetARN:                       fleetARN,
		Name:                           req.Name,
		Description:                    req.Description,
		BuildID:                        req.BuildID,
		Status:                         fleetStatusActive,
		FleetType:                      defaultString(req.FleetType, "ON_DEMAND"),
		InstanceType:                   defaultString(req.EC2InstanceType, "c5.large"),
		ServerLaunchPath:               req.ServerLaunchPath,
		NewGameSessionProtectionPolicy: defaultString(req.NewGameSessionProtectionPolicy, "NoProtection"),
		CreationTime:                   time.Now(),
	}

	m.fleets[fleetID] = fleet

	return fleet, nil
}

// DescribeFleetAttributes describes fleet attributes.
func (m *MemoryStorage) DescribeFleetAttributes(_ context.Context, fleetIDs []string) ([]*Fleet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(fleetIDs) == 0 {
		// Return all fleets
		result := make([]*Fleet, 0, len(m.fleets))
		for _, fleet := range m.fleets {
			result = append(result, fleet)
		}

		return result, nil
	}

	// Return specified fleets
	result := make([]*Fleet, 0, len(fleetIDs))

	for _, fleetID := range fleetIDs {
		if fleet, exists := m.fleets[fleetID]; exists {
			result = append(result, fleet)
		}
	}

	return result, nil
}

// ListFleets lists fleet IDs.
func (m *MemoryStorage) ListFleets(_ context.Context, buildID string, limit int32) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, 0, len(m.fleets))

	for _, fleet := range m.fleets {
		if buildID != "" && fleet.BuildID != buildID {
			continue
		}

		result = append(result, fleet.FleetID)

		//nolint:gosec // len(result) is bounded by the number of fleets which is limited.
		if limit > 0 && int32(len(result)) >= limit {
			break
		}
	}

	return result, nil
}

// DeleteFleet deletes a fleet.
func (m *MemoryStorage) DeleteFleet(_ context.Context, fleetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.fleets[fleetID]; !exists {
		return &Error{Code: errNotFoundException, Message: "Fleet not found: " + fleetID}
	}

	delete(m.fleets, fleetID)

	return nil
}

// CreateGameSession creates a new game session.
func (m *MemoryStorage) CreateGameSession(_ context.Context, req *CreateGameSessionRequest) (*GameSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.FleetID == "" {
		return nil, &Error{Code: errInvalidRequestException, Message: "FleetId is required"}
	}

	fleet, exists := m.fleets[req.FleetID]
	if !exists {
		return nil, &Error{Code: errNotFoundException, Message: "Fleet not found: " + req.FleetID}
	}

	gameSessionID := "gsess-" + uuid.New().String()[:8]
	gameSessionARN := generateGameSessionARN(m.region, m.accountID, fleet.FleetID, gameSessionID)

	gameSession := &GameSession{
		GameSessionID:             gameSessionID,
		GameSessionARN:            gameSessionARN,
		FleetID:                   fleet.FleetID,
		FleetARN:                  fleet.FleetARN,
		Name:                      req.Name,
		Status:                    gameSessionStatusActive,
		CurrentPlayerSessionCount: 0,
		MaximumPlayerSessionCount: req.MaximumPlayerSessionCount,
		IPAddress:                 defaultIPAddress,
		Port:                      defaultPort,
		CreationTime:              time.Now(),
	}

	m.gameSessions[gameSessionID] = gameSession

	return gameSession, nil
}

// DescribeGameSessions describes game sessions.
func (m *MemoryStorage) DescribeGameSessions(_ context.Context, fleetID, gameSessionID string) ([]*GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if gameSessionID != "" {
		session, exists := m.gameSessions[gameSessionID]
		if !exists {
			return []*GameSession{}, nil
		}

		return []*GameSession{session}, nil
	}

	result := make([]*GameSession, 0, len(m.gameSessions))

	for _, session := range m.gameSessions {
		if fleetID != "" && session.FleetID != fleetID {
			continue
		}

		result = append(result, session)
	}

	return result, nil
}

// UpdateGameSession updates a game session.
func (m *MemoryStorage) UpdateGameSession(_ context.Context, req *UpdateGameSessionRequest) (*GameSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.gameSessions[req.GameSessionID]
	if !exists {
		return nil, &Error{Code: errNotFoundException, Message: "Game session not found: " + req.GameSessionID}
	}

	if req.MaximumPlayerSessionCount != nil {
		session.MaximumPlayerSessionCount = *req.MaximumPlayerSessionCount
	}

	if req.Name != "" {
		session.Name = req.Name
	}

	return session, nil
}

// CreatePlayerSession creates a new player session.
func (m *MemoryStorage) CreatePlayerSession(_ context.Context, gameSessionID, playerID string) (*PlayerSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	gameSession, exists := m.gameSessions[gameSessionID]
	if !exists {
		return nil, &Error{Code: errNotFoundException, Message: "Game session not found: " + gameSessionID}
	}

	if gameSession.CurrentPlayerSessionCount >= gameSession.MaximumPlayerSessionCount {
		return nil, &Error{Code: errInvalidRequestException, Message: "Game session is full"}
	}

	playerSessionID := "psess-" + uuid.New().String()[:8]

	playerSession := &PlayerSession{
		PlayerSessionID: playerSessionID,
		GameSessionID:   gameSessionID,
		FleetID:         gameSession.FleetID,
		FleetARN:        gameSession.FleetARN,
		PlayerID:        playerID,
		Status:          playerSessionStatusReserved,
		IPAddress:       gameSession.IPAddress,
		Port:            gameSession.Port,
		CreationTime:    time.Now(),
	}

	m.playerSessions[playerSessionID] = playerSession
	gameSession.CurrentPlayerSessionCount++

	return playerSession, nil
}

// CreatePlayerSessions creates multiple player sessions.
func (m *MemoryStorage) CreatePlayerSessions(_ context.Context, gameSessionID string, playerIDs []string) ([]*PlayerSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	gameSession, exists := m.gameSessions[gameSessionID]
	if !exists {
		return nil, &Error{Code: errNotFoundException, Message: "Game session not found: " + gameSessionID}
	}

	//nolint:gosec // len(playerIDs) is bounded by the request, which is limited by AWS SDK.
	if gameSession.CurrentPlayerSessionCount+int32(len(playerIDs)) > gameSession.MaximumPlayerSessionCount {
		return nil, &Error{Code: errInvalidRequestException, Message: "Not enough capacity for all players"}
	}

	result := make([]*PlayerSession, 0, len(playerIDs))

	for _, playerID := range playerIDs {
		playerSessionID := "psess-" + uuid.New().String()[:8]

		playerSession := &PlayerSession{
			PlayerSessionID: playerSessionID,
			GameSessionID:   gameSessionID,
			FleetID:         gameSession.FleetID,
			FleetARN:        gameSession.FleetARN,
			PlayerID:        playerID,
			Status:          playerSessionStatusReserved,
			IPAddress:       gameSession.IPAddress,
			Port:            gameSession.Port,
			CreationTime:    time.Now(),
		}

		m.playerSessions[playerSessionID] = playerSession
		result = append(result, playerSession)
	}

	//nolint:gosec // len(playerIDs) is bounded by the request, which is limited by AWS SDK.
	gameSession.CurrentPlayerSessionCount += int32(len(playerIDs))

	return result, nil
}

// DescribePlayerSessions describes player sessions.
func (m *MemoryStorage) DescribePlayerSessions(_ context.Context, gameSessionID, playerSessionID, playerID string) ([]*PlayerSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if playerSessionID != "" {
		session, exists := m.playerSessions[playerSessionID]
		if !exists {
			return []*PlayerSession{}, nil
		}

		return []*PlayerSession{session}, nil
	}

	result := make([]*PlayerSession, 0, len(m.playerSessions))

	for _, session := range m.playerSessions {
		if gameSessionID != "" && session.GameSessionID != gameSessionID {
			continue
		}

		if playerID != "" && session.PlayerID != playerID {
			continue
		}

		result = append(result, session)
	}

	return result, nil
}

// Helper functions.

func generateBuildARN(region, accountID, buildID string) string {
	return fmt.Sprintf("arn:aws:gamelift:%s:%s:build/%s", region, accountID, buildID)
}

func generateFleetARN(region, accountID, fleetID string) string {
	return fmt.Sprintf("arn:aws:gamelift:%s:%s:fleet/%s", region, accountID, fleetID)
}

func generateGameSessionARN(region, accountID, fleetID, gameSessionID string) string {
	return fmt.Sprintf("arn:aws:gamelift:%s:%s:gamesession/%s/%s", region, accountID, fleetID, gameSessionID)
}

func defaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}
