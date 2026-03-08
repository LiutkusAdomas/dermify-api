-- +goose Up
-- +goose StatementBegin
CREATE TABLE ndyag_module_details (
    id              BIGSERIAL PRIMARY KEY,
    module_id       BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    device_id       BIGINT NOT NULL REFERENCES devices(id),
    handpiece_id    BIGINT REFERENCES handpieces(id),
    wavelength      VARCHAR(100),
    spot_size       VARCHAR(100),
    fluence         DECIMAL(6,2),
    pulse_duration  DECIMAL(8,2),
    repetition_rate DECIMAL(8,2),
    cooling_type    VARCHAR(100),
    total_pulses    INTEGER,
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_ndyag_module_details_module_id ON ndyag_module_details(module_id);
CREATE INDEX idx_ndyag_module_details_device_id ON ndyag_module_details(device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ndyag_module_details;
-- +goose StatementEnd
