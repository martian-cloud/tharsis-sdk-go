package tharsis

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetModuleVersion(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleVersionID := "1"

	type graphqlModuleVersionPayload struct {
		Node *graphQLTerraformModuleVersion `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload     interface{}
		expectModuleVersion *types.TerraformModuleVersion
		name                string
		expectErrorMessage  string
	}

	testCases := []testCase{
		{
			name: "Successfully return module version by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModuleVersionPayload{
					Node: &graphQLTerraformModuleVersion{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   "1",
						},
						ID:          graphql.String(moduleVersionID),
						Version:     "1.0.0",
						SHASum:      "7ae471ed18395339572f5265b835860e28a2f85016455214cb214bafe4422c7d",
						Status:      "pending",
						Error:       "error",
						Diagnostics: "error on line 2",
						Submodules:  []string{"submodule1"},
						Examples:    []string{"example1"},
						Module: graphQLTerraformModule{
							ID: "module-1",
						},
					},
				},
			},
			expectModuleVersion: &types.TerraformModuleVersion{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleVersionID,
					Version:              "1",
				},
				Version:     "1.0.0",
				SHASum:      "7ae471ed18395339572f5265b835860e28a2f85016455214cb214bafe4422c7d",
				Status:      "pending",
				Error:       "error",
				Diagnostics: "error on line 2",
				Submodules:  []string{"submodule1"},
				Examples:    []string{"example1"},
				ModuleID:    "module-1",
			},
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModuleVersionPayload{},
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
			name: "returns nil because module version does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlModuleVersionPayload{},
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
			client.TerraformModuleVersion = NewTerraformModuleVersion(client)

			// Call the method being tested.
			moduleVersion, actualError := client.TerraformModuleVersion.GetModuleVersion(
				ctx,
				&types.GetTerraformModuleVersionInput{ID: moduleVersionID},
			)

			checkError(t, test.expectErrorMessage, actualError)

			if test.expectModuleVersion != nil {
				require.NotNil(t, moduleVersion)
				assert.Equal(t, moduleVersion, test.expectModuleVersion)
			}
		})
	}
}

func TestCreateModuleVersion(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	moduleVersionID := "1"

	type graphqlCreateModuleVersionMutation struct {
		ModuleVersion *graphQLTerraformModuleVersion `json:"moduleVersion"`
		Problems      []fakeGraphqlResponseProblem   `json:"problems"`
	}

	type graphqlCreateModuleVersionPayload struct {
		CreateTerraformModuleVersion graphqlCreateModuleVersionMutation `json:"createTerraformModuleVersion"`
	}

	// test cases
	type testCase struct {
		responsePayload     interface{}
		expectModuleVersion *types.TerraformModuleVersion
		name                string
		expectErrorMessage  string
	}

	testCases := []testCase{
		{
			name: "Successfully created terraform moduleVersion",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateModuleVersionPayload{
					CreateTerraformModuleVersion: graphqlCreateModuleVersionMutation{
						ModuleVersion: &graphQLTerraformModuleVersion{
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   "1",
							},
							ID:          graphql.String(moduleVersionID),
							Version:     "1.0.0",
							SHASum:      "7ae471ed18395339572f5265b835860e28a2f85016455214cb214bafe4422c7d",
							Status:      "pending",
							Error:       "",
							Diagnostics: "",
							Submodules:  []string{"submodule1"},
							Examples:    []string{"example1"},
							Module: graphQLTerraformModule{
								ID: "module-1",
							},
						},
					},
				},
			},
			expectModuleVersion: &types.TerraformModuleVersion{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   moduleVersionID,
					Version:              "1",
				},
				Version:    "1.0.0",
				SHASum:     "7ae471ed18395339572f5265b835860e28a2f85016455214cb214bafe4422c7d",
				Status:     "pending",
				Submodules: []string{"submodule1"},
				Examples:   []string{"example1"},
				ModuleID:   "module-1",
			},
		},
		{
			name: "create module version returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlCreateModuleVersionPayload{
					CreateTerraformModuleVersion: graphqlCreateModuleVersionMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "module version already exists",
								Type:    internal.Conflict,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems creating module version: module version already exists",
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
			client.TerraformModuleVersion = NewTerraformModuleVersion(client)

			// Call the method being tested.
			moduleVersion, actualError := client.TerraformModuleVersion.CreateModuleVersion(ctx, &types.CreateTerraformModuleVersionInput{})

			checkError(t, test.expectErrorMessage, actualError)

			if test.expectModuleVersion != nil {
				require.NotNil(t, moduleVersion)
				assert.Equal(t, moduleVersion, test.expectModuleVersion)
			}
		})
	}
}

func TestDeleteModuleVersion(t *testing.T) {
	type graphqlDeleteModuleVersionMutation struct {
		ModuleVersion *graphQLTerraformModuleVersion `json:"moduleVersion"`
		Problems      []fakeGraphqlResponseProblem   `json:"problems"`
	}

	type graphqlDeleteModuleVersionPayload struct {
		DeleteTerraformModuleVersion graphqlDeleteModuleVersionMutation `json:"deleteTerraformModuleVersion"`
	}

	// test cases
	type testCase struct {
		responsePayload    interface{}
		name               string
		expectErrorMessage string
	}

	testCases := []testCase{
		{
			name: "Successful deletion of terraform moduleVersion",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteModuleVersionPayload{
					DeleteTerraformModuleVersion: graphqlDeleteModuleVersionMutation{},
				},
			},
		},
		{
			name: "delete module version returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteModuleVersionPayload{
					DeleteTerraformModuleVersion: graphqlDeleteModuleVersionMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "module version not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			expectErrorMessage: "problems deleting module version: module version not found",
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
			client.TerraformModuleVersion = NewTerraformModuleVersion(client)

			// Call the method being tested.
			err = client.TerraformModuleVersion.DeleteModuleVersion(ctx, &types.DeleteTerraformModuleVersionInput{})

			checkError(t, test.expectErrorMessage, err)
		})
	}
}

func TestUploadModuleVersion(t *testing.T) {
	// test cases
	type testCase struct {
		name               string
		expectErrorMessage string
		payloadToReturn    interface{}
		statusToReturn     int
	}

	testCases := []testCase{
		{
			name:           "successful module version upload",
			statusToReturn: http.StatusOK,
		},
		{
			name:               "failed module version upload",
			payloadToReturn:    fakeRESTError{Detail: "internal server error"},
			statusToReturn:     http.StatusInternalServerError,
			expectErrorMessage: "PUT request recieved http status code 500: {\"detail\":\"internal server error\"}",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			payloadBuf, err := json.Marshal(test.payloadToReturn)
			require.Nil(t, err)

			httpClient := newTestClient(func(req *http.Request) *http.Response {
				defer req.Body.Close()

				return &http.Response{
					StatusCode: test.statusToReturn,
					Body:       ioutil.NopCloser(bytes.NewReader(payloadBuf)),
					Header:     make(http.Header),
				}
			})

			client := &Client{
				httpClient: httpClient,
				cfg:        &config.Config{Endpoint: "http://test", TokenProvider: &fakeTokenProvider{token: "secret"}},
			}
			client.TerraformModuleVersion = NewTerraformModuleVersion(client)

			// Call the method being tested.
			err = client.TerraformModuleVersion.UploadModuleVersion(ctx, "module-1", strings.NewReader("test data"))

			checkError(t, test.expectErrorMessage, err)
		})
	}
}
