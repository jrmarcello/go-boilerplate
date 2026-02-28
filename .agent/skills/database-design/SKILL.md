---
name: database-design
description: PostgreSQL design patterns — sqlx, Goose migrations, DBCluster writer/reader, repository pattern, indexing
---

# Database Design

## Repository Pattern

Interface defined in use cases, implemented in infrastructure:

```go
// internal/usecases/entity_example/interfaces/repository.go
type Repository interface {
    FindByID(ctx context.Context, id string) (*entity.EntityExample, error)
    Save(ctx context.Context, e *entity.EntityExample) error
    Update(ctx context.Context, e *entity.EntityExample) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter entity.Filter) ([]entity.EntityExample, int, error)
}
```

Implementation uses sqlx with DBCluster:

```go
func (r *repository) FindByID(ctx context.Context, id string) (*entity.EntityExample, error) {
    var e entity.EntityExample
    readErr := r.cluster.Reader().GetContext(ctx, &e,
        `SELECT id, name, email, created_at, updated_at FROM entities WHERE id = $1`, id)
    if readErr != nil {
        return nil, readErr
    }
    return &e, nil
}
```

## Writer/Reader Split

Via `pkg/database.DBCluster`:

```go
// Writer: INSERT, UPDATE, DELETE
r.cluster.Writer().ExecContext(ctx, `INSERT INTO entities ...`)

// Reader: SELECT (fallback to writer if no reader)
r.cluster.Reader().GetContext(ctx, &result, `SELECT ...`)
```

Config: `DB_READER_DSN` env var (optional).

## Goose Migrations

```bash
make migrate-create NAME=add_column
make migrate-up
```

Format:

```sql
-- +goose Up
ALTER TABLE entities ADD COLUMN phone VARCHAR(20);
CREATE INDEX CONCURRENTLY idx_entities_phone ON entities(phone);

-- +goose Down
DROP INDEX IF EXISTS idx_entities_phone;
ALTER TABLE entities DROP COLUMN IF EXISTS phone;
```

Rules:

- Always have `-- +goose Down`
- Add columns as NULL first, populate, then NOT NULL
- Create indexes with CONCURRENTLY
- Never alter applied production migrations

## IDs — ULID

All entity IDs use ULID (Universally Unique Lexicographically Sortable Identifier). See `docs/adr/002-ulid.md`.

```sql
CREATE TABLE entities (
    id    CHAR(26) PRIMARY KEY,  -- ULID
    name  VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE
);
```

## Connection Pool

Configure via environment:

- `DB_MAX_OPEN_CONNS` (default: 25)
- `DB_MAX_IDLE_CONNS` (default: 5)
- `DB_CONN_MAX_LIFETIME` (default: 5m)

Monitor via `pkg/telemetry.RegisterDBPoolMetrics()`.

## Query Optimization

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id, name FROM entities WHERE email = 'test@example.com';
```
