package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/paginators"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// User implements functions related to Tharsis users.
type User interface {
	GetUsers(ctx context.Context, input *types.GetUsersInput) (*types.GetUsersOutput, error)
	GetUserPaginator(ctx context.Context, input *types.GetUsersInput) (*UserPaginator, error)
}

type user struct {
	client *Client
}

// NewUser returns a User.
func NewUser(client *Client) User {
	return &user{client: client}
}

// GetUsers returns a list of user objects.
//
// Based on the 'first' and 'after' fields of the PaginationOptions within the GetUsersInput,
// it returns the first 'first' items after the 'after' element.  That can be equivalent to the
//
//	first page from a paged query.
func (u *user) GetUsers(ctx context.Context, input *types.GetUsersInput) (*types.GetUsersOutput, error) {

	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := getUsers(ctx, u.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	// Convert and repackage the type-specific results.
	userResults := make([]types.User, len(queryStruct.Users.Edges))
	for ix, userCustom := range queryStruct.Users.Edges {
		userResults[ix] = userFromGraphQL(userCustom.Node)
	}

	return &types.GetUsersOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.Users.TotalCount),
			HasNextPage: bool(queryStruct.Users.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.Users.PageInfo.EndCursor),
		},
		Users: userResults,
	}, nil

}

func (u *user) GetUserPaginator(_ context.Context, input *types.GetUsersInput) (*UserPaginator, error) {
	paginator := newUserPaginator(*u.client, input)
	return &paginator, nil
}

//////////////////////////////////////////////////////////////////////////////

// The GetUsers paginator:

// UserPaginator is a type-specific paginator.
type UserPaginator struct {
	generic paginators.Paginator
}

// newUserPaginator returns a new user paginator.
func newUserPaginator(client Client, input *types.GetUsersInput) UserPaginator {
	inputCopy := &types.GetUsersInput{
		Sort:              input.Sort,
		PaginationOptions: input.PaginationOptions,
		Filter:            input.Filter,
	}

	// First return value is a GetUsersOutput, which implements PaginatedResponse.
	queryCallback := func(ctx context.Context, after *string) (interface{}, error) {
		inputCopy.PaginationOptions.Cursor = after
		return client.User.GetUsers(ctx, inputCopy)
	}

	genericPaginator := paginators.NewPaginator(queryCallback)

	return UserPaginator{
		generic: genericPaginator,
	}
}

// HasMore returns a boolean, whether there is another page (or more):
func (gp *UserPaginator) HasMore() bool {
	return gp.generic.HasMore()
}

// Next returns the next page of results:
func (gp *UserPaginator) Next(ctx context.Context) (*types.GetUsersOutput, error) {

	// The generic paginator runs the query.
	untyped, err := gp.generic.Next(ctx)
	if err != nil {
		return nil, err
	}

	// We know the returned data is a *GetUsersOutput:
	return untyped.(*types.GetUsersOutput), nil
}

//////////////////////////////////////////////////////////////////////////////

// getUsers runs the query and returns the results.
func getUsers(ctx context.Context, client graphqlClient,
	input *types.GetUsersInput, after *string) (*getUsersQuery, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getUsersQuery{}

	// Build the variables for filtering, sorting, and pagination.
	variables := map[string]interface{}{}

	// Shared input variables--possible candidates to factor out:
	if input.PaginationOptions.Limit != nil {
		variables["first"] = graphql.Int(*input.PaginationOptions.Limit)
	}
	if input.PaginationOptions.Cursor == nil {
		variables["after"] = (*graphql.String)(nil)
	} else {
		variables["after"] = graphql.String(*input.PaginationOptions.Cursor)
	}

	// after overrides input
	if after != nil {
		variables["after"] = graphql.String(*after)
	}

	// Resource type specific settings:

	// Make sure to pass the expected types for these variables.
	var search *graphql.String
	if (input.Filter != nil) && (input.Filter.Search != nil) {
		searchString := graphql.String(*input.Filter.Search)
		search = &searchString
	} else {
		search = nil
	}
	variables["search"] = search

	if input.Sort != nil {
		type UserSort string
		variables["sort"] = UserSort(*input.Sort)
	}

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	return queryStructP, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getUsersQuery is the query structure for GetUsers.
// It contains the tag with the include-everything argument list.
type getUsersQuery struct {
	Users struct {
		PageInfo struct {
			EndCursor   graphql.String
			HasNextPage graphql.Boolean
		}
		Edges      []struct{ Node graphQLUser }
		TotalCount graphql.Int
	} `graphql:"users(first: $first, after: $after, search: $search, sort: $sort)"`
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLUser represents a Tharsis user with graphQL types.
type graphQLUser struct {
	ID             graphql.String
	Metadata       internal.GraphQLMetadata
	Username       graphql.String
	Email          graphql.String
	SCIMExternalID graphql.String
	Admin          graphql.Boolean
	Active         graphql.Boolean
}

// userFromGraphQL converts a graphQL user to an external Tharsis user.
func userFromGraphQL(g graphQLUser) types.User {
	return types.User{
		Metadata:       internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Username:       string(g.Username),
		Email:          string(g.Email),
		SCIMExternalID: string(g.SCIMExternalID),
		Admin:          bool(g.Admin),
		Active:         bool(g.Active),
	}
}
