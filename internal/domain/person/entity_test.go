package person

import (
	"testing"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"github.com/stretchr/testify/assert"
)

func TestNewPerson(t *testing.T) {
	// Arrange
	name := "John Doe"
	cpf, _ := vo.NewCPF("52998224725") // CPF válido
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("john@example.com")

	// Act
	c := NewPerson(name, cpf, phone, email)

	// Assert
	assert.NotNil(t, c)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, name, c.Name)
	assert.Equal(t, cpf, c.CPF)
	assert.Equal(t, phone, c.Phone)
	assert.Equal(t, email, c.Email)
	assert.True(t, c.Active)
	assert.NotZero(t, c.CreatedAt)
	assert.NotZero(t, c.UpdatedAt)
}
