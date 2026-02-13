package acm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// handleRequest routes ACM requests based on the X-Amz-Target header.
func (s *Service) handleRequest(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	if target == "" {
		writeError(w, errInvalidParameter, "Missing X-Amz-Target header", http.StatusBadRequest)

		return
	}

	// Extract operation from target (e.g., "CertificateManager.RequestCertificate")
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		writeError(w, errInvalidParameter, "Invalid X-Amz-Target header", http.StatusBadRequest)

		return
	}

	operation := parts[1]

	switch operation {
	case "RequestCertificate":
		s.RequestCertificate(w, r)
	case "DescribeCertificate":
		s.DescribeCertificate(w, r)
	case "ListCertificates":
		s.ListCertificates(w, r)
	case "DeleteCertificate":
		s.DeleteCertificate(w, r)
	case "GetCertificate":
		s.GetCertificate(w, r)
	case "ImportCertificate":
		s.ImportCertificate(w, r)
	default:
		writeError(w, errInvalidParameter, fmt.Sprintf("Unknown operation: %s", operation), http.StatusBadRequest)
	}
}

// RequestCertificate handles the RequestCertificate operation.
func (s *Service) RequestCertificate(w http.ResponseWriter, r *http.Request) {
	var req RequestCertificateInput
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	cert, err := s.storage.RequestCertificate(r.Context(), &req)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	writeJSONResponse(w, RequestCertificateOutput{
		CertificateArn: cert.CertificateArn,
	})
}

// DescribeCertificate handles the DescribeCertificate operation.
func (s *Service) DescribeCertificate(w http.ResponseWriter, r *http.Request) {
	var req DescribeCertificateInput
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.CertificateArn == "" {
		writeError(w, errInvalidParameter, "CertificateArn is required", http.StatusBadRequest)

		return
	}

	cert, err := s.storage.DescribeCertificate(r.Context(), req.CertificateArn)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	createdAt := cert.CreatedAt
	detail := &CertificateDetail{
		CertificateArn:          cert.CertificateArn,
		DomainName:              cert.DomainName,
		SubjectAlternativeNames: cert.SubjectAlternativeNames,
		DomainValidationOptions: cert.DomainValidationOptions,
		Serial:                  cert.Serial,
		Subject:                 cert.Subject,
		Issuer:                  cert.Issuer,
		CreatedAt:               &createdAt,
		IssuedAt:                cert.IssuedAt,
		ImportedAt:              cert.ImportedAt,
		Status:                  cert.Status,
		NotBefore:               cert.NotBefore,
		NotAfter:                cert.NotAfter,
		KeyAlgorithm:            cert.KeyAlgorithm,
		SignatureAlgorithm:      cert.SignatureAlgorithm,
		InUseBy:                 cert.InUseBy,
		FailureReason:           cert.FailureReason,
		Type:                    cert.Type,
		KeyUsages:               cert.KeyUsages,
		ExtendedKeyUsages:       cert.ExtendedKeyUsages,
		RenewalEligibility:      cert.RenewalEligibility,
		Options:                 cert.Options,
	}

	writeJSONResponse(w, DescribeCertificateOutput{
		Certificate: detail,
	})
}

// ListCertificates handles the ListCertificates operation.
func (s *Service) ListCertificates(w http.ResponseWriter, r *http.Request) {
	var req ListCertificatesInput
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	certs, nextToken, err := s.storage.ListCertificates(r.Context(), req.CertificateStatuses, req.MaxItems, req.NextToken)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	summaries := make([]CertificateSummary, 0, len(certs))

	for _, cert := range certs {
		createdAt := cert.CreatedAt
		summary := CertificateSummary{
			CertificateArn:          cert.CertificateArn,
			DomainName:              cert.DomainName,
			SubjectAlternativeNames: cert.SubjectAlternativeNames,
			Status:                  cert.Status,
			Type:                    cert.Type,
			KeyAlgorithm:            cert.KeyAlgorithm,
			CreatedAt:               &createdAt,
			IssuedAt:                cert.IssuedAt,
			ImportedAt:              cert.ImportedAt,
			NotBefore:               cert.NotBefore,
			NotAfter:                cert.NotAfter,
			RenewalEligibility:      cert.RenewalEligibility,
			InUse:                   len(cert.InUseBy) > 0,
		}

		summaries = append(summaries, summary)
	}

	writeJSONResponse(w, ListCertificatesOutput{
		CertificateSummaryList: summaries,
		NextToken:              nextToken,
	})
}

// DeleteCertificate handles the DeleteCertificate operation.
func (s *Service) DeleteCertificate(w http.ResponseWriter, r *http.Request) {
	var req DeleteCertificateInput
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.CertificateArn == "" {
		writeError(w, errInvalidParameter, "CertificateArn is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteCertificate(r.Context(), req.CertificateArn); err != nil {
		handleStorageError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
}

// GetCertificate handles the GetCertificate operation.
func (s *Service) GetCertificate(w http.ResponseWriter, r *http.Request) {
	var req GetCertificateInput
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.CertificateArn == "" {
		writeError(w, errInvalidParameter, "CertificateArn is required", http.StatusBadRequest)

		return
	}

	cert, err := s.storage.GetCertificate(r.Context(), req.CertificateArn)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	writeJSONResponse(w, GetCertificateOutput{
		Certificate:      cert.CertificateBody,
		CertificateChain: cert.CertificateChain,
	})
}

// ImportCertificate handles the ImportCertificate operation.
func (s *Service) ImportCertificate(w http.ResponseWriter, r *http.Request) {
	var req ImportCertificateInput
	if err := readJSONRequest(r, &req); err != nil {
		writeError(w, errInvalidParameter, "Invalid request body", http.StatusBadRequest)

		return
	}

	cert, err := s.storage.ImportCertificate(r.Context(), &req)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	writeJSONResponse(w, ImportCertificateOutput{
		CertificateArn: cert.CertificateArn,
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

// writeJSONResponse writes a JSON response with HTTP 200 OK.
func writeJSONResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)

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

// handleStorageError handles storage errors and writes appropriate response.
func handleStorageError(w http.ResponseWriter, err error) {
	var acmErr *Error
	if errors.As(err, &acmErr) {
		status := http.StatusBadRequest
		if acmErr.Code == errNotFound {
			status = http.StatusNotFound
		}

		writeError(w, acmErr.Code, acmErr.Message, status)

		return
	}

	writeError(w, "InternalServiceError", "Internal server error", http.StatusInternalServerError)
}
