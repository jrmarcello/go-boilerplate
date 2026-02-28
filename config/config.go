package config

import (
	"os"
	"strconv"
	"time"

	"bitbucket.org/appmax-space/go-boilerplate/pkg/database"
	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Otel    OtelConfig
	Redis   RedisConfig
	Auth    AuthConfig
	Swagger SwaggerConfig
}

type AuthConfig struct {
	// ServiceKeys no formato "service1:key1,service2:key2"
	// Se vazio, auth é desabilitada (dev mode)
	ServiceKeys string
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	// Formato Postgres: postgres://user:password@host:port/database?sslmode=disable
	DSN string
	// ReaderDSN é o DSN do banco de leitura (opcional, fallback para writer)
	ReaderDSN       string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// ReaderConfig retorna a configuração do reader como *database.Config.
// Retorna nil se ReaderDSN não estiver configurado.
func (d *DBConfig) ReaderConfig() *database.Config {
	if d.ReaderDSN == "" {
		return nil
	}
	return &database.Config{
		DSN:             d.ReaderDSN,
		MaxOpenConns:    d.MaxOpenConns,
		MaxIdleConns:    d.MaxIdleConns,
		ConnMaxLifetime: d.ConnMaxLifetime,
	}
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

type SwaggerConfig struct {
	Enabled bool
}

// Load configura a aplicação lendo do ambiente.
// Prioridade:
// 1. Variáveis de Ambiente (maior prioridade)
// 2. Arquivo .env (desenvolvimento local)
// 3. Defaults (fallback seguro)
func Load() (*Config, error) {
	// Carrega .env se existir (ignora erro se não existir)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			DSN:             getEnv("DB_DSN", "postgres://user:password@localhost:5432/entities?sslmode=disable"),
			ReaderDSN:       getEnv("DB_READER_DSN", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Otel: OtelConfig{
			ServiceName:  getEnv("OTEL_SERVICE_NAME", "entity-service"),
			CollectorURL: getEnv("OTEL_COLLECTOR_URL", ""),
		},
		Redis: RedisConfig{
			URL:     getEnv("REDIS_URL", "redis://localhost:6379"),
			TTL:     getEnv("REDIS_TTL", "5m"),
			Enabled: getEnvBool("REDIS_ENABLED", false),
		},
		Auth: AuthConfig{
			ServiceKeys: getEnv("SERVICE_KEYS", ""),
		},
		Swagger: SwaggerConfig{
			Enabled: getEnvBool("SWAGGER_ENABLED", true),
		},
	}, nil
}

// getEnv retorna o valor da variável de ambiente ou o fallback se não existir.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvBool retorna o valor booleano da variável de ambiente ou o fallback.
func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}

// getEnvInt retorna o valor inteiro da variável de ambiente ou o fallback.
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}

// getEnvDuration retorna o valor de duração da variável de ambiente ou o fallback.
// Aceita formatos como "5m", "1h", "30s".
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, parseErr := time.ParseDuration(value)
		if parseErr != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}

// GetRedisTTL retorna o TTL do Redis como time.Duration.
func (c *Config) GetRedisTTL() time.Duration {
	d, err := time.ParseDuration(c.Redis.TTL)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}
