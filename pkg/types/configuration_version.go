package types

// ConfigurationVersion holds information about a Tharsis configuration version.
// It is used as input to and output from some operations.
// ID resides in the metadata
type ConfigurationVersion struct {
	Metadata    ResourceMetadata
	Status      string
	WorkspaceID string
	Speculative bool
}

// GetConfigurationVersionInput is the input to specify
// a single configuration version to fetch or download.
type GetConfigurationVersionInput struct {
	ID string
}

// CreateConfigurationVersionInput is the input for creating a new configuration version.
type CreateConfigurationVersionInput struct {
	Speculative   *bool  `json:"speculative"`
	WorkspacePath string `json:"workspacePath"`
}

// GraphiQL does not mention any operation to update or destroy a configuration version.

// UploadConfigurationVersionInput is the input for uploading a new configuration version.
type UploadConfigurationVersionInput struct {
	WorkspacePath          string `json:"workspacePath"`
	ConfigurationVersionID string `json:"configurationVersionId"`
	DirectoryPath          string `json:"directoryPath"`
}

// The End.
