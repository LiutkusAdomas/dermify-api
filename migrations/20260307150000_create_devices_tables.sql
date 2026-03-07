-- +goose Up
-- +goose StatementBegin
CREATE TABLE devices (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(200) NOT NULL,
    manufacturer VARCHAR(200) NOT NULL,
    model        VARCHAR(200) NOT NULL,
    device_type  VARCHAR(50) NOT NULL CHECK (device_type IN ('ipl', 'ndyag', 'co2', 'rf')),
    active       BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE handpieces (
    id         BIGSERIAL PRIMARY KEY,
    device_id  BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name       VARCHAR(200) NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_handpieces_device_id ON handpieces(device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS handpieces;
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd
