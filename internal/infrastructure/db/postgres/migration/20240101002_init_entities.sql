-- +goose Up
CREATE TABLE entities (
    id CHAR(26) PRIMARY KEY, -- ULID cabe em 26 chars
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_entities_active_created ON entities(created_at DESC) WHERE active = true;

-- +goose Down
DROP TABLE IF EXISTS entities;
