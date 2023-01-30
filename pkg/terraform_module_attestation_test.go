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

func TestCreateModuleAttestation(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleAttestationID := "1"

	type graphqlCreateTerraformModuleAttestationMutation struct {
		ModuleAttestation *graphQLTerraformModuleAttestation `json:"moduleAttestation"`
		Problems          []fakeGraphqlResponseProblem       `json:"problems"`
	}

	type graphqlCreateModuleAttestationPayload struct {
		CreateTerraformModuleAttestation graphqlCreateTerraformModuleAttestationMutation `json:"createTerraformModuleAttestation"`
	}

	// test cases
	type testCase struct {
		responsePayload   interface{}
		expectAttestation *types.TerraformModuleAttestation
		name              string
		expectErrorCode   ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully created terraform module attestation",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateModuleAttestationPayload{
					CreateTerraformModuleAttestation: graphqlCreateTerraformModuleAttestationMutation{
						ModuleAttestation: &graphQLTerraformModuleAttestation{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   "1",
							},
							ID:            graphql.String(moduleAttestationID),
							Description:   "some-description",
							SchemaType:    "https://in-toto.io/Statement/v0.1",
							PredicateType: "cosign.sigstore.dev/attestation/v1",
							Data:          "some-attestation-data",
							Digests:       []string{"some-digest"},
							Module: graphQLTerraformModule{
								ID: "module-1",
							},
						},
					},
				},
			},
			expectAttestation: &types.TerraformModuleAttestation{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleAttestationID,
					Version:              "1",
				},
				Description:   "some-description",
				SchemaType:    "https://in-toto.io/Statement/v0.1",
				PredicateType: "cosign.sigstore.dev/attestation/v1",
				Data:          "some-attestation-data",
				Digests:       []string{"some-digest"},
				ModuleID:      "module-1",
			},
		},
		{
			name: "create module attestation returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateModuleAttestationPayload{
					CreateTerraformModuleAttestation: graphqlCreateTerraformModuleAttestationMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "module attestation already exists",
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
				})}
			client.TerraformModuleAttestation = NewTerraformModuleAttestation(client)

			// Call the method being tested.
			actualAttestation, actualError := client.TerraformModuleAttestation.CreateModuleAttestation(ctx, &types.CreateTerraformModuleAttestationInput{})

			checkError(t, test.expectErrorCode, actualError)
			checkModuleAttestations(t, test.expectAttestation, actualAttestation)
		})
	}
}

func TestUpdateModuleAttestation(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleAttestationID := "module-attestation-id-1"
	moduleAttestationVersion := "module-attestation-1"
	moduleID := "managed-identity-id-1"

	type graphqlUpdateTerraformModuleAttestationMutation struct {
		ModuleAttestation *graphQLTerraformModuleAttestation `json:"moduleAttestation"`
		Problems          []fakeGraphqlResponseProblem       `json:"problems"`
	}

	type graphqlUpdateTerraformModuleAttestationPayload struct {
		UpdateTerraformModuleAttestation graphqlUpdateTerraformModuleAttestationMutation `json:"updateTerraformModuleAttestation"`
	}

	// test cases
	type testCase struct {
		responsePayload   interface{}
		input             *types.UpdateTerraformModuleAttestationInput
		expectAttestation *types.TerraformModuleAttestation
		name              string
		expectErrorCode   ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully updated terraform module attestation",
			input: &types.UpdateTerraformModuleAttestationInput{
				ID:          moduleAttestationID,
				Description: "a new updated description",
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateTerraformModuleAttestationPayload{
					UpdateTerraformModuleAttestation: graphqlUpdateTerraformModuleAttestationMutation{
						ModuleAttestation: &graphQLTerraformModuleAttestation{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(moduleAttestationVersion),
							},
							ID:            graphql.String(moduleAttestationID),
							Description:   "a new updated description",
							SchemaType:    "https://in-toto.io/Statement/v0.1",
							PredicateType: "cosign.sigstore.dev/attestation/v1",
							Data:          "some-attestation-data",
							Digests:       []string{"some-digest"},
							Module: graphQLTerraformModule{
								ID: graphql.String(moduleID),
							},
						},
					},
				},
			},
			expectAttestation: &types.TerraformModuleAttestation{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleAttestationID,
					Version:              moduleAttestationVersion,
				},
				Description:   "a new updated description",
				SchemaType:    "https://in-toto.io/Statement/v0.1",
				PredicateType: "cosign.sigstore.dev/attestation/v1",
				Data:          "some-attestation-data",
				Digests:       []string{"some-digest"},
				ModuleID:      moduleID,
			},
		},

		{
			name:  "negative: terraform module attestation update query returns error",
			input: &types.UpdateTerraformModuleAttestationInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Argument \"input\" has invalid value",
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

		{
			name:  "negative: query behaves as if the specified terraform module attestation did not exist",
			input: &types.UpdateTerraformModuleAttestationInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateTerraformModuleAttestationPayload{
					UpdateTerraformModuleAttestation: graphqlUpdateTerraformModuleAttestationMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Terraform module attestation with ID module-attestation-id-1 not found",
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
			client.TerraformModuleAttestation = NewTerraformModuleAttestation(client)

			// Call the method being tested.
			actualAttestation, actualError := client.TerraformModuleAttestation.UpdateModuleAttestation(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkModuleAttestations(t, test.expectAttestation, actualAttestation)
		})
	}
}

func TestDeleteModuleAttestation(t *testing.T) {
	moduleAttestationID := "module-attestation-id-1"

	type graphqlDeleteTerraformModuleAttestationMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteTerraformModuleAttestationPayload struct {
		DeleteTerraformModuleAttestation graphqlDeleteTerraformModuleAttestationMutation `json:"deleteTerraformModuleAttestation"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.DeleteTerraformModuleAttestationInput
		name            string
		expectErrorCode ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully deleted terraform module attestation",
			input: &types.DeleteTerraformModuleAttestationInput{
				ID: moduleAttestationID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteTerraformModuleAttestationPayload{
					DeleteTerraformModuleAttestation: graphqlDeleteTerraformModuleAttestationMutation{},
				},
			},
		},

		// negative: mutation returns error
		{
			name:  "negative: managed identity alias delete mutation returns error",
			input: &types.DeleteTerraformModuleAttestationInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02)",
						Path: []string{
							"deleteTerraformModuleAttestation",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
		},

		// negative: mutation behaves as if the specified alias did not exist
		{
			name:  "negative: mutation behaves as if the specified alias did not exist",
			input: &types.DeleteTerraformModuleAttestationInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteTerraformModuleAttestationPayload{
					DeleteTerraformModuleAttestation: graphqlDeleteTerraformModuleAttestationMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "Terraform module attestation with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
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
			client.TerraformModuleAttestation = NewTerraformModuleAttestation(client)

			// Call the method being tested.
			actualError := client.TerraformModuleAttestation.DeleteModuleAttestation(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}
}

// Utility functions.

func checkModuleAttestations(t *testing.T, expectAttestation, actualAttestation *types.TerraformModuleAttestation) {
	if expectAttestation != nil {
		require.NotNil(t, actualAttestation)
		assert.Equal(t, expectAttestation, actualAttestation)
	} else {
		assert.Equal(t, (*types.TerraformModuleAttestation)(nil), actualAttestation)
	}
}
