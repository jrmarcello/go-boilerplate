# .agent/ Architecture

## Overview

This directory contains the AI agent toolkit for the go-boilerplate project.
It provides structured knowledge, automated workflows, and quality gates
for AI agents working on the codebase.

## Directory Structure

```text
.agent/
├── ARCHITECTURE.md           # This file — master index
├── .markdownlint.json        # Markdown linting config
├── rules/
│   └── RULES.md              # AI governance rules and routing
├── agents/                   # Agent definitions (14 agents)
│   ├── backend-specialist.md
│   ├── orchestrator.md
│   ├── test-engineer.md
│   ├── debugger.md
│   ├── database-architect.md
│   ├── devops-engineer.md
│   ├── security-auditor.md
│   ├── performance-optimizer.md
│   ├── documentation-writer.md
│   ├── project-planner.md
│   ├── product-manager.md
│   ├── explorer-agent.md
│   ├── code-archaeologist.md
│   └── penetration-tester.md
├── skills/                   # Reusable knowledge modules (23 skills)
│   ├── doc.md                # Skill system documentation
│   ├── go-patterns/
│   ├── clean-code/
│   ├── api-patterns/
│   ├── testing-patterns/
│   ├── database-design/
│   ├── tdd-workflow/
│   ├── lint-and-validate/
│   ├── k8s-argocd-deploy/
│   ├── deployment-procedures/
│   ├── server-management/
│   ├── go-performance/
│   ├── code-review-checklist/
│   ├── vulnerability-scanner/
│   ├── systematic-debugging/
│   ├── brainstorming/
│   ├── plan-writing/
│   ├── documentation-templates/
│   ├── doc-coauthoring/
│   ├── skill-creator/
│   ├── context7/
│   ├── behavioral-modes/
│   ├── parallel-agents/
│   └── architecture/
├── workflows/                # Orchestration workflows (8 workflows)
│   ├── brainstorm.md
│   ├── debug.md
│   ├── deploy.md
│   ├── enhance.md
│   ├── orchestrate.md
│   ├── plan.md
│   ├── status.md
│   └── test.md
└── scripts/                  # Automation scripts
    ├── checklist.py          # Quality gate checklist (7 checks)
    └── verify_all.py         # Full pre-deploy verification suite
```

## Agent Catalog

| Agent | Specialty | When to Use |
| --- | --- | --- |
| backend-specialist | Go code, Clean Architecture | Feature implementation |
| orchestrator | Multi-agent coordination | Complex cross-cutting tasks |
| test-engineer | Testing strategy | Writing/fixing tests |
| debugger | Bug investigation | Error diagnosis |
| database-architect | PostgreSQL, migrations | Schema changes, queries |
| devops-engineer | K8s, CI/CD, Docker | Infrastructure tasks |
| security-auditor | Security review | Vulnerability assessment |
| performance-optimizer | Profiling, benchmarks | Performance issues |
| documentation-writer | ADRs, guides, docs | Documentation tasks |
| project-planner | Task decomposition | Planning & estimation |
| product-manager | Requirements, priorities | Feature definition |
| explorer-agent | Codebase exploration | Research & discovery |
| code-archaeologist | Legacy analysis | Tech debt, refactoring |
| penetration-tester | Security testing | Attack simulation |

## Workflow Catalog

| Workflow | Trigger | Purpose |
| --- | --- | --- |
| brainstorm | "explore options" | Design decision making |
| debug | "fix bug" | Systematic bug resolution |
| deploy | "deploy" | Pre-deploy to post-deploy |
| enhance | "add feature" | Feature implementation |
| orchestrate | "complex task" | Multi-agent coordination |
| plan | "plan" | Task planning & scoping |
| status | "status" | Project health report |
| test | "write tests" | Test creation & execution |

## Skill Categories

### Core Development (5)

go-patterns, clean-code, api-patterns, testing-patterns, database-design

### Infrastructure & Quality (9)

tdd-workflow, lint-and-validate, k8s-argocd-deploy, deployment-procedures,
server-management, go-performance, code-review-checklist, vulnerability-scanner,
systematic-debugging

### Meta & Process (9)

brainstorming, plan-writing, documentation-templates, doc-coauthoring,
skill-creator, context7, behavioral-modes, parallel-agents, architecture

## Quick Reference

```bash
# Run quality checks
python3 .agent/scripts/checklist.py

# Full pre-deploy verification
python3 .agent/scripts/verify_all.py

# Quick verification (skip slow checks)
python3 .agent/scripts/verify_all.py --quick
```
