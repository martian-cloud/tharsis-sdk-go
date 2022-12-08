package types

// TerraformModule represents a Terraform module.
type TerraformModule struct {
	Metadata          ResourceMetadata
	Name              string
	System            string
	ResourcePath      string
	RegistryNamespace string
	RepositoryURL     string
	Private           bool
}

// GetTerraformModuleInput is the input to specify a single module to fetch.
type GetTerraformModuleInput struct {
	ID string
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
	ID            string `json:"id"`
	Name          string `json:"name"`
	System        string `json:"system"`
	RepositoryURL string `json:"repositoryUrl"`
	Private       bool   `json:"private"`
}
