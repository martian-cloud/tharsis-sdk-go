package tharsis

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformModuleVersion implements functions related to Tharsis module versions.
type TerraformModuleVersion interface {
	GetModuleVersion(ctx context.Context, input *types.GetTerraformModuleVersionInput) (*types.TerraformModuleVersion, error)
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
	var target struct {
		Node *struct {
			TerraformModuleVersion graphQLTerraformModuleVersion `graphql:"...on TerraformModuleVersion"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, newError(ErrNotFound, "module version with id %s not found", input.ID)
	}

	result := moduleVersionFromGraphQL(target.Node.TerraformModuleVersion)
	return &result, nil
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

	if err = errorFromGraphqlProblems(wrappedCreate.CreateTerraformModuleVersion.Problems); err != nil {
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

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteTerraformModuleVersion.Problems); err != nil {
		return err
	}

	return nil
}

func (p *moduleVersion) UploadModuleVersion(ctx context.Context, moduleVersionID string, reader io.Reader) error {
	url := fmt.Sprintf("%s/v1/module-registry/versions/%s/upload", p.client.cfg.Endpoint, moduleVersionID)
	req, err := http.NewRequest("PUT", url, reader)
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
		return errorFromHTTPResponse(resp)
	}

	return nil
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
