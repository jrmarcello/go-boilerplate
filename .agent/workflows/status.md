---
name: status
description: Project status report — health checks, code quality, test coverage, infrastructure state
trigger: "status|health|check|how are we|overview"
---

# Status Workflow

## Load Skills

- `lint-and-validate`
- `server-management`

## Steps

### 1. Code Quality

```bash
# Lint
make lint
make lint-full

# Format check
gofmt -l .
```

### 2. Tests

```bash
# Run all tests
make test

# Coverage
make test-coverage

# Race detection
go test -race ./internal/...
```

### 3. Security

```bash
# Static analysis
make security

# Dependency vulnerabilities
govulncheck ./...
```

### 4. Dependencies

```bash
# Check for outdated
go list -u -m all

# Tidy
go mod tidy
git diff go.mod go.sum
```

### 5. Infrastructure (if running)

```bash
# Docker
docker compose -f docker/docker-compose.yml ps

# Kubernetes
kubectl get pods -n <namespace>
kubectl top pods -n <namespace>
```

### 6. Report

Present summary:

```text
📊 Project Status Report
========================
Lint:       ✅ Pass / ❌ N issues
Tests:      ✅ Pass / ❌ N failures
Coverage:   XX%
Security:   ✅ Clean / ❌ N findings
Deps:       ✅ Current / ⚠️ N outdated
```
