package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// GPGKey implements functions related to Tharsis GPG keys.
type GPGKey interface {
	GetGPGKey(ctx context.Context, input *types.GetGPGKeyInput) (*types.GPGKey, error)
	CreateGPGKey(ctx context.Context, input *types.CreateGPGKeyInput) (*types.GPGKey, error)
	DeleteGPGKey(ctx context.Context, input *types.DeleteGPGKeyInput) (*types.GPGKey, error)
}

type gpgKey struct {
	client *Client
}

// NewGPGKey returns a GPG key.
func NewGPGKey(client *Client) GPGKey {
	return &gpgKey{client: client}
}

// GetGPGKey returns everything about the GPG key.
func (gk *gpgKey) GetGPGKey(ctx context.Context, input *types.GetGPGKeyInput) (*types.GPGKey, error) {

	if input.ID == "" {
		return nil, errors.NewError(types.ErrBadRequest, "must specify ID when calling GetGPGKey")
	}

	// Node query by ID (supports both UUIDs and TRNs).
	var target struct {
		Node *struct {
			GPGKey graphQLGPGKey `graphql:"...on GPGKey"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := gk.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "GPG key with id %s not found", input.ID)
	}

	gotKey := gpgKeyFromGraphQL(target.Node.GPGKey)
	return &gotKey, nil
}

// CreateGPGKey creates a new GPG key and returns its content.
func (gk *gpgKey) CreateGPGKey(ctx context.Context, input *types.CreateGPGKeyInput) (*types.GPGKey, error) {

	var wrappedCreate struct {
		CreateGPGKey struct {
			GPGKey   graphQLGPGKey
			Problems []internal.GraphQLProblem
		} `graphql:"createGPGKey(input: $input)"`
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

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateGPGKey.Problems); err != nil {
		return nil, err
	}

	createdKey := gpgKeyFromGraphQL(wrappedCreate.CreateGPGKey.GPGKey)
	return &createdKey, nil
}

func (gk *gpgKey) DeleteGPGKey(ctx context.Context, input *types.DeleteGPGKeyInput) (*types.GPGKey, error) {

	var wrappedDelete struct {
		DeleteGPGKey struct {
			GPGKey   graphQLGPGKey
			Problems []internal.GraphQLProblem
		} `graphql:"deleteGPGKey(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := gk.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteGPGKey.Problems); err != nil {
		return nil, err
	}

	deletedKey := gpgKeyFromGraphQL(wrappedDelete.DeleteGPGKey.GPGKey)
	return &deletedKey, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLGPGKey represents (most of) the insides of the query structure, with graphql types.
type graphQLGPGKey struct {
	ID           graphql.String
	Metadata     internal.GraphQLMetadata
	CreatedBy    graphql.String
	ASCIIArmor   graphql.String
	Fingerprint  graphql.String
	GPGKeyID     graphql.String
	GroupPath    graphql.String
	ResourcePath graphql.String
}

// gpgKeyFromGraphQL converts a GraphQL GPG key to an external GPG key.
func gpgKeyFromGraphQL(ggk graphQLGPGKey) types.GPGKey {
	result := types.GPGKey{
		Metadata:     internal.MetadataFromGraphQL(ggk.Metadata, ggk.ID),
		CreatedBy:    string(ggk.CreatedBy),
		ASCIIArmor:   string(ggk.ASCIIArmor),
		Fingerprint:  string(ggk.Fingerprint),
		GPGKeyID:     string(ggk.GPGKeyID),
		GroupPath:    string(ggk.GroupPath),
		ResourcePath: string(ggk.ResourcePath),
	}
	return result
}
