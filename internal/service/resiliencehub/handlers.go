package resiliencehub

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// CreateApp handles the CreateApp API.
func (s *Service) CreateApp(w http.ResponseWriter, r *http.Request) {
	var req CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Name is required",
		})

		return
	}

	app, err := s.storage.CreateApp(&req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &CreateAppResponse{App: app})
}

// DescribeApp handles the DescribeApp API.
func (s *Service) DescribeApp(w http.ResponseWriter, r *http.Request) {
	appARN := r.URL.Query().Get("appArn")
	if appARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "appArn is required",
		})

		return
	}

	app, err := s.storage.DescribeApp(appARN)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DescribeAppResponse{App: app})
}

// UpdateApp handles the UpdateApp API.
func (s *Service) UpdateApp(w http.ResponseWriter, r *http.Request) {
	var req UpdateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.AppARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "appArn is required",
		})

		return
	}

	app, err := s.storage.UpdateApp(&req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &UpdateAppResponse{App: app})
}

// DeleteApp handles the DeleteApp API.
func (s *Service) DeleteApp(w http.ResponseWriter, r *http.Request) {
	appARN := r.URL.Query().Get("appArn")
	if appARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "appArn is required",
		})

		return
	}

	if err := s.storage.DeleteApp(appARN); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DeleteAppResponse{AppARN: appARN})
}

// ListApps handles the ListApps API.
func (s *Service) ListApps(w http.ResponseWriter, r *http.Request) {
	req := &ListAppsRequest{
		AppARN: r.URL.Query().Get("appArn"),
		Name:   r.URL.Query().Get("name"),
	}

	apps, nextToken, err := s.storage.ListApps(req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &ListAppsResponse{
		AppSummaries: apps,
		NextToken:    nextToken,
	})
}

// CreateResiliencyPolicy handles the CreateResiliencyPolicy API.
func (s *Service) CreateResiliencyPolicy(w http.ResponseWriter, r *http.Request) {
	var req CreateResiliencyPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.PolicyName == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "policyName is required",
		})

		return
	}

	if req.Tier == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "tier is required",
		})

		return
	}

	policy, err := s.storage.CreateResiliencyPolicy(&req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &CreateResiliencyPolicyResponse{Policy: policy})
}

// DescribeResiliencyPolicy handles the DescribeResiliencyPolicy API.
func (s *Service) DescribeResiliencyPolicy(w http.ResponseWriter, r *http.Request) {
	policyARN := r.URL.Query().Get("policyArn")
	if policyARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "policyArn is required",
		})

		return
	}

	policy, err := s.storage.DescribeResiliencyPolicy(policyARN)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DescribeResiliencyPolicyResponse{Policy: policy})
}

// UpdateResiliencyPolicy handles the UpdateResiliencyPolicy API.
func (s *Service) UpdateResiliencyPolicy(w http.ResponseWriter, r *http.Request) {
	var req UpdateResiliencyPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.PolicyARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "policyArn is required",
		})

		return
	}

	policy, err := s.storage.UpdateResiliencyPolicy(&req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &UpdateResiliencyPolicyResponse{Policy: policy})
}

// DeleteResiliencyPolicy handles the DeleteResiliencyPolicy API.
func (s *Service) DeleteResiliencyPolicy(w http.ResponseWriter, r *http.Request) {
	policyARN := r.URL.Query().Get("policyArn")
	if policyARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "policyArn is required",
		})

		return
	}

	if err := s.storage.DeleteResiliencyPolicy(policyARN); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DeleteResiliencyPolicyResponse{PolicyARN: policyARN})
}

// ListResiliencyPolicies handles the ListResiliencyPolicies API.
func (s *Service) ListResiliencyPolicies(w http.ResponseWriter, r *http.Request) {
	req := &ListResiliencyPoliciesRequest{
		PolicyName: r.URL.Query().Get("policyName"),
	}

	policies, nextToken, err := s.storage.ListResiliencyPolicies(req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &ListResiliencyPoliciesResponse{
		ResiliencyPolicies: policies,
		NextToken:          nextToken,
	})
}

// StartAppAssessment handles the StartAppAssessment API.
func (s *Service) StartAppAssessment(w http.ResponseWriter, r *http.Request) {
	var req StartAppAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.AppARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "appArn is required",
		})

		return
	}

	if req.AppVersion == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "appVersion is required",
		})

		return
	}

	if req.AssessmentName == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "assessmentName is required",
		})

		return
	}

	assessment, err := s.storage.StartAppAssessment(&req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &StartAppAssessmentResponse{Assessment: assessment})
}

// DescribeAppAssessment handles the DescribeAppAssessment API.
func (s *Service) DescribeAppAssessment(w http.ResponseWriter, r *http.Request) {
	assessmentARN := r.URL.Query().Get("assessmentArn")
	if assessmentARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "assessmentArn is required",
		})

		return
	}

	assessment, err := s.storage.DescribeAppAssessment(assessmentARN)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DescribeAppAssessmentResponse{Assessment: assessment})
}

// DeleteAppAssessment handles the DeleteAppAssessment API.
func (s *Service) DeleteAppAssessment(w http.ResponseWriter, r *http.Request) {
	assessmentARN := r.URL.Query().Get("assessmentArn")
	if assessmentARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "assessmentArn is required",
		})

		return
	}

	if err := s.storage.DeleteAppAssessment(assessmentARN); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DeleteAppAssessmentResponse{
		AssessmentARN:    assessmentARN,
		AssessmentStatus: "Success",
	})
}

// ListAppAssessments handles the ListAppAssessments API.
func (s *Service) ListAppAssessments(w http.ResponseWriter, r *http.Request) {
	req := &ListAppAssessmentsRequest{
		AppARN:           r.URL.Query().Get("appArn"),
		AssessmentName:   r.URL.Query().Get("assessmentName"),
		AssessmentStatus: r.URL.Query().Get("assessmentStatus"),
		ComplianceStatus: r.URL.Query().Get("complianceStatus"),
		Invoker:          r.URL.Query().Get("invoker"),
	}

	assessments, nextToken, err := s.storage.ListAppAssessments(req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &ListAppAssessmentsResponse{
		AssessmentSummaries: assessments,
		NextToken:           nextToken,
	})
}

// TagResource handles the TagResource API.
func (s *Service) TagResource(w http.ResponseWriter, r *http.Request) {
	resourceARN := r.URL.Query().Get("resourceArn")
	if resourceARN == "" {
		// Try to extract from path
		parts := strings.Split(r.URL.Path, "/tags/")
		if len(parts) > 1 {
			resourceARN = parts[1]
		}
	}

	if resourceARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "resourceArn is required",
		})

		return
	}

	var req TagResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if err := s.storage.TagResource(resourceARN, req.Tags); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &TagResourceResponse{})
}

// UntagResource handles the UntagResource API.
func (s *Service) UntagResource(w http.ResponseWriter, r *http.Request) {
	resourceARN := r.URL.Query().Get("resourceArn")
	if resourceARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "resourceArn is required",
		})

		return
	}

	tagKeys := r.URL.Query()["tagKeys"]

	if err := s.storage.UntagResource(resourceARN, tagKeys); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &UntagResourceResponse{})
}

// ListTagsForResource handles the ListTagsForResource API.
func (s *Service) ListTagsForResource(w http.ResponseWriter, r *http.Request) {
	resourceARN := r.URL.Query().Get("resourceArn")
	if resourceARN == "" {
		// Try to extract from path
		parts := strings.Split(r.URL.Path, "/tags/")
		if len(parts) > 1 {
			resourceARN = parts[1]
		}
	}

	if resourceARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "resourceArn is required",
		})

		return
	}

	tags, err := s.storage.ListTagsForResource(resourceARN)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &ListTagsForResourceResponse{Tags: tags})
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, statusCode int, e *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(e)
}

// handleError handles storage errors.
func handleError(w http.ResponseWriter, err error) {
	var e *Error
	if errors.As(err, &e) {
		var statusCode int

		switch e.Code {
		case errResourceNotFound:
			statusCode = http.StatusNotFound
		case errConflict:
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusBadRequest
		}

		writeError(w, statusCode, e)

		return
	}

	writeError(w, http.StatusInternalServerError, &Error{
		Code:    "InternalServerException",
		Message: err.Error(),
	})
}
