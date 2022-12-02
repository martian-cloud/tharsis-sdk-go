package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/likexian/gokit/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestCreateGroup(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	groupName := "group-name-1"
	groupID := "group-id-1"
	groupVersion := "group-version-1"
	groupDescription := "group description 1"
	nestedPath := parentGroupName + "/" + groupName

	type graphqlCreateGroupMutation struct {
		Group    graphQLGroup                 `json:"group"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateGroupPayload struct {
		CreateGroup graphqlCreateGroupMutation `json:"createGroup"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		input              *types.CreateGroupInput
		expectGroup        *types.Group
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive, top-level
		{
			name: "Successfully created top-level group",
			input: &types.CreateGroupInput{
				Name:        groupName,
				ParentPath:  nil,
				Description: groupDescription,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateGroupPayload{
					CreateGroup: graphqlCreateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							Name:        graphql.String(groupName),
							Description: graphql.String(groupDescription),
							FullPath:    graphql.String(groupName),
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
				Name:        groupName,
				Description: groupDescription,
				FullPath:    groupName,
			},
		},

		// positive, nested
		{
			name: "Successfully created nested group",
			input: &types.CreateGroupInput{
				Name:        groupName,
				ParentPath:  &parentGroupName,
				Description: groupDescription,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateGroupPayload{
					CreateGroup: graphqlCreateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							Name:        graphql.String(groupName),
							Description: graphql.String(groupDescription),
							FullPath:    graphql.String(nestedPath),
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
				Name:        groupName,
				Description: groupDescription,
				FullPath:    nestedPath,
			},
		},

		// negative: query returns error
		{
			name:  "negative: query to create group returned error",
			input: &types.CreateGroupInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateGroupPayload{
					CreateGroup: graphqlCreateGroupMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Group with path non-existent not found",
								Type:    internal.Conflict,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems creating group: Group with path non-existent not found",
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
			client.Group = NewGroup(client)

			// Call the method being tested.
			actualGroup, actualError := client.Group.CreateGroup(ctx, test.input)

			checkError(t, test.expectErrorMessage, actualError)
			checkGroup(t, test.expectGroup, actualGroup)
		})
	}
}

func TestGetGroup(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	groupName := "group-name-1"
	groupDescription := "group description 1"
	nestedPath := parentGroupName + "/" + groupName
	groupID := "group-id-1"
	groupVersion := "group-version-1"

	type graphqlGroupPayload struct {
		Node *graphQLGroup `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		expectGroup        *types.Group
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return group by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayload{
					Node: &graphQLGroup{
						ID: graphql.String(groupID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(groupVersion),
						},
						Name:        graphql.String(groupName),
						Description: graphql.String(groupDescription),
						FullPath:    graphql.String(nestedPath),
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
				Name:        groupName,
				Description: groupDescription,
				FullPath:    nestedPath,
			},
		},

		// negative: query returns error, invalid ID--payload taken from GraphiQL
		{
			name: "negative: query returns error, invalid ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"6f3106c6-c342-4790-a667-963a850d9�d4\" (SQLSTATE 22P02)",
						Path: []string{
							"node",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorMessage: "Message: ERROR: invalid input syntax for type uuid: \"6f3106c6-c342-4790-a667-963a850d9�d4\" (SQLSTATE 22P02), Locations: []",
		},

		// negative: query returns error, not found error--payload taken from GraphiQL
		{
			name: "negative: query returns error, not found error",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Group with id 6f3106c6-c342-4790-a667-963a850d34d4 not found",
						Path: []string{
							"node",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "NOT_FOUND",
						},
					},
				},
			},
			expectErrorMessage: "Message: Group with id 6f3106c6-c342-4790-a667-963a850d34d4 not found, Locations: []",
		},

		// negative: theoretical quiet not found
		{
			name: "negative: theoretical quiet not found",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGroupPayload{},
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
			client.Group = NewGroup(client)

			// Call the method being tested.
			actualGroup, actualError := client.Group.GetGroup(
				ctx,
				&types.GetGroupInput{ID: &groupID},
			)

			checkError(t, test.expectErrorMessage, actualError)
			checkGroup(t, test.expectGroup, actualGroup)
		})
	}
}

/*

func TestUpdateGroup(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	groupVersion := "group-version-1"
	managedIdentityID := "managed-identity-id-1"

	type graphqlUpdateGroupMutation struct {
		Group    graphQLGroup                 `json:"group"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateGroupPayload struct {
		UpdateGroup graphqlUpdateGroupMutation `json:"updateGroup"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		input              *types.UpdateGroupInput
		expectGroup        *types.Group
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully updated group",
			input: &types.UpdateGroupInput{
				RunStage:               types.JobApplyType,
				AllowedUsers:           []string{"test-user-3", "test-user-4"},
				AllowedServiceAccounts: []string{"test-service-account-5", "test-service-account-6"},
				AllowedTeams:           []string{"test-team-7", "test team-8"},
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupPayload{
					UpdateGroup: graphqlUpdateGroupMutation{
						Group: graphQLGroup{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(groupVersion),
							},
							RunStage: graphql.String(types.JobPlanType),
							AllowedUsers: []graphQLUser{
								{Username: "test-user-3"},
								{Username: "test-user-4"},
							},
							AllowedServiceAccounts: []graphQLServiceAccount{
								{Name: "test-service-account-5"},
								{Name: "test-service-account-6"},
							},
							AllowedTeams: []graphQLTeam{
								{Name: "test-team-7"},
								{Name: "test-team-8"},
							},
							ManagedIdentity: GraphQLManagedIdentity{
								ID: graphql.String(managedIdentityID),
							},
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
				RunStage: types.JobPlanType,
				AllowedUsers: []types.User{
					{Username: "test-user-3"},
					{Username: "test-user-4"},
				},
				AllowedServiceAccounts: []types.ServiceAccount{
					{Name: "test-service-account-5"},
					{Name: "test-service-account-6"},
				},
				AllowedTeams: []types.Team{
					{Name: "test-team-7"},
					{Name: "test-team-8"},
				},
				ManagedIdentityID: managedIdentityID,
			},
		},

		// negative: group update query returns error
		{
			name:  "negative: group update query returns error",
			input: &types.UpdateGroupInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Argument \"input\" has invalid value {id: \"TVJfZmUyZWI1NjQtNjMxMS00MmFlLTkwMWYtOTE5NTEyNWNhOTJh\", runStage: invalid, allowedUsers: [\"robert.richesjr\"], allowedServiceAccounts: [\"provider-test-parent-group/sa1\"], allowedTeams: [\"team1\", \"team2\"]}.\nIn field \"runStage\": Expected type \"JobType\", found invalid.",
						Locations: []fakeGraphqlResponseLocation{
							{
								Line:   3,
								Column: 12,
							},
						},
					},
				},
			},
			expectErrorMessage: "Message: Argument \"input\" has invalid value {id: \"TVJfZmUyZWI1NjQtNjMxMS00MmFlLTkwMWYtOTE5NTEyNWNhOTJh\", runStage: invalid, allowedUsers: [\"robert.richesjr\"], allowedServiceAccounts: [\"provider-test-parent-group/sa1\"], allowedTeams: [\"team1\", \"team2\"]}.\nIn field \"runStage\": Expected type \"JobType\", found invalid., Locations: [{Line:3 Column:12}]",
		},

		// negative: query behaves as if the specified access rule did not exist
		{
			name:  "negative: query behaves as if the specified access rule did not exist",
			input: &types.UpdateGroupInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupPayload{
					UpdateGroup: graphqlUpdateGroupMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "group with ID fe2eb564-6311-52ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems updating group: group with ID fe2eb564-6311-52ae-901f-9195125ca92a not found",
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualGroup, actualError := client.ManagedIdentity.UpdateGroup(ctx, test.input)

			checkError(t, test.expectErrorMessage, actualError)
			checkGroup(t, test.expectGroup, actualGroup)
		})
	}
}

func TestDeleteGroup(t *testing.T) {
	groupID := "group-id-1"

	// In GraphiQL, an 'group' element appeared here.  However, it would not unmarshal when run from a test.
	type graphqlDeleteGroupMutation struct {
		// Group graphQLGroup `json:"group"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteGroupPayload struct {
		DeleteGroup graphqlDeleteGroupMutation `json:"deleteGroup"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		input              *types.DeleteGroupInput
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully deleted group",
			input: &types.DeleteGroupInput{
				ID: groupID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteGroupPayload{
					DeleteGroup: graphqlDeleteGroupMutation{},
				},
			},
		},

		// negative: mutation returns error
		{
			name:  "negative: group delete mutation returns error",
			input: &types.DeleteGroupInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02)",
						Path: []string{
							"deleteGroup",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorMessage: "Message: ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02), Locations: []",
		},

		// negative: mutation behaves as if the specified access rule did not exist
		{
			name:  "negative: mutation behaves as if the specified access rule did not exist",
			input: &types.DeleteGroupInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteGroupPayload{
					DeleteGroup: graphqlDeleteGroupMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "group with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems deleting group: group with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualError := client.ManagedIdentity.DeleteGroup(ctx, test.input)

			checkError(t, test.expectErrorMessage, actualError)
		})
	}

}

*/

// Utility functions:

func checkServiceAccount(t *testing.T, expectGroup, actualGroup *types.Group) {
	if expectGroup != nil {
		require.NotNil(t, actualGroup)
		assert.Equal(t, actualGroup, expectGroup)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Group)(nil)
		assert.Equal(t, actualGroup, (*types.Group)(nil))
	}
}

// The End.
