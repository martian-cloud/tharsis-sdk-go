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

func TestGetCallerInfo(t *testing.T) {
	now := time.Now().UTC()

	userID := "user-id"
	userName := "test-user"
	userEmail := "test@example.com"
	userVersion := "user-version-1"

	saID := "sa-id"
	saName := "test-sa"
	saGroupPath := "group/path"
	saResourcePath := "group/path/test-sa"
	saVersion := "sa-version-1"

	type testCase struct {
		name            string
		responsePayload interface{}
		expectOutput    any
		expectErrorCode types.ErrorCode
	}

	type graphqlMeUser struct {
		Typename graphql.String `json:"__typename"`
		graphQLUser
	}

	type graphqlMeServiceAccount struct {
		Typename graphql.String `json:"__typename"`
		graphQLServiceAccount
	}

	type graphqlMePayload struct {
		Me interface{} `json:"me"`
	}

	testCases := []testCase{
		{
			name: "successfully get caller info for user",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMePayload{
					Me: graphqlMeUser{
						Typename: "User",
						graphQLUser: graphQLUser{
							ID: graphql.String(userID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(userVersion),
							},
							Username: graphql.String(userName),
							Email:    graphql.String(userEmail),
							Admin:    graphql.Boolean(false),
							Active:   graphql.Boolean(true),
						},
					},
				},
			},
			expectOutput: &types.User{
				Metadata: types.ResourceMetadata{
					ID:                   userID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              userVersion,
				},
				Username: userName,
				Email:    userEmail,
				Admin:    false,
				Active:   true,
			},
		},
		{
			name: "successfully get caller info for service account",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMePayload{
					Me: graphqlMeServiceAccount{
						Typename: "ServiceAccount",
						graphQLServiceAccount: graphQLServiceAccount{
							ID: graphql.String(saID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(saVersion),
							},
							Name:         graphql.String(saName),
							GroupPath:    graphql.String(saGroupPath),
							ResourcePath: graphql.String(saResourcePath),
							Description:  graphql.String("test description"),
						},
					},
				},
			},
			expectOutput: &types.ServiceAccount{
				Metadata: types.ResourceMetadata{
					ID:                   saID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              saVersion,
				},
				Name:              saName,
				GroupPath:         saGroupPath,
				ResourcePath:      saResourcePath,
				Description:       "test description",
				OIDCTrustPolicies: []types.OIDCTrustPolicy{},
			},
		},
		{
			name: "verify that correct error is returned",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlMePayload{},
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

			client := &Client{
				graphqlClient: newGraphQLClientForTest(testClientInput{
					statusToReturn:  http.StatusOK,
					payloadToReturn: string(payload),
				}),
			}
			client.Me = NewMe(client)

			actualOutput, actualError := client.Me.GetCallerInfo(ctx)

			checkError(t, test.expectErrorCode, actualError)

			if test.expectOutput != nil {
				require.NotNil(t, actualOutput)
				assert.Equal(t, test.expectOutput, actualOutput)
			} else {
				assert.Nil(t, actualOutput)
			}
		})
	}
}
