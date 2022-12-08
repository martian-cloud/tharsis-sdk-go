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

// TODO: This module has unit tests only for newer managed identity (access rule) methods
// added around November, 2022.  TODO: The other methods should also have unit tests added.

func TestGetManagedIdentity(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	managedIdentityID := "managed-identity-id-1"
	managedIdentityVersion := "managed-identity-version-1"

	// Field name taken from GraphiQL.
	type graphqlManagedIdentityPayload struct {
		ManagedIdentity *GraphQLManagedIdentity `json:"managedIdentity"`
	}

	// test cases
	type testCase struct {
		responsePayload       interface{}
		expectManagedIdentity *types.ManagedIdentity
		name                  string
		expectErrorMessage    string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return managed identity by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayload{
					ManagedIdentity: &GraphQLManagedIdentity{
						ID: graphql.String(managedIdentityID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(managedIdentityVersion),
						},
						Type:         "t01",
						ResourcePath: "rp01",
						Name:         "nm01",
						Description:  "de01",
						Data:         "da01",
						CreatedBy:    "cr01",
					},
				},
			},
			expectManagedIdentity: &types.ManagedIdentity{
				Metadata: types.ResourceMetadata{
					ID:                   managedIdentityID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              managedIdentityVersion,
				},
				Type:         "t01",
				ResourcePath: "rp01",
				Name:         "nm01",
				Description:  "de01",
				Data:         "da01",
				CreatedBy:    "cr01",
			},
		},

		// query returns error as if the ID is invalid
		{
			name: "query returns error as if the ID is invalid",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02)",
						Path: []string{
							"managedIdentity",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorMessage: "Message: ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02), Locations: []",
		},

		// query returns nil managed identity, as if the specified managed identity does not exist.
		{
			name: "query returns nil managed identity, as if the specified managed identity does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayload{},
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualIdentity, actualError := client.ManagedIdentity.GetManagedIdentity(
				ctx,
				&types.GetManagedIdentityInput{ID: managedIdentityID},
			)

			checkError(t, test.expectErrorMessage, actualError)
			checkIdentity(t, test.expectManagedIdentity, actualIdentity)
		})
	}
}

func TestCreateManagedIdentityAccessRule(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	accessRuleID := "access-rule-id-1"
	accessRuleVersion := "access-rule-version-1"
	managedIdentityID := "managed-identity-id-1"

	type graphqlCreateManagedIdentityAccessRuleMutation struct {
		AccessRule graphQLManagedIdentityAccessRule `json:"accessRule"`
		Problems   []fakeGraphqlResponseProblem     `json:"problems"`
	}

	type graphqlCreateManagedIdentityAccessRulePayload struct {
		CreateManagedIdentityAccessRule graphqlCreateManagedIdentityAccessRuleMutation `json:"createManagedIdentityAccessRule"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		input              *types.CreateManagedIdentityAccessRuleInput
		expectAccessRule   *types.ManagedIdentityAccessRule
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully created managed identity access rule",
			input: &types.CreateManagedIdentityAccessRuleInput{
				RunStage:               types.JobPlanType,
				AllowedUsers:           []string{"test-user-1", "test-user-2"},
				AllowedServiceAccounts: []string{"test-service-account-1", "test-service-account-2"},
				AllowedTeams:           []string{"test-team-1", "test team-2"},
				ManagedIdentityID:      managedIdentityID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateManagedIdentityAccessRulePayload{
					CreateManagedIdentityAccessRule: graphqlCreateManagedIdentityAccessRuleMutation{
						AccessRule: graphQLManagedIdentityAccessRule{
							ID: graphql.String(accessRuleID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(accessRuleVersion),
							},
							RunStage: graphql.String(types.JobPlanType),
							AllowedUsers: []graphQLUser{
								{Username: "test-user-1"},
								{Username: "test-user-2"},
							},
							AllowedServiceAccounts: []graphQLServiceAccount{
								{Name: "test-service-account-1"},
								{Name: "test-service-account-2"},
							},
							AllowedTeams: []graphQLTeam{
								{Name: "test-team-1"},
								{Name: "test-team-2"},
							},
							ManagedIdentity: GraphQLManagedIdentity{
								ID: graphql.String(managedIdentityID),
							},
						},
					},
				},
			},
			expectAccessRule: &types.ManagedIdentityAccessRule{
				Metadata: types.ResourceMetadata{
					ID:                   accessRuleID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              accessRuleVersion,
				},
				RunStage: types.JobPlanType,
				AllowedUsers: []types.User{
					{Username: "test-user-1"},
					{Username: "test-user-2"},
				},
				AllowedServiceAccounts: []types.ServiceAccount{
					{Name: "test-service-account-1", OIDCTrustPolicies: []types.OIDCTrustPolicy{}},
					{Name: "test-service-account-2", OIDCTrustPolicies: []types.OIDCTrustPolicy{}},
				},
				AllowedTeams: []types.Team{
					{Name: "test-team-1"},
					{Name: "test-team-2"},
				},
				ManagedIdentityID: managedIdentityID,
			},
		},

		// negative: query returns error
		{
			name:  "negative: query to create managed identity access rule returned error",
			input: &types.CreateManagedIdentityAccessRuleInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateManagedIdentityAccessRulePayload{
					CreateManagedIdentityAccessRule: graphqlCreateManagedIdentityAccessRuleMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Rule for run stage apply already exists",
								Type:    internal.Conflict,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems creating managed identity access rule: Rule for run stage apply already exists",
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
			actualAccessRule, actualError := client.ManagedIdentity.CreateManagedIdentityAccessRule(ctx, test.input)

			checkError(t, test.expectErrorMessage, actualError)
			checkAccessRule(t, test.expectAccessRule, actualAccessRule)
		})
	}
}

func TestGetManagedIdentityAccessRule(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	accessRuleID := "access-rule-id-1"
	accessRuleVersion := "access-rule-version-1"
	managedIdentityID := "managed-identity-id-1"

	type graphqlManagedIdentityAccessRulePayload struct {
		Node *graphQLManagedIdentityAccessRule `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		expectAccessRule   *types.ManagedIdentityAccessRule
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return managed identity access rule by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityAccessRulePayload{
					Node: &graphQLManagedIdentityAccessRule{
						ID: graphql.String(accessRuleID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(accessRuleVersion),
						},
						RunStage:               graphql.String(types.JobPlanType),
						AllowedUsers:           []graphQLUser{},
						AllowedServiceAccounts: []graphQLServiceAccount{},
						AllowedTeams:           []graphQLTeam{},
						ManagedIdentity: GraphQLManagedIdentity{
							ID: graphql.String(managedIdentityID),
						},
					},
				},
			},
			expectAccessRule: &types.ManagedIdentityAccessRule{
				Metadata: types.ResourceMetadata{
					ID:                   accessRuleID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              accessRuleVersion,
				},
				RunStage:               types.JobPlanType,
				AllowedUsers:           []types.User{},
				AllowedServiceAccounts: []types.ServiceAccount{},
				AllowedTeams:           []types.Team{},
				ManagedIdentityID:      managedIdentityID,
			},
		},

		// negative: query returns error, invalid ID--payload taken from GraphiQL
		{
			name: "negative: query returns error, invalid ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityAccessRulePayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Invalid ID",
						Path: []string{
							"node",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "BAD_REQUEST",
						},
					},
				},
			},
			expectErrorMessage: "Message: Invalid ID, Locations: []",
		},

		// negative: query returns error, not found error--payload taken from GraphiQL
		{
			name: "negative: query returns error, not found error",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityAccessRulePayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Managed identity access rule with ID 6f9666fb-1a4e-5755-b4a5-d8dfe15aa187 not found",
						Path: []string{
							"node",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "NOT_FOUND",
						},
					},
				},
			},
			expectErrorMessage: "Message: Managed identity access rule with ID 6f9666fb-1a4e-5755-b4a5-d8dfe15aa187 not found, Locations: []",
		},

		// negative: theoretical quiet not found
		{
			name: "negative: theoretical quiet not found",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityAccessRulePayload{},
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualAccessRule, actualError := client.ManagedIdentity.GetManagedIdentityAccessRule(
				ctx,
				&types.GetManagedIdentityAccessRuleInput{ID: accessRuleID},
			)

			checkError(t, test.expectErrorMessage, actualError)
			checkAccessRule(t, test.expectAccessRule, actualAccessRule)
		})
	}
}

func TestUpdateManagedIdentityAccessRule(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	accessRuleID := "access-rule-id-1"
	accessRuleVersion := "access-rule-version-1"
	managedIdentityID := "managed-identity-id-1"

	type graphqlUpdateManagedIdentityAccessRuleMutation struct {
		AccessRule graphQLManagedIdentityAccessRule `json:"accessRule"`
		Problems   []fakeGraphqlResponseProblem     `json:"problems"`
	}

	type graphqlUpdateManagedIdentityAccessRulePayload struct {
		UpdateManagedIdentityAccessRule graphqlUpdateManagedIdentityAccessRuleMutation `json:"updateManagedIdentityAccessRule"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		input              *types.UpdateManagedIdentityAccessRuleInput
		expectAccessRule   *types.ManagedIdentityAccessRule
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully updated managed identity access rule",
			input: &types.UpdateManagedIdentityAccessRuleInput{
				RunStage:               types.JobApplyType,
				AllowedUsers:           []string{"test-user-3", "test-user-4"},
				AllowedServiceAccounts: []string{"test-service-account-5", "test-service-account-6"},
				AllowedTeams:           []string{"test-team-7", "test team-8"},
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateManagedIdentityAccessRulePayload{
					UpdateManagedIdentityAccessRule: graphqlUpdateManagedIdentityAccessRuleMutation{
						AccessRule: graphQLManagedIdentityAccessRule{
							ID: graphql.String(accessRuleID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(accessRuleVersion),
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
			expectAccessRule: &types.ManagedIdentityAccessRule{
				Metadata: types.ResourceMetadata{
					ID:                   accessRuleID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              accessRuleVersion,
				},
				RunStage: types.JobPlanType,
				AllowedUsers: []types.User{
					{Username: "test-user-3"},
					{Username: "test-user-4"},
				},
				AllowedServiceAccounts: []types.ServiceAccount{
					{Name: "test-service-account-5", OIDCTrustPolicies: []types.OIDCTrustPolicy{}},
					{Name: "test-service-account-6", OIDCTrustPolicies: []types.OIDCTrustPolicy{}},
				},
				AllowedTeams: []types.Team{
					{Name: "test-team-7"},
					{Name: "test-team-8"},
				},
				ManagedIdentityID: managedIdentityID,
			},
		},

		// negative: managed identity access rule update query returns error
		{
			name:  "negative: managed identity access rule update query returns error",
			input: &types.UpdateManagedIdentityAccessRuleInput{},
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
			input: &types.UpdateManagedIdentityAccessRuleInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateManagedIdentityAccessRulePayload{
					UpdateManagedIdentityAccessRule: graphqlUpdateManagedIdentityAccessRuleMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Managed identity access rule with ID fe2eb564-6311-52ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems updating managed identity access rule: Managed identity access rule with ID fe2eb564-6311-52ae-901f-9195125ca92a not found",
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
			actualAccessRule, actualError := client.ManagedIdentity.UpdateManagedIdentityAccessRule(ctx, test.input)

			checkError(t, test.expectErrorMessage, actualError)
			checkAccessRule(t, test.expectAccessRule, actualAccessRule)
		})
	}
}

func TestDeleteManagedIdentityAccessRule(t *testing.T) {
	accessRuleID := "access-rule-id-1"

	// In GraphiQL, an 'accessRule' element appeared here.  However, it would not unmarshal when run from a test.
	type graphqlDeleteManagedIdentityAccessRuleMutation struct {
		// AccessRule graphQLManagedIdentityAccessRule `json:"accessRule"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteManagedIdentityAccessRulePayload struct {
		DeleteManagedIdentityAccessRule graphqlDeleteManagedIdentityAccessRuleMutation `json:"deleteManagedIdentityAccessRule"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		input              *types.DeleteManagedIdentityAccessRuleInput
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully deleted managed identity access rule",
			input: &types.DeleteManagedIdentityAccessRuleInput{
				ID: accessRuleID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteManagedIdentityAccessRulePayload{
					DeleteManagedIdentityAccessRule: graphqlDeleteManagedIdentityAccessRuleMutation{},
				},
			},
		},

		// negative: mutation returns error
		{
			name:  "negative: managed identity access rule delete mutation returns error",
			input: &types.DeleteManagedIdentityAccessRuleInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02)",
						Path: []string{
							"deleteManagedIdentityAccessRule",
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
			input: &types.DeleteManagedIdentityAccessRuleInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteManagedIdentityAccessRulePayload{
					DeleteManagedIdentityAccessRule: graphqlDeleteManagedIdentityAccessRuleMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Managed identity access rule with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems deleting managed identity access rule: Managed identity access rule with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
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
			actualError := client.ManagedIdentity.DeleteManagedIdentityAccessRule(ctx, test.input)

			checkError(t, test.expectErrorMessage, actualError)
		})
	}

}

// Utility functions:

func checkIdentity(t *testing.T, expectIdentity, actualIdentity *types.ManagedIdentity) {
	if expectIdentity != nil {
		require.NotNil(t, actualIdentity)
		assert.Equal(t, expectIdentity, actualIdentity)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.ManagedIdentity)(nil)
		assert.Equal(t, (*types.ManagedIdentity)(nil), actualIdentity)
	}
}

func checkAccessRule(t *testing.T, expectRule, actualRule *types.ManagedIdentityAccessRule) {
	if expectRule != nil {
		require.NotNil(t, actualRule)
		assert.Equal(t, expectRule, actualRule)
	} else {
		assert.Equal(t, (*types.ManagedIdentityAccessRule)(nil), actualRule)
	}
}

// The End.
