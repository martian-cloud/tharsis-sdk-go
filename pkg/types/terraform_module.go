package types

// Supporting structs for the TerraformModule paginator:

// TerraformModuleSortableField represents the fields that a TerraformModule can be sorted by
type TerraformModuleSortableField string

// TerraformModuleSortableField constants
const (
	TerraformModuleSortableFieldNameAsc       TerraformModuleSortableField = "NAME_ASC"
	TerraformModuleSortableFieldNameDesc      TerraformModuleSortableField = "NAME_DESC"
	TerraformModuleSortableFieldUpdatedAtAsc  TerraformModuleSortableField = "UPDATED_AT_ASC"
	TerraformModuleSortableFieldUpdatedAtDesc TerraformModuleSortableField = "UPDATED_AT_DESC"
)

// TerraformModuleFilter contains the supported fields for filtering TerraformModule resources
type TerraformModuleFilter struct {
	Search *string
}

// GetTerraformModulesInput is the input for listing TerraformModules
type GetTerraformModulesInput struct {
	// Sort specifies the field to sort on and direction
	Sort *TerraformModuleSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// Filter is used to filter the results
	Filter *TerraformModuleFilter
}

// GetTerraformModulesOutput is the output when listing TerraformModules
type GetTerraformModulesOutput struct {
	PageInfo         *PageInfo
	TerraformModules []TerraformModule
}

// GetPageInfo allows GetTerraformModulesOutput to implement the PaginatedResponse interface.
func (gtm *GetTerraformModulesOutput) GetPageInfo() *PageInfo {
	return gtm.PageInfo
}

// TerraformModule represents a Terraform module.
type TerraformModule struct {
	Metadata          ResourceMetadata
	Name              string
	System            string
	GroupPath         string
	ResourcePath      string
	RegistryNamespace string
	RepositoryURL     string
	Private           bool
}

// GetTerraformModuleInput is the input to specify a single module to fetch.
type GetTerraformModuleInput struct {
	ID   *string
	Path *string
	TRN  *string
}

// DeleteTerraformModuleInput is the input for deleting a terraform module
type DeleteTerraformModuleInput struct {
	ID string `json:"id"`
}

// CreateTerraformModuleInput is the input for creating a new module.
type CreateTerraformModuleInput struct {
	Name          string `json:"name"`
	System        string `json:"system"`
	GroupPath     string `json:"groupPath"`
	RepositoryURL string `json:"repositoryUrl"`
	Private       bool   `json:"private"`
}

// UpdateTerraformModuleInput is the input for updating a module.
type UpdateTerraformModuleInput struct {
	RepositoryURL *string `json:"repositoryUrl"`
	Private       *bool   `json:"private"`
	ID            string  `json:"id"`
}
