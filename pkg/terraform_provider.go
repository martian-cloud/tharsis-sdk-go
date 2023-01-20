package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformProvider implements functions related to Tharsis providers.
type TerraformProvider interface {
	GetProvider(ctx context.Context, input *types.GetTerraformProviderInput) (*types.TerraformProvider, error)
	CreateProvider(ctx context.Context, input *types.CreateTerraformProviderInput) (*types.TerraformProvider, error)
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
		return nil, newError(ErrNotFound, "terraform provider with id %s not found", input.ID)
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

	if err = errorFromGraphqlProblems(wrappedCreate.CreateTerraformProvider.Problems); err != nil {
		return nil, err
	}

	created := providerFromGraphQL(wrappedCreate.CreateTerraformProvider.Provider)
	return &created, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformProvider struct {
	ID                graphql.String
	Metadata          internal.GraphQLMetadata
	Name              string
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
		ResourcePath:      p.ResourcePath,
		RegistryNamespace: p.RegistryNamespace,
		Private:           p.Private,
		RepositoryURL:     p.RepositoryURL,
	}
	return result
}

// The End.
