package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newRecordingProvider returns a TracerProvider wired to an in-memory exporter
// so tests can assert on finished span data (names, attributes, etc.).
func newRecordingProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
	})
	return tp, exporter
}

func TestSpanRename_GET_UsersByID(t *testing.T) {
	// TC-UC-57: GET /v1/users/:id -> root span name = http.get.users_by_id
	tp, exporter := newRecordingProvider(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(otelgin.Middleware("test", otelgin.WithTracerProvider(tp)))
	r.Use(SpanRename())
	r.GET("/v1/users/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/abc-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "expected exactly one root span from otelgin")
	assert.Equal(t, "http.get.users_by_id", spans[0].Name)
}

func TestSpanRename_POST_Users(t *testing.T) {
	// TC-UC-58: POST /v1/users -> root span name = http.post.users
	tp, exporter := newRecordingProvider(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(otelgin.Middleware("test", otelgin.WithTracerProvider(tp)))
	r.Use(SpanRename())
	r.POST("/v1/users", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "http.post.users", spans[0].Name)
}

func TestSpanRename_UnmatchedRoute_Unknown(t *testing.T) {
	// Edge: 404 on unmatched route -> c.FullPath() is empty -> name = http.get.unknown
	tp, exporter := newRecordingProvider(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(otelgin.Middleware("test", otelgin.WithTracerProvider(tp)))
	r.Use(SpanRename())
	// no routes registered

	req := httptest.NewRequest(http.MethodGet, "/does/not/exist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "http.get.unknown", spans[0].Name)
}
