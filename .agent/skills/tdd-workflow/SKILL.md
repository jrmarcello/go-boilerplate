---
name: tdd-workflow
description: Test-Driven Development for Go — Red-Green-Refactor cycle applied to Value Objects, Use Cases, and Handlers
---

# TDD Workflow — Go

## Red-Green-Refactor Cycle

```text
🔴 RED    → Write a failing test (define expected behavior)
🟢 GREEN  → Write minimum code to pass the test
🔵 REFACT → Improve code without changing behavior
```

## TDD for Value Objects

```go
// 1. RED: Write test first
func TestNewEmail_InvalidFormat(t *testing.T) {
    _, createErr := vo.NewEmail("invalid")
    if createErr == nil {
        t.Error("expected error for invalid email")
    }
}

// 2. GREEN: Implement minimum
func NewEmail(s string) (Email, error) {
    if !strings.Contains(s, "@") {
        return Email{}, ErrInvalidEmail
    }
    return Email{value: s}, nil
}

// 3. REFACTOR: Improve validation
func NewEmail(s string) (Email, error) {
    if _, parseErr := mail.ParseAddress(s); parseErr != nil {
        return Email{}, ErrInvalidEmail
    }
    return Email{value: s}, nil
}
```

## TDD for Use Cases

1. Write test with mock repository
2. Implement use case Execute()
3. Add edge cases (not found, duplicate, validation error)

## Rules

- Never write production code without a failing test
- Tests describe behavior, not implementation
- Refactor only when tests are green
- Run `make test` after every refactor cycle
