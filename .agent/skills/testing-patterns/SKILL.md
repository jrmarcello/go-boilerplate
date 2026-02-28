---
name: testing-patterns
description: Go testing patterns — table-driven tests, subtests, manual mocks, TestContainers, httptest, benchmarks
---

# Testing Patterns

## Table-Driven Tests (Standard)

```go
func TestNewEmail(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "invalid", true},
        {"empty", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, createErr := vo.NewEmail(tt.input)
            if (createErr != nil) != tt.wantErr {
                t.Errorf("NewEmail(%q) error = %v, wantErr %v", tt.input, createErr, tt.wantErr)
            }
        })
    }
}
```

## Manual Mocks (Project Standard)

Each use case package has `mocks_test.go`:

```go
type mockRepository struct {
    findByIDFn func(ctx context.Context, id string) (*entity.EntityExample, error)
    saveFn     func(ctx context.Context, e *entity.EntityExample) error
    deleteFn   func(ctx context.Context, id string) error
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*entity.EntityExample, error) {
    return m.findByIDFn(ctx, id)
}
```

## Use Case Tests

```go
func TestGetUseCase_Execute(t *testing.T) {
    t.Run("returns entity when found", func(t *testing.T) {
        expected := &entity.EntityExample{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Name: "Test"}
        repo := &mockRepository{
            findByIDFn: func(_ context.Context, _ string) (*entity.EntityExample, error) {
                return expected, nil
            },
        }
        uc := NewGetUseCase(repo)

        result, execErr := uc.Execute(context.Background(), "01ARZ3NDEKTSV4RRFFQ69G5FAV")
        if execErr != nil {
            t.Fatalf("unexpected error: %v", execErr)
        }
        if result.Name != "Test" {
            t.Errorf("got %q, want %q", result.Name, "Test")
        }
    })

    t.Run("returns error when not found", func(t *testing.T) {
        repo := &mockRepository{
            findByIDFn: func(_ context.Context, _ string) (*entity.EntityExample, error) {
                return nil, entity.ErrNotFound
            },
        }
        uc := NewGetUseCase(repo)

        _, execErr := uc.Execute(context.Background(), "nonexistent")
        if !errors.Is(execErr, entity.ErrNotFound) {
            t.Errorf("got %v, want %v", execErr, entity.ErrNotFound)
        }
    })
}
```

## E2E with TestContainers

Located in `tests/e2e/`. Uses real Postgres + Redis containers.

```bash
make test-e2e  # requires Docker
```

## Benchmarks

```go
func BenchmarkCreateEntity(b *testing.B) {
    repo := &mockRepository{saveFn: func(_ context.Context, _ *entity.EntityExample) error { return nil }}
    uc := NewCreateUseCase(repo)
    input := dto.CreateInput{Name: "Bench", Email: "bench@test.com"}

    for i := 0; i < b.N; i++ {
        uc.Execute(context.Background(), input)
    }
}
```

## Commands

```bash
make test           # All tests
make test-unit      # Unit only (pkg/ + config/ + internal/)
make test-e2e       # E2E (Docker required)
make test-coverage  # HTML coverage report
go test -race ./internal/...  # Race detector
```
