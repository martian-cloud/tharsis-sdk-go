package tharsis

import (
	"errors"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// IsNotFoundError returns true if the error is a tharsis Error and contains the ErrNotFound code.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var tErr *types.Error
	return errors.As(err, &tErr) && tErr.Code == types.ErrNotFound
}

// IsConflictError returns true if the error is a tharsis Error and contains the ErrConflict code.
func IsConflictError(err error) bool {
	if err == nil {
		return false
	}

	var tErr *types.Error
	return errors.As(err, &tErr) && tErr.Code == types.ErrConflict
}
