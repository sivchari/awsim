package gamelift

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// handlerFunc is a type alias for handler functions.
type handlerFunc func(http.ResponseWriter, *http.Request)

// getActionHandlers returns a map of action names to handler functions.
func (s *Service) getActionHandlers() map[string]handlerFunc {
	return map[string]handlerFunc{
		// Build operations
		"CreateBuild":   s.CreateBuild,
		"DescribeBuild": s.DescribeBuild,
		"ListBuilds":    s.ListBuilds,
		"DeleteBuild":   s.DeleteBuild,
		// Fleet operations
		"CreateFleet":             s.CreateFleet,
		"DescribeFleetAttributes": s.DescribeFleetAttributes,
		"ListFleets":              s.ListFleets,
		"DeleteFleet":             s.DeleteFleet,
		// Game session operations
		"CreateGameSession":    s.CreateGameSession,
		"DescribeGameSessions": s.DescribeGameSessions,
		"UpdateGameSession":    s.UpdateGameSession,
		// Player session operations
		"CreatePlayerSession":    s.CreatePlayerSession,
		"CreatePlayerSessions":   s.CreatePlayerSessions,
		"DescribePlayerSessions": s.DescribePlayerSessions,
	}
}

// DispatchAction dispatches the request to the appropriate handler.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "GameLift.")

	handlers := s.getActionHandlers()
	if handler, ok := handlers[action]; ok {
		handler(w, r)

		return
	}

	writeError(w, "UnknownOperationException", "The operation "+action+" is not valid.", http.StatusBadRequest)
}

// CreateBuild handles the CreateBuild API.
func (s *Service) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var req CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	build, err := s.storage.CreateBuild(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreateBuildResponse{
		Build: convertToBuildOutput(build),
	}

	writeResponse(w, resp)
}

// DescribeBuild handles the DescribeBuild API.
func (s *Service) DescribeBuild(w http.ResponseWriter, r *http.Request) {
	var req DescribeBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.BuildId == "" {
		writeError(w, "InvalidRequestException", "BuildId is required", http.StatusBadRequest)

		return
	}

	build, err := s.storage.DescribeBuild(r.Context(), req.BuildId)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &DescribeBuildResponse{
		Build: convertToBuildOutput(build),
	}

	writeResponse(w, resp)
}

// ListBuilds handles the ListBuilds API.
func (s *Service) ListBuilds(w http.ResponseWriter, r *http.Request) {
	var req ListBuildsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	limit := int32(100)
	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
	}

	builds, err := s.storage.ListBuilds(r.Context(), req.Status, limit)
	if err != nil {
		handleError(w, err)

		return
	}

	buildOutputs := make([]BuildOutput, 0, len(builds))
	for _, build := range builds {
		buildOutputs = append(buildOutputs, *convertToBuildOutput(build))
	}

	resp := &ListBuildsResponse{
		Builds: buildOutputs,
	}

	writeResponse(w, resp)
}

// DeleteBuild handles the DeleteBuild API.
func (s *Service) DeleteBuild(w http.ResponseWriter, r *http.Request) {
	var req DeleteBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.BuildId == "" {
		writeError(w, "InvalidRequestException", "BuildId is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteBuild(r.Context(), req.BuildId); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &DeleteBuildResponse{})
}

// CreateFleet handles the CreateFleet API.
func (s *Service) CreateFleet(w http.ResponseWriter, r *http.Request) {
	var req CreateFleetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	fleet, err := s.storage.CreateFleet(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreateFleetResponse{
		FleetAttributes: convertToFleetAttributesOutput(fleet),
	}

	writeResponse(w, resp)
}

// DescribeFleetAttributes handles the DescribeFleetAttributes API.
func (s *Service) DescribeFleetAttributes(w http.ResponseWriter, r *http.Request) {
	var req DescribeFleetAttributesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	fleets, err := s.storage.DescribeFleetAttributes(r.Context(), req.FleetIds)
	if err != nil {
		handleError(w, err)

		return
	}

	fleetOutputs := make([]FleetAttributesOutput, 0, len(fleets))
	for _, fleet := range fleets {
		fleetOutputs = append(fleetOutputs, *convertToFleetAttributesOutput(fleet))
	}

	resp := &DescribeFleetAttributesResponse{
		FleetAttributes: fleetOutputs,
	}

	writeResponse(w, resp)
}

// ListFleets handles the ListFleets API.
func (s *Service) ListFleets(w http.ResponseWriter, r *http.Request) {
	var req ListFleetsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	limit := int32(100)
	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
	}

	fleetIDs, err := s.storage.ListFleets(r.Context(), req.BuildId, limit)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &ListFleetsResponse{
		FleetIds: fleetIDs,
	}

	writeResponse(w, resp)
}

// DeleteFleet handles the DeleteFleet API.
func (s *Service) DeleteFleet(w http.ResponseWriter, r *http.Request) {
	var req DeleteFleetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.FleetId == "" {
		writeError(w, "InvalidRequestException", "FleetId is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteFleet(r.Context(), req.FleetId); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &DeleteFleetResponse{})
}

// CreateGameSession handles the CreateGameSession API.
func (s *Service) CreateGameSession(w http.ResponseWriter, r *http.Request) {
	var req CreateGameSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	gameSession, err := s.storage.CreateGameSession(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreateGameSessionResponse{
		GameSession: convertToGameSessionOutput(gameSession),
	}

	writeResponse(w, resp)
}

// DescribeGameSessions handles the DescribeGameSessions API.
func (s *Service) DescribeGameSessions(w http.ResponseWriter, r *http.Request) {
	var req DescribeGameSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	sessions, err := s.storage.DescribeGameSessions(r.Context(), req.FleetId, req.GameSessionId)
	if err != nil {
		handleError(w, err)

		return
	}

	sessionOutputs := make([]GameSessionOutput, 0, len(sessions))
	for _, session := range sessions {
		sessionOutputs = append(sessionOutputs, *convertToGameSessionOutput(session))
	}

	resp := &DescribeGameSessionsResponse{
		GameSessions: sessionOutputs,
	}

	writeResponse(w, resp)
}

// UpdateGameSession handles the UpdateGameSession API.
func (s *Service) UpdateGameSession(w http.ResponseWriter, r *http.Request) {
	var req UpdateGameSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.GameSessionId == "" {
		writeError(w, "InvalidRequestException", "GameSessionId is required", http.StatusBadRequest)

		return
	}

	gameSession, err := s.storage.UpdateGameSession(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &UpdateGameSessionResponse{
		GameSession: convertToGameSessionOutput(gameSession),
	}

	writeResponse(w, resp)
}

// CreatePlayerSession handles the CreatePlayerSession API.
func (s *Service) CreatePlayerSession(w http.ResponseWriter, r *http.Request) {
	var req CreatePlayerSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.GameSessionId == "" {
		writeError(w, "InvalidRequestException", "GameSessionId is required", http.StatusBadRequest)

		return
	}

	if req.PlayerId == "" {
		writeError(w, "InvalidRequestException", "PlayerId is required", http.StatusBadRequest)

		return
	}

	playerSession, err := s.storage.CreatePlayerSession(r.Context(), req.GameSessionId, req.PlayerId)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreatePlayerSessionResponse{
		PlayerSession: convertToPlayerSessionOutput(playerSession),
	}

	writeResponse(w, resp)
}

// CreatePlayerSessions handles the CreatePlayerSessions API.
func (s *Service) CreatePlayerSessions(w http.ResponseWriter, r *http.Request) {
	var req CreatePlayerSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.GameSessionId == "" {
		writeError(w, "InvalidRequestException", "GameSessionId is required", http.StatusBadRequest)

		return
	}

	if len(req.PlayerIds) == 0 {
		writeError(w, "InvalidRequestException", "PlayerIds is required", http.StatusBadRequest)

		return
	}

	playerSessions, err := s.storage.CreatePlayerSessions(r.Context(), req.GameSessionId, req.PlayerIds)
	if err != nil {
		handleError(w, err)

		return
	}

	sessionOutputs := make([]PlayerSessionOutput, 0, len(playerSessions))
	for _, session := range playerSessions {
		sessionOutputs = append(sessionOutputs, *convertToPlayerSessionOutput(session))
	}

	resp := &CreatePlayerSessionsResponse{
		PlayerSessions: sessionOutputs,
	}

	writeResponse(w, resp)
}

// DescribePlayerSessions handles the DescribePlayerSessions API.
func (s *Service) DescribePlayerSessions(w http.ResponseWriter, r *http.Request) {
	var req DescribePlayerSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "InvalidRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	sessions, err := s.storage.DescribePlayerSessions(r.Context(), req.GameSessionId, req.PlayerSessionId, req.PlayerId)
	if err != nil {
		handleError(w, err)

		return
	}

	sessionOutputs := make([]PlayerSessionOutput, 0, len(sessions))
	for _, session := range sessions {
		sessionOutputs = append(sessionOutputs, *convertToPlayerSessionOutput(session))
	}

	resp := &DescribePlayerSessionsResponse{
		PlayerSessions: sessionOutputs,
	}

	writeResponse(w, resp)
}

// Helper functions.

// convertToBuildOutput converts a Build to BuildOutput.
func convertToBuildOutput(build *Build) *BuildOutput {
	return &BuildOutput{
		BuildID:         build.BuildID,
		BuildARN:        build.BuildARN,
		Name:            build.Name,
		Version:         build.Version,
		Status:          build.Status,
		SizeOnDisk:      build.SizeOnDisk,
		OperatingSystem: build.OperatingSystem,
		CreationTime:    float64(build.CreationTime.Unix()),
	}
}

// convertToFleetAttributesOutput converts a Fleet to FleetAttributesOutput.
func convertToFleetAttributesOutput(fleet *Fleet) *FleetAttributesOutput {
	return &FleetAttributesOutput{
		FleetId:                        fleet.FleetID,
		FleetArn:                       fleet.FleetARN,
		FleetType:                      fleet.FleetType,
		InstanceType:                   fleet.InstanceType,
		Description:                    fleet.Description,
		Name:                           fleet.Name,
		CreationTime:                   float64(fleet.CreationTime.Unix()),
		Status:                         fleet.Status,
		BuildId:                        fleet.BuildID,
		ServerLaunchPath:               fleet.ServerLaunchPath,
		NewGameSessionProtectionPolicy: fleet.NewGameSessionProtectionPolicy,
	}
}

// convertToGameSessionOutput converts a GameSession to GameSessionOutput.
func convertToGameSessionOutput(session *GameSession) *GameSessionOutput {
	return &GameSessionOutput{
		GameSessionId:             session.GameSessionID,
		Name:                      session.Name,
		FleetId:                   session.FleetID,
		FleetArn:                  session.FleetARN,
		CreationTime:              float64(session.CreationTime.Unix()),
		CurrentPlayerSessionCount: int32(session.CurrentPlayerSessionCount),
		MaximumPlayerSessionCount: int32(session.MaximumPlayerSessionCount),
		Status:                    session.Status,
		IpAddress:                 session.IPAddress,
		Port:                      int32(session.Port),
	}
}

// convertToPlayerSessionOutput converts a PlayerSession to PlayerSessionOutput.
func convertToPlayerSessionOutput(session *PlayerSession) *PlayerSessionOutput {
	return &PlayerSessionOutput{
		PlayerSessionId: session.PlayerSessionID,
		PlayerId:        session.PlayerID,
		GameSessionId:   session.GameSessionID,
		FleetId:         session.FleetID,
		FleetArn:        session.FleetARN,
		CreationTime:    float64(session.CreationTime.Unix()),
		Status:          session.Status,
		IpAddress:       session.IPAddress,
		Port:            int32(session.Port),
	}
}

// writeResponse writes a JSON response.
func writeResponse(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{
		Type:    code,
		Message: message,
	})
}

// handleError handles service errors.
func handleError(w http.ResponseWriter, err error) {
	var glErr *Error
	if errors.As(err, &glErr) {
		status := getErrorStatus(glErr.Code)
		writeError(w, glErr.Code, glErr.Message, status)

		return
	}

	writeError(w, "InternalServiceException", err.Error(), http.StatusInternalServerError)
}

// getErrorStatus returns the HTTP status code for a given error code.
func getErrorStatus(code string) int {
	switch code {
	case errNotFoundException:
		return http.StatusNotFound
	case errInvalidRequestException:
		return http.StatusBadRequest
	case errConflictException:
		return http.StatusConflict
	case errLimitExceededException:
		return http.StatusTooManyRequests
	default:
		return http.StatusBadRequest
	}
}
