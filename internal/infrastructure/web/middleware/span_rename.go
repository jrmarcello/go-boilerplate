package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/jrmarcello/gopherplate/pkg/telemetry"
)

// SpanRename renames the root span created by otelgin so it follows the
// project's naming convention `http.<verb>.<resource>` (snake_case).
//
// Ordering: this middleware MUST be registered immediately after
// `otelgin.Middleware(...)`. It also defers the rename to AFTER `c.Next()`
// because Gin only populates `c.FullPath()` once routing has matched the
// incoming request — calling SetName before Next would see an empty template
// and fall back to `http.<verb>.unknown` for every successful request.
//
// For unmatched routes (404) `c.FullPath()` stays empty and the resource part
// becomes `unknown`, which is the desired behavior.
func SpanRename() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // let routing populate c.FullPath()
		span := trace.SpanFromContext(c.Request.Context())
		if span == nil {
			return
		}
		span.SetName(telemetry.HTTPSpanName(c.Request.Method, c.FullPath()))
	}
}
