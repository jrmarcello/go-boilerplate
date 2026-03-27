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

// IPRateLimiter gerencia rate limiters por IP
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   RateLimiterConfig
}

// NewIPRateLimiter cria um novo rate limiter por IP
func NewIPRateLimiter(config RateLimiterConfig) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}
}

// getLimiter retorna o limiter para um IP específico
func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.limiters[ip]
	i.mu.RUnlock()

	if exists {
		return limiter
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Double check após adquirir lock exclusivo
	if limiter, exists = i.limiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rate.Limit(i.config.RequestsPerSecond), i.config.BurstSize)
	i.limiters[ip] = limiter

	return limiter
}

// RateLimit retorna um middleware de rate limiting por IP.
// The ctx parameter controls the lifetime of the background cleanup goroutine;
// pass the server's shutdown context so the goroutine stops on graceful shutdown.
func RateLimit(ctx context.Context, config RateLimiterConfig) gin.HandlerFunc {
	limiter := NewIPRateLimiter(config)

	// Goroutine para limpar limiters antigos periodicamente.
	// Exits when ctx is canceled (graceful shutdown).
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("rate limiter cleanup goroutine stopped")
				return
			case <-ticker.C:
				limiter.mu.Lock()
				limiter.limiters = make(map[string]*rate.Limiter)
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
