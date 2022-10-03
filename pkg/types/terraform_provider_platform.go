package types

// TerraformProviderPlatform represents a Tharsis provider platform.
type TerraformProviderPlatform struct {
	Metadata          ResourceMetadata
	ProviderVersionID string
	OperatingSystem   string
	Architecture      string
	SHASum            string
	Filename          string
	BinaryUploaded    bool
}

// GetTerraformProviderPlatformInput is the input to specify a single provider platform to fetch.
type GetTerraformProviderPlatformInput struct {
	ID string
}

// CreateTerraformProviderPlatformInput is the input for creating a new provider platform.
type CreateTerraformProviderPlatformInput struct {
	ProviderVersionID string `json:"providerVersionId"`
	OperatingSystem   string `json:"os"`
	Architecture      string `json:"arch"`
	SHASum            string `json:"shaSum"`
	Filename          string `json:"filename"`
}
