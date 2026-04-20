package user

import (
	"errors"

	userdomain "github.com/jrmarcello/gopherplate/internal/domain/user"
	"github.com/jrmarcello/gopherplate/internal/domain/user/vo"
	ucshared "github.com/jrmarcello/gopherplate/internal/usecases/shared"
	"github.com/jrmarcello/gopherplate/pkg/apperror"
)

// expectedErrors per use case — used by ucshared.ClassifyError to record
// semantic span attributes for expected outcomes (validation, not-found,
// conflict). Anything outside these slices is routed to FailSpan as an
// unexpected infra failure. AttrKey values come from the shared
// AttrKey* constants in internal/usecases/shared/attrkeys.go to keep the
// trace vocabulary consistent across domains.
var (
	createExpectedErrors = []ucshared.ExpectedError{
		{Err: vo.ErrInvalidEmail, AttrKey: ucshared.AttrKeyAppValidationError},
		{Err: userdomain.ErrDuplicateEmail, AttrKey: ucshared.AttrKeyAppResult, AttrValue: "duplicate_email"},
	}
	getExpectedErrors = []ucshared.ExpectedError{
		{Err: vo.ErrInvalidID, AttrKey: ucshared.AttrKeyAppValidationError},
		{Err: userdomain.ErrUserNotFound, AttrKey: ucshared.AttrKeyAppResult, AttrValue: "not_found"},
	}
	updateExpectedErrors = []ucshared.ExpectedError{
		{Err: vo.ErrInvalidID, AttrKey: ucshared.AttrKeyAppValidationError},
		{Err: vo.ErrInvalidEmail, AttrKey: ucshared.AttrKeyAppValidationError},
		{Err: userdomain.ErrUserNotFound, AttrKey: ucshared.AttrKeyAppResult, AttrValue: "not_found"},
	}
	deleteExpectedErrors = []ucshared.ExpectedError{
		{Err: vo.ErrInvalidID, AttrKey: ucshared.AttrKeyAppValidationError},
		{Err: userdomain.ErrUserNotFound, AttrKey: ucshared.AttrKeyAppResult, AttrValue: "not_found"},
	}
	// list has no expected errors — only infra errors are possible.
)

// userToAppError maps domain/validation errors to structured AppError.
// This is the single source of truth for user error translation in the use case layer.
func userToAppError(err error) *apperror.AppError {
	switch {
	case errors.Is(err, vo.ErrInvalidEmail):
		return apperror.Wrap(err, apperror.CodeInvalidRequest, "invalid email")
	case errors.Is(err, vo.ErrInvalidID):
		return apperror.Wrap(err, apperror.CodeInvalidRequest, "invalid ID")
	case errors.Is(err, userdomain.ErrUserNotFound):
		return apperror.Wrap(err, apperror.CodeNotFound, "user not found")
	case errors.Is(err, userdomain.ErrDuplicateEmail):
		return apperror.Wrap(err, apperror.CodeConflict, "email already exists")
	default:
		return apperror.Wrap(err, apperror.CodeInternalError, "internal server error")
	}
}
