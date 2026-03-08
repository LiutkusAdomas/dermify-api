-- +goose Up
-- +goose StatementBegin
CREATE TABLE co2_module_details (
    id               BIGSERIAL PRIMARY KEY,
    module_id        BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    device_id        BIGINT NOT NULL REFERENCES devices(id),
    handpiece_id     BIGINT REFERENCES handpieces(id),
    mode             VARCHAR(100),
    scanner_pattern  VARCHAR(100),
    power            DECIMAL(6,2),
    pulse_energy     DECIMAL(6,2),
    pulse_duration   DECIMAL(8,2),
    density          DECIMAL(5,2),
    pattern          VARCHAR(100),
    passes           INTEGER,
    anaesthesia_used VARCHAR(200),
    notes            TEXT,
    version          INTEGER NOT NULL DEFAULT 1,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by       BIGINT NOT NULL REFERENCES users(id),
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by       BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_co2_module_details_module_id ON co2_module_details(module_id);
CREATE INDEX idx_co2_module_details_device_id ON co2_module_details(device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS co2_module_details;
-- +goose StatementEnd
