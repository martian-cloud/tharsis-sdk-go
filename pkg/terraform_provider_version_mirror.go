package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/paginators"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformProviderVersionMirror implements functionalities related to Terraform provider version mirroring.
type TerraformProviderVersionMirror interface {
	GetProviderVersionMirror(
		ctx context.Context,
		input *types.GetTerraformProviderVersionMirrorInput,
	) (*types.TerraformProviderVersionMirror, error)
	GetProviderVersionMirrorByAddress(
		ctx context.Context,
		input *types.GetTerraformProviderVersionMirrorByAddressInput,
	) (*types.TerraformProviderVersionMirror, error)
	GetProviderVersionMirrors(
		ctx context.Context,
		input *types.GetTerraformProviderVersionMirrorsInput,
	) (*types.GetTerraformProviderVersionMirrorsOutput, error)
	GetProviderVersionMirrorPaginator(ctx context.Context,
		input *types.GetTerraformProviderVersionMirrorsInput,
	) (*GetTerraformProviderVersionMirrorsPaginator, error)
	GetAvailableProviderVersions(
		ctx context.Context,
		input *types.GetAvailableProviderVersionsInput,
	) (map[string]struct{}, error)
	CreateProviderVersionMirror(
		ctx context.Context,
		input *types.CreateTerraformProviderVersionMirrorInput,
	) (*types.TerraformProviderVersionMirror, error)
	DeleteProviderVersionMirror(
		ctx context.Context,
		input *types.DeleteTerraformProviderVersionMirrorInput,
	) error
}

type providerVersionMirror struct {
	client *Client
}

// NewTerraformProviderVersionMirror returns a TerraformProviderVersionMirror.
func NewTerraformProviderVersionMirror(client *Client) TerraformProviderVersionMirror {
	return &providerVersionMirror{client: client}
}

// GetTerraformProviderVersionMirror returns a provider version mirror.
func (p *providerVersionMirror) GetProviderVersionMirror(
	ctx context.Context,
	input *types.GetTerraformProviderVersionMirrorInput,
) (*types.TerraformProviderVersionMirror, error) {
	var target struct {
		Node *struct {
			TerraformProviderVersionMirror graphQLTerraformProviderVersionMirror `graphql:"...on TerraformProviderVersionMirror"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider version mirror with id %s not found", input.ID)
	}

	result := providerVersionMirrorFromGraphQL(target.Node.TerraformProviderVersionMirror)
	return &result, nil
}

// GetProviderVersionMirrorByAddress returns a provider version mirror by address.
func (p *providerVersionMirror) GetProviderVersionMirrorByAddress(
	ctx context.Context,
	input *types.GetTerraformProviderVersionMirrorByAddressInput,
) (*types.TerraformProviderVersionMirror, error) {
	var target struct {
		TerraformProviderVersionMirror *graphQLTerraformProviderVersionMirror `graphql:"terraformProviderVersionMirror(registryNamespace: $registryNamespace, registryHostname: $registryHostname, type: $type, version: $version, groupPath: $groupPath)"`
	}

	variables := map[string]interface{}{
		"registryNamespace": graphql.String(input.RegistryNamespace),
		"registryHostname":  graphql.String(input.RegistryHostname),
		"type":              graphql.String(input.Type),
		"version":           graphql.String(input.Version),
		"groupPath":         graphql.String(input.GroupPath),
	}

	err := p.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.TerraformProviderVersionMirror == nil {
		return nil, errors.NewError(types.ErrNotFound, "terraform provider version mirror with address not found")
	}

	result := providerVersionMirrorFromGraphQL(*target.TerraformProviderVersionMirror)
	return &result, nil
}

// GetProviderVersionMirrors returns a paginated list of TerraformProviderVersionMirrors.
func (p *providerVersionMirror) GetProviderVersionMirrors(
	ctx context.Context,
	input *types.GetTerraformProviderVersionMirrorsInput,
) (*types.GetTerraformProviderVersionMirrorsOutput, error) {

	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := getTerraformProviderVersionMirrors(ctx, p.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	if queryStruct.Group == nil {
		return nil, errors.NewError(types.ErrNotFound, "group with path %s not found", input.GroupPath)
	}

	// Convert and repackage the type-specific results.
	versionMirrorResults := make([]types.TerraformProviderVersionMirror, len(queryStruct.Group.TerraformProviderMirrors.Edges))
	for ix, edge := range queryStruct.Group.TerraformProviderMirrors.Edges {
		versionMirrorResults[ix] = providerVersionMirrorFromGraphQL(edge.Node)
	}

	return &types.GetTerraformProviderVersionMirrorsOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.Group.TerraformProviderMirrors.TotalCount),
			HasNextPage: bool(queryStruct.Group.TerraformProviderMirrors.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.Group.TerraformProviderMirrors.PageInfo.EndCursor),
		},
		VersionMirrors: versionMirrorResults,
	}, nil
}

// GetProviderVersionMirrorPaginator returns the provider mirror paginator.
func (p *providerVersionMirror) GetProviderVersionMirrorPaginator(_ context.Context,
	input *types.GetTerraformProviderVersionMirrorsInput) (*GetTerraformProviderVersionMirrorsPaginator, error) {

	paginator := newProviderVersionMirrorPaginator(*p.client, input)
	return &paginator, nil
}

// GetAvailableProviderVersions returns all cached versions for a provider via REST API.
func (p *providerVersionMirror) GetAvailableProviderVersions(
	ctx context.Context,
	input *types.GetAvailableProviderVersionsInput,
) (map[string]struct{}, error) {
	endpoint, err := url.JoinPath(
		p.client.cfg.Endpoint,
		"v1/provider-mirror/providers",
		input.GroupPath,
		input.RegistryHostname,
		input.RegistryNamespace,
		input.Type,
		"index.json",
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	authToken, err := p.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrorFromHTTPResponse(resp)
	}

	var result struct {
		Versions map[string]struct{} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Versions, nil
}

// CreateProviderVersionMirror creates a provider version mirror.
func (p *providerVersionMirror) CreateProviderVersionMirror(
	ctx context.Context,
	input *types.CreateTerraformProviderVersionMirrorInput,
) (*types.TerraformProviderVersionMirror, error) {
	var wrappedCreate struct {
		CreateTerraformProviderVersionMirror struct {
			VersionMirror graphQLTerraformProviderVersionMirror
			Problems      []internal.GraphQLProblem
		} `graphql:"createTerraformProviderVersionMirror(input: $input)"`
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

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTerraformProviderVersionMirror.Problems); err != nil {
		return nil, err
	}

	created := providerVersionMirrorFromGraphQL(wrappedCreate.CreateTerraformProviderVersionMirror.VersionMirror)
	return &created, nil
}

// DeleteProviderVersionMirror deletes a provider version mirror.
func (p *providerVersionMirror) DeleteProviderVersionMirror(
	ctx context.Context,
	input *types.DeleteTerraformProviderVersionMirrorInput,
) error {
	var wrappedDelete struct {
		DeleteTerraformProviderVersionMirror struct {
			VersionMirror graphQLTerraformProviderVersionMirror
			Problems      []internal.GraphQLProblem
		} `graphql:"deleteTerraformProviderVersionMirror(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := p.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteTerraformProviderVersionMirror.Problems)
}

// The GetTerraformProviderVersionMirrorsPaginator:

// GetTerraformProviderVersionMirrorsPaginator is a type-specific paginator.
type GetTerraformProviderVersionMirrorsPaginator struct {
	generic paginators.Paginator
}

// newProviderVersionMirrorPaginator returns a new TerraformProviderVersionMirror paginator.
func newProviderVersionMirrorPaginator(
	client Client,
	input *types.GetTerraformProviderVersionMirrorsInput,
) GetTerraformProviderVersionMirrorsPaginator {
	inputCopy := &types.GetTerraformProviderVersionMirrorsInput{
		Sort:              input.Sort,
		PaginationOptions: input.PaginationOptions,
		IncludeInherited:  input.IncludeInherited,
		GroupPath:         input.GroupPath,
	}

	queryCallback := func(ctx context.Context, after *string) (interface{}, error) {
		inputCopy.PaginationOptions.Cursor = after
		return client.TerraformProviderVersionMirror.GetProviderVersionMirrors(ctx, inputCopy)
	}

	genericPaginator := paginators.NewPaginator(queryCallback)

	return GetTerraformProviderVersionMirrorsPaginator{
		generic: genericPaginator,
	}
}

// HasMore returns a boolean, whether there is another page (or more):
func (p *GetTerraformProviderVersionMirrorsPaginator) HasMore() bool {
	return p.generic.HasMore()
}

// Next returns the next page of results:
func (p *GetTerraformProviderVersionMirrorsPaginator) Next(ctx context.Context) (*types.GetTerraformProviderVersionMirrorsOutput, error) {

	// The generic paginator runs the query.
	untyped, err := p.generic.Next(ctx)
	if err != nil {
		return nil, err
	}

	// We know the returned data is a *GetTerraformProviderVersionMirrorsOutput:
	return untyped.(*types.GetTerraformProviderVersionMirrorsOutput), nil
}

func getTerraformProviderVersionMirrors(
	ctx context.Context,
	client graphqlClient,
	input *types.GetTerraformProviderVersionMirrorsInput,
	after *string,
) (*getTerraformProviderVersionMirrorsQuery, error) {
	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getTerraformProviderVersionMirrorsQuery{}

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

	if input.IncludeInherited != nil {
		variables["includeInherited"] = *input.IncludeInherited
	} else {
		variables["includeInherited"] = (*graphql.Boolean)(nil)
	}

	variables["fullPath"] = input.GroupPath

	type TerraformProviderVersionMirrorSort string
	if input.Sort != nil {
		variables["sort"] = TerraformProviderVersionMirrorSort(*input.Sort)
	} else {
		variables["sort"] = (*TerraformProviderVersionMirrorSort)(nil)
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

// getTerraformProviderVersionMirrorsQuery is the query structure for GetProviderVersionMirrors.
type getTerraformProviderVersionMirrorsQuery struct {
	Group *struct {
		TerraformProviderMirrors struct {
			PageInfo struct {
				EndCursor   graphql.String
				HasNextPage graphql.Boolean
			}
			Edges []struct {
				Node graphQLTerraformProviderVersionMirror
			}
			TotalCount graphql.Int
		} `graphql:"terraformProviderMirrors(first: $first, after: $after, sort: $sort, includeInherited: $includeInherited)"`
	} `graphql:"group(fullPath: $fullPath)"`
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

type graphQLTerraformProviderVersionMirror struct {
	ID                graphql.String
	Metadata          internal.GraphQLMetadata
	Version           string
	RegistryNamespace string
	RegistryHostname  string
	Type              string
}

// providerVersionMirrorFromGraphQL converts a GraphQL TerraformProviderVersionMirror to an external TerraformProviderVersionMirror.
func providerVersionMirrorFromGraphQL(p graphQLTerraformProviderVersionMirror) types.TerraformProviderVersionMirror {
	return types.TerraformProviderVersionMirror{
		Metadata:          internal.MetadataFromGraphQL(p.Metadata, p.ID),
		SemanticVersion:   p.Version,
		RegistryHostname:  p.RegistryHostname,
		RegistryNamespace: p.RegistryNamespace,
		Type:              p.Type,
	}
}
