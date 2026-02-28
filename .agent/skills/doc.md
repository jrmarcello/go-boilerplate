# Skill System Documentation

## Overview

Skills are reusable knowledge modules that agents load on-demand to perform specialized tasks.
Each skill is a directory containing a `SKILL.md` file with structured knowledge.

## Loading Skills

Skills are loaded by agents based on:

1. The workflow being executed
2. The type of task (coding, testing, debugging, deploying)
3. Explicit user request

## Skill Categories

### Core Development

| Skill | Purpose |
| --- | --- |
| `go-patterns` | Go idioms, error handling, DI, Value Objects |
| `clean-code` | Naming, functions, anti-patterns |
| `api-patterns` | HTTP responses, pagination, Swagger |
| `testing-patterns` | Table-driven, mocks, E2E, benchmarks |
| `database-design` | Repositories, migrations, ULID, pools |

### Infrastructure & Quality

| Skill | Purpose |
| --- | --- |
| `tdd-workflow` | Red-Green-Refactor cycle |
| `lint-and-validate` | go vet, gofmt, golangci-lint |
| `k8s-argocd-deploy` | Kubernetes, ArgoCD, Kustomize |
| `deployment-procedures` | Deploy checklist, rollback |
| `server-management` | Probes, HPA, OTel, monitoring |
| `go-performance` | pprof, benchmarks, k6 |
| `code-review-checklist` | Review criteria |
| `vulnerability-scanner` | gosec, govulncheck, OWASP |
| `systematic-debugging` | 4-phase debugging |

### Meta & Process

| Skill | Purpose |
| --- | --- |
| `brainstorming` | Socratic questioning, decision-making |
| `plan-writing` | Task decomposition and sequencing |
| `documentation-templates` | ADR, godoc, changelog templates |
| `doc-coauthoring` | Documentation co-authoring workflow |
| `skill-creator` | Guide for creating new skills |
| `context7` | MCP for library documentation |
| `behavioral-modes` | 6 operational modes |
| `parallel-agents` | Subagent orchestration patterns |
| `architecture` | Architecture decision framework |

## Creating New Skills

See `skills/skill-creator/SKILL.md` for the complete guide.

## Directory Structure

```text
.agent/skills/
├── doc.md                    # This file
├── <skill-name>/
│   ├── SKILL.md              # Main skill definition (required)
│   ├── references/           # Supporting documents (optional)
│   └── scripts/              # Automation scripts (optional)
```
