-- +goose Up
CREATE TABLE people (
    id CHAR(26) PRIMARY KEY, -- ULID cabe em 26 chars
    name VARCHAR(255) NOT NULL,
    document VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(11) NOT NULL,
    
    -- Endereço
    street VARCHAR(255),           -- Logradouro (Rua, Av, etc)
    number VARCHAR(20),            -- Número
    complement VARCHAR(100),       -- Complemento (Apto, Bloco, etc)
    neighborhood VARCHAR(100),     -- Bairro
    city VARCHAR(100),             -- Cidade
    state CHAR(2),                 -- UF (SP, RJ, MG, etc)
    zip_code VARCHAR(10),          -- CEP (formato: 00000-000 ou 00000000)
    
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_people_email ON people(email);
CREATE INDEX idx_people_city_state ON people(city, state);

-- +goose Down
DROP TABLE people;
