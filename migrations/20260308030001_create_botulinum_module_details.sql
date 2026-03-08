-- +goose Up
-- +goose StatementBegin
CREATE TABLE botulinum_module_details (
    id                       BIGSERIAL PRIMARY KEY,
    module_id                BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    product_id               BIGINT NOT NULL REFERENCES products(id),
    batch_number             VARCHAR(100),
    expiry_date              DATE,
    diluent                  VARCHAR(100),
    dilution_volume          DECIMAL(6,2),
    resulting_concentration  VARCHAR(100),
    total_units              DECIMAL(8,2),
    injection_sites          JSONB,
    notes                    TEXT,
    version                  INTEGER NOT NULL DEFAULT 1,
    created_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by               BIGINT NOT NULL REFERENCES users(id),
    updated_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by               BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_botulinum_module_details_module_id ON botulinum_module_details(module_id);
CREATE INDEX idx_botulinum_module_details_product_id ON botulinum_module_details(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS botulinum_module_details;
-- +goose StatementEnd
