package route53

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateHostedZone handles the CreateHostedZone API.
func (s *Service) CreateHostedZone(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Failed to read request body")
		return
	}

	var req CreateHostedZoneRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Failed to parse request body")
		return
	}

	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Name is required")
		return
	}

	if req.CallerReference == "" {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "CallerReference is required")
		return
	}

	// Ensure name ends with a dot
	name := req.Name
	if !strings.HasSuffix(name, ".") {
		name = name + "."
	}

	zoneID := uuid.New().String()
	zone := &HostedZone{
		ID:                     "/hostedzone/" + zoneID,
		Name:                   name,
		CallerReference:        req.CallerReference,
		Config:                 req.HostedZoneConfig,
		ResourceRecordSetCount: 0,
	}

	if err := s.storage.CreateHostedZone(zone); err != nil {
		if errors.Is(err, ErrHostedZoneAlreadyExists) {
			writeErrorResponse(w, http.StatusConflict, "HostedZoneAlreadyExists", "Hosted zone already exists")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "InternalError", err.Error())
		return
	}

	resp := CreateHostedZoneResponse{
		XMLNS:      xmlns,
		HostedZone: *zone,
		ChangeInfo: ChangeInfo{
			ID:          "/change/" + uuid.New().String(),
			Status:      "INSYNC",
			SubmittedAt: time.Now().UTC().Format(time.RFC3339),
		},
		DelegationSet: DelegationSet{
			NameServers: []string{
				"ns-1.awsim.local",
				"ns-2.awsim.local",
				"ns-3.awsim.local",
				"ns-4.awsim.local",
			},
		},
	}

	w.Header().Set("Location", fmt.Sprintf("https://route53.amazonaws.com/2013-04-01%s", zone.ID))
	writeXMLResponse(w, http.StatusCreated, resp)
}

// GetHostedZone handles the GetHostedZone API.
func (s *Service) GetHostedZone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Hosted zone ID is required")
		return
	}

	zoneID := "/hostedzone/" + id
	zone, err := s.storage.GetHostedZone(zoneID)
	if err != nil {
		if errors.Is(err, ErrHostedZoneNotFound) {
			writeErrorResponse(w, http.StatusNotFound, "NoSuchHostedZone", "Hosted zone not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "InternalError", err.Error())
		return
	}

	resp := GetHostedZoneResponse{
		XMLNS:      xmlns,
		HostedZone: *zone,
		DelegationSet: DelegationSet{
			NameServers: []string{
				"ns-1.awsim.local",
				"ns-2.awsim.local",
				"ns-3.awsim.local",
				"ns-4.awsim.local",
			},
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// ListHostedZones handles the ListHostedZones API.
func (s *Service) ListHostedZones(w http.ResponseWriter, r *http.Request) {
	zones, err := s.storage.ListHostedZones()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "InternalError", err.Error())
		return
	}

	hostedZones := make([]HostedZone, 0, len(zones))
	for _, z := range zones {
		hostedZones = append(hostedZones, *z)
	}

	resp := ListHostedZonesResponse{
		XMLNS:       xmlns,
		HostedZones: hostedZones,
		IsTruncated: false,
		MaxItems:    "100",
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// DeleteHostedZone handles the DeleteHostedZone API.
func (s *Service) DeleteHostedZone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Hosted zone ID is required")
		return
	}

	zoneID := "/hostedzone/" + id
	if err := s.storage.DeleteHostedZone(zoneID); err != nil {
		if errors.Is(err, ErrHostedZoneNotFound) {
			writeErrorResponse(w, http.StatusNotFound, "NoSuchHostedZone", "Hosted zone not found")
			return
		}
		if errors.Is(err, ErrHostedZoneNotEmpty) {
			writeErrorResponse(w, http.StatusBadRequest, "HostedZoneNotEmpty", "Hosted zone is not empty")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "InternalError", err.Error())
		return
	}

	resp := DeleteHostedZoneResponse{
		XMLNS: xmlns,
		ChangeInfo: ChangeInfo{
			ID:          "/change/" + uuid.New().String(),
			Status:      "INSYNC",
			SubmittedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// ChangeResourceRecordSets handles the ChangeResourceRecordSets API.
func (s *Service) ChangeResourceRecordSets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Hosted zone ID is required")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Failed to read request body")
		return
	}

	var req ChangeResourceRecordSetsRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Failed to parse request body")
		return
	}

	if len(req.ChangeBatch.Changes) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "At least one change is required")
		return
	}

	zoneID := "/hostedzone/" + id
	if err := s.storage.ChangeRecordSets(zoneID, req.ChangeBatch.Changes); err != nil {
		if errors.Is(err, ErrHostedZoneNotFound) {
			writeErrorResponse(w, http.StatusNotFound, "NoSuchHostedZone", "Hosted zone not found")
			return
		}
		if errors.Is(err, ErrRecordSetAlreadyExists) {
			writeErrorResponse(w, http.StatusBadRequest, "ResourceRecordAlreadyExists", "Resource record already exists")
			return
		}
		if errors.Is(err, ErrRecordSetNotFound) {
			writeErrorResponse(w, http.StatusBadRequest, "InvalidChangeBatch", "Resource record not found")
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Invalid change action")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "InternalError", err.Error())
		return
	}

	resp := ChangeResourceRecordSetsResponse{
		XMLNS: xmlns,
		ChangeInfo: ChangeInfo{
			ID:          "/change/" + uuid.New().String(),
			Status:      "INSYNC",
			SubmittedAt: time.Now().UTC().Format(time.RFC3339),
			Comment:     req.ChangeBatch.Comment,
		},
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// ListResourceRecordSets handles the ListResourceRecordSets API.
func (s *Service) ListResourceRecordSets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "InvalidInput", "Hosted zone ID is required")
		return
	}

	zoneID := "/hostedzone/" + id
	records, err := s.storage.GetRecordSets(zoneID)
	if err != nil {
		if errors.Is(err, ErrHostedZoneNotFound) {
			writeErrorResponse(w, http.StatusNotFound, "NoSuchHostedZone", "Hosted zone not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "InternalError", err.Error())
		return
	}

	resp := ListResourceRecordSetsResponse{
		XMLNS:              xmlns,
		ResourceRecordSets: records,
		IsTruncated:        false,
		MaxItems:           "100",
	}

	writeXMLResponse(w, http.StatusOK, resp)
}

// writeXMLResponse writes an XML response.
func writeXMLResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	io.WriteString(w, xml.Header)
	xml.NewEncoder(w).Encode(v)
}

// writeErrorResponse writes an error response.
func writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{
		XMLNS: xmlns,
		Error: Error{
			Type:    "Sender",
			Code:    code,
			Message: message,
		},
		RequestId: uuid.New().String(),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amzn-RequestId", resp.RequestId)
	w.WriteHeader(status)
	io.WriteString(w, xml.Header)
	xml.NewEncoder(w).Encode(resp)
}
