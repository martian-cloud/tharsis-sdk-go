package types

// WorkspaceVCSProviderLink holds the information about a VCS Provider.
type WorkspaceVCSProviderLink struct {
	// ID resides in the metadata
	Metadata            ResourceMetadata
	CreatedBy           string
	WorkspaceID         string
	WorkspacePath       string
	VCSProviderID       string
	RepositoryPath      string
	WebhookID           *string
	ModuleDirectory     *string
	Branch              string
	TagRegex            *string
	GlobPatterns        []string
	AutoSpeculativePlan bool
	WebhookDisabled     bool
}

// GetWorkspaceVCSProviderLinkInput is the input for retrieving a workspace VCS provider link.
type GetWorkspaceVCSProviderLinkInput struct {
	ID string `json:"id"`
}

// CreateWorkspaceVCSProviderLinkInput is the input to create a VCS Provider.
type CreateWorkspaceVCSProviderLinkInput struct {
	ModuleDirectory     *string  `json:"moduleDirectory"`
	RepositoryPath      string   `json:"repositoryPath"`
	WorkspacePath       string   `json:"workspacePath"`
	ProviderID          string   `json:"providerId"`
	Branch              *string  `json:"branch"`
	TagRegex            *string  `json:"tagRegex"`
	GlobPatterns        []string `json:"globPatterns"`
	AutoSpeculativePlan bool     `json:"autoSpeculativePlan"`
	WebhookDisabled     bool     `json:"webhookDisabled"`
}

// UpdateWorkspaceVCSProviderLinkInput is the input for creating a new workspace VCS provider link.
type UpdateWorkspaceVCSProviderLinkInput struct {
	ID                  string   `json:"id"`
	ModuleDirectory     *string  `json:"moduleDirectory"`
	Branch              *string  `json:"branch"`
	TagRegex            *string  `json:"tagRegex"`
	GlobPatterns        []string `json:"globPatterns"`
	AutoSpeculativePlan bool     `json:"autoSpeculativePlan"`
	WebhookDisabled     bool     `json:"webhookDisabled"`
}

// DeleteWorkspaceVCSProviderLinkInput is the input for deleting a workspace VCS provider link.
type DeleteWorkspaceVCSProviderLinkInput struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}

// CreateWorkspaceVCSProviderLinkResponse is the output from creating a VCS Provider.
type CreateWorkspaceVCSProviderLinkResponse struct {
	WebhookToken    *string                  `json:"webhookToken"`
	WebhookURL      *string                  `json:"webhookUrl"`
	VCSProviderLink WorkspaceVCSProviderLink `json:"vcsProviderLink"`
}

// The End.
