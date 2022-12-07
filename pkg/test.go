package tharsis

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
)

// See http://hassansin.github.io/Unit-Testing-http-client-in-Go#2-by-replacing-httptransport
// as the basis for the fake client and related functions.

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

// testClientInput
type testClientInput struct {
	statusToReturn  int
	payloadToReturn string
}

// When developing tests that will use this fake Tharsis client, first run a successful query using GraphiQL.
// Then, capture both the query and response from GraphiQL.
// Use that to write the payload to send back from this fake Tharsis client.

// newGraphQLClientForTest returns a fake Tharsis client
// with the HTTP client replaced by a fake one.
func newGraphQLClientForTest(input testClientInput) *graphql.Client {

	httpClient := newTestClient(func(req *http.Request) *http.Response {
		defer req.Body.Close()

		return &http.Response{
			StatusCode: input.statusToReturn,
			Body:       ioutil.NopCloser(bytes.NewBufferString(input.payloadToReturn)),
			Header:     make(http.Header),
		}
	})

	return graphql.NewClient("graphql-client-url", httpClient)
}

// Types for building response payloads to be returned by the fake http transport.
type fakeGraphqlResponseErrorExtension struct {
	Code string `json:"code"`
}

type fakeGraphqlResponseLocation struct {
	Line   int
	Column int
}

// Actual responses seen in GraphiQL have path and locations elements in the error response.
type fakeGraphqlResponseError struct {
	Message    string                            `json:"message"`
	Path       []string                          `json:"path"`
	Locations  []fakeGraphqlResponseLocation     `json:"locations"`
	Extensions fakeGraphqlResponseErrorExtension `json:"extensions"`
}

type fakeGraphqlResponsePayload struct {
	Data   interface{}                `json:"data"`
	Errors []fakeGraphqlResponseError `json:"errors,omitempty"`
}

type fakeGraphqlResponseProblem struct {
	Message string
	Field   []string
	Type    internal.GraphQLProblemType
}

// Utility function(s):

func checkError(t *testing.T, expectedMsg string, actualError error) {
	if expectedMsg == "" {
		assert.Nil(t, actualError)
	} else {
		// Uses require rather than assert to avoid a nil pointer dereference.
		require.NotNil(t, actualError)
		assert.Equal(t, expectedMsg, actualError.Error())
	}
}

// The End.
