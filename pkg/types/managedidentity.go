package types

// ManagedIdentityType represents the supported managed identity types
type ManagedIdentityType string

// ManagedIdentityType constants
const (
	ManagedIdentityAzureFederated ManagedIdentityType = "azure_federated"
	ManagedIdentityAWSFederated   ManagedIdentityType = "aws_federated"
)

// CreateManagedIdentityCredentialsInput is the input for creating managed identity credentials
type CreateManagedIdentityCredentialsInput struct {
	ID string `json:"id"`
}

// AssignManagedIdentityInput is the input for assigning a managed identity to a workspace.
type AssignManagedIdentityInput struct {
	ManagedIdentityID   *string `json:"managedIdentityId"`
	ManagedIdentityPath *string `json:"managedIdentityPath"`
	WorkspacePath       string  `json:"workspacePath"`
}

// ManagedIdentity holds information about a Tharsis managed identity.
// It is used as input to and output from some operations.
type ManagedIdentity struct {
	// ID resides in the metadata
	Metadata     ResourceMetadata
	Type         ManagedIdentityType
	ResourcePath string
	Name         string
	Description  string
	Data         string
}

// The End.
