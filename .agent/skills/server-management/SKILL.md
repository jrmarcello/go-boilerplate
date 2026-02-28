---
name: server-management
description: Server management — K8s probes, HPA, OpenTelemetry metrics, Redis monitoring, PostgreSQL pool management
---

# Server Management

## Health Probes

| Probe | Endpoint | Purpose |
| --- | --- | --- |
| Liveness | `/health` | App alive? Restart if fails |
| Readiness | `/ready` | Ready for traffic? Remove from LB if fails |

## HPA (Horizontal Pod Autoscaler)

Configured in `deploy/base/hpa.yaml`. Scales based on CPU/memory.

## OpenTelemetry

Setup via `pkg/telemetry.Setup()`:

- **Traces**: gRPC exporter to OTel Collector
- **HTTP Metrics**: Request count, duration, Apdex scoring
- **DB Pool Metrics**: Open, in-use, idle, max connections

```go
// In server.go
provider, _ := pkgtelemetry.Setup(ctx, pkgtelemetry.Config{
    ServiceName:  cfg.OTel.ServiceName,
    CollectorURL: cfg.OTel.CollectorURL,
})
```

## PostgreSQL Pool

Configure via env:

- `DB_MAX_OPEN_CONNS=25`
- `DB_MAX_IDLE_CONNS=5`
- `DB_CONN_MAX_LIFETIME=5m`

Monitor via `pkg/telemetry.RegisterDBPoolMetrics(db)`.

## Redis

Via `pkg/cache.NewRedisClient()`. Nil-safe operations (returns ErrCacheMiss on miss).

## Monitoring

- Grafana/Kibana for logs and metrics
- OpenTelemetry Collector for traces
- Apdex thresholds: 500ms satisfied, 2s tolerating
