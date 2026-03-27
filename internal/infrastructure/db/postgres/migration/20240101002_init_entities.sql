-- +goose Up
CREATE TABLE entities (
    id VARCHAR(26) PRIMARY KEY CHECK (char_length(id) = 26), -- ULID: always exactly 26 chars
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_entities_active_created ON entities(created_at DESC) WHERE active = true;

-- For name search performance at scale, consider:
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- CREATE INDEX idx_entities_name_trgm ON entities USING gin(name gin_trgm_ops);

-- +goose Down
DROP TABLE IF EXISTS entities;
