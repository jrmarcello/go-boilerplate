package shared

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Local sentinel errors used by ClassifyError tests. We intentionally do NOT
// import user/role domain packages here — the shared classifier must work
// against arbitrary sentinels with only the stdlib `errors` package.
var (
	errFoo      = errors.New("foo")
	errBar      = errors.New("bar")
	errInfraOOM = errors.New("connection refused")
)

// newTestSpan creates a real, recording span backed by an in-memory exporter.
// Callers must call span.End() and tp.ForceFlush() before inspecting the
// exporter's finished spans.
func newTestSpan(exp *tracetest.InMemoryExporter) (sdktrace.ReadWriteSpan, *sdktrace.TracerProvider) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
	)
	_, span := tp.Tracer("test").Start(context.Background(), "test-span")
	return span.(sdktrace.ReadWriteSpan), tp
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expected       []ExpectedError
		contextMsg     string
		wantStatus     codes.Code
		wantAttrKey    string
		wantAttrValue  string
		wantNoSpanCall bool
	}{
		{
			name: "TC-UC-04: matched ExpectedError with explicit AttrValue applies semantic attribute",
			err:  errFoo,
			expected: []ExpectedError{
				{Err: errFoo, AttrKey: AttrKeyAppResult, AttrValue: "not_found"},
			},
			contextMsg:    "getting foo",
			wantStatus:    codes.Unset,
			wantAttrKey:   AttrKeyAppResult,
			wantAttrValue: "not_found",
		},
		{
			name: "TC-UC-05: matched ExpectedError with empty AttrValue falls back to err.Error()",
			err:  errFoo,
			expected: []ExpectedError{
				{Err: errFoo, AttrKey: AttrKeyAppValidationError},
			},
			contextMsg:    "validating foo",
			wantStatus:    codes.Unset,
			wantAttrKey:   AttrKeyAppValidationError,
			wantAttrValue: errFoo.Error(),
		},
		{
			name: "TC-UC-06: unmatched error routes to FailSpan (status=Error, description=context msg)",
			err:  errInfraOOM,
			expected: []ExpectedError{
				{Err: errFoo, AttrKey: AttrKeyAppResult, AttrValue: "not_found"},
			},
			contextMsg: "fetching widget",
			wantStatus: codes.Error,
		},
		{
			name: "TC-UC-07: nil error is a no-op",
			err:  nil,
			expected: []ExpectedError{
				{Err: errFoo, AttrKey: AttrKeyAppResult, AttrValue: "not_found"},
			},
			contextMsg:     "noop",
			wantNoSpanCall: true,
		},
		{
			name: "TC-UC-08: wrapped expected error still matches via errors.Is",
			err:  fmt.Errorf("repo: %w", errFoo),
			expected: []ExpectedError{
				{Err: errBar, AttrKey: AttrKeyAppResult, AttrValue: "duplicate"},
				{Err: errFoo, AttrKey: AttrKeyAppResult, AttrValue: "not_found"},
			},
			contextMsg:    "loading foo",
			wantStatus:    codes.Unset,
			wantAttrKey:   AttrKeyAppResult,
			wantAttrValue: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := tracetest.NewInMemoryExporter()
			span, tp := newTestSpan(exp)

			ClassifyError(span, tt.err, tt.expected, tt.contextMsg)

			span.End()
			flushErr := tp.ForceFlush(context.Background())
			require.NoError(t, flushErr)

			finished := exp.GetSpans()
			require.Len(t, finished, 1)

			stub := finished[0]

			if tt.wantNoSpanCall {
				assert.Equal(t, codes.Unset, stub.Status.Code, "nil error must leave span status Unset")
				assert.Empty(t, stub.Events, "nil error must not record any events")
				assert.Empty(t, stub.Attributes, "nil error must not set any attributes")
				return
			}

			assert.Equal(t, tt.wantStatus, stub.Status.Code)

			if tt.wantAttrKey != "" {
				wantAttr := attribute.String(tt.wantAttrKey, tt.wantAttrValue)
				assert.Contains(t, stub.Attributes, wantAttr,
					"expected attribute %s=%s on span", tt.wantAttrKey, tt.wantAttrValue)
			}

			if tt.wantStatus == codes.Error {
				assert.Equal(t, tt.contextMsg, stub.Status.Description,
					"FailSpan must set status description to contextMsg")
				// FailSpan calls RecordError, which produces an "exception" event.
				// Note: the richer error.type attribute is added by TASK-1 (FailSpan
				// enrichment) and validated by TC-UC-32 in TASK-3 — not asserted here
				// since TASK-1 may not be merged into this worktree's base yet.
				foundException := false
				for _, ev := range stub.Events {
					if ev.Name == "exception" {
						foundException = true
					}
				}
				assert.True(t, foundException,
					"FailSpan must record the error as an 'exception' event")
			}
		})
	}
}
