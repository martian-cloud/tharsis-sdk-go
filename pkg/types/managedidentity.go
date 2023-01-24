package types

// ManagedIdentityType represents the supported managed identity types
type ManagedIdentityType string

// ManagedIdentityType constants
const (
	ManagedIdentityAzureFederated   ManagedIdentityType = "azure_federated"
	ManagedIdentityAWSFederated     ManagedIdentityType = "aws_federated"
	ManagedIdentityTharsisFederated ManagedIdentityType = "tharsis_federated"
)

// GetManagedIdentityInput is the input for retrieving
// a managed identity and/or its access rules.
type GetManagedIdentityInput struct {
	ID string `json:"id"`
}

// ManagedIdentityAccessRuleInput is the input for managed identity access rules
// when created at the same time as the managed identity.
type ManagedIdentityAccessRuleInput struct {
	RunStage               JobType  `json:"runStage"`
	AllowedUsers           []string `json:"allowedUsers"`
	AllowedServiceAccounts []string `json:"allowedServiceAccounts"`
	AllowedTeams           []string `json:"allowedTeams"`
}

// CreateManagedIdentityInput is the input for creating a managed identity.
type CreateManagedIdentityInput struct {
	Type        ManagedIdentityType              `json:"type"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	GroupPath   string                           `json:"groupPath"`
	Data        string                           `json:"data"`
	AccessRules []ManagedIdentityAccessRuleInput `json:"accessRules"`
}

// UpdateManagedIdentityInput is the input for updating a managed identity.
type UpdateManagedIdentityInput struct {
	Data        string `json:"data"`
	ID          string `json:"id"`
	Description string `json:"description"`
}

// DeleteManagedIdentityInput is the input for deleting a managed identity.
type DeleteManagedIdentityInput struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}

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
	Metadata      ResourceMetadata
	Type          ManagedIdentityType
	AliasSourceID *string
	ResourcePath  string
	Name          string
	Description   string
	Data          string
	CreatedBy     string
	IsAlias       bool
}

// ManagedIdentityAccessRule represents an access rule for a managed identity.
type ManagedIdentityAccessRule struct {
	Metadata               ResourceMetadata
	RunStage               JobType
	ManagedIdentityID      string
	AllowedUsers           []User
	AllowedServiceAccounts []ServiceAccount
	AllowedTeams           []Team
}

// GetManagedIdentityAccessRuleInput is the input for retrieving a managed identity access rule.
type GetManagedIdentityAccessRuleInput struct {
	ID string `json:"id"`
}

// CreateManagedIdentityAccessRuleInput is the input for creating a managed identity access rule.
type CreateManagedIdentityAccessRuleInput struct {
	ManagedIdentityID      string   `json:"managedIdentityId"`
	RunStage               JobType  `json:"runStage"`
	AllowedUsers           []string `json:"allowedUsers"`
	AllowedServiceAccounts []string `json:"allowedServiceAccounts"`
	AllowedTeams           []string `json:"allowedTeams"`
}

// UpdateManagedIdentityAccessRuleInput is the input for updating a managed identity access rule.
type UpdateManagedIdentityAccessRuleInput struct {
	ID                     string   `json:"id"`
	RunStage               JobType  `json:"runStage"`
	AllowedUsers           []string `json:"allowedUsers"`
	AllowedServiceAccounts []string `json:"allowedServiceAccounts"`
	AllowedTeams           []string `json:"allowedTeams"`
}

// DeleteManagedIdentityAccessRuleInput is the input for deleting a managed identity access rule.
type DeleteManagedIdentityAccessRuleInput struct {
	ID string `json:"id"`
}

// CreateManagedIdentityAliasInput is the input for creating a managed identity alias.
type CreateManagedIdentityAliasInput struct {
	AliasSourceID   *string `json:"aliasSourceId"`
	AliasSourcePath *string `json:"aliasSourcePath"`
	Name            string  `json:"name"`
	GroupPath       string  `json:"groupPath"`
}

// DeleteManagedIdentityAliasInput is the input for deleting a managed identity alias.
type DeleteManagedIdentityAliasInput struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}

// The End.
