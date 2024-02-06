package types

// Team represents a team of (human) users
type Team struct {
	Name           string
	Description    string
	SCIMExternalID string
	Metadata       ResourceMetadata
}
