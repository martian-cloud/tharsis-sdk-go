package types

// The Team type is (initially at least) used only for testing other resources.

// Team represents a team of (human) users.
type Team struct {
	Name           string
	Description    string
	SCIMExternalID string
	Metadata       ResourceMetadata
}

// TeamMember represents one team member.
type TeamMember struct {
	Metadata     ResourceMetadata
	UserID       string
	TeamID       string
	IsMaintainer bool
}

// CreateTeamInput is the input for creating a new team.
type CreateTeamInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddUserToTeamInput is the input for adding a member to a team.
type AddUserToTeamInput struct {
	Username     string `json:"username"`
	TeamName     string `json:"teamName"`
	IsMaintainer bool   `json:"isMaintainer"`
}

// DeleteTeamInput is the input for deleting a Team.
type DeleteTeamInput struct {
	Name string `json:"name"`
}
