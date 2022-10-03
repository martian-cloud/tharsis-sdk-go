package types

import "time"

// Supporting structs for the Run paginator:

// RunSortableField represents the fields that a workspace can be sorted by
type RunSortableField string

// VariableCategory specifies if the variable is a terraform
// or environment variable
type VariableCategory string

// RunStatus represents the various states for a Run resource
type RunStatus string

// RunSortableField, VariableCategory, RunStatus constants
const (
	RunSortableFieldCreatedAtAsc  RunSortableField = "CREATED_AT_ASC"
	RunSortableFieldCreatedAtDesc RunSortableField = "CREATED_AT_DESC"
	RunSortableFieldUpdatedAtAsc  RunSortableField = "UPDATED_AT_ASC"
	RunSortableFieldUpdatedAtDesc RunSortableField = "UPDATED_AT_DESC"

	TerraformVariableCategory   VariableCategory = "terraform"
	EnvironmentVariableCategory VariableCategory = "environment"

	// Run Status Types
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"
)

// RunFilter contains the supported fields for filtering Run resources
type RunFilter struct {
	WorkspacePath *string
	WorkspaceID   *string
}

// GetRunsInput is the input for listing runs
type GetRunsInput struct {
	// Sort specifies the field to sort on and direction
	Sort *RunSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// Filter is used to filter the results
	Filter *RunFilter
}

// GetRunsOutput is the output when listing runs
type GetRunsOutput struct {
	PageInfo *PageInfo
	Runs     []Run
}

// GetPageInfo allows GetRunsOutput to implement the PaginatedResponse interface.
func (gro *GetRunsOutput) GetPageInfo() *PageInfo {
	return gro.PageInfo
}

//////////////////////////////////////////////////////////////////////////////

// Run holds information about a Tharsis run.
// It is used as input to and output from some operations.
// ID resides in the metadata
type Run struct {
	ModuleSource           *string
	ConfigurationVersionID *string
	ForceCanceledBy        *string
	ModuleVersion          *string
	Plan                   *Plan
	Apply                  *Apply
	ForceCancelAvailableAt *time.Time
	Metadata               ResourceMetadata
	WorkspaceID            string
	WorkspacePath          string
	Status                 RunStatus
	CreatedBy              string
	TerraformVersion       string
	IsDestroy              bool
	ForceCanceled          bool
}

// RunVariable holds information about a run variable
type RunVariable struct {
	Value         *string          `json:"value"`
	NamespacePath *string          `json:"namespacePath"`
	Key           string           `json:"key"`
	Category      VariableCategory `json:"category"`
	Hcl           bool             `json:"hcl"`
}

// GetRunInput is the input to specify a single run to fetch.
type GetRunInput struct {
	ID string
}

// RunSubscriptionInput is the input for subscribing to run events.
type RunSubscriptionInput struct {
	RunID         *string `json:"runId"`
	WorkspacePath string  `json:"workspacePath"`
}

// CreateRunInput is the input for creating a new run.
type CreateRunInput struct {
	ConfigurationVersionID *string       `json:"configurationVersionId"`
	ModuleSource           *string       `json:"moduleSource"`
	ModuleVersion          *string       `json:"moduleVersion"`
	TerraformVersion       *string       `json:"terraformVersion"`
	WorkspacePath          string        `json:"workspacePath"`
	Variables              []RunVariable `json:"variables"`
	IsDestroy              bool          `json:"isDestroy"`
}

// ApplyRunInput is the input for applying a run.
type ApplyRunInput struct {
	Comment *string `json:"comment,omitempty"`
	RunID   string  `json:"runId"`
}

// CancelRunInput is the input for canceling a run.
type CancelRunInput struct {
	Comment *string `json:"comment,omitempty"`
	Force   *bool   `json:"force"`
	RunID   string  `json:"runId"`
}

// The End.
