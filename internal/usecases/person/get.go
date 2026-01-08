package person

import (
	"context"
	"log/slog"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/interfaces"
)

// GetUseCase implementa o caso de uso de buscar cliente por ID.
type GetUseCase struct {
	Repo  interfaces.Repository
	Cache interfaces.Cache // opcional, pode ser nil
}

// NewGetUseCase cria uma nova instância do GetUseCase.
func NewGetUseCase(repo interfaces.Repository, cache interfaces.Cache) *GetUseCase {
	return &GetUseCase{
		Repo:  repo,
		Cache: cache,
	}
}

// Execute busca um cliente pelo ID.
//
// Fluxo com cache:
//  1. Tenta buscar no cache
//  2. Se cache miss, busca no DB
//  3. Armazena no cache para próximas requisições
func (uc *GetUseCase) Execute(ctx context.Context, input dto.GetInput) (*dto.GetOutput, error) {
	// Validar e converter ID
	id, err := vo.ParseID(input.ID)
	if err != nil {
		return nil, err
	}

	cacheKey := "person:" + input.ID

	// 1. Tentar cache primeiro
	if uc.Cache != nil {
		var cached dto.GetOutput
		if cacheErr := uc.Cache.Get(ctx, cacheKey, &cached); cacheErr == nil {
			slog.Debug("cache hit", "key", cacheKey)
			return &cached, nil
		}
	}

	// 2. Buscar no repositório (cache miss)
	entity, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Converter para DTO de saída
	output := &dto.GetOutput{
		ID:        entity.ID.String(),
		Name:      entity.Name,
		Document:  maskCPF(entity.CPF.String()),
		Phone:     entity.Phone.String(),
		Email:     entity.Email.String(),
		Active:    entity.Active,
		CreatedAt: entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
	}

	// Adicionar endereço se existir
	if !entity.Address.IsEmpty() {
		output.Address = &dto.AddressDTO{
			Street:       entity.Address.Street,
			Number:       entity.Address.Number,
			Complement:   entity.Address.Complement,
			Neighborhood: entity.Address.Neighborhood,
			City:         entity.Address.City,
			State:        entity.Address.State,
			ZipCode:      entity.Address.ZipCode,
		}
	}

	// 4. Armazenar no cache
	if uc.Cache != nil {
		if err := uc.Cache.Set(ctx, cacheKey, output); err != nil {
			slog.Warn("failed to cache person", "key", cacheKey, "error", err)
		}
	}

	return output, nil
}

// maskCPF mascara o CPF para exibição (ex: ***.***.***-25)
func maskCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return "***.***.***-" + cpf[9:11]
}
