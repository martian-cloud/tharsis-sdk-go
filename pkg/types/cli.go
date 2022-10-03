package types

// CreateTerraformCLIDownloadURLInput is the input for CreateTerraformCLIDownloadURL.
type CreateTerraformCLIDownloadURLInput struct {
	Version      string `json:"version"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}
