package tharsis

import (
	"bytes"
	goerrors "errors"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
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

// Dummy token provider for this test.
type dummyTokenProvider struct {
}

func (d dummyTokenProvider) GetToken() (string, error) {
	return "", nil
}

var (
	_ auth.TokenProvider = dummyTokenProvider{}
)

// testClientInput
type testClientInput struct {
	payloadToReturn string
	statusToReturn  int
}

// When developing tests that will use this fake Tharsis client, first run a successful query using GraphiQL.
// Then, capture both the query and response from GraphiQL.
// Use that to write the payload to send back from this fake Tharsis client.

// newGraphQLClientForTest returns a fake Tharsis client
// with the HTTP client replaced by a fake one.
func newGraphQLClientForTest(input testClientInput) graphqlClient {

	httpClient := newTestClient(func(req *http.Request) *http.Response {
		defer req.Body.Close()

		return &http.Response{
			StatusCode: input.statusToReturn,
			Body:       io.NopCloser(bytes.NewBufferString(input.payloadToReturn)),
			Header:     make(http.Header),
		}
	})

	return newGraphqlClientWrapper("graphql-client-url", httpClient, dummyTokenProvider{}, log.Default())
}

type fakeTokenProvider struct {
	token string
}

func (tp *fakeTokenProvider) GetToken() (string, error) {
	return tp.token, nil
}

type fakeRESTError struct {
	Detail string `json:"detail"`
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
	Extensions fakeGraphqlResponseErrorExtension `json:"extensions"`
	Path       []string                          `json:"path"`
	Locations  []fakeGraphqlResponseLocation     `json:"locations"`
}

type fakeGraphqlResponsePayload struct {
	Data   interface{}                `json:"data"`
	Errors []fakeGraphqlResponseError `json:"errors,omitempty"`
}

type fakeGraphqlResponseProblem struct {
	Message string
	Type    internal.GraphQLProblemType
	Field   []string
}

// Utility function(s):

func checkError(t *testing.T, expectCode types.ErrorCode, actualError error) {
	if expectCode == "" {
		assert.Nil(t, actualError)
	} else {
		// Uses require rather than assert to avoid a nil pointer dereference.
		require.NotNil(t, actualError)
		var tErr *types.Error
		if goerrors.As(actualError, &tErr) {
			assert.Equal(t, expectCode, tErr.Code)
		} else {
			t.Fatalf("expected tharsis error with code %s but received error: %v", expectCode, actualError)
		}
	}
}

// The End.
