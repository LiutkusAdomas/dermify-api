-- +goose Up
-- +goose StatementBegin
CREATE TABLE rf_module_details (
    id              BIGSERIAL PRIMARY KEY,
    module_id       BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    device_id       BIGINT NOT NULL REFERENCES devices(id),
    handpiece_id    BIGINT REFERENCES handpieces(id),
    rf_mode         VARCHAR(100),
    tip_type        VARCHAR(100),
    depth           DECIMAL(6,2),
    energy_level    DECIMAL(6,2),
    overlap         DECIMAL(5,2),
    pulses_per_zone INTEGER,
    total_pulses    INTEGER,
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_rf_module_details_module_id ON rf_module_details(module_id);
CREATE INDEX idx_rf_module_details_device_id ON rf_module_details(device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rf_module_details;
-- +goose StatementEnd
