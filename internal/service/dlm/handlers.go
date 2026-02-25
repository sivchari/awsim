package dlm

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// CreateLifecyclePolicy handles the CreateLifecyclePolicy API.
func (s *Service) CreateLifecyclePolicy(w http.ResponseWriter, r *http.Request) {
	var req CreateLifecyclePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errInvalidRequest, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.Description == "" {
		writeError(w, errInvalidRequest, "Description is required", http.StatusBadRequest)

		return
	}

	if req.ExecutionRoleArn == "" {
		writeError(w, errInvalidRequest, "ExecutionRoleArn is required", http.StatusBadRequest)

		return
	}

	if req.State == "" {
		writeError(w, errInvalidRequest, "State is required", http.StatusBadRequest)

		return
	}

	policy, err := s.storage.CreateLifecyclePolicy(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &CreateLifecyclePolicyResponse{
		PolicyID: policy.PolicyID,
	})
}

// GetLifecyclePolicy handles the GetLifecyclePolicy API.
func (s *Service) GetLifecyclePolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("policyId")
	if policyID == "" {
		writeError(w, errInvalidRequest, "PolicyId is required", http.StatusBadRequest)

		return
	}

	policy, err := s.storage.GetLifecyclePolicy(r.Context(), policyID)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &GetLifecyclePolicyResponse{
		Policy: policy,
	})
}

// GetLifecyclePolicies handles the GetLifecyclePolicies API.
func (s *Service) GetLifecyclePolicies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var policyIDs []string

	if ids := query.Get("policyIds"); ids != "" {
		policyIDs = strings.Split(ids, ",")
	}

	state := query.Get("state")

	var resourceTypes []string

	if rt := query.Get("resourceTypes"); rt != "" {
		resourceTypes = strings.Split(rt, ",")
	}

	var targetTags []string

	if tt := query.Get("targetTags"); tt != "" {
		targetTags = strings.Split(tt, ",")
	}

	policies, err := s.storage.GetLifecyclePolicies(r.Context(), policyIDs, state, resourceTypes, targetTags)
	if err != nil {
		handleError(w, err)

		return
	}

	summaries := make([]LifecyclePolicySummary, 0, len(policies))
	for _, p := range policies {
		summaries = append(summaries, *p)
	}

	writeResponse(w, &GetLifecyclePoliciesResponse{
		Policies: summaries,
	})
}

// UpdateLifecyclePolicy handles the UpdateLifecyclePolicy API.
func (s *Service) UpdateLifecyclePolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("policyId")
	if policyID == "" {
		writeError(w, errInvalidRequest, "PolicyId is required", http.StatusBadRequest)

		return
	}

	var req UpdateLifecyclePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errInvalidRequest, "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.UpdateLifecyclePolicy(r.Context(), policyID, &req); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, struct{}{})
}

// DeleteLifecyclePolicy handles the DeleteLifecyclePolicy API.
func (s *Service) DeleteLifecyclePolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("policyId")
	if policyID == "" {
		writeError(w, errInvalidRequest, "PolicyId is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteLifecyclePolicy(r.Context(), policyID); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, struct{}{})
}

// writeResponse writes a JSON response.
func writeResponse(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.Header().Set("x-amzn-ErrorType", code)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{
		Message: message,
		Code:    code,
	})
}

// handleError handles service errors.
func handleError(w http.ResponseWriter, err error) {
	var sErr *Error
	if errors.As(err, &sErr) {
		status := getErrorStatus(sErr.Code)
		writeError(w, sErr.Code, sErr.Message, status)

		return
	}

	writeError(w, errInternalServerError, err.Error(), http.StatusInternalServerError)
}

// getErrorStatus returns the HTTP status code for a given error code.
func getErrorStatus(code string) int {
	switch code {
	case errResourceNotFound:
		return http.StatusNotFound
	case errInvalidRequest:
		return http.StatusBadRequest
	case errLimitExceeded:
		return http.StatusTooManyRequests
	case errInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
