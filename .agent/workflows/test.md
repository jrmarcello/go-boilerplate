---
name: test
description: Testing workflow — write, run, and validate tests across all layers
trigger: "test|coverage|write tests|test this"
---

# Test Workflow

## Load Skills

- `testing-patterns`
- `tdd-workflow`
- `go-patterns`

## Steps

### 1. ASSESS

- What needs testing? (new code, bug fix, refactor)
- Which layer? (domain, usecases, infrastructure)
- What kind of test? (unit, integration, e2e)

### 2. WRITE TESTS

#### Domain Tests

```bash
# Value Objects, entities, business rules
go test ./internal/domain/entity_example/ -v
```

#### Use Case Tests

```bash
# With manual mocks
go test ./internal/usecases/entity_example/ -v -run TestCreate
```

#### E2E Tests (requires Docker)

```bash
make docker-up
make test-e2e
```

### 3. RUN

```bash
# All tests
make test

# Unit tests only
make test-unit

# Specific test
go test ./internal/usecases/entity_example/ -run TestCreateUseCase -v

# With race detector
go test -race ./internal/...

# With coverage
make test-coverage
```

### 4. VALIDATE

- All tests pass?
- Edge cases covered? (nil, empty, invalid, not found, duplicate)
- Error paths tested?
- Table-driven tests where multiple inputs make sense?

### 5. COVERAGE CHECK

```bash
make test-coverage
# Opens HTML report

# Quick coverage stats
go test ./internal/... -cover
```

### Benchmark (optional)

```bash
go test ./internal/usecases/entity_example/ -bench=. -benchmem
```
