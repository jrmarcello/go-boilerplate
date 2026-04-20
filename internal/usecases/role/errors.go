package role

import (
	"errors"

	roledomain "github.com/jrmarcello/gopherplate/internal/domain/role"
	"github.com/jrmarcello/gopherplate/internal/domain/user/vo"
	ucshared "github.com/jrmarcello/gopherplate/internal/usecases/shared"
	"github.com/jrmarcello/gopherplate/pkg/apperror"
)

// Expected errors per use case — consumed by ucshared.ClassifyError to label
// the active span with a semantic attribute (key + value) instead of marking
// the span as failed. Only sentinels listed here are treated as expected
// outcomes; anything else falls through to telemetry.FailSpan.
//
// AttrKey values must come from the package-level constants in
// internal/usecases/shared/attrkeys.go — never raw string literals.
var (
	createExpectedErrors = []ucshared.ExpectedError{
		{
			Err:       roledomain.ErrDuplicateRoleName,
			AttrKey:   ucshared.AttrKeyAppResult,
			AttrValue: "duplicate_role_name",
		},
	}

	deleteExpectedErrors = []ucshared.ExpectedError{
		{
			Err:     vo.ErrInvalidID,
			AttrKey: ucshared.AttrKeyAppValidationError,
			// AttrValue intentionally empty — ClassifyError falls back to
			// err.Error() so the recorded value carries the underlying
			// validation message.
		},
		{
			Err:       roledomain.ErrRoleNotFound,
			AttrKey:   ucshared.AttrKeyAppResult,
			AttrValue: "not_found",
		},
	}
	// listExpectedErrors is intentionally nil — list only produces infra errors.
)

// roleToAppError maps domain errors to structured AppError codes.
// Unknown/infra errors are wrapped with CodeInternalError.
func roleToAppError(err error) *apperror.AppError {
	switch {
	case errors.Is(err, vo.ErrInvalidID):
		return apperror.Wrap(err, apperror.CodeInvalidRequest, "invalid role ID")
	case errors.Is(err, roledomain.ErrRoleNotFound):
		return apperror.Wrap(err, apperror.CodeNotFound, "role not found")
	case errors.Is(err, roledomain.ErrDuplicateRoleName):
		return apperror.Wrap(err, apperror.CodeConflict, "role name already exists")
	default:
		return apperror.Wrap(err, apperror.CodeInternalError, "internal server error")
	}
}
