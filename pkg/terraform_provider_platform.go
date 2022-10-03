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
			ID                        graphql.String
			TerraformProviderPlatform graphQLTerraformProviderPlatform `graphql:"...on TerraformProviderPlatform"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil || target.Node.TerraformProviderPlatform.ID == "" {
		return nil, nil
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

	err := p.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedCreate.CreateTerraformProviderPlatform.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems creating provider platform: %v", err)
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
	if _, err := p.client.httpClient.Do(req); err != nil {
		return err
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
