package usecases

import (
	"fmt"
	"net/http"
)

// Códigos de Erro Comuns
const (
	CodeInternalError   = "INTERNAL_ERROR"
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeValidationError = "VALIDATION_ERROR"
)

// AppError é o erro base da aplicação.
// Implementa a interface error e suporta unwrapping.
type AppError struct {
	Code       string         // Código único (ex: "INVALID_CPF")
	Message    string         // Mensagem amigável para o usuário
	HTTPStatus int            // Status HTTP sugerido (400, 404, 500...)
	Details    map[string]any // Detalhes extras (campo, valor inválido, etc.)
	Err        error          // Erro original (para wrapping)
}

// Error implementa a interface error.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap permite usar errors.Is() e errors.As() com o erro original.
func (e *AppError) Unwrap() error {
	return e.Err
}

// IsDomainError retorna true se este é um erro de cliente (4xx).
func (e *AppError) IsDomainError() bool {
	return e.HTTPStatus >= 400 && e.HTTPStatus < 500
}

// WithDetails retorna uma cópia com detalhes adicionais.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// WithError retorna uma cópia com o erro original wrapped.
func (e *AppError) WithError(err error) *AppError {
	newErr := *e
	newErr.Err = err
	return &newErr
}

// =============================================================================
// Construtores
// =============================================================================

// New cria um novo AppError.
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// BadRequest cria um erro 400 Bad Request.
func BadRequest(code, message string) *AppError {
	return New(code, message, http.StatusBadRequest)
}

// NotFound cria um erro 404 Not Found.
func NotFound(code, message string) *AppError {
	return New(code, message, http.StatusNotFound)
}

// Conflict cria um erro 409 Conflict.
func Conflict(code, message string) *AppError {
	return New(code, message, http.StatusConflict)
}

// Internal cria um erro 500 Internal Server Error.
func Internal(code, message string) *AppError {
	return New(code, message, http.StatusInternalServerError)
}
