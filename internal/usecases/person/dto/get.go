package dto

// =============================================================================
// Get Person DTOs
// =============================================================================

// GetInput representa os dados de entrada para buscar um cliente.
type GetInput struct {
	ID string `json:"id"` // ULID do cliente
}

// GetOutput representa os dados de saída do cliente encontrado.
type GetOutput struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Document  string      `json:"document"` // CPF mascarado para segurança
	Phone     string      `json:"phone"`
	Email     string      `json:"email"`
	Address   *AddressDTO `json:"address,omitempty"`
	Active    bool        `json:"active"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// AddressDTO representa o endereço do cliente.
type AddressDTO struct {
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	ZipCode      string `json:"zip_code,omitempty"`
}
