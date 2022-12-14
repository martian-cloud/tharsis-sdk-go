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

// TODO: This module has unit tests only for newer method(s) added in December, 2022.
// The other methods should also have unit tests added, including a TestGetWorkspaceByPath.

func TestGetWorkspaceByID(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	workspaceID := "workspace-id-1"
	workspaceVersion := "workspace-version-1"

	// Field name taken from GraphiQL.
	type graphqlNodeWorkspacePayload struct {
		Node *graphQLWorkspace `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		expectWorkspace    *types.Workspace
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return workspace by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeWorkspacePayload{
					Node: &graphQLWorkspace{
						ID: graphql.String(workspaceID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(workspaceVersion),
						},
						Name:        "nm01",
						Description: "de01",
						FullPath:    "fp01",
					},
				},
			},
			expectWorkspace: &types.Workspace{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              workspaceVersion,
				},
				Name:        "nm01",
				Description: "de01",
				FullPath:    "fp01",
			},
		},

		// query returns error as if the ID is invalid
		{
			name: "query returns error as if the ID is invalid",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeWorkspacePayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02)",
						Path: []string{
							"workspace",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorMessage: "Message: ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02), Locations: []",
		},

		// query returns nil workspace, as if the specified workspace does not exist.
		{
			name: "query returns nil workspace, as if the specified workspace does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeWorkspacePayload{},
			},
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
				graphqlClient: *newGraphQLClientForTest(testClientInput{
					statusToReturn:  http.StatusOK,
					payloadToReturn: string(payload),
				}),
			}
			client.Workspaces = NewWorkspaces(client)

			// Call the method being tested.
			actualWorkspace, actualError := client.Workspaces.GetWorkspace(
				ctx,
				&types.GetWorkspaceInput{ID: &workspaceID},
			)

			checkError(t, test.expectErrorMessage, actualError)
			checkWorkspace(t, test.expectWorkspace, actualWorkspace)
		})
	}
}

// Utility functions:

func checkWorkspace(t *testing.T, expectWorkspace, actualWorkspace *types.Workspace) {
	if expectWorkspace != nil {
		require.NotNil(t, actualWorkspace)
		assert.Equal(t, expectWorkspace, actualWorkspace)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Workspace)(nil)
		assert.Equal(t, (*types.Workspace)(nil), actualWorkspace)
	}
}

// The End.
