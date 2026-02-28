---
name: architecture
description: Architecture decision framework — evaluate changes for layer compliance, dependency direction, and SOLID principles
---

# Architecture

## Clean Architecture Layers

```text
┌────────────────────────────────────────┐
│           Infrastructure               │  Gin, sqlx, Redis, OTel
│  ┌──────────────────────────────────┐  │
│  │          Use Cases               │  │  Business logic, interfaces
│  │  ┌────────────────────────────┐  │  │
│  │  │        Domain             │  │  │  Entities, VOs, errors
│  │  └────────────────────────────┘  │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
```

## Dependency Rule

Dependencies ALWAYS point inward:

- `infrastructure` → `usecases` → `domain` ✅
- `domain` → `usecases` ❌
- `usecases` → `infrastructure` ❌

## Decision Framework

Before any change, ask:

### Layer Check

- Which layer does this belong to?
- Does it cross any layer boundary?
- Are the dependencies pointing inward?

### Interface Check

- Is the interface defined in usecases?
- Does infrastructure implement the interface?
- Is dependency injected via constructor?

### SOLID Check

- **S**: Does this struct have a single responsibility?
- **O**: Can it be extended without modification?
- **L**: Are subtypes substitutable?
- **I**: Are interfaces minimal and focused?
- **D**: Do we depend on abstractions?

## ADR Trigger

Create an ADR (`docs/adr/NNN-title.md`) when:

- Adding a new external dependency
- Changing the data model
- Modifying API contracts
- Introducing a new pattern
- Making irreversible decisions

## Import Validation

```bash
# Check for prohibited imports
grep -rn "infrastructure" internal/domain/ internal/usecases/
# Should return nothing
```
