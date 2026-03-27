package database

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("returns config with provided DSN", func(t *testing.T) {
		dsn := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
		cfg := DefaultConfig(dsn)

		assert.Equal(t, dsn, cfg.DSN)
	})

	t.Run("returns sensible default values", func(t *testing.T) {
		cfg := DefaultConfig("any-dsn")

		assert.Equal(t, 25, cfg.MaxOpenConns)
		assert.Equal(t, 10, cfg.MaxIdleConns)
		assert.Equal(t, 5*time.Minute, cfg.ConnMaxLifetime)
		assert.Equal(t, 90*time.Second, cfg.ConnMaxIdleTime)
	})

	t.Run("accepts empty DSN", func(t *testing.T) {
		cfg := DefaultConfig("")

		assert.Equal(t, "", cfg.DSN)
		assert.Equal(t, 25, cfg.MaxOpenConns)
	})
}

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, mockErr := sqlmock.New()
	require.NoError(t, mockErr)
	return sqlx.NewDb(mockDB, "sqlmock"), mock
}

func TestNewDBClusterFromDB(t *testing.T) {
	t.Run("creates cluster with writer equals reader (fallback)", func(t *testing.T) {
		db, _ := newMockDB(t)
		defer db.Close()

		cluster := NewDBClusterFromDB(db)

		require.NotNil(t, cluster)
		assert.Equal(t, db, cluster.Writer())
		assert.Equal(t, db, cluster.Reader(), "reader should fall back to writer")
	})

	t.Run("HasSeparateReader returns false", func(t *testing.T) {
		db, _ := newMockDB(t)
		defer db.Close()

		cluster := NewDBClusterFromDB(db)

		assert.False(t, cluster.HasSeparateReader())
	})
}

func TestDBCluster_Writer(t *testing.T) {
	t.Run("returns the writer connection", func(t *testing.T) {
		db, _ := newMockDB(t)
		defer db.Close()

		cluster := NewDBClusterFromDB(db)

		assert.Same(t, db, cluster.Writer())
	})
}

func TestDBCluster_Reader(t *testing.T) {
	t.Run("returns writer when no separate reader", func(t *testing.T) {
		writerDB, _ := newMockDB(t)
		defer writerDB.Close()

		cluster := NewDBClusterFromDB(writerDB)

		assert.Same(t, writerDB, cluster.Reader())
	})

	t.Run("returns separate reader when configured", func(t *testing.T) {
		writerDB, _ := newMockDB(t)
		defer writerDB.Close()

		readerDB, _ := newMockDB(t)
		defer readerDB.Close()

		cluster := &DBCluster{writer: writerDB, reader: readerDB}

		assert.Same(t, readerDB, cluster.Reader())
		assert.NotSame(t, writerDB, cluster.Reader())
	})
}

func TestDBCluster_HasSeparateReader(t *testing.T) {
	t.Run("returns false when reader is nil", func(t *testing.T) {
		db, _ := newMockDB(t)
		defer db.Close()

		cluster := NewDBClusterFromDB(db)

		assert.False(t, cluster.HasSeparateReader())
	})

	t.Run("returns true when separate reader is configured", func(t *testing.T) {
		writerDB, _ := newMockDB(t)
		defer writerDB.Close()

		readerDB, _ := newMockDB(t)
		defer readerDB.Close()

		cluster := &DBCluster{writer: writerDB, reader: readerDB}

		assert.True(t, cluster.HasSeparateReader())
	})
}

func TestDBCluster_Close(t *testing.T) {
	t.Run("closes writer-only cluster without error", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectClose()

		cluster := NewDBClusterFromDB(db)

		closeErr := cluster.Close()
		assert.NoError(t, closeErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("closes both writer and reader", func(t *testing.T) {
		writerDB, writerMock := newMockDB(t)
		writerMock.ExpectClose()

		readerDB, readerMock := newMockDB(t)
		readerMock.ExpectClose()

		cluster := &DBCluster{writer: writerDB, reader: readerDB}

		closeErr := cluster.Close()
		assert.NoError(t, closeErr)
		assert.NoError(t, writerMock.ExpectationsWereMet())
		assert.NoError(t, readerMock.ExpectationsWereMet())
	})
}

func TestNewConnection(t *testing.T) {
	t.Run("empty DSN returns error", func(t *testing.T) {
		cfg := Config{DSN: ""}

		db, connErr := NewConnection(cfg)

		assert.Nil(t, db)
		assert.Error(t, connErr)
		assert.Contains(t, connErr.Error(), "failed to connect to database")
	})

	t.Run("invalid DSN returns error", func(t *testing.T) {
		cfg := Config{DSN: "not-a-valid-dsn://!!!"}

		db, connErr := NewConnection(cfg)

		assert.Nil(t, db)
		assert.Error(t, connErr)
	})
}
