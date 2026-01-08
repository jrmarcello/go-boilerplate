package person

import (
	"context"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// MockRepository - Mock do repositório de Person para testes unitários
// =============================================================================

// MockRepository implementa a interface Repository para testes
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, c *person.Person) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id vo.ID) (*person.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*person.Person), args.Error(1)
}

func (m *MockRepository) FindByCPF(ctx context.Context, cpf vo.CPF) (*person.Person, error) {
	args := m.Called(ctx, cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*person.Person), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, filter person.ListFilter) (*person.ListResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*person.ListResult), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, c *person.Person) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id vo.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// =============================================================================
// MockCache - Mock da interface de Cache para testes unitários
// =============================================================================

// MockCache implementa a interface Cache para testes
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
