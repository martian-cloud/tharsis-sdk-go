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

func TestGetTerraformProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	tfpID := "tf-provider-id-1"
	tfpVersion := "tf-provider-version-1"
	tfpName := "tf-provider-name-1"
	tfpResourcePath := "tf-provider-resource-path"
	tfpRegistryNamespace := "tf-provider-registry-namespace"
	tfpRepositoryURL := "tf-provider-repository-url"
	tfpPrivate := true

	type graphqlTerraformProviderPayload struct {
		Node *graphQLTerraformProvider `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		input                   *types.GetTerraformProviderInput
		expectTerraformProvider *types.TerraformProvider
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return Terraform provider by ID",
			input: &types.GetTerraformProviderInput{
				ID: tfpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlTerraformProviderPayload{
					Node: &graphQLTerraformProvider{
						ID: graphql.String(tfpID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(tfpVersion),
						},
						Name:              tfpName,
						ResourcePath:      tfpResourcePath,
						RegistryNamespace: tfpRegistryNamespace,
						RepositoryURL:     tfpRepositoryURL,
						Private:           tfpPrivate,
					},
				},
			},
			expectTerraformProvider: &types.TerraformProvider{
				Metadata: types.ResourceMetadata{
					ID:                   tfpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              tfpVersion,
				},
				Name:              tfpName,
				ResourcePath:      tfpResourcePath,
				RegistryNamespace: tfpRegistryNamespace,
				RepositoryURL:     tfpRepositoryURL,
				Private:           tfpPrivate,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetTerraformProviderInput{
				ID: tfpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlTerraformProviderPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
		},
		{
			name: "query returns nil Terraform provider, as if the specified Terraform provider does not exist",
			input: &types.GetTerraformProviderInput{
				ID: tfpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlTerraformProviderPayload{},
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
			client.TerraformProvider = NewTerraformProvider(client)

			// Call the method being tested.
			actualTerraformProvider, actualError := client.TerraformProvider.GetProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTerraformProvider(t, test.expectTerraformProvider, actualTerraformProvider)
		})
	}
}

func TestCreateTerraformProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	tfpID := "tf-provider-id-1"
	tfpVersion := "tf-provider-version-1"
	tfpName := "tf-provider-name-1"
	tfpGroupPath := "tf-provider-group-path-1"
	tfpResourcePath := "tf-provider-resource-path"
	tfpRegistryNamespace := "tf-provider-registry-namespace"
	tfpRepositoryURL := "tf-provider-repository-url"
	tfpPrivate := true

	type graphqlCreateTerraformProviderMutation struct {
		Provider graphQLTerraformProvider     `json:"provider"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateTerraformProviderPayload struct {
		CreateTerraformProvider graphqlCreateTerraformProviderMutation `json:"createTerraformProvider"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		input                   *types.CreateTerraformProviderInput
		expectTerraformProvider *types.TerraformProvider
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully create Terraform provider",
			input: &types.CreateTerraformProviderInput{
				Name:          tfpName,
				GroupPath:     tfpGroupPath,
				RepositoryURL: tfpRepositoryURL,
				Private:       tfpPrivate,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateTerraformProviderPayload{
					CreateTerraformProvider: graphqlCreateTerraformProviderMutation{
						Provider: graphQLTerraformProvider{
							ID: graphql.String(tfpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(tfpVersion),
							},
							Name:              tfpName,
							ResourcePath:      tfpResourcePath,
							RegistryNamespace: tfpRegistryNamespace,
							RepositoryURL:     tfpRepositoryURL,
							Private:           tfpPrivate,
						},
					},
				},
			},
			expectTerraformProvider: &types.TerraformProvider{
				Metadata: types.ResourceMetadata{
					ID:                   tfpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              tfpVersion,
				},
				Name:              tfpName,
				ResourcePath:      tfpResourcePath,
				RegistryNamespace: tfpRegistryNamespace,
				RepositoryURL:     tfpRepositoryURL,
				Private:           tfpPrivate,
			},
		},
		{
			name:  "verify that correct error is returned",
			input: &types.CreateTerraformProviderInput{},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateTerraformProviderPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
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
			client.TerraformProvider = NewTerraformProvider(client)

			// Call the method being tested.
			actualTerraformProvider, actualError := client.TerraformProvider.CreateProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTerraformProvider(t, test.expectTerraformProvider, actualTerraformProvider)
		})
	}
}

func TestUpdateTerraformProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	tfpID := "tf-provider-id-1"
	tfpVersion := "tf-provider-version-1"
	tfpName := "tf-provider-name-1"
	tfpResourcePath := "tf-provider-resource-path"
	tfpRegistryNamespace := "tf-provider-registry-namespace"
	tfpRepositoryURL := "tf-provider-repository-url"
	tfpPrivate := true

	type graphqlUpdateTerraformProviderMutation struct {
		Provider graphQLTerraformProvider     `json:"provider"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateTerraformProviderPayload struct {
		UpdateTerraformProvider graphqlUpdateTerraformProviderMutation `json:"updateTerraformProvider"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		input                   *types.UpdateTerraformProviderInput
		expectTerraformProvider *types.TerraformProvider
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully update Terraform provider",
			input: &types.UpdateTerraformProviderInput{
				ID:            tfpID,
				Name:          tfpName,
				RepositoryURL: tfpRepositoryURL,
				Private:       tfpPrivate,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateTerraformProviderPayload{
					UpdateTerraformProvider: graphqlUpdateTerraformProviderMutation{
						Provider: graphQLTerraformProvider{
							ID: graphql.String(tfpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(tfpVersion),
							},
							Name:              tfpName,
							ResourcePath:      tfpResourcePath,
							RegistryNamespace: tfpRegistryNamespace,
							RepositoryURL:     tfpRepositoryURL,
							Private:           tfpPrivate,
						},
					},
				},
			},
			expectTerraformProvider: &types.TerraformProvider{
				Metadata: types.ResourceMetadata{
					ID:                   tfpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              tfpVersion,
				},
				Name:              tfpName,
				ResourcePath:      tfpResourcePath,
				RegistryNamespace: tfpRegistryNamespace,
				RepositoryURL:     tfpRepositoryURL,
				Private:           tfpPrivate,
			},
		},
		{
			name:  "verify that correct error is returned",
			input: &types.UpdateTerraformProviderInput{},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateTerraformProviderPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
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
			client.TerraformProvider = NewTerraformProvider(client)

			// Call the method being tested.
			actualTerraformProvider, actualError := client.TerraformProvider.UpdateProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTerraformProvider(t, test.expectTerraformProvider, actualTerraformProvider)
		})
	}
}

func TestDeleteTerraformProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	tfpID := "tf-provider-id-1"
	tfpVersion := "tf-provider-version-1"
	tfpName := "tf-provider-name-1"
	tfpResourcePath := "tf-provider-resource-path"
	tfpRegistryNamespace := "tf-provider-registry-namespace"
	tfpRepositoryURL := "tf-provider-repository-url"
	tfpPrivate := true

	type graphqlDeleteTerraformProviderMutation struct {
		Provider graphQLTerraformProvider     `json:"provider"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteTerraformProviderPayload struct {
		DeleteTerraformProvider graphqlDeleteTerraformProviderMutation `json:"deleteTerraformProvider"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		input                   *types.DeleteTerraformProviderInput
		expectTerraformProvider *types.TerraformProvider
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully delete Terraform provider",
			input: &types.DeleteTerraformProviderInput{
				ID: tfpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteTerraformProviderPayload{
					DeleteTerraformProvider: graphqlDeleteTerraformProviderMutation{
						Provider: graphQLTerraformProvider{
							ID: graphql.String(tfpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(tfpVersion),
							},
							Name:              tfpName,
							ResourcePath:      tfpResourcePath,
							RegistryNamespace: tfpRegistryNamespace,
							RepositoryURL:     tfpRepositoryURL,
							Private:           tfpPrivate,
						},
					},
				},
			},
			expectTerraformProvider: &types.TerraformProvider{
				Metadata: types.ResourceMetadata{
					ID:                   tfpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              tfpVersion,
				},
				Name:              tfpName,
				ResourcePath:      tfpResourcePath,
				RegistryNamespace: tfpRegistryNamespace,
				RepositoryURL:     tfpRepositoryURL,
				Private:           tfpPrivate,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.DeleteTerraformProviderInput{
				ID: tfpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteTerraformProviderPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "an error occurred",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
		},
		{
			name: "query returns nil Terraform provider, as if the specified Terraform provider does not exist",
			input: &types.DeleteTerraformProviderInput{
				ID: tfpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteTerraformProviderPayload{
					DeleteTerraformProvider: graphqlDeleteTerraformProviderMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Type:    "NOT_FOUND",
								Message: "Terraform provider with ID something not found",
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
			client.TerraformProvider = NewTerraformProvider(client)

			// Call the method being tested.
			actualTerraformProvider, actualError := client.TerraformProvider.DeleteProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTerraformProvider(t, test.expectTerraformProvider, actualTerraformProvider)
		})
	}

}

// Utility functions:

func checkTerraformProvider(t *testing.T, expectTerraformProvider, actualTerraformProvider *types.TerraformProvider) {
	if expectTerraformProvider != nil {
		require.NotNil(t, actualTerraformProvider)
		assert.Equal(t, expectTerraformProvider, actualTerraformProvider)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.TerraformProvider)(nil)
		assert.Equal(t, (*types.TerraformProvider)(nil), actualTerraformProvider)
	}
}

// The End.
