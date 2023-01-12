package tharsis

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
)

// ErrorCode represents the types of errors that can be returned by the SDK
type ErrorCode string

// Error code constants
const (
	ErrInternal        ErrorCode = "internal error"
	ErrNotImplemented  ErrorCode = "not implemented"
	ErrNotFound        ErrorCode = "not found"
	ErrConflict        ErrorCode = "conflict"
	ErrOptimisticLock  ErrorCode = "optimistic lock"
	ErrForbidden       ErrorCode = "forbidden"
	ErrTooManyRequests ErrorCode = "too many requests"
	ErrUnauthorized    ErrorCode = "unauthorized"
	ErrTooLarge        ErrorCode = "request too large"
	ErrBadRequest      ErrorCode = "bad request"
)

var graphqlErrorCodeToSDKErrorCode = map[string]ErrorCode{
	"INTERNAL_SERVER_ERROR": ErrInternal,
	"BAD_REQUEST":           ErrBadRequest,
	"NOT_IMPLEMENTED":       ErrNotImplemented,
	"CONFLICT":              ErrConflict,
	"OPTIMISTIC_LOCK":       ErrOptimisticLock,
	"NOT_FOUND":             ErrNotFound,
	"FORBIDDEN":             ErrForbidden,
	"RATE_LIMIT_EXCEEDED":   ErrTooManyRequests,
	"UNAUTHENTICATED":       ErrUnauthorized,
	"UNAUTHORIZED":          ErrUnauthorized,
}

var graphqlProblemTypeToSDKErrorCode = map[internal.GraphQLProblemType]ErrorCode{
	internal.Conflict:   ErrConflict,
	internal.BadRequest: ErrBadRequest,
	internal.NotFound:   ErrNotFound,
	internal.Forbidden:  ErrForbidden,
}

var httpStatusCodeToSDKErrorCode = map[int]ErrorCode{
	http.StatusInternalServerError:   ErrInternal,
	http.StatusNotImplemented:        ErrNotFound,
	http.StatusBadRequest:            ErrBadRequest,
	http.StatusConflict:              ErrConflict,
	http.StatusNotFound:              ErrNotFound,
	http.StatusForbidden:             ErrForbidden,
	http.StatusTooManyRequests:       ErrTooManyRequests,
	http.StatusUnauthorized:          ErrUnauthorized,
	http.StatusRequestEntityTooLarge: ErrTooLarge,
}

// NotFoundError returns true if the error is a tharsis Error and contains the ErrNotFound code
func NotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var tErr *Error
	return errors.As(err, &tErr) && tErr.Code == ErrNotFound
}

// Error represents an error returned by the Tharsis API
type Error struct {
	Err  error
	Code ErrorCode
	Msg  string
}

func (e *Error) Error() string {
	if e.Msg != "" && e.Err != nil {
		var b strings.Builder
		b.WriteString(e.Msg)
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
		return b.String()
	} else if e.Msg != "" {
		return e.Msg
	} else if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("<%s>", e.Code)
}

// Unwrap unwraps an error.
func (e *Error) Unwrap() error {
	return e.Err
}

func newError(code ErrorCode, format string, args ...interface{}) error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func errorFromGraphqlError(err error) error {
	// Check if this is a graphql error
	var graphqlErrors graphql.Errors
	if errors.As(err, &graphqlErrors) && len(graphqlErrors) > 0 {
		var result error
		for _, graphqlError := range graphqlErrors {
			sdkErrorCode := ErrInternal
			if code, ok := graphqlError.Extensions["code"]; ok {
				// Attempt to map graphql error code to SDK error code
				if mappedCode, ok := graphqlErrorCodeToSDKErrorCode[code.(string)]; ok {
					sdkErrorCode = mappedCode
				}
			}
			// Use multierror here since graphql can return multiple errors
			result = multierror.Append(result, &Error{Code: sdkErrorCode, Msg: graphqlError.Message})
		}
		return result
	}
	// Return an internal server error code if this is not a graphql error type
	return &Error{Code: ErrInternal, Err: err}
}

func errorFromGraphqlProblems(problems []internal.GraphQLProblem) error {
	if len(problems) == 0 {
		return nil
	}

	var result error
	for _, problem := range problems {
		sdkErrorCode := ErrInternal
		// Attempt to map graphql problem type to SDK error code
		if mappedCode, ok := graphqlProblemTypeToSDKErrorCode[problem.Type]; ok {
			sdkErrorCode = mappedCode
		}
		// Use multierror here since a single graphql response may contain multiple problems
		result = multierror.Append(result, &Error{Code: sdkErrorCode, Msg: string(problem.Message)})
	}
	return result
}

func errorFromHTTPResponse(r *http.Response) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return newError(ErrInternal, "failed to read http response: %v", err)
	}

	sdkErrorCode := ErrInternal
	// Attempt to map graphql problem type to SDK error code
	if mappedCode, ok := httpStatusCodeToSDKErrorCode[r.StatusCode]; ok {
		sdkErrorCode = mappedCode
	}

	return newError(sdkErrorCode, "http request recieved http status code %d: %s", r.StatusCode, string(bodyBytes))
}
