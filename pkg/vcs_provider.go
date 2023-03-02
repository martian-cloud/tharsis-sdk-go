package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// VCSProvider implements functions related to Tharsis VCS Providers.
type VCSProvider interface {
	GetProvider(ctx context.Context, input *types.GetVCSProviderInput) (*types.VCSProvider, error)
	CreateProvider(ctx context.Context, input *types.CreateVCSProviderInput) (*types.CreateVCSProviderResponse, error)
	UpdateProvider(ctx context.Context, input *types.UpdateVCSProviderInput) (*types.VCSProvider, error)
	DeleteProvider(ctx context.Context, input *types.DeleteVCSProviderInput) (*types.VCSProvider, error)
}

type vcsProvider struct {
	client *Client
}

// NewVCSProvider returns a VCS provider.
func NewVCSProvider(client *Client) VCSProvider {
	return &vcsProvider{client: client}
}

func (vp *vcsProvider) GetProvider(ctx context.Context, input *types.GetVCSProviderInput) (*types.VCSProvider, error) {

	// Node query by ID.
	var target struct {
		Node *struct {
			VCSProvider graphQLVCSProvider `graphql:"...on VCSProvider"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := vp.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "VCS provider with id %s not found", input.ID)
	}

	gotVCSProvider := vcsProviderFromGraphQL(target.Node.VCSProvider)
	return &gotVCSProvider, nil
}

func (vp *vcsProvider) CreateProvider(ctx context.Context,
	input *types.CreateVCSProviderInput) (*types.CreateVCSProviderResponse, error) {

	var wrappedCreate struct {
		CreateVCSProvider struct {
			VCSProvider           graphQLVCSProvider
			OAuthAuthorizationURL graphql.String
			Problems              []internal.GraphQLProblem
		} `graphql:"createVCSProvider(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := vp.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateVCSProvider.Problems); err != nil {
		return nil, err
	}

	created := vcsProviderFromGraphQL(wrappedCreate.CreateVCSProvider.VCSProvider)
	return &types.CreateVCSProviderResponse{
		VCSProvider:           &created,
		OAuthAuthorizationURL: string(wrappedCreate.CreateVCSProvider.OAuthAuthorizationURL),
	}, nil
}

func (vp *vcsProvider) UpdateProvider(ctx context.Context, input *types.UpdateVCSProviderInput) (*types.VCSProvider, error) {

	var wrappedUpdate struct {
		UpdateVCSProvider struct {
			VCSProvider graphQLVCSProvider
			Problems    []internal.GraphQLProblem
		} `graphql:"updateVCSProvider(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := vp.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateVCSProvider.Problems); err != nil {
		return nil, err
	}

	updated := vcsProviderFromGraphQL(wrappedUpdate.UpdateVCSProvider.VCSProvider)
	return &updated, nil
}

func (vp *vcsProvider) DeleteProvider(ctx context.Context, input *types.DeleteVCSProviderInput) (*types.VCSProvider, error) {

	var wrappedDelete struct {
		DeleteVCSProvider struct {
			VCSProvider graphQLVCSProvider
			Problems    []internal.GraphQLProblem
		} `graphql:"deleteVCSProvider(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := vp.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteVCSProvider.Problems); err != nil {
		return nil, err
	}

	deleted := vcsProviderFromGraphQL(wrappedDelete.DeleteVCSProvider.VCSProvider)
	return &deleted, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLVCSProvider represents (most of) the insides of the query structure, with graphql types.
type graphQLVCSProvider struct {
	ID                 graphql.String
	Metadata           internal.GraphQLMetadata
	CreatedBy          graphql.String
	Name               graphql.String
	Description        graphql.String
	Hostname           graphql.String
	GroupPath          graphql.String
	ResourcePath       graphql.String
	Type               graphql.String
	AutoCreateWebhooks graphql.Boolean
}

// vcsProviderFromGraphQL converts a GraphQL VCS provider to an external VCS provider.
func vcsProviderFromGraphQL(gvp graphQLVCSProvider) types.VCSProvider {
	result := types.VCSProvider{
		Metadata:           internal.MetadataFromGraphQL(gvp.Metadata, gvp.ID),
		CreatedBy:          string(gvp.CreatedBy),
		Name:               string(gvp.Name),
		Description:        string(gvp.Description),
		Hostname:           string(gvp.Hostname),
		GroupPath:          string(gvp.GroupPath),
		ResourcePath:       string(gvp.ResourcePath),
		Type:               types.VCSProviderType(gvp.Type),
		AutoCreateWebhooks: bool(gvp.AutoCreateWebhooks),
	}
	return result
}

// The End.
