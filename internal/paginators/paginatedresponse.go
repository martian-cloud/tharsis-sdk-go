package paginators

import (
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module declares an interface to allow the generic paginator to
// extract information from the specific query output types.
//
// Each specific query output structure will need to implement this method.

// PaginatedResponse interface helps the generic paginator to get information from
// a resource-specific paginator.
type PaginatedResponse interface {
	GetPageInfo() *types.PageInfo
}
