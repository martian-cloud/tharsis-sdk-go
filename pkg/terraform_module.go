package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformModule implements functions related to Tharsis modules.
type TerraformModule interface {
	GetModule(ctx context.Context, input *types.GetTerraformModuleInput) (*types.TerraformModule, error)
	CreateModule(ctx context.Context, input *types.CreateTerraformModuleInput) (*types.TerraformModule, error)
	UpdateModule(ctx context.Context, input *types.UpdateTerraformModuleInput) (*types.TerraformModule, error)
	DeleteModule(ctx context.Context, input *types.DeleteTerraformModuleInput) error
}

type module struct {
	client *Client
}

// NewTerraformModule returns a TerraformModule.
func NewTerraformModule(client *Client) TerraformModule {
	return &module{client: client}
}

// GetTerraformModule returns a module
func (p *module) GetModule(ctx context.Context, input *types.GetTerraformModuleInput) (*types.TerraformModule, error) {
	var target struct {
		Node *struct {
			Module graphQLTerraformModule `graphql:"...on TerraformModule"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, newError(ErrNotFound, "module with id %s not found", input.ID)
	}

	result := moduleFromGraphQL(target.Node.Module)
	return &result, nil
}

func (p *module) CreateModule(ctx context.Context, input *types.CreateTerraformModuleInput) (*types.TerraformModule, error) {
	var wrappedCreate struct {
		CreateTerraformModule struct {
			Module   graphQLTerraformModule
			Problems []internal.GraphQLProblem
		} `graphql:"createTerraformModule(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateTerraformModule.Problems); err != nil {
		return nil, err
	}

	created := moduleFromGraphQL(wrappedCreate.CreateTerraformModule.Module)
	return &created, nil
}

func (p *module) UpdateModule(ctx context.Context, input *types.UpdateTerraformModuleInput) (*types.TerraformModule, error) {
	var wrappedUpdate struct {
		UpdateTerraformModule struct {
			Module   graphQLTerraformModule
			Problems []internal.GraphQLProblem
		} `graphql:"updateTerraformModule(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedUpdate.UpdateTerraformModule.Problems); err != nil {
		return nil, err
	}

	module := moduleFromGraphQL(wrappedUpdate.UpdateTerraformModule.Module)
	return &module, nil
}

func (p *module) DeleteModule(ctx context.Context, input *types.DeleteTerraformModuleInput) error {
	var wrappedDelete struct {
		DeleteTerraformModule struct {
			Module   graphQLTerraformModule
			Problems []internal.GraphQLProblem
		} `graphql:"deleteTerraformModule(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteTerraformModule.Problems); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformModule struct {
	ID                graphql.String           `json:"id"`
	Metadata          internal.GraphQLMetadata `json:"metadata"`
	Name              string                   `json:"name"`
	System            string                   `json:"system"`
	ResourcePath      string                   `json:"resourcePath"`
	RegistryNamespace string                   `json:"registryNamespace"`
	RepositoryURL     string                   `json:"repositoryUrl"`
	Private           bool                     `json:"private"`
}

// moduleFromGraphQL converts a GraphQL TerraformModule to an external TerraformModule.
func moduleFromGraphQL(p graphQLTerraformModule) types.TerraformModule {
	result := types.TerraformModule{
		Metadata:          internal.MetadataFromGraphQL(p.Metadata, p.ID),
		Name:              p.Name,
		System:            p.System,
		ResourcePath:      p.ResourcePath,
		RegistryNamespace: p.RegistryNamespace,
		Private:           p.Private,
		RepositoryURL:     p.RepositoryURL,
	}
	return result
}
