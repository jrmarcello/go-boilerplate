---
name: orchestrate
description: Multi-agent orchestration workflow for complex cross-cutting tasks
trigger: "orchestrate|coordinate|multi-step|complex task"
---

# Orchestrate Workflow

## Load Skills

- `parallel-agents`
- `plan-writing`
- `architecture`

## Steps

### 1. DECOMPOSE

Break the task into independent tracks:

```text
Task: [Main Goal]
├── Track A: [Domain changes]
├── Track B: [Infrastructure changes]
├── Track C: [Test updates]
└── Track D: [Documentation]
```

### 2. RESEARCH (parallel)

Spawn subagents for read-only discovery:

- Agent 1: Analyze domain impact
- Agent 2: Analyze infrastructure impact
- Agent 3: Analyze test coverage
- Merge findings

### 3. PLAN

Create unified execution plan with:

- Ordered steps respecting layer dependencies
- File ownership (no two agents edit same file)
- Validation checkpoints

### 4. EXECUTE (sequential with parallel validation)

```text
Phase 1: Domain changes (sequential)
  → Validate: tests pass

Phase 2: Use case changes (sequential)
  → Validate: tests pass

Phase 3: Infrastructure changes (sequential)
  → Validate: lint + tests pass

Phase 4: Documentation (parallel)
  → ADR, Swagger, README
```

### 5. FINAL VALIDATION

```bash
make lint
make test
# Check architecture boundaries
grep -rn "infrastructure" internal/domain/ internal/usecases/
```
