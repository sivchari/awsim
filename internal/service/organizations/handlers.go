package organizations

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handlerFunc is a type alias for handler functions.
type handlerFunc func(http.ResponseWriter, *http.Request)

// getActionHandlers returns a map of action names to handler functions.
func (s *Service) getActionHandlers() map[string]handlerFunc {
	return map[string]handlerFunc{
		// Organization operations
		"CreateOrganization":   s.CreateOrganization,
		"DeleteOrganization":   s.DeleteOrganization,
		"DescribeOrganization": s.DescribeOrganization,
		// Account operations
		"CreateAccount":   s.CreateAccount,
		"DescribeAccount": s.DescribeAccount,
		"ListAccounts":    s.ListAccounts,
		// Organizational unit operations
		"CreateOrganizationalUnit":         s.CreateOrganizationalUnit,
		"ListOrganizationalUnitsForParent": s.ListOrganizationalUnitsForParent,
		// Policy operations
		"AttachPolicy": s.AttachPolicy,
		"DetachPolicy": s.DetachPolicy,
		// Root operations
		"ListRoots": s.ListRoots,
	}
}

// DispatchAction dispatches the request to the appropriate handler.
func (s *Service) DispatchAction(w http.ResponseWriter, r *http.Request) {
	target := r.Header.Get("X-Amz-Target")
	action := strings.TrimPrefix(target, "AWSOrganizationsV20161128.")

	handlers := s.getActionHandlers()
	if handler, ok := handlers[action]; ok {
		handler(w, r)

		return
	}

	writeError(w, "UnknownOperationException", "The operation "+action+" is not valid.", http.StatusBadRequest)
}

// CreateOrganization handles the CreateOrganization action.
func (s *Service) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var input CreateOrganizationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	org, err := s.storage.CreateOrganization(r.Context(), input.FeatureSet)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, CreateOrganizationOutput{Organization: org})
}

// DeleteOrganization handles the DeleteOrganization action.
func (s *Service) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	if err := s.storage.DeleteOrganization(r.Context()); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// DescribeOrganization handles the DescribeOrganization action.
func (s *Service) DescribeOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := s.storage.DescribeOrganization(r.Context())
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, DescribeOrganizationOutput{Organization: org})
}

// CreateAccount handles the CreateAccount action.
func (s *Service) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var input CreateAccountInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	status, err := s.storage.CreateAccount(r.Context(), &input)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, CreateAccountOutput{CreateAccountStatus: status})
}

// DescribeAccount handles the DescribeAccount action.
func (s *Service) DescribeAccount(w http.ResponseWriter, r *http.Request) {
	var input DescribeAccountInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	account, err := s.storage.DescribeAccount(r.Context(), input.AccountID)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, DescribeAccountOutput{Account: account})
}

// ListAccounts handles the ListAccounts action.
func (s *Service) ListAccounts(w http.ResponseWriter, r *http.Request) {
	var input ListAccountsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	accounts, nextToken, err := s.storage.ListAccounts(r.Context(), input.MaxResults, input.NextToken)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, ListAccountsOutput{Accounts: accounts, NextToken: nextToken})
}

// CreateOrganizationalUnit handles the CreateOrganizationalUnit action.
func (s *Service) CreateOrganizationalUnit(w http.ResponseWriter, r *http.Request) {
	var input CreateOrganizationalUnitInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	ou, err := s.storage.CreateOrganizationalUnit(r.Context(), input.Name, input.ParentID)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, CreateOrganizationalUnitOutput{OrganizationalUnit: ou})
}

// ListOrganizationalUnitsForParent handles the ListOrganizationalUnitsForParent action.
func (s *Service) ListOrganizationalUnitsForParent(w http.ResponseWriter, r *http.Request) {
	var input ListOrganizationalUnitsForParentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	ous, nextToken, err := s.storage.ListOrganizationalUnitsForParent(r.Context(), input.ParentID, input.MaxResults, input.NextToken)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, ListOrganizationalUnitsForParentOutput{OrganizationalUnits: ous, NextToken: nextToken})
}

// AttachPolicy handles the AttachPolicy action.
func (s *Service) AttachPolicy(w http.ResponseWriter, r *http.Request) {
	var input AttachPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.AttachPolicy(r.Context(), input.PolicyID, input.TargetID); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// DetachPolicy handles the DetachPolicy action.
func (s *Service) DetachPolicy(w http.ResponseWriter, r *http.Request) {
	var input DetachPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := s.storage.DetachPolicy(r.Context(), input.PolicyID, input.TargetID); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListRoots handles the ListRoots action.
func (s *Service) ListRoots(w http.ResponseWriter, r *http.Request) {
	var input ListRootsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errInvalidInputException, "Invalid request body", http.StatusBadRequest)

		return
	}

	roots, nextToken, err := s.storage.ListRoots(r.Context(), input.MaxResults, input.NextToken)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, ListRootsOutput{Roots: roots, NextToken: nextToken})
}

// Helper functions.

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.WriteHeader(statusCode)

	errResp := &Error{
		Code:    code,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleError(w http.ResponseWriter, err error) {
	if e, ok := err.(*Error); ok {
		statusCode := http.StatusBadRequest

		switch e.Code {
		case errAWSOrganizationsNotInUseException:
			statusCode = http.StatusBadRequest
		case errAccountNotFoundException, errParentNotFoundException, errPolicyNotFoundException:
			statusCode = http.StatusNotFound
		case errAlreadyInOrganizationException, errOrganizationNotEmptyException, errDuplicateOrganizationalUnitException, errDuplicatePolicyAttachmentException:
			statusCode = http.StatusConflict
		}

		writeError(w, e.Code, e.Message, statusCode)

		return
	}

	writeError(w, "InternalServiceException", err.Error(), http.StatusInternalServerError)
}
