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
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetStateVersion(t *testing.T) {
	now := time.Now().UTC()

	stateVersionID := "state-version-id-1"
	stateVersionVersion := "state-version-version-1"
	outputVersion := "output-version-1"
	outputValue := "output-value-1"
	runID := "run-id-1"

	bytes, err := ctyjson.Marshal(cty.StringVal(outputValue), cty.String)
	if err != nil {
		t.Fatal(err)
	}

	typeBytes, err := ctyjson.MarshalType(cty.String)
	if err != nil {
		t.Fatal(err)
	}

	type graphqlStateVersionPayload struct {
		Node *GraphQLStateVersion `json:"node"`
	}

	type testCase struct {
		responsePayload    interface{}
		expectStateVersion *types.StateVersion
		name               string
		expectErrorCode    types.ErrorCode
	}

	testCases := []testCase{

		// positive
		{
			name: "Successfully return state-version by ID",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlStateVersionPayload{
					Node: &GraphQLStateVersion{
						ID: graphql.String(stateVersionID),
						Metadata: internal.GraphQLMetadata{
							CreatedAt: &now,
							UpdatedAt: &now,
							Version:   graphql.String(stateVersionVersion),
						},
						Run: struct{ ID graphql.String }{
							ID: graphql.String(runID),
						},
						Outputs: []GraphQLStateVersionOutput{
							{
								ID: graphql.String("some-id"),
								Metadata: internal.GraphQLMetadata{
									CreatedAt: &now,
									UpdatedAt: &now,
									Version:   graphql.String(outputVersion),
								},
								Name:  graphql.String("some-name"),
								Value: graphql.String(string(bytes)),
								Type:  graphql.String(string(typeBytes)),
							},
						},
					},
				},
			},
			expectStateVersion: &types.StateVersion{
				Metadata: types.ResourceMetadata{
					ID:                   stateVersionID,
					CreationTimestamp:    &now,
					LastUpdatedTimestamp: &now,
					Version:              stateVersionVersion,
				},
				RunID: runID,
				Outputs: []types.StateVersionOutput{
					{
						Metadata: types.ResourceMetadata{
							ID:                   "some-id",
							CreationTimestamp:    &now,
							LastUpdatedTimestamp: &now,
							Version:              outputVersion,
						},
						Name:  "some-name",
						Type:  cty.String,
						Value: cty.StringVal(outputValue),
					},
				},
			},
		},

		// query returns error as if the ID is invalid
		{
			name: "query returns error as if the ID is invalid",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlStateVersionPayload{},
				Errors: []fakeGraphqlResponseError{
					{
						Message: "ERROR: invalid input syntax for type uuid: \"invalid\n\" (SQLSTATE 22P02)",
						Path: []string{
							"stateVersion",
						},
						Extensions: fakeGraphqlResponseErrorExtension{
							Code: "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
			expectErrorCode: types.ErrInternal,
		},

		{
			name: "query returns nil state version, as if the specified state version does not exist",
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphqlStateVersionPayload{},
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
			client.StateVersion = NewStateVersion(client)

			// Call the method being tested.
			actualStateVersion, actualError := client.StateVersion.GetStateVersion(
				ctx,
				&types.GetStateVersionInput{ID: stateVersionID},
			)

			checkError(t, test.expectErrorCode, actualError)

			if test.expectStateVersion != nil {
				require.NotNil(t, actualStateVersion)
				assert.Equal(t, test.expectStateVersion, actualStateVersion)
			}
		})
	}
}
