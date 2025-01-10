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

func TestGetRunnerAgent(t *testing.T) {
	now := time.Now().UTC()

	runnerAgentID := "runner-agent-id-1"
	runnerAgentVersion := "runner-agent-version-1"

	type graphqlRunnerAgentPayload struct {
		Node *graphQLRunnerAgent `json:"node"`
	}

	type testCase struct {
		responsePayload   interface{}
		expectRunnerAgent *types.RunnerAgent
		name              string
		expectErrorCode   types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return runner agent by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlRunnerAgentPayload{
					Node: &graphQLRunnerAgent{
						ID: graphql.String(runnerAgentID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(runnerAgentVersion),
						},
						Type:            "t01",
						GroupPath:       "gp01",
						ResourcePath:    "rp01",
						Name:            "nm01",
						Description:     "de01",
						CreatedBy:       "cr01",
						Tags:            []graphql.String{"tag1", "tag2"},
						RunUntaggedJobs: graphql.Boolean(true),
					},
				},
			},
			expectRunnerAgent: &types.RunnerAgent{
				Metadata: types.ResourceMetadata{
					ID:                   runnerAgentID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              runnerAgentVersion,
				},
				Type:            "t01",
				GroupPath:       "gp01",
				ResourcePath:    "rp01",
				Name:            "nm01",
				Description:     "de01",
				CreatedBy:       "cr01",
				Tags:            []string{"tag1", "tag2"},
				RunUntaggedJobs: true,
			},
		},

		// query returns error as if the ID is invalid
		{
			name: "query returns error as if the ID is invalid",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlRunnerAgentPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02)",
						Path: []string{
							"runner",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: types.ErrInternal,
		},

		{
			name: "query returns nil runner agent, as if the specified runner agent does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlRunnerAgentPayload{},
			},
			expectErrorCode: types.ErrNotFound,
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
			client.RunnerAgent = NewRunnerAgent(client)

			// Call the method being tested.
			actualRunnerAgent, actualError := client.RunnerAgent.GetRunnerAgent(
				ctx,
				&types.GetRunnerInput{ID: runnerAgentID},
			)

			checkError(t, test.expectErrorCode, actualError)

			if test.expectRunnerAgent != nil {
				require.NotNil(t, actualRunnerAgent)
				assert.Equal(t, test.expectRunnerAgent, actualRunnerAgent)
			}
		})
	}
}

func TestCreateRunnerAgent(t *testing.T) {
	now := time.Now().UTC()

	runnerAgentID := "runner-agent-1"

	type graphqlCreateRunnerAgentMutation struct {
		RunnerAgent *graphQLRunnerAgent          `json:"runner"`
		Problems    []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateRunnerAgentPayload struct {
		CreateRunnerAgent graphqlCreateRunnerAgentMutation `json:"createRunner"`
	}

	type testCase struct {
		responsePayload   interface{}
		expectRunnerAgent *types.RunnerAgent
		name              string
		expectErrorCode   types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully created runner agent",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateRunnerAgentPayload{
					CreateRunnerAgent: graphqlCreateRunnerAgentMutation{
						RunnerAgent: &graphQLRunnerAgent{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   "1",
							},
							ID:              graphql.String(runnerAgentID),
							Name:            "awesome-runner",
							GroupPath:       "groupA",
							ResourcePath:    "groupA/awesome-runner",
							Type:            "group",
							Description:     "a new runner",
							CreatedBy:       "someone",
							Tags:            []graphql.String{"tag1", "tag2"},
							RunUntaggedJobs: graphql.Boolean(true),
						},
					},
				},
			},
			expectRunnerAgent: &types.RunnerAgent{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   runnerAgentID,
					Version:              "1",
				},
				Name:            "awesome-runner",
				GroupPath:       "groupA",
				ResourcePath:    "groupA/awesome-runner",
				Type:            "group",
				Description:     "a new runner",
				CreatedBy:       "someone",
				Tags:            []string{"tag1", "tag2"},
				RunUntaggedJobs: true,
			},
		},
		{
			name: "create runner agent returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateRunnerAgentPayload{
					CreateRunnerAgent: graphqlCreateRunnerAgentMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "runner agent already exists",
								Type:    internal.Conflict,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorCode: types.ErrConflict,
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
				})}
			client.RunnerAgent = NewRunnerAgent(client)

			// Call the method being tested.
			runnerAgent, actualError := client.RunnerAgent.CreateRunnerAgent(ctx, &types.CreateRunnerInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectRunnerAgent != nil {
				require.NotNil(t, runnerAgent)
				assert.Equal(t, test.expectRunnerAgent, runnerAgent)
			}
		})
	}
}

func TestUpdateRunnerAgent(t *testing.T) {
	now := time.Now().UTC()

	runnerAgentID := "1"

	type graphqlUpdateRunnerAgentMutation struct {
		RunnerAgent *graphQLRunnerAgent          `json:"runner"`
		Problems    []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateRunnerAgentPayload struct {
		UpdateRunnerAgent graphqlUpdateRunnerAgentMutation `json:"updateRunner"`
	}

	type testCase struct {
		responsePayload   interface{}
		expectRunnerAgent *types.RunnerAgent
		name              string
		expectErrorCode   types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successful update of terraform module",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateRunnerAgentPayload{
					UpdateRunnerAgent: graphqlUpdateRunnerAgentMutation{
						RunnerAgent: &graphQLRunnerAgent{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   "1",
							},
							ID:              graphql.String(runnerAgentID),
							Name:            "awesome-runner",
							GroupPath:       "groupA",
							ResourcePath:    "groupA/awesome-runner",
							Description:     "a new description",
							CreatedBy:       "someone",
							Type:            "group",
							Tags:            []graphql.String{"newtag"},
							RunUntaggedJobs: graphql.Boolean(true),
						},
					},
				},
			},
			expectRunnerAgent: &types.RunnerAgent{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   runnerAgentID,
					Version:              "1",
				},
				Name:            "awesome-runner",
				GroupPath:       "groupA",
				ResourcePath:    "groupA/awesome-runner",
				Description:     "a new description",
				CreatedBy:       "someone",
				Type:            "group",
				Tags:            []string{"newtag"},
				RunUntaggedJobs: true,
			},
		},
		{
			name: "update module returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateRunnerAgentPayload{
					UpdateRunnerAgent: graphqlUpdateRunnerAgentMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "runner not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorCode: types.ErrNotFound,
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
				})}
			client.RunnerAgent = NewRunnerAgent(client)

			runnerAgent, actualError := client.RunnerAgent.UpdateRunnerAgent(ctx, &types.UpdateRunnerInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectRunnerAgent != nil {
				require.NotNil(t, runnerAgent)
				assert.Equal(t, test.expectRunnerAgent, runnerAgent)
			}
		})
	}
}

func TestDeleteRunnerAgent(t *testing.T) {
	type graphqlDeleteRunnerAgentMutation struct {
		RunnerAgent *graphQLRunnerAgent          `json:"runner"`
		Problems    []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteRunnerAgentPayload struct {
		DeleteRunnerAgent graphqlDeleteRunnerAgentMutation `json:"deleteRunner"`
	}

	type testCase struct {
		responsePayload interface{}
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successful deletion of runner agent",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteRunnerAgentPayload{
					DeleteRunnerAgent: graphqlDeleteRunnerAgentMutation{},
				},
			},
		},
		{
			name: "delete runner agent returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteRunnerAgentPayload{
					DeleteRunnerAgent: graphqlDeleteRunnerAgentMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "runner not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorCode: types.ErrNotFound,
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
				})}
			client.RunnerAgent = NewRunnerAgent(client)

			// Call the method being tested.
			err = client.RunnerAgent.DeleteRunnerAgent(ctx, &types.DeleteRunnerInput{})

			checkError(t, test.expectErrorCode, err)
		})
	}
}

func TestAssignServiceAccountToRunner(t *testing.T) {
	type graphqlAssignServiceAccountToRunnerAgentMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlAssignServiceAccountToRunnerAgentPayload struct {
		AssignServiceAccountToRunner graphqlAssignServiceAccountToRunnerAgentMutation `json:"assignServiceAccountToRunner"`
	}

	type testCase struct {
		responsePayload interface{}
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully assign service account to runner agent",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAssignServiceAccountToRunnerAgentPayload{
					AssignServiceAccountToRunner: graphqlAssignServiceAccountToRunnerAgentMutation{},
				},
			},
		},
		{
			name: "assigning service account to runner agent returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlAssignServiceAccountToRunnerAgentPayload{
					AssignServiceAccountToRunner: graphqlAssignServiceAccountToRunnerAgentMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "runner not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorCode: types.ErrNotFound,
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
				})}
			client.RunnerAgent = NewRunnerAgent(client)

			// Call the method being tested.
			err = client.RunnerAgent.AssignServiceAccountToRunnerAgent(ctx, &types.AssignServiceAccountToRunnerInput{})

			checkError(t, test.expectErrorCode, err)
		})
	}
}

func TestUnassignServiceAccountFromRunner(t *testing.T) {
	type graphqlUnassignServiceAccountFromRunnerAgentMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUnassignServiceAccountFromRunnerAgentPayload struct {
		UnassignServiceAccountFromRunner graphqlUnassignServiceAccountFromRunnerAgentMutation `json:"unassignServiceAccountFromRunner"`
	}

	type testCase struct {
		responsePayload interface{}
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully unassign service account from runner agent",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUnassignServiceAccountFromRunnerAgentPayload{
					UnassignServiceAccountFromRunner: graphqlUnassignServiceAccountFromRunnerAgentMutation{},
				},
			},
		},
		{
			name: "un-assigning service account from runner agent returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUnassignServiceAccountFromRunnerAgentPayload{
					UnassignServiceAccountFromRunner: graphqlUnassignServiceAccountFromRunnerAgentMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "runner not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorCode: types.ErrNotFound,
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
				})}
			client.RunnerAgent = NewRunnerAgent(client)

			// Call the method being tested.
			err = client.RunnerAgent.UnassignServiceAccountFromRunnerAgent(ctx, &types.AssignServiceAccountToRunnerInput{})

			checkError(t, test.expectErrorCode, err)
		})
	}
}
