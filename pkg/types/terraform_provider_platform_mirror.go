package types

import "io"

// TerraformProviderPlatformMirror represents a Terraform provider platform mirror.
type TerraformProviderPlatformMirror struct {
	Metadata      ResourceMetadata
	VersionMirror TerraformProviderVersionMirror
	OS            string
	Arch          string
}

// GetTerraformProviderPlatformMirrorInput is the input for retrieving a TerraformProviderPlatformMirror.
type GetTerraformProviderPlatformMirrorInput struct {
	ID string `json:"id"`
}

// GetTerraformProviderPlatformMirrorsByVersionInput is the input for retrieving a
// list of TerraformProviderPlatformMirrors by the version mirror's ID.
type GetTerraformProviderPlatformMirrorsByVersionInput struct {
	VersionMirrorID string `json:"id"`
}

// DeleteTerraformProviderPlatformMirrorInput is the input for deleting a TerraformProviderPlatformMirror.
type DeleteTerraformProviderPlatformMirrorInput struct {
	ID string `json:"id"`
}

// UploadProviderPlatformPackageToMirrorInput is the input for uploading a Terraform provider package.
type UploadProviderPlatformPackageToMirrorInput struct {
	Reader          io.Reader
	VersionMirrorID string
	OS              string
	Arch            string
}
