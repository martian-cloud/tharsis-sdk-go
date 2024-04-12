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

func TestGetTeam(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	teamName := "team-1"
	teamID := "team-id"
	teamExternalSCIMID := "team-external-SCIM-id"
	description := "a team description"
	teamVersion := "team-version-1"

	type graphqlTeamPayload struct {
		Team *graphQLTeam `json:"team"`
	}

	// test cases
	type testCase struct {
		name            string
		input           *types.GetTeamInput
		responsePayload interface{}
		expectTeam      *types.Team
		expectErrorCode types.ErrorCode
	}

	/*
		Test case template:

		name            string
		input           *types.GetTeamInput
		responsePayload interface{}
		expectTeam      *types.Team
		expectErrorCode types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "successfully get a team by name",
			input: &types.GetTeamInput{
				Name: &teamName,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlTeamPayload{
					Team: &graphQLTeam{
						ID: graphql.String(teamID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(teamVersion),
						},
						Name:           graphql.String(teamName),
						Description:    graphql.String(description),
						SCIMExternalID: graphql.String(teamExternalSCIMID),
					},
				},
			},
			expectTeam: &types.Team{
				Metadata: types.ResourceMetadata{
					ID:                   teamID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              teamVersion,
				},
				Name:           teamName,
				Description:    description,
				SCIMExternalID: teamExternalSCIMID,
			},
		},
		{
			name: "verify that correct error is returned",
			input: &types.GetTeamInput{
				Name: &teamName,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlTeamPayload{},
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

		// query returns nil team, as if the specified team does not exist.
		{
			name: "query returns nil team, as if the specified team does not exist",
			input: &types.GetTeamInput{
				Name: &teamName,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlTeamPayload{},
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
			client.Team = NewTeam(client)

			// Call the method being tested.
			actualTeam, actualError := client.Team.GetTeam(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTeam(t, test.expectTeam, actualTeam)
		})
	}
}

func TestCreateTeam(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	teamName := "team-1"
	teamID := "team-id"
	description := "a team description"
	teamVersion := "team-version-1"

	type graphqlTeamMutation struct {
		Team     graphQLTeam
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlCreateTeamPayload struct {
		CreateTeam graphqlTeamMutation `json:"createTeam"`
	}

	// test cases
	type testCase struct {
		name            string
		input           *types.CreateTeamInput
		responsePayload interface{}
		expectTeam      *types.Team
		expectErrorCode types.ErrorCode
	}

	/*
		Test case template:

		name            string
		input           *types.CreateTeamInput
		responsePayload interface{}
		expectTeam      *types.Team
		expectErrorCode types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "successfully create team",
			input: &types.CreateTeamInput{
				Name:        teamName,
				Description: description,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlCreateTeamPayload{
					CreateTeam: graphqlTeamMutation{
						Team: graphQLTeam{
							ID: graphql.String(teamID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(teamVersion),
							},
							Name:        graphql.String(teamName),
							Description: graphql.String(description),
						},
					},
				},
			},
			expectTeam: &types.Team{
				Name:        teamName,
				Description: description,
				Metadata: types.ResourceMetadata{
					ID:                   teamID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              teamVersion,
				},
			},
		},
		{
			name: "fail to create team",
			input: &types.CreateTeamInput{
				Name:        teamName,
				Description: description,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Failed to create team.",
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
			client.Team = NewTeam(client)

			// Call the method being tested.
			actualTeam, actualError := client.Team.CreateTeam(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTeam(t, test.expectTeam, actualTeam)
		})
	}
}

func TestDeleteTeam(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	teamName := "team-1"
	teamID := "team-id"
	description := "a team description"
	teamVersion := "team-version-1"

	type graphqlTeamMutation struct {
		Team     graphQLTeam
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlDeleteTeamPayload struct {
		DeleteTeam graphqlTeamMutation `json:"deleteTeam"`
	}

	// test cases
	type testCase struct {
		name            string
		input           *types.DeleteTeamInput
		responsePayload interface{}
		expectErrorCode types.ErrorCode
	}

	/*
		Test case template:

		name            string
		input           *types.DeleteTeamInput
		responsePayload interface{}
		expectErrorCode types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "successfully delete team",
			input: &types.DeleteTeamInput{
				Name: teamName,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlDeleteTeamPayload{
					DeleteTeam: graphqlTeamMutation{
						Team: graphQLTeam{
							ID: graphql.String(teamID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(teamVersion),
							},
							Name:        graphql.String(teamName),
							Description: graphql.String(description),
						},
					},
				},
			},
		},
		{
			name: "fail to delete team",
			input: &types.DeleteTeamInput{
				Name: teamName,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Failed to delete team.",
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
			client.Team = NewTeam(client)

			// Call the method being tested.
			actualError := client.Team.DeleteTeam(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
		})
	}
}

func TestAddTeamMember(t *testing.T) {
	now := time.Now().UTC() // Getting rid of local timezone makes equality checks work better.

	teamName := "team-1"
	teamID := "team-id"
	teamExternalSCIMID := "team-external-SCIM-id"
	description := "a team description"
	teamVersion := "team-version-1"
	username := "user-1"
	userID := "user-id"
	userVersion := "user-version-1"
	userEmail := "user-email"
	userAdmin := true
	userActive := false
	userExternalSCIMID := "user-external-SCIM-id"
	isMaintainer := false

	type graphqlAddUserToTeamMutation struct {
		TeamMember graphQLTeamMember
		Problems   []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlAddUserToTeamPayload struct {
		AddUserToTeam graphqlAddUserToTeamMutation `json:"addUserToTeam"`
	}

	// test cases
	type testCase struct {
		name             string
		input            *types.AddUserToTeamInput
		responsePayload  interface{}
		expectTeamMember *types.TeamMember
		expectErrorCode  types.ErrorCode
	}

	/*
		Test case template:

		name             string
		input            *types.AddUserToTeamInput
		responsePayload  interface{}
		expectTeamMember *types.TeamMember
		expectErrorCode  types.ErrorCode
	*/

	testCases := []testCase{
		{
			name: "successfully add user to team",
			input: &types.AddUserToTeamInput{
				Username:     username,
				TeamName:     teamName,
				IsMaintainer: isMaintainer,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlAddUserToTeamPayload{
					AddUserToTeam: graphqlAddUserToTeamMutation{
						TeamMember: graphQLTeamMember{
							ID: graphql.String(teamID),
							Metadata: internal.GraphQLMetadata{
								CreatedAt: &now,
								UpdatedAt: &now,
								Version:   graphql.String(teamVersion),
							},
							User: graphQLUser{
								ID: graphql.String(userID),
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String(userVersion),
								},
								Username:       graphql.String(username),
								Email:          graphql.String(userEmail),
								Admin:          graphql.Boolean(userAdmin),
								Active:         graphql.Boolean(userActive),
								SCIMExternalID: graphql.String(userExternalSCIMID),
							},
							Team: graphQLTeam{
								ID: graphql.String(teamID),
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String(teamVersion),
								},
								Name:           graphql.String(teamName),
								Description:    graphql.String(description),
								SCIMExternalID: graphql.String(teamExternalSCIMID),
							},
							IsMaintainer: graphql.Boolean(isMaintainer),
						},
					},
				},
			},
			expectTeamMember: &types.TeamMember{
				Metadata: types.ResourceMetadata{
					ID:                   teamID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              teamVersion,
				},
				UserID:       userID,
				TeamID:       teamID,
				IsMaintainer: isMaintainer,
			},
		},
		{
			name: "fail to add user to team",
			input: &types.AddUserToTeamInput{
				Username:     username,
				TeamName:     teamName,
				IsMaintainer: isMaintainer,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Errors: []fakeGraphqlResponseError{
					{
						Message: "Failed to create team.",
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
			client.Team = NewTeam(client)

			// Call the method being tested.
			actualTeamMember, actualError := client.Team.AddTeamMember(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			checkTeamMember(t, test.expectTeamMember, actualTeamMember)
		})
	}
}

// Utility functions:

func checkTeam(t *testing.T, expectTeam, actualTeam *types.Team) {
	if expectTeam != nil {
		require.NotNil(t, actualTeam)
		assert.Equal(t, expectTeam, actualTeam)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Team)(nil)
		assert.Equal(t, (*types.Team)(nil), actualTeam)
	}
}

func checkTeamMember(t *testing.T, expectTeamMember, actualTeamMember *types.TeamMember) {
	if expectTeamMember != nil {
		require.NotNil(t, actualTeamMember)
		assert.Equal(t, expectTeamMember, actualTeamMember)
	} else {
		// Plain assert.Nil reports expected <nil>, but got (*types.Team)(nil)
		assert.Equal(t, (*types.TeamMember)(nil), actualTeamMember)
	}
}
