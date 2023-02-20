package tharsis

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformProviderPlatform implements functions related to Tharsis provider platforms.
type TerraformProviderPlatform interface {
	GetProviderPlatform(ctx context.Context, input *types.GetTerraformProviderPlatformInput) (*types.TerraformProviderPlatform, error)
	CreateProviderPlatform(ctx context.Context, input *types.CreateTerraformProviderPlatformInput) (*types.TerraformProviderPlatform, error)
	UploadProviderPlatformBinary(ctx context.Context, providerPlatformID string, reader io.Reader) error
}

type providerPlatform struct {
	client *Client
}

// NewTerraformProviderPlatform returns a TerraformProviderPlatform.
func NewTerraformProviderPlatform(client *Client) TerraformProviderPlatform {
	return &providerPlatform{client: client}
}

func (p *providerPlatform) GetProviderPlatform(ctx context.Context, input *types.GetTerraformProviderPlatformInput) (*types.TerraformProviderPlatform, error) {
	var target struct {
		Node *struct {
			TerraformProviderPlatform graphQLTerraformProviderPlatform `graphql:"...on TerraformProviderPlatform"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider platform with id %s not found", input.ID)
	}

	result := providerPlatformFromGraphQL(target.Node.TerraformProviderPlatform)
	return &result, nil
}

func (p *providerPlatform) CreateProviderPlatform(ctx context.Context, input *types.CreateTerraformProviderPlatformInput) (*types.TerraformProviderPlatform, error) {

	var wrappedCreate struct {
		CreateTerraformProviderPlatform struct {
			Problems         []internal.GraphQLProblem
			ProviderPlatform graphQLTerraformProviderPlatform
		} `graphql:"createTerraformProviderPlatform(input: $input)"`
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

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTerraformProviderPlatform.Problems); err != nil {
		return nil, err
	}

	created := providerPlatformFromGraphQL(wrappedCreate.CreateTerraformProviderPlatform.ProviderPlatform)
	return &created, nil
}

func (p *providerPlatform) UploadProviderPlatformBinary(ctx context.Context, providerPlatformID string, reader io.Reader) error {
	url := fmt.Sprintf(
		"%s/v1/provider-registry/platforms/%s/upload",
		p.client.cfg.Endpoint,
		providerPlatformID,
	)

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
		return errors.ErrorFromHTTPResponse(resp)
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conplatform functions:

type graphQLTerraformProviderPlatform struct {
	Metadata        internal.GraphQLMetadata
	ID              graphql.String
	OS              string
	Arch            string
	SHASum          string
	Filename        string
	ProviderVersion graphQLTerraformProviderVersion
	BinaryUploaded  bool
}

// providerPlatformFromGraphQL converts a GraphQL TerraformProviderPlatform to an external TerraformProviderPlatform.
func providerPlatformFromGraphQL(p graphQLTerraformProviderPlatform) types.TerraformProviderPlatform {
	result := types.TerraformProviderPlatform{
		Metadata:          internal.MetadataFromGraphQL(p.Metadata, p.ID),
		ProviderVersionID: string(p.ProviderVersion.ID),
		OperatingSystem:   p.OS,
		Architecture:      p.Arch,
		SHASum:            p.SHASum,
		Filename:          p.Filename,
		BinaryUploaded:    p.BinaryUploaded,
	}
	return result
}

// The End.
