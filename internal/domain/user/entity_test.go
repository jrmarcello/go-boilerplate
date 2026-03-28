package user

import (
	"testing"

	"bitbucket.org/appmax-space/go-boilerplate/internal/domain/user/vo"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	email, _ := vo.NewEmail("test@example.com")

	u := NewUser("Test Name", email)

	assert.NotEmpty(t, u.ID)
	assert.Equal(t, "Test Name", u.Name)
	assert.Equal(t, "test@example.com", u.Email.String())
	assert.True(t, u.Active)
	assert.NotZero(t, u.CreatedAt)
	assert.NotZero(t, u.UpdatedAt)
}

func TestEntity_Deactivate(t *testing.T) {
	email, _ := vo.NewEmail("test@example.com")
	u := NewUser("Test Name", email)

	u.Deactivate()

	assert.False(t, u.Active)
}

func TestEntity_Activate(t *testing.T) {
	email, _ := vo.NewEmail("test@example.com")
	u := NewUser("Test Name", email)
	u.Deactivate()

	u.Activate()

	assert.True(t, u.Active)
}

func TestEntity_UpdateName(t *testing.T) {
	email, _ := vo.NewEmail("test@example.com")
	u := NewUser("Old Name", email)

	u.UpdateName("New Name")

	assert.Equal(t, "New Name", u.Name)
}

func TestEntity_UpdateEmail(t *testing.T) {
	email, _ := vo.NewEmail("old@example.com")
	u := NewUser("Test Name", email)
	newEmail, _ := vo.NewEmail("new@example.com")

	u.UpdateEmail(newEmail)

	assert.Equal(t, "new@example.com", u.Email.String())
}
