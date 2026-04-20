package role

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newRecordingContext returns a context carrying a recording span backed by an
// in-memory exporter. Callers must invoke endAndFlush before inspecting the
// exporter's finished spans, so attributes/status are guaranteed to be visible.
func newRecordingContext(t *testing.T) (context.Context, *tracetest.InMemoryExporter, func()) {
	t.Helper()

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	ctx, span := tp.Tracer("role-test").Start(context.Background(), "test-span")

	endAndFlush := func() {
		span.End()
		_ = tp.ForceFlush(context.Background())
	}

	return ctx, exp, endAndFlush
}
