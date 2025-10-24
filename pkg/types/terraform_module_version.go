package types

// Supporting structs for the TerraformModuleVersion paginator:

// TerraformModuleVersionSortableField represents the fields that a TerraformModuleVersion can be sorted by
type TerraformModuleVersionSortableField string

// TerraformModuleVersionSortableField constants
const (
	TerraformModuleVersionSortableFieldUpdatedAtAsc  TerraformModuleVersionSortableField = "UPDATED_AT_ASC"
	TerraformModuleVersionSortableFieldUpdatedAtDesc TerraformModuleVersionSortableField = "UPDATED_AT_DESC"
	TerraformModuleVersionSortableFieldCreatedAtAsc  TerraformModuleVersionSortableField = "CREATED_AT_ASC"
	TerraformModuleVersionSortableFieldCreatedAtDesc TerraformModuleVersionSortableField = "CREATED_AT_DESC"
)

// GetTerraformModuleVersionsInput is the input for listing TerraformModules
type GetTerraformModuleVersionsInput struct {
	// Sort specifies the field to sort on and direction
	Sort *TerraformModuleVersionSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// TerraformModuleID is the Terraform module to get versions for.
	TerraformModuleID string
}

// GetTerraformModuleVersionsOutput is the output when listing TerraformModuleVersions
type GetTerraformModuleVersionsOutput struct {
	PageInfo       *PageInfo
	ModuleVersions []TerraformModuleVersion
}

// GetPageInfo allows GetTerraformModuleVersionsOutput to implement the PaginatedResponse interface.
func (ggo *GetTerraformModuleVersionsOutput) GetPageInfo() *PageInfo {
	return ggo.PageInfo
}

// TerraformModuleVersion represents a Tharsis module version.
type TerraformModuleVersion struct {
	Metadata    ResourceMetadata
	ModuleID    string
	Version     string
	SHASum      string
	Status      string
	Error       string
	Diagnostics string
	Submodules  []string
	Examples    []string
	Latest      bool
}

// GetTerraformModuleVersionInput is the input to specify a single module version to fetch.
type GetTerraformModuleVersionInput struct {
	ID         *string
	ModulePath *string
	Version    *string
	TRN        *string
}

// DeleteTerraformModuleVersionInput is the input to delete a terraform module version
type DeleteTerraformModuleVersionInput struct {
	ID string `json:"id"`
}

// CreateTerraformModuleVersionInput is the input for creating a new module version.
type CreateTerraformModuleVersionInput struct {
	ModulePath string `json:"modulePath"`
	Version    string `json:"version"`
	SHASum     string `json:"shaSum"`
}
