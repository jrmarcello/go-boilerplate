package telemetry

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RecordEvent attaches a named event to the span with optional attributes.
//
// Naming convention is `<subsystem>.<action>` in snake_case (for example,
// "cache.hit", "cache.miss", "idempotency.replayed"). Use this helper to
// signal noteworthy occurrences within a span without altering its status —
// for status changes, use FailSpan or WarnSpan instead.
//
// No-op when span is nil.
func RecordEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if span == nil {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}
