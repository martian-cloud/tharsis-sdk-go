package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestCreateFederatedRegistryTokens(t *testing.T) {
	testHostname := "test.remote.registry.example"
	testTokenString := "test-token-string"

	type graphqlCreateFederatedRegistryTokensMutation struct {
		Tokens   []graphQLFederatedRegistryToken `json:"tokens"`
		Problems []fakeGraphqlResponseProblem    `json:"problems"`
	}

	type graphqlCreateFederatedRegistryTokensPayload struct {
		CreateFederatedRegistryTokens graphqlCreateFederatedRegistryTokensMutation `json:"createFederatedRegistryTokens"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		expectTokens    []types.FederatedRegistryToken
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully created federated registry token(s)",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateFederatedRegistryTokensPayload{
					CreateFederatedRegistryTokens: graphqlCreateFederatedRegistryTokensMutation{
						Tokens: []graphQLFederatedRegistryToken{
							{
								Hostname: graphql.String(testHostname),
								Token:    graphql.String(testTokenString),
							},
						},
					},
				},
			},
			expectTokens: []types.FederatedRegistryToken{
				{
					Hostname: testHostname,
					Token:    testTokenString,
				},
			},
		},
		{
			name: "create federated registry token returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateFederatedRegistryTokensPayload{
					CreateFederatedRegistryTokens: graphqlCreateFederatedRegistryTokensMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "failed to create federated registry tokens",
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
			client.FederatedRegistry = NewFederatedRegistry(client)

			// Call the method being tested.
			actualTokens, actualError := client.FederatedRegistry.CreateFederatedRegistryTokens(ctx,
				&types.CreateFederatedRegistryTokensInput{},
			)

			checkError(t, test.expectErrorCode, actualError)

			if test.expectTokens != nil {
				require.NotNil(t, actualTokens)
				assert.ElementsMatch(t, test.expectTokens, actualTokens)
			}
		})
	}
}
