package types

// TerraformProvider represents a Terraform provider.
type TerraformProvider struct {
	Metadata          ResourceMetadata
	Name              string
	ResourcePath      string
	RegistryNamespace string
	RepositoryURL     string
	Private           bool
}

// GetTerraformProviderInput is the input to specify a single provider to fetch.
type GetTerraformProviderInput struct {
	ID string
}

// CreateTerraformProviderInput is the input for creating a new provider.
type CreateTerraformProviderInput struct {
	Name          string `json:"name"`
	GroupPath     string `json:"groupPath"`
	RepositoryURL string `json:"repositoryUrl"`
	Private       bool   `json:"private"`
}
