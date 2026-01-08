package person

import (
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
)

// Person é a Entidade principal (Aggregate Root) do domínio.
type Person struct {
	ID        vo.ID
	Name      string
	CPF       vo.CPF
	Phone     vo.Phone
	Email     vo.Email
	Address   vo.Address
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPerson(name string, cpf vo.CPF, phone vo.Phone, email vo.Email) *Person {
	return &Person{
		ID:        vo.NewID(),
		Name:      name,
		CPF:       cpf,
		Phone:     phone,
		Email:     email,
		Address:   vo.Address{},
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *Person) SetAddress(addr vo.Address) {
	c.Address = addr
	c.UpdatedAt = time.Now()
}

func (c *Person) Deactivate() {
	c.Active = false
	c.UpdatedAt = time.Now()
}

func (c *Person) Activate() {
	c.Active = true
	c.UpdatedAt = time.Now()
}

func (c *Person) UpdateEmail(email vo.Email) {
	c.Email = email
	c.UpdatedAt = time.Now()
}

func (c *Person) UpdateName(name string) {
	c.Name = name
	c.UpdatedAt = time.Now()
}

func (c *Person) UpdatePhone(phone vo.Phone) {
	c.Phone = phone
	c.UpdatedAt = time.Now()
}
