// Package cognito provides AWS Cognito Identity Provider service emulation.
package cognito

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
		"CreateUserPool":         s.CreateUserPool,
		"DescribeUserPool":       s.DescribeUserPool,
		"ListUserPools":          s.ListUserPools,
		"DeleteUserPool":         s.DeleteUserPool,
		"CreateUserPoolClient":   s.CreateUserPoolClient,
		"DescribeUserPoolClient": s.DescribeUserPoolClient,
		"ListUserPoolClients":    s.ListUserPoolClients,
		"DeleteUserPoolClient":   s.DeleteUserPoolClient,
		"AdminCreateUser":        s.AdminCreateUser,
		"AdminGetUser":           s.AdminGetUser,
		"AdminDeleteUser":        s.AdminDeleteUser,
		"ListUsers":              s.ListUsers,
		"SignUp":                 s.SignUp,
		"ConfirmSignUp":          s.ConfirmSignUp,
		"InitiateAuth":           s.InitiateAuth,
	}
}

// DispatchAction dispatches the request to the appropriate handler.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "AWSCognitoIdentityProviderService.")

	handlers := s.getActionHandlers()
	if handler, ok := handlers[action]; ok {
		handler(w, r)

		return
	}

	writeError(w, "InvalidAction", "The action "+action+" is not valid for this endpoint.", http.StatusBadRequest)
}

// CreateUserPool handles the CreateUserPool API.
func (s *Service) CreateUserPool(w http.ResponseWriter, r *http.Request) {
	var req CreateUserPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	pool, err := s.storage.CreateUserPool(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreateUserPoolResponse{
		UserPool: userPoolToOutput(pool),
	}

	writeResponse(w, resp)
}

// DescribeUserPool handles the DescribeUserPool API.
func (s *Service) DescribeUserPool(w http.ResponseWriter, r *http.Request) {
	var req DescribeUserPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	pool, err := s.storage.GetUserPool(r.Context(), req.UserPoolID)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &DescribeUserPoolResponse{
		UserPool: userPoolToOutput(pool),
	}

	writeResponse(w, resp)
}

// ListUserPools handles the ListUserPools API.
func (s *Service) ListUserPools(w http.ResponseWriter, r *http.Request) {
	var req ListUserPoolsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	pools, nextToken, err := s.storage.ListUserPools(r.Context(), req.MaxResults, req.NextToken)
	if err != nil {
		handleError(w, err)

		return
	}

	outputs := make([]UserPoolOutput, len(pools))

	for i, pool := range pools {
		outputs[i] = *userPoolToOutput(pool)
	}

	resp := &ListUserPoolsResponse{
		UserPools: outputs,
		NextToken: nextToken,
	}

	writeResponse(w, resp)
}

// DeleteUserPool handles the DeleteUserPool API.
func (s *Service) DeleteUserPool(w http.ResponseWriter, r *http.Request) {
	var req DeleteUserPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteUserPool(r.Context(), req.UserPoolID); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &DeleteUserPoolResponse{})
}

// CreateUserPoolClient handles the CreateUserPoolClient API.
func (s *Service) CreateUserPoolClient(w http.ResponseWriter, r *http.Request) {
	var req CreateUserPoolClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	client, err := s.storage.CreateUserPoolClient(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreateUserPoolClientResponse{
		UserPoolClient: userPoolClientToOutput(client),
	}

	writeResponse(w, resp)
}

// DescribeUserPoolClient handles the DescribeUserPoolClient API.
func (s *Service) DescribeUserPoolClient(w http.ResponseWriter, r *http.Request) {
	var req DescribeUserPoolClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	client, err := s.storage.GetUserPoolClient(r.Context(), req.UserPoolID, req.ClientID)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &DescribeUserPoolClientResponse{
		UserPoolClient: userPoolClientToOutput(client),
	}

	writeResponse(w, resp)
}

// ListUserPoolClients handles the ListUserPoolClients API.
func (s *Service) ListUserPoolClients(w http.ResponseWriter, r *http.Request) {
	var req ListUserPoolClientsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	clients, nextToken, err := s.storage.ListUserPoolClients(r.Context(), req.UserPoolID, req.MaxResults, req.NextToken)
	if err != nil {
		handleError(w, err)

		return
	}

	outputs := make([]UserPoolClientOutput, len(clients))

	for i, client := range clients {
		outputs[i] = *userPoolClientToOutput(client)
	}

	resp := &ListUserPoolClientsResponse{
		UserPoolClients: outputs,
		NextToken:       nextToken,
	}

	writeResponse(w, resp)
}

// DeleteUserPoolClient handles the DeleteUserPoolClient API.
func (s *Service) DeleteUserPoolClient(w http.ResponseWriter, r *http.Request) {
	var req DeleteUserPoolClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteUserPoolClient(r.Context(), req.UserPoolID, req.ClientID); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &DeleteUserPoolClientResponse{})
}

// AdminCreateUser handles the AdminCreateUser API.
func (s *Service) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req AdminCreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	user, err := s.storage.AdminCreateUser(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &AdminCreateUserResponse{
		User: userToOutput(user),
	}

	writeResponse(w, resp)
}

// AdminGetUser handles the AdminGetUser API.
func (s *Service) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	var req AdminGetUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	user, err := s.storage.AdminGetUser(r.Context(), req.UserPoolID, req.Username)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &AdminGetUserResponse{
		Username:             user.Username,
		UserAttributes:       convertAttributes(user.Attributes),
		UserCreateDate:       float64(user.UserCreateDate.Unix()),
		UserLastModifiedDate: float64(user.UserLastModified.Unix()),
		Enabled:              user.Enabled,
		UserStatus:           string(user.UserStatus),
	}

	writeResponse(w, resp)
}

// AdminDeleteUser handles the AdminDeleteUser API.
func (s *Service) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req AdminDeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.AdminDeleteUser(r.Context(), req.UserPoolID, req.Username); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &AdminDeleteUserResponse{})
}

// ListUsers handles the ListUsers API.
func (s *Service) ListUsers(w http.ResponseWriter, r *http.Request) {
	var req ListUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	users, paginationToken, err := s.storage.ListUsers(r.Context(), req.UserPoolID, req.Limit, req.PaginationToken)
	if err != nil {
		handleError(w, err)

		return
	}

	outputs := make([]UserOutput, len(users))

	for i, user := range users {
		outputs[i] = *userToOutput(user)
	}

	resp := &ListUsersResponse{
		Users:           outputs,
		PaginationToken: paginationToken,
	}

	writeResponse(w, resp)
}

// SignUp handles the SignUp API.
func (s *Service) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	user, err := s.storage.SignUp(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	// Get user sub from attributes.
	userSub := uuid.New().String()

	for _, attr := range user.Attributes {
		if attr.Name == "sub" {
			userSub = attr.Value

			break
		}
	}

	resp := &SignUpResponse{
		UserConfirmed: user.UserStatus == UserStatusConfirmed,
		UserSub:       userSub,
	}

	writeResponse(w, resp)
}

// ConfirmSignUp handles the ConfirmSignUp API.
func (s *Service) ConfirmSignUp(w http.ResponseWriter, r *http.Request) {
	var req ConfirmSignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.ConfirmSignUp(r.Context(), req.ClientID, req.Username, req.ConfirmationCode); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &ConfirmSignUpResponse{})
}

// InitiateAuth handles the InitiateAuth API.
func (s *Service) InitiateAuth(w http.ResponseWriter, r *http.Request) {
	var req InitiateAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ValidationException", "Invalid request body", http.StatusBadRequest)

		return
	}

	resp, err := s.storage.InitiateAuth(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, resp)
}

// userPoolToOutput converts a UserPool to UserPoolOutput.
func userPoolToOutput(pool *UserPool) *UserPoolOutput {
	output := &UserPoolOutput{
		ID:                     pool.ID,
		Name:                   pool.Name,
		Status:                 string(pool.Status),
		CreationDate:           float64(pool.CreationDate.Unix()),
		LastModifiedDate:       float64(pool.LastModifiedDate.Unix()),
		AutoVerifiedAttributes: pool.AutoVerifiedAttrs,
		UsernameAttributes:     pool.UsernameAttributes,
		MfaConfiguration:       pool.MFAConfiguration,
	}

	if pool.Policies != nil && pool.Policies.PasswordPolicy != nil {
		output.Policies = &UserPoolPoliciesOutput{
			PasswordPolicy: &PasswordPolicyOutput{
				MinimumLength:                 pool.Policies.PasswordPolicy.MinimumLength,
				RequireUppercase:              pool.Policies.PasswordPolicy.RequireUppercase,
				RequireLowercase:              pool.Policies.PasswordPolicy.RequireLowercase,
				RequireNumbers:                pool.Policies.PasswordPolicy.RequireNumbers,
				RequireSymbols:                pool.Policies.PasswordPolicy.RequireSymbols,
				TemporaryPasswordValidityDays: pool.Policies.PasswordPolicy.TemporaryPasswordValidityDays,
			},
		}
	}

	return output
}

// userPoolClientToOutput converts a UserPoolClient to UserPoolClientOutput.
func userPoolClientToOutput(client *UserPoolClient) *UserPoolClientOutput {
	return &UserPoolClientOutput{
		ClientID:                        client.ClientID,
		ClientName:                      client.ClientName,
		UserPoolID:                      client.UserPoolID,
		ClientSecret:                    client.ClientSecret,
		CreationDate:                    float64(client.CreationDate.Unix()),
		LastModifiedDate:                float64(client.LastModifiedDate.Unix()),
		RefreshTokenValidity:            client.RefreshTokenValidity,
		AccessTokenValidity:             client.AccessTokenValidity,
		IDTokenValidity:                 client.IDTokenValidity,
		ExplicitAuthFlows:               client.ExplicitAuthFlows,
		SupportedIdentityProviders:      client.SupportedIdentityProviders,
		CallbackURLs:                    client.CallbackURLs,
		LogoutURLs:                      client.LogoutURLs,
		AllowedOAuthFlows:               client.AllowedOAuthFlows,
		AllowedOAuthScopes:              client.AllowedOAuthScopes,
		AllowedOAuthFlowsUserPoolClient: client.AllowedOAuthFlowsUserPoolClient,
	}
}

// userToOutput converts a User to UserOutput.
func userToOutput(user *User) *UserOutput {
	return &UserOutput{
		Username:             user.Username,
		Attributes:           convertAttributes(user.Attributes),
		UserCreateDate:       float64(user.UserCreateDate.Unix()),
		UserLastModifiedDate: float64(user.UserLastModified.Unix()),
		Enabled:              user.Enabled,
		UserStatus:           string(user.UserStatus),
	}
}

// convertAttributes converts UserAttribute slice to UserAttributeOutput slice.
func convertAttributes(attrs []UserAttribute) []UserAttributeOutput {
	if attrs == nil {
		return nil
	}

	outputs := make([]UserAttributeOutput, len(attrs))

	for i, attr := range attrs {
		outputs[i] = UserAttributeOutput{
			Name:  attr.Name,
			Value: attr.Value,
		}
	}

	return outputs
}

// writeResponse writes a JSON response.
func writeResponse(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{
		Type:    code,
		Message: message,
	})
}

// handleError handles service errors.
func handleError(w http.ResponseWriter, err error) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		status := getErrorStatus(svcErr.Code)
		writeError(w, svcErr.Code, svcErr.Message, status)

		return
	}

	writeError(w, "InternalServiceError", err.Error(), http.StatusInternalServerError)
}

// getErrorStatus returns the HTTP status code for a given error code.
func getErrorStatus(code string) int {
	switch code {
	case "ResourceNotFoundException", "UserNotFoundException":
		return http.StatusNotFound
	case "NotAuthorizedException":
		return http.StatusUnauthorized
	case "UsernameExistsException":
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}
