package tharsis

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetJobLogs(t *testing.T) {
	jobID := "job-id-1"
	startOffset := int32(0)
	logLimit := int32(2 * 1024 * 1024)
	expectLogs := "some-logs"
	expectLogSize := int32(1 * 1024)

	type graphQLJobLogsPayload struct {
		Job *struct {
			Logs    graphql.String `json:"logs"`
			LogSize graphql.Int
		} `json:"job"`
	}

	// test cases
	type testCase struct {
		responsePayload interface{}
		input           *types.GetJobLogsInput
		expectOutput    *types.GetJobLogsOutput
		name            string
		expectErrorCode ErrorCode
	}

	testCases := []testCase{
		{
			name: "successfully return job logs",
			input: &types.GetJobLogsInput{
				ID:          jobID,
				StartOffset: startOffset,
				Limit:       logLimit,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphQLJobLogsPayload{
					Job: &struct {
						Logs    graphql.String "json:\"logs\""
						LogSize graphql.Int
					}{
						Logs:    graphql.String(expectLogs),
						LogSize: graphql.Int(expectLogSize),
					},
				},
			},
			expectOutput: &types.GetJobLogsOutput{
				Logs:    expectLogs,
				LogSize: expectLogSize,
			},
		},
		{
			name: "verify correct error is returned",
			input: &types.GetJobLogsInput{
				ID:          jobID,
				StartOffset: startOffset,
				Limit:       logLimit,
			},
			responsePayload: fakeGraphqlResponsePayload{
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
			name: "query returns error as if job doesn't exist",
			input: &types.GetJobLogsInput{
				ID:          jobID,
				StartOffset: startOffset,
				Limit:       logLimit,
			},
			responsePayload: fakeGraphqlResponsePayload{
				Data: graphQLJobLogsPayload{},
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
			client.Job = NewJob(client)

			// Call the method being tested.
			actualOutput, actualError := client.Job.GetJobLogs(ctx, test.input)

			checkError(t, test.expectErrorCode, actualError)
			assert.Equal(t, test.expectOutput, actualOutput)
		})
	}
}
