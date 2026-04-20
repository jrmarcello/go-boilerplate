package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrmarcello/gopherplate/pkg/httputil"
	"github.com/jrmarcello/gopherplate/pkg/idempotency"
	"github.com/jrmarcello/gopherplate/pkg/logutil"
	"github.com/jrmarcello/gopherplate/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	// IdempotencyKeyHeader is the header used for idempotency.
	IdempotencyKeyHeader = "Idempotency-Key"

	// idempotencyKeyPrefix is the prefix used in Redis keys.
	idempotencyKeyPrefix = "idempotency:"
)

// Idempotency returns a middleware that ensures idempotency for POST requests.
// The Idempotency-Key header is optional: if absent, the request is processed normally.
// If Redis is unavailable, the middleware operates in fail-open mode (degrades gracefully).
func Idempotency(store idempotency.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Only applies to POST requests
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		// 2. Header is optional — if absent, process normally
		key := c.GetHeader(IdempotencyKeyHeader)
		if key == "" {
			c.Next()
			return
		}

		// 3. Read and buffer body for fingerprint.
		// Memory safety relies on an upstream body cap (see middleware.BodyLimit).
		// When the cap trips, io.ReadAll surfaces *http.MaxBytesError and we
		// return 413 here instead of letting the handler respond 400. Other
		// read errors fall through to fail-open so a transient network hiccup
		// does not break requests that would otherwise succeed.
		reqBody, readErr := io.ReadAll(c.Request.Body)
		if readErr != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(readErr, &maxBytesErr) {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, httputil.ErrorResponse{
					Errors: httputil.ErrorDetail{Message: "request body too large"},
				})
				return
			}
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		fingerprint := bodyFingerprint(reqBody)

		// 4. Build Redis key with optional service namespace
		serviceName := c.GetHeader("X-Service-Name")
		fullKey := buildIdempotencyKey(serviceName, key)

		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)
		keyAttr := attribute.String("idempotency.key", key)

		// 5. Attempt to acquire lock
		acquired, lockErr := store.Lock(ctx, fullKey, fingerprint)
		if lockErr != nil {
			// Redis unavailable -> fail-open; emit event for trace correlation
			// AND keep the log for operator visibility per REQ-4 (observability.md
			// permits logutil on infra-unreachable branches).
			telemetry.RecordEvent(span, "idempotency.store_unavailable", keyAttr)
			// fail-open infra-unreachable branch; span event already emitted for correlation
			// nosemgrep: gopherplate-usecase-no-slog-in-flow
			logutil.LogWarn(ctx, "idempotency store unavailable, proceeding without",
				"error", lockErr.Error(), "idempotency_key", key)
			c.Next()
			return
		}

		if !acquired {
			// Key already exists: check state
			entry, getErr := store.Get(ctx, fullKey)
			if getErr != nil {
				// fail-open infra-unreachable branch; log retained for operator visibility
				// nosemgrep: gopherplate-usecase-no-slog-in-flow
				logutil.LogWarn(ctx, "idempotency store get failed, proceeding without",
					"error", getErr.Error(), "idempotency_key", key)
				c.Next()
				return
			}

			if entry == nil {
				// Key existed but expired between Lock and Get (rare race condition)
				c.Next()
				return
			}

			if entry.Status == idempotency.StatusProcessing {
				// Previous request still in progress -> 409 Conflict
				telemetry.RecordEvent(span, "idempotency.locked", keyAttr)
				c.AbortWithStatusJSON(http.StatusConflict, httputil.ErrorResponse{
					Errors: httputil.ErrorDetail{Message: "A request with this Idempotency-Key is already being processed"},
				})
				return
			}

			// COMPLETED -> verify fingerprint before replay
			if entry.Fingerprint != "" && fingerprint != entry.Fingerprint {
				telemetry.RecordEvent(span, "idempotency.fingerprint_mismatch", keyAttr)
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, httputil.ErrorResponse{
					Errors: httputil.ErrorDetail{Message: "Idempotency-Key already used with a different request body"},
				})
				return
			}

			// Replay stored response
			telemetry.RecordEvent(span, "idempotency.replayed", keyAttr,
				attribute.Int("idempotency.status_code", entry.StatusCode))
			c.Data(entry.StatusCode, "application/json; charset=utf-8", entry.Body)
			c.Abort()
			return
		}

		// 6. First request with this key — capture response
		telemetry.RecordEvent(span, "idempotency.key_acquired", keyAttr)
		rw := &idempotencyResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = rw

		// 7. Execute handler
		c.Next()

		// 8. Store or release based on status code
		statusCode := rw.Status()
		if shouldStoreResponse(statusCode) {
			completeErr := store.Complete(ctx, fullKey, &idempotency.Entry{
				StatusCode:  statusCode,
				Body:        rw.body.Bytes(),
				Fingerprint: fingerprint,
			})
			if completeErr != nil {
				// fail-open infra-unreachable branch; log retained for operator visibility
				// nosemgrep: gopherplate-usecase-no-slog-in-flow
				logutil.LogWarn(ctx, "failed to store idempotency response",
					"error", completeErr.Error(), "idempotency_key", key)
			} else {
				telemetry.RecordEvent(span, "idempotency.stored", keyAttr,
					attribute.Int("idempotency.status_code", statusCode))
			}
		} else {
			// 5xx error -> unlock to allow retry
			unlockErr := store.Unlock(ctx, fullKey)
			if unlockErr != nil {
				// fail-open infra-unreachable branch; log retained for operator visibility
				// nosemgrep: gopherplate-usecase-no-slog-in-flow
				logutil.LogWarn(ctx, "failed to unlock idempotency key",
					"error", unlockErr.Error(), "idempotency_key", key)
			} else {
				telemetry.RecordEvent(span, "idempotency.released", keyAttr)
			}
		}
	}
}

// buildIdempotencyKey builds the Redis key with namespace.
// Format: idempotency:{service-name}:{key} or idempotency:{key}
func buildIdempotencyKey(serviceName, key string) string {
	if serviceName != "" {
		return idempotencyKeyPrefix + serviceName + ":" + key
	}
	return idempotencyKeyPrefix + key
}

// shouldStoreResponse determines whether the response should be stored for replay.
// 2xx and 4xx are deterministic and should be stored.
// 5xx are transient and should allow retry.
func shouldStoreResponse(statusCode int) bool {
	return statusCode >= 200 && statusCode < 500
}

// idempotencyResponseWriter wraps gin.ResponseWriter to capture the response body.
type idempotencyResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *idempotencyResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// bodyFingerprint calculates the SHA-256 of the request body to detect reuse
// of Idempotency-Key with a different body.
func bodyFingerprint(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}
