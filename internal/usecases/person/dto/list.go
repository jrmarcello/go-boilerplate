package dto

// =============================================================================
// List People DTOs
// =============================================================================

// ListInput representa os filtros e paginação para listagem.
type ListInput struct {
	Name       string `form:"name"`        // Filtro por nome (parcial)
	Email      string `form:"email"`       // Filtro por email (parcial)
	City       string `form:"city"`        // Filtro por cidade
	State      string `form:"state"`       // Filtro por estado (UF)
	ActiveOnly bool   `form:"active_only"` // Apenas ativos
	Page       int    `form:"page"`        // Página (default: 1)
	Limit      int    `form:"limit"`       // Itens por página (default: 20)
}

// ListOutput representa o resultado paginado da listagem.
type ListOutput struct {
	Data       []PersonDTO   `json:"data"`
	Pagination PaginationDTO `json:"pagination"`
}

// PersonDTO representa um cliente na listagem (dados resumidos).
type PersonDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

// PaginationDTO contém informações de paginação.
type PaginationDTO struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}
