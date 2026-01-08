package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Otel   OtelConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	// Formato Postgres: postgres://user:password@host:port/database?sslmode=disable
	DSN string
}

type OtelConfig struct {
	ServiceName  string
	CollectorURL string
}

type RedisConfig struct {
	URL     string
	TTL     string // ex: "5m", "1h"
	Enabled bool
}

func Load() (*Config, error) {
	// godotenv.Load() carrega variáveis do arquivo .env para o ambiente.
	// O underscore ignora o erro se o arquivo não existir (ok em produção).
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			DSN: buildDSN(),
		},
		Otel: OtelConfig{
			ServiceName:  getEnv("OTEL_SERVICE_NAME", "entity-service"),
			CollectorURL: getEnv("OTEL_COLLECTOR_URL", ""),
		},
		Redis: RedisConfig{
			URL:     getEnv("REDIS_URL", "redis://localhost:6379"),
			TTL:     getEnv("REDIS_TTL", "5m"),
			Enabled: getEnv("REDIS_ENABLED", "false") == "true",
		},
	}, nil
}

// buildDSN constrói a connection string do Postgres.
// Se DB_DSN estiver definida, usa ela diretamente.
// Caso contrário, constrói a partir das variáveis POSTGRES_*.
func buildDSN() string {
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("POSTGRES_USER", "user"),
		getEnv("POSTGRES_PASSWORD", "password"),
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_DB", "entities"),
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
