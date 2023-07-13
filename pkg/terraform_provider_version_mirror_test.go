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

func TestGetProviderVersionMirror(t *testing.T) {
	now := time.Now().UTC()

	versionMirrorID := "version-mirror-1"

	type graphQLProviderVersionMirrorPayload struct {
		Node *graphQLTerraformProviderVersionMirror `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		expectMirror    *types.TerraformProviderVersionMirror
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "successfully return a version mirror",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphQLProviderVersionMirrorPayload{
					Node: &graphQLTerraformProviderVersionMirror{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String("1"),
						},
						ID:                graphql.String(versionMirrorID),
						Version:           "0.0.1",
						RegistryHostname:  "registry.terraform.io",
						RegistryNamespace: "hashicorp",
						Type:              "aws",
					},
				},
			},
			expectMirror: &types.TerraformProviderVersionMirror{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   versionMirrorID,
					Version:              "1",
				},
				SemanticVersion:   "0.0.1",
				RegistryHostname:  "registry.terraform.io",
				RegistryNamespace: "hashicorp",
				Type:              "aws",
			},
		},
		{
			name:            "terraform provider version mirror does not exist",
			responsePayload: graphQLProviderVersionMirrorPayload{},
			expectErrorCode: types.ErrNotFound,
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphQLProviderVersionMirrorPayload{},
				Errors: []fakeGraphqlResponseError{{
					Message: "an error occurred",
					Extensions: fakeGraphqlResponseErrorExtension{
						Code: "INTERNAL_SERVER_ERROR",
					},
				}},
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
				})}
			client.TerraformProviderVersionMirror = NewTerraformProviderVersionMirror(client)

			// Call the method being tested.
			versionMirror, actualError := client.TerraformProviderVersionMirror.GetProviderVersionMirror(ctx, &types.GetTerraformProviderVersionMirrorInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectMirror != nil {
				require.NotNil(t, versionMirror)
				assert.Equal(t, versionMirror, test.expectMirror)
			}
		})
	}
}

func TestGetProviderVersionMirrorByAddress(t *testing.T) {
	now := time.Now().UTC()

	versionMirrorID := "version-mirror-1"

	type graphQLProviderVersionMirrorPayload struct {
		TerraformProviderVersionMirror *graphQLTerraformProviderVersionMirror `json:"terraformProviderVersionMirror"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		expectMirror    *types.TerraformProviderVersionMirror
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "successfully return a version mirror",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphQLProviderVersionMirrorPayload{
					TerraformProviderVersionMirror: &graphQLTerraformProviderVersionMirror{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String("1"),
						},
						ID:                graphql.String(versionMirrorID),
						Version:           "0.0.1",
						RegistryHostname:  "registry.terraform.io",
						RegistryNamespace: "hashicorp",
						Type:              "aws",
					},
				},
			},
			expectMirror: &types.TerraformProviderVersionMirror{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   versionMirrorID,
					Version:              "1",
				},
				SemanticVersion:   "0.0.1",
				RegistryHostname:  "registry.terraform.io",
				RegistryNamespace: "hashicorp",
				Type:              "aws",
			},
		},
		{
			name:            "terraform provider version mirror does not exist",
			responsePayload: graphQLProviderVersionMirrorPayload{},
			expectErrorCode: types.ErrNotFound,
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphQLProviderVersionMirrorPayload{},
				Errors: []fakeGraphqlResponseError{{
					Message: "an error occurred",
					Extensions: fakeGraphqlResponseErrorExtension{
						Code: "INTERNAL_SERVER_ERROR",
					},
				}},
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
				})}
			client.TerraformProviderVersionMirror = NewTerraformProviderVersionMirror(client)

			// Call the method being tested.
			versionMirror, actualError := client.TerraformProviderVersionMirror.GetProviderVersionMirrorByAddress(ctx, &types.GetTerraformProviderVersionMirrorByAddressInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectMirror != nil {
				require.NotNil(t, versionMirror)
				assert.Equal(t, versionMirror, test.expectMirror)
			}
		})
	}
}

func TestCreateProviderVersionMirror(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	versionMirrorID := "version-mirror-1"

	type graphqlTerraformProviderVersionMirrorMutation struct {
		VersionMirror *graphQLTerraformProviderVersionMirror `json:"versionMirror"`
		Problems      []fakeGraphqlResponseProblem           `json:"problems"`
	}

	type graphqlCreateTerraformProviderVersionMirrorPayload struct {
		CreateTerraformProviderVersionMirror graphqlTerraformProviderVersionMirrorMutation `json:"createTerraformProviderVersionMirror"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		expectMirror    *types.TerraformProviderVersionMirror
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "successfully create a version mirror",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateTerraformProviderVersionMirrorPayload{
					CreateTerraformProviderVersionMirror: graphqlTerraformProviderVersionMirrorMutation{
						VersionMirror: &graphQLTerraformProviderVersionMirror{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String("1"),
							},
							ID:                graphql.String(versionMirrorID),
							Version:           "0.0.1",
							RegistryHostname:  "registry.terraform.io",
							RegistryNamespace: "hashicorp",
							Type:              "aws",
						},
					},
				},
			},
			expectMirror: &types.TerraformProviderVersionMirror{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   versionMirrorID,
					Version:              "1",
				},
				SemanticVersion:   "0.0.1",
				RegistryHostname:  "registry.terraform.io",
				RegistryNamespace: "hashicorp",
				Type:              "aws",
			},
		},
		{
			name: "create version mirror returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateTerraformProviderVersionMirrorPayload{
					CreateTerraformProviderVersionMirror: graphqlTerraformProviderVersionMirrorMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "version mirror already exists",
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
			client.TerraformProviderVersionMirror = NewTerraformProviderVersionMirror(client)

			// Call the method being tested.
			versionMirror, actualError := client.TerraformProviderVersionMirror.CreateProviderVersionMirror(ctx, &types.CreateTerraformProviderVersionMirrorInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectMirror != nil {
				require.NotNil(t, versionMirror)
				assert.Equal(t, versionMirror, test.expectMirror)
			}
		})
	}
}

func TestDeleteProviderVersionMirror(t *testing.T) {
	type graphqlDeleteProviderVersionMirrorMutation struct {
		VersionMirror graphQLTerraformProviderVersionMirror `json:"versionMirror"`
		Problems      []fakeGraphqlResponseProblem          `json:"problems"`
	}

	type graphqlDeleteProviderVersionMirrorPayload struct {
		DeleteTerraformProviderVersionMirror graphqlDeleteProviderVersionMirrorMutation `json:"deleteTerraformProviderVersionMirror"`
	}

	type testCase struct {
		responsePayload interface{}
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successful deletion of provider version mirror",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteProviderVersionMirrorPayload{
					DeleteTerraformProviderVersionMirror: graphqlDeleteProviderVersionMirrorMutation{},
				},
			},
		},
		{
			name: "delete provider version mirror returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteProviderVersionMirrorPayload{
					DeleteTerraformProviderVersionMirror: graphqlDeleteProviderVersionMirrorMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "provider version mirror not found",
								Type:    internal.NotFound,
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
				})}
			client.TerraformProviderVersionMirror = NewTerraformProviderVersionMirror(client)

			// Call the method being tested.
			actualError := client.TerraformProviderVersionMirror.DeleteProviderVersionMirror(ctx, &types.DeleteTerraformProviderVersionMirrorInput{})

			checkError(t, test.expectErrorCode, actualError)
		})
	}
}
