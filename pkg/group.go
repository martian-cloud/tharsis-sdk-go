package tharsis

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/paginators"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Group implements functions related to Tharsis groups.
type Group interface {
	GetGroup(ctx context.Context, input *types.GetGroupInput) (*types.Group, error)
	GetGroups(ctx context.Context, input *types.GetGroupsInput) (*types.GetGroupsOutput, error)
	GetGroupPaginator(ctx context.Context, input *types.GetGroupsInput) (*GroupPaginator, error)
	CreateGroup(ctx context.Context, input *types.CreateGroupInput) (*types.Group, error)
	UpdateGroup(ctx context.Context, input *types.UpdateGroupInput) (*types.Group, error)
	DeleteGroup(ctx context.Context, input *types.DeleteGroupInput) error
}

type group struct {
	client *Client
}

// NewGroup returns a Group.
func NewGroup(client *Client) Group {
	return &group{client: client}
}

// GetGroup returns everything about the group _EXCEPT_ the subgroups/descendentGroups and workspaces.
// There are separate calls to get each of those.
func (g *group) GetGroup(ctx context.Context, input *types.GetGroupInput) (*types.Group, error) {
	switch {
	case input.Path != nil:
		// Group query by path.

		var target struct {
			Group *graphQLGroup `graphql:"group(fullPath: $fullPath)"`
		}
		variables := map[string]interface{}{"fullPath": graphql.String(*input.Path)}

		err := g.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Group == nil {
			return nil, newError(ErrNotFound, "group with path %s not found", *input.Path)
		}

		result := groupFromGraphQL(*target.Group)
		return &result, nil
	case input.ID != nil:
		// Node query by ID.

		var target struct {
			Node *struct {
				Group graphQLGroup `graphql:"...on Group"`
			} `graphql:"node(id: $id)"`
		}
		variables := map[string]interface{}{"id": graphql.String(*input.ID)}

		err := g.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Node == nil {
			return nil, newError(ErrNotFound, "group with id %s not found", *input.ID)
		}

		result := groupFromGraphQL(target.Node.Group)
		return &result, nil
	default:
		// Didn't ask for anything.

		return nil, fmt.Errorf("must specify path or ID when calling GetGroup")
	}
}

// GetGroups returns a list of group objects.
//
// Based on the 'first' and 'after' fields of the PaginationOptions within the GetGroupsInput,
// it returns the first 'first' items after the 'after' element.  That can be equivalent to the
//
//	first page from a paged query.
func (g *group) GetGroups(ctx context.Context,
	input *types.GetGroupsInput) (*types.GetGroupsOutput, error) {

	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := getGroups(ctx, g.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	// Convert and repackage the type-specific results.
	groupResults := make([]types.Group, len(queryStruct.Groups.Edges))
	for ix, groupCustom := range queryStruct.Groups.Edges {
		groupResults[ix] = groupFromGraphQL(groupCustom.Node)
	}

	return &types.GetGroupsOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.Groups.TotalCount),
			HasNextPage: bool(queryStruct.Groups.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.Groups.PageInfo.EndCursor),
		},
		Groups: groupResults,
	}, nil
}

func (g *group) GetGroupPaginator(ctx context.Context,
	input *types.GetGroupsInput) (*GroupPaginator, error) {

	paginator := newGroupPaginator(*g.client, input)
	return &paginator, nil
}

// CreateGroup creates a new group and returns its content.
func (g *group) CreateGroup(ctx context.Context, input *types.CreateGroupInput) (*types.Group, error) {

	var wrappedCreate struct {
		CreateGroup struct {
			Group    graphQLGroup
			Problems []internal.GraphQLProblem
		} `graphql:"createGroup(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := g.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateGroup.Problems); err != nil {
		return nil, err
	}

	created := groupFromGraphQL(wrappedCreate.CreateGroup.Group)
	return &created, nil
}

// UpdateGroup updates a group and returns its content.
func (g *group) UpdateGroup(ctx context.Context, input *types.UpdateGroupInput) (*types.Group, error) {

	var wrappedUpdate struct {
		UpdateGroup struct {
			Group    graphQLGroup
			Problems []internal.GraphQLProblem
		} `graphql:"updateGroup(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := g.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedUpdate.UpdateGroup.Problems); err != nil {
		return nil, err
	}

	updated := groupFromGraphQL(wrappedUpdate.UpdateGroup.Group)
	return &updated, nil
}

func (g *group) DeleteGroup(ctx context.Context, input *types.DeleteGroupInput) error {

	var wrappedDelete struct {
		DeleteGroup struct {
			Problems []internal.GraphQLProblem
		} `graphql:"deleteGroup(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := g.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteGroup.Problems); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////

// The GetGroups paginator:

// GroupPaginator is a type-specific paginator.
type GroupPaginator struct {
	generic paginators.Paginator
}

// newGroupPaginator returns a new group paginator.
func newGroupPaginator(client Client, input *types.GetGroupsInput) GroupPaginator {
	inputCopy := &types.GetGroupsInput{
		Sort:              input.Sort,
		PaginationOptions: input.PaginationOptions,
		Filter:            input.Filter,
	}

	// First return value is a GetGroupsOutput, which implements PaginatedResponse.
	queryCallback := func(ctx context.Context, after *string) (interface{}, error) {
		inputCopy.PaginationOptions.Cursor = after
		return client.Group.GetGroups(ctx, inputCopy)
	}

	genericPaginator := paginators.NewPaginator(queryCallback)

	return GroupPaginator{
		generic: genericPaginator,
	}
}

// HasMore returns a boolean, whether there is another page (or more):
func (gp *GroupPaginator) HasMore() bool {
	return gp.generic.HasMore()
}

// Next returns the next page of results:
func (gp *GroupPaginator) Next(ctx context.Context) (*types.GetGroupsOutput, error) {

	// The generic paginator runs the query.
	untyped, err := gp.generic.Next(ctx)
	if err != nil {
		return nil, err
	}

	// We know the returned data is a *GetGroupsOutput:
	return untyped.(*types.GetGroupsOutput), nil
}

//////////////////////////////////////////////////////////////////////////////

// getGroups runs the query and returns the results.
func getGroups(ctx context.Context, client graphqlClient,
	input *types.GetGroupsInput, after *string) (*getGroupsQuery, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getGroupsQuery{}

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
	var parentPath *graphql.String
	if input.Filter != nil {
		parentPathString := graphql.String(*input.Filter.ParentPath)
		parentPath = &parentPathString
	} else {
		parentPath = nil
	}
	variables["parentPath"] = parentPath

	type GroupSort string
	variables["sort"] = GroupSort(*input.Sort)

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	return queryStructP, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getGroupsQuery is the query structure for GetGroups.
// It contains the tag with the include-everything argument list.
type getGroupsQuery struct {
	Groups struct {
		PageInfo struct {
			EndCursor   graphql.String
			HasNextPage graphql.Boolean
		}
		Edges      []struct{ Node graphQLGroup }
		TotalCount graphql.Int
	} `graphql:"groups(first: $first, after: $after, parentPath: $parentPath, sort: $sort)"`
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLGroup represents (most of) the insides of the query structure,
// everything (except descendent groups and workspaces) in the group object,
// and with graphql types.
//
// NOTE: Early on, DescendentGroups were represented here as []graphQLGroup.
// That caused the go-graphql-client library to go into infinite cross-recursion.
type graphQLGroup struct {
	ID          graphql.String
	Metadata    internal.GraphQLMetadata
	Name        graphql.String
	Description graphql.String
	FullPath    graphql.String
}

// groupFromGraphQL converts a GraphQL Group to an external Group.
func groupFromGraphQL(g graphQLGroup) types.Group {
	result := types.Group{
		Metadata:    internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Name:        string(g.Name),
		Description: string(g.Description),
		FullPath:    string(g.FullPath),
	}
	return result
}

// The End.
