package ec2

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

const ec2XMLNS = "http://ec2.amazonaws.com/doc/2016-11-15/"

// Error codes for EC2.
const (
	errInvalidParameter = "InvalidParameterValue"
	errInternalError    = "InternalError"
	errInvalidAction    = "InvalidAction"
)

// RunInstances handles the RunInstances action.
func (s *Service) RunInstances(w http.ResponseWriter, r *http.Request) {
	var req RunInstancesRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.ImageID == "" {
		writeEC2Error(w, errInvalidParameter, "ImageId is required", http.StatusBadRequest)

		return
	}

	instances, reservationID, err := s.storage.RunInstances(r.Context(), &req)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	xmlInstances := make([]XMLInstance, 0, len(instances))
	for _, inst := range instances {
		xmlInstances = append(xmlInstances, convertToXMLInstance(inst))
	}

	writeEC2XMLResponse(w, XMLRunInstancesResponse{
		Xmlns:         ec2XMLNS,
		RequestID:     uuid.New().String(),
		ReservationID: reservationID,
		OwnerID:       defaultAccountID,
		InstancesSet:  XMLInstancesSet{Items: xmlInstances},
	})
}

// TerminateInstances handles the TerminateInstances action.
func (s *Service) TerminateInstances(w http.ResponseWriter, r *http.Request) {
	var req TerminateInstancesRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if len(req.InstanceIDs) == 0 {
		writeEC2Error(w, errInvalidParameter, "InstanceIds is required", http.StatusBadRequest)

		return
	}

	changes, err := s.storage.TerminateInstances(r.Context(), req.InstanceIDs)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLTerminateInstancesResponse{
		Xmlns:        ec2XMLNS,
		RequestID:    uuid.New().String(),
		InstancesSet: convertToXMLInstanceStateChangeSet(changes),
	})
}

// DescribeInstances handles the DescribeInstances action.
func (s *Service) DescribeInstances(w http.ResponseWriter, r *http.Request) {
	var req DescribeInstancesRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	reservations, err := s.storage.DescribeInstances(r.Context(), req.InstanceIDs)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	xmlReservations := make([]XMLReservation, 0, len(reservations))
	for _, res := range reservations {
		xmlInstances := make([]XMLInstance, 0, len(res.Instances))
		for _, inst := range res.Instances {
			xmlInstances = append(xmlInstances, convertToXMLInstance(inst))
		}

		xmlReservations = append(xmlReservations, XMLReservation{
			ReservationID: res.ReservationID,
			OwnerID:       res.OwnerID,
			InstancesSet:  XMLInstancesSet{Items: xmlInstances},
		})
	}

	writeEC2XMLResponse(w, XMLDescribeInstancesResponse{
		Xmlns:          ec2XMLNS,
		RequestID:      uuid.New().String(),
		ReservationSet: XMLReservationSet{Items: xmlReservations},
	})
}

// StartInstances handles the StartInstances action.
func (s *Service) StartInstances(w http.ResponseWriter, r *http.Request) {
	var req StartInstancesRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if len(req.InstanceIDs) == 0 {
		writeEC2Error(w, errInvalidParameter, "InstanceIds is required", http.StatusBadRequest)

		return
	}

	changes, err := s.storage.StartInstances(r.Context(), req.InstanceIDs)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLStartInstancesResponse{
		Xmlns:        ec2XMLNS,
		RequestID:    uuid.New().String(),
		InstancesSet: convertToXMLInstanceStateChangeSet(changes),
	})
}

// StopInstances handles the StopInstances action.
func (s *Service) StopInstances(w http.ResponseWriter, r *http.Request) {
	var req StopInstancesRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if len(req.InstanceIDs) == 0 {
		writeEC2Error(w, errInvalidParameter, "InstanceIds is required", http.StatusBadRequest)

		return
	}

	changes, err := s.storage.StopInstances(r.Context(), req.InstanceIDs)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLStopInstancesResponse{
		Xmlns:        ec2XMLNS,
		RequestID:    uuid.New().String(),
		InstancesSet: convertToXMLInstanceStateChangeSet(changes),
	})
}

// CreateSecurityGroup handles the CreateSecurityGroup action.
func (s *Service) CreateSecurityGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateSecurityGroupRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.GroupName == "" {
		writeEC2Error(w, errInvalidParameter, "GroupName is required", http.StatusBadRequest)

		return
	}

	if req.Description == "" {
		writeEC2Error(w, errInvalidParameter, "Description is required", http.StatusBadRequest)

		return
	}

	sg, err := s.storage.CreateSecurityGroup(r.Context(), &req)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLCreateSecurityGroupResponse{
		Xmlns:     ec2XMLNS,
		RequestID: uuid.New().String(),
		Return:    true,
		GroupID:   sg.GroupID,
	})
}

// DeleteSecurityGroup handles the DeleteSecurityGroup action.
func (s *Service) DeleteSecurityGroup(w http.ResponseWriter, r *http.Request) {
	var req DeleteSecurityGroupRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.GroupID == "" && req.GroupName == "" {
		writeEC2Error(w, errInvalidParameter, "GroupId or GroupName is required", http.StatusBadRequest)

		return
	}

	err := s.storage.DeleteSecurityGroup(r.Context(), req.GroupID, req.GroupName)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLDeleteSecurityGroupResponse{
		Xmlns:     ec2XMLNS,
		RequestID: uuid.New().String(),
		Return:    true,
	})
}

// AuthorizeSecurityGroupIngress handles the AuthorizeSecurityGroupIngress action.
func (s *Service) AuthorizeSecurityGroupIngress(w http.ResponseWriter, r *http.Request) {
	var req AuthorizeSecurityGroupIngressRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.GroupID == "" && req.GroupName == "" {
		writeEC2Error(w, errInvalidParameter, "GroupId or GroupName is required", http.StatusBadRequest)

		return
	}

	err := s.storage.AuthorizeSecurityGroupIngress(r.Context(), req.GroupID, req.GroupName, req.IPPermissions)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLAuthorizeSecurityGroupIngressResponse{
		Xmlns:     ec2XMLNS,
		RequestID: uuid.New().String(),
		Return:    true,
	})
}

// AuthorizeSecurityGroupEgress handles the AuthorizeSecurityGroupEgress action.
func (s *Service) AuthorizeSecurityGroupEgress(w http.ResponseWriter, r *http.Request) {
	var req AuthorizeSecurityGroupEgressRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.GroupID == "" {
		writeEC2Error(w, errInvalidParameter, "GroupId is required", http.StatusBadRequest)

		return
	}

	err := s.storage.AuthorizeSecurityGroupEgress(r.Context(), req.GroupID, req.IPPermissions)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLAuthorizeSecurityGroupEgressResponse{
		Xmlns:     ec2XMLNS,
		RequestID: uuid.New().String(),
		Return:    true,
	})
}

// CreateKeyPair handles the CreateKeyPair action.
func (s *Service) CreateKeyPair(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyPairRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.KeyName == "" {
		writeEC2Error(w, errInvalidParameter, "KeyName is required", http.StatusBadRequest)

		return
	}

	kp, err := s.storage.CreateKeyPair(r.Context(), req.KeyName, req.KeyType)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLCreateKeyPairResponse{
		Xmlns:          ec2XMLNS,
		RequestID:      uuid.New().String(),
		KeyName:        kp.KeyName,
		KeyFingerprint: kp.KeyFingerprint,
		KeyMaterial:    kp.KeyMaterial,
		KeyPairID:      kp.KeyPairID,
	})
}

// DeleteKeyPair handles the DeleteKeyPair action.
func (s *Service) DeleteKeyPair(w http.ResponseWriter, r *http.Request) {
	var req DeleteKeyPairRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	if req.KeyName == "" && req.KeyPairID == "" {
		writeEC2Error(w, errInvalidParameter, "KeyName or KeyPairId is required", http.StatusBadRequest)

		return
	}

	err := s.storage.DeleteKeyPair(r.Context(), req.KeyName, req.KeyPairID)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	writeEC2XMLResponse(w, XMLDeleteKeyPairResponse{
		Xmlns:     ec2XMLNS,
		RequestID: uuid.New().String(),
		Return:    true,
	})
}

// DescribeKeyPairs handles the DescribeKeyPairs action.
func (s *Service) DescribeKeyPairs(w http.ResponseWriter, r *http.Request) {
	var req DescribeKeyPairsRequest
	if err := readEC2JSONRequest(r, &req); err != nil {
		writeEC2Error(w, errInvalidParameter, "Failed to parse request body", http.StatusBadRequest)

		return
	}

	keyPairs, err := s.storage.DescribeKeyPairs(r.Context(), req.KeyNames, req.KeyPairIDs)
	if err != nil {
		handleEC2Error(w, err)

		return
	}

	xmlKeyPairs := make([]XMLKeyPairInfo, 0, len(keyPairs))
	for _, kp := range keyPairs {
		xmlKeyPairs = append(xmlKeyPairs, XMLKeyPairInfo{
			KeyName:        kp.KeyName,
			KeyFingerprint: kp.KeyFingerprint,
			KeyPairID:      kp.KeyPairID,
		})
	}

	writeEC2XMLResponse(w, XMLDescribeKeyPairsResponse{
		Xmlns:     ec2XMLNS,
		RequestID: uuid.New().String(),
		KeySet:    XMLKeyPairSet{Items: xmlKeyPairs},
	})
}

// DispatchAction routes the request to the appropriate handler based on Action parameter.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("Action")

	switch action {
	case "RunInstances":
		s.RunInstances(w, r)
	case "TerminateInstances":
		s.TerminateInstances(w, r)
	case "DescribeInstances":
		s.DescribeInstances(w, r)
	case "StartInstances":
		s.StartInstances(w, r)
	case "StopInstances":
		s.StopInstances(w, r)
	case "CreateSecurityGroup":
		s.CreateSecurityGroup(w, r)
	case "DeleteSecurityGroup":
		s.DeleteSecurityGroup(w, r)
	case "AuthorizeSecurityGroupIngress":
		s.AuthorizeSecurityGroupIngress(w, r)
	case "AuthorizeSecurityGroupEgress":
		s.AuthorizeSecurityGroupEgress(w, r)
	case "CreateKeyPair":
		s.CreateKeyPair(w, r)
	case "DeleteKeyPair":
		s.DeleteKeyPair(w, r)
	case "DescribeKeyPairs":
		s.DescribeKeyPairs(w, r)
	default:
		writeEC2Error(w, errInvalidAction, fmt.Sprintf("The action '%s' is not valid", action), http.StatusBadRequest)
	}
}

// convertToXMLInstance converts an Instance to XMLInstance.
func convertToXMLInstance(inst *Instance) XMLInstance {
	groupSet := make([]XMLGroupIdentifier, 0, len(inst.SecurityGroups))
	for _, sg := range inst.SecurityGroups {
		groupSet = append(groupSet, XMLGroupIdentifier{
			GroupID:   sg.GroupID,
			GroupName: sg.GroupName,
		})
	}

	return XMLInstance{
		InstanceID:       inst.InstanceID,
		ImageID:          inst.ImageID,
		InstanceType:     inst.InstanceType,
		InstanceState:    XMLInstanceState{Code: inst.State.Code, Name: inst.State.Name},
		PrivateIPAddress: inst.PrivateIPAddress,
		IPAddress:        inst.PublicIPAddress,
		KeyName:          inst.KeyName,
		LaunchTime:       inst.LaunchTime.Format("2006-01-02T15:04:05.000Z"),
		GroupSet:         XMLGroupSet{Items: groupSet},
	}
}

// convertToXMLInstanceStateChangeSet converts instance state changes to XML format.
func convertToXMLInstanceStateChangeSet(changes []InstanceStateChange) XMLInstanceStateChangeSet {
	items := make([]XMLInstanceStateChange, 0, len(changes))
	for _, c := range changes {
		items = append(items, XMLInstanceStateChange{
			InstanceID:    c.InstanceID,
			CurrentState:  XMLInstanceState{Code: c.CurrentState.Code, Name: c.CurrentState.Name},
			PreviousState: XMLInstanceState{Code: c.PreviousState.Code, Name: c.PreviousState.Name},
		})
	}

	return XMLInstanceStateChangeSet{Items: items}
}

// readEC2JSONRequest reads and decodes JSON request body.
func readEC2JSONRequest(r *http.Request, v any) error {
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

// writeEC2XMLResponse writes an XML response with HTTP 200 OK.
func writeEC2XMLResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(v)
}

// writeEC2Error writes an EC2 error response in XML format.
func writeEC2Error(w http.ResponseWriter, code, message string, status int) {
	requestID := uuid.New().String()

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("x-amzn-RequestId", requestID)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(XMLErrorResponse{
		Errors: XMLErrors{
			Error: XMLError{
				Code:    code,
				Message: message,
			},
		},
		RequestID: requestID,
	})
}

// handleEC2Error handles EC2 errors and writes the appropriate response.
func handleEC2Error(w http.ResponseWriter, err error) {
	var ec2Err *EC2Error
	if errors.As(err, &ec2Err) {
		writeEC2Error(w, ec2Err.Code, ec2Err.Message, http.StatusBadRequest)

		return
	}

	writeEC2Error(w, errInternalError, "Internal server error", http.StatusInternalServerError)
}
