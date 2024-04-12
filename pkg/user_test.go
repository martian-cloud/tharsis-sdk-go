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

// TODO: This module has unit tests only for newer method(s) added in April 2024.
// The other methods should also have unit tests added.

func TestGetUsers(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	userName := "user-1"
	userID := "user-id"
	userEmail := "user-email@example.invalid"
	userExternalSCIMID := "user-external-SCIM-id"
	userIsAdmin := false
	userIsActive := true
	userVersion := "user-version-1"

	paginationSort := types.UserSortableFieldUpdatedAtDesc
	paginationLimit := int32(50)
	paginationCursor := "pagination-cursor"

	// test cases
	type testCase struct {
		name            string
		input           *types.GetUsersInput
		responsePayload interface{}
		expectOutput    *types.GetUsersOutput
		expectErrorCode types.ErrorCode
	}

	/*
		Test case template:

		name            string
		input           *types.GetUsersInput
		responsePayload interface{}
		expectOutput    *types.GetUsersOutput
		expectErrorCode types.ErrorCode
	*/

	type graphqlPageInfo struct {
		HasNextPage graphql.Boolean `json:"hasNextPage"`
	}

	type graphqlUserEdge struct {
		Node graphQLUser `json:"node"`
	}

	type graphqlUserMiddle struct {
		Edges      []graphqlUserEdge `json:"edges"`
		PageInfo   graphqlPageInfo   `json:"pageInfo"`
		TotalCount graphql.Int       `json:"totalCount"`
	}

	type graphqlGetUsersPayload struct {
		Users graphqlUserMiddle `json:"users"`
	}

	testCases := []testCase{
		{
			name: "successfully get a user by name",
			input: &types.GetUsersInput{
				Sort: &paginationSort,
				PaginationOptions: &types.PaginationOptions{
					Limit:  &paginationLimit,
					Cursor: &paginationCursor,
				},
				Filter: &types.UserFilter{
					Search: &userName,
				},
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGetUsersPayload{
					Users: graphqlUserMiddle{
						TotalCount: 1,
						PageInfo: graphqlPageInfo{
							HasNextPage: graphql.Boolean(false),
						},
						Edges: []graphqlUserEdge{
							{
								Node: graphQLUser{
									ID: graphql.String(userID),
									Metadata: internal.GraphQLMetadata{
										CreatedAt: &now,
										UpdatedAt: &now,
										Version:   graphql.String(userVersion),
									},
									Username:       graphql.String(userName),
									Email:          graphql.String(userEmail),
									SCIMExternalID: graphql.String(userExternalSCIMID),
									Admin:          graphql.Boolean(userIsAdmin),
									Active:         graphql.Boolean(userIsActive),
								},
							},
						},
					},
				},
			},
			expectOutput: &types.GetUsersOutput{
				PageInfo: &types.PageInfo{
					TotalCount:  1,
					HasNextPage: false,
				},
				Users: []types.User{
					{
						Metadata: types.ResourceMetadata{
							ID:                   userID,
							CreationTimestamp:    &now,
							LastUpdatedTimestamp: &now,
							Version:              userVersion,
						},
						Username:       userName,
						Email:          userEmail,
						SCIMExternalID: userExternalSCIMID,
						Admin:          userIsAdmin,
						Active:         userIsActive,
					},
				},
			},
		},
		{
			name: "query returns empty, as if the specified users do not exist",
			input: &types.GetUsersInput{
				Sort: &paginationSort,
				PaginationOptions: &types.PaginationOptions{
					Limit:  &paginationLimit,
					Cursor: &paginationCursor,
				},
				Filter: &types.UserFilter{
					Search: &userName,
				},
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGetUsersPayload{},
			},
			expectOutput: &types.GetUsersOutput{
				PageInfo: &types.PageInfo{
					TotalCount:  0,
					HasNextPage: false,
				},
				Users: []types.User{},
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetUsersInput{
				Sort: &paginationSort,
				PaginationOptions: &types.PaginationOptions{
					Limit:  &paginationLimit,
					Cursor: &paginationCursor,
				},
				Filter: &types.UserFilter{
					Search: &userName,
				},
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlGetUsersPayload{},
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
			client.User = NewUser(client)

			// Call the method being tested.
			actualOutput, actualError := client.User.GetUsers(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkGetUsersOutput(t, test.expectOutput, actualOutput)
		})
	}
}

func checkGetUsersOutput(t *testing.T, expectOutput, actualOutput *types.GetUsersOutput) {
	if expectOutput != nil {
		require.NotNil(t, actualOutput)
		assert.Equal(t, expectOutput.PageInfo.TotalCount, actualOutput.PageInfo.TotalCount)
		assert.Equal(t, expectOutput.PageInfo.HasNextPage, actualOutput.PageInfo.HasNextPage)
		assert.ElementsMatch(t, expectOutput.Users, actualOutput.Users)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.GetUsersOutput)(nil)
		assert.Equal(t, (*types.GetUsersOutput)(nil), actualOutput)
	}
}
