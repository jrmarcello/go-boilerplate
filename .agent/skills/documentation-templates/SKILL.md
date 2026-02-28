---
name: documentation-templates
description: Templates for ADRs, godoc, changelog, guides, and commit messages
---

# Documentation Templates

## ADR (Architecture Decision Record)

Location: `docs/adr/NNN-title.md`

```markdown
# NNN: Title

## Status

Accepted | Proposed | Deprecated

## Context

[Problem or question that needs a decision]

## Decision

[What we decided and why]

## Consequences

### Positive
- [benefit]

### Negative
- [trade-off]
```

## Godoc

```go
// Package entity_example defines the core domain entity and value objects.
//
// The entity uses ULID as a natural primary key, ensuring uniqueness
// and sortability without database coordination.
package entity_example
```

## Commit Messages

Format: `type(scope): description`

```text
feat(entity): add cache support with builder pattern
fix(handler): prevent error shadowing in list handler
refactor(usecases): extract repository interface
docs(adr): add ADR-009 for pagination strategy
test(usecases): add table-driven tests for create use case
chore(deps): update Go to 1.23
```

## Guide

Location: `docs/guides/topic.md`

```markdown
# Topic Name

## Overview
[What this guide covers]

## Prerequisites
[What you need before starting]

## Steps
[Step-by-step instructions with code examples]

## Troubleshooting
[Common issues and solutions]
```
