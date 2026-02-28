---
name: code-review-checklist
description: Code review checklist for Go — correctness, architecture, security, performance, testing
---

# Code Review Checklist

## Correctness

- [ ] Logic handles all edge cases
- [ ] Error handling is complete (no ignored errors)
- [ ] Error variables have unique names (no shadowing)
- [ ] Context propagated correctly

## Architecture

- [ ] Follows Clean Architecture layers
- [ ] No prohibited imports (infrastructure from domain/usecases)
- [ ] Interfaces defined in use cases layer
- [ ] DI via constructor, optional via builder

## Security

- [ ] No hardcoded secrets
- [ ] SQL uses parameterized queries ($1, $2)
- [ ] Input validated via Value Objects
- [ ] No sensitive data in logs

## Performance

- [ ] No N+1 queries
- [ ] Indexes for query patterns
- [ ] Cache for frequent reads
- [ ] Connection pool sized appropriately

## Testing

- [ ] Unit tests for new/changed code
- [ ] Table-driven tests where applicable
- [ ] Mocks for external dependencies
- [ ] `make test` passes

## Standards

- [ ] `make lint` passes
- [ ] Commit follows `type(scope): description`
- [ ] API responses use `httputil.SendSuccess/SendError`
- [ ] No commented-out code
