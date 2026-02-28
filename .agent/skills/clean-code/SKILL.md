---
name: clean-code
description: Clean, idiomatic Go code — naming, function design, error handling, interfaces, Value Objects, and anti-patterns
---

# Clean Code — Go

## Naming Conventions

### Packages

- Lowercase, single word, no underscores: `entity`, `vo`, `handler`
- No stuttering: `entity.EntityExample` not `entity.EntityExampleEntity`
- Package name provides context

### Variables & Functions

- Exported: `PascalCase` — `FindByID`, `ErrNotFound`
- Unexported: `camelCase` — `buildEntity`, `parseDate`
- Acronyms: all caps — `ID`, `HTTP`, `URL`, `ULID`
- Receivers: 1-2 letter abbreviation — `(uc *CreateUseCase)`, `(e *EntityExample)`

### Error Variables

Use unique names to **avoid shadowing**:

```go
// ✅ Unique error names
parseErr := Parse(input)
saveErr := repo.Save(ctx, entity)

// ❌ Reusing err — shadow bugs
err := Parse(input)
err := repo.Save(ctx, entity)
```

---

## Function Design

### Small Functions, Single Purpose

Each function does one thing. If it needs a comment explaining *what* it does, split it.

### Early Returns (Guard Clauses)

```go
// ✅ Guard clauses reduce nesting
func (uc *CreateUseCase) Execute(ctx context.Context, input dto.CreateInput) (*dto.Output, error) {
    if validateErr := input.Validate(); validateErr != nil {
        return nil, validateErr
    }

    entity, buildErr := entity.New(input.Name, input.Email)
    if buildErr != nil {
        return nil, buildErr
    }

    if saveErr := uc.Repo.Save(ctx, entity); saveErr != nil {
        return nil, saveErr
    }

    return toOutput(entity), nil
}
```

---

## Interface Design

### Accept Interfaces, Return Structs

```go
// ✅ Constructor accepts interface
func NewGetUseCase(repo interfaces.Repository) *GetUseCase {
    return &GetUseCase{Repo: repo}
}
```

### Keep Interfaces Small

1-2 methods. Compose multiple small interfaces if needed.

---

## Anti-Patterns to Avoid

| Anti-Pattern | Why Bad | Fix |
| --- | --- | --- |
| `err` reuse | Shadow bugs | Unique names: `parseErr`, `saveErr` |
| Logic in handler | Violates architecture | Move to use case |
| `c.JSON` directly | Inconsistent responses | Use `httputil.SendSuccess/SendError` |
| `panic` for validation | Crashes server | Return error |
| Commented code | Dead weight | Delete it |
| Infrastructure in domain | Layer violation | Use interfaces |

---

## Commit Messages

Format: `type(scope): description`

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

```text
feat(entity): add email validation value object
fix(handler): correct error status for not found
refactor(usecase): extract cache logic to builder pattern
```
