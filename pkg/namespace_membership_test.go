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

func TestGetNamespaceMemberships(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	namespaceID := "group-or-workspace-id-1"
	namespaceName := "group-or-workspace-name-1"
	namespaceFullPath := "parent-group-1/" + namespaceName
	namespaceMembershipVersion := "membership-namespace-version-1"
	userID := "user-id"
	serviceAccountID := "service-account-id"
	teamID := "team-id"

	type testGraphQLMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	type testGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       testGraphQLMember
		Role         graphQLRole
	}

	type testGraphQLNamespace struct {
		Memberships []testGraphQLNamespaceMembership `json:"memberships"`
	}

	type testGraphqlNamespacePayload struct {
		Namespace *testGraphQLNamespace `json:"namespace"`
	}

	// test cases
	type testCase struct {
		name              string
		input             *types.GetNamespaceMembershipsInput
		responsePayload   interface{}
		expectMemberships []types.NamespaceMembership
		expectErrorCode   types.ErrorCode
	}

	/*
		Test case template:

		name              string
		input             *types.GetNamespaceMembershipsInput
		responsePayload   interface{}
		expectMemberships []types.NamespaceMembership
		expectErrorCode   types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "list three memberships",
			input: &types.GetNamespaceMembershipsInput{
				NamespacePath: namespaceFullPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: testGraphqlNamespacePayload{
					Namespace: &testGraphQLNamespace{
						Memberships: []testGraphQLNamespaceMembership{
							{
								ID: graphql.String(namespaceID),
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String(namespaceMembershipVersion),
								},
								Member: testGraphQLMember{
									Typename: "User",
									ID:       graphql.String(userID),
								},
								Role: graphQLRole{
									Name: graphql.String("viewer"),
								},
							},
							{
								ID: graphql.String(namespaceID),
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String(namespaceMembershipVersion),
								},
								Member: testGraphQLMember{
									Typename:  "ServiceAccount",
									ID:        graphql.String(serviceAccountID),
									GroupPath: graphql.String(namespaceFullPath),
								},
								Role: graphQLRole{
									Name: graphql.String("deployer"),
								},
							},
							{
								ID: graphql.String(namespaceID),
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String(namespaceMembershipVersion),
								},
								Member: testGraphQLMember{
									Typename: "Team",
									ID:       graphql.String(teamID),
								},
								Role: graphQLRole{
									Name: graphql.String("owner"),
								},
							},
						},
					},
				},
			},
			expectMemberships: []types.NamespaceMembership{
				{
					Metadata: types.ResourceMetadata{
						ID:                   namespaceID,
						CreationTimestamp:    &now,
						LastUpdatedTimestamp: &now,
						Version:              namespaceMembershipVersion,
					},
					UserID: &userID,
					Role:   "viewer",
				},
				{
					Metadata: types.ResourceMetadata{
						ID:                   namespaceID,
						CreationTimestamp:    &now,
						LastUpdatedTimestamp: &now,
						Version:              namespaceMembershipVersion,
					},
					ServiceAccountID: &serviceAccountID,
					Role:             "deployer",
				},
				{
					Metadata: types.ResourceMetadata{
						ID:                   namespaceID,
						CreationTimestamp:    &now,
						LastUpdatedTimestamp: &now,
						Version:              namespaceMembershipVersion,
					},
					TeamID: &teamID,
					Role:   "owner",
				},
			},
		},
		{
			name: "empty list but no error",
			input: &types.GetNamespaceMembershipsInput{
				NamespacePath: namespaceFullPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: testGraphqlNamespacePayload{
					Namespace: &testGraphQLNamespace{
						Memberships: []testGraphQLNamespaceMembership{},
					},
				},
			},
			expectMemberships: []types.NamespaceMembership{},
		},
		{
			name: "namespace not found",
			input: &types.GetNamespaceMembershipsInput{
				NamespacePath: namespaceFullPath + "/bogus",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: testGraphqlNamespacePayload{},
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMemberships, actualError := client.NamespaceMembership.GetMemberships(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			assert.Equal(t, test.expectMemberships == nil, actualMemberships == nil)
			assert.Equal(t, len(test.expectMemberships), len(actualMemberships))
			if test.expectMemberships != nil {
				for i := 0; i < len(test.expectMemberships); i++ {
					checkMembership(t, &test.expectMemberships[i], &actualMemberships[i])
				}
			}
		})
	}
}

func TestAddGroupMembership(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	groupName := "group-name-1"
	groupFullPath := "parent-group-1/" + groupName
	namespaceMembershipVersion := "membership-namespace-version-1"
	username := "user-1"
	userID := "user-id"
	serviceAccountID := "service-account-id"
	teamName := "team-1"
	teamID := "team-id"

	// based on graphQLNamespaceMembership
	type forTestGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       interface{} // allow user, service account, or team
		Role         graphQLRole
	}

	type graphqlGroupMembershipMutation struct {
		Membership forTestGraphQLNamespaceMembership
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlAddGroupMembershipPayload struct {
		CreateNamespaceMembership graphqlGroupMembershipMutation `json:"createNamespaceMembership"`
	}

	type testGraphqlMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.CreateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.CreateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "add user viewer",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath: groupFullPath,
				Username:      &username,
				Role:          "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddGroupMembershipPayload{
					CreateNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "User",
								ID:       graphql.String(userID),
							},
							Role: graphQLRole{
								Name: graphql.String("viewer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				UserID: &userID,
				Role:   "viewer",
			},
		},
		{
			name: "add service account deployer",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath:    groupFullPath,
				ServiceAccountID: &serviceAccountID,
				Role:             "deployer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddGroupMembershipPayload{
					CreateNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename:  "ServiceAccount",
								ID:        graphql.String(serviceAccountID),
								GroupPath: graphql.String(groupFullPath),
							},
							Role: graphQLRole{
								Name: graphql.String("deployer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				ServiceAccountID: &serviceAccountID,
				Role:             "deployer",
			},
		},
		{
			name: "add team owner",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath: groupFullPath,
				TeamName:      &teamName,
				Role:          "owner",
			},

			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddGroupMembershipPayload{
					CreateNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "Team",
								ID:       graphql.String(teamID),
							},
							Role: graphQLRole{
								Name: graphql.String("owner"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				TeamID: &teamID,
				Role:   "owner",
			},
		},
		{
			name: "error in add, no user, no service account, no team",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath: groupFullPath,
				Role:          "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Exactly one of User, ServiceAccount, team field must be defined",
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMembership, actualError := client.NamespaceMembership.AddMembership(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkMembership(t, test.expectMembership, actualMembership)
		})
	}
}

func TestUpdateGroupMembership(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	groupName := "group-name-1"
	groupFullPath := "parent-group-1/" + groupName
	namespaceMembershipID := "namespace-membership-id-1"
	namespaceMembershipVersion := "namespace-membership-version-1"
	userID := "user-id"
	serviceAccountID := "service-account-id"
	teamID := "team-id"

	// based on graphQLNamespaceMembership
	type forTestGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       interface{} // allow user, service account, or team
		Role         graphQLRole
	}

	type graphqlGroupMembershipMutation struct {
		Membership forTestGraphQLNamespaceMembership
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateGroupMembershipPayload struct {
		UpdateNamespaceMembership graphqlGroupMembershipMutation `json:"updateNamespaceMembership"`
	}

	type testGraphqlMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.UpdateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.UpdateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "update to user viewer",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupMembershipPayload{
					UpdateNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "User",
								ID:       graphql.String(userID),
							},
							Role: graphQLRole{
								Name: graphql.String("viewer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				UserID: &userID,
				Role:   "viewer",
			},
		},
		{
			name: "update to service account deployer",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "deployer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupMembershipPayload{
					UpdateNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename:  "ServiceAccount",
								ID:        graphql.String(serviceAccountID),
								GroupPath: graphql.String(groupFullPath),
							},
							Role: graphQLRole{
								Name: graphql.String("deployer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				ServiceAccountID: &serviceAccountID,
				Role:             "deployer",
			},
		},
		{
			name: "update to team owner",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "owner",
			},

			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateGroupMembershipPayload{
					UpdateNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "Team",
								ID:       graphql.String(teamID),
							},
							Role: graphQLRole{
								Name: graphql.String("owner"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				TeamID: &teamID,
				Role:   "owner",
			},
		},
		{
			name: "error in update, invalid ID",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Invalid namespace membership ID",
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMembership, actualError := client.NamespaceMembership.UpdateMembership(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkMembership(t, test.expectMembership, actualMembership)
		})
	}
}

func TestDeleteGroupMembership(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	groupID := "group-id-1"
	namespaceMembershipID := "namespace-membership-id-1"
	namespaceMembershipVersion := "namespace-membership-version-1"
	userID := "user-id"

	// based on graphQLNamespaceMembership
	type forTestGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       interface{} // allow user, service account, or team
		Role         graphQLRole
	}

	type graphqlGroupMembershipMutation struct {
		Membership forTestGraphQLNamespaceMembership
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteGroupMembershipPayload struct {
		DeleteNamespaceMembership graphqlGroupMembershipMutation `json:"deleteNamespaceMembership"`
	}

	type testGraphqlMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.DeleteNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.DeleteNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "successfully delete",
			input: &types.DeleteNamespaceMembershipInput{
				ID: namespaceMembershipID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteGroupMembershipPayload{
					DeleteNamespaceMembership: graphqlGroupMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(groupID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "User",
								ID:       graphql.String(userID),
							},
							Role: graphQLRole{
								Name: graphql.String("viewer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   groupID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				UserID: &userID,
				Role:   "viewer",
			},
		},
		{
			name: "error in delete, invalid ID",
			input: &types.DeleteNamespaceMembershipInput{
				ID: namespaceMembershipID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Invalid namespace membership ID",
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMembership, actualError := client.NamespaceMembership.DeleteMembership(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkMembership(t, test.expectMembership, actualMembership)
		})
	}
}

func TestAddWorkspaceMembership(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	workspaceID := "workspace-id-1"
	workspaceName := "workspace-name-1"
	parentGroupName := "parent-group-1"
	workspaceFullPath := parentGroupName + "/" + workspaceName
	namespaceMembershipVersion := "membership-namespace-version-1"
	username := "user-1"
	userID := "user-id"
	serviceAccountID := "service-account-id"
	teamName := "team-1"
	teamID := "team-id"

	// based on graphQLNamespaceMembership
	type forTestGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       interface{} // allow user, service account, or team
		Role         graphQLRole
	}

	type graphqlWorkspaceMembershipMutation struct {
		Membership forTestGraphQLNamespaceMembership
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlAddWorkspaceMembershipPayload struct {
		CreateNamespaceMembership graphqlWorkspaceMembershipMutation `json:"createNamespaceMembership"`
	}

	type testGraphqlMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.CreateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.CreateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "add user viewer",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath: workspaceFullPath,
				Username:      &username,
				Role:          "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddWorkspaceMembershipPayload{
					CreateNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "User",
								ID:       graphql.String(userID),
							},
							Role: graphQLRole{
								Name: graphql.String("viewer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				UserID: &userID,
				Role:   "viewer",
			},
		},
		{
			name: "add service account deployer",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath:    workspaceFullPath,
				ServiceAccountID: &serviceAccountID,
				Role:             "deployer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddWorkspaceMembershipPayload{
					CreateNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename:  "ServiceAccount",
								ID:        graphql.String(serviceAccountID),
								GroupPath: graphql.String(parentGroupName),
							},
							Role: graphQLRole{
								Name: graphql.String("deployer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				ServiceAccountID: &serviceAccountID,
				Role:             "deployer",
			},
		},
		{
			name: "add team owner",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath: workspaceFullPath,
				TeamName:      &teamName,
				Role:          "owner",
			},

			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddWorkspaceMembershipPayload{
					CreateNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "Team",
								ID:       graphql.String(teamID),
							},
							Role: graphQLRole{
								Name: graphql.String("owner"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				TeamID: &teamID,
				Role:   "owner",
			},
		},
		{
			name: "error in add, no user, no service account, no team",
			input: &types.CreateNamespaceMembershipInput{
				NamespacePath: workspaceFullPath,
				Role:          "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Exactly one of User, ServiceAccount, team field must be defined",
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMembership, actualError := client.NamespaceMembership.AddMembership(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkMembership(t, test.expectMembership, actualMembership)
		})
	}
}

func TestUpdateWorkspaceMembership(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	workspaceID := "workspace-id-1"
	parentGroupName := "parent-group-1"
	namespaceMembershipID := "namespace-membership-id-1"
	namespaceMembershipVersion := "namespace-membership-version-1"
	userID := "user-id"
	serviceAccountID := "service-account-id"
	teamID := "team-id"

	// based on graphQLNamespaceMembership
	type forTestGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       interface{} // allow user, service account, or team
		Role         graphQLRole
	}

	type graphqlWorkspaceMembershipMutation struct {
		Membership forTestGraphQLNamespaceMembership
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateWorkspaceMembershipPayload struct {
		UpdateNamespaceMembership graphqlWorkspaceMembershipMutation `json:"updateNamespaceMembership"`
	}

	type testGraphqlMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.UpdateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.UpdateNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "update to user viewer",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspaceMembershipPayload{
					UpdateNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "User",
								ID:       graphql.String(userID),
							},
							Role: graphQLRole{
								Name: graphql.String("viewer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				UserID: &userID,
				Role:   "viewer",
			},
		},
		{
			name: "update to service account deployer",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "deployer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspaceMembershipPayload{
					UpdateNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename:  "ServiceAccount",
								ID:        graphql.String(serviceAccountID),
								GroupPath: graphql.String(parentGroupName),
							},
							Role: graphQLRole{
								Name: graphql.String("deployer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				ServiceAccountID: &serviceAccountID,
				Role:             "deployer",
			},
		},
		{
			name: "update to team owner",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "owner",
			},

			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspaceMembershipPayload{
					UpdateNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "Team",
								ID:       graphql.String(teamID),
							},
							Role: graphQLRole{
								Name: graphql.String("owner"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				TeamID: &teamID,
				Role:   "owner",
			},
		},
		{
			name: "error in update, invalid ID",
			input: &types.UpdateNamespaceMembershipInput{
				ID:   namespaceMembershipID,
				Role: "viewer",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Invalid namespace membership ID",
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMembership, actualError := client.NamespaceMembership.UpdateMembership(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkMembership(t, test.expectMembership, actualMembership)
		})
	}
}

func TestDeleteWorkspaceMembership(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	workspaceID := "workspace-id-1"
	namespaceMembershipID := "namespace-membership-id-1"
	namespaceMembershipVersion := "namespace-membership-version-1"
	userID := "user-id"

	// based on graphQLNamespaceMembership
	type forTestGraphQLNamespaceMembership struct {
		ID           graphql.String
		Metadata     internal.GraphQLMetadata
		ResourcePath graphql.String
		Member       interface{} // allow user, service account, or team
		Role         graphQLRole
	}

	type graphqlWorkspaceMembershipMutation struct {
		Membership forTestGraphQLNamespaceMembership
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteWorkspaceMembershipPayload struct {
		DeleteNamespaceMembership graphqlWorkspaceMembershipMutation `json:"deleteNamespaceMembership"`
	}

	type testGraphqlMember struct {
		Typename  string         `json:"__typename"`
		GroupPath graphql.String // from graphQLServiceAccount
		ID        graphql.String // common among the three below
		graphQLUser
		graphQLServiceAccount
		graphQLTeam
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.DeleteNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.DeleteNamespaceMembershipInput
		responsePayload  interface{}
		expectMembership *types.NamespaceMembership
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "successfully delete",
			input: &types.DeleteNamespaceMembershipInput{
				ID: namespaceMembershipID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteWorkspaceMembershipPayload{
					DeleteNamespaceMembership: graphqlWorkspaceMembershipMutation{
						Membership: forTestGraphQLNamespaceMembership{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(namespaceMembershipVersion),
							},
							Member: testGraphqlMember{
								Typename: "User",
								ID:       graphql.String(userID),
							},
							Role: graphQLRole{
								Name: graphql.String("viewer"),
							},
						},
					},
				},
			},
			expectMembership: &types.NamespaceMembership{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceMembershipVersion,
				},
				UserID: &userID,
				Role:   "viewer",
			},
		},
		{
			name: "error in delete, invalid ID",
			input: &types.DeleteNamespaceMembershipInput{
				ID: namespaceMembershipID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Invalid namespace membership ID",
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
			client.NamespaceMembership = NewNamespaceMembership(client)

			// Call the method being tested.
			actualMembership, actualError := client.NamespaceMembership.DeleteMembership(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkMembership(t, test.expectMembership, actualMembership)
		})
	}
}

// Utility functions:

func checkMembership(t *testing.T, expectMembership, actualMembership *types.NamespaceMembership) {
	if expectMembership != nil {
		require.NotNil(t, actualMembership)
		assert.Equal(t, expectMembership, actualMembership)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Group)(nil)
		assert.Equal(t, (*types.NamespaceMembership)(nil), actualMembership)
	}
}
