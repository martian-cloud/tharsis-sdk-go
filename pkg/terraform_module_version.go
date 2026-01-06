package tharsis

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformModuleVersion implements functions related to Tharsis module versions.
type TerraformModuleVersion interface {
	GetModuleVersion(ctx context.Context, input *types.GetTerraformModuleVersionInput) (*types.TerraformModuleVersion, error)
	GetModuleVersions(ctx context.Context, input *types.GetTerraformModuleVersionsInput) (*types.GetTerraformModuleVersionsOutput, error)
	CreateModuleVersion(ctx context.Context, input *types.CreateTerraformModuleVersionInput) (*types.TerraformModuleVersion, error)
	UploadModuleVersion(ctx context.Context, moduleVersionID string, reader io.Reader) error
	DeleteModuleVersion(ctx context.Context, input *types.DeleteTerraformModuleVersionInput) error
}

type moduleVersion struct {
	client *Client
}

// NewTerraformModuleVersion returns a TerraformModuleVersion.
func NewTerraformModuleVersion(client *Client) TerraformModuleVersion {
	return &moduleVersion{client: client}
}

// GetTerraformModuleVersion returns a module version
func (p *moduleVersion) GetModuleVersion(ctx context.Context, input *types.GetTerraformModuleVersionInput) (*types.TerraformModuleVersion, error) {
	switch {
	case input.ModulePath != nil:
		pathParts := strings.Split(*input.ModulePath, "/")
		if len(pathParts) < 3 {
			return nil, errors.NewError(types.ErrBadRequest, "module path %s is not valid", *input.ModulePath)
		}

		var target struct {
			TerraformModuleVersion *graphQLTerraformModuleVersion `graphql:"terraformModuleVersion(registryNamespace: $registryNamespace, moduleName: $moduleName, system: $system, version: $version)"`
		}

		variables := map[string]interface{}{
			"registryNamespace": graphql.String(pathParts[0]),
			"moduleName":        graphql.String(pathParts[len(pathParts)-2]),
			"system":            graphql.String(pathParts[len(pathParts)-1]),
		}

		var version *graphql.String
		if input.Version != nil {
			versionString := graphql.String(*input.Version)
			version = &versionString
		} else {
			version = nil
		}
		variables["version"] = version

		err := p.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.TerraformModuleVersion == nil {
			return nil, errors.NewError(types.ErrNotFound, "terraform module version with module path %s not found", *input.ModulePath)
		}

		result := moduleVersionFromGraphQL(*target.TerraformModuleVersion)
		return &result, nil
	case input.ID != nil:
		var target struct {
			Node *struct {
				TerraformModuleVersion graphQLTerraformModuleVersion `graphql:"...on TerraformModuleVersion"`
			} `graphql:"node(id: $id)"`
		}
		variables := map[string]interface{}{"id": graphql.String(*input.ID)}

		err := p.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if target.Node == nil {
			return nil, errors.NewError(types.ErrNotFound, "module version with id %s not found", *input.ID)
		}

		result := moduleVersionFromGraphQL(target.Node.TerraformModuleVersion)
		return &result, nil
	default:
		return nil, errors.NewError(types.ErrBadRequest, "must specify ID or ModulePath (optionally version) when calling GetModuleVersion")
	}
}

func (p *moduleVersion) GetModuleVersions(ctx context.Context, input *types.GetTerraformModuleVersionsInput) (*types.GetTerraformModuleVersionsOutput, error) {
	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := p.getTerraformModuleVersions(ctx, p.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	if queryStruct.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "module with id %s not found", input.TerraformModuleID)
	}

	// Convert and repackage the type-specific results.
	versionResults := make([]types.TerraformModuleVersion, len(queryStruct.Node.TerraformModule.Versions.Edges))
	for ix, versionCustom := range queryStruct.Node.TerraformModule.Versions.Edges {
		versionResults[ix] = moduleVersionFromGraphQL(versionCustom.Node)
	}

	return &types.GetTerraformModuleVersionsOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.Node.TerraformModule.Versions.TotalCount),
			HasNextPage: bool(queryStruct.Node.TerraformModule.Versions.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.Node.TerraformModule.Versions.PageInfo.EndCursor),
		},
		ModuleVersions: versionResults,
	}, nil
}

func (p *moduleVersion) CreateModuleVersion(ctx context.Context, input *types.CreateTerraformModuleVersionInput) (*types.TerraformModuleVersion, error) {
	var wrappedCreate struct {
		CreateTerraformModuleVersion struct {
			Problems      []internal.GraphQLProblem
			ModuleVersion graphQLTerraformModuleVersion
		} `graphql:"createTerraformModuleVersion(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTerraformModuleVersion.Problems); err != nil {
		return nil, err
	}

	created := moduleVersionFromGraphQL(wrappedCreate.CreateTerraformModuleVersion.ModuleVersion)
	return &created, nil
}

func (p *moduleVersion) DeleteModuleVersion(ctx context.Context, input *types.DeleteTerraformModuleVersionInput) error {
	var wrappedDelete struct {
		DeleteTerraformModuleVersion struct {
			Problems      []internal.GraphQLProblem
			ModuleVersion graphQLTerraformModuleVersion
		} `graphql:"deleteTerraformModuleVersion(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteTerraformModuleVersion.Problems)
}

func (p *moduleVersion) UploadModuleVersion(ctx context.Context, moduleVersionID string, reader io.Reader) error {
	url := fmt.Sprintf("%s/v1/module-registry/versions/%s/upload", p.client.cfg.Endpoint, moduleVersionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, reader)
	if err != nil {
		return err
	}

	// Get the authentication token.
	authToken, err := p.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	// Make the request.
	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.ErrorFromHTTPResponse(resp)
	}

	return nil
}

// getTerraformModuleVersions runs the query and returns the results.
func (p *moduleVersion) getTerraformModuleVersions(ctx context.Context, client graphqlClient,
	input *types.GetTerraformModuleVersionsInput, after *string) (*getTerraformModuleVersionsQuery, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getTerraformModuleVersionsQuery{}

	// Build the variables for filtering, sorting, and pagination.
	variables := map[string]interface{}{}

	variables["id"] = graphql.String(input.TerraformModuleID)

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

	type TerraformModuleVersionSort string
	if input.Sort != nil {
		variables["sort"] = TerraformModuleVersionSort(*input.Sort)
	} else {
		variables["sort"] = (*TerraformModuleVersionSort)(nil)
	}

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	return queryStructP, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getTerraformModuleVersionsQuery is the query structure for GetTerraformModuleVersions.
// It contains the tag with the include-everything argument list.
type getTerraformModuleVersionsQuery struct {
	Node *struct {
		TerraformModule struct {
			Versions struct {
				PageInfo struct {
					EndCursor   graphql.String
					HasNextPage graphql.Boolean
				}
				Edges      []struct{ Node graphQLTerraformModuleVersion }
				TotalCount graphql.Int
			} `graphql:"versions(first: $first, after: $after, sort: $sort)"`
		} `graphql:"...on TerraformModule"`
	} `graphql:"node(id: $id)"`
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformModuleVersion struct {
	Metadata    internal.GraphQLMetadata
	ID          graphql.String
	Version     string
	SHASum      string
	Status      string
	Error       string
	Diagnostics string
	Module      graphQLTerraformModule
	Submodules  []string
	Examples    []string
	Latest      bool
}

// moduleVersionFromGraphQL converts a GraphQL TerraformModuleVersion to an external TerraformModuleVersion.
func moduleVersionFromGraphQL(p graphQLTerraformModuleVersion) types.TerraformModuleVersion {
	result := types.TerraformModuleVersion{
		Metadata:    internal.MetadataFromGraphQL(p.Metadata, p.ID),
		ModuleID:    string(p.Module.ID),
		Version:     p.Version,
		SHASum:      p.SHASum,
		Status:      p.Status,
		Error:       p.Error,
		Diagnostics: p.Diagnostics,
		Submodules:  p.Submodules,
		Examples:    p.Examples,
		Latest:      p.Latest,
	}
	return result
}
