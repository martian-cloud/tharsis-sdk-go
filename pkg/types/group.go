package types

// Supporting structs for the Group paginator:

// GroupSortableField represents the fields that a group can be sorted by
type GroupSortableField string

// GroupSortableField constants
const (
	GroupSortableFieldFullPathAsc  GroupSortableField = "FULL_PATH_ASC"
	GroupSortableFieldFullPathDesc GroupSortableField = "FULL_PATH_DESC"
)

// GroupFilter contains the supported fields for filtering Group resources
type GroupFilter struct {
	ParentPath *string
}

// GetGroupsInput is the input for listing groups
type GetGroupsInput struct {
	// Sort specifies the field to sort on and direction
	Sort *GroupSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// Filter is used to filter the results
	Filter *GroupFilter
}

// GetGroupsOutput is the output when listing groups
type GetGroupsOutput struct {
	PageInfo *PageInfo
	Groups   []Group
}

// GetPageInfo allows GetGroupsOutput to implement the PaginatedResponse interface.
func (ggo *GetGroupsOutput) GetPageInfo() *PageInfo {
	return ggo.PageInfo
}

//////////////////////////////////////////////////////////////////////////////

// Group holds (most) information about a Tharsis group.
// It is used as input to and output from some operations.
//
// See below for structs that handle DescendentGroups and Workspaces.
type Group struct {
	// ID resides in the metadata
	Metadata    ResourceMetadata
	Name        string
	Description string
	FullPath    string
}

// GetGroupInput is the input to specify a single group to fetch.
type GetGroupInput struct {
	Path *string
	ID   *string
	TRN  *string
}

// CreateGroupInput is the input for creating a new group.
type CreateGroupInput struct {
	Name        string  `json:"name"`
	ParentPath  *string `json:"parentPath"` // is allowed to be nil
	Description string  `json:"description"`
}

// UpdateGroupInput is the input for updating a group.
// One (and only one) of ID, GroupPath, or TRN finds the group to update.
// Description is modified.
type UpdateGroupInput struct {
	GroupPath   *string `json:"groupPath"`
	ID          *string `json:"id"`
	TRN         *string `json:"trn"`
	Description string  `json:"description"`
}

// DeleteGroupInput is the input for deleting a group.
type DeleteGroupInput struct {
	Force     *bool   `json:"force"`
	GroupPath *string `json:"groupPath"`
	ID        *string `json:"id"`
	TRN       *string `json:"trn"`
}

// MigrateGroupInput is the input for migrating a group.
// One (and only one) of ID and GroupPath finds the group to migrate.
// If NewParentPath is nil, that means move the group to top-level.
type MigrateGroupInput struct {
	NewParentPath *string `json:"newParentPath"`
	GroupPath     string  `json:"groupPath"`
}
