package scheduler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// CreateSchedule handles the CreateSchedule API.
func (s *Service) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule name is required", http.StatusBadRequest)

		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errValidationException, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.ScheduleExpression == "" {
		writeError(w, errValidationException, "ScheduleExpression is required", http.StatusBadRequest)

		return
	}

	if req.FlexibleTimeWindow == nil {
		writeError(w, errValidationException, "FlexibleTimeWindow is required", http.StatusBadRequest)

		return
	}

	if req.Target == nil {
		writeError(w, errValidationException, "Target is required", http.StatusBadRequest)

		return
	}

	schedule, err := s.storage.CreateSchedule(r.Context(), name, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &CreateScheduleResponse{
		ScheduleArn: schedule.ARN,
	})
}

// GetSchedule handles the GetSchedule API.
func (s *Service) GetSchedule(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule name is required", http.StatusBadRequest)

		return
	}

	groupName := r.URL.Query().Get("groupName")

	schedule, err := s.storage.GetSchedule(r.Context(), name, groupName)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := &GetScheduleResponse{
		ActionAfterCompletion:      schedule.ActionAfterCompletion,
		Arn:                        schedule.ARN,
		CreationDate:               float64(schedule.CreationDate.Unix()),
		Description:                schedule.Description,
		FlexibleTimeWindow:         schedule.FlexibleTimeWindow,
		GroupName:                  schedule.GroupName,
		KmsKeyArn:                  schedule.KmsKeyArn,
		LastModificationDate:       float64(schedule.LastModificationDate.Unix()),
		Name:                       schedule.Name,
		ScheduleExpression:         schedule.ScheduleExpression,
		ScheduleExpressionTimezone: schedule.ScheduleExpressionTimezone,
		State:                      schedule.State,
		Target:                     schedule.Target,
	}

	if schedule.StartDate != nil {
		startDate := schedule.StartDate.Format("2006-01-02T15:04:05Z")
		resp.StartDate = &startDate
	}

	if schedule.EndDate != nil {
		endDate := schedule.EndDate.Format("2006-01-02T15:04:05Z")
		resp.EndDate = &endDate
	}

	writeResponse(w, resp)
}

// UpdateSchedule handles the UpdateSchedule API.
func (s *Service) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule name is required", http.StatusBadRequest)

		return
	}

	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errValidationException, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.ScheduleExpression == "" {
		writeError(w, errValidationException, "ScheduleExpression is required", http.StatusBadRequest)

		return
	}

	if req.FlexibleTimeWindow == nil {
		writeError(w, errValidationException, "FlexibleTimeWindow is required", http.StatusBadRequest)

		return
	}

	if req.Target == nil {
		writeError(w, errValidationException, "Target is required", http.StatusBadRequest)

		return
	}

	schedule, err := s.storage.UpdateSchedule(r.Context(), name, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &UpdateScheduleResponse{
		ScheduleArn: schedule.ARN,
	})
}

// DeleteSchedule handles the DeleteSchedule API.
func (s *Service) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule name is required", http.StatusBadRequest)

		return
	}

	groupName := r.URL.Query().Get("groupName")

	if err := s.storage.DeleteSchedule(r.Context(), name, groupName); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, struct{}{})
}

// ListSchedules handles the ListSchedules API.
func (s *Service) ListSchedules(w http.ResponseWriter, r *http.Request) {
	groupName := r.URL.Query().Get("GroupName")

	schedules, err := s.storage.ListSchedules(r.Context(), groupName, 100)
	if err != nil {
		handleError(w, err)

		return
	}

	summaries := make([]ScheduleSummary, 0, len(schedules))

	for _, schedule := range schedules {
		summary := ScheduleSummary{
			Arn:                  schedule.ARN,
			CreationDate:         float64(schedule.CreationDate.Unix()),
			GroupName:            schedule.GroupName,
			LastModificationDate: float64(schedule.LastModificationDate.Unix()),
			Name:                 schedule.Name,
			State:                schedule.State,
		}

		if schedule.Target != nil {
			summary.Target = &TargetSummary{
				Arn: schedule.Target.Arn,
			}
		}

		summaries = append(summaries, summary)
	}

	writeResponse(w, &ListSchedulesResponse{
		Schedules: summaries,
	})
}

// CreateScheduleGroup handles the CreateScheduleGroup API.
func (s *Service) CreateScheduleGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule group name is required", http.StatusBadRequest)

		return
	}

	var req CreateScheduleGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is OK for CreateScheduleGroup.
		req = CreateScheduleGroupRequest{}
	}

	group, err := s.storage.CreateScheduleGroup(r.Context(), name, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &CreateScheduleGroupResponse{
		ScheduleGroupArn: group.ARN,
	})
}

// GetScheduleGroup handles the GetScheduleGroup API.
func (s *Service) GetScheduleGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule group name is required", http.StatusBadRequest)

		return
	}

	group, err := s.storage.GetScheduleGroup(r.Context(), name)
	if err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, &GetScheduleGroupResponse{
		Arn:                  group.ARN,
		CreationDate:         float64(group.CreationDate.Unix()),
		LastModificationDate: float64(group.CreationDate.Unix()),
		Name:                 group.Name,
		State:                group.State,
	})
}

// DeleteScheduleGroup handles the DeleteScheduleGroup API.
func (s *Service) DeleteScheduleGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, errValidationException, "Schedule group name is required", http.StatusBadRequest)

		return
	}

	if err := s.storage.DeleteScheduleGroup(r.Context(), name); err != nil {
		handleError(w, err)

		return
	}

	writeResponse(w, struct{}{})
}

// ListScheduleGroups handles the ListScheduleGroups API.
func (s *Service) ListScheduleGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.storage.ListScheduleGroups(r.Context(), 100)
	if err != nil {
		handleError(w, err)

		return
	}

	summaries := make([]ScheduleGroupSummary, 0, len(groups))

	for _, group := range groups {
		summaries = append(summaries, ScheduleGroupSummary{
			Arn:                  group.ARN,
			CreationDate:         float64(group.CreationDate.Unix()),
			LastModificationDate: float64(group.CreationDate.Unix()),
			Name:                 group.Name,
			State:                group.State,
		})
	}

	writeResponse(w, &ListScheduleGroupsResponse{
		ScheduleGroups: summaries,
	})
}

// writeResponse writes a JSON response.
func writeResponse(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.Header().Set("x-amzn-ErrorType", code)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{
		Message: message,
	})
}

// handleError handles service errors.
func handleError(w http.ResponseWriter, err error) {
	var sErr *Error
	if errors.As(err, &sErr) {
		status := getErrorStatus(sErr.Code)
		writeError(w, sErr.Code, sErr.Message, status)

		return
	}

	writeError(w, "InternalServerException", err.Error(), http.StatusInternalServerError)
}

// getErrorStatus returns the HTTP status code for a given error code.
func getErrorStatus(code string) int {
	switch code {
	case errResourceNotFound:
		return http.StatusNotFound
	case errConflictException:
		return http.StatusConflict
	case errValidationException:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
