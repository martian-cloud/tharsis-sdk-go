package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// FederatedRegistry implements functions related to federated registry features.
type FederatedRegistry interface {
	CreateFederatedRegistryTokens(ctx context.Context, input *types.CreateFederatedRegistryTokensInput) ([]types.FederatedRegistryToken, error)
}

type federatedRegistry struct {
	client *Client
}

// NewFederatedRegistry returns a FederatedRegistry.
func NewFederatedRegistry(client *Client) FederatedRegistry {
	return &federatedRegistry{client: client}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Method to create a federated registry token:

// CreateFederatedRegistryTokens creates one or more new federated registry tokens and returns their content.
func (p *federatedRegistry) CreateFederatedRegistryTokens(ctx context.Context,
	input *types.CreateFederatedRegistryTokensInput,
) ([]types.FederatedRegistryToken, error) {

	var wrappedCreate struct {
		CreateFederatedRegistryTokens struct {
			Tokens   []graphQLFederatedRegistryToken
			Problems []internal.GraphQLProblem
		} `graphql:"createFederatedRegistryTokens(input: $input)"`
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

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateFederatedRegistryTokens.Problems); err != nil {
		return nil, err
	}

	created := make([]types.FederatedRegistryToken, len(wrappedCreate.CreateFederatedRegistryTokens.Tokens))
	for ix, token := range wrappedCreate.CreateFederatedRegistryTokens.Tokens {
		created[ix] = federatedRegistryTokenFromGraphQL(token)
	}

	return created, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLFederatedRegistryToken represents the insides of the query structure,
type graphQLFederatedRegistryToken struct {
	Hostname graphql.String
	Token    graphql.String
}

// federatedRegistryTokenFromGraphQL converts a GraphQL FederatedRegistryToken to an external FederatedRegistryToken.
func federatedRegistryTokenFromGraphQL(p graphQLFederatedRegistryToken) types.FederatedRegistryToken {
	result := types.FederatedRegistryToken{
		Hostname: string(p.Hostname),
		Token:    string(p.Token),
	}
	return result
}
