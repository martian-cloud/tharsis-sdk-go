package types

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
	ID string
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
