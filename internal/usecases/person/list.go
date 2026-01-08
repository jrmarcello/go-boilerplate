package person

import (
	"context"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/interfaces"
)

// ListUseCase implementa o caso de uso de listar pessoas.
type ListUseCase struct {
	Repo interfaces.Repository
}

// NewListUseCase cria uma nova instância do ListUseCase.
func NewListUseCase(repo interfaces.Repository) *ListUseCase {
	return &ListUseCase{Repo: repo}
}

// Execute lista pessoas com filtros e paginação.
func (uc *ListUseCase) Execute(ctx context.Context, input dto.ListInput) (*dto.ListOutput, error) {
	// Converter DTO para filtro de domínio
	filter := person.ListFilter{
		Name:       input.Name,
		Email:      input.Email,
		City:       input.City,
		State:      input.State,
		ActiveOnly: input.ActiveOnly,
		Page:       input.Page,
		Limit:      input.Limit,
	}

	// Buscar no repositório
	result, err := uc.Repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Converter para DTOs de saída
	persons := make([]dto.PersonDTO, 0, len(result.Persons))
	for _, c := range result.Persons {
		persons = append(persons, dto.PersonDTO{
			ID:        c.ID.String(),
			Name:      c.Name,
			Phone:     c.Phone.String(),
			Email:     c.Email.String(),
			City:      c.Address.City,
			State:     c.Address.State,
			Active:    c.Active,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.ListOutput{
		Data: persons,
		Pagination: dto.PaginationDTO{
			Page:       result.Page,
			Limit:      result.Limit,
			Total:      result.Total,
			TotalPages: result.TotalPages(),
			HasNext:    result.HasNextPage(),
			HasPrev:    result.HasPrevPage(),
		},
	}, nil
}
