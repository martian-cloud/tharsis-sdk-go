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

// TODO: This module has unit tests only for newer managed identity (access rule) methods
// added around November, 2022.  TODO: The other methods should also have unit tests added.

func TestGetManagedIdentity(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	managedIdentityID := "managed-identity-id-1"
	managedIdentityVersion := "managed-identity-version-1"
	managedIdentityPath := "gp01/nm01"

	type graphqlManagedIdentityPayloadByID struct {
		Node *GraphQLManagedIdentity `json:"node"`
	}

	type graphqlManagedIdentityPayloadByPath struct {
		ManagedIdentity *GraphQLManagedIdentity `json:"managedIdentity"`
	}

	// test cases
	type testCase struct {
		responsePayload       interface{}
		expectManagedIdentity *types.ManagedIdentity
		name                  string
		input                 *types.GetManagedIdentityInput
		expectErrorCode       types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return managed identity by ID",
			input: &types.GetManagedIdentityInput{
				ID: &managedIdentityID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByID{
					Node: &GraphQLManagedIdentity{
						ID: graphql.String(managedIdentityID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(managedIdentityVersion),
						},
						Type:         "t01",
						GroupPath:    "gp01",
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
				GroupPath:    "gp01",
				ResourcePath: "rp01",
				Name:         "nm01",
				Description:  "de01",
				Data:         "da01",
				CreatedBy:    "cr01",
			},
		},
		{
			name: "Successfully return managed identity by resource path",
			input: &types.GetManagedIdentityInput{
				Path: &managedIdentityPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByPath{
					ManagedIdentity: &GraphQLManagedIdentity{
						ID: graphql.String(managedIdentityID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(managedIdentityVersion),
						},
						Type:         "t01",
						GroupPath:    "gp01",
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
				GroupPath:    "gp01",
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
			input: &types.GetManagedIdentityInput{
				ID: ptr.String("invalid-ID"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByID{},
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
			expectErrorCode: types.ErrInternal,
		},

		// query returns nil managed identity, as if the specified managed identity does not exist.
		{
			name: "query returns nil managed identity, as if the specified managed identity does not exist",
			input: &types.GetManagedIdentityInput{
				ID: ptr.String("not-found-ID"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByID{},
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualIdentity, actualError := client.ManagedIdentity.GetManagedIdentity(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkIdentity(t, test.expectManagedIdentity, actualIdentity)
		})
	}
}

func TestGetManagedIdentityAccessRules(t *testing.T) {
	managedIdentityID := "managed-identity-id-1"
	managedIdentityPath := "gp01/nm01"

	// graphqlManagedIdentityWithAccessRules has access rules but no other fields.
	// The regular GraphQLManagedIdentity does not have access rules, because it would be a circular reference.
	type graphqlManagedIdentityWithAccessRules struct {
		AccessRules []graphQLManagedIdentityAccessRule `json:"accessRules"`
	}

	type graphqlManagedIdentityPayloadByID struct {
		Node *graphqlManagedIdentityWithAccessRules `json:"node"`
	}

	type graphqlManagedIdentityPayloadByPath struct {
		ManagedIdentity *struct {
			AccessRules []graphQLManagedIdentityAccessRule `json:"accessRules"`
		} `json:"managedIdentity"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.GetManagedIdentityInput
		name            string
		expectErrorCode types.ErrorCode
		expectRules     []types.ManagedIdentityAccessRule
	}

	testCases := []testCase{
		// positive
		{
			name: "Successfully return managed identity rules by managed identity ID",
			input: &types.GetManagedIdentityInput{
				ID: &managedIdentityID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: &graphqlManagedIdentityPayloadByID{
					Node: &graphqlManagedIdentityWithAccessRules{
						AccessRules: []graphQLManagedIdentityAccessRule{
							{
								ID: "ar01",
							},
							{
								ID: "ar02",
							},
						},
					},
				},
			},
			expectRules: []types.ManagedIdentityAccessRule{
				{
					Metadata: types.ResourceMetadata{
						ID: "ar01",
					},
					AllowedUsers:              []types.User{},
					AllowedServiceAccounts:    []types.ServiceAccount{},
					AllowedTeams:              []types.Team{},
					ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				},
				{
					Metadata: types.ResourceMetadata{
						ID: "ar02",
					},
					AllowedUsers:              []types.User{},
					AllowedServiceAccounts:    []types.ServiceAccount{},
					AllowedTeams:              []types.Team{},
					ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				},
			},
		},
		{
			name: "Successfully return managed identity access rules by managed identity resource path",
			input: &types.GetManagedIdentityInput{
				Path: &managedIdentityPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByPath{
					ManagedIdentity: &struct {
						AccessRules []graphQLManagedIdentityAccessRule `json:"accessRules"`
					}{
						AccessRules: []graphQLManagedIdentityAccessRule{
							{
								ID: "ar01",
							},
							{
								ID: "ar02",
							},
						},
					},
				},
			},
			expectRules: []types.ManagedIdentityAccessRule{
				{
					Metadata: types.ResourceMetadata{
						ID: "ar01",
					},
					AllowedUsers:              []types.User{},
					AllowedServiceAccounts:    []types.ServiceAccount{},
					AllowedTeams:              []types.Team{},
					ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				},
				{
					Metadata: types.ResourceMetadata{
						ID: "ar02",
					},
					AllowedUsers:              []types.User{},
					AllowedServiceAccounts:    []types.ServiceAccount{},
					AllowedTeams:              []types.Team{},
					ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				},
			},
		},

		// query returns error as if the ID is invalid
		{
			name: "query returns error as if the ID is invalid",
			input: &types.GetManagedIdentityInput{
				ID: ptr.String("invalid-ID"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByID{},
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
			expectErrorCode: types.ErrInternal,
		},

		// query returns nil managed identity, as if the specified managed identity does not exist.
		{
			name: "query returns nil managed identity, as if the specified managed identity does not exist",
			input: &types.GetManagedIdentityInput{
				ID: ptr.String("not-found-ID"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityPayloadByID{},
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualRules, actualError := client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			assert.ElementsMatch(t, test.expectRules, actualRules)
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
		responsePayload  interface{}
		input            *types.CreateManagedIdentityAccessRuleInput
		expectAccessRule *types.ManagedIdentityAccessRule
		name             string
		expectErrorCode  types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully created managed identity access rule",
			input: &types.CreateManagedIdentityAccessRuleInput{
				Type:                   types.ManagedIdentityAccessRuleEligiblePrincipals,
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
							Type:     graphql.String(types.ManagedIdentityAccessRuleEligiblePrincipals),
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
				Type:     types.ManagedIdentityAccessRuleEligiblePrincipals,
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
				ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				ManagedIdentityID:         managedIdentityID,
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
				}),
			}
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualAccessRule, actualError := client.ManagedIdentity.CreateManagedIdentityAccessRule(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
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
		responsePayload  interface{}
		expectAccessRule *types.ManagedIdentityAccessRule
		name             string
		expectErrorCode  types.ErrorCode
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
						Type:                      graphql.String(types.ManagedIdentityAccessRuleEligiblePrincipals),
						RunStage:                  graphql.String(types.JobPlanType),
						AllowedUsers:              []graphQLUser{},
						AllowedServiceAccounts:    []graphQLServiceAccount{},
						AllowedTeams:              []graphQLTeam{},
						ModuleAttestationPolicies: []graphQLAccessRuleModuleAttestationPolicy{},
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
				Type:                      types.ManagedIdentityAccessRuleEligiblePrincipals,
				RunStage:                  types.JobPlanType,
				AllowedUsers:              []types.User{},
				AllowedServiceAccounts:    []types.ServiceAccount{},
				AllowedTeams:              []types.Team{},
				ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				ManagedIdentityID:         managedIdentityID,
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
			expectErrorCode: types.ErrBadRequest,
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
			expectErrorCode: types.ErrNotFound,
		},

		// negative: theoretical quiet not found
		{
			name: "negative: theoretical quiet not found",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlManagedIdentityAccessRulePayload{},
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualAccessRule, actualError := client.ManagedIdentity.GetManagedIdentityAccessRule(
				ctx,
				&types.GetManagedIdentityAccessRuleInput{ID: accessRuleID},
			)

			checkError(t, test.expectErrorCode, actualError)
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
		responsePayload  interface{}
		input            *types.UpdateManagedIdentityAccessRuleInput
		expectAccessRule *types.ManagedIdentityAccessRule
		name             string
		expectErrorCode  types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully updated managed identity access rule",
			input: &types.UpdateManagedIdentityAccessRuleInput{
				ID:                     accessRuleID,
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
							Type:     graphql.String(types.ManagedIdentityAccessRuleEligiblePrincipals),
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
				Type:     types.ManagedIdentityAccessRuleEligiblePrincipals,
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
				ModuleAttestationPolicies: []types.ManagedIdentityAccessRuleModuleAttestationPolicy{},
				ManagedIdentityID:         managedIdentityID,
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
			expectErrorCode: types.ErrInternal,
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualAccessRule, actualError := client.ManagedIdentity.UpdateManagedIdentityAccessRule(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkAccessRule(t, test.expectAccessRule, actualAccessRule)
		})
	}
}

func TestDeleteManagedIdentityAccessRule(t *testing.T) {
	accessRuleID := "access-rule-id-1"

	type graphqlDeleteManagedIdentityAccessRuleMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteManagedIdentityAccessRulePayload struct {
		DeleteManagedIdentityAccessRule graphqlDeleteManagedIdentityAccessRuleMutation `json:"deleteManagedIdentityAccessRule"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.DeleteManagedIdentityAccessRuleInput
		name            string
		expectErrorCode types.ErrorCode
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
			expectErrorCode: types.ErrInternal,
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
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualError := client.ManagedIdentity.DeleteManagedIdentityAccessRule(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}

}

func TestCreateManagedIdentityAlias(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	aliasID := "test-alias-1"
	aliasVersion := "alias-version-1"
	aliasSourceID := "test-managed-identity-1"

	type graphqlCreateManagedIdentityAliasMutation struct {
		ManagedIdentity GraphQLManagedIdentity       `json:"managedIdentity"`
		Problems        []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateManagedIdentityAliasPayload struct {
		CreateManagedIdentityAlias graphqlCreateManagedIdentityAliasMutation `json:"createManagedIdentityAlias"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.CreateManagedIdentityAliasInput
		expectAlias     *types.ManagedIdentity
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "successfully created managed identity alias",
			input: &types.CreateManagedIdentityAliasInput{
				Name:          "test-alias",
				AliasSourceID: &aliasSourceID,
				GroupPath:     "test/group",
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateManagedIdentityAliasPayload{
					CreateManagedIdentityAlias: graphqlCreateManagedIdentityAliasMutation{
						ManagedIdentity: GraphQLManagedIdentity{
							ID: graphql.String(aliasID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(aliasVersion),
							},
							Type:          graphql.String(types.ManagedIdentityAWSFederated),
							GroupPath:     graphql.String("alias-group-path"),
							ResourcePath:  graphql.String("alias-resource-path"),
							Name:          graphql.String("test-alias"),
							Description:   graphql.String("some-description"),
							Data:          graphql.String("some-data"),
							CreatedBy:     graphql.String("some-creator"),
							AliasSourceID: graphql.NewString(graphql.String(aliasSourceID)),
							IsAlias:       graphql.Boolean(true),
						},
					},
				},
			},
			expectAlias: &types.ManagedIdentity{
				Metadata: types.ResourceMetadata{
					ID:                   aliasID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              aliasVersion,
				},
				Type:          types.ManagedIdentityAWSFederated,
				GroupPath:     "alias-group-path",
				ResourcePath:  "alias-resource-path",
				Name:          "test-alias",
				Description:   "some-description",
				Data:          "some-data",
				CreatedBy:     "some-creator",
				AliasSourceID: &aliasSourceID,
				IsAlias:       true,
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
				graphqlClient: newGraphQLClientForTest(testClientInput{
					statusToReturn:  http.StatusOK,
					payloadToReturn: string(payload),
				}),
			}
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualAlias, actualError := client.ManagedIdentity.CreateManagedIdentityAlias(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkIdentity(t, test.expectAlias, actualAlias)
		})
	}
}

func TestDeleteManagedIdentityAlias(t *testing.T) {
	aliasID := "test-alias-1"

	type graphqlDeleteManagedIdentityAliasMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteManagedIdentityAliasPayload struct {
		DeleteManagedIdentityAlias graphqlDeleteManagedIdentityAliasMutation `json:"deleteManagedIdentityAlias"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.DeleteManagedIdentityAliasInput
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully deleted managed identity alias",
			input: &types.DeleteManagedIdentityAliasInput{
				ID: aliasID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteManagedIdentityAliasPayload{
					DeleteManagedIdentityAlias: graphqlDeleteManagedIdentityAliasMutation{},
				},
			},
		},

		// negative: mutation returns error
		{
			name:  "negative: managed identity alias delete mutation returns error",
			input: &types.DeleteManagedIdentityAliasInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02)",
						Path: []string{
							"deleteManagedIdentityAlias",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: types.ErrInternal,
		},

		// negative: mutation behaves as if the specified alias did not exist
		{
			name:  "negative: mutation behaves as if the specified alias did not exist",
			input: &types.DeleteManagedIdentityAliasInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteManagedIdentityAliasPayload{
					DeleteManagedIdentityAlias: graphqlDeleteManagedIdentityAliasMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Managed identity with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
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
				}),
			}
			client.ManagedIdentity = NewManagedIdentity(client)

			// Call the method being tested.
			actualError := client.ManagedIdentity.DeleteManagedIdentityAlias(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
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
