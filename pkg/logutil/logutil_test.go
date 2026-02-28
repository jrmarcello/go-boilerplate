package logutil

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectAndExtract(t *testing.T) {
	ctx := context.Background()
	lc := LogContext{
		RequestID: "req-123",
		TraceID:   "trace-456",
		Step:      StepHandler,
	}

	ctx = Inject(ctx, lc)
	extracted, ok := Extract(ctx)

	assert.True(t, ok)
	assert.Equal(t, "req-123", extracted.RequestID)
	assert.Equal(t, "trace-456", extracted.TraceID)
	assert.Equal(t, StepHandler, extracted.Step)
}

func TestExtractFromEmptyContext(t *testing.T) {
	ctx := context.Background()
	_, ok := Extract(ctx)
	assert.False(t, ok)
}

func TestLogContext_WithStep(t *testing.T) {
	lc := LogContext{RequestID: "req-123", Step: StepHandler}
	ucLC := lc.WithStep(StepUseCase)

	assert.Equal(t, StepUseCase, ucLC.Step)
	assert.Equal(t, StepHandler, lc.Step) // original not mutated
}

func TestLogContext_ToSlogAttrs(t *testing.T) {
	lc := LogContext{
		RequestID:     "req-123",
		TraceID:       "trace-456",
		Step:          StepUseCase,
		Resource:      "entity",
		Action:        "create",
		CallerService: "api-gateway",
	}

	attrs := lc.ToSlogAttrs()
	assert.GreaterOrEqual(t, len(attrs), 5)

	keys := make(map[string]bool)
	for _, attr := range attrs {
		keys[attr.Key] = true
	}
	assert.True(t, keys["request_id"])
	assert.True(t, keys["trace_id"])
	assert.True(t, keys["step"])
	assert.True(t, keys["resource"])
	assert.True(t, keys["action"])
}

func TestErrorLogFields_DomainError(t *testing.T) {
	err := errors.New("validation failed")
	attrs := ErrorLogFields(err, true)

	hasStack := false
	for _, attr := range attrs {
		if attr.Key == "stack_trace" {
			hasStack = true
		}
	}
	assert.False(t, hasStack, "domain errors should NOT have stack trace")
}

func TestErrorLogFields_InternalError(t *testing.T) {
	err := errors.New("db connection failed")
	attrs := ErrorLogFields(err, false)

	hasStack := false
	for _, attr := range attrs {
		if attr.Key == "stack_trace" {
			hasStack = true
		}
	}
	assert.True(t, hasStack, "internal errors SHOULD have stack trace")
}
