package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestSetVariablesIncludedInTFConfig(t *testing.T) {
	runID := "run-1"

	type graphqlSetVariablesIncludedInTFConfigMutation struct {
		Problems []fakeGraphqlResponseProblem `json:"problems"`
	}

	type graphqlUpdateRunVariablesUsagePayload struct {
		SetVariablesIncludedInTFConfig graphqlSetVariablesIncludedInTFConfigMutation `json:"setVariablesIncludedInTFConfig"`
	}

	type testCase struct {
		responsePayload interface{}
		input           *types.SetVariablesIncludedInTFConfigInput
		name            string
		expectErrorCode types.ErrorCode
	}

	testCases := []testCase{
		{
			name: "Successful",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateRunVariablesUsagePayload{
					SetVariablesIncludedInTFConfig: graphqlSetVariablesIncludedInTFConfigMutation{},
				},
			},
			input: &types.SetVariablesIncludedInTFConfigInput{
				RunID:        runID,
				VariableKeys: []string{"my_var"},
			},
		},
		{
			name: "Failed",
			responsePayload: &fakeGraphqlResponsePayload{
				Data: graphqlUpdateRunVariablesUsagePayload{
					SetVariablesIncludedInTFConfig: graphqlSetVariablesIncludedInTFConfigMutation{
						Problems: []fakeGraphqlResponseProblem{
							{
								Message: "run not found",
								Type:    internal.NotFound,
								Field:   []string{},
							},
						},
					},
				},
			},
			input:           &types.SetVariablesIncludedInTFConfigInput{},
			expectErrorCode: types.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			payload, err := json.Marshal(&tc.responsePayload)
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
			client.Run = NewRun(client)

			// Call the method being tested.
			actualError := client.Run.SetVariablesIncludedInTFConfig(ctx, tc.input)
			checkError(t, tc.expectErrorCode, actualError)
		})
	}
}
