package bootstrap

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, _, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	t.Cleanup(func() { _ = db.Close() })
	return sqlx.NewDb(db, "sqlmock")
}

func TestNew_ReturnsPopulatedContainer(t *testing.T) {
	db := newMockDB(t)

	// TC-U-01: bootstrap.New returns Container with all fields non-nil
	c := New(db, db, nil, nil)

	require.NotNil(t, c)
	assert.NotNil(t, c.Repos.User, "Repos.User")
	assert.NotNil(t, c.Repos.Role, "Repos.Role")
	assert.NotNil(t, c.UserUseCases.Create, "UserUseCases.Create")
	assert.NotNil(t, c.UserUseCases.Get, "UserUseCases.Get")
	assert.NotNil(t, c.UserUseCases.List, "UserUseCases.List")
	assert.NotNil(t, c.UserUseCases.Update, "UserUseCases.Update")
	assert.NotNil(t, c.UserUseCases.Delete, "UserUseCases.Delete")
	assert.NotNil(t, c.RoleUseCases.Create, "RoleUseCases.Create")
	assert.NotNil(t, c.RoleUseCases.List, "RoleUseCases.List")
	assert.NotNil(t, c.RoleUseCases.Delete, "RoleUseCases.Delete")
	assert.NotNil(t, c.Handlers.User, "Handlers.User")
	assert.NotNil(t, c.Handlers.Role, "Handlers.Role")
}

func TestNew_ReposPopulated(t *testing.T) {
	db := newMockDB(t)

	// TC-U-02: Container.Repos has all repositories populated
	c := New(db, db, nil, nil)

	assert.NotNil(t, c.Repos.User)
	assert.NotNil(t, c.Repos.Role)
}

func TestNew_UserUseCasesPopulated(t *testing.T) {
	db := newMockDB(t)

	// TC-U-03: Container.UserUseCases has all use cases populated
	c := New(db, db, nil, nil)

	assert.NotNil(t, c.UserUseCases.Create)
	assert.NotNil(t, c.UserUseCases.Get)
	assert.NotNil(t, c.UserUseCases.List)
	assert.NotNil(t, c.UserUseCases.Update)
	assert.NotNil(t, c.UserUseCases.Delete)
}

func TestNew_RoleUseCasesPopulated(t *testing.T) {
	db := newMockDB(t)

	// TC-U-04: Container.RoleUseCases has all use cases populated
	c := New(db, db, nil, nil)

	assert.NotNil(t, c.RoleUseCases.Create)
	assert.NotNil(t, c.RoleUseCases.List)
	assert.NotNil(t, c.RoleUseCases.Delete)
}

func TestNew_HandlersPopulated(t *testing.T) {
	db := newMockDB(t)

	// TC-U-05: Container.Handlers has all handlers populated
	c := New(db, db, nil, nil)

	assert.NotNil(t, c.Handlers.User)
	assert.NotNil(t, c.Handlers.Role)
}
