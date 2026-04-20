package user

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newRecordingSpanContext returns a context carrying a recording span backed
// by an in-memory exporter, plus a finalize function that ends the span,
// flushes the provider and returns the captured span snapshot.
//
// Callers use this to assert OTel attributes / status / events emitted by
// the use case under test (e.g. `app.result=duplicate_email`,
// `app.validation_error=<msg>`, FailSpan with `error.type`).
func newRecordingSpanContext(t *testing.T) (context.Context, func() tracetest.SpanStub) {
	t.Helper()

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	ctx, span := tp.Tracer("test").Start(context.Background(), "use-case-test")

	finalize := func() tracetest.SpanStub {
		t.Helper()
		span.End()
		if flushErr := tp.ForceFlush(context.Background()); flushErr != nil {
			t.Fatalf("flush span exporter: %v", flushErr)
		}
		spans := exp.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("expected exactly one finished span, got %d", len(spans))
		}
		return spans[0]
	}

	return ctx, finalize
}

// hasAttr returns true when the span snapshot contains an attribute whose
// key matches the supplied key and whose string value equals want. When want
// is empty, only the key presence is required (callers can read the value
// directly off the returned attribute).
func hasAttr(stub tracetest.SpanStub, key, want string) bool {
	for _, attr := range stub.Attributes {
		if string(attr.Key) != key {
			continue
		}
		if want == "" {
			return true
		}
		if attr.Value.AsString() == want {
			return true
		}
	}
	return false
}

// attrValue returns the string value of the named attribute, or "" when the
// key is absent.
func attrValue(stub tracetest.SpanStub, key string) string {
	for _, attr := range stub.Attributes {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

// hasExceptionEventAttr returns true when the span recorded an `exception`
// event carrying an attribute whose key equals attrKey. We probe events
// rather than top-level attributes because trace.RecordError attaches the
// `error.type` attribute to the exception event itself, not to the span.
func hasExceptionEventAttr(stub tracetest.SpanStub, attrKey string) bool {
	for _, ev := range stub.Events {
		if ev.Name != "exception" {
			continue
		}
		for _, attr := range ev.Attributes {
			if string(attr.Key) == attrKey {
				return true
			}
		}
	}
	return false
}

// eventNames returns the names of all events on the span, in emission order.
func eventNames(stub tracetest.SpanStub) []string {
	names := make([]string, 0, len(stub.Events))
	for _, ev := range stub.Events {
		names = append(names, ev.Name)
	}
	return names
}

// hasEvent reports whether the span emitted an event with the given name.
func hasEvent(stub tracetest.SpanStub, name string) bool {
	for _, ev := range stub.Events {
		if ev.Name == name {
			return true
		}
	}
	return false
}

// eventAttr returns the string value of the named attribute on the first
// event matching eventName, or "" when the event or attribute is absent.
func eventAttr(stub tracetest.SpanStub, eventName, attrKey string) string {
	for _, ev := range stub.Events {
		if ev.Name != eventName {
			continue
		}
		for _, attr := range ev.Attributes {
			if string(attr.Key) == attrKey {
				return attr.Value.AsString()
			}
		}
	}
	return ""
}
