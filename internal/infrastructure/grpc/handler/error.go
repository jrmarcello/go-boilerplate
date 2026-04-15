package handler

import (
	"errors"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jrmarcello/gopherplate/pkg/apperror"
)

// codeToGRPCStatus maps AppError codes to gRPC status codes.
var codeToGRPCStatus = map[string]codes.Code{
	apperror.CodeInvalidRequest:      codes.InvalidArgument,
	apperror.CodeValidationError:     codes.InvalidArgument,
	apperror.CodeNotFound:            codes.NotFound,
	apperror.CodeConflict:            codes.AlreadyExists,
	apperror.CodeUnauthorized:        codes.Unauthenticated,
	apperror.CodeForbidden:           codes.PermissionDenied,
	apperror.CodeUnprocessableEntity: codes.InvalidArgument,
	apperror.CodeInternalError:       codes.Internal,
}

// toGRPCStatus converts an error to a gRPC status error.
// If the error is an AppError, it maps the code to a gRPC status code.
// Otherwise, it returns codes.Internal.
func toGRPCStatus(err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		code, ok := codeToGRPCStatus[appErr.Code]
		if !ok {
			code = codes.Internal
		}
		return status.Error(code, appErr.Message)
	}
	return status.Error(codes.Internal, "internal server error")
}

// toInt32 safely converts int to int32, clamping to [0, MaxInt32].
func toInt32(v int) int32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(v) //nolint:gosec // bounds checked above
}
