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

func TestCreateServiceAccount(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	serviceAccountName := "service-account-name-1"
	serviceAccountID := "service-account-id-1"
	serviceAccountVersion := "service-account-version-1"
	serviceAccountDescription := "service account description 1"
	groupPath := parentGroupName
	resourcePath := parentGroupName + "/" + serviceAccountName
	trustPolicyIssuer1 := "https://trust-policy-issuer-1"
	boundClaimName1a := "bound-claim-name-1a"
	boundClaimValue1a := "bound-claim-value-1a"
	boundClaimName1b := "bound-claim-name-1b"
	boundClaimValue1b := "bound-claim-value-1b"
	trustPolicyIssuer2 := "https://trust-policy-issuer-2"
	boundClaimName2a := "bound-claim-name-2a"
	boundClaimValue2a := "bound-claim-value-2a"
	boundClaimName2b := "bound-claim-name-2b"
	boundClaimValue2b := "bound-claim-value-2b"

	type graphqlCreateServiceAccountMutation struct {
		ServiceAccount graphQLServiceAccount        `json:"serviceAccount"`
		Problems       []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateServiceAccountPayload struct {
		CreateServiceAccount graphqlCreateServiceAccountMutation `json:"createServiceAccount"`
	}

	// test cases
	type testCase struct {
		responsePayload      interface{}
		input                *types.CreateServiceAccountInput
		expectServiceAccount *types.ServiceAccount
		name                 string
		expectErrorCode      ErrorCode
	}

	testCases := []testCase{

		// positive, nested
		{
			name: "Successfully created service account",
			input: &types.CreateServiceAccountInput{
				Name:        serviceAccountName,
				GroupPath:   parentGroupName,
				Description: serviceAccountDescription,
				OIDCTrustPolicies: []types.OIDCTrustPolicy{
					{
						Issuer: trustPolicyIssuer1,
						BoundClaims: map[string]string{
							boundClaimName1a: boundClaimValue1a,
							boundClaimName1b: boundClaimValue1b,
						},
					},
					{
						Issuer: trustPolicyIssuer2,
						BoundClaims: map[string]string{
							boundClaimName2a: boundClaimValue2a,
							boundClaimName2b: boundClaimValue2b,
						},
					},
				},
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateServiceAccountPayload{
					CreateServiceAccount: graphqlCreateServiceAccountMutation{
						ServiceAccount: graphQLServiceAccount{
							ID: graphql.String(serviceAccountID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(serviceAccountVersion),
							},
							GroupPath:    graphql.String(groupPath),
							ResourcePath: graphql.String(resourcePath),
							Name:         graphql.String(serviceAccountName),
							Description:  graphql.String(serviceAccountDescription),
							OIDCTrustPolicies: []graphQLTrustPolicy{
								{
									Issuer: graphql.String(trustPolicyIssuer1),
									BoundClaims: []graphQLBoundClaim{
										{
											Name:  graphql.String(boundClaimName1a),
											Value: graphql.String(boundClaimValue1a),
										},
										{
											Name:  graphql.String(boundClaimName1b),
											Value: graphql.String(boundClaimValue1b),
										},
									},
								},
								{
									Issuer: graphql.String(trustPolicyIssuer2),
									BoundClaims: []graphQLBoundClaim{
										{
											Name:  graphql.String(boundClaimName2a),
											Value: graphql.String(boundClaimValue2a),
										},
										{
											Name:  graphql.String(boundClaimName2b),
											Value: graphql.String(boundClaimValue2b),
										},
									},
								},
							},
						},
					},
				},
			},
			expectServiceAccount: &types.ServiceAccount{
				Metadata: types.ResourceMetadata{
					ID:                   serviceAccountID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              serviceAccountVersion,
				},
				GroupPath:    groupPath,
				ResourcePath: resourcePath,
				Name:         serviceAccountName,
				Description:  serviceAccountDescription,
				OIDCTrustPolicies: []types.OIDCTrustPolicy{
					{
						Issuer: trustPolicyIssuer1,
						BoundClaims: map[string]string{
							boundClaimName1a: boundClaimValue1a,
							boundClaimName1b: boundClaimValue1b,
						},
					},
					{
						Issuer: trustPolicyIssuer2,
						BoundClaims: map[string]string{
							boundClaimName2a: boundClaimValue2a,
							boundClaimName2b: boundClaimValue2b,
						},
					},
				},
			},
		},

		// negative: query returns error
		{
			name:  "negative: query to create service account returned error",
			input: &types.CreateServiceAccountInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateServiceAccountPayload{
					CreateServiceAccount: graphqlCreateServiceAccountMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "service account with path non-existent not found",
								Type:    internal.Conflict,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorCode: ErrConflict,
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
			client.ServiceAccount = NewServiceAccount(client)

			// Call the method being tested.
			actualServiceAccount, actualError := client.ServiceAccount.CreateServiceAccount(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkServiceAccount(t, test.expectServiceAccount, actualServiceAccount)
		})
	}
}

func TestGetServiceAccount(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	serviceAccountName := "service-account-name-1"
	serviceAccountDescription := "service account description 1"
	groupPath := parentGroupName
	resourcePath := parentGroupName + "/" + serviceAccountName
	serviceAccountID := "service-account-id-1"
	serviceAccountVersion := "service-account-version-1"
	trustPolicyIssuer1 := "https://trust-policy-issuer-1"
	boundClaimName1a := "bound-claim-name-1a"
	boundClaimValue1a := "bound-claim-value-1a"
	boundClaimName1b := "bound-claim-name-1b"
	boundClaimValue1b := "bound-claim-value-1b"
	trustPolicyIssuer2 := "https://trust-policy-issuer-2"
	boundClaimName2a := "bound-claim-name-2a"
	boundClaimValue2a := "bound-claim-value-2a"
	boundClaimName2b := "bound-claim-name-2b"
	boundClaimValue2b := "bound-claim-value-2b"

	type graphqlServiceAccountPayload struct {
		ServiceAccount *graphQLServiceAccount `json:"serviceAccount"`
	}

	// test cases
	type testCase struct {
		responsePayload      interface{}
		expectServiceAccount *types.ServiceAccount
		name                 string
		expectErrorCode      ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return service account by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlServiceAccountPayload{
					ServiceAccount: &graphQLServiceAccount{
						ID: graphql.String(serviceAccountID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(serviceAccountVersion),
						},
						GroupPath:    graphql.String(groupPath),
						ResourcePath: graphql.String(resourcePath),
						Name:         graphql.String(serviceAccountName),
						Description:  graphql.String(serviceAccountDescription),
						OIDCTrustPolicies: []graphQLTrustPolicy{
							{
								Issuer: graphql.String(trustPolicyIssuer1),
								BoundClaims: []graphQLBoundClaim{
									{
										Name:  graphql.String(boundClaimName1a),
										Value: graphql.String(boundClaimValue1a),
									},
									{
										Name:  graphql.String(boundClaimName1b),
										Value: graphql.String(boundClaimValue1b),
									},
								},
							},
							{
								Issuer: graphql.String(trustPolicyIssuer2),
								BoundClaims: []graphQLBoundClaim{
									{
										Name:  graphql.String(boundClaimName2a),
										Value: graphql.String(boundClaimValue2a),
									},
									{
										Name:  graphql.String(boundClaimName2b),
										Value: graphql.String(boundClaimValue2b),
									},
								},
							},
						},
					},
				},
			},
			expectServiceAccount: &types.ServiceAccount{
				Metadata: types.ResourceMetadata{
					ID:                   serviceAccountID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              serviceAccountVersion,
				},
				GroupPath:    groupPath,
				ResourcePath: resourcePath,
				Name:         serviceAccountName,
				Description:  serviceAccountDescription,
				OIDCTrustPolicies: []types.OIDCTrustPolicy{
					{
						Issuer: trustPolicyIssuer1,
						BoundClaims: map[string]string{
							boundClaimName1a: boundClaimValue1a,
							boundClaimName1b: boundClaimValue1b,
						},
					},
					{
						Issuer: trustPolicyIssuer2,
						BoundClaims: map[string]string{
							boundClaimName2a: boundClaimValue2a,
							boundClaimName2b: boundClaimValue2b,
						},
					},
				},
			},
		},

		// negative: query returns error, invalid ID--payload taken from GraphiQL
		{
			name: "negative: query returns error, invalid ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlServiceAccountPayload{},
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
			expectErrorCode: ErrInternal,
		},

		// negative: query returns error, not found error--payload taken from GraphiQL
		{
			name: "negative: query returns error, not found error",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlServiceAccountPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "service account with id 6f3106c6-c342-4790-a667-963a850d34d4 not found",
						Path: []string{
							"node",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "NOT_FOUND",
						},
					},
				},
			},
			expectErrorCode: ErrNotFound,
		},

		// negative: theoretical quiet not found
		{
			name: "negative: theoretical quiet not found",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlServiceAccountPayload{},
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
			client.ServiceAccount = NewServiceAccount(client)

			// Call the method being tested.
			actualServiceAccount, actualError := client.ServiceAccount.GetServiceAccount(
				ctx,
				&types.GetServiceAccountInput{ID: serviceAccountID},
			)

			checkError(t, test.expectErrorCode, actualError)
			checkServiceAccount(t, test.expectServiceAccount, actualServiceAccount)
		})
	}
}

func TestUpdateServiceAccount(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	serviceAccountName := "service-account-name-1"
	serviceAccountDescription := "service account description 1"
	groupPath := parentGroupName
	resourcePath := parentGroupName + "/" + serviceAccountName
	serviceAccountID := "service-account-id-1"
	serviceAccountVersion := "service-account-version-1"

	trustPolicyIssuer1 := "https://trust-policy-issuer-1"
	boundClaimName1a := "bound-claim-name-1a"
	boundClaimValue1a := "bound-claim-value-1a"
	boundClaimName1b := "bound-claim-name-1b"
	boundClaimValue1b := "bound-claim-value-1b"
	trustPolicyIssuer2 := "https://trust-policy-issuer-2"
	boundClaimName2a := "bound-claim-name-2a"
	boundClaimValue2a := "bound-claim-value-2a"
	boundClaimName2b := "bound-claim-name-2b"
	boundClaimValue2b := "bound-claim-value-2b"

	type graphqlUpdateServiceAccountMutation struct {
		ServiceAccount graphQLServiceAccount        `json:"serviceAccount"`
		Problems       []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateServiceAccountPayload struct {
		UpdateServiceAccount graphqlUpdateServiceAccountMutation `json:"updateServiceAccount"`
	}

	// test cases
	type testCase struct {
		responsePayload      interface{}
		input                *types.UpdateServiceAccountInput
		expectServiceAccount *types.ServiceAccount
		name                 string
		expectErrorCode      ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully updated service account",
			input: &types.UpdateServiceAccountInput{
				ID:          serviceAccountID,
				Description: serviceAccountDescription,
				OIDCTrustPolicies: []types.OIDCTrustPolicy{
					{
						Issuer: trustPolicyIssuer1,
						BoundClaims: map[string]string{
							boundClaimName1a: boundClaimValue1a,
							boundClaimName1b: boundClaimValue1b,
						},
					},
					{
						Issuer: trustPolicyIssuer2,
						BoundClaims: map[string]string{
							boundClaimName2a: boundClaimValue2a,
							boundClaimName2b: boundClaimValue2b,
						},
					},
				},
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateServiceAccountPayload{
					UpdateServiceAccount: graphqlUpdateServiceAccountMutation{
						ServiceAccount: graphQLServiceAccount{
							ID: graphql.String(serviceAccountID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(serviceAccountVersion),
							},
							GroupPath:    graphql.String(groupPath),
							ResourcePath: graphql.String(resourcePath),
							Name:         graphql.String(serviceAccountName),
							Description:  graphql.String(serviceAccountDescription),
							OIDCTrustPolicies: []graphQLTrustPolicy{
								{
									Issuer: graphql.String(trustPolicyIssuer1),
									BoundClaims: []graphQLBoundClaim{
										{
											Name:  graphql.String(boundClaimName1a),
											Value: graphql.String(boundClaimValue1a),
										},
										{
											Name:  graphql.String(boundClaimName1b),
											Value: graphql.String(boundClaimValue1b),
										},
									},
								},
								{
									Issuer: graphql.String(trustPolicyIssuer2),
									BoundClaims: []graphQLBoundClaim{
										{
											Name:  graphql.String(boundClaimName2a),
											Value: graphql.String(boundClaimValue2a),
										},
										{
											Name:  graphql.String(boundClaimName2b),
											Value: graphql.String(boundClaimValue2b),
										},
									},
								},
							},
						},
					},
				},
			},
			expectServiceAccount: &types.ServiceAccount{
				Metadata: types.ResourceMetadata{
					ID:                   serviceAccountID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              serviceAccountVersion,
				},
				GroupPath:    groupPath,
				ResourcePath: resourcePath,
				Name:         serviceAccountName,
				Description:  serviceAccountDescription,
				OIDCTrustPolicies: []types.OIDCTrustPolicy{
					{
						Issuer: trustPolicyIssuer1,
						BoundClaims: map[string]string{
							boundClaimName1a: boundClaimValue1a,
							boundClaimName1b: boundClaimValue1b,
						},
					},
					{
						Issuer: trustPolicyIssuer2,
						BoundClaims: map[string]string{
							boundClaimName2a: boundClaimValue2a,
							boundClaimName2b: boundClaimValue2b,
						},
					},
				},
			},
		},

		// negative: service account update query returns error
		{
			name:  "negative: service account update query returns error",
			input: &types.UpdateServiceAccountInput{},
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
			expectErrorCode: ErrInternal,
		},

		// negative: query behaves as if the specified service account did not exist
		{
			name:  "negative: query behaves as if the specified service account did not exist",
			input: &types.UpdateServiceAccountInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateServiceAccountPayload{
					UpdateServiceAccount: graphqlUpdateServiceAccountMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "service account with ID fe2eb564-6311-52ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
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
			client.ServiceAccount = NewServiceAccount(client)

			// Call the method being tested.
			actualServiceAccount, actualError := client.ServiceAccount.UpdateServiceAccount(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkServiceAccount(t, test.expectServiceAccount, actualServiceAccount)
		})
	}
}

func TestDeleteServiceAccount(t *testing.T) {
	serviceAccountID := "service-account-id-1"

	type graphqlDeleteServiceAccountMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteServiceAccountPayload struct {
		DeleteServiceAccount graphqlDeleteServiceAccountMutation `json:"deleteServiceAccount"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.DeleteServiceAccountInput
		name            string
		expectErrorCode ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully deleted service account",
			input: &types.DeleteServiceAccountInput{
				ID: serviceAccountID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteServiceAccountPayload{
					DeleteServiceAccount: graphqlDeleteServiceAccountMutation{},
				},
			},
		},

		// negative: mutation returns error
		{
			name:  "negative: service account delete mutation returns error",
			input: &types.DeleteServiceAccountInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02)",
						Path: []string{
							"deleteServiceAccount",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
		},

		// negative: mutation behaves as if the specified service account did not exist
		{
			name:  "negative: mutation behaves as if the specified service account did not exist",
			input: &types.DeleteServiceAccountInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteServiceAccountPayload{
					DeleteServiceAccount: graphqlDeleteServiceAccountMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "service account with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
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
			client.ServiceAccount = NewServiceAccount(client)

			// Call the method being tested.
			actualError := client.ServiceAccount.DeleteServiceAccount(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}

}

func TestServiceAccountCreateToken(t *testing.T) {
	serviceAccountPath := "service/account/path"
	inputToken := "some-input-token"
	outputToken := "some-output-token"
	testExpiresIn := 25 * time.Second
	testExpiresInSeconds := int(testExpiresIn / time.Second)

	type graphqlServiceAccountCreateTokenMutation struct {
		Token     string                       `json:"token"`
		Problems  []fakeGraphqlResponseProblem `json:"problems"`
		ExpiresIn int                          `json:"expiresIn"`
	}

	type graphqlServiceAccountCreateTokenPayload struct {
		ServiceAccountCreateToken graphqlServiceAccountCreateTokenMutation `json:"serviceAccountCreateToken"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.ServiceAccountCreateTokenInput
		name            string
		expectToken     string
		expectErrorCode ErrorCode
		expectExpiresIn time.Duration
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully logged in to service account",
			input: &types.ServiceAccountCreateTokenInput{
				ServiceAccountPath: serviceAccountPath,
				Token:              inputToken,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlServiceAccountCreateTokenPayload{
					ServiceAccountCreateToken: graphqlServiceAccountCreateTokenMutation{
						Token:     outputToken,
						ExpiresIn: testExpiresInSeconds,
					},
				},
			},
			expectToken:     outputToken,
			expectExpiresIn: testExpiresIn,
		},

		// negative: mutation returns error
		{
			name:  "negative: service account token creation mutation returns error",
			input: &types.ServiceAccountCreateTokenInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: JWT is missing issuer claim",
						Path: []string{
							"serviceAccountCreateToken",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "UNAUTHENTICATED", // could also use UNAUTHORIZED
						},
					},
				},
			},
			expectErrorCode: ErrUnauthorized,
		},

		// negative: mutation behaves as if the specified service account did not exist
		{
			name:  "negative: mutation behaves as if the specified service account did not exist",
			input: &types.ServiceAccountCreateTokenInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlServiceAccountCreateTokenPayload{
					ServiceAccountCreateToken: graphqlServiceAccountCreateTokenMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "service account with path does/not/exist not found",
								Type:    "NOT_FOUND",
								Field:   []string{},
							},
						},
					},
				},
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
			client.ServiceAccount = NewServiceAccount(client)

			// Call the method being tested.
			actualResponse, actualError := client.ServiceAccount.CreateToken(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			if test.expectToken != "" {
				assert.Equal(t, test.expectToken, actualResponse.Token)
				assert.Equal(t, test.expectExpiresIn, actualResponse.ExpiresIn)
			}
		})
	}
}

// Utility functions:

func checkServiceAccount(t *testing.T, expectServiceAccount, actualServiceAccount *types.ServiceAccount) {
	if expectServiceAccount != nil {
		require.NotNil(t, actualServiceAccount)
		assert.Equal(t, expectServiceAccount, actualServiceAccount)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.ServiceAccount)(nil)
		assert.Equal(t, (*types.ServiceAccount)(nil), actualServiceAccount)
	}
}

// The End.
