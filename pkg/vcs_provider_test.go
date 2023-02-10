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

func TestGetVCSProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	vpID := "vcs-provider-id-1"
	vpVersion := "vcs-provider-version-1"
	vpCreatedBy := "vcs-provider-created-by"
	vpName := "vcs-provider-name-1"
	vpDescription := "vcs-provider-description-1"
	vpHostname := "vcs-provider-hostname"
	vpResourcePath := parentGroupName + "/" + vpName
	vpType := types.VCSProviderTypeGitlab
	vpAutoCreateWebhooks := true

	type graphqlVCSProviderPayload struct {
		Node *graphQLVCSProvider `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload   interface{}
		input             *types.GetVCSProviderInput
		expectVCSProvider *types.VCSProvider
		name              string
		expectErrorCode   ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return VCS provider by ID",
			input: &types.GetVCSProviderInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlVCSProviderPayload{
					Node: &graphQLVCSProvider{
						ID: graphql.String(vpID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(vpVersion),
						},
						CreatedBy:          graphql.String(vpCreatedBy),
						Name:               graphql.String(vpName),
						Description:        graphql.String(vpDescription),
						Hostname:           graphql.String(vpHostname),
						ResourcePath:       graphql.String(vpResourcePath),
						Type:               graphql.String(vpType),
						AutoCreateWebhooks: graphql.Boolean(vpAutoCreateWebhooks),
					},
				},
			},
			expectVCSProvider: &types.VCSProvider{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:          vpCreatedBy,
				Name:               vpName,
				Description:        vpDescription,
				Hostname:           vpHostname,
				ResourcePath:       vpResourcePath,
				Type:               vpType,
				AutoCreateWebhooks: vpAutoCreateWebhooks,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetVCSProviderInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlVCSProviderPayload{},
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
			name: "query returns nil VCS provider, as if the specified VCS provider does not exist",
			input: &types.GetVCSProviderInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlVCSProviderPayload{},
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
			client.VCSProvider = NewVCSProvider(client)

			// Call the method being tested.
			actualVCSProvider, actualError := client.VCSProvider.GetProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkVCSProvider(t, test.expectVCSProvider, actualVCSProvider)
		})
	}
}

func TestCreateVCSProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	vpID := "vcs-provider-id-1"
	vpVersion := "vcs-provider-version-1"
	vpCreatedBy := "vcs-provider-created-by-1"
	vpName := "vcs-provider-name-1"
	vpDescription := "vcs-provider-description-1"
	vpGroupPath := parentGroupName
	vpHostname := "vcs-provider-hostname-1"
	vpOAuthClientID := "vcs-provider-client-id-1"
	vpOAuthClientSecret := "vcs-provider-client-secret-1"
	vpResourcePath := vpGroupPath + "/" + vpName
	vpType := types.VCSProviderTypeGitlab
	vpAutoCreateWebhooks := true

	type graphqlCreateVCSProviderMutation struct {
		VCSProvider graphQLVCSProvider           `json:"vcsProvider"`
		Problems    []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateVCSProviderPayload struct {
		CreateVCSProvider graphqlCreateVCSProviderMutation `json:"createVCSProvider"`
	}

	// test cases
	type testCase struct {
		responsePayload   interface{}
		input             *types.CreateVCSProviderInput
		expectVCSProvider *types.VCSProvider
		name              string
		expectErrorCode   ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully create VCS provider",
			input: &types.CreateVCSProviderInput{
				Name:               vpName,
				Description:        vpDescription,
				GroupPath:          vpGroupPath,
				Hostname:           &vpHostname,
				OAuthClientID:      vpOAuthClientID,
				OAuthClientSecret:  vpOAuthClientSecret,
				Type:               vpType,
				AutoCreateWebhooks: vpAutoCreateWebhooks,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateVCSProviderPayload{
					CreateVCSProvider: graphqlCreateVCSProviderMutation{
						VCSProvider: graphQLVCSProvider{
							ID: graphql.String(vpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(vpVersion),
							},
							CreatedBy:          graphql.String(vpCreatedBy),
							Name:               graphql.String(vpName),
							Description:        graphql.String(vpDescription),
							Hostname:           graphql.String(vpHostname),
							ResourcePath:       graphql.String(vpResourcePath),
							Type:               graphql.String(vpType),
							AutoCreateWebhooks: graphql.Boolean(vpAutoCreateWebhooks),
						},
					},
				},
			},
			expectVCSProvider: &types.VCSProvider{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:          vpCreatedBy,
				Name:               vpName,
				Description:        vpDescription,
				Hostname:           vpHostname,
				ResourcePath:       vpResourcePath,
				Type:               vpType,
				AutoCreateWebhooks: vpAutoCreateWebhooks,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.CreateVCSProviderInput{
				Name:               vpName,
				Description:        vpDescription,
				GroupPath:          vpGroupPath,
				Hostname:           &vpHostname,
				OAuthClientID:      vpOAuthClientID,
				OAuthClientSecret:  vpOAuthClientSecret,
				Type:               vpType,
				AutoCreateWebhooks: vpAutoCreateWebhooks,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateVCSProviderPayload{},
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
			client.VCSProvider = NewVCSProvider(client)

			// Call the method being tested.
			actualVCSProvider, actualError := client.VCSProvider.CreateProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkVCSProvider(t, test.expectVCSProvider, actualVCSProvider)
		})
	}
}

func TestUpdateVCSProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	vpID := "vcs-provider-id-1"
	vpVersion := "vcs-provider-version-1"
	vpCreatedBy := "vcs-provider-created-by-1"
	vpName := "vcs-provider-name-1"
	vpDescription := "vcs-provider-updated-description-1"
	vpGroupPath := parentGroupName
	vpHostname := "vcs-provider-hostname-1"
	vpOAuthClientID := "vcs-provider-updated-client-id-1"
	vpOAuthClientSecret := "vcs-provider-updated-client-secret-1"
	vpResourcePath := vpGroupPath + "/" + vpName
	vpType := types.VCSProviderTypeGitlab
	vpAutoCreateWebhooks := true

	type graphqlUpdateVCSProviderMutation struct {
		VCSProvider graphQLVCSProvider           `json:"vcsProvider"`
		Problems    []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateVCSProviderPayload struct {
		UpdateVCSProvider graphqlUpdateVCSProviderMutation `json:"updateVCSProvider"`
	}

	// test cases
	type testCase struct {
		responsePayload   interface{}
		input             *types.UpdateVCSProviderInput
		expectVCSProvider *types.VCSProvider
		name              string
		expectErrorCode   ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully update VCS provider",
			input: &types.UpdateVCSProviderInput{
				ID:                vpID,
				Description:       &vpDescription,
				OAuthClientID:     &vpOAuthClientID,
				OAuthClientSecret: &vpOAuthClientSecret,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateVCSProviderPayload{
					UpdateVCSProvider: graphqlUpdateVCSProviderMutation{
						VCSProvider: graphQLVCSProvider{
							ID: graphql.String(vpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(vpVersion),
							},
							CreatedBy:          graphql.String(vpCreatedBy),
							Name:               graphql.String(vpName),
							Description:        graphql.String(vpDescription),
							Hostname:           graphql.String(vpHostname),
							ResourcePath:       graphql.String(vpResourcePath),
							Type:               graphql.String(vpType),
							AutoCreateWebhooks: graphql.Boolean(vpAutoCreateWebhooks),
						},
					},
				},
			},
			expectVCSProvider: &types.VCSProvider{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:          vpCreatedBy,
				Name:               vpName,
				Description:        vpDescription,
				Hostname:           vpHostname,
				ResourcePath:       vpResourcePath,
				Type:               vpType,
				AutoCreateWebhooks: vpAutoCreateWebhooks,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.UpdateVCSProviderInput{
				ID:                vpID,
				Description:       &vpDescription,
				OAuthClientID:     &vpOAuthClientID,
				OAuthClientSecret: &vpOAuthClientSecret,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateVCSProviderPayload{},
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
			client.VCSProvider = NewVCSProvider(client)

			// Call the method being tested.
			actualVCSProvider, actualError := client.VCSProvider.UpdateProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkVCSProvider(t, test.expectVCSProvider, actualVCSProvider)
		})
	}
}

func TestDeleteVCSProvider(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	parentGroupName := "parent-group-name"
	vpID := "vcs-provider-id-1"
	vpVersion := "vcs-provider-version-1"
	vpCreatedBy := "vcs-provider-created-by-1"
	vpName := "vcs-provider-name-1"
	vpDescription := "vcs-provider-updated-description-1"
	vpGroupPath := parentGroupName
	vpHostname := "vcs-provider-hostname-1"
	vpResourcePath := vpGroupPath + "/" + vpName
	vpType := types.VCSProviderTypeGitlab
	vpAutoCreateWebhooks := true

	type graphqlDeleteVCSProviderMutation struct {
		VCSProvider graphQLVCSProvider           `json:"vcsProvider"`
		Problems    []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteVCSProviderPayload struct {
		DeleteVCSProvider graphqlDeleteVCSProviderMutation `json:"deleteVCSProvider"`
	}

	// test cases
	type testCase struct {
		responsePayload   interface{}
		input             *types.DeleteVCSProviderInput
		expectVCSProvider *types.VCSProvider
		name              string
		expectErrorCode   ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully delete VCS provider",
			input: &types.DeleteVCSProviderInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteVCSProviderPayload{
					DeleteVCSProvider: graphqlDeleteVCSProviderMutation{
						VCSProvider: graphQLVCSProvider{
							ID: graphql.String(vpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(vpVersion),
							},
							CreatedBy:          graphql.String(vpCreatedBy),
							Name:               graphql.String(vpName),
							Description:        graphql.String(vpDescription),
							Hostname:           graphql.String(vpHostname),
							ResourcePath:       graphql.String(vpResourcePath),
							Type:               graphql.String(vpType),
							AutoCreateWebhooks: graphql.Boolean(vpAutoCreateWebhooks),
						},
					},
				},
			},
			expectVCSProvider: &types.VCSProvider{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:          vpCreatedBy,
				Name:               vpName,
				Description:        vpDescription,
				Hostname:           vpHostname,
				ResourcePath:       vpResourcePath,
				Type:               vpType,
				AutoCreateWebhooks: vpAutoCreateWebhooks,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.DeleteVCSProviderInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteVCSProviderPayload{},
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
			name: "query returns nil VCS provider, as if the specified VCS provider does not exist",
			input: &types.DeleteVCSProviderInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteVCSProviderPayload{
					DeleteVCSProvider: graphqlDeleteVCSProviderMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Type:    "NOT_FOUND",
								Message: "VCS provider with ID something not found",
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
			client.VCSProvider = NewVCSProvider(client)

			// Call the method being tested.
			actualVCSProvider, actualError := client.VCSProvider.DeleteProvider(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkVCSProvider(t, test.expectVCSProvider, actualVCSProvider)
		})
	}

}

// Utility functions:

func checkVCSProvider(t *testing.T, expectVCSProvider, actualVCSProvider *types.VCSProvider) {
	if expectVCSProvider != nil {
		require.NotNil(t, actualVCSProvider)
		assert.Equal(t, expectVCSProvider, actualVCSProvider)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.VCSProvider)(nil)
		assert.Equal(t, (*types.VCSProvider)(nil), actualVCSProvider)
	}
}

// The End.
