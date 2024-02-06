package types

// GPGKey holds (most) information about a Tharsis GPG key.
type GPGKey struct {
	// ID resides in the metadata
	Metadata     ResourceMetadata
	CreatedBy    string
	ASCIIArmor   string
	Fingerprint  string
	GPGKeyID     string // string of hex digits of size to fit in a uint64
	GroupPath    string
	ResourcePath string
}

// GetGPGKeyInput is the input to specify a single GPG key to fetch.
type GetGPGKeyInput struct {
	ID string `json:"id"`
}

// CreateGPGKeyInput is the input for creating a new GPG key.
type CreateGPGKeyInput struct {
	ASCIIArmor string `json:"asciiArmor"`
	GroupPath  string `json:"groupPath"`
}

// DeleteGPGKeyInput is the input for deleting a GPG key.
type DeleteGPGKeyInput struct {
	ID string `json:"id"`
}
