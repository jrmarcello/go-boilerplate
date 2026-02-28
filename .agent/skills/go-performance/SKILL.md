---
name: go-performance
description: Go performance optimization — pprof, benchmarks, k6 load testing, memory optimization, common pitfalls
---

# Go Performance

## Profiling with pprof

```bash
# CPU
go test -cpuprofile=cpu.prof -bench=. ./internal/usecases/entity_example/
go tool pprof -http=:6060 cpu.prof

# Memory
go test -memprofile=mem.prof -bench=. ./internal/usecases/entity_example/

# Goroutines (running app)
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

## Benchmarks

```bash
# Run
go test ./internal/... -bench=. -benchmem -count=5 > results.txt

# Compare before/after
benchstat before.txt after.txt
```

## k6 Load Testing

```bash
k6 run tests/load/scenarios.js --env SCENARIO=smoke
k6 run tests/load/scenarios.js --env SCENARIO=load
```

## Common Pitfalls

| Pitfall | Fix |
| --- | --- |
| String concatenation in loops | Use `strings.Builder` |
| Frequent small allocations | Use `sync.Pool` |
| Unindexed DB queries | Add indexes, `EXPLAIN ANALYZE` |
| Missing connection pool limits | Set `DB_MAX_OPEN_CONNS` |
| No cache for hot reads | Add Redis cache with builder pattern |
| SELECT * | Select only needed columns |

## Optimization Rules

1. **Measure first**: pprof before touching code
2. **Benchmark**: before.txt and after.txt
3. **Simple first**: indexes and cache before algorithm changes
4. **Validate**: `make test` after every optimization
