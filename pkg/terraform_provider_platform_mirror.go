package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformProviderPlatformMirror implements functionalities related to TerraformProviderPlatformMirrors.
type TerraformProviderPlatformMirror interface {
	GetProviderPlatformMirror(
		ctx context.Context,
		input *types.GetTerraformProviderPlatformMirrorInput,
	) (*types.TerraformProviderPlatformMirror, error)
	GetProviderPlatformMirrorsByVersion(
		ctx context.Context,
		input *types.GetTerraformProviderPlatformMirrorsByVersionInput,
	) ([]types.TerraformProviderPlatformMirror, error)
	DeleteProviderPlatformMirror(
		ctx context.Context,
		input *types.DeleteTerraformProviderPlatformMirrorInput,
	) error
	UploadProviderPlatformPackageToMirror(
		ctx context.Context,
		input *types.UploadProviderPlatformPackageToMirrorInput,
	) error
	GetProviderPlatformPackageDownloadURL(
		ctx context.Context,
		input *types.GetProviderPlatformPackageDownloadURLInput,
	) (*types.ProviderPlatformPackageInfo, error)
}

type providerPlatformMirror struct {
	client *Client
}

// NewTerraformProviderPlatformMirror returns a TerraformProviderPlatformMirror.
func NewTerraformProviderPlatformMirror(client *Client) TerraformProviderPlatformMirror {
	return &providerPlatformMirror{client: client}
}

// GetProviderPlatformMirror returns a TerraformProviderPlatformMirror by ID.
func (p *providerPlatformMirror) GetProviderPlatformMirror(
	ctx context.Context,
	input *types.GetTerraformProviderPlatformMirrorInput,
) (*types.TerraformProviderPlatformMirror, error) {
	var target struct {
		Node *struct {
			TerraformProviderPlatformMirror graphQLTerraformProviderPlatformMirror `graphql:"...on TerraformProviderPlatformMirror"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider platform mirror with id %s not found", input.ID)
	}

	result := providerPlatformMirrorFromGraphQL(target.Node.TerraformProviderPlatformMirror)
	return &result, nil
}

// GetProviderPlatformMirrorsByVersion returns a TerraformProviderPlatformMirror by version mirror.
func (p *providerPlatformMirror) GetProviderPlatformMirrorsByVersion(
	ctx context.Context,
	input *types.GetTerraformProviderPlatformMirrorsByVersionInput,
) ([]types.TerraformProviderPlatformMirror, error) {
	var target struct {
		Node *struct {
			TerraformProviderVersionMirror struct {
				PlatformMirrors []graphQLTerraformProviderPlatformMirror
			} `graphql:"...on TerraformProviderVersionMirror"`
		} `graphql:"node(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.String(input.VersionMirrorID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider version mirror with id %s not found", input.VersionMirrorID)
	}

	converted := make([]types.TerraformProviderPlatformMirror, 0, len(target.Node.TerraformProviderVersionMirror.PlatformMirrors))
	for _, pm := range target.Node.TerraformProviderVersionMirror.PlatformMirrors {
		converted = append(converted, providerPlatformMirrorFromGraphQL(pm))
	}

	return converted, nil
}

// DeleteProviderPlatformMirror deletes a provider platform mirror.
func (p *providerPlatformMirror) DeleteProviderPlatformMirror(
	ctx context.Context,
	input *types.DeleteTerraformProviderPlatformMirrorInput,
) error {
	var wrappedDelete struct {
		DeleteTerraformProviderPlatformMirror struct {
			PlatformMirror graphQLTerraformProviderPlatformMirror
			Problems       []internal.GraphQLProblem
		} `graphql:"deleteTerraformProviderPlatformMirror(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := p.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteTerraformProviderPlatformMirror.Problems)
}

// UploadProviderPlatformPackageToMirror uploads the provider platform package to the mirror.
func (p *providerPlatformMirror) UploadProviderPlatformPackageToMirror(
	ctx context.Context,
	input *types.UploadProviderPlatformPackageToMirrorInput,
) error {
	endpoint, err := url.JoinPath(
		p.client.cfg.Endpoint,
		"v1/provider-mirror/providers",
		input.VersionMirrorID,
		input.OS,
		input.Arch,
		"upload",
	)
	if err != nil {
		return err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, input.Reader)
	if err != nil {
		return err
	}

	authToken, err := p.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return err
	}

	r.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := p.client.httpClient.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.ErrorFromHTTPResponse(resp)
	}

	return nil
}

// GetProviderPlatformPackageDownloadURL returns the download URL and hashes for a provider platform package.
func (p *providerPlatformMirror) GetProviderPlatformPackageDownloadURL(
	ctx context.Context,
	input *types.GetProviderPlatformPackageDownloadURLInput,
) (*types.ProviderPlatformPackageInfo, error) {
	endpoint, err := url.JoinPath(
		p.client.cfg.Endpoint,
		"v1/provider-mirror/providers",
		url.PathEscape(input.GroupPath),
		input.RegistryHostname,
		input.RegistryNamespace,
		input.Type,
		input.Version,
		input.OS,
		input.Arch,
	)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	authToken, err := p.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return nil, err
	}

	r.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := p.client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrorFromHTTPResponse(resp)
	}

	var result types.ProviderPlatformPackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformProviderPlatformMirror struct {
	Metadata      internal.GraphQLMetadata
	ID            graphql.String
	OS            string
	Arch          string
	VersionMirror graphQLTerraformProviderVersionMirror
}

// providerPlatformMirrorFromGraphQL converts a GraphQL TerraformProviderPlatformMirror to an external TerraformProviderPLatformMirror.
func providerPlatformMirrorFromGraphQL(p graphQLTerraformProviderPlatformMirror) types.TerraformProviderPlatformMirror {
	return types.TerraformProviderPlatformMirror{
		Metadata:      internal.MetadataFromGraphQL(p.Metadata, p.ID),
		VersionMirror: providerVersionMirrorFromGraphQL(p.VersionMirror),
		OS:            p.OS,
		Arch:          p.Arch,
	}
}
