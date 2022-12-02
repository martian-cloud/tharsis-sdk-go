package tharsis

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/paginators"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Workspaces implements functions related to Tharsis workspaces.
type Workspaces interface {
	GetWorkspace(ctx context.Context, input *types.GetWorkspaceInput) (*types.Workspace, error)
	GetWorkspaces(ctx context.Context, input *types.GetWorkspacesInput) (*types.GetWorkspacesOutput, error)
	GetWorkspacePaginator(ctx context.Context, input *types.GetWorkspacesInput) (*GetWorkspacesPaginator, error)
	CreateWorkspace(ctx context.Context, workspace *types.CreateWorkspaceInput) (*types.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *types.UpdateWorkspaceInput) (*types.Workspace, error)
	DeleteWorkspace(ctx context.Context, workspace *types.DeleteWorkspaceInput) error
	SetWorkspaceVariables(ctx context.Context, input *types.SetNamespaceVariablesInput) error
	GetAssignedManagedIdentities(ctx context.Context, input *types.GetAssignedManagedIdentitiesInput) ([]types.ManagedIdentity, error)
}

type workspaces struct {
	client *Client
}

// NewWorkspaces returns a Workspaces.
func NewWorkspaces(client *Client) Workspaces {
	return &workspaces{client: client}
}

func (ws *workspaces) GetWorkspace(ctx context.Context, input *types.GetWorkspaceInput) (*types.Workspace, error) {
	switch {
	case input.Path != nil:
		// Workspace query by path.

		var target struct {
			Workspace *graphQLWorkspace `graphql:"workspace(fullPath: $fullPath)"`
		}
		variables := map[string]interface{}{
			"fullPath": graphql.String(*input.Path),
		}

		err := ws.client.graphqlClient.Query(ctx, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Workspace == nil {
			return nil, nil
		}

		result, err := workspaceFromGraphQL(*target.Workspace)
		if err != nil {
			return nil, err
		}

		return result, nil
	case input.ID != nil:
		// Node query by ID.

		var target struct {
			Node *struct {
				ID        graphql.String
				Workspace graphQLWorkspace `graphql:"...on Workspace"`
			} `graphql:"node(id: $id)"`
		}

		variables := map[string]interface{}{"id": graphql.String(*input.ID)}

		err := ws.client.graphqlClient.Query(ctx, &target, variables)
		if err != nil {
			return nil, err
		}

		if target.Node == nil || target.Node.Workspace.ID == "" {
			return nil, nil
		}

		return workspaceFromGraphQL(target.Node.Workspace)
	default:

		// Didn't ask for anything; won't get anything.
		return nil, nil
	}
}

// GetWorkspaces returns a list of workspace objects.
//
// Based on the 'first' and 'after' fields of the PaginationOptions within the GetWorkspacesInput,
// it returns the first 'first' items after the 'after' element.  That can be equivalent to the
// first page from a paged query.
func (ws *workspaces) GetWorkspaces(ctx context.Context,
	input *types.GetWorkspacesInput) (*types.GetWorkspacesOutput, error) {

	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := getWorkspaces(ctx, ws.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	// Convert and repackage the type-specific results.
	workspaceResults := make([]types.Workspace, len(queryStruct.Workspaces.Edges))
	for ix, workspaceCustom := range queryStruct.Workspaces.Edges {
		workspace, err := workspaceFromGraphQL(workspaceCustom.Node)
		if err != nil {
			return nil, err
		}
		workspaceResults[ix] = *workspace
	}

	return &types.GetWorkspacesOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.Workspaces.TotalCount),
			HasNextPage: bool(queryStruct.Workspaces.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.Workspaces.PageInfo.EndCursor),
		},
		Workspaces: workspaceResults,
	}, nil
}

func (ws *workspaces) GetWorkspacePaginator(ctx context.Context,
	input *types.GetWorkspacesInput) (*GetWorkspacesPaginator, error) {

	paginator := newWorkspacePaginator(*ws.client, input)
	return &paginator, nil
}

func (ws *workspaces) CreateWorkspace(ctx context.Context,
	input *types.CreateWorkspaceInput) (*types.Workspace, error) {

	var wrappedCreate struct {
		CreateWorkspace struct {
			Problems  []internal.GraphQLProblem
			Workspace graphQLWorkspace
		} `graphql:"createWorkspace(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := ws.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedCreate.CreateWorkspace.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems creating workspace: %v", err)
	}

	created, err := workspaceFromGraphQL(wrappedCreate.CreateWorkspace.Workspace)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (ws *workspaces) UpdateWorkspace(ctx context.Context,
	input *types.UpdateWorkspaceInput) (*types.Workspace, error) {

	var wrappedUpdate struct {
		UpdateWorkspace struct {
			Problems  []internal.GraphQLProblem
			Workspace graphQLWorkspace
		} `graphql:"updateWorkspace(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ws.client.graphqlClient.Mutate(ctx, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}
	err = internal.ProblemsToError(wrappedUpdate.UpdateWorkspace.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems updating workspace: %v", err)
	}

	updated, err := workspaceFromGraphQL(wrappedUpdate.UpdateWorkspace.Workspace)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (ws *workspaces) DeleteWorkspace(ctx context.Context,
	input *types.DeleteWorkspaceInput) error {

	var wrappedDelete struct {
		DeleteWorkspace struct {
			// It appears it's not possible to return the deleted object.
			// Workspace internal.GraphQLWorkspace
			Problems []internal.GraphQLProblem
		} `graphql:"deleteWorkspace(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ws.client.graphqlClient.Mutate(ctx, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	err = internal.ProblemsToError(wrappedDelete.DeleteWorkspace.Problems)
	if err != nil {
		return fmt.Errorf("problems deleting workspace: %v", err)
	}

	return nil
}

func (ws *workspaces) SetWorkspaceVariables(ctx context.Context, input *types.SetNamespaceVariablesInput) error {
	var wrappedSet struct {
		SetNamespaceVariables struct {
			Problems []internal.GraphQLProblem
		} `graphql:"setNamespaceVariables(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := ws.client.graphqlClient.Mutate(ctx, &wrappedSet, variables)
	if err != nil {
		return err
	}

	err = internal.ProblemsToError(wrappedSet.SetNamespaceVariables.Problems)
	if err != nil {
		return fmt.Errorf("problems setting workspace variables: %v", err)
	}

	return nil
}

func (ws *workspaces) GetAssignedManagedIdentities(ctx context.Context,
	input *types.GetAssignedManagedIdentitiesInput) ([]types.ManagedIdentity, error) {
	var target struct {
		Workspace *struct {
			ManagedIdentities []GraphQLManagedIdentity `graphql:"assignedManagedIdentities"`
		} `graphql:"workspace(fullPath: $fullPath)"`
	}

	variables := map[string]interface{}{
		"fullPath": graphql.String(input.Path),
	}

	err := ws.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.Workspace == nil {
		return nil, nil
	}

	return sliceManagedIdentitiesFromGraphQL(target.Workspace.ManagedIdentities), nil
}

//////////////////////////////////////////////////////////////////////////////

// The GetWorkspaces paginator:

// GetWorkspacesPaginator is a type-specific paginator.
type GetWorkspacesPaginator struct {
	generic paginators.Paginator
}

// newWorkspacePaginator returns a new workspace paginator.
func newWorkspacePaginator(client Client, input *types.GetWorkspacesInput) GetWorkspacesPaginator {
	inputCopy := &types.GetWorkspacesInput{
		Sort:              input.Sort,
		PaginationOptions: input.PaginationOptions,
		Filter:            input.Filter,
	}

	// First return value is a GetWorkspacesOutput, which implements PaginatedResponse.
	queryCallback := func(ctx context.Context, after *string) (interface{}, error) {
		inputCopy.PaginationOptions.Cursor = after
		return client.Workspaces.GetWorkspaces(ctx, inputCopy)
	}

	genericPaginator := paginators.NewPaginator(queryCallback)

	return GetWorkspacesPaginator{
		generic: genericPaginator,
	}
}

// HasMore returns a boolean, whether there is another page (or more):
func (wp *GetWorkspacesPaginator) HasMore() bool {
	return wp.generic.HasMore()
}

// Next returns the next page of results:
func (wp *GetWorkspacesPaginator) Next(ctx context.Context) (*types.GetWorkspacesOutput, error) {

	// The generic paginator runs the query.
	untyped, err := wp.generic.Next(ctx)
	if err != nil {
		return nil, err
	}

	// We know the returned data is a *GetWorkspacesOutput:
	return untyped.(*types.GetWorkspacesOutput), nil
}

//////////////////////////////////////////////////////////////////////////////

// getWorkspaces runs the query and returns the results.
func getWorkspaces(ctx context.Context, client graphql.Client,
	input *types.GetWorkspacesInput, after *string) (*getWorkspacesQuery, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getWorkspacesQuery{}

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
	var groupPath *graphql.String
	if input.Filter != nil {
		groupPathString := graphql.String(*input.Filter.GroupPath)
		groupPath = &groupPathString
	} else {
		groupPath = nil
	}
	variables["groupPath"] = groupPath

	type WorkspaceSort string
	variables["sort"] = WorkspaceSort(*input.Sort)

	// Now, do the query.
	err := client.Query(ctx, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	return queryStructP, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getWorkspacesQuery is the query structure for GetWorkspaces.
// It contains the tag with the include-everything argument list.
type getWorkspacesQuery struct {
	Workspaces struct {
		PageInfo struct {
			EndCursor   graphql.String
			HasNextPage graphql.Boolean
		}
		Edges      []struct{ Node graphQLWorkspace }
		TotalCount graphql.Int
	} `graphql:"workspaces(first: $first, after: $after, groupPath: $groupPath, sort: $sort)"`
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLWorkspace represents the insides of the query structure,
// everything in the workspace object, and with graphql types.
type graphQLWorkspace struct {
	CurrentStateVersion *GraphQLStateVersion
	Metadata            internal.GraphQLMetadata
	ID                  graphql.String
	Name                graphql.String
	Description         graphql.String
	FullPath            graphql.String
	TerraformVersion    graphql.String
	MaxJobDuration      graphql.Int
	PreventDestroyPlan  graphql.Boolean
}

// workspaceFromGraphQL converts a GraphQL Workspace to an external Workspace.
func workspaceFromGraphQL(g graphQLWorkspace) (*types.Workspace, error) {
	currentStateVersion, err := stateVersionFromGraphQL(g.CurrentStateVersion)
	if err != nil {
		return nil, err
	}

	return &types.Workspace{
		Metadata:            internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Name:                string(g.Name),
		FullPath:            string(g.FullPath),
		Description:         string(g.Description),
		CurrentStateVersion: currentStateVersion,
		MaxJobDuration:      int32(g.MaxJobDuration),
		TerraformVersion:    string(g.TerraformVersion),
		PreventDestroyPlan:  bool(g.PreventDestroyPlan),
	}, nil
}

// The End.
