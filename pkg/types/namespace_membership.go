package types

// NamespaceMembership holds information about a Tharsis group or workspace membership.
type NamespaceMembership struct {
	// ID resides in the metadata
	Metadata         ResourceMetadata
	UserID           *string
	ServiceAccountID *string
	TeamID           *string
	Role             string
}

// CreateNamespaceMembershipInput is the input for adding a membership to a group or workspace.
type CreateNamespaceMembershipInput struct {
	NamespacePath    string  `json:"namespacePath"`
	Username         *string `json:"username"`
	ServiceAccountID *string `json:"serviceAccountId"`
	TeamName         *string `json:"teamName"`
	Role             string  `json:"role"`
}

// UpdateNamespaceMembershipInput is the input for updating a membership on a group or workspace.
type UpdateNamespaceMembershipInput struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// DeleteNamespaceMembershipInput is the input for updating a membership from a group or workspace.
type DeleteNamespaceMembershipInput struct {
	ID string `json:"id"`
}
