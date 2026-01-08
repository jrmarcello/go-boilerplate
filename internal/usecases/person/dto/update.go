package dto

// =============================================================================
// Update Person DTOs
// =============================================================================

// UpdateInput representa os dados de entrada para atualizar um cliente.
// CPF não pode ser alterado (identificador único).
type UpdateInput struct {
	ID      string            `json:"id"`                // ULID do cliente (obrigatório)
	Name    *string           `json:"name,omitempty"`    // Nome (opcional)
	Phone   *string           `json:"phone,omitempty"`   // Telefone (opcional)
	Email   *string           `json:"email,omitempty"`   // Email (opcional)
	Address *UpdateAddressDTO `json:"address,omitempty"` // Endereço (opcional)
}

// UpdateAddressDTO representa o endereço para atualização.
type UpdateAddressDTO struct {
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	ZipCode      string `json:"zip_code,omitempty"`
}

// UpdateOutput representa os dados do cliente atualizado.
type UpdateOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	UpdatedAt string `json:"updated_at"`
}
