package tharsis

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hasura/go-graphql-client"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// StateVersion implements functions related to a Tharsis state version.
type StateVersion interface {
	CreateStateVersion(ctx context.Context, input *types.CreateStateVersionInput) (*types.StateVersion, error)
	DownloadStateVersion(ctx context.Context, input *types.DownloadStateVersionInput, writer io.WriterAt) error
}

type stateVersion struct {
	client *Client
}

// NewStateVersion returns a StateVersion.
func NewStateVersion(client *Client) StateVersion {
	return &stateVersion{client: client}
}

// CreateStateVersion creates a State Version and returns its contents.
func (s *stateVersion) CreateStateVersion(ctx context.Context,
	input *types.CreateStateVersionInput) (*types.StateVersion, error) {

	var wrappedCreate struct {
		CreateStateVersion struct {
			StateVersion GraphQLStateVersion
			Problems     []internal.GraphQLProblem
		} `graphql:"createStateVersion(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := s.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedCreate.CreateStateVersion.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems creating state version: %v", err)
	}

	created, err := stateVersionFromGraphQL(&wrappedCreate.CreateStateVersion.StateVersion)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// DownloadStateVersion downloads a state version and returns the response.
func (s *stateVersion) DownloadStateVersion(ctx context.Context,
	input *types.DownloadStateVersionInput, writer io.WriterAt) error {

	// Create the URL and request.
	url := strings.Join([]string{s.client.cfg.Endpoint, "v1", "state-versions", input.ID, "content"}, "/")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Get the authentication token.
	authToken, err := s.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Accept", "application/json")

	// Make the request.
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download state version response code: %d", resp.StatusCode)
	}

	return copyFromResponseBody(resp, writer)
}

// GraphQLStateVersion represents the insides of the query structure,
// everything in the state version object, and with graphql types.
type GraphQLStateVersion struct {
	Metadata internal.GraphQLMetadata
	ID       graphql.String
	Run      struct {
		ID graphql.String
	}
	Outputs []GraphQLStateVersionOutput
}

// GraphQLStateVersionOutput represents the insides of the query structure,
// everything in the state version output object, and with graphql types.
type GraphQLStateVersionOutput struct {
	Metadata  internal.GraphQLMetadata
	ID        graphql.String
	Name      graphql.String
	Value     graphql.String
	Type      graphql.String
	Sensitive graphql.Boolean
}

// stateVersionFromGraphQL converts a GraphQL State Version to a model type.
func stateVersionFromGraphQL(input *GraphQLStateVersion) (*types.StateVersion, error) {
	if input == nil {
		return nil, nil
	}

	outputs, err := sliceStateVersionOutputsFromGraphQL(input.Outputs)
	if err != nil {
		return nil, err
	}

	return &types.StateVersion{
		Metadata: internal.MetadataFromGraphQL(input.Metadata, input.ID),
		RunID:    string(input.Run.ID),
		Outputs:  outputs,
	}, nil
}

// sliceStateVersionOutputsFromGraphQL converts a slice of GraphQL State Version Outputs
// to a slice of state version outputs model type.
func sliceStateVersionOutputsFromGraphQL(inputs []GraphQLStateVersionOutput) ([]types.StateVersionOutput, error) {
	result := make([]types.StateVersionOutput, len(inputs))

	for i, input := range inputs {
		val, err := stateVersionOutputFromGraphQL(input)
		if err != nil {
			return result, err
		}
		result[i] = *val
	}

	return result, nil
}

// stateVersionOutputFromGraphQL converts a GraphQL State Version Output to a
// state version output model type.
func stateVersionOutputFromGraphQL(g GraphQLStateVersionOutput) (*types.StateVersionOutput, error) {
	ty, err := ctyjson.UnmarshalType([]byte(g.Type))
	if err != nil {
		return nil, err
	}

	val, err := ctyjson.Unmarshal([]byte(g.Value), ty)
	if err != nil {
		return nil, err
	}

	return &types.StateVersionOutput{
		Metadata:  internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Name:      string(g.Name),
		Value:     val,
		Type:      ty,
		Sensitive: bool(g.Sensitive),
	}, nil
}
