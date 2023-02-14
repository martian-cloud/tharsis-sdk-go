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

func TestGetModule(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleID := "1"
	modulePath := "groupA/awesome-module/aws"
	groupPath := "groupA"

	type graphqlModulePayloadByID struct {
		Node *graphQLTerraformModule `json:"node"`
	}

	type graphqlModulePayloadByPath struct {
		TerraformModule *graphQLTerraformModule `json:"terraformModule"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.GetTerraformModuleInput
		expectModule    *types.TerraformModule
		name            string
		expectErrorCode ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully return module by ID",
			input: &types.GetTerraformModuleInput{
				ID: &moduleID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayloadByID{
					Node: &graphQLTerraformModule{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   "1",
						},
						ID:                graphql.String(moduleID),
						Name:              "awesome-module",
						System:            "aws",
						GroupPath:         groupPath,
						ResourcePath:      modulePath,
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
				GroupPath:         groupPath,
				ResourcePath:      modulePath,
				RegistryNamespace: "groupA",
				Private:           true,
			},
		},
		{
			name: "Successfully return module by path",
			input: &types.GetTerraformModuleInput{
				Path: &modulePath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayloadByPath{
					TerraformModule: &graphQLTerraformModule{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   "1",
						},
						ID:                graphql.String(moduleID),
						Name:              "awesome-module",
						System:            "aws",
						GroupPath:         groupPath,
						ResourcePath:      modulePath,
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
				GroupPath:         groupPath,
				ResourcePath:      modulePath,
				RegistryNamespace: "groupA",
				Private:           true,
			},
		},
		{
			name:            "returns an error since ID and path are unspecified",
			input:           &types.GetTerraformModuleInput{},
			expectErrorCode: ErrBadRequest,
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetTerraformModuleInput{
				ID: &moduleID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayloadByID{},
				Errors: []fakeGraphqlResponseError{{
					Message: "an error occurred",
					Extensions: fakeGraphqlResponseErrorExtension{
						Code: "INTERNAL_SERVER_ERROR",
					},
				}},
			},
			expectErrorCode: ErrInternal,
		},
		{
			name: "returns nil because module does not exist",
			input: &types.GetTerraformModuleInput{
				ID: &moduleID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModulePayloadByID{},
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			module, actualError := client.TerraformModule.GetModule(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)

			if test.expectModule != nil {
				require.NotNil(t, module)
				assert.Equal(t, test.expectModule, module)
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
		responsePayload interface{}
		expectModule    *types.TerraformModule
		name            string
		expectErrorCode ErrorCode
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
							GroupPath:         "groupA",
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
				GroupPath:         "groupA",
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
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			module, actualError := client.TerraformModule.CreateModule(ctx, &types.CreateTerraformModuleInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectModule != nil {
				require.NotNil(t, module)
				assert.Equal(t, test.expectModule, module)
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
		responsePayload interface{}
		expectModule    *types.TerraformModule
		name            string
		expectErrorCode ErrorCode
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
							GroupPath:         "groupA",
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
				GroupPath:         "groupA",
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			module, actualError := client.TerraformModule.UpdateModule(ctx, &types.UpdateTerraformModuleInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectModule != nil {
				require.NotNil(t, module)
				assert.Equal(t, test.expectModule, module)
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
		responsePayload interface{}
		name            string
		expectErrorCode ErrorCode
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
				})}
			client.TerraformModule = NewTerraformModule(client)

			// Call the method being tested.
			err = client.TerraformModule.DeleteModule(ctx, &types.DeleteTerraformModuleInput{})

			checkError(t, test.expectErrorCode, err)
		})
	}
}
