package person

import (
	"context"
	"log/slog"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/interfaces"
)

// DeleteUseCase implementa o caso de uso de deletar cliente (soft delete).
type DeleteUseCase struct {
	Repo  interfaces.Repository
	Cache interfaces.Cache
}

// NewDeleteUseCase cria uma nova instância do DeleteUseCase.
func NewDeleteUseCase(repo interfaces.Repository, cache interfaces.Cache) *DeleteUseCase {
	return &DeleteUseCase{
		Repo:  repo,
		Cache: cache,
	}
}

// Execute realiza soft delete de um cliente (active=false).
// Cache é invalidado após delete bem-sucedido.
//
// Retorna ErrPersonNotFound se o cliente não existir ou já estiver deletado.
func (uc *DeleteUseCase) Execute(ctx context.Context, input dto.DeleteInput) (*dto.DeleteOutput, error) {
	// Validar e converter ID
	id, err := vo.ParseID(input.ID)
	if err != nil {
		return nil, err
	}

	// Realizar soft delete
	if err := uc.Repo.Delete(ctx, id); err != nil {
		return nil, err
	}

	// Invalidar cache
	if uc.Cache != nil {
		cacheKey := "person:" + input.ID
		if err := uc.Cache.Delete(ctx, cacheKey); err != nil {
			slog.Warn("failed to invalidate cache", "key", cacheKey, "error", err)
		}
	}

	return &dto.DeleteOutput{
		ID:        input.ID,
		DeletedAt: time.Now().Format(time.RFC3339),
		Message:   "Cliente desativado com sucesso",
	}, nil
}
