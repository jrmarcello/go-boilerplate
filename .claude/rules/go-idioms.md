---
applies-to: "**/*.go"
---
# Go Language Idioms

Canonical Go idioms distilled from the official references. Project-specific
conventions (Clean Architecture layout, apperror, span classification, DI, etc.)
live in [`go-conventions.md`](go-conventions.md) — this file covers
**language-level** idioms only.

## Canonical References

| Topic | Source |
| ----- | ------ |
| Language semantics (authoritative) | https://go.dev/ref/spec |
| Idiomatic style (curated) | https://go.dev/doc/effective_go |
| Error handling guidance | https://go.dev/blog/error-handling-and-go, https://go.dev/blog/go1.13-errors |
| Concurrency patterns | https://go.dev/blog/pipelines, https://go.dev/ref/mem (memory model) |
| Code review checklist (upstream) | https://go.dev/wiki/CodeReviewComments |

When a rule below has a `Ref:` pointer, it names the canonical section — follow
the link before arguing with the rule.

## Naming

- Use `MixedCaps` / `mixedCaps`, never `snake_case` or `SCREAMING_SNAKE`.
- Acronyms stay in one case: `UserID`, `HTTPClient`, `parseURL` — not `UserId`, `HttpClient`, `parseUrl`.
- Short names in small scopes (`i`, `r`, `ctx`, `err`); descriptive names as scope grows.
- Receiver names: short (1–3 letters), consistent across all methods of the same type — `u *User`, not a mix of `u` / `user` / `self`.
- Interface names: `-er` suffix for single-method interfaces (`Reader`, `Closer`, `UserFinder`). Multi-method interfaces take a domain noun (`UserRepository`).
- Getter without `Get` prefix: `user.Name()` — not `user.GetName()`. Setters keep `Set`.
- Package names: short, lowercase, no underscores, no plurals (`user`, not `users` or `user_pkg`).
- Error variables prefixed `Err`: `ErrNotFound`, `ErrDuplicateEmail`. Error *types* suffixed `Error`: `ValidationError`.
- Ref: Effective Go §Names, CodeReviewComments §Initialisms.

## Formatting

- All code must pass `gofmt` and `goimports` — non-negotiable, enforced by pre-commit hook.
- Import groups separated by blank lines: stdlib / third-party / internal. `goimports` handles this.
- Line length: no hard limit, but break long function signatures for readability. Don't line-wrap string literals.
- Ref: Effective Go §Formatting.

## Zero Values

- Design types so the **zero value is useful**. A caller should be able to declare `var s MyStruct` and use it without calling `New`.
- Examples from stdlib: `sync.Mutex{}` is ready-to-use, `bytes.Buffer{}` is ready-to-write.
- If a zero value is invalid, force construction through a constructor that returns an error, and never expose the struct fields.
- Ref: Effective Go §The zero value, Go Spec §The zero value.

## `make` vs `new`

- `new(T)` allocates zeroed storage and returns `*T`. Rarely needed — use `&T{}` for clarity.
- `make(T, …)` initializes slices, maps, and channels (the only types that require internal structure). Never use `new` on these.
- `new(map[string]int)` is a trap: it returns a `*map` pointing at a nil map.
- Ref: Effective Go §Allocation with new, §Allocation with make.

## Slices, Maps, Channels

- Nil slices are valid — `len(nil) == 0`, `append(nil, x)` works. Don't pre-allocate with `make([]T, 0)` unless you need a specific capacity.
- When capacity is known, `make([]T, 0, n)` avoids reallocations in a tight append loop.
- Nil maps are **read-safe** but **write-panics**. Always `make` before writing.
- Range over a map has non-deterministic order — never rely on iteration order.
- Channel direction in types narrows intent: `chan<- T` (send-only), `<-chan T` (receive-only).
- The **sender** closes the channel, never the receiver. Closing twice panics.
- Receive from a closed channel returns the zero value immediately (use `v, ok := <-ch`).

## Interfaces

- Accept interfaces, return concrete types — unless returning an interface is the explicit contract (e.g., `io.Reader`).
- Keep interfaces small. `io.Reader` and `io.Writer` are the gold standard.
- Define the interface on the **consumer** side (the package that uses it), not the provider. This keeps implementations decoupled.
- Don't export an interface "just in case" — export concrete types, add interfaces when a second implementation appears.
- Embedding: `type ReadWriter interface { Reader; Writer }` composes behavior. Prefer composition over large flat interfaces.
- An empty interface accepts anything; prefer `any` (Go 1.18+) over `interface{}` — they're aliases, but `any` reads better.
- Ref: Effective Go §Interfaces and other types, CodeReviewComments §Interfaces.

## Method Sets (Pointer vs Value Receiver)

- Value receiver (`func (u User)`) is cheap for small structs and safe for concurrent use.
- Pointer receiver (`func (u *User)`) is required when the method mutates state or when the struct is large or contains a `sync.Mutex`.
- **Be consistent**: if one method on a type has a pointer receiver, give all methods pointer receivers. Mixed sets confuse interface satisfaction rules.
- Pointer method sets include value methods, but not vice versa — `*T` satisfies interfaces defined over both `T` and `*T`; a value `T` does not satisfy interfaces with pointer-receiver methods.
- Ref: Go Spec §Method sets, CodeReviewComments §Receiver Type.

## Errors

- Return errors as the **last** return value. Never use `panic` for expected failure.
- Wrap with context using `%w`: `fmt.Errorf("creating user %q: %w", id, err)`. Quote variable data with `%q`.
- Inspect with `errors.Is(err, Target)` for sentinels, `errors.As(err, &target)` for typed errors. Never compare with `==` when wrapping is possible.
- Sentinel errors are package-level `var Err... = errors.New(...)`. Custom error types implement `Error() string` and optionally `Unwrap() error`.
- Error messages: lowercase, no trailing punctuation, no `ERROR:` prefix. They're composed into larger messages by callers.
- Don't log *and* return — the caller logs or handles. Logging twice pollutes output.
- Ref: https://go.dev/blog/go1.13-errors, CodeReviewComments §Error Strings.

## Control Flow

- `if`/`else`: prefer early returns over nested conditionals. Happy path stays at the leftmost indent.
- Initialization in `if`: `if v, err := fn(); err != nil { ... }` keeps scope tight.
- `switch` without a tag is the idiomatic long `if-else` chain. `switch x.(type)` for type switches.
- `for` is the only loop keyword. Infinite loops are just `for { ... }`.
- Never shadow loop variables in goroutines: in Go 1.22+ the loop variable is per-iteration, but older patterns `x := x` before `go func()` are still seen in third-party code.
- Ref: Effective Go §Control structures.

## Defer

- Arguments to the deferred call are evaluated **at the `defer` statement**, not when the call executes.
- Deferred calls run in **LIFO** order when the surrounding function returns, even on panic.
- Classic pattern: `defer f.Close()` immediately after the acquiring call (`f, err := os.Open(...)`), once the err check passes.
- Avoid `defer` in tight loops — each defer has an allocation cost, and the calls only run when the enclosing function returns.
- Ref: Effective Go §Defer.

## Constants and `iota`

- `const` is compile-time; `var` is run-time. Prefer `const` whenever a value is known at compile time.
- `iota` generates sequential typed constants inside a `const (...)` block, resetting per block:
  ```go
  const (
      StatusDraft Status = iota // 0
      StatusActive              // 1
      StatusArchived            // 2
  )
  ```
- For bitmasks: `1 << iota`.
- Ref: Go Spec §Iota, Effective Go §Constants.

## Concurrency

- A goroutine without a clear exit path is a leak. Every `go func()` should have: a `context.Context`, a `done` channel, a bounded workload, or a `sync.WaitGroup`.
- Propagate `context.Context` as the **first** parameter (`ctx context.Context, ...`). Never store it in a struct.
- `sync.Mutex`: zero-value ready. Embed or field it, never pointer to it. Name the mutex for what it protects (`mu` next to the field, or a comment).
- `sync.Once` for one-shot initialization (also `sync.OnceFunc`, Go 1.21+).
- Don't communicate by sharing memory — share memory by communicating (channels). But a mutex is correct when the shared state is a single variable.
- Always cancel a `context.WithCancel` / `WithTimeout` — `defer cancel()` on the next line.
- Ref: https://go.dev/blog/pipelines, https://go.dev/ref/mem.

## `any`, Generics, `comparable`

- `any` is the preferred alias for `interface{}` (Go 1.18+).
- Use generics when the alternative is code duplication across types with identical logic (e.g., slice utilities). Don't reach for generics to "future-proof" a single-type API.
- Type parameter naming: single uppercase letter (`T`, `K`, `V`) or short PascalCase (`Elem`).
- Ref: https://go.dev/blog/intro-generics.

## Tests (Language Level)

(Project-specific test conventions — table-driven, mock hygiene, TC-IDs — live in
[`go-conventions.md`](go-conventions.md) and [`.claude/agents/test-reviewer.md`](../agents/test-reviewer.md).)

- Test files end in `_test.go`; test functions match `TestXxx(t *testing.T)`.
- `t.Helper()` inside assertion helpers so failures point at the caller.
- `t.Parallel()` for independent tests. Unsafe when tests share mutable global state (env vars, working directory).
- `t.Cleanup(fn)` over `defer` for test teardown — runs even when the test fails mid-flight.
- Example functions (`ExampleXxx`) double as documentation and are verified by `go test`.
- Ref: https://pkg.go.dev/testing, https://go.dev/blog/subtests.

## Packages and Imports

- One package per directory; the package name matches the directory (almost always).
- Circular imports are a compile error — the compiler is right; restructure.
- Dot imports (`import . "pkg"`) are banned outside test DSLs.
- Blank imports (`import _ "pkg"`) only for side effects (drivers, init registration). Comment *why* next to the line.
- Internal packages (`internal/...`) can only be imported from the module that contains them — use this to enforce boundaries.
- Ref: Effective Go §Package names, Go Spec §Import declarations.

## What NOT to Do

- Don't name returns unless it aids documentation or is required by `defer` to mutate a result. Naked returns in long functions are a readability loss.
- Don't write getters prefixed with `Get`.
- Don't wrap every error with `fmt.Errorf` — add context only when it's useful to the caller.
- Don't swallow errors silently (`_ = fn()`). Either handle or propagate.
- Don't use `panic` for control flow. `panic` is for truly unrecoverable invariant violations.
- Don't create "utility" packages named `util`, `common`, or `helpers` — name by what they *do*.
