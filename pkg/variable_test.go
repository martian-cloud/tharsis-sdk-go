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

func TestCreateNamespaceVariable(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	namespacePath := "parent-group-name"
	namespaceVariableID := "namespace-variable-id-1"
	namespaceVariableVersion := "namespace-variable-version-1"
	namespaceVariable1Key := "variable-1-key"
	namespaceVariable1Value := "variable-1-value"
	namespaceVariable2Key := "variable-2-key"
	namespaceVariable2Value := "variable-2-value"

	type graphQLNamespace struct {
		Variables []graphQLNamespaceVariable `json:"variables"`
	}

	type graphqlCreateNamespaceVariableMutation struct {
		Namespace graphQLNamespace             `json:"namespace"`
		Problems  []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateNamespaceVariablePayload struct {
		CreateNamespaceVariable graphqlCreateNamespaceVariableMutation `json:"createNamespaceVariable"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		input                   *types.CreateNamespaceVariableInput
		expectNamespaceVariable *types.NamespaceVariable
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive, Terraform HCL
		{
			name: "Successfully created a Terraform HCL namespace variable",
			input: &types.CreateNamespaceVariableInput{
				NamespacePath: namespacePath,
				Category:      types.TerraformVariableCategory,
				HCL:           true,
				Key:           namespaceVariable1Key,
				Value:         namespaceVariable1Value,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateNamespaceVariablePayload{
					CreateNamespaceVariable: graphqlCreateNamespaceVariableMutation{
						Namespace: graphQLNamespace{
							Variables: []graphQLNamespaceVariable{
								{
									ID: graphql.String(namespaceVariableID),
									Metadata: internal.GraphQLMetadata{
										CreatedAt: &now,
										UpdatedAt: &now,
										Version:   graphql.String(namespaceVariableVersion),
									},
									NamespacePath: graphql.String(namespacePath),
									Category:      graphql.String(types.TerraformVariableCategory),
									HCL:           true,
									Key:           graphql.String(namespaceVariable1Key),
									Value:         (*graphql.String)(&namespaceVariable1Value),
								},
							},
						},
					},
				},
			},
			expectNamespaceVariable: &types.NamespaceVariable{
				Metadata: types.ResourceMetadata{
					ID:                   namespaceVariableID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceVariableVersion,
				},
				NamespacePath: namespacePath,
				Category:      types.TerraformVariableCategory,
				HCL:           true,
				Key:           namespaceVariable1Key,
				Value:         &namespaceVariable1Value,
			},
		},

		// positive, environment, string
		{
			name: "Successfully created an environment non-HCL namespace variable",
			input: &types.CreateNamespaceVariableInput{
				NamespacePath: namespacePath,
				Category:      types.EnvironmentVariableCategory,
				HCL:           false,
				Key:           namespaceVariable2Key,
				Value:         namespaceVariable2Value,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateNamespaceVariablePayload{
					CreateNamespaceVariable: graphqlCreateNamespaceVariableMutation{
						Namespace: graphQLNamespace{
							Variables: []graphQLNamespaceVariable{
								{
									ID: graphql.String(namespaceVariableID),
									Metadata: internal.GraphQLMetadata{
										CreatedAt: &now,
										UpdatedAt: &now,
										Version:   graphql.String(namespaceVariableVersion),
									},
									NamespacePath: graphql.String(namespacePath),
									Category:      graphql.String(types.EnvironmentVariableCategory),
									HCL:           false,
									Key:           graphql.String(namespaceVariable2Key),
									Value:         (*graphql.String)(&namespaceVariable2Value),
								},
							},
						},
					},
				},
			},
			expectNamespaceVariable: &types.NamespaceVariable{
				Metadata: types.ResourceMetadata{
					ID:                   namespaceVariableID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceVariableVersion,
				},
				NamespacePath: namespacePath,
				Category:      types.EnvironmentVariableCategory,
				HCL:           false,
				Key:           namespaceVariable2Key,
				Value:         &namespaceVariable2Value,
			},
		},

		// negative: query returns error
		{
			name:  "negative: query to create namespace variable returned error",
			input: &types.CreateNamespaceVariableInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateNamespaceVariablePayload{
					CreateNamespaceVariable: graphqlCreateNamespaceVariableMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "namespace variable with path non-existent not found",
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
			client.Variable = NewVariable(client)

			// Call the method being tested.
			actualNamespaceVariable, actualError := client.Variable.CreateVariable(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkNamespaceVariable(t, test.expectNamespaceVariable, actualNamespaceVariable)
		})
	}
}

func TestGetNamespaceVariable(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	namespacePath := "parent-group-name"
	namespaceVariableID := "namespace-variable-id-1"
	namespaceVariableVersion := "namespace-variable-version-1"
	namespaceVariableKey := "variable-key"
	namespaceVariableValue := "variable-value"

	// Field name taken from GraphiQL.
	type graphqlNodeNamespaceVariablePayload struct {
		Node *graphQLNamespaceVariable `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		expectNamespaceVariable *types.NamespaceVariable
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return namespace variable by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeNamespaceVariablePayload{
					Node: &graphQLNamespaceVariable{
						ID: graphql.String(namespaceVariableID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(namespaceVariableVersion),
						},
						NamespacePath: graphql.String(namespacePath),
						Category:      graphql.String(types.TerraformVariableCategory),
						HCL:           true,
						Key:           graphql.String(namespaceVariableKey),
						Value:         (*graphql.String)(&namespaceVariableValue),
					},
				},
			},
			expectNamespaceVariable: &types.NamespaceVariable{
				Metadata: types.ResourceMetadata{
					ID:                   namespaceVariableID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceVariableVersion,
				},
				NamespacePath: namespacePath,
				Category:      types.TerraformVariableCategory,
				HCL:           true,
				Key:           namespaceVariableKey,
				Value:         &namespaceVariableValue,
			},
		},

		// negative: query returns error, invalid ID--payload taken from GraphiQL
		{
			name: "negative: query returns error, invalid ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlNodeNamespaceVariablePayload{},
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
				Data: graphqlNodeNamespaceVariablePayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "namespace variable with id 6f3106c6-c342-4790-a667-963a850d34d4 not found",
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
				Data: graphqlNodeNamespaceVariablePayload{},
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
			client.Variable = NewVariable(client)

			// Call the method being tested.
			actualNamespaceVariable, actualError := client.Variable.GetVariable(
				ctx,
				&types.GetNamespaceVariableInput{ID: namespaceVariableID},
			)

			checkError(t, test.expectErrorCode, actualError)
			checkNamespaceVariable(t, test.expectNamespaceVariable, actualNamespaceVariable)
		})
	}
}

func TestUpdateNamespaceVariable(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	namespacePath := "parent-group-name"
	namespaceVariableID := "namespace-variable-id-1"
	namespaceVariableVersion := "namespace-variable-version-1"
	namespaceVariableKey := "variable-key"
	namespaceVariableValue := "variable-value"

	type graphQLNamespace struct {
		Variables []graphQLNamespaceVariable `json:"variables"`
	}

	type graphqlUpdateNamespaceVariableMutation struct {
		Namespace graphQLNamespace             `json:"namespace"`
		Problems  []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateNamespaceVariablePayload struct {
		UpdateNamespaceVariable graphqlUpdateNamespaceVariableMutation `json:"updateNamespaceVariable"`
	}

	// test cases
	type testCase struct {
		responsePayload         interface{}
		input                   *types.UpdateNamespaceVariableInput
		expectNamespaceVariable *types.NamespaceVariable
		name                    string
		expectErrorCode         ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully updated namespace variable",
			input: &types.UpdateNamespaceVariableInput{
				ID:    namespaceVariableID,
				HCL:   true,
				Key:   namespaceVariableKey,
				Value: namespaceVariableValue,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateNamespaceVariablePayload{
					UpdateNamespaceVariable: graphqlUpdateNamespaceVariableMutation{
						Namespace: graphQLNamespace{
							Variables: []graphQLNamespaceVariable{
								{
									ID: graphql.String(namespaceVariableID),
									Metadata: internal.GraphQLMetadata{
										CreatedAt: &now,
										UpdatedAt: &now,
										Version:   graphql.String(namespaceVariableVersion),
									},
									NamespacePath: graphql.String(namespacePath),
									Category:      graphql.String(types.TerraformVariableCategory),
									HCL:           true,
									Key:           graphql.String(namespaceVariableKey),
									Value:         (*graphql.String)(&namespaceVariableValue),
								},
							},
						},
					},
				},
			},
			expectNamespaceVariable: &types.NamespaceVariable{
				Metadata: types.ResourceMetadata{
					ID:                   namespaceVariableID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              namespaceVariableVersion,
				},
				NamespacePath: namespacePath,
				Category:      types.TerraformVariableCategory,
				HCL:           true,
				Key:           namespaceVariableKey,
				Value:         &namespaceVariableValue,
			},
		},

		// negative: namespace variable update query returns error
		{
			name:  "negative: namespace variable update query returns error",
			input: &types.UpdateNamespaceVariableInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Argument \"input\" has invalid value {id: \"TVJfZmUyZWI1NjQtNjMxMS00MmFlLTkwMWYtOTE5NTEyNWNhOTJh\", runStage: invalid, allowedUsers: [\"robert.richesjr\"], allowedNamespaceVariables: [\"provider-test-parent-group/sa1\"], allowedTeams: [\"team1\", \"team2\"]}.\nIn field \"runStage\": Expected type \"JobType\", found invalid.",
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

		// negative: query behaves as if the specified namespace variable did not exist
		{
			name:  "negative: query behaves as if the specified namespace variable did not exist",
			input: &types.UpdateNamespaceVariableInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateNamespaceVariablePayload{
					UpdateNamespaceVariable: graphqlUpdateNamespaceVariableMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "namespace variable with ID fe2eb564-6311-52ae-901f-9195125ca92a not found",
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
			client.Variable = NewVariable(client)

			// Call the method being tested.
			actualNamespaceVariable, actualError := client.Variable.UpdateVariable(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkNamespaceVariable(t, test.expectNamespaceVariable, actualNamespaceVariable)
		})
	}
}

func TestDeleteNamespaceVariable(t *testing.T) {
	namespaceVariableID := "namespace-variable-id-1"

	type graphqlDeleteNamespaceVariableMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteNamespaceVariablePayload struct {
		DeleteNamespaceVariable graphqlDeleteNamespaceVariableMutation `json:"deleteNamespaceVariable"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.DeleteNamespaceVariableInput
		name            string
		expectErrorCode ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully deleted namespace variable",
			input: &types.DeleteNamespaceVariableInput{
				ID: namespaceVariableID,
			},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteNamespaceVariablePayload{
					DeleteNamespaceVariable: graphqlDeleteNamespaceVariableMutation{},
				},
			},
		},

		// negative: mutation returns error
		{
			name:  "negative: namespace variable delete mutation returns error",
			input: &types.DeleteNamespaceVariableInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"fe2eb564-6311-42qe-901f-9195125ca92a\" (SQLSTATE 22P02)",
						Path: []string{
							"deleteNamespaceVariable",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: ErrInternal,
		},

		// negative: mutation behaves as if the specified namespace variable did not exist
		{
			name:  "negative: mutation behaves as if the specified namespace variable did not exist",
			input: &types.DeleteNamespaceVariableInput{},
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteNamespaceVariablePayload{
					DeleteNamespaceVariable: graphqlDeleteNamespaceVariableMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "namespace variable with ID fe2eb564-6311-42ae-901f-9195125ca92a not found",
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
			client.Variable = NewVariable(client)

			// Call the method being tested.
			actualError := client.Variable.DeleteVariable(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}

}

// Utility functions:

func checkNamespaceVariable(t *testing.T, expectNamespaceVariable, actualNamespaceVariable *types.NamespaceVariable) {
	if expectNamespaceVariable != nil {
		require.NotNil(t, actualNamespaceVariable)
		assert.Equal(t, expectNamespaceVariable, actualNamespaceVariable)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.NamespaceVariable)(nil)
		assert.Equal(t, (*types.NamespaceVariable)(nil), actualNamespaceVariable)
	}
}

// The End.
