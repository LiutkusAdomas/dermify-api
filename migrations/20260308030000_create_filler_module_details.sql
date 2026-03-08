-- +goose Up
-- +goose StatementBegin
CREATE TABLE filler_module_details (
    id               BIGSERIAL PRIMARY KEY,
    module_id        BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    product_id       BIGINT NOT NULL REFERENCES products(id),
    batch_number     VARCHAR(100),
    expiry_date      DATE,
    syringe_volume   DECIMAL(6,2),
    total_volume     DECIMAL(6,2),
    needle_type      VARCHAR(100),
    injection_plane  VARCHAR(100),
    anatomical_sites TEXT,
    endpoint         VARCHAR(200),
    notes            TEXT,
    version          INTEGER NOT NULL DEFAULT 1,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by       BIGINT NOT NULL REFERENCES users(id),
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by       BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_filler_module_details_module_id ON filler_module_details(module_id);
CREATE INDEX idx_filler_module_details_product_id ON filler_module_details(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS filler_module_details;
-- +goose StatementEnd
