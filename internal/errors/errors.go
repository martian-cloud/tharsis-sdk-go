// Package errors contains the mappings and functions that
// standardize errors returned from the API to SDK's clients.
package errors

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-multierror"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

var graphqlErrorCodeToSDKErrorCode = map[string]types.ErrorCode{
	"INTERNAL_SERVER_ERROR": types.ErrInternal,
	"BAD_REQUEST":           types.ErrBadRequest,
	"NOT_IMPLEMENTED":       types.ErrNotImplemented,
	"CONFLICT":              types.ErrConflict,
	"OPTIMISTIC_LOCK":       types.ErrOptimisticLock,
	"NOT_FOUND":             types.ErrNotFound,
	"FORBIDDEN":             types.ErrForbidden,
	"RATE_LIMIT_EXCEEDED":   types.ErrTooManyRequests,
	"UNAUTHENTICATED":       types.ErrUnauthorized,
	"UNAUTHORIZED":          types.ErrUnauthorized,
	"SERVICE_UNAVAILABLE":   types.ErrServiceUnavailable,
}

var graphqlProblemTypeToSDKErrorCode = map[internal.GraphQLProblemType]types.ErrorCode{
	internal.Conflict:           types.ErrConflict,
	internal.BadRequest:         types.ErrBadRequest,
	internal.NotFound:           types.ErrNotFound,
	internal.Forbidden:          types.ErrForbidden,
	internal.ServiceUnavailable: types.ErrServiceUnavailable,
}

var httpStatusCodeToSDKErrorCode = map[int]types.ErrorCode{
	http.StatusInternalServerError:   types.ErrInternal,
	http.StatusNotImplemented:        types.ErrNotFound,
	http.StatusBadRequest:            types.ErrBadRequest,
	http.StatusConflict:              types.ErrConflict,
	http.StatusNotFound:              types.ErrNotFound,
	http.StatusForbidden:             types.ErrForbidden,
	http.StatusTooManyRequests:       types.ErrTooManyRequests,
	http.StatusUnauthorized:          types.ErrUnauthorized,
	http.StatusRequestEntityTooLarge: types.ErrTooLarge,
	http.StatusServiceUnavailable:    types.ErrServiceUnavailable,
}

// NewError returns a new Error.
func NewError(code types.ErrorCode, format string, args ...interface{}) error {
	return &types.Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// ErrorFromGraphqlError returns an SDK error type from a GraphQL error.
func ErrorFromGraphqlError(err error) error {
	// Check if this is a graphql error
	var graphqlErrors graphql.Errors
	if errors.As(err, &graphqlErrors) && len(graphqlErrors) > 0 {
		var result error
		for _, graphqlError := range graphqlErrors {
			sdkErrorCode := types.ErrInternal
			if code, ok := graphqlError.Extensions["code"]; ok {
				// Attempt to map graphql error code to SDK error code
				if mappedCode, ok := graphqlErrorCodeToSDKErrorCode[code.(string)]; ok {
					sdkErrorCode = mappedCode
				}
			}
			// Use multierror here since graphql can return multiple errors
			result = multierror.Append(result, &types.Error{Code: sdkErrorCode, Msg: graphqlError.Message})
		}
		return result
	}
	// Return an internal server error code if this is not a graphql error type
	return &types.Error{Code: types.ErrInternal, Err: err}
}

// ErrorFromGraphqlProblems returns an SDK error from GraphQL problems.
func ErrorFromGraphqlProblems(problems []internal.GraphQLProblem) error {
	if len(problems) == 0 {
		return nil
	}

	var result error
	for _, problem := range problems {
		sdkErrorCode := types.ErrInternal
		// Attempt to map graphql problem type to SDK error code
		if mappedCode, ok := graphqlProblemTypeToSDKErrorCode[problem.Type]; ok {
			sdkErrorCode = mappedCode
		}
		// Use multierror here since a single graphql response may contain multiple problems
		result = multierror.Append(result, &types.Error{Code: sdkErrorCode, Msg: string(problem.Message)})
	}
	return result
}

// ErrorFromHTTPResponse returns an SDK error from an HTTP response.
func ErrorFromHTTPResponse(r *http.Response) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return NewError(types.ErrInternal, "failed to read http response: %v", err)
	}

	sdkErrorCode := types.ErrInternal
	// Attempt to map graphql problem type to SDK error code
	if mappedCode, ok := httpStatusCodeToSDKErrorCode[r.StatusCode]; ok {
		sdkErrorCode = mappedCode
	}

	return NewError(sdkErrorCode, "http request received http status code %d: %s", r.StatusCode, string(bodyBytes))
}
