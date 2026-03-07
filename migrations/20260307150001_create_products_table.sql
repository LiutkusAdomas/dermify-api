-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(200) NOT NULL,
    manufacturer  VARCHAR(200) NOT NULL,
    product_type  VARCHAR(50) NOT NULL CHECK (product_type IN ('filler', 'botulinum_toxin')),
    concentration VARCHAR(100),
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
