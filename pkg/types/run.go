package types

import "time"

// Supporting structs for the Run paginator:

// RunSortableField represents the fields that a workspace can be sorted by
type RunSortableField string

// VariableCategory type declaration moved to the variable module.

// RunStatus represents the various states for a Run resource
type RunStatus string

// RunSortableField, VariableCategory, RunStatus constants
const (
	RunSortableFieldCreatedAtAsc  RunSortableField = "CREATED_AT_ASC"
	RunSortableFieldCreatedAtDesc RunSortableField = "CREATED_AT_DESC"
	RunSortableFieldUpdatedAtAsc  RunSortableField = "UPDATED_AT_ASC"
	RunSortableFieldUpdatedAtDesc RunSortableField = "UPDATED_AT_DESC"

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
	ModuleDigest           *string
	StateVersionID         *string
	Plan                   *Plan
	Apply                  *Apply
	ForceCancelAvailableAt *time.Time
	Metadata               ResourceMetadata
	WorkspaceID            string
	WorkspacePath          string
	Status                 RunStatus
	CreatedBy              string
	TerraformVersion       string
	TargetAddresses        []string
	IsDestroy              bool
	ForceCanceled          bool
	Refresh                bool
	RefreshOnly            bool
	Speculative            bool
}

// RunVariable holds information about a run variable
type RunVariable struct {
	Value              *string          `json:"value"`
	NamespacePath      *string          `json:"namespacePath"`
	Key                string           `json:"key"`
	Category           VariableCategory `json:"category"`
	HCL                bool             `json:"hcl"`
	Sensitive          bool             `json:"sensitive"`
	VersionID          *string          `json:"versionId"`
	IncludedInTFConfig *bool            `json:"includedInTfConfig"`
}

// SetVariablesIncludedInTFConfigInput is the input for setting
// variables that are included in the Terraform config.
type SetVariablesIncludedInTFConfigInput struct {
	RunID        string   `json:"runId"`
	VariableKeys []string `json:"variableKeys"`
}

// GetRunInput is the input to specify a single run to fetch.
type GetRunInput struct {
	ID  string  `json:"id"`
}

// GetRunVariablesInput is the input for getting run variables
type GetRunVariablesInput struct {
	RunID                  string
	IncludeSensitiveValues bool
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
	Speculative            *bool         `json:"speculative"`
	WorkspacePath          string        `json:"workspacePath"`
	Variables              []RunVariable `json:"variables"`
	TargetAddresses        []string      `json:"targetAddresses"`
	IsDestroy              bool          `json:"isDestroy"`
	Refresh                bool          `json:"refresh"`
	RefreshOnly            bool          `json:"refreshOnly"`
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
