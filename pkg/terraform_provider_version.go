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

// TerraformProviderVersion implements functions related to Tharsis providerVersions.
type TerraformProviderVersion interface {
	GetProviderVersion(ctx context.Context, input *types.GetTerraformProviderVersionInput) (*types.TerraformProviderVersion, error)
	CreateProviderVersion(ctx context.Context, input *types.CreateTerraformProviderVersionInput) (*types.TerraformProviderVersion, error)
	UploadProviderReadme(ctx context.Context, providerVersionID string, reader io.Reader) error
	UploadProviderChecksums(ctx context.Context, providerVersionID string, reader io.Reader) error
	UploadProviderChecksumSignature(ctx context.Context, providerVersionID string, reader io.Reader) error
}

type providerVersion struct {
	client *Client
}

// NewTerraformProviderVersion returns a TerraformProviderVersion.
func NewTerraformProviderVersion(client *Client) TerraformProviderVersion {
	return &providerVersion{client: client}
}

// GetTerraformProviderVersion returns a provider version
func (p *providerVersion) GetProviderVersion(ctx context.Context, input *types.GetTerraformProviderVersionInput) (*types.TerraformProviderVersion, error) {
	var target struct {
		Node *struct {
			TerraformProviderVersion graphQLTerraformProviderVersion `graphql:"...on TerraformProviderVersion"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider version with id %s not found", input.ID)
	}

	result := providerVersionFromGraphQL(target.Node.TerraformProviderVersion)
	return &result, nil
}

func (p *providerVersion) CreateProviderVersion(ctx context.Context, input *types.CreateTerraformProviderVersionInput) (*types.TerraformProviderVersion, error) {

	var wrappedCreate struct {
		CreateTerraformProviderVersion struct {
			Problems        []internal.GraphQLProblem
			ProviderVersion graphQLTerraformProviderVersion
		} `graphql:"createTerraformProviderVersion(input: $input)"`
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

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTerraformProviderVersion.Problems); err != nil {
		return nil, err
	}

	created := providerVersionFromGraphQL(wrappedCreate.CreateTerraformProviderVersion.ProviderVersion)
	return &created, nil
}

func (p *providerVersion) UploadProviderReadme(ctx context.Context, providerVersionID string, reader io.Reader) error {
	url := fmt.Sprintf("%s/v1/provider-registry/versions/%s/readme/upload", p.client.cfg.Endpoint, providerVersionID)
	return p.uploadProviderFile(ctx, reader, url)
}

func (p *providerVersion) UploadProviderChecksums(ctx context.Context, providerVersionID string, reader io.Reader) error {
	url := fmt.Sprintf("%s/v1/provider-registry/versions/%s/checksums/upload", p.client.cfg.Endpoint, providerVersionID)
	return p.uploadProviderFile(ctx, reader, url)
}

func (p *providerVersion) UploadProviderChecksumSignature(ctx context.Context, providerVersionID string, reader io.Reader) error {
	url := fmt.Sprintf("%s/v1/provider-registry/versions/%s/signature/upload", p.client.cfg.Endpoint, providerVersionID)
	return p.uploadProviderFile(ctx, reader, url)
}

func (p *providerVersion) uploadProviderFile(ctx context.Context, reader io.Reader, url string) error {
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

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformProviderVersion struct {
	ID                 graphql.String
	Metadata           internal.GraphQLMetadata
	Provider           graphQLTerraformProvider
	Version            string
	GPGKeyID           *string
	GPGASCIIArmor      *string `graphql:"gpgAsciiArmor"`
	Protocols          []string
	SHASumsUploaded    bool
	SHASumsSigUploaded bool
	ReadmeUploaded     bool
}

// providerVersionFromGraphQL converts a GraphQL TerraformProviderVersion to an external TerraformProviderVersion.
func providerVersionFromGraphQL(p graphQLTerraformProviderVersion) types.TerraformProviderVersion {
	result := types.TerraformProviderVersion{
		Metadata:                 internal.MetadataFromGraphQL(p.Metadata, p.ID),
		ProviderID:               string(p.Provider.ID),
		Version:                  p.Version,
		GPGKeyID:                 p.GPGKeyID,
		GPGASCIIArmor:            p.GPGASCIIArmor,
		Protocols:                p.Protocols,
		SHASumsUploaded:          p.SHASumsUploaded,
		SHASumsSignatureUploaded: p.SHASumsSigUploaded,
		ReadmeUploaded:           p.ReadmeUploaded,
	}
	return result
}
