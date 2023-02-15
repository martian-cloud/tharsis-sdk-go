package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// WorkspaceVCSProviderLink implements functions related to Tharsis workspace VCS provider links.
type WorkspaceVCSProviderLink interface {
	GetLink(ctx context.Context,
		input *types.GetWorkspaceVCSProviderLinkInput) (*types.WorkspaceVCSProviderLink, error)
	CreateLink(ctx context.Context,
		input *types.CreateWorkspaceVCSProviderLinkInput) (*types.CreateWorkspaceVCSProviderLinkResponse, error)
	UpdateLink(ctx context.Context,
		input *types.UpdateWorkspaceVCSProviderLinkInput) (*types.WorkspaceVCSProviderLink, error)
	DeleteLink(ctx context.Context,
		input *types.DeleteWorkspaceVCSProviderLinkInput) (*types.WorkspaceVCSProviderLink, error)
}

type workspaceVCSProviderLink struct {
	client *Client
}

// NewWorkspaceVCSProviderLink returns a workspace VCS provider link.
func NewWorkspaceVCSProviderLink(client *Client) WorkspaceVCSProviderLink {
	return &workspaceVCSProviderLink{client: client}
}

// GetLink returns everything about the workspace VCS provider link.
func (gk *workspaceVCSProviderLink) GetLink(ctx context.Context,
	input *types.GetWorkspaceVCSProviderLinkInput) (*types.WorkspaceVCSProviderLink, error) {

	// Node query by ID.
	var target struct {
		Node *struct {
			WorkspaceVCSProviderLink graphQLWorkspaceVCSProviderLink `graphql:"...on WorkspaceVCSProviderLink"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := gk.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, newError(ErrNotFound, "workspace VCS provider link with id %s not found", input.ID)
	}

	gotLink := workspaceVCSProviderLinkFromGraphQL(target.Node.WorkspaceVCSProviderLink)
	return &gotLink, nil
}

// CreateLink creates a new workspace VCS provider link and returns its content.
func (gk *workspaceVCSProviderLink) CreateLink(ctx context.Context,
	input *types.CreateWorkspaceVCSProviderLinkInput) (*types.CreateWorkspaceVCSProviderLinkResponse, error) {

	var wrappedCreate struct {
		CreateWorkspaceVCSProviderLink struct {
			VCSProviderLink graphQLWorkspaceVCSProviderLink
			WebhookToken    *graphql.String
			WebhookURL      *graphql.String
			Problems        []internal.GraphQLProblem
		} `graphql:"createWorkspaceVCSProviderLink(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := gk.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateWorkspaceVCSProviderLink.Problems); err != nil {
		return nil, err
	}

	createdLink := workspaceVCSProviderLinkFromGraphQL(
		wrappedCreate.CreateWorkspaceVCSProviderLink.VCSProviderLink)
	result := types.CreateWorkspaceVCSProviderLinkResponse{
		VCSProviderLink: createdLink,
		WebhookToken:    internal.StringPointerFromGraphQL(wrappedCreate.CreateWorkspaceVCSProviderLink.WebhookToken),
		WebhookURL:      internal.StringPointerFromGraphQL(wrappedCreate.CreateWorkspaceVCSProviderLink.WebhookURL),
	}
	return &result, nil
}

// UpdateLink updates a workspace VCS provider link and returns its content.
func (gk *workspaceVCSProviderLink) UpdateLink(ctx context.Context,
	input *types.UpdateWorkspaceVCSProviderLinkInput) (*types.WorkspaceVCSProviderLink, error) {

	var wrappedUpdate struct {
		UpdateWorkspaceVCSProviderLink struct {
			WorkspaceVCSProviderLink graphQLWorkspaceVCSProviderLink
			Problems                 []internal.GraphQLProblem
		} `graphql:"updateWorkspaceVCSProviderLink(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := gk.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedUpdate.UpdateWorkspaceVCSProviderLink.Problems); err != nil {
		return nil, err
	}

	updatedLink := workspaceVCSProviderLinkFromGraphQL(wrappedUpdate.UpdateWorkspaceVCSProviderLink.WorkspaceVCSProviderLink)
	return &updatedLink, nil
}

// DeleteLink deletes a workspace VCS provider link and returns the content of the now-deleted object.
func (gk *workspaceVCSProviderLink) DeleteLink(ctx context.Context,
	input *types.DeleteWorkspaceVCSProviderLinkInput) (*types.WorkspaceVCSProviderLink, error) {

	var wrappedDelete struct {
		DeleteWorkspaceVCSProviderLink struct {
			WorkspaceVCSProviderLink graphQLWorkspaceVCSProviderLink
			Problems                 []internal.GraphQLProblem
		} `graphql:"deleteWorkspaceVCSProviderLink(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := gk.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteWorkspaceVCSProviderLink.Problems); err != nil {
		return nil, err
	}

	deletedLink := workspaceVCSProviderLinkFromGraphQL(wrappedDelete.DeleteWorkspaceVCSProviderLink.WorkspaceVCSProviderLink)
	return &deletedLink, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLWorkspaceVCSProviderLink represents (most of) the insides of the query structure, with graphql types.
type graphQLWorkspaceVCSProviderLink struct {
	ID                  graphql.String
	Metadata            internal.GraphQLMetadata
	CreatedBy           graphql.String
	Workspace           graphQLWorkspace
	VCSProvider         graphQLVCSProvider
	RepositoryPath      graphql.String
	WebhookID           *graphql.String
	ModuleDirectory     *graphql.String
	Branch              graphql.String
	TagRegex            *graphql.String
	GlobPatterns        []graphql.String
	AutoSpeculativePlan graphql.Boolean
	WebhookDisabled     graphql.Boolean
}

// workspaceVCSProviderLinkFromGraphQL converts a GraphQL workspace VCS provider link
// to an external workspace VCS provider link.
func workspaceVCSProviderLinkFromGraphQL(g graphQLWorkspaceVCSProviderLink) types.WorkspaceVCSProviderLink {
	result := types.WorkspaceVCSProviderLink{
		Metadata:            internal.MetadataFromGraphQL(g.Metadata, g.ID),
		CreatedBy:           string(g.CreatedBy),
		WorkspaceID:         string(g.Workspace.ID),
		VCSProviderID:       string(g.VCSProvider.ID),
		RepositoryPath:      string(g.RepositoryPath),
		WebhookID:           internal.StringPointerFromGraphQL(g.WebhookID),
		ModuleDirectory:     internal.StringPointerFromGraphQL(g.ModuleDirectory),
		Branch:              string(g.Branch),
		TagRegex:            internal.StringPointerFromGraphQL(g.TagRegex),
		GlobPatterns:        internal.StringSliceFromGraphQL(g.GlobPatterns),
		AutoSpeculativePlan: bool(g.AutoSpeculativePlan),
		WebhookDisabled:     bool(g.WebhookDisabled),
	}
	return result
}

// The End.
