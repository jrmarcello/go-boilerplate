package vo

import "errors"

// =============================================================================
// ERROS DE VALUE OBJECTS (PUROS)
// =============================================================================
//
// Estes erros são usados pelos Value Objects (CPF, Email).
// Ficam no pacote `vo` para evitar dependência circular com `person`.
//
// Uso:
//   if errors.Is(err, vo.ErrInvalidCPF) { ... }

var (
	// ErrInvalidCPF indica que o CPF informado não é válido.
	ErrInvalidCPF = errors.New("CPF inválido")

	// ErrInvalidEmail indica que o email informado não é válido.
	ErrInvalidEmail = errors.New("email inválido")

	// ErrInvalidPhone indica que o telefone informado não é válido.
	ErrInvalidPhone = errors.New("telefone inválido")
)
