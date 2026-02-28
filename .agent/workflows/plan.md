---
name: plan
description: Planning workflow — scope, decompose, estimate, and sequence tasks
trigger: "plan|roadmap|how should we|strategy"
---

# Plan Workflow

## Load Skills

- `plan-writing`
- `architecture`
- `brainstorming`

## Steps

### 1. UNDERSTAND

- What is the goal?
- What are the constraints?
- What exists today? (research codebase)

### 2. SCOPE

Define boundaries clearly:

```text
IN SCOPE:
- [Feature/change 1]
- [Feature/change 2]

OUT OF SCOPE:
- [Thing we won't do now]
- [Thing for future iteration]
```

### 3. DECOMPOSE

Break into actionable tasks:

```text
1. [ ] [Task] — [Layer] — [Risk: Low/Med/High]
2. [ ] [Task] — [Layer] — [Risk: Low/Med/High]
3. [ ] Validate: make lint && make test
...
```

### 4. SEQUENCE

Order by:

1. Domain first (innermost layer)
2. Use Cases next
3. Infrastructure last
4. Documentation and tests throughout

### 5. RISKS

| Risk | Impact | Mitigation |
| --- | --- | --- |
| [Risk 1] | [Impact] | [How to mitigate] |

### 6. PRESENT

Present the plan for user review before execution.
