package types

// Supporting structs for the GetWorkspaces paginator:

// WorkspaceSortableField represents the fields that a workspace can be sorted by
type WorkspaceSortableField string

// WorkspaceSortableField constants
const (
	WorkspaceSortableFieldFullPathAsc   WorkspaceSortableField = "FULL_PATH_ASC"
	WorkspaceSortableFieldFullPathDesc  WorkspaceSortableField = "FULL_PATH_DESC"
	WorkspaceSortableFieldUpdatedAtAsc  WorkspaceSortableField = "UPDATED_AT_ASC"
	WorkspaceSortableFieldUpdatedAtDesc WorkspaceSortableField = "UPDATED_AT_DESC"
)

// WorkspaceFilter contains the supported field(s) for filtering Workspace resources
type WorkspaceFilter struct {
	GroupPath *string
}

// GetWorkspacesInput is the input for listing workspaces
type GetWorkspacesInput struct {
	// Sort specifies the field to sort on and direction
	Sort *WorkspaceSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// Filter is used to filter the results
	Filter *WorkspaceFilter
}

// GetWorkspacesOutput is the output when listing groups
type GetWorkspacesOutput struct {
	PageInfo   *PageInfo
	Workspaces []Workspace
}

// GetPageInfo allows GetWorkspacesOutput to implement the PaginatedResponse interface.
func (ggo *GetWorkspacesOutput) GetPageInfo() *PageInfo {
	return ggo.PageInfo
}

//////////////////////////////////////////////////////////////////////////////

// Workspace holds information about a Tharsis workspace.
// It is used as input to and output from some operations.
//
// Tharsis API has CurrentRunID and CurrentStateVersionID, but
// GraphQL has {assigned,available}ManagedIdentities
//
type Workspace struct {
	CurrentStateVersion       *StateVersion
	Metadata                  ResourceMetadata
	Name                      string
	FullPath                  string
	Description               string
	TerraformVersion          string
	AssignedManagedIdentities []ManagedIdentity
	MaxJobDuration            int32
}

// GetWorkspaceInput is the input to specify a single workspace to fetch.
type GetWorkspaceInput struct {
	Path string
}

// CreateWorkspaceInput is the input for creating a new workspace.
type CreateWorkspaceInput struct {
	MaxJobDuration   *int32  `json:"maxJobDuration"`
	TerraformVersion *string `json:"terraformVersion"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	GroupPath        string  `json:"groupPath"`
}

// UpdateWorkspaceInput is the input for updating a workspace.
type UpdateWorkspaceInput struct {
	MaxJobDuration   *int32  `json:"maxJobDuration"`
	TerraformVersion *string `json:"terraformVersion"`
	WorkspacePath    string  `json:"workspacePath"`
	Description      string  `json:"description"`
}

// DeleteWorkspaceInput is the input for deleting a workspace.
type DeleteWorkspaceInput struct {
	Force         *bool  `json:"force"`
	WorkspacePath string `json:"workspacePath"`
}

// The End.
