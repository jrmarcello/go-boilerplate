package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/appmax-space/go-boilerplate/pkg/httputil"
	"bitbucket.org/appmax-space/go-boilerplate/pkg/logutil"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig contém configurações do rate limiter
type RateLimiterConfig struct {
	// RequestsPerSecond é o número de requisições permitidas por segundo
	RequestsPerSecond float64
	// BurstSize é o tamanho máximo do burst
	BurstSize int
}

// limiterEntry holds a rate limiter and the last time it was accessed.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter gerencia rate limiters por IP
type IPRateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	config   RateLimiterConfig
}

// NewIPRateLimiter cria um novo rate limiter por IP
func NewIPRateLimiter(config RateLimiterConfig) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*limiterEntry),
		config:   config,
	}
}

// getLimiter retorna o limiter para um IP específico
func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	now := time.Now()

	i.mu.RLock()
	entry, exists := i.limiters[ip]
	i.mu.RUnlock()

	if exists {
		// Update lastSeen under write lock
		i.mu.Lock()
		entry.lastSeen = now
		i.mu.Unlock()
		return entry.limiter
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Double check após adquirir lock exclusivo
	if entry, exists = i.limiters[ip]; exists {
		entry.lastSeen = now
		return entry.limiter
	}

	entry = &limiterEntry{
		limiter:  rate.NewLimiter(rate.Limit(i.config.RequestsPerSecond), i.config.BurstSize),
		lastSeen: now,
	}
	i.limiters[ip] = entry

	return entry.limiter
}

// RateLimit retorna um middleware de rate limiting por IP.
// The ctx parameter controls the lifetime of the background cleanup goroutine;
// pass the server's shutdown context so the goroutine stops on graceful shutdown.
func RateLimit(ctx context.Context, config RateLimiterConfig) gin.HandlerFunc {
	limiter := NewIPRateLimiter(config)

	// Goroutine para limpar limiters inativos periodicamente.
	// Evicts entries not seen for 10+ minutes instead of replacing the entire map.
	// Exits when ctx is canceled (graceful shutdown).
	go func() {
		const idleThreshold = 10 * time.Minute
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("rate limiter cleanup goroutine stopped")
				return
			case <-ticker.C:
				now := time.Now()
				limiter.mu.Lock()
				for ip, entry := range limiter.limiters {
					if now.Sub(entry.lastSeen) > idleThreshold {
						delete(limiter.limiters, ip)
					}
				}
				limiter.mu.Unlock()
			}
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.getLimiter(ip)

		if !l.Allow() {
			// logutil extracts LogContext from request context when available
			// (e.g., request_id, trace_id injected by upstream middleware),
			// providing richer structured logs for rate-limit events.
			logutil.LogWarn(c.Request.Context(), "rate limit exceeded",
				"ip", ip,
				"requests_per_second", config.RequestsPerSecond,
				"burst_size", config.BurstSize,
			)

			httputil.SendError(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultRateLimitConfig retorna configuração padrão de rate limiting
func DefaultRateLimitConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}
}
