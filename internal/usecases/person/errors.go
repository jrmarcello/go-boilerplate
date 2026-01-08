package person

import "bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases"

// Estes erros são retornados pelos use cases de Person.
// Usam AppError para incluir status HTTP e códigos de erro.

// Erros de validação (400 Bad Request)
var (
	ErrInvalidCPF   = usecases.BadRequest("INVALID_CPF", "CPF informado é inválido")
	ErrInvalidEmail = usecases.BadRequest("INVALID_EMAIL", "Email informado é inválido")
	ErrInvalidPhone = usecases.BadRequest("INVALID_PHONE", "Telefone informado é inválido")
	ErrInvalidID    = usecases.BadRequest("INVALID_ID", "ID informado é inválido")
)

// Erros de não encontrado (404 Not Found)
var (
	ErrPersonNotFound = usecases.NotFound("PERSON_NOT_FOUND", "Pessoa não encontrada")
)

// Erros de conflito (409 Conflict)
var (
	ErrDuplicateCPF   = usecases.Conflict("DUPLICATE_CPF", "CPF já cadastrado")
	ErrDuplicateEmail = usecases.Conflict("DUPLICATE_EMAIL", "Email já cadastrado")
)
