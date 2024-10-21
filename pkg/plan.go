package tharsis

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// planWithProviderSchemasPayload is a struct that contains the plan and provider schemas for uploading to the API
type planWithProviderSchemasPayload struct {
	Plan            *tfjson.Plan            `json:"plan"`
	ProviderSchemas *tfjson.ProviderSchemas `json:"provider_schemas"`
}

// Plan implements functions related to Tharsis Plan.
type Plan interface {
	UpdatePlan(ctx context.Context, input *types.UpdatePlanInput) (*types.Plan, error)
	DownloadPlanCache(ctx context.Context, id string, writer io.WriterAt) error
	UploadPlanCache(ctx context.Context, id string, body io.Reader) error
	UploadPlanData(ctx context.Context, id string, tfPlan *tfjson.Plan, tfProviderSchemas *tfjson.ProviderSchemas) error
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

// UploadPlanData uploads plan data such as the plan json and provider schemas
func (p *plan) UploadPlanData(ctx context.Context, id string, tfPlan *tfjson.Plan, tfProviderSchemas *tfjson.ProviderSchemas) error {
	planWithSchemas := &planWithProviderSchemasPayload{
		Plan:            tfPlan,
		ProviderSchemas: tfProviderSchemas,
	}

	planData, err := json.Marshal(planWithSchemas)
	if err != nil {
		return fmt.Errorf("failed to marshal plan json: %w", err)
	}

	// Compress the data using gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err = writer.Write(planData); err != nil {
		return fmt.Errorf("failed to compress plan data: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Create the PUT request
	url := strings.Join([]string{p.client.cfg.Endpoint, "v1", "plans", id, "content.json"}, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create http request for uploading plan data: %w", err)
	}

	// Set the Content-Encoding header to gzip
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/octet-stream")

	_, err = p.doRequest(req)
	if err != nil {
		return fmt.Errorf("failed to upload plan data: %w", err)
	}

	return nil
}

// do prepares, makes a request with appropriate headers and returns the response.
func (p *plan) doRequest(req *http.Request) (*http.Response, error) {
	// Get the authentication token.
	authToken, err := p.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Accept", "application/json")

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

func (p *plan) do(ctx context.Context,
	method string, url string, body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Set appropriate request headers.
	if req.Method == http.MethodPut {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	return p.doRequest(req)
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
