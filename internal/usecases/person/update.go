package person

import (
	"context"
	"log/slog"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/interfaces"
)

// UpdateUseCase implementa o caso de uso de atualizar cliente.
type UpdateUseCase struct {
	Repo  interfaces.Repository
	Cache interfaces.Cache
}

// NewUpdateUseCase cria uma nova instância do UpdateUseCase.
func NewUpdateUseCase(repo interfaces.Repository, cache interfaces.Cache) *UpdateUseCase {
	return &UpdateUseCase{
		Repo:  repo,
		Cache: cache,
	}
}

// Execute atualiza um cliente existente.
//
// CPF não pode ser alterado (identificador único do cliente).
// Apenas campos informados serão atualizados.
// Cache é invalidado após update bem-sucedido.
//
// Retorna ErrPersonNotFound se o cliente não existir.
func (uc *UpdateUseCase) Execute(ctx context.Context, input dto.UpdateInput) (*dto.UpdateOutput, error) {
	// Validar e converter ID
	id, err := vo.ParseID(input.ID)
	if err != nil {
		return nil, err
	}

	// Buscar cliente atual
	entity, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Atualizar campos informados
	if input.Name != nil && *input.Name != "" {
		entity.UpdateName(*input.Name)
	}

	if input.Phone != nil && *input.Phone != "" {
		phone, err := vo.NewPhone(*input.Phone)
		if err != nil {
			return nil, err
		}
		entity.UpdatePhone(phone)
	}

	if input.Email != nil && *input.Email != "" {
		email, err := vo.NewEmail(*input.Email)
		if err != nil {
			return nil, err
		}
		entity.UpdateEmail(email)
	}

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

	// Persistir alterações
	if err := uc.Repo.Update(ctx, entity); err != nil {
		return nil, err
	}

	// Invalidar cache
	if uc.Cache != nil {
		cacheKey := "person:" + input.ID
		if err := uc.Cache.Delete(ctx, cacheKey); err != nil {
			slog.Warn("failed to invalidate cache", "key", cacheKey, "error", err)
		}
	}

	return &dto.UpdateOutput{
		ID:        entity.ID.String(),
		Name:      entity.Name,
		Phone:     entity.Phone.String(),
		Email:     entity.Email.String(),
		UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
	}, nil
}
