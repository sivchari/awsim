package secretsmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Error codes for Secrets Manager.
const (
	errResourceNotFound     = "ResourceNotFoundException"
	errInvalidParameter     = "InvalidParameterException"
	errInternalServiceError = "InternalServiceError"
	errInvalidAction        = "InvalidAction"
)

// CreateSecret handles the CreateSecret action.
func (s *Service) CreateSecret(w http.ResponseWriter, r *http.Request) {
	var req CreateSecretRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.Name == "" {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide a value for the Name parameter.", http.StatusBadRequest)

		return
	}

	secret, err := s.storage.CreateSecret(r.Context(), &req)
	if err != nil {
		var sErr *SecretError
		if errors.As(err, &sErr) {
			writeSecretsManagerError(w, sErr.Code, sErr.Message, http.StatusBadRequest)

			return
		}

		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, CreateSecretResponse{
		ARN:       secret.ARN,
		Name:      secret.Name,
		VersionID: secret.VersionID,
	})
}

// GetSecretValue handles the GetSecretValue action.
func (s *Service) GetSecretValue(w http.ResponseWriter, r *http.Request) {
	var req GetSecretValueRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.SecretID == "" {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide a value for the SecretId parameter.", http.StatusBadRequest)

		return
	}

	secret, version, err := s.storage.GetSecretValue(r.Context(), req.SecretID, req.VersionID, req.VersionStage)
	if err != nil {
		var sErr *SecretError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errResourceNotFound {
				status = http.StatusNotFound
			}

			writeSecretsManagerError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, GetSecretValueResponse{
		ARN:           secret.ARN,
		Name:          secret.Name,
		VersionID:     version.VersionID,
		SecretBinary:  version.SecretBinary,
		SecretString:  version.SecretString,
		VersionStages: version.VersionStages,
		CreatedDate:   float64(version.CreatedDate.Unix()),
	})
}

// PutSecretValue handles the PutSecretValue action.
func (s *Service) PutSecretValue(w http.ResponseWriter, r *http.Request) {
	var req PutSecretValueRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.SecretID == "" {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide a value for the SecretId parameter.", http.StatusBadRequest)

		return
	}

	if req.SecretString == "" && len(req.SecretBinary) == 0 {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide either SecretString or SecretBinary.", http.StatusBadRequest)

		return
	}

	secret, version, err := s.storage.PutSecretValue(r.Context(), req.SecretID, req.ClientRequestToken, req.SecretString, req.SecretBinary, req.VersionStages)
	if err != nil {
		var sErr *SecretError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errResourceNotFound {
				status = http.StatusNotFound
			}

			writeSecretsManagerError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, PutSecretValueResponse{
		ARN:           secret.ARN,
		Name:          secret.Name,
		VersionID:     version.VersionID,
		VersionStages: version.VersionStages,
	})
}

// DeleteSecret handles the DeleteSecret action.
func (s *Service) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	var req DeleteSecretRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.SecretID == "" {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide a value for the SecretId parameter.", http.StatusBadRequest)

		return
	}

	secret, err := s.storage.DeleteSecret(r.Context(), req.SecretID, req.RecoveryWindowInDays, req.ForceDeleteWithoutRecovery)
	if err != nil {
		var sErr *SecretError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errResourceNotFound {
				status = http.StatusNotFound
			}

			writeSecretsManagerError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	var deletionDate float64
	if secret.DeletedDate != nil {
		deletionDate = float64(secret.DeletedDate.Unix())
	}

	if secret.ScheduledDeletionDate != nil {
		deletionDate = float64(secret.ScheduledDeletionDate.Unix())
	}

	writeJSONResponse(w, DeleteSecretResponse{
		ARN:          secret.ARN,
		Name:         secret.Name,
		DeletionDate: deletionDate,
	})
}

// ListSecrets handles the ListSecrets action.
func (s *Service) ListSecrets(w http.ResponseWriter, r *http.Request) {
	var req ListSecretsRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	secrets, nextToken, err := s.storage.ListSecrets(r.Context(), req.MaxResults, req.NextToken, req.IncludePlannedDeletion)
	if err != nil {
		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	secretList := convertSecretsToListEntries(secrets)

	writeJSONResponse(w, ListSecretsResponse{
		SecretList: secretList,
		NextToken:  nextToken,
	})
}

// convertSecretsToListEntries converts secrets to list entries.
func convertSecretsToListEntries(secrets []*Secret) []SecretListEntry {
	secretList := make([]SecretListEntry, 0, len(secrets))

	for _, secret := range secrets {
		entry := SecretListEntry{
			ARN:             secret.ARN,
			Name:            secret.Name,
			Description:     secret.Description,
			KmsKeyID:        secret.KmsKeyID,
			RotationEnabled: secret.RotationEnabled,
			RotationRules:   secret.RotationRules,
			Tags:            secret.Tags,
			CreatedDate:     float64(secret.CreatedDate.Unix()),
			PrimaryRegion:   secret.PrimaryRegion,
			OwningService:   secret.OwningService,
		}

		if secret.LastChangedDate.Unix() > 0 {
			lastChanged := float64(secret.LastChangedDate.Unix())
			entry.LastChangedDate = &lastChanged
		}

		if secret.LastAccessedDate != nil {
			lastAccessed := float64(secret.LastAccessedDate.Unix())
			entry.LastAccessedDate = &lastAccessed
		}

		if secret.DeletedDate != nil {
			deleted := float64(secret.DeletedDate.Unix())
			entry.DeletedDate = &deleted
		}

		if secret.LastRotationDate != nil {
			lastRotated := float64(secret.LastRotationDate.Unix())
			entry.LastRotatedDate = &lastRotated
		}

		if secret.NextRotationDate != nil {
			nextRotation := float64(secret.NextRotationDate.Unix())
			entry.NextRotationDate = &nextRotation
		}

		entry.SecretVersionsToStages = make(map[string][]string)

		for versionID, version := range secret.VersionIDs {
			entry.SecretVersionsToStages[versionID] = version.VersionStages
		}

		secretList = append(secretList, entry)
	}

	return secretList
}

// DescribeSecret handles the DescribeSecret action.
func (s *Service) DescribeSecret(w http.ResponseWriter, r *http.Request) {
	var req DescribeSecretRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.SecretID == "" {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide a value for the SecretId parameter.", http.StatusBadRequest)

		return
	}

	secret, err := s.storage.DescribeSecret(r.Context(), req.SecretID)
	if err != nil {
		var sErr *SecretError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errResourceNotFound {
				status = http.StatusNotFound
			}

			writeSecretsManagerError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := buildDescribeSecretResponse(secret)

	writeJSONResponse(w, resp)
}

// buildDescribeSecretResponse builds the DescribeSecret response.
func buildDescribeSecretResponse(secret *Secret) DescribeSecretResponse {
	resp := DescribeSecretResponse{
		ARN:               secret.ARN,
		Name:              secret.Name,
		Description:       secret.Description,
		KmsKeyID:          secret.KmsKeyID,
		RotationEnabled:   secret.RotationEnabled,
		RotationLambdaARN: secret.RotationLambdaARN,
		RotationRules:     secret.RotationRules,
		Tags:              secret.Tags,
		CreatedDate:       float64(secret.CreatedDate.Unix()),
		PrimaryRegion:     secret.PrimaryRegion,
		OwningService:     secret.OwningService,
		ReplicationStatus: secret.ReplicationStatus,
	}

	if secret.LastChangedDate.Unix() > 0 {
		lastChanged := float64(secret.LastChangedDate.Unix())
		resp.LastChangedDate = &lastChanged
	}

	if secret.LastAccessedDate != nil {
		lastAccessed := float64(secret.LastAccessedDate.Unix())
		resp.LastAccessedDate = &lastAccessed
	}

	if secret.DeletedDate != nil {
		deleted := float64(secret.DeletedDate.Unix())
		resp.DeletedDate = &deleted
	}

	if secret.LastRotationDate != nil {
		lastRotated := float64(secret.LastRotationDate.Unix())
		resp.LastRotatedDate = &lastRotated
	}

	if secret.NextRotationDate != nil {
		nextRotation := float64(secret.NextRotationDate.Unix())
		resp.NextRotationDate = &nextRotation
	}

	resp.VersionIDsToStages = make(map[string][]string)

	for versionID, version := range secret.VersionIDs {
		resp.VersionIDsToStages[versionID] = version.VersionStages
	}

	return resp
}

// UpdateSecret handles the UpdateSecret action.
func (s *Service) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	var req UpdateSecretRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeSecretsManagerError(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.SecretID == "" {
		writeSecretsManagerError(w, errInvalidParameter, "You must provide a value for the SecretId parameter.", http.StatusBadRequest)

		return
	}

	secret, version, err := s.storage.UpdateSecret(r.Context(), &req)
	if err != nil {
		var sErr *SecretError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errResourceNotFound {
				status = http.StatusNotFound
			}

			writeSecretsManagerError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeSecretsManagerError(w, errInternalServiceError, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := UpdateSecretResponse{
		ARN:  secret.ARN,
		Name: secret.Name,
	}

	if version != nil {
		resp.VersionID = version.VersionID
	}

	writeJSONResponse(w, resp)
}

// readJSONRequest reads and decodes JSON request body.
func readJSONRequest(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// writeJSONResponse writes a JSON response with HTTP 200 OK.
func writeJSONResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// writeSecretsManagerError writes a Secrets Manager error response in JSON format.
func writeSecretsManagerError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Type:    code,
		Message: message,
	})
}

// DispatchAction routes the request to the appropriate handler based on X-Amz-Target header.
// This method implements the JSONProtocolService interface.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "secretsmanager.")

	switch action {
	case "CreateSecret":
		s.CreateSecret(w, r)
	case "GetSecretValue":
		s.GetSecretValue(w, r)
	case "PutSecretValue":
		s.PutSecretValue(w, r)
	case "DeleteSecret":
		s.DeleteSecret(w, r)
	case "ListSecrets":
		s.ListSecrets(w, r)
	case "DescribeSecret":
		s.DescribeSecret(w, r)
	case "UpdateSecret":
		s.UpdateSecret(w, r)
	default:
		writeSecretsManagerError(w, errInvalidAction, "The action "+action+" is not valid", http.StatusBadRequest)
	}
}
