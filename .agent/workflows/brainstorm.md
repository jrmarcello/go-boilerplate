---
name: brainstorm
description: Guided brainstorming workflow for design decisions
trigger: "brainstorm|ideate|think about|explore options"
---

# Brainstorm Workflow

## Load Skills

- `brainstorming`
- `architecture`

## Steps

### 1. Define Problem

- What are we trying to solve?
- What constraints exist (architecture, time, complexity)?
- What does success look like?

### 2. Generate Options

Generate at least 3 approaches. For each option:

```text
Option N: [Name]
- How it works: [Brief description]
- Pros: [Benefits]
- Cons: [Trade-offs]
- Complexity: Low | Medium | High
- Architecture impact: None | Minor | Major (needs ADR)
```

### 3. Evaluate

Rate each option on:

- Fits Clean Architecture?
- Maintainability
- Performance implications
- Implementation effort
- Testing complexity

### 4. Recommend

Present recommendation with rationale. If architectural, propose ADR.

### 5. Decide

Wait for user confirmation before proceeding to implementation.
