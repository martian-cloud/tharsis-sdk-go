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

func TestGetWorkspaceVCSProviderLink(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	vpID := "workspace-vcs-provider-link-id-1"
	vpVersion := "workspace-vcs-provider-link-version-1"
	vpCreatedBy := "workspace-vcs-provider-link-created-by"
	vpWorkspaceID := "workspace-vcs-provider-link-workspace-id-1"
	vpWorkspacePath := "workspace/vcs/provider/link/workspace/path-1"
	vpVCSProviderID := "workspace-vcs-provider-link-vcs-provider-id-1"
	vpRepositoryPath := "workspace-vcs-provider-link-repository-path-1"
	vpWebhookID := "workspace-vcs-provider-link-webhook-id-1"
	vpModuleDirectory := "workspace-vcs-provider-link-module-directory-1"
	vpBranch := "workspace-vcs-provider-link-branch-1"
	vpTagRegex := "workspace-vcs-provider-link-tag-regex-1"
	vpGlobPatterns := []string{"workspace-vcs-provider-link-0", "workspace-vcs-provider-link-1"}
	vpAutoSpeculativePlan := true
	vpWebhookDisabled := false

	type graphqlWorkspaceVCSProviderLinkPayload struct {
		Node *graphQLWorkspaceVCSProviderLink `json:"node"`
	}

	// test cases
	type testCase struct {
		responsePayload                interface{}
		input                          *types.GetWorkspaceVCSProviderLinkInput
		expectWorkspaceVCSProviderLink *types.WorkspaceVCSProviderLink
		name                           string
		expectErrorCode                types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return workspace VCS provider link by ID",
			input: &types.GetWorkspaceVCSProviderLinkInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspaceVCSProviderLinkPayload{
					Node: &graphQLWorkspaceVCSProviderLink{
						ID: graphql.String(vpID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(vpVersion),
						},
						CreatedBy: graphql.String(vpCreatedBy),
						Workspace: graphQLWorkspace{
							ID:       graphql.String(vpWorkspaceID),
							FullPath: graphql.String(vpWorkspacePath),
						},
						VCSProvider: graphQLVCSProvider{
							ID: graphql.String(vpVCSProviderID),
						},
						RepositoryPath:      graphql.String(vpRepositoryPath),
						WebhookID:           graphqlStringPointerFromString(vpWebhookID),
						ModuleDirectory:     graphqlStringPointerFromString(vpModuleDirectory),
						Branch:              graphql.String(vpBranch),
						TagRegex:            graphqlStringPointerFromString(vpTagRegex),
						GlobPatterns:        graphqlStringSliceFromStrings(vpGlobPatterns),
						AutoSpeculativePlan: graphql.Boolean(vpAutoSpeculativePlan),
						WebhookDisabled:     graphql.Boolean(vpWebhookDisabled),
					},
				},
			},
			expectWorkspaceVCSProviderLink: &types.WorkspaceVCSProviderLink{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:           vpCreatedBy,
				WorkspaceID:         vpWorkspaceID,
				WorkspacePath:       vpWorkspacePath,
				VCSProviderID:       vpVCSProviderID,
				RepositoryPath:      vpRepositoryPath,
				WebhookID:           &vpWebhookID,
				ModuleDirectory:     &vpModuleDirectory,
				Branch:              vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetWorkspaceVCSProviderLinkInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspaceVCSProviderLinkPayload{},
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
			name: "query returns nil workspace VCS provider link, as if the specified workspace VCS provider link does not exist",
			input: &types.GetWorkspaceVCSProviderLinkInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspaceVCSProviderLinkPayload{},
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
			client.WorkspaceVCSProviderLink = NewWorkspaceVCSProviderLink(client)

			// Call the method being tested.
			actualWorkspaceVCSProviderLink, actualError := client.WorkspaceVCSProviderLink.GetLink(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkWorkspaceVCSProviderLink(t, test.expectWorkspaceVCSProviderLink, actualWorkspaceVCSProviderLink)
		})
	}
}

func TestCreateWorkspaceVCSProviderLink(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	vpID := "workspace-vcs-provider-link-id-1"
	vpVersion := "workspace-vcs-provider-link-version-1"
	vpCreatedBy := "workspace-vcs-provider-link-created-by"
	vpWorkspaceID := "workspace-vcs-provider-link-workspace-id-1"
	vpWorkspacePath := "workspace/vcs/provider/link/workspace/path-1"
	vpVCSProviderID := "workspace-vcs-provider-link-vcs-provider-id-1"
	vpRepositoryPath := "workspace-vcs-provider-link-repository-path-1"
	vpWebhookID := "workspace-vcs-provider-link-webhook-id-1"
	vpModuleDirectory := "workspace-vcs-provider-link-module-directory-1"
	vpBranch := "workspace-vcs-provider-link-branch-1"
	vpTagRegex := "workspace-vcs-provider-link-tag-regex-1"
	vpGlobPatterns := []string{"workspace-vcs-provider-link-0", "workspace-vcs-provider-link-1"}
	vpAutoSpeculativePlan := true
	vpWebhookDisabled := false
	vpResponseWebhookToken := "workspace-vcs-provider-link-webhook-token"
	vpResponseWebhookURL := "workspace-vcs-provider-link-webhook-url"

	type graphqlCreateWorkspaceVCSProviderLinkMutation struct {
		WebhookToken    *graphql.String
		WebhookURL      *graphql.String
		Problems        []fakeGraphqlResponseProblem    `json:"problems"`
		VCSProviderLink graphQLWorkspaceVCSProviderLink `json:"vcsProviderLink"`
	}

	type graphqlCreateWorkspaceVCSProviderLinkPayload struct {
		CreateWorkspaceVCSProviderLink graphqlCreateWorkspaceVCSProviderLinkMutation `json:"createWorkspaceVCSProviderLink"`
	}

	// test cases
	type testCase struct {
		responsePayload                interface{}
		input                          *types.CreateWorkspaceVCSProviderLinkInput
		expectWorkspaceVCSProviderLink *types.WorkspaceVCSProviderLink
		name                           string
		expectErrorCode                types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully create workspace VCS provider link",
			input: &types.CreateWorkspaceVCSProviderLinkInput{
				ModuleDirectory:     &vpModuleDirectory,
				RepositoryPath:      vpRepositoryPath,
				WorkspacePath:       vpWorkspacePath,
				ProviderID:          vpVCSProviderID,
				Branch:              &vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateWorkspaceVCSProviderLinkPayload{
					CreateWorkspaceVCSProviderLink: graphqlCreateWorkspaceVCSProviderLinkMutation{
						VCSProviderLink: graphQLWorkspaceVCSProviderLink{
							ID: graphql.String(vpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(vpVersion),
							},
							CreatedBy: graphql.String(vpCreatedBy),
							Workspace: graphQLWorkspace{
								ID:       graphql.String(vpWorkspaceID),
								FullPath: graphql.String(vpWorkspacePath),
							},
							VCSProvider: graphQLVCSProvider{
								ID: graphql.String(vpVCSProviderID),
							},
							RepositoryPath:      graphql.String(vpRepositoryPath),
							WebhookID:           graphqlStringPointerFromString(vpWebhookID),
							ModuleDirectory:     graphqlStringPointerFromString(vpModuleDirectory),
							Branch:              graphql.String(vpBranch),
							TagRegex:            graphqlStringPointerFromString(vpTagRegex),
							GlobPatterns:        graphqlStringSliceFromStrings(vpGlobPatterns),
							AutoSpeculativePlan: graphql.Boolean(vpAutoSpeculativePlan),
							WebhookDisabled:     graphql.Boolean(vpWebhookDisabled),
						},
						WebhookToken: graphqlStringPointerFromString(vpResponseWebhookToken),
						WebhookURL:   graphqlStringPointerFromString(vpResponseWebhookURL),
					},
				},
			},
			expectWorkspaceVCSProviderLink: &types.WorkspaceVCSProviderLink{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:           vpCreatedBy,
				WorkspaceID:         vpWorkspaceID,
				WorkspacePath:       vpWorkspacePath,
				VCSProviderID:       vpVCSProviderID,
				RepositoryPath:      vpRepositoryPath,
				WebhookID:           &vpWebhookID,
				ModuleDirectory:     &vpModuleDirectory,
				Branch:              vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.CreateWorkspaceVCSProviderLinkInput{
				ModuleDirectory:     &vpModuleDirectory,
				RepositoryPath:      vpRepositoryPath,
				WorkspacePath:       vpWorkspacePath,
				ProviderID:          vpVCSProviderID,
				Branch:              &vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateWorkspaceVCSProviderLinkPayload{},
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
			client.WorkspaceVCSProviderLink = NewWorkspaceVCSProviderLink(client)

			// Call the method being tested.
			actualPayload, actualError := client.WorkspaceVCSProviderLink.CreateLink(ctx, test.input)
			checkError(t, test.expectErrorCode, actualError)

			// Link is inside the payload, so check is more complex.
			assert.Equal(t, (test.expectWorkspaceVCSProviderLink == nil), (actualPayload == nil))
			if test.expectWorkspaceVCSProviderLink != nil {
				checkWorkspaceVCSProviderLink(t, test.expectWorkspaceVCSProviderLink, &actualPayload.VCSProviderLink)
				assert.NotNil(t, actualPayload.WebhookToken)
				assert.NotNil(t, actualPayload.WebhookURL)
			}
		})
	}
}

func TestUpdateWorkspaceVCSProviderLink(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	vpID := "workspace-vcs-provider-link-id-1"
	vpVersion := "workspace-vcs-provider-link-version-1"
	vpCreatedBy := "workspace-vcs-provider-link-created-by"
	vpWorkspaceID := "workspace-vcs-provider-link-workspace-id-1"
	vpWorkspacePath := "workspace/vcs/provider/link/workspace/path-1"
	vpVCSProviderID := "workspace-vcs-provider-link-vcs-provider-id-1"
	vpRepositoryPath := "workspace-vcs-provider-link-repository-path-1"
	vpWebhookID := "workspace-vcs-provider-link-webhook-id-1"
	vpModuleDirectory := "workspace-vcs-provider-link-module-directory-1"
	vpBranch := "workspace-vcs-provider-link-branch-1"
	vpTagRegex := "workspace-vcs-provider-link-tag-regex-1"
	vpGlobPatterns := []string{"workspace-vcs-provider-link-0", "workspace-vcs-provider-link-1"}
	vpAutoSpeculativePlan := true
	vpWebhookDisabled := false

	type graphqlUpdateWorkspaceVCSProviderLinkMutation struct {
		Problems        []fakeGraphqlResponseProblem    `json:"problems"`
		VCSProviderLink graphQLWorkspaceVCSProviderLink `json:"vcsProviderLink"`
	}

	type graphqlUpdateWorkspaceVCSProviderLinkPayload struct {
		UpdateWorkspaceVCSProviderLink graphqlUpdateWorkspaceVCSProviderLinkMutation `json:"updateWorkspaceVCSProviderLink"`
	}

	// test cases
	type testCase struct {
		responsePayload                interface{}
		input                          *types.UpdateWorkspaceVCSProviderLinkInput
		expectWorkspaceVCSProviderLink *types.WorkspaceVCSProviderLink
		name                           string
		expectErrorCode                types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully update workspace VCS provider link",
			input: &types.UpdateWorkspaceVCSProviderLinkInput{
				ID:                  vpID,
				ModuleDirectory:     &vpModuleDirectory,
				Branch:              &vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspaceVCSProviderLinkPayload{
					UpdateWorkspaceVCSProviderLink: graphqlUpdateWorkspaceVCSProviderLinkMutation{
						VCSProviderLink: graphQLWorkspaceVCSProviderLink{
							ID: graphql.String(vpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(vpVersion),
							},
							CreatedBy: graphql.String(vpCreatedBy),
							Workspace: graphQLWorkspace{
								ID:       graphql.String(vpWorkspaceID),
								FullPath: graphql.String(vpWorkspacePath),
							},
							VCSProvider: graphQLVCSProvider{
								ID: graphql.String(vpVCSProviderID),
							},
							RepositoryPath:      graphql.String(vpRepositoryPath),
							WebhookID:           graphqlStringPointerFromString(vpWebhookID),
							ModuleDirectory:     graphqlStringPointerFromString(vpModuleDirectory),
							Branch:              graphql.String(vpBranch),
							TagRegex:            graphqlStringPointerFromString(vpTagRegex),
							GlobPatterns:        graphqlStringSliceFromStrings(vpGlobPatterns),
							AutoSpeculativePlan: graphql.Boolean(vpAutoSpeculativePlan),
							WebhookDisabled:     graphql.Boolean(vpWebhookDisabled),
						},
					},
				},
			},
			expectWorkspaceVCSProviderLink: &types.WorkspaceVCSProviderLink{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:           vpCreatedBy,
				WorkspaceID:         vpWorkspaceID,
				WorkspacePath:       vpWorkspacePath,
				VCSProviderID:       vpVCSProviderID,
				RepositoryPath:      vpRepositoryPath,
				WebhookID:           &vpWebhookID,
				ModuleDirectory:     &vpModuleDirectory,
				Branch:              vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.UpdateWorkspaceVCSProviderLinkInput{
				ID:                  vpID,
				ModuleDirectory:     &vpModuleDirectory,
				Branch:              &vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspaceVCSProviderLinkPayload{},
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
			client.WorkspaceVCSProviderLink = NewWorkspaceVCSProviderLink(client)

			// Call the method being tested.
			actualWorkspaceVCSProviderLink, actualError := client.WorkspaceVCSProviderLink.UpdateLink(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkWorkspaceVCSProviderLink(t, test.expectWorkspaceVCSProviderLink, actualWorkspaceVCSProviderLink)
		})
	}
}

func TestDeleteWorkspaceVCSProviderLink(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	vpID := "workspace-vcs-provider-link-id-1"
	vpVersion := "workspace-vcs-provider-link-version-1"
	vpCreatedBy := "workspace-vcs-provider-link-created-by"
	vpWorkspaceID := "workspace-vcs-provider-link-workspace-id-1"
	vpWorkspacePath := "workspace/vcs/provider/link/workspace/path-1"
	vpVCSProviderID := "workspace-vcs-provider-link-vcs-provider-id-1"
	vpRepositoryPath := "workspace-vcs-provider-link-repository-path-1"
	vpWebhookID := "workspace-vcs-provider-link-webhook-id-1"
	vpModuleDirectory := "workspace-vcs-provider-link-module-directory-1"
	vpBranch := "workspace-vcs-provider-link-branch-1"
	vpTagRegex := "workspace-vcs-provider-link-tag-regex-1"
	vpGlobPatterns := []string{"workspace-vcs-provider-link-0", "workspace-vcs-provider-link-1"}
	vpAutoSpeculativePlan := true
	vpWebhookDisabled := false

	type graphqlDeleteWorkspaceVCSProviderLinkMutation struct {
		Problems        []fakeGraphqlResponseProblem    `json:"problems"`
		VCSProviderLink graphQLWorkspaceVCSProviderLink `json:"vcsProviderLink"`
	}

	type graphqlDeleteWorkspaceVCSProviderLinkPayload struct {
		DeleteWorkspaceVCSProviderLink graphqlDeleteWorkspaceVCSProviderLinkMutation `json:"deleteWorkspaceVCSProviderLink"`
	}

	// test cases
	type testCase struct {
		responsePayload                interface{}
		input                          *types.DeleteWorkspaceVCSProviderLinkInput
		expectWorkspaceVCSProviderLink *types.WorkspaceVCSProviderLink
		name                           string
		expectErrorCode                types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully delete workspace VCS provider link",
			input: &types.DeleteWorkspaceVCSProviderLinkInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteWorkspaceVCSProviderLinkPayload{
					DeleteWorkspaceVCSProviderLink: graphqlDeleteWorkspaceVCSProviderLinkMutation{
						VCSProviderLink: graphQLWorkspaceVCSProviderLink{
							ID: graphql.String(vpID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(vpVersion),
							},
							CreatedBy: graphql.String(vpCreatedBy),
							Workspace: graphQLWorkspace{
								ID:       graphql.String(vpWorkspaceID),
								FullPath: graphql.String(vpWorkspacePath),
							},
							VCSProvider: graphQLVCSProvider{
								ID: graphql.String(vpVCSProviderID),
							},
							RepositoryPath:      graphql.String(vpRepositoryPath),
							WebhookID:           graphqlStringPointerFromString(vpWebhookID),
							ModuleDirectory:     graphqlStringPointerFromString(vpModuleDirectory),
							Branch:              graphql.String(vpBranch),
							TagRegex:            graphqlStringPointerFromString(vpTagRegex),
							GlobPatterns:        graphqlStringSliceFromStrings(vpGlobPatterns),
							AutoSpeculativePlan: graphql.Boolean(vpAutoSpeculativePlan),
							WebhookDisabled:     graphql.Boolean(vpWebhookDisabled),
						},
					},
				},
			},
			expectWorkspaceVCSProviderLink: &types.WorkspaceVCSProviderLink{
				Metadata: types.ResourceMetadata{
					ID:                   vpID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              vpVersion,
				},
				CreatedBy:           vpCreatedBy,
				WorkspaceID:         vpWorkspaceID,
				WorkspacePath:       vpWorkspacePath,
				VCSProviderID:       vpVCSProviderID,
				RepositoryPath:      vpRepositoryPath,
				WebhookID:           &vpWebhookID,
				ModuleDirectory:     &vpModuleDirectory,
				Branch:              vpBranch,
				TagRegex:            &vpTagRegex,
				GlobPatterns:        vpGlobPatterns,
				AutoSpeculativePlan: vpAutoSpeculativePlan,
				WebhookDisabled:     vpWebhookDisabled,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.DeleteWorkspaceVCSProviderLinkInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteWorkspaceVCSProviderLinkPayload{},
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
			name: "query returns nil workspace VCS provider link, as if the specified workspace VCS provider link does not exist",
			input: &types.DeleteWorkspaceVCSProviderLinkInput{
				ID: vpID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteWorkspaceVCSProviderLinkPayload{
					DeleteWorkspaceVCSProviderLink: graphqlDeleteWorkspaceVCSProviderLinkMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Type:    "NOT_FOUND",
								Message: "workspace VCS provider link with ID something not found",
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
			client.WorkspaceVCSProviderLink = NewWorkspaceVCSProviderLink(client)

			// Call the method being tested.
			actualWorkspaceVCSProviderLink, actualError := client.WorkspaceVCSProviderLink.DeleteLink(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkWorkspaceVCSProviderLink(t, test.expectWorkspaceVCSProviderLink, actualWorkspaceVCSProviderLink)
		})
	}

}

// Utility functions:

func checkWorkspaceVCSProviderLink(t *testing.T, expectWorkspaceVCSProviderLink, actualWorkspaceVCSProviderLink *types.WorkspaceVCSProviderLink) {
	if expectWorkspaceVCSProviderLink != nil {
		require.NotNil(t, actualWorkspaceVCSProviderLink)
		assert.Equal(t, expectWorkspaceVCSProviderLink, actualWorkspaceVCSProviderLink)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.WorkspaceVCSProviderLink)(nil)
		assert.Equal(t, (*types.WorkspaceVCSProviderLink)(nil), actualWorkspaceVCSProviderLink)
	}
}

func graphqlStringPointerFromString(s string) *graphql.String {
	gs := graphql.String(s)
	return &gs
}

func graphqlStringSliceFromStrings(arg []string) []graphql.String {
	result := []graphql.String{}
	for _, s := range arg {
		result = append(result, graphql.String(s))
	}
	return result
}

// The End.
