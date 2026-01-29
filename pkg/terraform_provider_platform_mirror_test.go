package tharsis

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

func TestGetProviderPlatformMirror(t *testing.T) {
	now := time.Now().UTC()

	platformMirrorID := "platform-mirror-1"
	versionMirrorID := "version-mirror-1"

	type graphqlProviderPlatformMirrorPayloadByID struct {
		Node *graphQLTerraformProviderPlatformMirror `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		expectMirror    *types.TerraformProviderPlatformMirror
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully return platform mirror by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlProviderPlatformMirrorPayloadByID{
					Node: &graphQLTerraformProviderPlatformMirror{
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String("1"),
						},
						ID:   graphql.String(platformMirrorID),
						OS:   "windows",
						Arch: "amd64",
						VersionMirror: graphQLTerraformProviderVersionMirror{
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
			expectMirror: &types.TerraformProviderPlatformMirror{
				Metadata: types.ResourceMetadata{
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					ID:                   platformMirrorID,
					Version:              "1",
				},
				OS:   "windows",
				Arch: "amd64",
				VersionMirror: types.TerraformProviderVersionMirror{
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
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlProviderPlatformMirrorPayloadByID{},
				Errors: []fakeGraphqlResponseError{{
					Message: "an error occurred",
					Extensions: fakeGraphqlResponseErrorExtension{
						Code: "INTERNAL_SERVER_ERROR",
					},
				}},
			},
			expectErrorCode: types.ErrInternal,
		},
		{
			name: "returns nil because platform mirror does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlProviderPlatformMirrorPayloadByID{},
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
			client.TerraformProviderPlatformMirror = NewTerraformProviderPlatformMirror(client)

			// Call the method being tested.
			platformMirror, actualError := client.TerraformProviderPlatformMirror.GetProviderPlatformMirror(ctx, &types.GetTerraformProviderPlatformMirrorInput{})

			checkError(t, test.expectErrorCode, actualError)

			if test.expectMirror != nil {
				require.NotNil(t, platformMirror)
				assert.Equal(t, platformMirror, test.expectMirror)
			}
		})
	}
}

func TestGetProviderPlatformMirrorsByVersion(t *testing.T) {
	now := time.Now().UTC()

	platformMirrorID := "platform-mirror-1"
	versionMirrorID := "version-mirror-1"

	type customGraphQLVersionMirror struct {
		PlatformMirrors []graphQLTerraformProviderPlatformMirror `json:"platformMirrors"`
	}

	type graphqlProviderPlatformMirrorPayloadByVersion struct {
		Node *customGraphQLVersionMirror `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		name            string
		expectErrorCode types.ErrorCode
		expectMirrors   []types.TerraformProviderPlatformMirror
	}

	testCases := []testCase{
		{
			name: "Successfully return a list of provider platform mirrors",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlProviderPlatformMirrorPayloadByVersion{
					Node: &customGraphQLVersionMirror{
						PlatformMirrors: []graphQLTerraformProviderPlatformMirror{
							{
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String("1"),
								},
								ID:   graphql.String(platformMirrorID),
								OS:   "windows",
								Arch: "amd64",
								VersionMirror: graphQLTerraformProviderVersionMirror{
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
				},
			},
			expectMirrors: []types.TerraformProviderPlatformMirror{
				{
					Metadata: types.ResourceMetadata{
						CreationTimestamp:    &now,
						LastUpdatedTimestamp: &now,
						ID:                   platformMirrorID,
						Version:              "1",
					},
					OS:   "windows",
					Arch: "amd64",
					VersionMirror: types.TerraformProviderVersionMirror{
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
			},
		},
		{
			name:            "terraform provider version mirror does not exist",
			responsePayload: graphqlProviderPlatformMirrorPayloadByVersion{},
			expectErrorCode: types.ErrNotFound,
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlProviderPlatformMirrorPayloadByVersion{},
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
			client.TerraformProviderPlatformMirror = NewTerraformProviderPlatformMirror(client)

			// Call the method being tested.
			platformMirrors, actualError := client.TerraformProviderPlatformMirror.GetProviderPlatformMirrorsByVersion(ctx, &types.GetTerraformProviderPlatformMirrorsByVersionInput{})

			checkError(t, test.expectErrorCode, actualError)

			if len(test.expectMirrors) > 0 {
				require.NotNil(t, platformMirrors)
				assert.Equal(t, platformMirrors, test.expectMirrors)
			}
		})
	}
}

func TestDeleteProviderPlatformMirror(t *testing.T) {
	type graphqlDeleteProviderPlatformMirrorMutation struct {
		PlatformMirror graphQLTerraformProviderPlatformMirror `json:"platformMirror"`
		Problems       []fakeGraphqlResponseProblem           `json:"problems"`
	}

	type graphqlDeleteProviderPlatformMirrorPayload struct {
		DeleteTerraformProviderPlatformMirror graphqlDeleteProviderPlatformMirrorMutation `json:"deleteTerraformProviderPlatformMirror"`
	}

	type testCase struct {
		responsePayload interface{}
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successful deletion of provider platform mirror",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteProviderPlatformMirrorPayload{
					DeleteTerraformProviderPlatformMirror: graphqlDeleteProviderPlatformMirrorMutation{},
				},
			},
		},
		{
			name: "delete provider platform mirror returns a problem",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlDeleteProviderPlatformMirrorPayload{
					DeleteTerraformProviderPlatformMirror: graphqlDeleteProviderPlatformMirrorMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "provider platform mirror not found",
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
			client.TerraformProviderPlatformMirror = NewTerraformProviderPlatformMirror(client)

			// Call the method being tested.
			actualError := client.TerraformProviderPlatformMirror.DeleteProviderPlatformMirror(ctx, &types.DeleteTerraformProviderPlatformMirrorInput{})

			checkError(t, test.expectErrorCode, actualError)
		})
	}
}

func TestUploadProviderPlatformPackageToMirror(t *testing.T) {
	// test cases
	type testCase struct {
		payloadToReturn interface{}
		name            string
		expectErrorCode types.ErrorCode
		statusToReturn  int
	}

	testCases := []testCase{
		{
			name:           "successful platform package upload",
			statusToReturn: http.StatusOK,
		},
		{
			name:            "failed platform package upload",
			payloadToReturn: fakeRESTError{Detail: "internal server error"},
			statusToReturn:  http.StatusInternalServerError,
			expectErrorCode: types.ErrInternal,
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
					Body:       io.NopCloser(bytes.NewReader(payloadBuf)),
					Header:     make(http.Header),
				}
			})

			client := &Client{
				httpClient: httpClient,
				cfg:        &config.Config{Endpoint: "http://test", TokenProvider: &fakeTokenProvider{token: "secret"}},
			}
			client.TerraformProviderPlatformMirror = NewTerraformProviderPlatformMirror(client)

			input := &types.UploadProviderPlatformPackageToMirrorInput{
				Reader:          strings.NewReader("test-data"),
				VersionMirrorID: "version-mirror-1",
				OS:              "windows",
				Arch:            "arch",
			}

			// Call the method being tested.
			err = client.TerraformProviderPlatformMirror.UploadProviderPlatformPackageToMirror(ctx, input)

			checkError(t, test.expectErrorCode, err)
		})
	}
}

func TestGetProviderPlatformPackageDownloadURL(t *testing.T) {
	type testCase struct {
		name            string
		responsePayload *types.ProviderPlatformPackageInfo
		statusToReturn  int
		expectResult    *types.ProviderPlatformPackageInfo
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully get download URL",
			responsePayload: &types.ProviderPlatformPackageInfo{
				URL:    "https://example.com/download/provider.zip",
				Hashes: []string{"h1:abc123", "h1:def456"},
			},
			statusToReturn: http.StatusOK,
			expectResult: &types.ProviderPlatformPackageInfo{
				URL:    "https://example.com/download/provider.zip",
				Hashes: []string{"h1:abc123", "h1:def456"},
			},
		},
		{
			name:            "Returns error on not found",
			statusToReturn:  http.StatusNotFound,
			expectErrorCode: types.ErrNotFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			payloadBuf, err := json.Marshal(test.responsePayload)
			require.Nil(t, err)

			httpClient := newTestClient(func(_ *http.Request) *http.Response {
				return &http.Response{
					StatusCode: test.statusToReturn,
					Body:       io.NopCloser(bytes.NewReader(payloadBuf)),
					Header:     make(http.Header),
				}
			})

			client := &Client{
				httpClient: httpClient,
				cfg:        &config.Config{Endpoint: "http://test", TokenProvider: &fakeTokenProvider{token: "secret"}},
			}
			client.TerraformProviderPlatformMirror = NewTerraformProviderPlatformMirror(client)

			result, err := client.TerraformProviderPlatformMirror.GetProviderPlatformPackageDownloadURL(ctx,
				&types.GetProviderPlatformPackageDownloadURLInput{
					GroupPath:         "test-group",
					RegistryHostname:  "registry.terraform.io",
					RegistryNamespace: "hashicorp",
					Type:              "aws",
					Version:           "5.0.0",
					OS:                "linux",
					Arch:              "amd64",
				})

			checkError(t, test.expectErrorCode, err)
			if test.expectResult != nil {
				require.NotNil(t, result)
				assert.Equal(t, test.expectResult, result)
			}
		})
	}
}
