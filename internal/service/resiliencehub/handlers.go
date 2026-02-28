package resiliencehub

import (
	"encoding/json"
	"errors"
	"net/http"
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
	var req DescribeAppRequest
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

	app, err := s.storage.DescribeApp(req.AppARN)
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
	var req DeleteAppRequest
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

	if err := s.storage.DeleteApp(req.AppARN); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DeleteAppResponse{AppARN: req.AppARN})
}

// ListApps handles the ListApps API.
func (s *Service) ListApps(w http.ResponseWriter, r *http.Request) {
	var req ListAppsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// For list operations, empty body is acceptable
		req = ListAppsRequest{}
	}

	apps, nextToken, err := s.storage.ListApps(&req)
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
	var req DescribeResiliencyPolicyRequest
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

	policy, err := s.storage.DescribeResiliencyPolicy(req.PolicyARN)
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
	var req DeleteResiliencyPolicyRequest
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

	if err := s.storage.DeleteResiliencyPolicy(req.PolicyARN); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DeleteResiliencyPolicyResponse{PolicyARN: req.PolicyARN})
}

// ListResiliencyPolicies handles the ListResiliencyPolicies API.
func (s *Service) ListResiliencyPolicies(w http.ResponseWriter, r *http.Request) {
	var req ListResiliencyPoliciesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// For list operations, empty body is acceptable
		req = ListResiliencyPoliciesRequest{}
	}

	policies, nextToken, err := s.storage.ListResiliencyPolicies(&req)
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
	var req DescribeAppAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.AssessmentARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "assessmentArn is required",
		})

		return
	}

	assessment, err := s.storage.DescribeAppAssessment(req.AssessmentARN)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DescribeAppAssessmentResponse{Assessment: assessment})
}

// DeleteAppAssessment handles the DeleteAppAssessment API.
func (s *Service) DeleteAppAssessment(w http.ResponseWriter, r *http.Request) {
	var req DeleteAppAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.AssessmentARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "assessmentArn is required",
		})

		return
	}

	if err := s.storage.DeleteAppAssessment(req.AssessmentARN); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &DeleteAppAssessmentResponse{
		AssessmentARN:    req.AssessmentARN,
		AssessmentStatus: "Success",
	})
}

// ListAppAssessments handles the ListAppAssessments API.
func (s *Service) ListAppAssessments(w http.ResponseWriter, r *http.Request) {
	var req ListAppAssessmentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// For list operations, empty body is acceptable
		req = ListAppAssessmentsRequest{}
	}

	assessments, nextToken, err := s.storage.ListAppAssessments(&req)
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
	var req TagResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.ResourceARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "resourceArn is required",
		})

		return
	}

	if err := s.storage.TagResource(req.ResourceARN, req.Tags); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &TagResourceResponse{})
}

// UntagResource handles the UntagResource API.
func (s *Service) UntagResource(w http.ResponseWriter, r *http.Request) {
	var req UntagResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.ResourceARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "resourceArn is required",
		})

		return
	}

	if err := s.storage.UntagResource(req.ResourceARN, req.TagKeys); err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &UntagResourceResponse{})
}

// ListTagsForResource handles the ListTagsForResource API.
func (s *Service) ListTagsForResource(w http.ResponseWriter, r *http.Request) {
	var req ListTagsForResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "Invalid request body",
		})

		return
	}

	if req.ResourceARN == "" {
		writeError(w, http.StatusBadRequest, &Error{
			Code:    "ValidationException",
			Message: "resourceArn is required",
		})

		return
	}

	tags, err := s.storage.ListTagsForResource(req.ResourceARN)
	if err != nil {
		handleError(w, err)

		return
	}

	writeJSON(w, &ListTagsForResourceResponse{Tags: tags})
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, statusCode int, e *Error) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
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
