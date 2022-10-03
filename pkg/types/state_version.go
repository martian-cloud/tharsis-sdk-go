package types

import "github.com/zclconf/go-cty/cty"

// GetStateVersionInput is the input for retrieving a State Version for a workspace.
type GetStateVersionInput struct {
	WorkspacePath string
}

// CreateStateVersionInput is the input for creating a state version.
type CreateStateVersionInput struct {
	State string `json:"state"`
	RunID string `json:"runId"`
}

// DownloadStateVersionInput is the input for downloading a state version.
type DownloadStateVersionInput struct {
	ID string
}

// StateVersion represents a specific version of the the terraform state associated with a workspace
// It is used as input to and output from some operations.
type StateVersion struct {
	// ID resides in the metadata
	Metadata ResourceMetadata
	RunID    string
	Outputs  []StateVersionOutput
}

// StateVersionOutput represents a specific version of the the terraform state's outputs associated with a workspace
// ID resides in the metadata
type StateVersionOutput struct {
	Value     cty.Value
	Type      cty.Type
	Metadata  ResourceMetadata
	Name      string
	Sensitive bool
}

// The End.
