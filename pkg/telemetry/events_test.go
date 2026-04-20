package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

// TC-UC-09: RecordEvent attaches a named event with attributes to the span.
func TestRecordEvent_AddsEventWithAttributes(t *testing.T) {
	span, exporter := newTestSpan(t)

	RecordEvent(span, "cache.hit", attribute.String("cache.key", "user:1"))
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	recorded := spans[0]
	require.Len(t, recorded.Events, 1)

	evt := recorded.Events[0]
	assert.Equal(t, "cache.hit", evt.Name)

	foundAttr := false
	for _, attr := range evt.Attributes {
		if attr.Key == attribute.Key("cache.key") && attr.Value.AsString() == "user:1" {
			foundAttr = true
			break
		}
	}
	assert.True(t, foundAttr, "expected attribute cache.key=user:1 on the event")
}

// TC-UC-10: RecordEvent on a nil span is a no-op (must not panic).
func TestRecordEvent_NilSpanIsNoOp(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordEvent(nil, "cache.hit", attribute.String("cache.key", "user:1"))
	})
}

// TC-UC-11: RecordEvent called twice records both events in order.
func TestRecordEvent_MultipleEventsCapturedInOrder(t *testing.T) {
	span, exporter := newTestSpan(t)

	RecordEvent(span, "cache.miss", attribute.String("cache.key", "user:1"))
	RecordEvent(span, "idempotency.replayed", attribute.String("idempotency.key", "abc-123"))
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	recorded := spans[0]
	require.Len(t, recorded.Events, 2)

	assert.Equal(t, "cache.miss", recorded.Events[0].Name)
	assert.Equal(t, "idempotency.replayed", recorded.Events[1].Name)
}
