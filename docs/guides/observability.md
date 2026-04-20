# Observability Guide

Canonical guide for traces, span events, and the logs-vs-traces posture in this
template. Refines the observability decisions of
[ADR-009](../adr/009-error-handling.md) with the span-naming convention,
`ExpectedError` classification, and event catalog introduced by the
[otel-strategy-alignment](../../.specs/otel-strategy-alignment.md) spec.

See also: [error-handling.md](error-handling.md) for the use-case-side checklist
of adding new errors and mapping them to spans.

---

## Three Pillars in This Codebase

| Pillar | Primary use | Where it lives |
|--------|-------------|----------------|
| Traces (OpenTelemetry) | Per-request business observability — flow, cache, idempotency, DB queries, error classification | Spans + attributes + events |
| Metrics (OpenTelemetry) | Aggregates — HTTP rate/latency/Apdex, DB pool saturation, business counters | [pkg/telemetry/metrics_http.go](../../pkg/telemetry/metrics_http.go), [pkg/telemetry/db_pool_metrics.go](../../pkg/telemetry/db_pool_metrics.go), [internal/infrastructure/telemetry/](../../internal/infrastructure/telemetry/) |
| Logs (slog) | Emergency path only — startup, shutdown, panic, infra unreachable | [pkg/logutil/](../../pkg/logutil/), stdlib `log/slog` |

Prefer traces over logs for any state that describes the flow of a single
request. A trace tells the story of that request; a log line is a sentence
taken out of context.

---

## Span Naming Convention

All span names are **snake_case** and **lowercase**. The helpers in
[pkg/telemetry/naming.go](../../pkg/telemetry/naming.go) compute them
deterministically so handlers and repositories never hand-craft names.

| Subsystem | Pattern | Example |
|-----------|---------|---------|
| HTTP | `http.<verb>.<resource>` | `http.get.users`, `http.post.users`, `http.get.users_by_id` |
| DB | `db.<op>.<table>` | `db.insert.users`, `db.select.users_by_id`, `db.delete.roles` |
| Cache (event, not span) | `cache.<action>` | `cache.hit`, `cache.miss`, `cache.set_failed` |
| Idempotency (event, not span) | `idempotency.<action>` | `idempotency.replayed`, `idempotency.locked` |
| Singleflight (event) | `singleflight.<action>` | `singleflight.shared` |

The HTTP rename runs in
[internal/infrastructure/web/middleware/span_rename.go](../../internal/infrastructure/web/middleware/span_rename.go)
**after** `otelgin.Middleware` so `c.FullPath()` is already populated with the
matched route template when the rename executes.

### Soft-delete gotcha

[UserRepository.Delete](../../internal/infrastructure/db/postgres/repository/user.go)
is a **soft delete** implemented as an `UPDATE` (flips `deleted_at`). Its span
is therefore `db.update.users`, not `db.delete.users`. The only hard `DELETE`
on the repository layer is
[RoleRepository.Delete](../../internal/infrastructure/db/postgres/repository/role.go),
which emits `db.delete.roles`. Pick the span name by the **SQL verb executed**,
not by the Go method name.

### gRPC

gRPC spans keep `otelgrpc` default naming (`grpc.<package>.<Service>/<Method>`).
A gopherplate-flavored `grpc.<service>.<method>` convention is deferred to a
future spec — gRPC is out of scope for the current alignment.

---

## Error Classification on Spans

Three functions drive every span-side error decision:

```go
pkg/telemetry.FailSpan(span, err, msg)            // unexpected — status=Error + error.type + stack trace
pkg/telemetry.WarnSpan(span, attrKey, attrValue)  // expected — semantic attribute, span stays Unset/Ok
internal/usecases/shared.ClassifyError(span, err, expected, contextMsg)
```

`ClassifyError` takes `[]ExpectedError{Err, AttrKey, AttrValue}`:

```go
var createExpectedErrors = []ucshared.ExpectedError{
    {Err: vo.ErrInvalidEmail,         AttrKey: ucshared.AttrKeyAppValidationError},
    {Err: userdomain.ErrDuplicateEmail, AttrKey: ucshared.AttrKeyAppResult, AttrValue: "duplicate_email"},
}
```

When `AttrValue` is empty, `ClassifyError` falls back to `err.Error()`. Matching
uses `errors.Is` so wrapped errors (`fmt.Errorf("...: %w", sentinel)`) still
classify correctly.

Semantic keys live in
[internal/usecases/shared/attrkeys.go](../../internal/usecases/shared/attrkeys.go)
— domain packages reference the constants instead of raw string literals:

| Constant | Value | Used for |
|----------|-------|----------|
| `ucshared.AttrKeyAppResult` | `app.result` | Expected business outcomes (`not_found`, `duplicate_email`, `duplicate_role_name`) |
| `ucshared.AttrKeyAppValidationError` | `app.validation_error` | Expected validation failures (invalid email, invalid ID) — value is `err.Error()` |

Canonical examples: [internal/usecases/user/errors.go](../../internal/usecases/user/errors.go)
and [internal/usecases/role/errors.go](../../internal/usecases/role/errors.go).

---

## Span Events — Business Checkpoints

Use [pkg/telemetry.RecordEvent](../../pkg/telemetry/events.go) for flow
checkpoints. Event names follow `<subsystem>.<action>` in snake_case.

### Cache

| Event | Where | Attributes |
|-------|-------|------------|
| `cache.hit` | [user.GetUseCase](../../internal/usecases/user/get.go) | `cache.key` |
| `cache.miss` | [user.GetUseCase](../../internal/usecases/user/get.go) | `cache.key` |
| `cache.set` | [user.GetUseCase](../../internal/usecases/user/get.go) | `cache.key` |
| `cache.set_failed` | [user.GetUseCase](../../internal/usecases/user/get.go) | `cache.key`, `error.message` |
| `cache.invalidated` | [user.UpdateUseCase](../../internal/usecases/user/update.go), [user.DeleteUseCase](../../internal/usecases/user/delete.go) | `cache.key` |
| `cache.invalidate_failed` | [user.UpdateUseCase](../../internal/usecases/user/update.go), [user.DeleteUseCase](../../internal/usecases/user/delete.go) | `cache.key`, `error.message` |

### Singleflight

| Event | Where | Attributes |
|-------|-------|------------|
| `singleflight.shared` | [user.GetUseCase](../../internal/usecases/user/get.go) | `singleflight.key` |

Note: `x/sync/singleflight` reports `shared=true` to every caller that benefited
from dedup, including the leader — so the event appears on every participant
span, not just joiners. The load-bearing invariant is "repo called once per
dedup group," asserted in tests.

### Idempotency

| Event | Where | Attributes |
|-------|-------|------------|
| `idempotency.key_acquired` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key` |
| `idempotency.replayed` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key`, `idempotency.status_code` |
| `idempotency.locked` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key` |
| `idempotency.fingerprint_mismatch` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key` |
| `idempotency.stored` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key`, `idempotency.status_code` |
| `idempotency.released` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key` |
| `idempotency.store_unavailable` | [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go) | `idempotency.key`, `error.message` |

`idempotency.store_unavailable` is the one event that pairs with a retained
`logutil.LogWarn` call, because the store being unreachable is an
emergency-path signal operators need to see in logs.

---

## Logs: When (and Only When)

`slog.*` / `logutil.Log*` calls are allowed in exactly **four categories**:

1. **Startup / shutdown** — `cmd/api/**` lifecycle events, config validation,
   graceful shutdown.
2. **Panic recovery** — `middleware/recovery.go` (the stack trace belongs in a
   log line so it survives after the span is exported).
3. **Access log** — `middleware/logger.go` (per-request HTTP summary — the
   one place where one log line per request is the whole point).
4. **Unreachable-infra warnings** — the fail-open branches of
   [middleware/idempotency.go](../../internal/infrastructure/web/middleware/idempotency.go)
   (Lock/Get/Complete/Unlock errors when Redis is down). These branches carry a
   `// nosemgrep: gopherplate-usecase-no-slog-in-flow` pragma plus an inline
   reason — they are **always** paired with a matching span event
   (`idempotency.store_unavailable`) so traces stay self-contained.

Everything else — cache hit/miss, singleflight dedup, idempotency business
outcomes, validation warnings — goes through span attributes (`WarnSpan`) or
span events (`RecordEvent`).

**Not allowed:**

- `slog.Debug("cache hit")` — use `RecordEvent(span, "cache.hit", ...)`.
- `slog.Warn("failed to invalidate cache")` on a non-fatal error — use
  `RecordEvent(span, "cache.invalidate_failed", ...)`.
- Any `slog.Info` describing the progress of a single request.

Rule of thumb: **if the information helps debug a single request, it belongs on
the span.** If it describes process health, consider a metric before a log.

The rule is enforced by
[.semgrep/observability.yml](../../.semgrep/observability.yml)
(`gopherplate-usecase-no-slog-in-flow`). Run `make semgrep` locally before
committing.

---

## Layer Responsibilities (ADR-009 refinement)

| Layer | Span status | Span attributes | Span events | slog |
|-------|-------------|-----------------|-------------|------|
| **Domain** | No access | No access | No access | No |
| **Use case** | Owns `SetStatus` via `ClassifyError` | Owns `WarnSpan` via `ExpectedError` | Owns `RecordEvent` for flow checkpoints | No |
| **Infrastructure (repo, cache, etc.)** | NEVER sets status — errors bubble up | Child spans from `StartDBSpan`/`otelhttp` ok | Does NOT emit business events (belongs in use case) | Only emergency path (infra down) |
| **Handler** | Never touches span | Never touches span | Never touches span | No |
| **Middleware** | Renames span (`SpanRename`), records access log (`logger`) | Adds request-scoped attrs (`otelgin`) | Emits cross-cutting events (`idempotency`) | Only emergency path |

---

## Development & Debugging

- Export traces to a local collector (Jaeger/Tempo) in dev. Search by span
  name: `http.post.users` surfaces every user-creation request; filter by
  `app.result=not_found` to see the 404s.
- When diagnosing a "one request failed" ticket, open the trace by `trace_id`
  from the access log — the span events (`cache.hit`, `idempotency.replayed`,
  etc.) tell the story without grepping logs.
- Stack traces live in the span's `exception` event (emitted by `FailSpan`
  via `trace.WithStackTrace(true)`), not in logs.

---

## References

- [ADR-009 — Error Handling Refactor](../adr/009-error-handling.md)
- [error-handling.md — use-case implementation guide](error-handling.md)
- Spec: [.specs/otel-strategy-alignment.md](../../.specs/otel-strategy-alignment.md)
- Assis — "OpenTelemetry — Boas Práticas: Gestão de Erros em Spans com Go e DDD" (internal Confluence)
- Assis — "[Max] OpenTelemetry" (internal Confluence)
