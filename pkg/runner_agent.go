package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// RunnerAgent implements functions related to Tharsis Runners
type RunnerAgent interface {
	GetRunnerAgent(ctx context.Context, input *types.GetRunnerInput) (*types.RunnerAgent, error)
	CreateRunnerAgent(ctx context.Context, input *types.CreateRunnerInput) (*types.RunnerAgent, error)
	UpdateRunnerAgent(ctx context.Context, input *types.UpdateRunnerInput) (*types.RunnerAgent, error)
	DeleteRunnerAgent(ctx context.Context, input *types.DeleteRunnerInput) error
	AssignServiceAccountToRunnerAgent(ctx context.Context, input *types.AssignServiceAccountToRunnerInput) error
	UnassignServiceAccountFromRunnerAgent(ctx context.Context, input *types.AssignServiceAccountToRunnerInput) error
}

type runnerAgent struct {
	client *Client
}

// NewRunnerAgent returns a new RunnerAgent
func NewRunnerAgent(client *Client) RunnerAgent {
	return &runnerAgent{client: client}
}

func (r *runnerAgent) GetRunnerAgent(ctx context.Context, input *types.GetRunnerInput) (*types.RunnerAgent, error) {
	// Validate and resolve ID or TRN
	resolvedID, err := types.ValidateIDOrTRN(input.ID, input.TRN, "runner")
	if err != nil {
		return nil, errors.NewError(types.ErrBadRequest, err.Error())
	}

	var target struct {
		Node *struct {
			Runner graphQLRunnerAgent `graphql:"...on Runner"`
		} `graphql:"node(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.String(resolvedID)}

	if err := r.client.graphqlClient.Query(ctx, true, &target, variables); err != nil {
		return nil, err
	}

	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "runner id %s not found", resolvedID)
	}

	return runnerAgentFromGraphQL(target.Node.Runner), nil
}

func (r *runnerAgent) CreateRunnerAgent(ctx context.Context, input *types.CreateRunnerInput) (*types.RunnerAgent, error) {
	var wrappedCreate struct {
		CreateRunner struct {
			Runner   graphQLRunnerAgent
			Problems []internal.GraphQLProblem
		} `graphql:"createRunner(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := r.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables); err != nil {
		return nil, err
	}

	if err := errors.ErrorFromGraphqlProblems(wrappedCreate.CreateRunner.Problems); err != nil {
		return nil, err
	}

	return runnerAgentFromGraphQL(wrappedCreate.CreateRunner.Runner), nil
}

func (r *runnerAgent) UpdateRunnerAgent(ctx context.Context, input *types.UpdateRunnerInput) (*types.RunnerAgent, error) {
	var wrappedUpdate struct {
		UpdateRunner struct {
			Runner   graphQLRunnerAgent
			Problems []internal.GraphQLProblem
		} `graphql:"updateRunner(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := r.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables); err != nil {
		return nil, err
	}

	if err := errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateRunner.Problems); err != nil {
		return nil, err
	}

	return runnerAgentFromGraphQL(wrappedUpdate.UpdateRunner.Runner), nil
}

func (r *runnerAgent) DeleteRunnerAgent(ctx context.Context, input *types.DeleteRunnerInput) error {
	var wrappedDelete struct {
		DeleteRunner struct {
			Runner   graphQLRunnerAgent
			Problems []internal.GraphQLProblem
		} `graphql:"deleteRunner(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := r.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteRunner.Problems)
}

func (r *runnerAgent) AssignServiceAccountToRunnerAgent(ctx context.Context, input *types.AssignServiceAccountToRunnerInput) error {
	var wrappedAssign struct {
		AssignServiceAccountToRunner struct {
			Problems []internal.GraphQLProblem
		} `graphql:"assignServiceAccountToRunner(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := r.client.graphqlClient.Mutate(ctx, true, &wrappedAssign, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedAssign.AssignServiceAccountToRunner.Problems)
}

func (r *runnerAgent) UnassignServiceAccountFromRunnerAgent(ctx context.Context, input *types.AssignServiceAccountToRunnerInput) error {
	var wrappedUnassign struct {
		UnassignServiceAccountFromRunner struct {
			Problems []internal.GraphQLProblem
		} `graphql:"unassignServiceAccountFromRunner(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := r.client.graphqlClient.Mutate(ctx, true, &wrappedUnassign, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedUnassign.UnassignServiceAccountFromRunner.Problems)
}

// graphQLRunnerAgent is RunnerAgent with GraphQL types
type graphQLRunnerAgent struct {
	ID              graphql.String           `json:"id"`
	Metadata        internal.GraphQLMetadata `json:"metadata"`
	Name            string                   `json:"name"`
	Description     string                   `json:"description"`
	GroupPath       string                   `json:"groupPath"`
	ResourcePath    string                   `json:"resourcePath"`
	CreatedBy       string                   `json:"createdBy"`
	Type            graphql.String           `json:"type"`
	Tags            []graphql.String         `json:"tags"`
	RunUntaggedJobs graphql.Boolean          `json:"runUntaggedJobs"`
}

// runnerAgentFromGraphQL converts a GraphQL Runner to an external Runner
func runnerAgentFromGraphQL(r graphQLRunnerAgent) *types.RunnerAgent {
	result := types.RunnerAgent{
		Metadata:        internal.MetadataFromGraphQL(r.Metadata, r.ID),
		Name:            r.Name,
		Description:     r.Description,
		GroupPath:       r.GroupPath,
		ResourcePath:    r.ResourcePath,
		CreatedBy:       r.CreatedBy,
		Type:            types.RunnerType(r.Type),
		RunUntaggedJobs: bool(r.RunUntaggedJobs),
	}

	for _, tag := range r.Tags {
		result.Tags = append(result.Tags, string(tag))
	}

	return &result
}
