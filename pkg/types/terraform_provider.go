package types

// TerraformProvider represents a Terraform provider.
type TerraformProvider struct {
	Metadata          ResourceMetadata
	Name              string
	GroupPath         string
	ResourcePath      string
	RegistryNamespace string
	RepositoryURL     string
	Private           bool
}

// GetTerraformProviderInput is the input to specify a single provider to fetch.
type GetTerraformProviderInput struct {
	ID string `json:"id"`
}

// CreateTerraformProviderInput is the input for creating a new provider.
type CreateTerraformProviderInput struct {
	Name          string `json:"name"`
	GroupPath     string `json:"groupPath"`
	RepositoryURL string `json:"repositoryUrl"`
	Private       bool   `json:"private"`
}

// UpdateTerraformProviderInput is the input for updating a TF provider.
type UpdateTerraformProviderInput struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	RepositoryURL string `json:"repositoryUrl"`
	Private       bool   `json:"private"`
}

// DeleteTerraformProviderInput is the input for deleting a TF provider.
type DeleteTerraformProviderInput struct {
	ID string `json:"id"`
}
