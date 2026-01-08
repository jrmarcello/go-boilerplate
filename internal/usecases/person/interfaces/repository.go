package interfaces

import (
	"context"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
)

// Repository define o CONTRATO para persistência de Person.
//
// Esta é uma INTERFACE definida na camada de USE CASES e IMPLEMENTADA na INFRASTRUCTURE.
// Isso é a essência da inversão de dependência (Dependency Inversion Principle):
//
//	┌─────────────┐         ┌────────────────────┐
//	│   UseCases  │◄────────│   Infrastructure   │
//	│ (Interface) │         │  (Implementação)   │
//	└─────────────┘         └────────────────────┘
//
// A camada de use cases define O QUE precisa, mas não COMO fazer.
// A infrastructure implementa o COMO (Postgres, MySQL, MongoDB, etc).
//
// Benefícios:
//   - Use cases não sabem nada sobre banco de dados
//   - Fácil trocar implementação (Postgres → MySQL)
//   - Fácil criar mocks para testes
type Repository interface {
	// Create persiste um novo Person no banco de dados.
	// Retorna erro se falhar (ex: CPF duplicado, erro de conexão).
	Create(ctx context.Context, c *person.Person) error

	// FindByID busca um Person pelo ID (ULID).
	// Retorna ErrPersonNotFound se não encontrar.
	FindByID(ctx context.Context, id vo.ID) (*person.Person, error)

	// FindByCPF busca um Person pelo CPF.
	// Retorna ErrPersonNotFound se não encontrar.
	// Útil para verificar duplicidade antes de criar.
	FindByCPF(ctx context.Context, cpf vo.CPF) (*person.Person, error)

	// List retorna uma lista paginada de Persons com filtros opcionais.
	List(ctx context.Context, filter person.ListFilter) (*person.ListResult, error)

	// Update atualiza um Person existente.
	// Retorna ErrPersonNotFound se o ID não existir.
	Update(ctx context.Context, c *person.Person) error

	// Delete realiza soft delete (active=false) de um Person.
	// Retorna ErrPersonNotFound se o ID não existir.
	Delete(ctx context.Context, id vo.ID) error
}
