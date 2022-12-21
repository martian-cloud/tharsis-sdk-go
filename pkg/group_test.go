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
// The other methods should also have unit tests added, including a TestGetGroupByPath.

func TestGetGroupByID(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	groupVersion := "group-version-1"

	// Field name taken from GraphiQL.
	type graphqlNodeGroupPayload struct {
		Node *graphQLGroup `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		expectGroup     *types.Group
		name            string
		expectErrorCode ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return group by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeGroupPayload{
					Node: &graphQLGroup{
						ID: graphql.String(groupID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(groupVersion),
						},
						Name:        "nm01",
						Description: "de01",
						FullPath:    "fp01",
					},
				},
			},
			expectGroup: &types.Group{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              groupVersion,
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
				Data: graphqlNodeGroupPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02)",
						Path: []string{
							"group",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
		},

		// query returns nil group, as if the specified group does not exist.
		{
			name: "query returns nil group, as if the specified group does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeGroupPayload{},
			},
			expectErrorCode: ErrNotFound,
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
			client.Group = NewGroup(client)

			// Call the method being tested.
			actualGroup, actualError := client.Group.GetGroup(
				ctx,
				&types.GetGroupInput{ID: &groupID},
			)

			checkError(t, test.expectErrorCode, actualError)
			checkGroup(t, test.expectGroup, actualGroup)
		})
	}
}

// Utility functions:

func checkGroup(t *testing.T, expectGroup, actualGroup *types.Group) {
	if expectGroup != nil {
		require.NotNil(t, actualGroup)
		assert.Equal(t, expectGroup, actualGroup)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Group)(nil)
		assert.Equal(t, (*types.Group)(nil), actualGroup)
	}
}

// The End.
