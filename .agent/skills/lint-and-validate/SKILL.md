---
name: lint-and-validate
description: Go linting and validation — go vet, gofmt, golangci-lint, Lefthook pre-commit hooks
---

# Lint and Validate

## Commands

```bash
make lint          # go vet + gofmt (quick)
make lint-full     # golangci-lint (same as CI)
make security      # gosec only
```

## Lefthook Pre-Commit

Configured in `lefthook.yml`. Runs automatically on `git commit`:

- `gofmt` on staged `.go` files
- `go vet` on staged packages
- `golangci-lint` on staged files

## Common Lint Issues

| Issue | Fix |
| --- | --- |
| `err` shadowing | Use unique names: `parseErr`, `saveErr` |
| Unused variable | Remove or use `_` |
| Missing error check | Handle error or assign to `_` with comment |
| Formatting | Run `gofmt -w .` |
| Import order | Group: stdlib, external, internal |

## golangci-lint

```bash
# Run all linters
golangci-lint run ./...

# Run specific linter
golangci-lint run --enable-only gosec ./...

# Auto-fix when possible
golangci-lint run --fix ./...
```

## Validation Cycle

Before any commit or PR:

1. `make lint` — basic checks
2. `make test` — all tests pass
3. `gofmt -l .` — no unformatted files
