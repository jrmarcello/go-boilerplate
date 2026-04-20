package shared

import (
	"errors"

	"github.com/jrmarcello/gopherplate/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

// ExpectedError describes a sentinel error a use case considers part of its
// normal outcome (e.g. validation failures, not-found, duplicate). When
// ClassifyError matches an incoming error against ExpectedError.Err via
// errors.Is, it records a semantic attribute on the active span keyed by
// AttrKey — leaving the span status Unset/Ok rather than marking it failed.
//
// AttrValue is optional: when empty, ClassifyError falls back to err.Error().
// Use AttrValue when the desired vocabulary differs from the raw error
// message (e.g. "not_found", "duplicate_email").
type ExpectedError struct {
	// Err is the sentinel checked via errors.Is, so it sees through
	// fmt.Errorf("%w") wrapping.
	Err error
	// AttrKey is the OpenTelemetry attribute key recorded on the span.
	// Prefer the package-level constants (AttrKeyAppResult,
	// AttrKeyAppValidationError) over raw string literals.
	AttrKey string
	// AttrValue is the recorded attribute value. When empty, ClassifyError
	// substitutes err.Error().
	AttrValue string
}

// ClassifyError inspects err against a list of expected outcomes and routes
// telemetry accordingly:
//   - nil err: no-op (returns immediately).
//   - Match in expected (via errors.Is, supports wrapping): records a
//     semantic attribute via telemetry.WarnSpan keyed by the matched
//     ExpectedError.AttrKey. The value is ExpectedError.AttrValue when
//     non-empty, otherwise err.Error(). Span status remains Unset/Ok.
//   - No match: marks the span as failed via telemetry.FailSpan, attaching
//     contextMsg as the human-readable description.
//
// Use cases own the contract by declaring a domain-local
// `[]ExpectedError` slice next to their error sentinels and passing it on
// every call.
func ClassifyError(span trace.Span, err error, expected []ExpectedError, contextMsg string) {
	if err == nil {
		return
	}

	for _, e := range expected {
		if errors.Is(err, e.Err) {
			value := e.AttrValue
			if value == "" {
				value = err.Error()
			}
			telemetry.WarnSpan(span, e.AttrKey, value)
			return
		}
	}

	telemetry.FailSpan(span, err, contextMsg)
}
