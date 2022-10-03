// Package paginators handles pagination logic
package paginators

import (
	"context"
)

// This module defines the generic cursor-based paginator for the Tharsis SDK
// for Go. It is intended to serve for all queries that return lists that
// might need to be paginated.

// Paginator is the generic paginator struct with methods.
//
// One quirk of the go-graphql-client is it requires a _FRESH_ query
// structure for each page. Otherwise, the second page panics in reflect with
// a slice index out of range.
type Paginator struct {
	NextCursor    *string // nil means no next page unless hasDoneQuery is false
	queryCallback func(ctx context.Context, after *string) (interface{}, error)
	hasDoneQuery  bool // false means we don't know whether there is a next page
}

// NewPaginator creates a brand-new paginator.
func NewPaginator(queryCallback func(ctx context.Context, after *string) (interface{}, error)) Paginator {
	return Paginator{
		queryCallback: queryCallback,
		// Leave hasDoneQuery false and NextCursor nil.
	}
}

// HasMore tells whether a paginator has more pages to read.
// If no page has not yet been read, it will return true, even if the page will be empty.
func (p *Paginator) HasMore() bool {

	if !p.hasDoneQuery {
		// No query has yet been attempted, so there must be more, even if it's empty.
		return true
	}

	return p.NextCursor != nil
}

// Next returns the next page of results.
func (p *Paginator) Next(ctx context.Context) (interface{}, error) {

	// Do the query via the query callback.
	response, err := p.queryCallback(ctx, p.NextCursor)
	if err != nil {
		return nil, err
	}

	// Be sure to mark that we have done a query.
	p.hasDoneQuery = true

	// Get the next cursor, with a minor correction.
	pageInfo := (response.(PaginatedResponse)).GetPageInfo()
	nextCursor := &pageInfo.Cursor
	if !pageInfo.HasNextPage {
		// On the last page with data, the library returns HasNextPage == false,
		// but it also returns a cursor (which will return an empty page).
		// This forces the cursor to nil in that case, avoiding return of an wasted empty page.
		nextCursor = nil
	}
	p.NextCursor = nextCursor

	return response, nil
}

// The End.
