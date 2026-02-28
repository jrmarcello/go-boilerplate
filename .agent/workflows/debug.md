---
name: debug
description: Systematic debugging workflow using 4-phase approach
trigger: "debug|fix bug|error|not working|broken"
---

# Debug Workflow

## Load Skills

- `systematic-debugging`
- `go-patterns`
- `testing-patterns`

## Steps

### 1. REPRODUCE

- Understand the error/behavior
- Get exact reproduction steps
- Check logs for stack traces

```bash
# Recent logs
make dev   # run locally and reproduce

# If in K8s
kubectl logs -f deployment/go-boilerplate -n <namespace> --tail=100
```

### 2. ISOLATE

- Identify affected layer (domain? usecase? handler? infrastructure?)
- Check recent changes: `git log --oneline -10`
- Run targeted tests:

```bash
go test ./internal/usecases/entity_example/ -run TestSpecific -v
go test -race ./internal/...
```

### 3. ROOT CAUSE

Apply "5 Whys":

1. Why does [symptom] occur?
2. Why does [reason 1] happen?
3. Continue until root cause is found

Common causes by layer:

- **Domain**: Invalid Value Object construction
- **Use Case**: Missing error handling, wrong interface call
- **Handler**: Error translation, binding errors
- **Infrastructure**: SQL query, cache miss handling, connection issues

### 4. FIX AND VERIFY

1. Write a failing test that exposes the bug
2. Fix the root cause
3. Verify: `make lint && make test`
4. Check for similar patterns in codebase
