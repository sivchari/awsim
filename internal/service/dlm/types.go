// Package dlm provides Data Lifecycle Manager service emulation for awsim.
package dlm

import "time"

// LifecyclePolicy represents a DLM lifecycle policy.
type LifecyclePolicy struct {
	PolicyID         string            `json:"PolicyId"`
	Description      string            `json:"Description"`
	State            string            `json:"State"`
	ExecutionRoleArn string            `json:"ExecutionRoleArn"`
	DateCreated      time.Time         `json:"DateCreated"`
	DateModified     time.Time         `json:"DateModified"`
	PolicyDetails    *PolicyDetails    `json:"PolicyDetails,omitempty"`
	Tags             map[string]string `json:"Tags,omitempty"`
	PolicyArn        string            `json:"PolicyArn"`
	DefaultPolicy    bool              `json:"DefaultPolicy,omitempty"`
}

// PolicyDetails contains configuration details for a lifecycle policy.
type PolicyDetails struct {
	PolicyType        string       `json:"PolicyType,omitempty"`
	ResourceTypes     []string     `json:"ResourceTypes,omitempty"`
	ResourceLocations []string     `json:"ResourceLocations,omitempty"`
	TargetTags        []Tag        `json:"TargetTags,omitempty"`
	Schedules         []Schedule   `json:"Schedules,omitempty"`
	Parameters        *Parameters  `json:"Parameters,omitempty"`
	Actions           []Action     `json:"Actions,omitempty"`
	EventSource       *EventSource `json:"EventSource,omitempty"`
	CopyTags          bool         `json:"CopyTags,omitempty"`
}

// Tag represents a key-value tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// Schedule represents a schedule for a lifecycle policy.
type Schedule struct {
	Name                 string                `json:"Name,omitempty"`
	CopyTags             bool                  `json:"CopyTags,omitempty"`
	TagsToAdd            []Tag                 `json:"TagsToAdd,omitempty"`
	VariableTags         []Tag                 `json:"VariableTags,omitempty"`
	CreateRule           *CreateRule           `json:"CreateRule,omitempty"`
	RetainRule           *RetainRule           `json:"RetainRule,omitempty"`
	FastRestoreRule      *FastRestoreRule      `json:"FastRestoreRule,omitempty"`
	CrossRegionCopyRules []CrossRegionCopyRule `json:"CrossRegionCopyRules,omitempty"`
	ShareRules           []ShareRule           `json:"ShareRules,omitempty"`
	DeprecateRule        *DeprecateRule        `json:"DeprecateRule,omitempty"`
	ArchiveRule          *ArchiveRule          `json:"ArchiveRule,omitempty"`
}

// CreateRule specifies when snapshots are created.
type CreateRule struct {
	Location       string   `json:"Location,omitempty"`
	Interval       int32    `json:"Interval,omitempty"`
	IntervalUnit   string   `json:"IntervalUnit,omitempty"`
	Times          []string `json:"Times,omitempty"`
	CronExpression string   `json:"CronExpression,omitempty"`
}

// RetainRule specifies retention for snapshots.
type RetainRule struct {
	Count        int32  `json:"Count,omitempty"`
	Interval     int32  `json:"Interval,omitempty"`
	IntervalUnit string `json:"IntervalUnit,omitempty"`
}

// FastRestoreRule specifies fast snapshot restore settings.
type FastRestoreRule struct {
	Count             int32    `json:"Count,omitempty"`
	Interval          int32    `json:"Interval,omitempty"`
	IntervalUnit      string   `json:"IntervalUnit,omitempty"`
	AvailabilityZones []string `json:"AvailabilityZones,omitempty"`
}

// CrossRegionCopyRule specifies cross-region copy settings.
type CrossRegionCopyRule struct {
	TargetRegion  string                        `json:"TargetRegion,omitempty"`
	Target        string                        `json:"Target,omitempty"`
	Encrypted     bool                          `json:"Encrypted,omitempty"`
	CmkArn        string                        `json:"CmkArn,omitempty"`
	CopyTags      bool                          `json:"CopyTags,omitempty"`
	RetainRule    *CrossRegionCopyRetainRule    `json:"RetainRule,omitempty"`
	DeprecateRule *CrossRegionCopyDeprecateRule `json:"DeprecateRule,omitempty"`
}

// CrossRegionCopyRetainRule specifies retention for cross-region copies.
type CrossRegionCopyRetainRule struct {
	Interval     int32  `json:"Interval,omitempty"`
	IntervalUnit string `json:"IntervalUnit,omitempty"`
}

// CrossRegionCopyDeprecateRule specifies deprecation for cross-region copies.
type CrossRegionCopyDeprecateRule struct {
	Interval     int32  `json:"Interval,omitempty"`
	IntervalUnit string `json:"IntervalUnit,omitempty"`
}

// ShareRule specifies snapshot sharing settings.
type ShareRule struct {
	TargetAccounts      []string `json:"TargetAccounts,omitempty"`
	UnshareInterval     int32    `json:"UnshareInterval,omitempty"`
	UnshareIntervalUnit string   `json:"UnshareIntervalUnit,omitempty"`
}

// DeprecateRule specifies AMI deprecation settings.
type DeprecateRule struct {
	Count        int32  `json:"Count,omitempty"`
	Interval     int32  `json:"Interval,omitempty"`
	IntervalUnit string `json:"IntervalUnit,omitempty"`
}

// ArchiveRule specifies archive settings.
type ArchiveRule struct {
	RetainRule *ArchiveRetainRule `json:"RetainRule,omitempty"`
}

// ArchiveRetainRule specifies retention for archived snapshots.
type ArchiveRetainRule struct {
	RetentionArchiveTier *RetentionArchiveTier `json:"RetentionArchiveTier,omitempty"`
}

// RetentionArchiveTier specifies archive tier retention.
type RetentionArchiveTier struct {
	Count        int32  `json:"Count,omitempty"`
	Interval     int32  `json:"Interval,omitempty"`
	IntervalUnit string `json:"IntervalUnit,omitempty"`
}

// Parameters contains additional parameters for a policy.
type Parameters struct {
	ExcludeBootVolume     bool  `json:"ExcludeBootVolume,omitempty"`
	NoReboot              bool  `json:"NoReboot,omitempty"`
	ExcludeDataVolumeTags []Tag `json:"ExcludeDataVolumeTags,omitempty"`
}

// Action represents an action for event-based policies.
type Action struct {
	Name            string                  `json:"Name"`
	CrossRegionCopy []CrossRegionCopyAction `json:"CrossRegionCopy,omitempty"`
}

// CrossRegionCopyAction specifies cross-region copy action.
type CrossRegionCopyAction struct {
	Target                  string                     `json:"Target"`
	EncryptionConfiguration *EncryptionConfiguration   `json:"EncryptionConfiguration,omitempty"`
	RetainRule              *CrossRegionCopyRetainRule `json:"RetainRule,omitempty"`
}

// EncryptionConfiguration specifies encryption settings.
type EncryptionConfiguration struct {
	Encrypted bool   `json:"Encrypted"`
	CmkArn    string `json:"CmkArn,omitempty"`
}

// EventSource specifies the event source for event-based policies.
type EventSource struct {
	Type       string           `json:"Type"`
	Parameters *EventParameters `json:"Parameters,omitempty"`
}

// EventParameters contains event source parameters.
type EventParameters struct {
	EventType        string   `json:"EventType"`
	SnapshotOwner    []string `json:"SnapshotOwner,omitempty"`
	DescriptionRegex string   `json:"DescriptionRegex,omitempty"`
}

// Policy states.
const (
	StateEnabled  = "ENABLED"
	StateDisabled = "DISABLED"
	StateError    = "ERROR"
)

// Policy types.
const (
	PolicyTypeEBSSnapshotManagement = "EBS_SNAPSHOT_MANAGEMENT"
	PolicyTypeImageManagement       = "IMAGE_MANAGEMENT"
	PolicyTypeEventBasedPolicy      = "EVENT_BASED_POLICY"
)

// CreateLifecyclePolicyRequest represents the CreateLifecyclePolicy API request.
type CreateLifecyclePolicyRequest struct {
	Description            string                  `json:"Description"`
	ExecutionRoleArn       string                  `json:"ExecutionRoleArn"`
	State                  string                  `json:"State"`
	PolicyDetails          *PolicyDetails          `json:"PolicyDetails,omitempty"`
	Tags                   map[string]string       `json:"Tags,omitempty"`
	DefaultPolicy          string                  `json:"DefaultPolicy,omitempty"`
	CreateInterval         int32                   `json:"CreateInterval,omitempty"`
	RetainInterval         int32                   `json:"RetainInterval,omitempty"`
	CopyTags               bool                    `json:"CopyTags,omitempty"`
	ExtendDeletion         bool                    `json:"ExtendDeletion,omitempty"`
	CrossRegionCopyTargets []CrossRegionCopyTarget `json:"CrossRegionCopyTargets,omitempty"`
	Exclusions             *Exclusions             `json:"Exclusions,omitempty"`
}

// CrossRegionCopyTarget specifies a cross-region copy target.
type CrossRegionCopyTarget struct {
	TargetRegion string `json:"TargetRegion,omitempty"`
}

// Exclusions specifies resources to exclude from the policy.
type Exclusions struct {
	ExcludeBootVolumes bool     `json:"ExcludeBootVolumes,omitempty"`
	ExcludeVolumeTypes []string `json:"ExcludeVolumeTypes,omitempty"`
	ExcludeTags        []Tag    `json:"ExcludeTags,omitempty"`
}

// CreateLifecyclePolicyResponse represents the CreateLifecyclePolicy API response.
type CreateLifecyclePolicyResponse struct {
	PolicyID string `json:"PolicyId"`
}

// GetLifecyclePolicyResponse represents the GetLifecyclePolicy API response.
type GetLifecyclePolicyResponse struct {
	Policy *LifecyclePolicy `json:"Policy"`
}

// GetLifecyclePoliciesResponse represents the GetLifecyclePolicies API response.
type GetLifecyclePoliciesResponse struct {
	Policies []LifecyclePolicySummary `json:"Policies"`
}

// LifecyclePolicySummary represents a summary of a lifecycle policy.
type LifecyclePolicySummary struct {
	PolicyID      string            `json:"PolicyId"`
	Description   string            `json:"Description"`
	State         string            `json:"State"`
	Tags          map[string]string `json:"Tags,omitempty"`
	PolicyType    string            `json:"PolicyType,omitempty"`
	DefaultPolicy bool              `json:"DefaultPolicy,omitempty"`
}

// UpdateLifecyclePolicyRequest represents the UpdateLifecyclePolicy API request.
type UpdateLifecyclePolicyRequest struct {
	Description            string                  `json:"Description,omitempty"`
	ExecutionRoleArn       string                  `json:"ExecutionRoleArn,omitempty"`
	State                  string                  `json:"State,omitempty"`
	PolicyDetails          *PolicyDetails          `json:"PolicyDetails,omitempty"`
	CreateInterval         int32                   `json:"CreateInterval,omitempty"`
	RetainInterval         int32                   `json:"RetainInterval,omitempty"`
	CopyTags               *bool                   `json:"CopyTags,omitempty"`
	ExtendDeletion         *bool                   `json:"ExtendDeletion,omitempty"`
	CrossRegionCopyTargets []CrossRegionCopyTarget `json:"CrossRegionCopyTargets,omitempty"`
	Exclusions             *Exclusions             `json:"Exclusions,omitempty"`
}

// Error represents a service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"Message"`
	Code    string `json:"Code,omitempty"`
}
