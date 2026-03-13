package backup

// Vault represents an AWS Backup vault.
type Vault struct {
	BackupVaultArn         string  `json:"BackupVaultArn"`
	BackupVaultName        string  `json:"BackupVaultName"`
	CreationDate           float64 `json:"CreationDate"`
	CreatorRequestID       string  `json:"CreatorRequestId,omitempty"`
	EncryptionKeyArn       string  `json:"EncryptionKeyArn,omitempty"`
	NumberOfRecoveryPoints int64   `json:"NumberOfRecoveryPoints"`
}

// Rule represents a backup rule in a plan.
type Rule struct {
	RuleName                string `json:"RuleName"`
	RuleID                  string `json:"RuleId"`
	TargetBackupVaultName   string `json:"TargetBackupVaultName"`
	ScheduleExpression      string `json:"ScheduleExpression,omitempty"`
	StartWindowMinutes      *int64 `json:"StartWindowMinutes,omitempty"`
	CompletionWindowMinutes *int64 `json:"CompletionWindowMinutes,omitempty"`
}

// PlanData represents the body of a backup plan.
type PlanData struct {
	BackupPlanName string `json:"BackupPlanName"`
	Rules          []Rule `json:"Rules"`
}

// Plan represents an AWS Backup plan with metadata.
type Plan struct {
	BackupPlanArn string    `json:"BackupPlanArn"`
	BackupPlanID  string    `json:"BackupPlanId"`
	BackupPlan    *PlanData `json:"BackupPlan"`
	CreationDate  float64   `json:"CreationDate"`
	VersionID     string    `json:"VersionId"`
}

// SelectionData represents the body of a backup selection.
type SelectionData struct {
	SelectionName string   `json:"SelectionName"`
	IamRoleArn    string   `json:"IamRoleArn"`
	Resources     []string `json:"Resources,omitempty"`
}

// Selection represents an AWS Backup selection with metadata.
type Selection struct {
	BackupPlanID    string         `json:"BackupPlanId"`
	SelectionID     string         `json:"SelectionId"`
	BackupSelection *SelectionData `json:"BackupSelection"`
	CreationDate    float64        `json:"CreationDate"`
}

// CreateBackupVaultInput represents a CreateBackupVault request body.
type CreateBackupVaultInput struct {
	BackupVaultTags  map[string]string `json:"BackupVaultTags,omitempty"`
	CreatorRequestID string            `json:"CreatorRequestId,omitempty"`
	EncryptionKeyArn string            `json:"EncryptionKeyArn,omitempty"`
}

// CreateBackupVaultResponse represents a CreateBackupVault response.
type CreateBackupVaultResponse struct {
	BackupVaultArn  string  `json:"BackupVaultArn"`
	BackupVaultName string  `json:"BackupVaultName"`
	CreationDate    float64 `json:"CreationDate"`
}

// CreateBackupPlanInput represents a CreateBackupPlan request body.
type CreateBackupPlanInput struct {
	BackupPlan       *PlanInputData    `json:"BackupPlan"`
	BackupPlanTags   map[string]string `json:"BackupPlanTags,omitempty"`
	CreatorRequestID string            `json:"CreatorRequestId,omitempty"`
}

// PlanInputData represents the input body for a backup plan.
type PlanInputData struct {
	BackupPlanName string      `json:"BackupPlanName"`
	Rules          []RuleInput `json:"Rules"`
}

// RuleInput represents a backup rule in a create plan request.
type RuleInput struct {
	RuleName                string `json:"RuleName"`
	TargetBackupVaultName   string `json:"TargetBackupVaultName"`
	ScheduleExpression      string `json:"ScheduleExpression,omitempty"`
	StartWindowMinutes      *int64 `json:"StartWindowMinutes,omitempty"`
	CompletionWindowMinutes *int64 `json:"CompletionWindowMinutes,omitempty"`
}

// CreateBackupPlanResponse represents a CreateBackupPlan response.
type CreateBackupPlanResponse struct {
	BackupPlanArn string  `json:"BackupPlanArn"`
	BackupPlanID  string  `json:"BackupPlanId"`
	CreationDate  float64 `json:"CreationDate"`
	VersionID     string  `json:"VersionId"`
}

// CreateBackupSelectionInput represents a CreateBackupSelection request body.
type CreateBackupSelectionInput struct {
	BackupSelection  *SelectionData `json:"BackupSelection"`
	CreatorRequestID string         `json:"CreatorRequestId,omitempty"`
}

// CreateBackupSelectionResponse represents a CreateBackupSelection response.
type CreateBackupSelectionResponse struct {
	BackupPlanID string  `json:"BackupPlanId"`
	CreationDate float64 `json:"CreationDate"`
	SelectionID  string  `json:"SelectionId"`
}

// ListBackupVaultsResponse represents a ListBackupVaults response.
type ListBackupVaultsResponse struct {
	BackupVaultList []Vault `json:"BackupVaultList"`
}

// ListBackupPlansResponse represents a ListBackupPlans response.
type ListBackupPlansResponse struct {
	BackupPlansList []PlanListMember `json:"BackupPlansList"`
}

// PlanListMember represents a backup plan in a list response.
type PlanListMember struct {
	BackupPlanArn  string  `json:"BackupPlanArn"`
	BackupPlanID   string  `json:"BackupPlanId"`
	BackupPlanName string  `json:"BackupPlanName"`
	CreationDate   float64 `json:"CreationDate"`
	VersionID      string  `json:"VersionId"`
}

// ListBackupSelectionsResponse represents a ListBackupSelections response.
type ListBackupSelectionsResponse struct {
	BackupSelectionsList []SelectionListMember `json:"BackupSelectionsList"`
}

// SelectionListMember represents a backup selection in a list response.
type SelectionListMember struct {
	BackupPlanID  string  `json:"BackupPlanId"`
	CreationDate  float64 `json:"CreationDate"`
	IamRoleArn    string  `json:"IamRoleArn"`
	SelectionID   string  `json:"SelectionId"`
	SelectionName string  `json:"SelectionName"`
}

// GetBackupPlanResponse represents a GetBackupPlan response.
type GetBackupPlanResponse struct {
	BackupPlan    *PlanData `json:"BackupPlan"`
	BackupPlanArn string    `json:"BackupPlanArn"`
	BackupPlanID  string    `json:"BackupPlanId"`
	CreationDate  float64   `json:"CreationDate"`
	VersionID     string    `json:"VersionId"`
}

// GetBackupSelectionResponse represents a GetBackupSelection response.
type GetBackupSelectionResponse struct {
	BackupPlanID    string         `json:"BackupPlanId"`
	BackupSelection *SelectionData `json:"BackupSelection"`
	CreationDate    float64        `json:"CreationDate"`
	SelectionID     string         `json:"SelectionId"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}
