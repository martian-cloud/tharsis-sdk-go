package types

import (
	"fmt"
	"strings"
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

// Error represents an error returned by the Tharsis API
type Error struct {
	Err  error
	Code ErrorCode
	Msg  string
}

// Error returns the error string.
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
