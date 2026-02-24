//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/stretchr/testify/require"
)

func newGameLiftClient(t *testing.T) *gamelift.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return gamelift.NewFromConfig(cfg, func(o *gamelift.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestGameLift_CreateAndDeleteBuild(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	buildName := "test-build"
	buildVersion := "1.0.0"

	// Create build.
	createOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String(buildName),
		Version: aws.String(buildVersion),
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.Build)
	require.NotEmpty(t, createOutput.Build.BuildId)
	require.Equal(t, buildName, *createOutput.Build.Name)
	require.Equal(t, buildVersion, *createOutput.Build.Version)

	buildID := createOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	// Describe build.
	descOutput, err := client.DescribeBuild(ctx, &gamelift.DescribeBuildInput{
		BuildId: buildID,
	})
	require.NoError(t, err)
	require.NotNil(t, descOutput.Build)
	require.Equal(t, *buildID, *descOutput.Build.BuildId)
	require.Equal(t, buildName, *descOutput.Build.Name)

	// Delete build.
	_, err = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
		BuildId: buildID,
	})
	require.NoError(t, err)

	// Verify build is deleted.
	_, err = client.DescribeBuild(ctx, &gamelift.DescribeBuildInput{
		BuildId: buildID,
	})
	require.Error(t, err)
}

func TestGameLift_ListBuilds(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create multiple builds.
	var buildIDs []*string

	for i := 0; i < 3; i++ {
		createOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
			Name:    aws.String("test-build-list"),
			Version: aws.String("1.0.0"),
		})
		require.NoError(t, err)
		buildIDs = append(buildIDs, createOutput.Build.BuildId)
	}

	t.Cleanup(func() {
		for _, buildID := range buildIDs {
			_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
				BuildId: buildID,
			})
		}
	})

	// List builds.
	listOutput, err := client.ListBuilds(ctx, &gamelift.ListBuildsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.Builds), 3)
}

func TestGameLift_CreateAndDeleteFleet(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create a build first.
	buildOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String("test-build-for-fleet"),
		Version: aws.String("1.0.0"),
	})
	require.NoError(t, err)

	buildID := buildOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	fleetName := "test-fleet"

	// Create fleet.
	createOutput, err := client.CreateFleet(ctx, &gamelift.CreateFleetInput{
		Name:            aws.String(fleetName),
		BuildId:         buildID,
		EC2InstanceType: "c5.large",
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.FleetAttributes)
	require.NotEmpty(t, createOutput.FleetAttributes.FleetId)
	require.Equal(t, fleetName, *createOutput.FleetAttributes.Name)

	fleetID := createOutput.FleetAttributes.FleetId

	t.Cleanup(func() {
		_, _ = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
			FleetId: fleetID,
		})
	})

	// Describe fleet attributes.
	descOutput, err := client.DescribeFleetAttributes(ctx, &gamelift.DescribeFleetAttributesInput{
		FleetIds: []string{*fleetID},
	})
	require.NoError(t, err)
	require.Len(t, descOutput.FleetAttributes, 1)
	require.Equal(t, *fleetID, *descOutput.FleetAttributes[0].FleetId)

	// Delete fleet.
	_, err = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
		FleetId: fleetID,
	})
	require.NoError(t, err)

	// Verify fleet is deleted.
	descOutput, err = client.DescribeFleetAttributes(ctx, &gamelift.DescribeFleetAttributesInput{
		FleetIds: []string{*fleetID},
	})
	require.NoError(t, err)
	require.Empty(t, descOutput.FleetAttributes)
}

func TestGameLift_ListFleets(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create a build first.
	buildOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String("test-build-for-list-fleets"),
		Version: aws.String("1.0.0"),
	})
	require.NoError(t, err)

	buildID := buildOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	// Create multiple fleets.
	var fleetIDs []*string

	for i := 0; i < 2; i++ {
		createOutput, err := client.CreateFleet(ctx, &gamelift.CreateFleetInput{
			Name:            aws.String("test-fleet-list"),
			BuildId:         buildID,
			EC2InstanceType: "c5.large",
		})
		require.NoError(t, err)
		fleetIDs = append(fleetIDs, createOutput.FleetAttributes.FleetId)
	}

	t.Cleanup(func() {
		for _, fleetID := range fleetIDs {
			_, _ = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
				FleetId: fleetID,
			})
		}
	})

	// List fleets.
	listOutput, err := client.ListFleets(ctx, &gamelift.ListFleetsInput{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listOutput.FleetIds), 2)
}

func TestGameLift_CreateAndDescribeGameSession(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create a build.
	buildOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String("test-build-for-gamesession"),
		Version: aws.String("1.0.0"),
	})
	require.NoError(t, err)

	buildID := buildOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	// Create a fleet.
	fleetOutput, err := client.CreateFleet(ctx, &gamelift.CreateFleetInput{
		Name:            aws.String("test-fleet-for-gamesession"),
		BuildId:         buildID,
		EC2InstanceType: "c5.large",
	})
	require.NoError(t, err)

	fleetID := fleetOutput.FleetAttributes.FleetId

	t.Cleanup(func() {
		_, _ = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
			FleetId: fleetID,
		})
	})

	// Create a game session.
	sessionName := "test-game-session"
	maxPlayers := int32(10)

	createSessionOutput, err := client.CreateGameSession(ctx, &gamelift.CreateGameSessionInput{
		FleetId:                   fleetID,
		Name:                      aws.String(sessionName),
		MaximumPlayerSessionCount: aws.Int32(maxPlayers),
	})
	require.NoError(t, err)
	require.NotNil(t, createSessionOutput.GameSession)
	require.NotEmpty(t, createSessionOutput.GameSession.GameSessionId)
	require.Equal(t, sessionName, *createSessionOutput.GameSession.Name)
	require.Equal(t, maxPlayers, *createSessionOutput.GameSession.MaximumPlayerSessionCount)

	gameSessionID := createSessionOutput.GameSession.GameSessionId

	// Describe game session.
	descOutput, err := client.DescribeGameSessions(ctx, &gamelift.DescribeGameSessionsInput{
		GameSessionId: gameSessionID,
	})
	require.NoError(t, err)
	require.Len(t, descOutput.GameSessions, 1)
	require.Equal(t, *gameSessionID, *descOutput.GameSessions[0].GameSessionId)
}

func TestGameLift_UpdateGameSession(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create a build.
	buildOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String("test-build-for-update"),
		Version: aws.String("1.0.0"),
	})
	require.NoError(t, err)

	buildID := buildOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	// Create a fleet.
	fleetOutput, err := client.CreateFleet(ctx, &gamelift.CreateFleetInput{
		Name:            aws.String("test-fleet-for-update"),
		BuildId:         buildID,
		EC2InstanceType: "c5.large",
	})
	require.NoError(t, err)

	fleetID := fleetOutput.FleetAttributes.FleetId

	t.Cleanup(func() {
		_, _ = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
			FleetId: fleetID,
		})
	})

	// Create a game session.
	createSessionOutput, err := client.CreateGameSession(ctx, &gamelift.CreateGameSessionInput{
		FleetId:                   fleetID,
		Name:                      aws.String("original-name"),
		MaximumPlayerSessionCount: aws.Int32(10),
	})
	require.NoError(t, err)

	gameSessionID := createSessionOutput.GameSession.GameSessionId

	// Update game session.
	newName := "updated-name"
	newMaxPlayers := int32(20)

	updateOutput, err := client.UpdateGameSession(ctx, &gamelift.UpdateGameSessionInput{
		GameSessionId:             gameSessionID,
		Name:                      aws.String(newName),
		MaximumPlayerSessionCount: aws.Int32(newMaxPlayers),
	})
	require.NoError(t, err)
	require.Equal(t, newName, *updateOutput.GameSession.Name)
	require.Equal(t, newMaxPlayers, *updateOutput.GameSession.MaximumPlayerSessionCount)
}

func TestGameLift_CreateAndDescribePlayerSession(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create a build.
	buildOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String("test-build-for-playersession"),
		Version: aws.String("1.0.0"),
	})
	require.NoError(t, err)

	buildID := buildOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	// Create a fleet.
	fleetOutput, err := client.CreateFleet(ctx, &gamelift.CreateFleetInput{
		Name:            aws.String("test-fleet-for-playersession"),
		BuildId:         buildID,
		EC2InstanceType: "c5.large",
	})
	require.NoError(t, err)

	fleetID := fleetOutput.FleetAttributes.FleetId

	t.Cleanup(func() {
		_, _ = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
			FleetId: fleetID,
		})
	})

	// Create a game session.
	createSessionOutput, err := client.CreateGameSession(ctx, &gamelift.CreateGameSessionInput{
		FleetId:                   fleetID,
		Name:                      aws.String("test-session"),
		MaximumPlayerSessionCount: aws.Int32(10),
	})
	require.NoError(t, err)

	gameSessionID := createSessionOutput.GameSession.GameSessionId

	// Create a player session.
	playerID := "player-123"

	createPlayerOutput, err := client.CreatePlayerSession(ctx, &gamelift.CreatePlayerSessionInput{
		GameSessionId: gameSessionID,
		PlayerId:      aws.String(playerID),
	})
	require.NoError(t, err)
	require.NotNil(t, createPlayerOutput.PlayerSession)
	require.NotEmpty(t, createPlayerOutput.PlayerSession.PlayerSessionId)
	require.Equal(t, playerID, *createPlayerOutput.PlayerSession.PlayerId)

	playerSessionID := createPlayerOutput.PlayerSession.PlayerSessionId

	// Describe player session.
	descOutput, err := client.DescribePlayerSessions(ctx, &gamelift.DescribePlayerSessionsInput{
		PlayerSessionId: playerSessionID,
	})
	require.NoError(t, err)
	require.Len(t, descOutput.PlayerSessions, 1)
	require.Equal(t, *playerSessionID, *descOutput.PlayerSessions[0].PlayerSessionId)
}

func TestGameLift_CreatePlayerSessions(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Create a build.
	buildOutput, err := client.CreateBuild(ctx, &gamelift.CreateBuildInput{
		Name:    aws.String("test-build-for-playersessions"),
		Version: aws.String("1.0.0"),
	})
	require.NoError(t, err)

	buildID := buildOutput.Build.BuildId

	t.Cleanup(func() {
		_, _ = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
			BuildId: buildID,
		})
	})

	// Create a fleet.
	fleetOutput, err := client.CreateFleet(ctx, &gamelift.CreateFleetInput{
		Name:            aws.String("test-fleet-for-playersessions"),
		BuildId:         buildID,
		EC2InstanceType: "c5.large",
	})
	require.NoError(t, err)

	fleetID := fleetOutput.FleetAttributes.FleetId

	t.Cleanup(func() {
		_, _ = client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
			FleetId: fleetID,
		})
	})

	// Create a game session.
	createSessionOutput, err := client.CreateGameSession(ctx, &gamelift.CreateGameSessionInput{
		FleetId:                   fleetID,
		Name:                      aws.String("test-session"),
		MaximumPlayerSessionCount: aws.Int32(10),
	})
	require.NoError(t, err)

	gameSessionID := createSessionOutput.GameSession.GameSessionId

	// Create multiple player sessions.
	playerIDs := []string{"player-1", "player-2", "player-3"}

	createPlayersOutput, err := client.CreatePlayerSessions(ctx, &gamelift.CreatePlayerSessionsInput{
		GameSessionId: gameSessionID,
		PlayerIds:     playerIDs,
	})
	require.NoError(t, err)
	require.Len(t, createPlayersOutput.PlayerSessions, 3)
}

func TestGameLift_BuildNotFound(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Describe non-existent build.
	_, err := client.DescribeBuild(ctx, &gamelift.DescribeBuildInput{
		BuildId: aws.String("non-existent-build"),
	})
	require.Error(t, err)

	// Delete non-existent build.
	_, err = client.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
		BuildId: aws.String("non-existent-build"),
	})
	require.Error(t, err)
}

func TestGameLift_FleetNotFound(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Delete non-existent fleet.
	_, err := client.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
		FleetId: aws.String("non-existent-fleet"),
	})
	require.Error(t, err)
}

func TestGameLift_GameSessionNotFound(t *testing.T) {
	client := newGameLiftClient(t)
	ctx := t.Context()

	// Update non-existent game session.
	_, err := client.UpdateGameSession(ctx, &gamelift.UpdateGameSessionInput{
		GameSessionId: aws.String("non-existent-session"),
		Name:          aws.String("new-name"),
	})
	require.Error(t, err)

	// Create player session with non-existent game session.
	_, err = client.CreatePlayerSession(ctx, &gamelift.CreatePlayerSessionInput{
		GameSessionId: aws.String("non-existent-session"),
		PlayerId:      aws.String("player-1"),
	})
	require.Error(t, err)
}
