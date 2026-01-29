package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/paginators"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Workspaces implements functions related to Tharsis workspaces.
type Workspaces interface {
	GetWorkspace(ctx context.Context, input *types.GetWorkspaceInput) (*types.Workspace, error)
	GetWorkspaces(ctx context.Context, input *types.GetWorkspacesInput) (*types.GetWorkspacesOutput, error)
	GetWorkspacePaginator(ctx context.Context, input *types.GetWorkspacesInput) (*GetWorkspacesPaginator, error)
	GetProviderMirrorEnabled(ctx context.Context, input *types.GetWorkspaceProviderMirrorEnabledInput) (*types.NamespaceProviderMirrorEnabled, error)
	CreateWorkspace(ctx context.Context, workspace *types.CreateWorkspaceInput) (*types.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *types.UpdateWorkspaceInput) (*types.Workspace, error)
	DeleteWorkspace(ctx context.Context, workspace *types.DeleteWorkspaceInput) error
	DestroyWorkspace(ctx context.Context, workspace *types.DestroyWorkspaceInput) (*types.Run, error)
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

		err := ws.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Workspace == nil {
			return nil, errors.NewError(types.ErrNotFound, "workspace with path %s not found", *input.Path)
		}

		return workspaceFromGraphQL(*target.Workspace)
	case input.ID != nil:
		// Node query by ID (supports both UUIDs and TRNs).

		var target struct {
			Node *struct {
				Workspace graphQLWorkspace `graphql:"...on Workspace"`
			} `graphql:"node(id: $id)"`
		}

		variables := map[string]interface{}{"id": graphql.String(*input.ID)}

		err := ws.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}

		if target.Node == nil {
			return nil, errors.NewError(types.ErrNotFound, "workspace with id %s not found", *input.ID)
		}

		return workspaceFromGraphQL(target.Node.Workspace)
	default:
		return nil, errors.NewError(types.ErrBadRequest, "must specify path or ID when calling GetWorkspace")
	}
}

// GetWorkspaces returns a list of workspace objects.
//
// Based on the 'first' and 'after' fields of the PaginationOptions within the GetWorkspacesInput,
// it returns the first 'first' items after the 'after' element.  That can be equivalent to the
// first page from a paged query.
func (ws *workspaces) GetWorkspaces(ctx context.Context,
	input *types.GetWorkspacesInput,
) (*types.GetWorkspacesOutput, error) {
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

func (ws *workspaces) GetWorkspacePaginator(_ context.Context,
	input *types.GetWorkspacesInput,
) (*GetWorkspacesPaginator, error) {
	paginator := newWorkspacePaginator(*ws.client, input)
	return &paginator, nil
}

// GetProviderMirrorEnabled retrieves the provider mirror enabled setting for a workspace.
func (ws *workspaces) GetProviderMirrorEnabled(ctx context.Context,
	input *types.GetWorkspaceProviderMirrorEnabledInput,
) (*types.NamespaceProviderMirrorEnabled, error) {
	var target struct {
		Node *struct {
			Workspace struct {
				ProviderMirrorEnabled struct {
					Inherited     bool
					NamespacePath string
					Value         bool
				}
			} `graphql:"...on Workspace"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := ws.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "workspace with id %s not found", input.ID)
	}

	return &types.NamespaceProviderMirrorEnabled{
		Inherited:     target.Node.Workspace.ProviderMirrorEnabled.Inherited,
		NamespacePath: target.Node.Workspace.ProviderMirrorEnabled.NamespacePath,
		Value:         target.Node.Workspace.ProviderMirrorEnabled.Value,
	}, nil
}

func (ws *workspaces) CreateWorkspace(ctx context.Context,
	input *types.CreateWorkspaceInput,
) (*types.Workspace, error) {
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

	err := ws.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateWorkspace.Problems); err != nil {
		return nil, err
	}

	created, err := workspaceFromGraphQL(wrappedCreate.CreateWorkspace.Workspace)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (ws *workspaces) UpdateWorkspace(ctx context.Context,
	input *types.UpdateWorkspaceInput,
) (*types.Workspace, error) {
	// Check that exactly one of ID and WorkspacePath are set in order to properly find the workspace to update.
	if (input.ID == nil) && (input.WorkspacePath == nil) {
		// Neither supplied.  Must have one.
		return nil, errors.NewError(types.ErrBadRequest, "must specify either ID or WorkspacePath")
	}
	if (input.ID != nil) && (input.WorkspacePath != nil) {
		// Both supplied.  Must have only one.
		return nil, errors.NewError(types.ErrBadRequest, "must specify only one of ID and WorkspacePath, not both")
	}

	var wrappedUpdate struct {
		UpdateWorkspace struct {
			Problems  []internal.GraphQLProblem
			Workspace graphQLWorkspace
		} `graphql:"updateWorkspace(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ws.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateWorkspace.Problems); err != nil {
		return nil, err
	}

	updated, err := workspaceFromGraphQL(wrappedUpdate.UpdateWorkspace.Workspace)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (ws *workspaces) DeleteWorkspace(ctx context.Context,
	input *types.DeleteWorkspaceInput,
) error {
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

	err := ws.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteWorkspace.Problems)
}

func (ws *workspaces) DestroyWorkspace(ctx context.Context, input *types.DestroyWorkspaceInput) (*types.Run, error) {
	var wrappedDestroy struct {
		DestroyWorkspace struct {
			Problems []internal.GraphQLProblem
			Run      graphQLRun
		} `graphql:"destroyWorkspace(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ws.client.graphqlClient.Mutate(ctx, true, &wrappedDestroy, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedDestroy.DestroyWorkspace.Problems); err != nil {
		return nil, err
	}

	run := runFromGraphQL(wrappedDestroy.DestroyWorkspace.Run)
	return &run, nil
}

func (ws *workspaces) GetAssignedManagedIdentities(ctx context.Context,
	input *types.GetAssignedManagedIdentitiesInput,
) ([]types.ManagedIdentity, error) {
	switch {
	case input.Path != nil:
		// Workspace query by path.

		var target struct {
			Workspace *struct {
				ManagedIdentities []GraphQLManagedIdentity `graphql:"assignedManagedIdentities"`
			} `graphql:"workspace(fullPath: $fullPath)"`
		}
		variables := map[string]interface{}{
			"fullPath": graphql.String(*input.Path),
		}

		err := ws.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Workspace == nil {
			return nil, errors.NewError(types.ErrNotFound, "workspace with path %s not found", *input.Path)
		}

		return sliceManagedIdentitiesFromGraphQL(target.Workspace.ManagedIdentities), nil
	case input.ID != nil:
		// Node query by ID.

		var target struct {
			Node *struct {
				Workspace struct {
					ManagedIdentities []GraphQLManagedIdentity `graphql:"assignedManagedIdentities"`
				} `graphql:"...on Workspace"`
			} `graphql:"node(id: $id)"`
		}
		variables := map[string]interface{}{"id": graphql.String(*input.ID)}

		err := ws.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}

		if target.Node == nil {
			return nil, errors.NewError(types.ErrNotFound, "workspace with id %s not found", *input.ID)
		}

		return sliceManagedIdentitiesFromGraphQL(target.Node.Workspace.ManagedIdentities), nil
	default:
		return nil, errors.NewError(types.ErrBadRequest, "must specify path or ID when calling GetAssignedManagedIdentities")
	}
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
func getWorkspaces(ctx context.Context, client graphqlClient,
	input *types.GetWorkspacesInput, after *string,
) (*getWorkspacesQuery, error) {
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
	if input.Filter != nil && input.Filter.GroupPath != nil {
		groupPathString := graphql.String(*input.Filter.GroupPath)
		groupPath = &groupPathString
	} else {
		groupPath = nil
	}
	variables["groupPath"] = groupPath

	type WorkspaceSort string
	if input.Sort != nil {
		variables["sort"] = WorkspaceSort(*input.Sort)
	} else {
		variables["sort"] = (*WorkspaceSort)(nil)
	}

	// Filter for workspaces with a specific set of labels
	var labelFilters *WorkspaceLabelsFilter
	if input.Filter != nil && len(input.Filter.Labels) > 0 {
		labelFilters = &WorkspaceLabelsFilter{
			Labels: []types.WorkspaceLabelInput{},
		}

		for _, label := range input.Filter.Labels {
			labelFilters.Labels = append(labelFilters.Labels, types.WorkspaceLabelInput(label))
		}
	} else {
		labelFilters = nil
	}
	variables["labelFilter"] = labelFilters

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
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
	} `graphql:"workspaces(first: $first, after: $after, groupPath: $groupPath, sort: $sort, labelFilter: $labelFilter)"`
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
	GroupPath           graphql.String
	FullPath            graphql.String
	TerraformVersion    graphql.String
	MaxJobDuration      graphql.Int
	PreventDestroyPlan  graphql.Boolean
	Labels              []graphQLWorkspaceLabel
}

// graphQLWorkspaceLabel represents the insides of the query structure,
// everything in the workspace label object, and with graphql types.
type graphQLWorkspaceLabel struct {
	Key   graphql.String
	Value graphql.String
}

// WorkspaceLabelsFilter represents the insides of the query variable structure,
// everything in the workspace label filter object.
type WorkspaceLabelsFilter struct {
	Labels []types.WorkspaceLabelInput `json:"labels,omitempty"`
}

// workspaceFromGraphQL converts a GraphQL Workspace to an external Workspace.
func workspaceFromGraphQL(g graphQLWorkspace) (*types.Workspace, error) {
	currentStateVersion, err := stateVersionFromGraphQL(g.CurrentStateVersion)
	if err != nil {
		return nil, err
	}

	var labels map[string]string
	for _, label := range g.Labels {
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[string(label.Key)] = string(label.Value)
	}

	return &types.Workspace{
		Metadata:            internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Name:                string(g.Name),
		GroupPath:           string(g.GroupPath),
		FullPath:            string(g.FullPath),
		Description:         string(g.Description),
		CurrentStateVersion: currentStateVersion,
		MaxJobDuration:      int32(g.MaxJobDuration),
		TerraformVersion:    string(g.TerraformVersion),
		PreventDestroyPlan:  bool(g.PreventDestroyPlan),
		Labels:              labels,
	}, nil
}
