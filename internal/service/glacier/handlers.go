// Package glacier provides AWS Glacier service emulation.
package glacier

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// CreateVault handles PUT /-/vaults/{vaultName}.
func (s *Service) CreateVault(w http.ResponseWriter, r *http.Request) {
	vaultName := r.PathValue("vaultName")
	if vaultName == "" {
		writeError(w, errVaultNotFound, "Vault name is required", http.StatusBadRequest)

		return
	}

	vault, err := s.storage.CreateVault(r.Context(), vaultName)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.Header().Set("Location", fmt.Sprintf("/%s/vaults/%s", defaultAccountID, vaultName))
	w.Header().Set("x-amz-glacier-version", "2012-06-01")

	_ = vault

	w.WriteHeader(http.StatusCreated)
}

// DescribeVault handles GET /-/vaults/{vaultName}.
func (s *Service) DescribeVault(w http.ResponseWriter, r *http.Request) {
	vaultName := r.PathValue("vaultName")
	if vaultName == "" {
		writeError(w, errVaultNotFound, "Vault name is required", http.StatusBadRequest)

		return
	}

	vault, err := s.storage.DescribeVault(r.Context(), vaultName)
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, vault)
}

// DeleteVault handles DELETE /-/vaults/{vaultName}.
func (s *Service) DeleteVault(w http.ResponseWriter, r *http.Request) {
	vaultName := r.PathValue("vaultName")
	if vaultName == "" {
		writeError(w, errVaultNotFound, "Vault name is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteVault(r.Context(), vaultName); err != nil {
		handleServiceError(w, err)

		return
	}

	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusNoContent)
}

// ListVaults handles GET /-/vaults.
func (s *Service) ListVaults(w http.ResponseWriter, r *http.Request) {
	vaults, err := s.storage.ListVaults(r.Context())
	if err != nil {
		handleServiceError(w, err)

		return
	}

	writeJSON(w, &ListVaultsResponse{
		VaultList: vaults,
	})
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
		Code:    code,
		Message: message,
		Type:    "client",
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		status := http.StatusBadRequest

		if svcErr.Code == errVaultNotFound {
			status = http.StatusNotFound
		}

		writeError(w, svcErr.Code, svcErr.Message, status)

		return
	}

	writeError(w, "InternalServiceError", err.Error(), http.StatusInternalServerError)
}
