package types

// PaginationOptions contain the cursor based pagination options
// Our paginators support only forward paging, not reverse.
type PaginationOptions struct {
	Limit  *int32
	Cursor *string
}

// PageInfo contains all three fields common to all queries that can be paginated.
//
// Please note that the internal struct called PageInfo inside the query
// structure is lacking the totalCount field.
type PageInfo struct {
	Cursor      string
	TotalCount  int
	HasNextPage bool
}
