// Package telemetry — span-naming helpers.
package telemetry

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// HTTPSpanName builds an HTTP root-span name in the form
// `http.<method>.<resource>`, following the Appmax OpenTelemetry naming
// convention (snake_case throughout).
//
// Conversion rules applied to routeTemplate:
//   - Lowercase the method (`GET` -> `get`).
//   - Strip the leading API-version segment if present (`/v1/`, `/api/`).
//   - Replace `:param` placeholders with `by_param`.
//   - Replace `/` with `_`, then collapse runs of underscores.
//   - Empty routeTemplate (e.g. unmatched route, 404) yields the resource `unknown`.
//
// Example: HTTPSpanName("GET", "/v1/users/:id") -> "http.get.users_by_id".
func HTTPSpanName(method, routeTemplate string) string {
	verb := strings.ToLower(strings.TrimSpace(method))
	if verb == "" {
		verb = "unknown"
	}

	resource := httpResource(routeTemplate)
	return "http." + verb + "." + resource
}

// httpResource turns a Gin route template into the snake_case resource segment
// of an HTTP span name. See HTTPSpanName for the full rule set.
func httpResource(routeTemplate string) string {
	r := strings.TrimSpace(routeTemplate)
	if r == "" {
		return "unknown"
	}

	// Strip leading version/api prefix if present.
	r = strings.TrimPrefix(r, "/")
	r = strings.TrimPrefix(r, "v1/")
	r = strings.TrimPrefix(r, "api/")

	// Walk segments and replace ":param" with "by_param".
	segments := strings.Split(r, "/")
	for i, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			segments[i] = "by_" + strings.TrimPrefix(seg, ":")
		}
	}

	joined := strings.Join(segments, "_")
	joined = collapseUnderscores(joined)
	joined = strings.Trim(joined, "_")

	if joined == "" {
		return "unknown"
	}
	return strings.ToLower(joined)
}

// collapseUnderscores compresses any run of `_` to a single `_`.
func collapseUnderscores(s string) string {
	if !strings.Contains(s, "__") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	prevUnderscore := false
	for _, r := range s {
		if r == '_' {
			if prevUnderscore {
				continue
			}
			prevUnderscore = true
		} else {
			prevUnderscore = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

// DBSpanName builds a database child-span name in the form
// `db.<op>.<table>`, following the Appmax OpenTelemetry naming convention.
//
// Conversion rules:
//   - Both op and table are lowercased and trimmed.
//   - Whitespace runs and uppercase letters collapse to snake_case
//     (e.g. "INSERT" -> "insert", "users by email" -> "users_by_email").
//   - Runs of underscores are collapsed to a single underscore and trimmed.
//
// Examples:
//
//	DBSpanName("insert", "users")               -> "db.insert.users"
//	DBSpanName("select", "users_by_id")         -> "db.select.users_by_id"
//	DBSpanName("INSERT", "Users")               -> "db.insert.users"
func DBSpanName(op, table string) string {
	return "db." + dbSegment(op) + "." + dbSegment(table)
}

// dbSegment normalizes a DB span-name segment to lowercase snake_case.
func dbSegment(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "unknown"
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '\t':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := collapseUnderscores(b.String())
	out = strings.Trim(out, "_")
	if out == "" {
		return "unknown"
	}
	return out
}

// StartDBSpan opens a child span for a database operation, named according to
// DBSpanName. The returned ctx carries the new span; callers must close the
// span themselves (typically via `defer span.End()`).
//
// Per ADR-009, infrastructure code MUST NOT mark the span as failed — repository
// methods let errors bubble up so the use case classifies them via
// telemetry.FailSpan / telemetry.WarnSpan / shared.ClassifyError.
func StartDBSpan(ctx context.Context, op, table string) (context.Context, trace.Span) {
	return otel.Tracer("db").Start(ctx, DBSpanName(op, table))
}
