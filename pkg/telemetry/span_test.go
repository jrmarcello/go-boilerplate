package telemetry

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// MyErr is a custom error type used to verify FailSpan records the
// fully-qualified Go type name as the `error.type` attribute.
type MyErr struct{ msg string }

func (e *MyErr) Error() string { return e.msg }

// newTestSpan creates a real SDK span backed by an in-memory recorder,
// returning both the span and the recorder so assertions can inspect
// the finished span data.
func newTestSpan(t *testing.T) (trace.Span, *tracetest.InMemoryExporter) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
	})

	_, span := tp.Tracer("test").Start(context.Background(), "test-op")
	return span, exporter
}

func TestFailSpan(t *testing.T) {
	t.Run("sets Error status and records error event", func(t *testing.T) {
		span, exporter := newTestSpan(t)
		testErr := errors.New("something went wrong")

		FailSpan(span, testErr, "operation failed")
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)

		recorded := spans[0]
		assert.Equal(t, codes.Error, recorded.Status.Code)
		assert.Equal(t, "operation failed", recorded.Status.Description)

		// Verify the error was recorded as an event
		require.NotEmpty(t, recorded.Events)

		foundErrEvent := false
		for _, evt := range recorded.Events {
			if evt.Name == "exception" {
				foundErrEvent = true
				break
			}
		}
		assert.True(t, foundErrEvent, "expected an exception event to be recorded")
	})

	t.Run("nil span is no-op", func(t *testing.T) {
		// Must not panic
		assert.NotPanics(t, func() {
			FailSpan(nil, errors.New("err"), "msg")
		})
	})

	// TC-UC-01: typed error → exception event has error.type = *telemetry.MyErr
	// and a stack trace attribute.
	t.Run("records error.type and stack trace for typed error", func(t *testing.T) {
		span, exporter := newTestSpan(t)
		typedErr := &MyErr{msg: "typed boom"}

		FailSpan(span, typedErr, "operation failed")
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)

		recorded := spans[0]
		assert.Equal(t, codes.Error, recorded.Status.Code)
		assert.Equal(t, "operation failed", recorded.Status.Description)

		require.NotEmpty(t, recorded.Events)

		var (
			foundException bool
			gotExcType     string
			gotErrTypeAttr string
			gotStack       string
		)
		for _, evt := range recorded.Events {
			if evt.Name != "exception" {
				continue
			}
			foundException = true
			for _, attr := range evt.Attributes {
				switch attr.Key {
				case attribute.Key("exception.type"):
					gotExcType = attr.Value.AsString()
				case attribute.Key("error.type"):
					gotErrTypeAttr = attr.Value.AsString()
				case attribute.Key("exception.stacktrace"):
					gotStack = attr.Value.AsString()
				}
			}
		}
		require.True(t, foundException, "expected an exception event to be recorded")

		expectedType := fmt.Sprintf("%T", typedErr)
		assert.Equal(t, expectedType, gotErrTypeAttr,
			"error.type attribute must equal fmt.Sprintf(%%T, err)")
		assert.Equal(t, expectedType, gotExcType,
			"exception.type (auto-set by SDK) must also reflect the typed error")
		assert.NotEmpty(t, gotStack,
			"exception.stacktrace must be present (FailSpan should pass WithStackTrace(true))")
	})

	// TC-UC-02: wrapped error → exception event reports the wrapper type
	// (*fmt.wrapError on Go 1.20+) and includes a stack trace.
	t.Run("records wrapper type for fmt.Errorf-wrapped error", func(t *testing.T) {
		span, exporter := newTestSpan(t)
		baseErr := errors.New("boom")
		wrappedErr := fmt.Errorf("ctx: %w", baseErr)

		FailSpan(span, wrappedErr, "wrapped failure")
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		recorded := spans[0]

		require.NotEmpty(t, recorded.Events)

		var (
			foundException bool
			gotErrTypeAttr string
			gotStack       string
		)
		for _, evt := range recorded.Events {
			if evt.Name != "exception" {
				continue
			}
			foundException = true
			for _, attr := range evt.Attributes {
				switch attr.Key {
				case attribute.Key("error.type"):
					gotErrTypeAttr = attr.Value.AsString()
				case attribute.Key("exception.stacktrace"):
					gotStack = attr.Value.AsString()
				}
			}
		}
		require.True(t, foundException, "expected an exception event to be recorded")

		expectedType := fmt.Sprintf("%T", wrappedErr)
		assert.Equal(t, expectedType, gotErrTypeAttr,
			"error.type must reflect the runtime wrapper type (e.g. *fmt.wrapError)")
		assert.NotEmpty(t, gotStack,
			"exception.stacktrace must be present for wrapped errors too")
	})

	// TC-UC-03: nil span with non-nil err must not panic and must not mutate
	// any state (covered by NotPanics — there is no state to mutate).
	t.Run("nil span with non-nil err is no-op", func(t *testing.T) {
		assert.NotPanics(t, func() {
			FailSpan(nil, errors.New("any"), "any msg")
		})
	})
}

func TestWarnSpan(t *testing.T) {
	t.Run("adds attribute without setting Error status", func(t *testing.T) {
		span, exporter := newTestSpan(t)

		WarnSpan(span, "warn.reason", "cache miss")
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)

		recorded := spans[0]

		// Status must NOT be Error
		assert.NotEqual(t, codes.Error, recorded.Status.Code)

		// Verify the attribute was added
		foundAttr := false
		for _, attr := range recorded.Attributes {
			if attr.Key == attribute.Key("warn.reason") && attr.Value.AsString() == "cache miss" {
				foundAttr = true
				break
			}
		}
		assert.True(t, foundAttr, "expected attribute warn.reason=cache miss to be present")
	})

	t.Run("nil span is no-op", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WarnSpan(nil, "key", "value")
		})
	})
}
