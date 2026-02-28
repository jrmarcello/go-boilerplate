---
name: systematic-debugging
description: Systematic debugging in 4 phases — Reproduce, Isolate, Understand, Fix with Go tools
---

# Systematic Debugging

## 4-Phase Process

### Phase 1: REPRODUCE

- Get exact reproduction steps
- Determine reproduction rate (100%? intermittent?)
- Document expected vs actual behavior

### Phase 2: ISOLATE

- When did it start? What changed?
- Which component is responsible?
- Create minimal reproduction case

```bash
# Check recent changes
git log --oneline -20

# Bisect to find breaking commit
git bisect start
git bisect bad HEAD
git bisect good <known-good-commit>
```

### Phase 3: UNDERSTAND (Root Cause)

Apply "5 Whys" technique:

```text
1. Why is the response 500? → Use case returns error
2. Why does use case error? → Repository returns nil
3. Why does repository return nil? → Query has wrong filter
4. Why is filter wrong? → Missing parameter validation
5. Why no validation? → Value Object not used ← ROOT CAUSE
```

### Phase 4: FIX AND VERIFY

1. Write regression test FIRST
2. Fix the root cause
3. Verify fix passes new test
4. Check similar code for same bug
5. `make lint && make test`

## Go Debug Tools

```bash
# Delve debugger
dlv test ./internal/usecases/entity_example/ -- -test.run TestCreate

# Race detector
go test -race ./internal/...

# CPU/Memory profiling
go test -cpuprofile=cpu.prof -bench=. ./path/
```
