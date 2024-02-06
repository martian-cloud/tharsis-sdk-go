package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TODO: This module has unit tests only for newer method(s) added in December, 2022.
// The other methods should also have unit tests added, including a TestGetWorkspaceByPath.

func TestGetWorkspaceByID(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	workspaceID := "workspace-id-1"
	workspacePath := "fp01"
	workspaceVersion := "workspace-version-1"

	type graphqlWorkspacePayloadByID struct {
		Node *graphQLWorkspace `json:"node"`
	}

	type graphqlWorkspacePayloadByPath struct {
		Workspace *graphQLWorkspace `json:"workspace"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.GetWorkspaceInput
		expectWorkspace *types.Workspace
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return workspace by ID",
			input: &types.GetWorkspaceInput{
				ID: &workspaceID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspacePayloadByID{
					Node: &graphQLWorkspace{
						ID: graphql.String(workspaceID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(workspaceVersion),
						},
						Name:        "nm01",
						Description: "de01",
						GroupPath:   "gp01",
						FullPath:    "fp01",
					},
				},
			},
			expectWorkspace: &types.Workspace{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              workspaceVersion,
				},
				Name:        "nm01",
				Description: "de01",
				GroupPath:   "gp01",
				FullPath:    "fp01",
			},
		},

		{
			name: "Successfully return workspace by path",
			input: &types.GetWorkspaceInput{
				Path: &workspacePath,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspacePayloadByPath{
					Workspace: &graphQLWorkspace{
						ID: graphql.String(workspaceID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(workspaceVersion),
						},
						Name:        "nm01",
						Description: "de01",
						GroupPath:   "gp01",
						FullPath:    "fp01",
					},
				},
			},
			expectWorkspace: &types.Workspace{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              workspaceVersion,
				},
				Name:        "nm01",
				Description: "de01",
				GroupPath:   "gp01",
				FullPath:    "fp01",
			},
		},

		{
			name:            "returns an error since ID and path were unspecified",
			input:           &types.GetWorkspaceInput{},
			expectErrorCode: types.ErrBadRequest,
		},

		{
			name: "verify that correct error is returned",
			input: &types.GetWorkspaceInput{
				ID: ptr.String("invalid"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspacePayloadByID{},
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

		// query returns nil workspace, as if the specified workspace does not exist.
		{
			name: "query returns nil workspace, as if the specified workspace does not exist",
			input: &types.GetWorkspaceInput{
				ID: &workspaceID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlWorkspacePayloadByID{},
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
			client.Workspaces = NewWorkspaces(client)

			// Call the method being tested.
			actualWorkspace, actualError := client.Workspaces.GetWorkspace(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkWorkspace(t, test.expectWorkspace, actualWorkspace)
		})
	}
}

func TestUpdateWorkspace(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	workspaceID := "workspace-id-1"
	workspaceVersion := "workspace-version-1"
	workspaceName := "workspace-name-1"
	workspaceGroupPath := "parent-group-1"
	workspaceFullPath := workspaceGroupPath + "/" + workspaceName
	workspaceDescription := "workspace-description-1"
	workspaceTerraformVersion := "1.2.3"
	workspaceMaxJobDuration := int32(1200)
	workspacePreventDestroyPlan := true

	type graphqlUpdateWorkspaceMutation struct {
		Workspace graphQLWorkspace             `json:"workspace"`
		Problems  []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateWorkspacePayload struct {
		UpdateWorkspace graphqlUpdateWorkspaceMutation `json:"updateWorkspace"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.UpdateWorkspaceInput
		expectWorkspace *types.Workspace
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successfully update workspace by ID",
			input: &types.UpdateWorkspaceInput{
				ID:                 &workspaceID,
				Description:        workspaceDescription,
				TerraformVersion:   &workspaceTerraformVersion,
				MaxJobDuration:     &workspaceMaxJobDuration,
				PreventDestroyPlan: &workspacePreventDestroyPlan,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspacePayload{
					UpdateWorkspace: graphqlUpdateWorkspaceMutation{
						Workspace: graphQLWorkspace{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(workspaceVersion),
							},
							Name:               "nm01",
							GroupPath:          "gp01",
							FullPath:           "fp01",
							Description:        "de01",
							TerraformVersion:   "tfv01",
							MaxJobDuration:     1200,
							PreventDestroyPlan: true,
						},
					},
				},
			},
			expectWorkspace: &types.Workspace{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              workspaceVersion,
				},
				Name:               "nm01",
				GroupPath:          "gp01",
				FullPath:           "fp01",
				Description:        "de01",
				TerraformVersion:   "tfv01",
				MaxJobDuration:     1200,
				PreventDestroyPlan: true,
			},
		},
		{
			name: "Successfully update workspace by path",
			input: &types.UpdateWorkspaceInput{
				WorkspacePath:      &workspaceFullPath,
				Description:        workspaceDescription,
				TerraformVersion:   &workspaceTerraformVersion,
				MaxJobDuration:     &workspaceMaxJobDuration,
				PreventDestroyPlan: &workspacePreventDestroyPlan,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspacePayload{
					UpdateWorkspace: graphqlUpdateWorkspaceMutation{
						Workspace: graphQLWorkspace{
							ID: graphql.String(workspaceID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(workspaceVersion),
							},
							Name:               "nm01",
							GroupPath:          "gp01",
							FullPath:           "fp01",
							Description:        "de01",
							TerraformVersion:   "tfv01",
							MaxJobDuration:     1200,
							PreventDestroyPlan: true,
						},
					},
				},
			},
			expectWorkspace: &types.Workspace{
				Metadata: types.ResourceMetadata{
					ID:                   workspaceID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              workspaceVersion,
				},
				Name:               "nm01",
				GroupPath:          "gp01",
				FullPath:           "fp01",
				Description:        "de01",
				TerraformVersion:   "tfv01",
				MaxJobDuration:     1200,
				PreventDestroyPlan: true,
			},
		},
		{
			name:            "returns an error since neither ID nor path was supplied",
			input:           &types.UpdateWorkspaceInput{},
			expectErrorCode: types.ErrBadRequest,
		},
		{
			name: "returns an error since both ID and path were unspecified",
			input: &types.UpdateWorkspaceInput{
				ID:            &workspaceID,
				WorkspacePath: &workspaceFullPath,
			},
			expectErrorCode: types.ErrBadRequest,
		},
		{
			name: "verify that correct error is returned",
			input: &types.UpdateWorkspaceInput{
				ID: ptr.String("invalid"),
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspacePayload{},
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

		// query returns nil workspace, as if the specified workspace does not exist.
		{
			name: "query returns nil workspace, as if the specified workspace does not exist",
			input: &types.UpdateWorkspaceInput{
				ID: &workspaceID,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlUpdateWorkspacePayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "workspace not found",
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "NOT_FOUND",
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
			client.Workspaces = NewWorkspaces(client)

			// Call the method being tested.
			actualWorkspace, actualError := client.Workspaces.UpdateWorkspace(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkWorkspace(t, test.expectWorkspace, actualWorkspace)
		})
	}
}

// Utility functions:

func checkWorkspace(t *testing.T, expectWorkspace, actualWorkspace *types.Workspace) {
	if expectWorkspace != nil {
		require.NotNil(t, actualWorkspace)
		assert.Equal(t, expectWorkspace, actualWorkspace)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Workspace)(nil)
		assert.Equal(t, (*types.Workspace)(nil), actualWorkspace)
	}
}
