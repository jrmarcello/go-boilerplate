---
applies-to: "**/*.go"
---
# Go Code Conventions

## Error Handling

- Use unique error variable names to avoid shadowing: `parseErr`, `saveErr`, `fetchErr` (never reuse `err`)
- Wrap errors with context: `fmt.Errorf("creating user: %w", err)`
- Domain errors are pure: `user.ErrNotFound`, `user.ErrDuplicateEmail`
- Never return HTTP status codes from domain or usecases
- Use cases return `*apperror.AppError` via a local `toAppError()` function that maps domain errors to structured application errors
- `apperror.Wrap(err, code, message)` preserves the error chain — `errors.Is()` works through `Unwrap()`
- Handler resolves errors generically via `errors.As(err, &appErr)` + `codeToStatus` map — zero domain imports in the handler layer
- Ref: `docs/guides/error-handling.md`, ADR-009 (created by spec `error-handling-refactor` — references valid after its execution)

## Span Error Classification (OTel)

> Note: The files referenced below (`pkg/telemetry/span.go`, `internal/usecases/shared/classify.go`, `docs/guides/error-handling.md`, ADR-009) are created by the spec `error-handling-refactor` — references will be valid after its execution.

- **Use case decides span status** — not the handler, not the infrastructure layer
- `telemetry.FailSpan(span, err, msg)` for **unexpected** errors (DB timeout, connection reset, 5xx) — marks span as Error and records the error event
- `telemetry.WarnSpan(span, key, value)` for **expected** errors (validation, not found, conflict) — adds a semantic attribute without marking the span as Error
- Handler NEVER calls `span.SetStatus()` or `span.RecordError()` — it only translates `*apperror.AppError` to HTTP status
- Domain layer has zero OTel dependency
- Each use case defines `var xxxExpectedErrors = []error{...}` and calls `shared.ClassifyError(span, err, expectedErrors, "context")`
- Ref: `pkg/telemetry/span.go`, `internal/usecases/shared/classify.go`

## Architecture

- Domain layer: zero external dependencies, only stdlib
- Use cases: define interfaces in `interfaces/` subdirectory, DTOs in `dto/`
- One use case per file: create.go, get.go, update.go, delete.go, list.go
- Handlers: always use `httpgin.SendSuccess(c, status, data)` and `httpgin.SendError(c, status, message)` (from `pkg/httputil/httpgin`)

## DI Pattern

- Constructor injection for required deps (interfaces)
- Builder methods for optional deps: `.WithCache()`
- All wiring in `cmd/api/server.go:buildDependencies()`
- **Tip**: Use `gopherplate wiring` to auto-regenerate `server.go`/`router.go`/`container.go` from detected domains (instead of manual edits)

## Reusable Packages (pkg/)

- Use `pkg/apperror` for structured errors with HTTP status
- Use `pkg/httputil` for standardized API responses
- Use `pkg/cache` for cache interface (not `internal/infrastructure/cache`)
- Use `pkg/database` for DB connections
- Use `pkg/logutil` for structured logging
- Use `pkg/telemetry` for OpenTelemetry setup

## Testing

- Table-driven tests with descriptive names
- Hand-written mocks in `mocks_test.go` per package
- go-sqlmock for repository tests
- No test frameworks beyond stdlib `testing` package + testify assertions
