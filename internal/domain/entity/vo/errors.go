package vo

import "errors"

// =============================================================================
// ERROS DE VALUE OBJECTS (PUROS)
// =============================================================================
//
// Estes erros são usados pelos Value Objects (Email).
// Ficam no pacote `vo` para evitar dependência circular com `entity`.

var (
	// ErrInvalidEmail indica que o email informado não é válido.
	ErrInvalidEmail = errors.New("email inválido")
)
