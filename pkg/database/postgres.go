package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config holds database connection configuration.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(dsn string) Config {
	return Config{
		DSN:             dsn,
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// DBCluster provides Writer/Reader split with automatic fallback.
// If no reader is configured, reader operations fall back to the writer.
type DBCluster struct {
	writer *sqlx.DB
	reader *sqlx.DB
}

// NewConnection creates a single database connection.
func NewConnection(cfg Config) (*sqlx.DB, error) {
	db, connectErr := sqlx.Connect("postgres", cfg.DSN)
	if connectErr != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", connectErr)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pingErr := db.PingContext(ctx); pingErr != nil {
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return db, nil
}

// NewDBCluster creates a DBCluster with a writer and optional reader.
// If readerCfg is nil, reader operations will fall back to the writer.
func NewDBCluster(writerCfg Config, readerCfg *Config) (*DBCluster, error) {
	writer, writerErr := NewConnection(writerCfg)
	if writerErr != nil {
		return nil, fmt.Errorf("failed to connect writer: %w", writerErr)
	}

	cluster := &DBCluster{writer: writer}

	if readerCfg != nil && readerCfg.DSN != "" {
		reader, readerErr := NewConnection(*readerCfg)
		if readerErr != nil {
			// Log warning but don't fail — fall back to writer
			fmt.Printf("WARNING: Failed to connect reader, falling back to writer: %v\n", readerErr)
		} else {
			cluster.reader = reader
		}
	}

	return cluster, nil
}

// Writer returns the writer database connection.
func (c *DBCluster) Writer() *sqlx.DB {
	return c.writer
}

// Reader returns the reader database connection.
// Falls back to writer if no reader is configured.
func (c *DBCluster) Reader() *sqlx.DB {
	if c.reader != nil {
		return c.reader
	}
	return c.writer
}

// HasSeparateReader returns true if a separate reader is configured.
func (c *DBCluster) HasSeparateReader() bool {
	return c.reader != nil
}

// PingAll pings all database connections for health checks.
func (c *DBCluster) PingAll(ctx context.Context) error {
	if pingErr := c.writer.PingContext(ctx); pingErr != nil {
		return fmt.Errorf("writer ping failed: %w", pingErr)
	}
	if c.reader != nil {
		if pingErr := c.reader.PingContext(ctx); pingErr != nil {
			return fmt.Errorf("reader ping failed: %w", pingErr)
		}
	}
	return nil
}

// Close closes all database connections.
func (c *DBCluster) Close() error {
	var closeErr error
	if writerErr := c.writer.Close(); writerErr != nil {
		closeErr = fmt.Errorf("failed to close writer: %w", writerErr)
	}
	if c.reader != nil {
		if readerErr := c.reader.Close(); readerErr != nil {
			if closeErr != nil {
				closeErr = fmt.Errorf("%w; failed to close reader: %w", closeErr, readerErr)
			} else {
				closeErr = fmt.Errorf("failed to close reader: %w", readerErr)
			}
		}
	}
	return closeErr
}
