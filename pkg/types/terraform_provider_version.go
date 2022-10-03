package types

// TerraformProviderVersion represents a Tharsis provider version.
type TerraformProviderVersion struct {
	Metadata                 ResourceMetadata
	ProviderID               string
	Version                  string
	GPGKeyID                 *string
	GPGASCIIArmor            *string
	Protocols                []string
	SHASumsUploaded          bool
	SHASumsSignatureUploaded bool
	ReadmeUploaded           bool
}

// GetTerraformProviderVersionInput is the input to specify a single provider version to fetch.
type GetTerraformProviderVersionInput struct {
	ID string
}

// CreateTerraformProviderVersionInput is the input for creating a new provider version.
type CreateTerraformProviderVersionInput struct {
	ProviderPath string   `json:"providerPath"`
	Version      string   `json:"version"`
	Protocols    []string `json:"protocols"`
}
