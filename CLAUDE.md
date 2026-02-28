# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go boilerplate/template for microservices, part of the Appmax ecosystem. Uses Clean Architecture with PostgreSQL, Redis cache, and OpenTelemetry observability. Hosted on Bitbucket, deployed to AWS EKS via ArgoCD with Kustomize overlays.

This project serves as a **starter template** — clone it and rename `entity_example` to your domain entity.

## Common Commands

```bash
make dev            # Start server with hot reload (air)
make lint           # Run go vet + gofmt
make lint-full      # Run golangci-lint (same as CI)
make test           # Run all tests: go test ./... -v
make test-unit      # Unit tests only: go test ./internal/... -v
make test-e2e       # E2E tests (requires Docker): go test ./tests/e2e/... -v -count=1
make test-coverage  # Generate HTML coverage report
make docker-up      # Start infrastructure containers (Postgres, Redis)
make docker-down    # Stop infrastructure containers
make migrate-up     # Run database migrations
make migrate-create NAME=add_something  # Create new migration
make kind-setup     # Full Kind cluster setup (cluster + db + migrate + deploy)
make help           # See all available make targets
```

Run a single test file or function:

```bash
go test ./internal/usecases/entity_example/ -run TestCreateUseCase -v
```

Generate Swagger docs (required before CI lint passes):

```bash
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

## Architecture

**Clean Architecture** with strict dependency rule: `domain` <- `usecases` <- `infrastructure`

### Layer Structure

- **`internal/domain/entity_example/`** - Entities, value objects (ID, Email), domain errors. Zero external dependencies.
- **`internal/usecases/entity_example/`** - One file per use case (create.go, get.go, update.go, delete.go, list.go). Each use case defines its own interfaces in `interfaces/` subdirectory. DTOs live in `dto/` subdirectory.
- **`internal/infrastructure/`** - All external concerns:
  - `web/handler/` - Gin HTTP handlers, translate domain errors to HTTP responses via `httputil.SendSuccess`/`httputil.SendError`
  - `web/router/` - Route registration, middleware wiring
  - `web/middleware/` - Logger, metrics, rate limiting, idempotency, service key auth
  - `db/postgres/repository/` - sqlx repository implementations
  - `cache/` - Internal Redis client (legacy, use `pkg/cache` for new code)
  - `telemetry/` - Business-specific metrics (entity counters)
- **`pkg/`** - Reusable packages shared across services:
  - `apperror/` - Structured application errors (AppError with code, message, HTTP status)
  - `httputil/` - Standardized API response helpers (SendSuccess, SendError)
  - `ctxkeys/` - Typed context key definitions
  - `logutil/` - Structured logging with context propagation
  - `telemetry/` - OpenTelemetry setup (traces + HTTP metrics + DB pool metrics)
  - `cache/` - Cache interface and Redis implementation
  - `database/` - PostgreSQL connection with Writer/Reader cluster
- **`config/`** - Configuration loading (godotenv + env vars)
- **`cmd/api/`** - Application entrypoint and manual DI wiring in `server.go`

### Key Patterns

- **Manual DI**: All wiring happens in `cmd/api/server.go:buildDependencies()`. No DI framework. Use cases accept interfaces via constructor, optional dependencies (cache) via `.WithCache()` builder method.
- **ID Strategy**: ULID for all entity IDs. See `docs/adr/002-ulid.md`.
- **DB Cluster**: Writer/Reader split via `pkg/database.DBCluster`. Reader is optional, falls back to writer.
- **API Response Format**: Always use `httputil.SendSuccess(c, status, data)` and `httputil.SendError(c, status, message)`. Responses wrap in `{"data": ...}` or `{"errors": {"message": ...}}`.
- **Error Handling**: Domain defines pure errors (`entity.ErrNotFound`, etc.). `pkg/apperror.AppError` provides structured errors with HTTP status. Handlers translate errors via `handler.HandleError()`. Never return HTTP concepts from domain layer.
- **Service Key Auth**: Optional service-to-service authentication via `X-Service-Name` + `X-Service-Key` headers. See `docs/adr/005-service-key-auth.md`.

### Conventions

- **Commit messages**: `type(scope): description` (feat, fix, refactor, docs, test, chore)
- **Error variable naming**: Use unique names to avoid shadowing (`parseErr`, `saveErr`, `bindErr` instead of reusing `err`)
- **Pre-commit hooks**: Lefthook runs `gofmt`, `go vet`, `golangci-lint` on staged `.go` files
- **Migrations**: Goose SQL files in `internal/infrastructure/db/postgres/migration/`
- **Tests**: Unit tests use hand-written mocks (`mocks_test.go` per package). E2E tests use TestContainers (Postgres + Redis).
- **Import rule**: Never import `infrastructure` from `domain` or `usecases`. Never import `usecases` from `domain`.

## CI Pipeline (Bitbucket)

PRs run: `swag init` -> `golangci-lint run` -> `go test ./internal/...` with coverage. Branch pushes to `develop`/`main` build Docker image, push to ECR, and update Kustomize image tags.

## Agent Toolkit (`.agent/`)

This project has a comprehensive AI agent toolkit in `.agent/`. **Read `.agent/ARCHITECTURE.md` for the full index.** Before starting any task, check if there is a relevant workflow, skill, or agent definition.

### Key resources

- **Rules**: `.agent/rules/RULES.md` — Governance rules, request classification, agent routing
- **Workflows**: `.agent/workflows/` — Step-by-step guides for common tasks (debug, enhance, deploy, test, plan, brainstorm, orchestrate, status)
- **Skills**: `.agent/skills/` — Reusable knowledge modules (go-patterns, clean-code, api-patterns, testing-patterns, database-design, architecture, etc.)
- **Agents**: `.agent/agents/` — Specialized agent definitions (backend-specialist, test-engineer, debugger, database-architect, etc.)
- **Scripts**: `.agent/scripts/checklist.py` (quality gate), `.agent/scripts/verify_all.py` (full verification suite)

### How to use

1. For feature work → read `.agent/workflows/enhance.md`
2. For debugging → read `.agent/workflows/debug.md`
3. For testing → read `.agent/workflows/test.md`
4. For deployment → read `.agent/workflows/deploy.md`
5. Load relevant skills as needed (e.g., `go-patterns`, `api-patterns`)

## Agent Workflow

Always follow these execution directives when working in this repository.

### 1) Mandatory implementation cycle

For every non-trivial task, execute in this order:

1. **Plan** — define scope, affected files, risks, and validation strategy.
2. **Implement** — apply changes strictly following project architecture and conventions.
3. **Test** — create/update tests when needed and run relevant automated checks.
4. **Validate** — perform complete functional verification to ensure the planned behavior actually works.

Do not finish a task without concrete validation evidence.

### 2) Prefer subagents and parallelization

- Use subagents whenever there are independent discovery/analysis tracks.
- Parallelize read-only investigation and validation tasks whenever possible.
- Merge findings into a single concise execution plan before coding.

### 3) Architecture guard

- Before creating or modifying files, verify the change respects layer boundaries.
- Run `make lint` before considering any task complete.
- Consult `AGENTS.md` for detailed rules and patterns.
