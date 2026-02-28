---
name: parallel-agents
description: Subagent orchestration — when and how to spawn parallel agents for efficient task completion
---

# Parallel Agents

## When to Parallelize

- Independent file reads across different directories
- Multiple search queries for different topics
- Creating files with no dependencies between them
- Validation checks that don't affect each other

## When NOT to Parallelize

- Sequential file edits (risk of conflicts)
- Operations that depend on previous results
- Terminal commands (must run sequentially)

## Patterns

### Research Fan-Out

```text
Orchestrator
├── Agent 1: Read domain layer files
├── Agent 2: Read infrastructure layer files
├── Agent 3: Read test files
└── Merge findings → single execution plan
```

### Implementation + Validation

```text
Step 1: Implement (sequential)
Step 2: Validate in parallel
├── Agent 1: Run lint
├── Agent 2: Run tests
└── Agent 3: Check architecture
```

### Batch Creation

```text
Orchestrator
├── Agent 1: Create files A, B, C (independent)
├── Agent 2: Create files D, E, F (independent)
└── Agent 3: Create files G, H, I (independent)
```

## Rules

- Each agent must be self-contained (include all context in prompt)
- Agents are stateless — one prompt, one response
- Prefer fewer, larger agent tasks over many small ones
- Always merge results before proceeding
