package person

import (
	"context"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/interfaces"
)

// CreateUseCase implementa o caso de uso de criação de cliente.
type CreateUseCase struct {
	Repo interfaces.Repository
}

// NewCreateUseCase cria uma nova instância do CreateUseCase.
func NewCreateUseCase(repo interfaces.Repository) *CreateUseCase {
	return &CreateUseCase{Repo: repo}
}

// Execute executa o caso de uso de criação de cliente.
//
// Fluxo:
//  1. Converte primitivos (string) para Value Objects (validação acontece aqui)
//  2. Cria a entidade Person usando a Factory
//  3. Persiste no banco via Repository
//  4. Retorna DTO com ID e timestamp
func (uc *CreateUseCase) Execute(ctx context.Context, input dto.CreateInput) (*dto.CreateOutput, error) {
	// PASSO 1: Converter primitivos para Value Objects
	cpfVO, err := vo.NewCPF(input.Document)
	if err != nil {
		return nil, err
	}

	phoneVO, err := vo.NewPhone(input.Phone)
	if err != nil {
		return nil, err
	}

	emailVO, err := vo.NewEmail(input.Email)
	if err != nil {
		return nil, err
	}

	// PASSO 2: Criar Entidade usando a Factory
	entity := person.NewPerson(input.Name, cpfVO, phoneVO, emailVO)

	// Setar endereço se fornecido
	if input.Address != nil {
		entity.SetAddress(vo.Address{
			Street:       input.Address.Street,
			Number:       input.Address.Number,
			Complement:   input.Address.Complement,
			Neighborhood: input.Address.Neighborhood,
			City:         input.Address.City,
			State:        input.Address.State,
			ZipCode:      input.Address.ZipCode,
		})
	}

	// PASSO 3: Persistir no banco via Repository
	if err := uc.Repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	// PASSO 4: Retornar Output DTO
	return &dto.CreateOutput{
		ID:        entity.ID.String(),
		CreatedAt: entity.CreatedAt.Format(time.RFC3339),
	}, nil
}
