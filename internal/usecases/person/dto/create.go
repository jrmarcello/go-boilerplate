package dto

// =============================================================================
// Create Person DTOs
// =============================================================================

// CreateInput representa os dados de entrada para criação de cliente.
type CreateInput struct {
	Name     string      `json:"name"`              // Nome do cliente
	Document string      `json:"document"`          // CPF (será validado no UseCase)
	Phone    string      `json:"phone"`             // Telefone (será validado no UseCase)
	Email    string      `json:"email"`             // Email (será validado no UseCase)
	Address  *AddressDTO `json:"address,omitempty"` // Endereço (opcional)
}

// CreateOutput representa os dados de saída após criação.
type CreateOutput struct {
	ID        string `json:"id"`         // ID gerado (ULID)
	CreatedAt string `json:"created_at"` // Timestamp no formato RFC3339
}
