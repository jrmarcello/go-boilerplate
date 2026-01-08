package person

import "errors"

var (
	// Erros de Validação (Value Objects)
	ErrInvalidCPF   = errors.New("CPF inválido")
	ErrInvalidEmail = errors.New("email inválido")

	// Erros de Entidade
	ErrPersonNotFound = errors.New("cliente não encontrado")

	// Erros de Conflito
	ErrDuplicateCPF   = errors.New("CPF já cadastrado")
	ErrDuplicateEmail = errors.New("email já cadastrado")
)
