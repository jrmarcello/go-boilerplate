---
name: go-patterns
description: Core Go idioms — error handling, interfaces, DI, value objects, context, concurrency, and Clean Architecture
---

# Go Patterns

## Error Handling

### Sentinel Errors (Domain Layer)

Define domain errors as package-level variables. Keep them pure — no HTTP concepts.

```go
// internal/domain/entity_example/errors.go
var (
    ErrNotFound     = errors.New("entity not found")
    ErrInvalidEmail = errors.New("invalid email")
    ErrEmptyName    = errors.New("name is required")
)
```

### AppError (pkg/apperror)

Use `pkg/apperror.AppError` for structured errors with HTTP status codes. These live in **pkg/**, reusable across services.

```go
// internal/usecases/entity_example/errors.go
var (
    ErrInvalidInput  = apperror.BadRequest("INVALID_INPUT", "dados inválidos")
    ErrEntityNotFound = apperror.NotFound("ENTITY_NOT_FOUND", "entidade não encontrada")
    ErrDuplicateEmail = apperror.Conflict("DUPLICATE_EMAIL", "email já cadastrado")
)
```

### Error Wrapping with `%w`

```go
if findErr := uc.Repo.FindByID(ctx, id); findErr != nil {
    return nil, fmt.Errorf("finding entity %s: %w", id, findErr)
}
```

### Unique Variable Names (Avoid Shadowing)

**Always** use descriptive, unique error variable names:

```go
// ✅ Correct — each error has a unique name
entityID, parseErr := vo.ParseID(input.ID)
if parseErr != nil {
    return nil, parseErr
}

entity, findErr := uc.Repo.FindByID(ctx, entityID.String())
if findErr != nil {
    return nil, findErr
}

if saveErr := uc.Repo.Save(ctx, entity); saveErr != nil {
    return nil, saveErr
}

// ❌ Wrong — reusing `err` causes shadowing
id, err := vo.ParseID(input.ID)
entity, err := uc.Repo.FindByID(ctx, id) // shadows previous err
```

### errors.Is / errors.As

```go
// Check sentinel
if errors.Is(findErr, sql.ErrNoRows) {
    return nil, ErrEntityNotFound
}

// Extract typed error
var appErr *apperror.AppError
if errors.As(execErr, &appErr) {
    // access appErr.Code, appErr.Message, appErr.HTTPStatus
}
```

---

## Interfaces

### Define Where Used (Use Cases Layer)

```go
// internal/usecases/entity_example/interfaces/repository.go
type Repository interface {
    FindByID(ctx context.Context, id string) (*entity.EntityExample, error)
    Save(ctx context.Context, e *entity.EntityExample) error
    Delete(ctx context.Context, id string) error
}
```

### Small Interfaces (1-2 methods preferred)

```go
// ✅ Good — focused interface
type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

---

## Dependency Injection

### Constructor Injection + Builder Pattern

```go
// Required dependency via constructor
func NewGetUseCase(repo interfaces.Repository) *GetUseCase {
    return &GetUseCase{Repo: repo}
}

// Optional dependency via builder
func (uc *GetUseCase) WithCache(cache interfaces.Cache) *GetUseCase {
    uc.Cache = cache
    return uc
}

// Wiring in server.go
getUC := NewGetUseCase(repo).WithCache(cache)
```

### Manual DI in server.go

All wiring happens in `cmd/api/server.go:buildDependencies()`. No DI framework.

---

## Value Objects

```go
// internal/domain/entity_example/vo/id.go
type ID struct {
    value string
}

func NewID() ID {
    return ID{value: ulid.Make().String()}
}

func ParseID(s string) (ID, error) {
    if _, parseErr := ulid.Parse(s); parseErr != nil {
        return ID{}, ErrInvalidID
    }
    return ID{value: s}, nil
}
```

---

## Context.Context

```go
// Always first parameter
func (uc *GetUseCase) Execute(ctx context.Context, id string) (*dto.Output, error) {
    return uc.Repo.FindByID(ctx, id)
}

// Set timeouts
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

---

## Concurrency

```go
// errgroup for parallel work with error handling
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return fetchA(ctx) })
g.Go(func() error { return fetchB(ctx) })
if waitErr := g.Wait(); waitErr != nil {
    return waitErr
}
```
