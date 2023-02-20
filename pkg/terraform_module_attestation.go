package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformModuleAttestation implements functions related to Tharsis module attestations.
type TerraformModuleAttestation interface {
	GetModuleAttestations(ctx context.Context, input *types.GetTerraformModuleAttestationsInput) (*types.GetTerraformModuleAttestationsOutput, error)
	CreateModuleAttestation(ctx context.Context, input *types.CreateTerraformModuleAttestationInput) (*types.TerraformModuleAttestation, error)
	UpdateModuleAttestation(ctx context.Context, input *types.UpdateTerraformModuleAttestationInput) (*types.TerraformModuleAttestation, error)
	DeleteModuleAttestation(ctx context.Context, input *types.DeleteTerraformModuleAttestationInput) error
}

type moduleAttestation struct {
	client *Client
}

// NewTerraformModuleAttestation returns a TerraformModuleAttestation.
func NewTerraformModuleAttestation(client *Client) TerraformModuleAttestation {
	return &moduleAttestation{client: client}
}

func (ma *moduleAttestation) GetModuleAttestations(ctx context.Context, input *types.GetTerraformModuleAttestationsInput) (*types.GetTerraformModuleAttestationsOutput, error) {
	if input.Filter == nil {
		return nil, errors.NewError(types.ErrBadRequest, "Filter must be non-nil when calling GetModuleAttestations")
	}

	switch {
	case input.Filter.TerraformModuleID != nil:
		return ma.getTerraformModuleAttestationsForModule(ctx, ma.client.graphqlClient, input, nil)
	case input.Filter.TerraformModuleVersionID != nil:
		return ma.getTerraformModuleAttestationsForModuleVersion(ctx, ma.client.graphqlClient, input, nil)
	default:
		return nil, errors.NewError(types.ErrBadRequest, "Either TerraformModuleID or TerraformModuleVersionID must be specified in filter input")
	}
}

func (ma *moduleAttestation) CreateModuleAttestation(ctx context.Context, input *types.CreateTerraformModuleAttestationInput) (*types.TerraformModuleAttestation, error) {
	var wrappedCreate struct {
		CreateTerraformModuleAttestation struct {
			ModuleAttestation graphQLTerraformModuleAttestation
			Problems          []internal.GraphQLProblem
		} `graphql:"createTerraformModuleAttestation(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ma.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTerraformModuleAttestation.Problems); err != nil {
		return nil, err
	}

	created := moduleAttestationFromGraphQL(wrappedCreate.CreateTerraformModuleAttestation.ModuleAttestation)
	return &created, nil
}

func (ma *moduleAttestation) UpdateModuleAttestation(ctx context.Context, input *types.UpdateTerraformModuleAttestationInput) (*types.TerraformModuleAttestation, error) {
	var wrappedUpdate struct {
		UpdateTerraformModuleAttestation struct {
			ModuleAttestation graphQLTerraformModuleAttestation
			Problems          []internal.GraphQLProblem
		} `graphql:"updateTerraformModuleAttestation(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ma.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateTerraformModuleAttestation.Problems); err != nil {
		return nil, err
	}

	attestation := moduleAttestationFromGraphQL(wrappedUpdate.UpdateTerraformModuleAttestation.ModuleAttestation)
	return &attestation, nil
}

func (ma *moduleAttestation) DeleteModuleAttestation(ctx context.Context, input *types.DeleteTerraformModuleAttestationInput) error {
	var wrappedDelete struct {
		DeleteTerraformModuleAttestation struct {
			ModuleAttestation graphQLTerraformModuleAttestation
			Problems          []internal.GraphQLProblem
		} `graphql:"deleteTerraformModuleAttestation(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := ma.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteTerraformModuleAttestation.Problems)
}

// getTerraformModuleAttestationsForModule runs the query and returns the results.
func (ma *moduleAttestation) getTerraformModuleAttestationsForModule(ctx context.Context, client graphqlClient,
	input *types.GetTerraformModuleAttestationsInput, after *string) (*types.GetTerraformModuleAttestationsOutput, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getTerraformModuleAttestationsForModuleQuery{}

	// Build the variables for filtering, sorting, and pagination.
	variables := map[string]interface{}{}

	variables["id"] = graphql.String(*input.Filter.TerraformModuleID)

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

	// Resource type specific settings:

	// Make sure to pass the expected types for these variables.
	var digest *graphql.String
	if input.Filter.Digest != nil {
		digestString := graphql.String(*input.Filter.Digest)
		digest = &digestString
	} else {
		digest = nil
	}
	variables["digest"] = digest

	type TerraformModuleAttestationSort string
	variables["sort"] = TerraformModuleAttestationSort(*input.Sort)

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	if queryStructP.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "module with id %s not found", *input.Filter.TerraformModuleID)
	}

	// Convert and repackage the type-specific results.
	attestationResults := make([]types.TerraformModuleAttestation, len(queryStructP.Node.TerraformModule.Attestations.Edges))
	for ix, attestationCustom := range queryStructP.Node.TerraformModule.Attestations.Edges {
		attestationResults[ix] = moduleAttestationFromGraphQL(attestationCustom.Node)
	}

	return &types.GetTerraformModuleAttestationsOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStructP.Node.TerraformModule.Attestations.TotalCount),
			HasNextPage: bool(queryStructP.Node.TerraformModule.Attestations.PageInfo.HasNextPage),
			Cursor:      string(queryStructP.Node.TerraformModule.Attestations.PageInfo.EndCursor),
		},
		ModuleAttestations: attestationResults,
	}, nil
}

// getTerraformModuleAttestationsForModuleVersion runs the query and returns the results.
func (ma *moduleAttestation) getTerraformModuleAttestationsForModuleVersion(ctx context.Context, client graphqlClient,
	input *types.GetTerraformModuleAttestationsInput, after *string) (*types.GetTerraformModuleAttestationsOutput, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getTerraformModuleAttestationsForVersionQuery{}

	// Build the variables for filtering, sorting, and pagination.
	variables := map[string]interface{}{}

	variables["id"] = graphql.String(*input.Filter.TerraformModuleVersionID)

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

	type TerraformModuleAttestationSort string
	variables["sort"] = TerraformModuleAttestationSort(*input.Sort)

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}

	if queryStructP.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "module with id %s not found", *input.Filter.TerraformModuleID)
	}

	// Convert and repackage the type-specific results.
	attestationResults := make([]types.TerraformModuleAttestation, len(queryStructP.Node.TerraformModuleVersion.Attestations.Edges))
	for ix, attestationCustom := range queryStructP.Node.TerraformModuleVersion.Attestations.Edges {
		attestationResults[ix] = moduleAttestationFromGraphQL(attestationCustom.Node)
	}

	return &types.GetTerraformModuleAttestationsOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStructP.Node.TerraformModuleVersion.Attestations.TotalCount),
			HasNextPage: bool(queryStructP.Node.TerraformModuleVersion.Attestations.PageInfo.HasNextPage),
			Cursor:      string(queryStructP.Node.TerraformModuleVersion.Attestations.PageInfo.EndCursor),
		},
		ModuleAttestations: attestationResults,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getTerraformModuleAttestationsForModuleQuery is the query structure for GetTerraformModuleAttestations.
// It contains the tag with the include-everything argument list.
type getTerraformModuleAttestationsForModuleQuery struct {
	Node *struct {
		TerraformModule struct {
			Attestations struct {
				PageInfo struct {
					EndCursor   graphql.String
					HasNextPage graphql.Boolean
				}
				Edges []struct {
					Node graphQLTerraformModuleAttestation
				}
				TotalCount graphql.Int
			} `graphql:"attestations(first: $first, after: $after, digest: $digest, sort: $sort)"`
		} `graphql:"...on TerraformModule"`
	} `graphql:"node(id: $id)"`
}

// getTerraformModuleAttestationsForVersionQuery is the query structure for GetTerraformModuleAttestations.
// It contains the tag with the include-everything argument list.
type getTerraformModuleAttestationsForVersionQuery struct {
	Node *struct {
		TerraformModuleVersion struct {
			Attestations struct {
				PageInfo struct {
					EndCursor   graphql.String
					HasNextPage graphql.Boolean
				}
				Edges []struct {
					Node graphQLTerraformModuleAttestation
				}
				TotalCount graphql.Int
			} `graphql:"attestations(first: $first, after: $after, sort: $sort)"`
		} `graphql:"...on TerraformModuleVersion"`
	} `graphql:"node(id: $id)"`
}

// Related types and conversion functions:
type graphQLTerraformModuleAttestation struct {
	Metadata      internal.GraphQLMetadata
	ID            graphql.String
	Module        graphQLTerraformModule
	Description   string
	SchemaType    string
	PredicateType string
	Data          string
	Digests       []string
}

// moduleAttestationFromGraphQL converts a GraphQL TerraformModuleAttestation to an external TerraformModuleAttestation.
func moduleAttestationFromGraphQL(tma graphQLTerraformModuleAttestation) types.TerraformModuleAttestation {
	result := types.TerraformModuleAttestation{
		Metadata:      internal.MetadataFromGraphQL(tma.Metadata, tma.ID),
		ModuleID:      string(tma.Module.ID),
		Description:   tma.Description,
		SchemaType:    tma.SchemaType,
		PredicateType: tma.PredicateType,
		Data:          tma.Data,
		Digests:       tma.Digests,
	}
	return result
}
