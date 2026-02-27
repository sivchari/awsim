// Package s3control implements the AWS S3 Control service.
package s3control

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

const (
	headerAccountID = "X-Amz-Account-Id"
)

// GetPublicAccessBlock handles the GetPublicAccessBlock API.
func (s *Service) GetPublicAccessBlock(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	config, err := s.storage.GetPublicAccessBlock(r.Context(), accountID)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &GetPublicAccessBlockOutput{
		PublicAccessBlockConfiguration: config,
	}

	writeXMLResponse(w, resp)
}

// PutPublicAccessBlock handles the PutPublicAccessBlock API.
func (s *Service) PutPublicAccessBlock(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	var config PublicAccessBlockConfiguration
	if err := xml.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, ErrInvalidRequest, "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.PutPublicAccessBlock(r.Context(), accountID, &config); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletePublicAccessBlock handles the DeletePublicAccessBlock API.
func (s *Service) DeletePublicAccessBlock(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeletePublicAccessBlock(r.Context(), accountID); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateAccessPoint handles the CreateAccessPoint API.
func (s *Service) CreateAccessPoint(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	name := r.PathValue("name")
	if name == "" {
		writeError(w, ErrInvalidRequest, "Missing access point name", http.StatusBadRequest)

		return
	}

	var input CreateAccessPointInput
	if err := xml.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, ErrInvalidRequest, "Invalid request body", http.StatusBadRequest)

		return
	}

	ap := &AccessPoint{
		Name:              name,
		Bucket:            input.Bucket,
		VpcConfiguration:  input.VpcConfiguration,
		PublicAccessBlock: input.PublicAccessBlockConfiguration,
		BucketAccountID:   input.BucketAccountID,
	}

	result, err := s.storage.CreateAccessPoint(r.Context(), accountID, ap)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &CreateAccessPointResult{
		AccessPointArn: result.AccessPointArn,
		Alias:          result.Alias,
	}

	writeXMLResponse(w, resp)
}

// GetAccessPoint handles the GetAccessPoint API.
func (s *Service) GetAccessPoint(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	name := r.PathValue("name")
	if name == "" {
		writeError(w, ErrInvalidRequest, "Missing access point name", http.StatusBadRequest)

		return
	}

	ap, err := s.storage.GetAccessPoint(r.Context(), accountID, name)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &GetAccessPointResult{
		Name:                           ap.Name,
		Bucket:                         ap.Bucket,
		NetworkOrigin:                  ap.NetworkOrigin,
		VpcConfiguration:               ap.VpcConfiguration,
		PublicAccessBlockConfiguration: ap.PublicAccessBlock,
		AccessPointArn:                 ap.AccessPointArn,
		Alias:                          ap.Alias,
		BucketAccountID:                ap.BucketAccountID,
	}

	if len(ap.Endpoints) > 0 {
		resp.Endpoints = &Endpoints{}

		for k, v := range ap.Endpoints {
			resp.Endpoints.Entry = append(resp.Endpoints.Entry, EndpointEntry{
				Key:   k,
				Value: v,
			})
		}
	}

	writeXMLResponse(w, resp)
}

// DeleteAccessPoint handles the DeleteAccessPoint API.
func (s *Service) DeleteAccessPoint(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	name := r.PathValue("name")
	if name == "" {
		writeError(w, ErrInvalidRequest, "Missing access point name", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteAccessPoint(r.Context(), accountID, name); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListAccessPoints handles the ListAccessPoints API.
func (s *Service) ListAccessPoints(w http.ResponseWriter, r *http.Request) {
	accountID := r.Header.Get(headerAccountID)
	if accountID == "" {
		writeError(w, ErrInvalidRequest, "Missing x-amz-account-id header", http.StatusBadRequest)

		return
	}

	bucket := r.URL.Query().Get("bucket")
	nextToken := r.URL.Query().Get("nextToken")
	maxResults := 1000

	if maxResultsStr := r.URL.Query().Get("maxResults"); maxResultsStr != "" {
		if val, err := strconv.Atoi(maxResultsStr); err == nil {
			maxResults = val
		}
	}

	accessPoints, newNextToken, err := s.storage.ListAccessPoints(r.Context(), accountID, bucket, maxResults, nextToken)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &ListAccessPointsResult{}

	for _, ap := range accessPoints {
		item := AccessPointListItem{
			Name:             ap.Name,
			NetworkOrigin:    ap.NetworkOrigin,
			VpcConfiguration: ap.VpcConfiguration,
			Bucket:           ap.Bucket,
			AccessPointArn:   ap.AccessPointArn,
			Alias:            ap.Alias,
			BucketAccountID:  ap.BucketAccountID,
		}
		resp.AccessPointList = append(resp.AccessPointList, item)
	}

	resp.NextToken = newNextToken

	writeXMLResponse(w, resp)
}

// handleError handles errors and writes appropriate response.
func handleError(w http.ResponseWriter, err error) {
	var s3Err *Error
	if errors.As(err, &s3Err) {
		status := http.StatusBadRequest

		switch s3Err.Code {
		case ErrNoSuchAccessPoint, ErrNoSuchPublicAccessBlockConfiguration:
			status = http.StatusNotFound
		case ErrAccessPointAlreadyOwnedByYou:
			status = http.StatusConflict
		case ErrInternalError:
			status = http.StatusInternalServerError
		}

		writeError(w, s3Err.Code, s3Err.Message, status)

		return
	}

	writeError(w, ErrInternalError, "Internal server error", http.StatusInternalServerError)
}

// writeXMLResponse writes an XML response with status 200 OK.
func writeXMLResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("X-Amz-Request-Id", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = xml.NewEncoder(w).Encode(v)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("X-Amz-Request-Id", uuid.New().String())
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(&Error{
		Code:      code,
		Message:   message,
		RequestID: uuid.New().String(),
	})
}
