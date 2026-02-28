package handler

import (
	"errors"
	"net/http"

	entity "bitbucket.org/appmax-space/go-boilerplate/internal/domain/entity_example"
	"bitbucket.org/appmax-space/go-boilerplate/internal/domain/entity_example/vo"
	"bitbucket.org/appmax-space/go-boilerplate/pkg/apperror"
	"bitbucket.org/appmax-space/go-boilerplate/pkg/httputil"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// HandleError handles errors in a centralized and consistent way.
// It supports AppError (structured) and falls back to domain error translation.
func HandleError(c *gin.Context, span trace.Span, err error) {
	// 1. Try AppError first (structured errors from use cases)
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		span.SetStatus(codes.Error, appErr.Code)
		if appErr.HTTPStatus >= 500 {
			span.RecordError(err)
		}
		httputil.SendError(c, appErr.HTTPStatus, appErr.Message)
		return
	}

	// 2. Fallback: translate domain errors to HTTP
	status, code, message := translateError(err)

	span.SetStatus(codes.Error, code)
	if status >= 500 {
		span.RecordError(err)
	}

	httputil.SendError(c, status, message)
}

// translateError maps domain errors to HTTP status codes.
// This is the fallback for errors that are not AppError.
func translateError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, vo.ErrInvalidEmail):
		return http.StatusBadRequest, apperror.CodeInvalidRequest, "Email inválido"
	case errors.Is(err, entity.ErrEntityNotFound):
		return http.StatusNotFound, apperror.CodeNotFound, "Entity não encontrada"
	default:
		// Error with message "invalid ULID" from vo.ParseID
		if err != nil && err.Error() == "invalid ULID" {
			return http.StatusBadRequest, apperror.CodeInvalidRequest, "ID inválido"
		}
		return http.StatusInternalServerError, apperror.CodeInternalError, "Erro interno do servidor"
	}
}
