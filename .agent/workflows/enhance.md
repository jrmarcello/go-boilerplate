---
name: enhance
description: Feature implementation workflow — plan, implement, test, validate
trigger: "implement|add feature|create|build|enhance"
---

# Enhance Workflow

## Load Skills

- `go-patterns`
- `clean-code`
- `testing-patterns`
- `architecture`

## Steps

### 1. PLAN

- Define the feature scope
- Identify affected layers (domain, usecases, infrastructure)
- List files to create/modify
- Check if ADR is needed

### 2. IMPLEMENT (by layer, inside-out)

#### Domain (if needed)

- Create/update entity in `internal/domain/`
- Add Value Objects in `vo/`
- Define domain errors in `errors.go`

#### Use Cases

- Create use case file: `internal/usecases/<entity>/<action>.go`
- Define interface in `interfaces/`
- Create DTO in `dto/`
- Implement `Execute()` method

#### Infrastructure

- Repository: `internal/infrastructure/db/postgres/repository/`
- Handler: `internal/infrastructure/web/handler/`
- Router: `internal/infrastructure/web/router/`
- Wire in `cmd/api/server.go:buildDependencies()`

### 3. TEST

```bash
# Write tests alongside implementation
go test ./internal/domain/<entity>/ -v
go test ./internal/usecases/<entity>/ -v
make test
```

### 4. VALIDATE

```bash
make lint
make test
# Update swagger if API changed
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

### 5. COMMIT

```bash
git add .
git commit -m "feat(<scope>): <description>"
```
