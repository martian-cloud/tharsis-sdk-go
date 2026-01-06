package tharsis

import (
	"context"

	"github.com/aws/smithy-go/ptr"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Apply implements functions related to Tharsis Apply.
type Apply interface {
	UpdateApply(ctx context.Context, input *types.UpdateApplyInput) (*types.Apply, error)
}

type apply struct {
	client *Client
}

// NewApply returns an apply.
func NewApply(client *Client) Apply {
	return &apply{client: client}
}

// UpdateApply updates an apply and returns its content.
func (a *apply) UpdateApply(ctx context.Context, input *types.UpdateApplyInput) (*types.Apply, error) {
	var wrappedUpdate struct {
		UpdateApply struct {
			Apply    graphQLApply
			Problems []internal.GraphQLProblem
		} `graphql:"updateApply(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := a.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateApply.Problems); err != nil {
		return nil, err
	}

	updated := applyFromGraphQL(&wrappedUpdate.UpdateApply.Apply)
	return updated, nil
}

// graphQLApply holds information about Tharsis Apply with GraphQL types.
type graphQLApply struct {
	Metadata     internal.GraphQLMetadata
	CurrentJob   *graphQLJob
	ID           graphql.String
	Status       graphql.String
	ErrorMessage *graphql.String
}

// applyFromGraphQL converts a GraphQL Apply to an external Apply.
func applyFromGraphQL(a *graphQLApply) *types.Apply {
	result := &types.Apply{
		Metadata: internal.MetadataFromGraphQL(a.Metadata, a.ID),
		Status:   types.ApplyStatus(a.Status),
	}

	if a.CurrentJob != nil {
		result.CurrentJobID = ptr.String(string(a.CurrentJob.ID))
	}

	if a.ErrorMessage != nil {
		result.ErrorMessage = ptr.String(string(*a.ErrorMessage))
	}

	return result
}
