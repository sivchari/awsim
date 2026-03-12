// Package ebs provides AWS EBS direct API service emulation.
package ebs

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

const errValidation = "ValidationException"

// StartSnapshotHandler handles POST /snapshots.
func (s *Service) StartSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	var req StartSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errValidation, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.VolumeSize <= 0 {
		writeError(w, errValidation, "VolumeSize is required and must be greater than 0", http.StatusBadRequest)

		return
	}

	snapshot, err := s.storage.StartSnapshot(r.Context(), &req)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(snapshot)
}

// CompleteSnapshotHandler handles POST /snapshots/completion/{snapshotId}.
func (s *Service) CompleteSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	snapshotID := r.PathValue("snapshotId")
	if snapshotID == "" {
		writeError(w, errValidation, "SnapshotId is required", http.StatusBadRequest)

		return
	}

	snapshot, err := s.storage.CompleteSnapshot(r.Context(), snapshotID)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, snapshot)
}

// ListSnapshotBlocksHandler handles GET /snapshots/{snapshotId}/blocks.
func (s *Service) ListSnapshotBlocksHandler(w http.ResponseWriter, r *http.Request) {
	snapshotID := r.PathValue("snapshotId")
	if snapshotID == "" {
		writeError(w, errValidation, "SnapshotId is required", http.StatusBadRequest)

		return
	}

	result, err := s.storage.ListSnapshotBlocks(r.Context(), snapshotID)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, result)
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
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{
		Message: message,
		Reason:  code,
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		status := http.StatusBadRequest

		if svcErr.Code == errSnapshotNotFound {
			status = http.StatusNotFound
		}

		writeError(w, svcErr.Code, svcErr.Message, status)

		return
	}

	writeError(w, "InternalException", err.Error(), http.StatusInternalServerError)
}
