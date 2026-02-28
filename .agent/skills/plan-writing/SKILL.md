---
name: plan-writing
description: Task decomposition — break complex tasks into actionable steps with clear deliverables
---

# Plan Writing

## Structure

```text
1. GOAL     → What is the end state?
2. SCOPE    → What's included/excluded?
3. STEPS    → Ordered, actionable tasks
4. RISKS    → What could go wrong?
5. VALIDATE → How do we verify success?
```

## Task Decomposition Rules

- Each step has a clear deliverable (file, test, config)
- Steps are small enough to verify individually
- Include "validate" steps after implementation blocks
- Mark dependencies between steps

## Template

```markdown
# Plan: [Title]

## Goal
[One sentence description of desired end state]

## Scope
- Include: [what's in scope]
- Exclude: [what's out of scope]

## Steps
1. [ ] [Action] → [Deliverable]
2. [ ] [Action] → [Deliverable]
3. [ ] Validate: [How to verify steps 1-2]
...

## Risks
- [Risk] → [Mitigation]

## Validation
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] [Domain-specific checks]
```

## Sizing

| Task Size | Steps | Self-Assessment |
| --- | --- | --- |
| Small | 1-3 | Single file change |
| Medium | 4-8 | Multiple files, same domain |
| Large | 9-15 | Cross-cutting, multi-layer |
| Epic | 15+ | Break into smaller plans |
