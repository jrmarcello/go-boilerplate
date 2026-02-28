package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Setup env vars for test - os.Getenv reads these!
	// Struct structure: Server.Port -> SERVER_PORT
	os.Setenv("SERVER_PORT", "9090")
	// DB.DSN -> DB_DSN
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test_db?sslmode=disable")
	os.Setenv("DB_READER_DSN", "postgres://test:test@reader:5432/test_db?sslmode=disable")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_CONN_MAX_LIFETIME", "10m")
	// Redis.Enabled -> REDIS_ENABLED
	os.Setenv("REDIS_ENABLED", "true")
	os.Setenv("SWAGGER_ENABLED", "false")
	defer os.Clearenv()

	cfg, loadErr := Load()
	assert.NoError(t, loadErr)
	assert.NotNil(t, cfg)

	// Verify overrides
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Contains(t, cfg.DB.DSN, "test_db")
	assert.Equal(t, "postgres://test:test@reader:5432/test_db?sslmode=disable", cfg.DB.ReaderDSN)
	assert.Equal(t, 50, cfg.DB.MaxOpenConns)
	assert.Equal(t, 10, cfg.DB.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, cfg.DB.ConnMaxLifetime)
	assert.True(t, cfg.Redis.Enabled)
	assert.False(t, cfg.Swagger.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.GetRedisTTL())

	// Verify ReaderConfig returns proper config
	readerCfg := cfg.DB.ReaderConfig()
	assert.NotNil(t, readerCfg)
	assert.Equal(t, cfg.DB.ReaderDSN, readerCfg.DSN)
	assert.Equal(t, cfg.DB.MaxOpenConns, readerCfg.MaxOpenConns)
}

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, loadErr := Load()
	assert.NoError(t, loadErr)

	// Verify defaults
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Contains(t, cfg.DB.DSN, "entities")
	assert.Equal(t, "", cfg.DB.ReaderDSN)
	assert.Equal(t, 25, cfg.DB.MaxOpenConns)
	assert.Equal(t, 25, cfg.DB.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.DB.ConnMaxLifetime)
	assert.False(t, cfg.Redis.Enabled)
	assert.True(t, cfg.Swagger.Enabled)

	// ReaderConfig should be nil when no reader DSN
	assert.Nil(t, cfg.DB.ReaderConfig())
}
