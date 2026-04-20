package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPSpanName(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		routeTemplate  string
		expectedResult string
	}{
		// TC-UC-12
		{
			name:           "TC-UC-12: GET /v1/users",
			method:         "GET",
			routeTemplate:  "/v1/users",
			expectedResult: "http.get.users",
		},
		// TC-UC-13
		{
			name:           "TC-UC-13: GET /v1/users/:id",
			method:         "GET",
			routeTemplate:  "/v1/users/:id",
			expectedResult: "http.get.users_by_id",
		},
		// TC-UC-14
		{
			name:           "TC-UC-14: POST /v1/roles",
			method:         "POST",
			routeTemplate:  "/v1/roles",
			expectedResult: "http.post.roles",
		},
		// TC-UC-15
		{
			name:           "TC-UC-15: GET with empty route",
			method:         "GET",
			routeTemplate:  "",
			expectedResult: "http.get.unknown",
		},
		// Edge: multi-segment with param
		{
			name:           "edge: GET /v1/roles/:id/permissions",
			method:         "GET",
			routeTemplate:  "/v1/roles/:id/permissions",
			expectedResult: "http.get.roles_by_id_permissions",
		},
		// Edge: lowercase method
		{
			name:           "edge: lowercase method normalized",
			method:         "delete",
			routeTemplate:  "/v1/users/:id",
			expectedResult: "http.delete.users_by_id",
		},
		// Edge: /api/ prefix stripped
		{
			name:           "edge: /api/ prefix stripped",
			method:         "GET",
			routeTemplate:  "/api/users",
			expectedResult: "http.get.users",
		},
		// Edge: collapses multiple underscores from consecutive params
		{
			name:           "edge: collapses underscores",
			method:         "GET",
			routeTemplate:  "/v1/users/:id/:sub",
			expectedResult: "http.get.users_by_id_by_sub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTTPSpanName(tt.method, tt.routeTemplate)
			assert.Equal(t, tt.expectedResult, got)
		})
	}
}

func TestDBSpanName(t *testing.T) {
	tests := []struct {
		name           string
		op             string
		table          string
		expectedResult string
	}{
		// TC-UC-16
		{
			name:           "TC-UC-16: insert users",
			op:             "insert",
			table:          "users",
			expectedResult: "db.insert.users",
		},
		// TC-UC-17
		{
			name:           "TC-UC-17: select users_by_id",
			op:             "select",
			table:          "users_by_id",
			expectedResult: "db.select.users_by_id",
		},
		// TC-UC-18: uppercase op + spaces normalized
		{
			name:           "TC-UC-18: uppercase op normalized",
			op:             "INSERT",
			table:          "Users",
			expectedResult: "db.insert.users",
		},
		{
			name:           "TC-UC-18: spaces collapse to underscores",
			op:             "SELECT BY",
			table:          "users by email",
			expectedResult: "db.select_by.users_by_email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DBSpanName(tt.op, tt.table)
			assert.Equal(t, tt.expectedResult, got)
		})
	}
}

// installTestTracerProvider wires an in-memory exporter as the global tracer
// provider for the duration of the test, restoring the previous provider on
// cleanup so cross-test pollution cannot occur.
func installTestTracerProvider(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	})
	return exporter
}

func TestStartDBSpan(t *testing.T) {
	exporter := installTestTracerProvider(t)

	_, span := StartDBSpan(context.Background(), "insert", "users")
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "db.insert.users", spans[0].Name)
}
