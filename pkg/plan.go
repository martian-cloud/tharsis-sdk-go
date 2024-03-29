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

// Plan implements functions related to Tharsis Plan.
type Plan interface {
	UpdatePlan(ctx context.Context, input *types.UpdatePlanInput) (*types.Plan, error)
	DownloadPlanCache(ctx context.Context, id string, writer io.WriterAt) error
	UploadPlanCache(ctx context.Context, id string, body io.Reader) error
}

type plan struct {
	client *Client
}

// NewPlan returns a plan.
func NewPlan(client *Client) Plan {
	return &plan{client: client}
}

// UpdatePlan updates a plan and returns its content.
func (p *plan) UpdatePlan(ctx context.Context, input *types.UpdatePlanInput) (*types.Plan, error) {
	var wrappedUpdate struct {
		UpdatePlan struct {
			Problems []internal.GraphQLProblem
			Plan     graphQLPlan
		} `graphql:"updatePlan(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := p.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdatePlan.Problems); err != nil {
		return nil, err
	}

	updated := planFromGraphQL(wrappedUpdate.UpdatePlan.Plan)
	return updated, nil
}

// DownloadPlanCache downloads a plan cache and returns the response.
func (p *plan) DownloadPlanCache(ctx context.Context, id string, writer io.WriterAt) error {
	tfeV2Endpoint, err := p.client.services.GetServiceURL("tfe.v2")
	if err != nil {
		return fmt.Errorf("failed to discover tfe.v2 endpoint: %w", err)
	}

	// Create the URL and request.
	url := tfeV2Endpoint.String() + strings.Join([]string{"plans", id, "content"}, "/")
	resp, err := p.do(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	return copyFromResponseBody(resp, writer)
}

// UploadPlanCache uploads a plan cache and returns any errors.
func (p *plan) UploadPlanCache(ctx context.Context, id string, body io.Reader) error {
	// Not a TFE endpoint
	url := strings.Join([]string{p.client.cfg.Endpoint, "v1", "plans", id, "content"}, "/")
	_, err := p.do(ctx, http.MethodPut, url, body)
	if err != nil {
		return err
	}

	return nil
}

// do prepares, makes a request with appropriate headers and returns the response.
func (p *plan) do(ctx context.Context,
	method string, url string, body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Get the authentication token.
	authToken, err := p.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Accept", "application/json")

	// Set appropriate request headers.
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	// Make the request.
	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrorFromHTTPResponse(resp)
	}

	return resp, nil
}

// graphQLPlan represents a Tharsis plan with GraphQL types.
type graphQLPlan struct {
	Metadata             internal.GraphQLMetadata
	ID                   graphql.String
	Status               graphql.String
	CurrentJob           graphQLJob
	ResourceAdditions    graphql.Int
	ResourceChanges      graphql.Int
	ResourceDestructions graphql.Int
	HasChanges           graphql.Boolean
}

// planFromGraphQL converts a GraphQL Plan to an external Plan.
func planFromGraphQL(p graphQLPlan) *types.Plan {
	jobID := string(p.CurrentJob.ID) // need to convert to *string
	result := &types.Plan{
		Metadata:             internal.MetadataFromGraphQL(p.Metadata, p.ID),
		Status:               types.PlanStatus(p.Status),
		HasChanges:           bool(p.HasChanges),
		ResourceAdditions:    int(p.ResourceAdditions),
		ResourceChanges:      int(p.ResourceChanges),
		ResourceDestructions: int(p.ResourceDestructions),
		CurrentJobID:         &jobID,
	}
	return result
}
