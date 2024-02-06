package types

// User represents a Tharsis user.
type User struct {
	Username       string
	Email          string
	SCIMExternalID string
	Metadata       ResourceMetadata
	Admin          bool
	Active         bool
}
