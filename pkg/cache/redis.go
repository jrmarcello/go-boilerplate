package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss indicates that the key was not found in cache.
var ErrCacheMiss = errors.New("cache miss")

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	URL     string
	TTL     string
	Enabled bool
}

// RedisClient implements the Cache interface using Redis.
type RedisClient struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisClient creates a new Redis cache client.
// Returns nil if cache is disabled (no-op pattern).
func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	if !cfg.Enabled {
		slog.Info("Redis cache disabled")
		return nil, nil
	}

	opts, parseErr := redis.ParseURL(cfg.URL)
	if parseErr != nil {
		return nil, parseErr
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pingErr := client.Ping(ctx).Err(); pingErr != nil {
		return nil, pingErr
	}

	ttl, ttlErr := time.ParseDuration(cfg.TTL)
	if ttlErr != nil {
		ttl = 5 * time.Minute
	}

	slog.Info("Redis cache initialized",
		"addr", opts.Addr,
		"ttl", ttl,
	)

	return &RedisClient{
		client: client,
		ttl:    ttl,
	}, nil
}

// Get retrieves a value from the cache.
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	if r == nil {
		return ErrCacheMiss
	}
	val, getErr := r.client.Get(ctx, key).Result()
	if errors.Is(getErr, redis.Nil) {
		return ErrCacheMiss
	}
	if getErr != nil {
		return getErr
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in the cache with the configured TTL.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}) error {
	if r == nil {
		return nil
	}
	data, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		return marshalErr
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// Delete removes a value from the cache.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if r == nil {
		return nil
	}
	return r.client.Del(ctx, key).Err()
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	if r == nil {
		return nil
	}
	return r.client.Close()
}

// RedisClient returns the underlying go-redis client for use by other packages
// (e.g., pkg/idempotency). Returns nil if the cache is disabled.
func (r *RedisClient) UnderlyingClient() *redis.Client {
	if r == nil {
		return nil
	}
	return r.client
}

// Ping checks the Redis connection.
func (r *RedisClient) Ping(ctx context.Context) error {
	if r == nil {
		return nil
	}
	return r.client.Ping(ctx).Err()
}
