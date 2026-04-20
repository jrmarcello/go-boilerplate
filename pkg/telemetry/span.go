package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// FailSpan marks the span as failed by setting its status to Error and
// recording the error as an `exception` event. The event is enriched with:
//
//   - `error.type` — the Go runtime type of err (via fmt.Sprintf("%T", err)),
//     useful for grouping by concrete error type even when the message varies.
//   - `exception.stacktrace` — captured at the call site via
//     trace.WithStackTrace(true), so tail-sampled traces retain enough
//     context to triage unexpected failures.
//
// It is a no-op if span is nil.
func FailSpan(span trace.Span, err error, msg string) {
	if span == nil {
		return
	}
	span.SetStatus(codes.Error, msg)
	span.RecordError(
		err,
		trace.WithAttributes(attribute.String("error.type", fmt.Sprintf("%T", err))),
		trace.WithStackTrace(true),
	)
}

// WarnSpan adds a semantic attribute to the span without changing its status
// to Error. Useful for annotating non-fatal conditions (e.g. cache misses,
// fallback paths) that are worth surfacing in traces.
// It is a no-op if span is nil.
func WarnSpan(span trace.Span, key, value string) {
	if span == nil {
		return
	}
	span.SetAttributes(attribute.String(key, value))
}
