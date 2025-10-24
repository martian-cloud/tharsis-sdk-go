package types

// VCSProviderType represents the supported VCS provider types
type VCSProviderType string

// VCSProviderType constants
const (
	VCSProviderTypeGitlab VCSProviderType = "gitlab"
	VCSProviderTypeGithub VCSProviderType = "github"
)

// VCSProvider holds the information about a VCS Provider.
type VCSProvider struct {
	// ID resides in the metadata
	Metadata           ResourceMetadata
	CreatedBy          string
	Name               string
	Description        string
	URL                string
	GroupPath          string
	ResourcePath       string
	Type               VCSProviderType
	AutoCreateWebhooks bool
}

// GetVCSProviderInput is the input for retrieving a VCS provider.
type GetVCSProviderInput struct {
	ID  string  `json:"id"`
	TRN *string `json:"trn"`
}

// CreateVCSProviderInput is the input to create a VCS Provider.
type CreateVCSProviderInput struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	GroupPath          string          `json:"groupPath"`
	URL                *string         `json:"url"`
	OAuthClientID      string          `json:"oAuthClientId"`
	OAuthClientSecret  string          `json:"oAuthClientSecret"`
	Type               VCSProviderType `json:"type"`
	AutoCreateWebhooks bool            `json:"autoCreateWebhooks"`
}

// UpdateVCSProviderInput is the input for creating a new VCS provider.
type UpdateVCSProviderInput struct {
	Description       *string `json:"description"`
	OAuthClientID     *string `json:"oAuthClientId"`
	OAuthClientSecret *string `json:"oAuthClientSecret"`
	ID                string  `json:"id"`
}

// DeleteVCSProviderInput is the input for deleting a VCS provider.
type DeleteVCSProviderInput struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}

// CreateVCSProviderResponse is the response from creating a new VCS provider.
type CreateVCSProviderResponse struct {
	VCSProvider           *VCSProvider
	OAuthAuthorizationURL string
}
