package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformProvider implements functions related to Tharsis providers.
type TerraformProvider interface {
	GetProvider(ctx context.Context, input *types.GetTerraformProviderInput) (*types.TerraformProvider, error)
	CreateProvider(ctx context.Context, input *types.CreateTerraformProviderInput) (*types.TerraformProvider, error)
	UpdateProvider(ctx context.Context, input *types.UpdateTerraformProviderInput) (*types.TerraformProvider, error)
	DeleteProvider(ctx context.Context, input *types.DeleteTerraformProviderInput) (*types.TerraformProvider, error)
}

type provider struct {
	client *Client
}

// NewTerraformProvider returns a TerraformProvider.
func NewTerraformProvider(client *Client) TerraformProvider {
	return &provider{client: client}
}

// GetTerraformProvider returns a provider
func (p *provider) GetProvider(ctx context.Context, input *types.GetTerraformProviderInput) (*types.TerraformProvider, error) {
	var target struct {
		Node *struct {
			Provider graphQLTerraformProvider `graphql:"...on TerraformProvider"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider with id %s not found", input.ID)
	}

	result := providerFromGraphQL(target.Node.Provider)
	return &result, nil
}

func (p *provider) CreateProvider(ctx context.Context, input *types.CreateTerraformProviderInput) (*types.TerraformProvider, error) {

	var wrappedCreate struct {
		CreateTerraformProvider struct {
			Provider graphQLTerraformProvider
			Problems []internal.GraphQLProblem
		} `graphql:"createTerraformProvider(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err := errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTerraformProvider.Problems); err != nil {
		return nil, err
	}

	created := providerFromGraphQL(wrappedCreate.CreateTerraformProvider.Provider)
	return &created, nil
}

func (p *provider) UpdateProvider(ctx context.Context, input *types.UpdateTerraformProviderInput) (*types.TerraformProvider, error) {

	var wrappedUpdate struct {
		UpdateTerraformProvider struct {
			Provider graphQLTerraformProvider
			Problems []internal.GraphQLProblem
		} `graphql:"updateTerraformProvider(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err := errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateTerraformProvider.Problems); err != nil {
		return nil, err
	}

	updated := providerFromGraphQL(wrappedUpdate.UpdateTerraformProvider.Provider)
	return &updated, nil
}

func (p *provider) DeleteProvider(ctx context.Context, input *types.DeleteTerraformProviderInput) (*types.TerraformProvider, error) {

	var wrappedDelete struct {
		DeleteTerraformProvider struct {
			Provider graphQLTerraformProvider
			Problems []internal.GraphQLProblem
		} `graphql:"deleteTerraformProvider(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return nil, err
	}

	if err := errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteTerraformProvider.Problems); err != nil {
		return nil, err
	}

	deleted := providerFromGraphQL(wrappedDelete.DeleteTerraformProvider.Provider)
	return &deleted, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformProvider struct {
	ID                graphql.String
	Metadata          internal.GraphQLMetadata
	Name              string
	GroupPath         string
	ResourcePath      string
	RegistryNamespace string
	RepositoryURL     string
	Private           bool
}

// providerFromGraphQL converts a GraphQL TerraformProvider to an external TerraformProvider.
func providerFromGraphQL(p graphQLTerraformProvider) types.TerraformProvider {
	result := types.TerraformProvider{
		Metadata:          internal.MetadataFromGraphQL(p.Metadata, p.ID),
		Name:              p.Name,
		GroupPath:         p.GroupPath,
		ResourcePath:      p.ResourcePath,
		RegistryNamespace: p.RegistryNamespace,
		Private:           p.Private,
		RepositoryURL:     p.RepositoryURL,
	}
	return result
}
