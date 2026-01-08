package dto

// =============================================================================
// Delete Person DTOs
// =============================================================================

// DeleteInput representa os dados de entrada para deletar um cliente.
type DeleteInput struct {
	ID string `json:"id"` // ULID do cliente
}

// DeleteOutput representa a confirmação da exclusão.
type DeleteOutput struct {
	ID        string `json:"id"`
	DeletedAt string `json:"deleted_at"`
	Message   string `json:"message"`
}
