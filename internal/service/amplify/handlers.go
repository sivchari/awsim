// Package amplify implements the AWS Amplify service handlers.
package amplify

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

const errValidation = "BadRequestException"

// CreateApp handles POST /apps.
func (s *Service) CreateApp(w http.ResponseWriter, r *http.Request) {
	var req CreateAppInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errValidation, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.Name == "" {
		writeError(w, errValidation, "Name is required", http.StatusBadRequest)

		return
	}

	app, err := s.storage.CreateApp(r.Context(), &req)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, AppResponse{App: app})
}

// GetApp handles GET /apps/{appId}.
func (s *Service) GetApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	if appID == "" {
		writeError(w, errValidation, "appId is required", http.StatusBadRequest)

		return
	}

	app, err := s.storage.GetApp(r.Context(), appID)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, AppResponse{App: app})
}

// ListApps handles GET /apps.
func (s *Service) ListApps(w http.ResponseWriter, r *http.Request) {
	apps, err := s.storage.ListApps(r.Context())
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, AppsResponse{Apps: apps})
}

// UpdateApp handles POST /apps/{appId}.
func (s *Service) UpdateApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	if appID == "" {
		writeError(w, errValidation, "appId is required", http.StatusBadRequest)

		return
	}

	var req UpdateAppInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errValidation, "Invalid request body", http.StatusBadRequest)

		return
	}

	app, err := s.storage.UpdateApp(r.Context(), appID, &req)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, AppResponse{App: app})
}

// DeleteApp handles DELETE /apps/{appId}.
func (s *Service) DeleteApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	if appID == "" {
		writeError(w, errValidation, "appId is required", http.StatusBadRequest)

		return
	}

	app, err := s.storage.DeleteApp(r.Context(), appID)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, AppResponse{App: app})
}

// CreateBranch handles POST /apps/{appId}/branches.
func (s *Service) CreateBranch(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	if appID == "" {
		writeError(w, errValidation, "appId is required", http.StatusBadRequest)

		return
	}

	var req CreateBranchInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errValidation, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.BranchName == "" {
		writeError(w, errValidation, "branchName is required", http.StatusBadRequest)

		return
	}

	branch, err := s.storage.CreateBranch(r.Context(), appID, &req)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, BranchResponse{Branch: branch})
}

// GetBranch handles GET /apps/{appId}/branches/{branchName}.
func (s *Service) GetBranch(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	branchName := r.PathValue("branchName")

	if appID == "" || branchName == "" {
		writeError(w, errValidation, "appId and branchName are required", http.StatusBadRequest)

		return
	}

	branch, err := s.storage.GetBranch(r.Context(), appID, branchName)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, BranchResponse{Branch: branch})
}

// ListBranches handles GET /apps/{appId}/branches.
func (s *Service) ListBranches(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	if appID == "" {
		writeError(w, errValidation, "appId is required", http.StatusBadRequest)

		return
	}

	branches, err := s.storage.ListBranches(r.Context(), appID)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, BranchesResponse{Branches: branches})
}

// DeleteBranch handles DELETE /apps/{appId}/branches/{branchName}.
func (s *Service) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	branchName := r.PathValue("branchName")

	if appID == "" || branchName == "" {
		writeError(w, errValidation, "appId and branchName are required", http.StatusBadRequest)

		return
	}

	branch, err := s.storage.DeleteBranch(r.Context(), appID, branchName)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, BranchResponse{Branch: branch})
}

// Helper functions.

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.Header().Set("x-amzn-ErrorType", code)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		writeError(w, svcErr.Code, svcErr.Message, svcErr.Status)

		return
	}

	writeError(w, "InternalFailureException", err.Error(), http.StatusInternalServerError)
}
