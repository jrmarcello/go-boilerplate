package user

import (
	"time"

	"bitbucket.org/appmax-space/go-boilerplate/internal/domain/user/vo"
)

// Entity é a Entidade principal (Aggregate Root) do domínio.
// Estrutura simplificada para o boilerplate.
type User struct {
	ID        vo.ID
	Name      string
	Email     vo.Email
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser cria uma nova Entity com valores padrão.
func NewUser(name string, email vo.Email) *User {
	return &User{
		ID:        vo.NewID(),
		Name:      name,
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Deactivate desativa a entity (soft delete).
func (e *User) Deactivate() {
	e.Active = false
	e.UpdatedAt = time.Now()
}

// Activate reativa a entity.
func (e *User) Activate() {
	e.Active = true
	e.UpdatedAt = time.Now()
}

// UpdateEmail atualiza o email da entity.
func (e *User) UpdateEmail(email vo.Email) {
	e.Email = email
	e.UpdatedAt = time.Now()
}

// UpdateName atualiza o nome da entity.
func (e *User) UpdateName(name string) {
	e.Name = name
	e.UpdatedAt = time.Now()
}
