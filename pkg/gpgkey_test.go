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

func TestGetGPGKeyByID(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	keyID := "gpg-key-id-1"
	keyVersion := "gpg-key-version-1"
	keyCreatedBy := "gpg-key-created-by"
	keyASCIIArmor := "gpg-key-ascii-armor"
	keyFingerprint := "gpg-key-fingerprint"
	keyGPGKeyID := "fedcba9876543210"

	type graphqlGPGKeyPayloadByID struct {
		Node *graphQLGPGKey `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.GetGPGKeyInput
		expectGPGKey    *types.GPGKey
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return GPG key by ID",
			input: &types.GetGPGKeyInput{
				ID: keyID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGPGKeyPayloadByID{
					Node: &graphQLGPGKey{
						ID: graphql.String(keyID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(keyVersion),
						},
						CreatedBy:   graphql.String(keyCreatedBy),
						ASCIIArmor:  graphql.String(keyASCIIArmor),
						Fingerprint: graphql.String(keyFingerprint),
						GPGKeyID:    graphql.String(keyGPGKeyID),
					},
				},
			},
			expectGPGKey: &types.GPGKey{
				Metadata: types.ResourceMetadata{
					ID:                   keyID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              keyVersion,
				},
				CreatedBy:   keyCreatedBy,
				ASCIIArmor:  keyASCIIArmor,
				Fingerprint: keyFingerprint,
				GPGKeyID:    keyGPGKeyID,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetGPGKeyInput{
				ID: keyID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGPGKeyPayloadByID{},
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
		{
			name: "query returns nil GPG key, as if the specified GPG key does not exist",
			input: &types.GetGPGKeyInput{
				ID: keyID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGPGKeyPayloadByID{},
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
			client.GPGKey = NewGPGKey(client)

			// Call the method being tested.
			actualGPGKey, actualError := client.GPGKey.GetGPGKey(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkGPGKey(t, test.expectGPGKey, actualGPGKey)
		})
	}
}

func TestCreateGPGKeyByID(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	keyID := "gpg-key-id-1"
	keyVersion := "gpg-key-version-1"
	keyCreatedBy := "gpg-key-created-by"
	keyGroupPath := "gpg-key-group-path"
	keyASCIIArmor := "gpg-key-ascii-armor"
	keyFingerprint := "gpg-key-fingerprint"
	keyGPGKeyID := "fedcba9876543210"

	type graphqlCreateGPGKeyMutation struct {
		GPGKey   graphQLGPGKey                `json:"gpgKey"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateGPGKeyPayload struct {
		CreateGPGKey graphqlCreateGPGKeyMutation `json:"createGPGKey"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.CreateGPGKeyInput
		expectGPGKey    *types.GPGKey
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully create GPG key",
			input: &types.CreateGPGKeyInput{
				ASCIIArmor: keyASCIIArmor,
				GroupPath:  keyGroupPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateGPGKeyPayload{
					CreateGPGKey: graphqlCreateGPGKeyMutation{
						GPGKey: graphQLGPGKey{
							ID: graphql.String(keyID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(keyVersion),
							},
							CreatedBy:   graphql.String(keyCreatedBy),
							ASCIIArmor:  graphql.String(keyASCIIArmor),
							Fingerprint: graphql.String(keyFingerprint),
							GPGKeyID:    graphql.String(keyGPGKeyID),
						},
					},
				},
			},
			expectGPGKey: &types.GPGKey{
				Metadata: types.ResourceMetadata{
					ID:                   keyID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              keyVersion,
				},
				CreatedBy:   keyCreatedBy,
				ASCIIArmor:  keyASCIIArmor,
				Fingerprint: keyFingerprint,
				GPGKeyID:    keyGPGKeyID,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.CreateGPGKeyInput{
				ASCIIArmor: keyASCIIArmor,
				GroupPath:  keyGroupPath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateGPGKeyPayload{},
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
			client.GPGKey = NewGPGKey(client)

			// Call the method being tested.
			actualGPGKey, actualError := client.GPGKey.CreateGPGKey(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkGPGKey(t, test.expectGPGKey, actualGPGKey)
		})
	}
}

func TestDeleteGPGKeyByID(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	keyID := "gpg-key-id-1"
	keyVersion := "gpg-key-version-1"
	keyCreatedBy := "gpg-key-created-by"
	keyASCIIArmor := "gpg-key-ascii-armor"
	keyFingerprint := "gpg-key-fingerprint"
	keyGPGKeyID := "fedcba9876543210"

	type graphqlDeleteGPGKeyMutation struct {
		GPGKey   graphQLGPGKey                `json:"gpgKey"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteGPGKeyPayload struct {
		DeleteGPGKey graphqlDeleteGPGKeyMutation `json:"deleteGPGKey"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.DeleteGPGKeyInput
		expectGPGKey    *types.GPGKey
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully delete GPG key",
			input: &types.DeleteGPGKeyInput{
				ID: keyID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteGPGKeyPayload{
					DeleteGPGKey: graphqlDeleteGPGKeyMutation{
						GPGKey: graphQLGPGKey{
							ID: graphql.String(keyID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(keyVersion),
							},
							CreatedBy:   graphql.String(keyCreatedBy),
							ASCIIArmor:  graphql.String(keyASCIIArmor),
							Fingerprint: graphql.String(keyFingerprint),
							GPGKeyID:    graphql.String(keyGPGKeyID),
						},
					},
				},
			},
			expectGPGKey: &types.GPGKey{
				Metadata: types.ResourceMetadata{
					ID:                   keyID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              keyVersion,
				},
				CreatedBy:   keyCreatedBy,
				ASCIIArmor:  keyASCIIArmor,
				Fingerprint: keyFingerprint,
				GPGKeyID:    keyGPGKeyID,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.DeleteGPGKeyInput{
				ID: keyID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteGPGKeyPayload{},
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
		{
			name: "query returns nil GPG key, as if the specified GPG key does not exist",
			input: &types.DeleteGPGKeyInput{
				ID: keyID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteGPGKeyPayload{
					DeleteGPGKey: graphqlDeleteGPGKeyMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Type:    "NOT_FOUND",
								Message: "GPG key with ID something not found",
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
			client.GPGKey = NewGPGKey(client)

			// Call the method being tested.
			actualGPGKey, actualError := client.GPGKey.DeleteGPGKey(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkGPGKey(t, test.expectGPGKey, actualGPGKey)
		})
	}

}

// Utility functions:

func checkGPGKey(t *testing.T, expectGPGKey, actualGPGKey *types.GPGKey) {
	if expectGPGKey != nil {
		require.NotNil(t, actualGPGKey)
		assert.Equal(t, expectGPGKey, actualGPGKey)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.GPGKey)(nil)
		assert.Equal(t, (*types.GPGKey)(nil), actualGPGKey)
	}
}

// The End.
