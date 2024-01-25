package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestCreateRunnerSession(t *testing.T) {
	now := time.Now().UTC()                                 // Getting rid of local timezone makes equality checks work better.
	then := now.Truncate(time.Second).Add(-5 * time.Second) // must drop fractional seconds due to formatting
	thenGraphQLString := graphql.String(then.Format(time.RFC3339))

	runnerPath := "runner-path"
	runnerSessionID := "runner-session-id"
	runnerSessionVersion := "runner-session-version"

	type graphqlCreateRunnerSessionMutation struct {
		RunnerSession graphQLRunnerSession         `json:"runnerSession"`
		Problems      []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateRunnerSessionPayload struct {
		CreateRunnerSession graphqlCreateRunnerSessionMutation `json:"createRunnerSession"`
	}

	// test cases
	type testCase struct {
		responsePayload     interface{}
		input               *types.CreateRunnerSessionInput
		expectRunnerSession *types.RunnerSession
		name                string
		expectErrorCode     types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully create a runner session",
			input: &types.CreateRunnerSessionInput{
				RunnerPath: runnerPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateRunnerSessionPayload{
					CreateRunnerSession: graphqlCreateRunnerSessionMutation{
						RunnerSession: graphQLRunnerSession{
							ID: graphql.String(runnerSessionID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(runnerSessionVersion),
							},
							Runner:        graphQLRunnerAgent{},
							Internal:      false,
							LastContacted: thenGraphQLString,
							ErrorCount:    0,
						},
					},
				},
			},
			expectRunnerSession: &types.RunnerSession{
				Metadata: types.ResourceMetadata{
					ID:                   runnerSessionID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              runnerSessionVersion,
				},
				Runner:        &types.RunnerAgent{},
				LastContacted: &then,
				ErrorCount:    0,
				Internal:      false,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.CreateRunnerSessionInput{
				RunnerPath: "invalid-runner-path",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateRunnerSessionPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: types.ErrInternal,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			payload, err := json.Marshal(&test.responsePayload)
			if err != nil {
				t.Fatal(err)
			}

			// Prepare to replace the http.transport that is used by the http client that is passed to the GraphQL client.
			client := &Client{
				graphqlClient: newGraphQLClientForTest(testClientInput{
					statusToReturn:  http.StatusOK,
					payloadToReturn: string(payload),
				}),
			}
			client.RunnerSession = NewRunnerSession(client)

			// Call the method being tested.
			actualRunnerSession, actualError := client.RunnerSession.CreateRunnerSession(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkRunnerSession(t, test.expectRunnerSession, actualRunnerSession)
		})
	}
}

func TestSendRunnerSessionHeartbeat(t *testing.T) {

	runnerSessionID := "runner-session-id"

	// SendRunnerSessionHeartbeat returns only a potential error.
	type graphqlSendRunnerSessionHeartbeatMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.RunnerSessionHeartbeatInput
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully send",
			input: &types.RunnerSessionHeartbeatInput{
				RunnerSessionID: runnerSessionID,
			},
			responsePayload: fakeGraphqlResponsePayload{},
		},
		{
			name: "verify that correct error is returned",
			input: &types.RunnerSessionHeartbeatInput{
				RunnerSessionID: runnerSessionID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: types.ErrInternal,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			payload, err := json.Marshal(&test.responsePayload)
			if err != nil {
				t.Fatal(err)
			}

			// Prepare to replace the http.transport that is used by the http client that is passed to the GraphQL client.
			client := &Client{
				graphqlClient: newGraphQLClientForTest(testClientInput{
					statusToReturn:  http.StatusOK,
					payloadToReturn: string(payload),
				}),
			}
			client.RunnerSession = NewRunnerSession(client)

			// Call the method being tested.
			actualError := client.RunnerSession.SendRunnerSessionHeartbeat(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}
}

func TestCreateRunnerSessionError(t *testing.T) {

	runnerSessionID := "runner-session-id"
	notFoundMessage := "not found"

	// CreateRunnerSessionError returns only a potential error.
	type graphqlCreateRunnerSessionErrorMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.CreateRunnerSessionErrorInput
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully send a not-found error",
			input: &types.CreateRunnerSessionErrorInput{
				RunnerSessionID: runnerSessionID,
				ErrorMessage:    notFoundMessage,
			},
			responsePayload: fakeGraphqlResponsePayload{},
		},
		{
			name: "failed to send an error, returned an internal error",
			input: &types.CreateRunnerSessionErrorInput{
				RunnerSessionID: runnerSessionID,
				ErrorMessage:    notFoundMessage,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: types.ErrInternal,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			payload, err := json.Marshal(&test.responsePayload)
			if err != nil {
				t.Fatal(err)
			}

			// Prepare to replace the http.transport that is used by the http client that is passed to the GraphQL client.
			client := &Client{
				graphqlClient: newGraphQLClientForTest(testClientInput{
					statusToReturn:  http.StatusOK,
					payloadToReturn: string(payload),
				}),
			}
			client.RunnerSession = NewRunnerSession(client)

			// Call the method being tested.
			actualError := client.RunnerSession.CreateRunnerSessionError(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}
}

// Utility functions:

func checkRunnerSession(t *testing.T, expectRunnerSession, actualRunnerSession *types.RunnerSession) {
	if expectRunnerSession != nil {
		require.NotNil(t, actualRunnerSession)
		assert.Equal(t, expectRunnerSession, actualRunnerSession)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.RunnerSession)(nil)
		assert.Equal(t, (*types.RunnerSession)(nil), actualRunnerSession)
	}
}
