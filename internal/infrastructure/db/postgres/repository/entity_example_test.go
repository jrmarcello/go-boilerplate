package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	entity "bitbucket.org/appmax-space/go-boilerplate/internal/domain/entity_example"
	"bitbucket.org/appmax-space/go-boilerplate/internal/domain/entity_example/vo"
	"bitbucket.org/appmax-space/go-boilerplate/pkg/database"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Unit Tests for internal conversions (não precisam de banco)
// =============================================================================

func TestEntityDB_ToEntity_Success(t *testing.T) {
	// Arrange
	now := time.Now().Truncate(time.Microsecond)
	dbModel := entityDB{
		ID:        "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:      "João Silva",
		Email:     "joao@example.com",
		Active:    true,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "01ARZ3NDEKTSV4RRFFQ69G5FAV", entity.ID.String())
	assert.Equal(t, "João Silva", entity.Name)
	assert.Equal(t, "joao@example.com", entity.Email.String())
	assert.True(t, entity.Active)
}

func TestEntityDB_ToEntity_InvalidID(t *testing.T) {
	// Arrange
	dbModel := entityDB{
		ID:    "invalid-id",
		Name:  "Test",
		Email: "test@example.com",
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "parsing ID")
}

func TestEntityDB_ToEntity_InvalidEmail(t *testing.T) {
	// Arrange
	dbModel := entityDB{
		ID:    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:  "Test",
		Email: "invalid-email",
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "parsing email")
}

func TestFromDomainEntity(t *testing.T) {
	// Arrange
	email, _ := vo.NewEmail("joao@example.com")
	now := time.Now().Truncate(time.Microsecond)

	domainEntity := &entity.Entity{
		ID:        vo.NewID(),
		Name:      "João Silva",
		Email:     email,
		Active:    true,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}

	// Act
	dbModel := fromDomainEntity(domainEntity)

	// Assert
	assert.Equal(t, domainEntity.ID.String(), dbModel.ID)
	assert.Equal(t, domainEntity.Name, dbModel.Name)
	assert.Equal(t, domainEntity.Email.String(), dbModel.Email)
	assert.Equal(t, domainEntity.Active, dbModel.Active)
	assert.Equal(t, domainEntity.CreatedAt, dbModel.CreatedAt)
	assert.Equal(t, domainEntity.UpdatedAt, dbModel.UpdatedAt)
}

func TestFromDomainEntity_InactiveEntity(t *testing.T) {
	// Arrange
	email, _ := vo.NewEmail("inactive@example.com")

	domainEntity := &entity.Entity{
		ID:        vo.NewID(),
		Name:      "Inactive User",
		Email:     email,
		Active:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	dbModel := fromDomainEntity(domainEntity)

	// Assert
	assert.False(t, dbModel.Active)
}

func TestFromDomainEntity_RoundTrip(t *testing.T) {
	// Teste que podemos converter entity -> dbModel -> entity sem perda de dados
	email, _ := vo.NewEmail("roundtrip@example.com")
	now := time.Now().Truncate(time.Microsecond)

	original := &entity.Entity{
		ID:        vo.NewID(),
		Name:      "Round Trip Test",
		Email:     email,
		Active:    true,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}

	// Convert to DB model
	dbModel := fromDomainEntity(original)

	// Convert back to entity
	restored, err := dbModel.toEntity()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Email.String(), restored.Email.String())
	assert.Equal(t, original.Active, restored.Active)
	// Timestamps devem ser iguais quando truncados para microseconds (Postgres precision)
	assert.Equal(t, original.CreatedAt, restored.CreatedAt)
	assert.Equal(t, original.UpdatedAt, restored.UpdatedAt)
}

// =============================================================================
// Helpers for sqlmock tests
// =============================================================================

func buildTestEntity() *entity.Entity {
	email, _ := vo.NewEmail("test@example.com")
	now := time.Now().Truncate(time.Microsecond)

	return &entity.Entity{
		ID:        vo.NewID(),
		Name:      "Test Entity",
		Email:     email,
		Active:    true,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}
}

// =============================================================================
// Unit Tests for EntityRepository with sqlmock
// =============================================================================

// --- Create ------------------------------------------------------------------

func TestEntityRepository_Create(t *testing.T) {
	db, mock, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	cluster := database.NewDBClusterFromDB(sqlxDB)
	repo := NewEntityRepository(cluster)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO entities").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		e := buildTestEntity()
		createErr := repo.Create(context.Background(), e)

		assert.NoError(t, createErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO entities").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		e := buildTestEntity()
		createErr := repo.Create(context.Background(), e)

		assert.Error(t, createErr)
		assert.ErrorIs(t, createErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// --- FindByID ----------------------------------------------------------------

func TestEntityRepository_FindByID(t *testing.T) {
	db, mock, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	cluster := database.NewDBClusterFromDB(sqlxDB)
	repo := NewEntityRepository(cluster)

	now := time.Now().Truncate(time.Microsecond)
	testID := vo.NewID()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"}).
			AddRow(testID.String(), "Test Entity", "test@example.com", true, now, now)

		mock.ExpectQuery("SELECT .+ FROM entities WHERE id").
			WithArgs(testID.String()).
			WillReturnRows(rows)

		result, findErr := repo.FindByID(context.Background(), testID)

		assert.NoError(t, findErr)
		require.NotNil(t, result)
		assert.Equal(t, testID, result.ID)
		assert.Equal(t, "Test Entity", result.Name)
		assert.Equal(t, "test@example.com", result.Email.String())
		assert.True(t, result.Active)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM entities WHERE id").
			WithArgs(testID.String()).
			WillReturnError(sql.ErrNoRows)

		result, findErr := repo.FindByID(context.Background(), testID)

		assert.Nil(t, result)
		assert.ErrorIs(t, findErr, entity.ErrEntityNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM entities WHERE id").
			WithArgs(testID.String()).
			WillReturnError(sql.ErrConnDone)

		result, findErr := repo.FindByID(context.Background(), testID)

		assert.Nil(t, result)
		assert.Error(t, findErr)
		assert.ErrorIs(t, findErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// --- FindByEmail -------------------------------------------------------------

func TestEntityRepository_FindByEmail(t *testing.T) {
	db, mock, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	cluster := database.NewDBClusterFromDB(sqlxDB)
	repo := NewEntityRepository(cluster)

	now := time.Now().Truncate(time.Microsecond)
	testID := vo.NewID()
	testEmail, _ := vo.NewEmail("test@example.com")

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"}).
			AddRow(testID.String(), "Test Entity", "test@example.com", true, now, now)

		mock.ExpectQuery("SELECT .+ FROM entities WHERE email").
			WithArgs(testEmail.String()).
			WillReturnRows(rows)

		result, findErr := repo.FindByEmail(context.Background(), testEmail)

		assert.NoError(t, findErr)
		require.NotNil(t, result)
		assert.Equal(t, testID, result.ID)
		assert.Equal(t, "Test Entity", result.Name)
		assert.Equal(t, "test@example.com", result.Email.String())
		assert.True(t, result.Active)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM entities WHERE email").
			WithArgs(testEmail.String()).
			WillReturnError(sql.ErrNoRows)

		result, findErr := repo.FindByEmail(context.Background(), testEmail)

		assert.Nil(t, result)
		assert.ErrorIs(t, findErr, entity.ErrEntityNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM entities WHERE email").
			WithArgs(testEmail.String()).
			WillReturnError(sql.ErrConnDone)

		result, findErr := repo.FindByEmail(context.Background(), testEmail)

		assert.Nil(t, result)
		assert.Error(t, findErr)
		assert.ErrorIs(t, findErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// --- List --------------------------------------------------------------------

func TestEntityRepository_List(t *testing.T) {
	db, mock, mockErr := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, mockErr)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	cluster := database.NewDBClusterFromDB(sqlxDB)
	repo := NewEntityRepository(cluster)

	now := time.Now().Truncate(time.Microsecond)
	testID := vo.NewID()

	t.Run("success with results", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities").
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"}).
			AddRow(testID.String(), "Test Entity", "test@example.com", true, now, now)
		mock.ExpectQuery("SELECT .+ FROM entities").
			WillReturnRows(dataRows)

		filter := entity.ListFilter{Page: 1, Limit: 20}
		result, listErr := repo.List(context.Background(), filter)

		assert.NoError(t, listErr)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Limit)
		require.Len(t, result.Entities, 1)
		assert.Equal(t, testID, result.Entities[0].ID)
		assert.Equal(t, "Test Entity", result.Entities[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities").
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT .+ FROM entities").
			WillReturnRows(dataRows)

		filter := entity.ListFilter{Page: 1, Limit: 20}
		result, listErr := repo.List(context.Background(), filter)

		assert.NoError(t, listErr)
		require.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
		assert.Empty(t, result.Entities)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with name filter", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities WHERE name ILIKE").
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"}).
			AddRow(testID.String(), "Test Entity", "test@example.com", true, now, now)
		mock.ExpectQuery("SELECT .+ FROM entities.+WHERE name ILIKE").
			WillReturnRows(dataRows)

		filter := entity.ListFilter{Page: 1, Limit: 20, Name: "Test"}
		result, listErr := repo.List(context.Background(), filter)

		assert.NoError(t, listErr)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		require.Len(t, result.Entities, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with email filter", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities WHERE email ILIKE").
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"}).
			AddRow(testID.String(), "Test Entity", "test@example.com", true, now, now)
		mock.ExpectQuery("SELECT .+ FROM entities.+WHERE email ILIKE").
			WillReturnRows(dataRows)

		filter := entity.ListFilter{Page: 1, Limit: 20, Email: "test@"}
		result, listErr := repo.List(context.Background(), filter)

		assert.NoError(t, listErr)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		require.Len(t, result.Entities, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with active only filter", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities WHERE active = true").
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{"id", "name", "email", "active", "created_at", "updated_at"}).
			AddRow(testID.String(), "Active Entity", "active@example.com", true, now, now)
		mock.ExpectQuery("SELECT .+ FROM entities.+WHERE active = true").
			WillReturnRows(dataRows)

		filter := entity.ListFilter{Page: 1, Limit: 20, ActiveOnly: true}
		result, listErr := repo.List(context.Background(), filter)

		assert.NoError(t, listErr)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		require.Len(t, result.Entities, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on count query", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities").
			WillReturnError(sql.ErrConnDone)

		filter := entity.ListFilter{Page: 1, Limit: 20}
		result, listErr := repo.List(context.Background(), filter)

		assert.Nil(t, result)
		assert.Error(t, listErr)
		assert.ErrorIs(t, listErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on data query", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM entities").
			WillReturnRows(countRows)

		mock.ExpectQuery("SELECT .+ FROM entities").
			WillReturnError(sql.ErrConnDone)

		filter := entity.ListFilter{Page: 1, Limit: 20}
		result, listErr := repo.List(context.Background(), filter)

		assert.Nil(t, result)
		assert.Error(t, listErr)
		assert.ErrorIs(t, listErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// --- Update ------------------------------------------------------------------

func TestEntityRepository_Update(t *testing.T) {
	db, mock, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	cluster := database.NewDBClusterFromDB(sqlxDB)
	repo := NewEntityRepository(cluster)

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		e := buildTestEntity()
		updateErr := repo.Update(context.Background(), e)

		assert.NoError(t, updateErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - zero rows affected", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()

		e := buildTestEntity()
		updateErr := repo.Update(context.Background(), e)

		assert.ErrorIs(t, updateErr, entity.ErrEntityNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on exec", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		e := buildTestEntity()
		updateErr := repo.Update(context.Background(), e)

		assert.Error(t, updateErr)
		assert.ErrorIs(t, updateErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction begin error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

		e := buildTestEntity()
		updateErr := repo.Update(context.Background(), e)

		assert.Error(t, updateErr)
		assert.Contains(t, updateErr.Error(), "beginning transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction commit error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

		e := buildTestEntity()
		updateErr := repo.Update(context.Background(), e)

		assert.Error(t, updateErr)
		assert.Contains(t, updateErr.Error(), "committing transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// --- Delete ------------------------------------------------------------------

func TestEntityRepository_Delete(t *testing.T) {
	db, mock, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	cluster := database.NewDBClusterFromDB(sqlxDB)
	repo := NewEntityRepository(cluster)

	testID := vo.NewID()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), testID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		deleteErr := repo.Delete(context.Background(), testID)

		assert.NoError(t, deleteErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - zero rows affected", func(t *testing.T) {
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), testID.String()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		deleteErr := repo.Delete(context.Background(), testID)

		assert.ErrorIs(t, deleteErr, entity.ErrEntityNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("UPDATE entities SET").
			WithArgs(sqlmock.AnyArg(), testID.String()).
			WillReturnError(sql.ErrConnDone)

		deleteErr := repo.Delete(context.Background(), testID)

		assert.Error(t, deleteErr)
		assert.ErrorIs(t, deleteErr, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
