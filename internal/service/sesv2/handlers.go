package sesv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// CreateEmailIdentity handles the CreateEmailIdentity operation.
func (s *Service) CreateEmailIdentity(w http.ResponseWriter, r *http.Request) {
	var req CreateEmailIdentityRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	identity, err := s.storage.CreateEmailIdentity(r.Context(), &req)
	if err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errAlreadyExists {
				status = http.StatusConflict
			}

			writeError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, http.StatusOK, CreateEmailIdentityResponse{
		IdentityType:             identity.IdentityType,
		VerifiedForSendingStatus: identity.VerifiedForSendingStatus,
		DkimAttributes:           identity.DkimAttributes,
	})
}

// GetEmailIdentity handles the GetEmailIdentity operation.
func (s *Service) GetEmailIdentity(w http.ResponseWriter, r *http.Request) {
	emailIdentity := extractPathParam(r.URL.Path, "/ses/v2/email/identities/")
	if emailIdentity == "" {
		writeError(w, errInvalidParameter, "EmailIdentity is required", http.StatusBadRequest)

		return
	}

	identity, err := s.storage.GetEmailIdentity(r.Context(), emailIdentity)
	if err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, http.StatusOK, GetEmailIdentityResponse{
		IdentityType:             identity.IdentityType,
		FeedbackForwardingStatus: true,
		VerifiedForSendingStatus: identity.VerifiedForSendingStatus,
		DkimAttributes:           identity.DkimAttributes,
	})
}

// ListEmailIdentities handles the ListEmailIdentities operation.
func (s *Service) ListEmailIdentities(w http.ResponseWriter, r *http.Request) {
	nextToken := r.URL.Query().Get("NextToken")

	var pageSize int32 = 100
	if ps := r.URL.Query().Get("PageSize"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	identities, nextTokenOut, err := s.storage.ListEmailIdentities(r.Context(), nextToken, pageSize)
	if err != nil {
		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	summaries := make([]EmailIdentitySummary, 0, len(identities))
	for _, identity := range identities {
		summaries = append(summaries, EmailIdentitySummary{
			IdentityName:   identity.IdentityName,
			IdentityType:   identity.IdentityType,
			SendingEnabled: identity.VerifiedForSendingStatus,
		})
	}

	writeJSONResponse(w, http.StatusOK, ListEmailIdentitiesResponse{
		EmailIdentities: summaries,
		NextToken:       nextTokenOut,
	})
}

// DeleteEmailIdentity handles the DeleteEmailIdentity operation.
func (s *Service) DeleteEmailIdentity(w http.ResponseWriter, r *http.Request) {
	emailIdentity := extractPathParam(r.URL.Path, "/ses/v2/email/identities/")
	if emailIdentity == "" {
		writeError(w, errInvalidParameter, "EmailIdentity is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteEmailIdentity(r.Context(), emailIdentity); err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateConfigurationSet handles the CreateConfigurationSet operation.
func (s *Service) CreateConfigurationSet(w http.ResponseWriter, r *http.Request) {
	var req CreateConfigurationSetRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	_, err := s.storage.CreateConfigurationSet(r.Context(), &req)
	if err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errAlreadyExists {
				status = http.StatusConflict
			}

			writeError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetConfigurationSet handles the GetConfigurationSet operation.
func (s *Service) GetConfigurationSet(w http.ResponseWriter, r *http.Request) {
	name := extractPathParam(r.URL.Path, "/ses/v2/email/configuration-sets/")
	if name == "" {
		writeError(w, errInvalidParameter, "ConfigurationSetName is required", http.StatusBadRequest)

		return
	}

	configSet, err := s.storage.GetConfigurationSet(r.Context(), name)
	if err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, http.StatusOK, GetConfigurationSetResponse{
		ConfigurationSetName: configSet.Name,
		DeliveryOptions:      configSet.DeliveryOptions,
		ReputationOptions:    configSet.ReputationOptions,
		SendingOptions:       configSet.SendingOptions,
		TrackingOptions:      configSet.TrackingOptions,
		Tags:                 configSet.Tags,
	})
}

// ListConfigurationSets handles the ListConfigurationSets operation.
func (s *Service) ListConfigurationSets(w http.ResponseWriter, r *http.Request) {
	nextToken := r.URL.Query().Get("NextToken")

	var pageSize int32 = 100
	if ps := r.URL.Query().Get("PageSize"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	names, nextTokenOut, err := s.storage.ListConfigurationSets(r.Context(), nextToken, pageSize)
	if err != nil {
		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, http.StatusOK, ListConfigurationSetsResponse{
		ConfigurationSets: names,
		NextToken:         nextTokenOut,
	})
}

// DeleteConfigurationSet handles the DeleteConfigurationSet operation.
func (s *Service) DeleteConfigurationSet(w http.ResponseWriter, r *http.Request) {
	name := extractPathParam(r.URL.Path, "/ses/v2/email/configuration-sets/")
	if name == "" {
		writeError(w, errInvalidParameter, "ConfigurationSetName is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteConfigurationSet(r.Context(), name); err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			status := http.StatusBadRequest
			if sErr.Code == errNotFound {
				status = http.StatusNotFound
			}

			writeError(w, sErr.Code, sErr.Message, status)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// SendEmail handles the SendEmail operation.
func (s *Service) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req SendEmailRequest
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	messageID, err := s.storage.SendEmail(r.Context(), &req)
	if err != nil {
		var sErr *IdentityError
		if errors.As(err, &sErr) {
			writeError(w, sErr.Code, sErr.Message, http.StatusBadRequest)

			return
		}

		writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)

		return
	}

	writeJSONResponse(w, http.StatusOK, SendEmailResponse{
		MessageId: messageID,
	})
}

// Helper functions.

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

// writeJSONResponse writes a JSON response.
func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)

	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Type:    code,
		Message: message,
	})
}

// extractPathParam extracts a path parameter from the URL.
func extractPathParam(path, prefix string) string {
	param, found := strings.CutPrefix(path, prefix)
	if !found {
		return ""
	}

	// Remove any trailing path segments.
	if idx := strings.Index(param, "/"); idx != -1 {
		param = param[:idx]
	}

	return param
}
