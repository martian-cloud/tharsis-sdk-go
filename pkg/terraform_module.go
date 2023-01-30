package tharsis

import (
	"context"
	"errors"
	"strings"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformModule implements functions related to Tharsis modules.
type TerraformModule interface {
	GetModule(ctx context.Context, input *types.GetTerraformModuleInput) (*types.TerraformModule, error)
	GetModules(ctx context.Context, input *types.GetTerraformModulesInput) (*types.GetTerraformModulesOutput, error)
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
	switch {
	case input.Path != nil:
		pathParts := strings.Split(*input.Path, "/")
		if len(pathParts) < 3 {
			return nil, errors.New("module path is not valid")
		}

		var target struct {
			TerraformModule *graphQLTerraformModule `graphql:"terraformModule(registryNamespace: $registryNamespace, moduleName: $moduleName, system: $system)"`
		}
		variables := map[string]interface{}{
			"registryNamespace": graphql.String(pathParts[0]),
			"moduleName":        graphql.String(pathParts[len(pathParts)-2]),
			"system":            graphql.String(pathParts[len(pathParts)-1]),
		}

		err := p.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.TerraformModule == nil {
			return nil, newError(ErrNotFound, "terraform module with path %s not found", *input.Path)
		}

		result := moduleFromGraphQL(*target.TerraformModule)
		return &result, nil
	case input.ID != nil:
		var target struct {
			Node *struct {
				Module graphQLTerraformModule `graphql:"...on TerraformModule"`
			} `graphql:"node(id: $id)"`
		}
		variables := map[string]interface{}{"id": graphql.String(*input.ID)}

		err := p.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Node == nil {
			return nil, newError(ErrNotFound, "module with id %s not found", *input.ID)
		}

		result := moduleFromGraphQL(target.Node.Module)
		return &result, nil
	default:
		return nil, newError(ErrBadRequest, "must specify ID or path must be specified when calling GetModule")
	}
}
func (p *module) GetModules(ctx context.Context, input *types.GetTerraformModulesInput) (*types.GetTerraformModulesOutput, error) {
	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := p.getTerraformModules(ctx, p.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	// Convert and repackage the type-specific results.
	moduleResults := make([]types.TerraformModule, len(queryStruct.TerraformModules.Edges))
	for ix, moduleCustom := range queryStruct.TerraformModules.Edges {
		moduleResults[ix] = moduleFromGraphQL(moduleCustom.Node)
	}

	return &types.GetTerraformModulesOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.TerraformModules.TotalCount),
			HasNextPage: bool(queryStruct.TerraformModules.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.TerraformModules.PageInfo.EndCursor),
		},
		TerraformModules: moduleResults,
	}, nil
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

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
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

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
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

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteTerraformModule.Problems); err != nil {
		return err
	}

	return nil
}

// getTerraformModules runs the query and returns the results.
func (p *module) getTerraformModules(ctx context.Context, client graphqlClient,
	input *types.GetTerraformModulesInput, after *string) (*getTerraformModulesQuery, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getTerraformModulesQuery{}

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
	var search *graphql.String
	if input.Filter != nil {
		searchString := graphql.String(*input.Filter.Search)
		search = &searchString
	} else {
		search = nil
	}
	variables["search"] = search

	type TerraformModuleSort string
	variables["sort"] = TerraformModuleSort(*input.Sort)

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	return queryStructP, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getTerraformModulesQuery is the query structure for GetTerraformModules.
// It contains the tag with the include-everything argument list.
type getTerraformModulesQuery struct {
	TerraformModules struct {
		PageInfo struct {
			EndCursor   graphql.String
			HasNextPage graphql.Boolean
		}
		Edges      []struct{ Node graphQLTerraformModule }
		TotalCount graphql.Int
	} `graphql:"terraformModules(first: $first, after: $after, search: $search, sort: $sort)"`
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
