package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// RunnerSession implements functions related to Tharsis Runner Sessions.
type RunnerSession interface {
	CreateRunnerSession(ctx context.Context, input *types.CreateRunnerSessionInput) (*types.RunnerSession, error)
	SendRunnerSessionHeartbeat(ctx context.Context, input *types.RunnerSessionHeartbeatInput) error
	CreateRunnerSessionError(ctx context.Context, input *types.CreateRunnerSessionErrorInput) error
}

type runnerSession struct {
	client *Client
}

// NewRunnerSession returns a new RunnerSession.
func NewRunnerSession(client *Client) RunnerSession {
	return &runnerSession{client: client}
}

func (s *runnerSession) CreateRunnerSession(ctx context.Context, input *types.CreateRunnerSessionInput) (*types.RunnerSession, error) {
	var wrappedCreate struct {
		CreateRunnerSession struct {
			Problems      []internal.GraphQLProblem
			RunnerSession graphQLRunnerSession
		} `graphql:"createRunnerSession(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := s.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables); err != nil {
		return nil, err
	}

	if err := errors.ErrorFromGraphqlProblems(wrappedCreate.CreateRunnerSession.Problems); err != nil {
		return nil, err
	}

	return runnerSessionFromGraphQL(wrappedCreate.CreateRunnerSession.RunnerSession), nil
}

func (s *runnerSession) SendRunnerSessionHeartbeat(ctx context.Context,
	input *types.RunnerSessionHeartbeatInput) error {
	var wrappedSend struct {
		RunnerSessionHeartbeat struct {
			Problems []internal.GraphQLProblem
		} `graphql:"runnerSessionHeartbeat(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := s.client.graphqlClient.Mutate(ctx, true, &wrappedSend, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedSend.RunnerSessionHeartbeat.Problems)
}

func (s *runnerSession) CreateRunnerSessionError(ctx context.Context,
	input *types.CreateRunnerSessionErrorInput) error {
	var wrappedSend struct {
		CreateRunnerSessionError struct {
			Problems []internal.GraphQLProblem
		} `graphql:"createRunnerSessionError(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := s.client.graphqlClient.Mutate(ctx, true, &wrappedSend, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedSend.CreateRunnerSessionError.Problems)
}

// graphQLRunnerSession is RunnerSession with GraphQL types
type graphQLRunnerSession struct {
	ID            graphql.String           `json:"id"`
	Metadata      internal.GraphQLMetadata `json:"metadata"`
	Runner        graphQLRunnerAgent       `json:"runner"`
	LastContacted graphql.String           `json:"lastContacted"`
	ErrorCount    int                      `json:"errorCount"`
	Internal      bool                     `json:"internal"`
}

// runnerSessionFromGraphQL converts a GraphQL RunnerSession to an external RunnerSession
func runnerSessionFromGraphQL(s graphQLRunnerSession) *types.RunnerSession {
	return &types.RunnerSession{
		Metadata:      internal.MetadataFromGraphQL(s.Metadata, s.ID),
		Runner:        runnerAgentFromGraphQL(s.Runner),
		Internal:      s.Internal,
		LastContacted: timeFromGraphQL(&s.LastContacted),
		ErrorCount:    s.ErrorCount,
	}
}
