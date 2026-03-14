// Package codeguruprofiler implements the AWS CodeGuru Profiler service.
package codeguruprofiler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{
		Message: message,
	})
}

// CreateProfilingGroup handles POST /profilingGroups.
func (s *Service) CreateProfilingGroup(w http.ResponseWriter, r *http.Request) {
	var input CreateProfilingGroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	group := s.storage.CreateProfilingGroup(&input)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(group)
}

// DescribeProfilingGroup handles GET /profilingGroups/{profilingGroupName}.
func (s *Service) DescribeProfilingGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("profilingGroupName")

	group, err := s.storage.DescribeProfilingGroup(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())

		return
	}

	writeJSON(w, group)
}

// UpdateProfilingGroup handles PUT /profilingGroups/{profilingGroupName}.
func (s *Service) UpdateProfilingGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("profilingGroupName")

	var input UpdateProfilingGroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	group, err := s.storage.UpdateProfilingGroup(name, &input)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())

		return
	}

	writeJSON(w, group)
}

// DeleteProfilingGroup handles DELETE /profilingGroups/{profilingGroupName}.
func (s *Service) DeleteProfilingGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("profilingGroupName")

	if err := s.storage.DeleteProfilingGroup(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProfilingGroups handles GET /profilingGroups.
func (s *Service) ListProfilingGroups(w http.ResponseWriter, _ *http.Request) {
	groups := s.storage.ListProfilingGroups()

	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}

	writeJSON(w, &ListProfilingGroupsResponse{
		ProfilingGroupNames: names,
		ProfilingGroups:     groups,
	})
}
