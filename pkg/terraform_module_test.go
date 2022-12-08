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

func TestGetModule(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleID := "1"

	type graphqlModulePayload struct {
		Node *graphQLTerraformModule `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		expectModule       *types.TerraformModule
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{
		{
			name: "Successfully return module by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayload{
					Node: &graphQLTerraformModule{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   "1",
						},
						ID:                graphql.String(moduleID),
						Name:              "awesome-module",
						System:            "aws",
						ResourcePath:      "groupA/awesome-module/aws",
						RegistryNamespace: "groupA",
						Private:           true,
					},
				},
			},
			expectModule: &types.TerraformModule{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleID,
					Version:              "1",
				},
				Name:              "awesome-module",
				System:            "aws",
				ResourcePath:      "groupA/awesome-module/aws",
				RegistryNamespace: "groupA",
				Private:           true,
			},
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayload{},
				Errors: []fakeGraphqlResponseError{{
					Message: "an error occurred",
					Extensions: fakeGraphqlResponseErrorExtension{
						Code: "INTERNAL_SERVER_ERROR",
					},
				}},
			},
			expectErrorMessage: "Message: an error occurred, Locations: []",
		},
		{
			name: "returns nil because module does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayload{},
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			module, actualError := client.TerraformModule.GetModule(
				ctx,
				&types.GetTerraformModuleInput{ID: moduleID},
			)

			checkError(t, test.expectErrorMessage, actualError)

			if test.expectModule != nil {
				require.NotNil(t, module)
				assert.Equal(t, module, test.expectModule)
			}
		})
	}
}

func TestCreateModule(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleID := "1"

	type graphqlCreateModuleMutation struct {
		Module   *graphQLTerraformModule      `json:"module"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateModulePayload struct {
		CreateTerraformModule graphqlCreateModuleMutation `json:"createTerraformModule"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		expectModule       *types.TerraformModule
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{
		{
			name: "Successfully created terraform module",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateModulePayload{
					CreateTerraformModule: graphqlCreateModuleMutation{
						Module: &graphQLTerraformModule{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   "1",
							},
							ID:                graphql.String(moduleID),
							Name:              "awesome-module",
							System:            "aws",
							ResourcePath:      "groupA/awesome-module/aws",
							RegistryNamespace: "groupA",
							Private:           true,
						},
					},
				},
			},
			expectModule: &types.TerraformModule{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleID,
					Version:              "1",
				},
				Name:              "awesome-module",
				System:            "aws",
				ResourcePath:      "groupA/awesome-module/aws",
				RegistryNamespace: "groupA",
				Private:           true,
			},
		},
		{
			name: "create module returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateModulePayload{
					CreateTerraformModule: graphqlCreateModuleMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "module already exists",
								Type:    internal.Conflict,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems creating module: module already exists",
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			module, actualError := client.TerraformModule.CreateModule(ctx, &types.CreateTerraformModuleInput{})

			checkError(t, test.expectErrorMessage, actualError)

			if test.expectModule != nil {
				require.NotNil(t, module)
				assert.Equal(t, module, test.expectModule)
			}
		})
	}
}

func TestUpdateModule(t *testing.T) {
	now := time.Now().UTC()

	moduleID := "1"

	type graphqlUpdateModuleMutation struct {
		Module   *graphQLTerraformModule      `json:"module"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateModulePayload struct {
		UpdateTerraformModule graphqlUpdateModuleMutation `json:"updateTerraformModule"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		expectModule       *types.TerraformModule
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{
		{
			name: "Successful update of terraform module",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateModulePayload{
					UpdateTerraformModule: graphqlUpdateModuleMutation{
						Module: &graphQLTerraformModule{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   "1",
							},
							ID:                graphql.String(moduleID),
							Name:              "awesome-module",
							System:            "aws",
							ResourcePath:      "groupA/awesome-module/aws",
							RegistryNamespace: "groupA",
							Private:           true,
						},
					},
				},
			},
			expectModule: &types.TerraformModule{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleID,
					Version:              "1",
				},
				Name:              "awesome-module",
				System:            "aws",
				ResourcePath:      "groupA/awesome-module/aws",
				RegistryNamespace: "groupA",
				Private:           true,
			},
		},
		{
			name: "update module returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateModulePayload{
					UpdateTerraformModule: graphqlUpdateModuleMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "module not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems updating module: module not found",
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			module, actualError := client.TerraformModule.UpdateModule(ctx, &types.UpdateTerraformModuleInput{})

			checkError(t, test.expectErrorMessage, actualError)

			if test.expectModule != nil {
				require.NotNil(t, module)
				assert.Equal(t, module, test.expectModule)
			}
		})
	}
}

func TestDeleteModule(t *testing.T) {
	type graphqlDeleteModuleMutation struct {
		Module   *graphQLTerraformModule      `json:"module"`
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteModulePayload struct {
		DeleteTerraformModule graphqlDeleteModuleMutation `json:"deleteTerraformModule"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{
		{
			name: "Successful deletion of terraform module",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteModulePayload{
					DeleteTerraformModule: graphqlDeleteModuleMutation{},
				},
			},
		},
		{
			name: "delete module returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteModulePayload{
					DeleteTerraformModule: graphqlDeleteModuleMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "module not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems deleting module: module not found",
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			err = client.TerraformModule.DeleteModule(ctx, &types.DeleteTerraformModuleInput{})

			checkError(t, test.expectErrorMessage, err)
		})
	}
}