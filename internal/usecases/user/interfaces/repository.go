package interfaces

import (
	"context"

	userdomain "bitbucket.org/appmax-space/go-boilerplate/internal/domain/user"
	"bitbucket.org/appmax-space/go-boilerplate/internal/domain/user/vo"
)

// Repository define o CONTRATO para persistência de Entity.
//
// Esta é uma INTERFACE definida na camada de USE CASES e IMPLEMENTADA na INFRASTRUCTURE.
// Isso é a essência da inversão de dependência (Dependency Inversion Principle).
//
// Benefícios:
//   - Use cases não sabem nada sobre banco de dados
//   - Fácil trocar implementação (Postgres → MySQL)
//   - Fácil criar mocks para testes
type Repository interface {
	// Create persiste uma nova Entity no banco de dados.
	Create(ctx context.Context, e *userdomain.User) error

	// FindByID busca uma Entity pelo ID (UUID v7).
	// Retorna ErrUserNotFound se não encontrar.
	FindByID(ctx context.Context, id vo.ID) (*userdomain.User, error)

	// FindByEmail busca uma Entity pelo email.
	// Retorna ErrUserNotFound se não encontrar.
	FindByEmail(ctx context.Context, email vo.Email) (*userdomain.User, error)

	// List retorna uma lista paginada de Entities com filtros opcionais.
	List(ctx context.Context, filter userdomain.ListFilter) (*userdomain.ListResult, error)

	// Update atualiza uma Entity existente.
	// Retorna ErrUserNotFound se o ID não existir.
	Update(ctx context.Context, e *userdomain.User) error

	// Delete realiza soft delete (active=false) de uma Entity.
	// Retorna ErrUserNotFound se o ID não existir.
	Delete(ctx context.Context, id vo.ID) error
}
