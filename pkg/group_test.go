package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/aws/smithy-go/ptr"
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
	groupPath := "fp01"
	groupVersion := "group-version-1"

	type graphqlGroupPayloadByID struct {
		Node *graphQLGroup `json:"node"`
	}

	type graphqlGroupPayloadByPath struct {
		Group *graphQLGroup `json:"group"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.GetGroupInput
		expectGroup     *types.Group
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return group by ID",
			input: &types.GetGroupInput{
				ID: &groupID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayloadByID{
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

		{
			name: "Successfully return group by path",
			input: &types.GetGroupInput{
				Path: &groupPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayloadByPath{
					Group: &graphQLGroup{
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

		{
			name:            "returns an error since ID and path were unspecified",
			input:           &types.GetGroupInput{},
			expectErrorCode: types.ErrBadRequest,
		},

		{
			name: "verify that correct error is returned",
			input: &types.GetGroupInput{
				ID: &groupID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayloadByID{},
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

		// query returns nil group, as if the specified group does not exist.
		{
			name: "query returns nil group, as if the specified group does not exist",
			input: &types.GetGroupInput{
				ID: &groupID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayloadByID{},
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
			client.Group = NewGroup(client)

			// Call the method being tested.
			actualGroup, actualError := client.Group.GetGroup(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkGroup(t, test.expectGroup, actualGroup)
		})
	}
}

func TestUpdateGroup(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	groupVersion := "group-version-1"
	groupName := "group-name-1"
	groupFullPath := "parent-group-1/" + groupName
	groupDescription := "group-description-1"

	type graphqlUpdateGroupMutation struct {
		Group    graphQLGroup                 `json:"group"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateGroupPayload struct {
		UpdateGroup graphqlUpdateGroupMutation `json:"updateGroup"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.UpdateGroupInput
		expectGroup     *types.Group
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully update group by ID",
			input: &types.UpdateGroupInput{
				ID:          &groupID,
				Description: groupDescription,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupPayload{
					UpdateGroup: graphqlUpdateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							Name:        "nm01",
							FullPath:    "fp01",
							Description: "de01",
						},
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
				FullPath:    "fp01",
				Description: "de01",
			},
		},
		{
			name: "Successfully update group by path",
			input: &types.UpdateGroupInput{
				GroupPath:   &groupFullPath,
				Description: groupDescription,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupPayload{
					UpdateGroup: graphqlUpdateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							Name:        "nm01",
							FullPath:    "fp01",
							Description: "de01",
						},
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
				FullPath:    "fp01",
				Description: "de01",
			},
		},
		{
			name:            "returns an error since neither ID nor path was supplied",
			input:           &types.UpdateGroupInput{},
			expectErrorCode: types.ErrBadRequest,
		},
		{
			name: "returns an error since both ID and path were unspecified",
			input: &types.UpdateGroupInput{
				ID:        &groupID,
				GroupPath: &groupFullPath,
			},
			expectErrorCode: types.ErrBadRequest,
		},
		{
			name: "verify that correct error is returned",
			input: &types.UpdateGroupInput{
				ID: ptr.String("invalid"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupPayload{},
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

		// query returns nil group, as if the specified group does not exist.
		{
			name: "query returns nil group, as if the specified group does not exist",
			input: &types.UpdateGroupInput{
				ID: &groupID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "group not found",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "NOT_FOUND",
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
				}),
			}
			client.Group = NewGroup(client)

			// Call the method being tested.
			actualGroup, actualError := client.Group.UpdateGroup(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkGroup(t, test.expectGroup, actualGroup)
		})
	}
}

func TestMigrateGroup(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	groupVersion := "group-version-1"
	groupName := "group-name-1"
	groupFullPath := "parent-group-1/" + groupName

	parentPath1 := "some-top-level/parent-1-name"

	type graphqlMigrateGroupMutation struct {
		Group    graphQLGroup                 `json:"group"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlMigrateGroupPayload struct {
		MigrateGroup graphqlMigrateGroupMutation `json:"migrateGroup"`
	}

	// test cases
	type testCase struct {
		name            string
		input           *types.MigrateGroupInput
		responsePayload interface{}
		expectGroup     *types.Group
		expectErrorCode types.ErrorCode
	}

	/*
		Test case template:

		name            string
		input           *types.MigrateGroupInput
		responsePayload interface{}
		expectGroup     *types.Group
		expectErrorCode types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "Successfully migrate group by ID",
			input: &types.MigrateGroupInput{
				GroupPath:     groupFullPath,
				NewParentPath: &parentPath1,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMigrateGroupPayload{
					MigrateGroup: graphqlMigrateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							Name:        "nm01",
							FullPath:    "fp01",
							Description: "de01",
						},
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
				FullPath:    "fp01",
				Description: "de01",
			},
		},
		{
			name: "Successfully migrate group by path",
			input: &types.MigrateGroupInput{
				GroupPath:     groupFullPath,
				NewParentPath: &parentPath1,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMigrateGroupPayload{
					MigrateGroup: graphqlMigrateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							Name:        "nm01",
							FullPath:    "fp01",
							Description: "de01",
						},
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
				FullPath:    "fp01",
				Description: "de01",
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.MigrateGroupInput{
				GroupPath: "invalid",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMigrateGroupPayload{},
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

		// query returns nil group, as if the specified group does not exist.
		{
			name: "query returns nil group, as if the specified group does not exist",
			input: &types.MigrateGroupInput{
				GroupPath: groupFullPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMigrateGroupPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "group not found",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "NOT_FOUND",
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
				}),
			}
			client.Group = NewGroup(client)

			// Call the method being tested.
			actualGroup, actualError := client.Group.MigrateGroup(ctx, test.input)

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
