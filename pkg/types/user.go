package types

// Supporting structs for the User paginator:

// UserSortableField represents the fields that a user can be sorted by
type UserSortableField string

// UserSortableField constants
const (
	UserSortableFieldUpdatedAtAsc  UserSortableField = "UPDATED_AT_ASC"
	UserSortableFieldUpdatedAtDesc UserSortableField = "UPDATED_AT_DESC"
)

// UserFilter contains the supported fields for filtering User resources
type UserFilter struct {
	Search *string
}

// GetUsersInput is the input for listing users
type GetUsersInput struct {
	// Sort specifies the field to sort on and direction
	Sort *UserSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// Filter is used to filter the results
	Filter *UserFilter
}

// GetUsersOutput is the output when listing users
type GetUsersOutput struct {
	PageInfo *PageInfo
	Users    []User
}

// GetPageInfo allows GetUsersOutput to implement the PaginatedResponse interface.
func (guo *GetUsersOutput) GetPageInfo() *PageInfo {
	return guo.PageInfo
}

//////////////////////////////////////////////////////////////////////////////

// User represents a Tharsis user.
type User struct {
	Username       string
	Email          string
	SCIMExternalID string
	Metadata       ResourceMetadata
	Admin          bool
	Active         bool
}
