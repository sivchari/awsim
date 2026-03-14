package codeguruprofiler

import "time"

// ProfilingGroup represents a CodeGuru Profiler profiling group.
type ProfilingGroup struct {
	AgentOrchestrationConfig *AgentOrchestrationConfig `json:"agentOrchestrationConfig,omitempty"`
	Arn                      string                    `json:"arn"`
	ComputePlatform          string                    `json:"computePlatform"`
	CreatedAt                time.Time                 `json:"createdAt"`
	Name                     string                    `json:"name"`
	ProfilingStatus          *ProfilingStatus          `json:"profilingStatus,omitempty"`
	Tags                     map[string]string         `json:"tags,omitempty"`
	UpdatedAt                time.Time                 `json:"updatedAt"`
}

// AgentOrchestrationConfig represents agent orchestration configuration.
type AgentOrchestrationConfig struct {
	ProfilingEnabled bool `json:"profilingEnabled"`
}

// ProfilingStatus represents the profiling status.
type ProfilingStatus struct {
	LatestAgentOrchestratedAt    *time.Time             `json:"latestAgentOrchestratedAt,omitempty"`
	LatestAgentProfileReportedAt *time.Time             `json:"latestAgentProfileReportedAt,omitempty"`
	LatestAggregatedProfile      *AggregatedProfileTime `json:"latestAggregatedProfile,omitempty"`
}

// AggregatedProfileTime represents an aggregated profile time.
type AggregatedProfileTime struct {
	Period string    `json:"period,omitempty"`
	Start  time.Time `json:"start,omitzero"`
}

// CreateProfilingGroupInput represents the request body for CreateProfilingGroup.
type CreateProfilingGroupInput struct {
	AgentOrchestrationConfig *AgentOrchestrationConfig `json:"agentOrchestrationConfig,omitempty"`
	ComputePlatform          string                    `json:"computePlatform,omitempty"`
	ProfilingGroupName       string                    `json:"profilingGroupName"`
	Tags                     map[string]string         `json:"tags,omitempty"`
}

// UpdateProfilingGroupInput represents the request body for UpdateProfilingGroup.
type UpdateProfilingGroupInput struct {
	AgentOrchestrationConfig *AgentOrchestrationConfig `json:"agentOrchestrationConfig"`
}

// ListProfilingGroupsResponse represents the response for ListProfilingGroups.
type ListProfilingGroupsResponse struct {
	ProfilingGroupNames []string         `json:"profilingGroupNames"`
	ProfilingGroups     []ProfilingGroup `json:"profilingGroups"`
	NextToken           string           `json:"nextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message"`
}
