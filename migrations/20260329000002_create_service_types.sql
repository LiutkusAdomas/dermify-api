-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS service_types (
    id                       BIGSERIAL PRIMARY KEY,
    org_id                   BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                     VARCHAR(255) NOT NULL,
    default_duration_minutes INTEGER NOT NULL DEFAULT 30,
    description              TEXT NOT NULL DEFAULT '',
    active                   BOOLEAN NOT NULL DEFAULT true,
    created_at               TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (org_id, name)
);

CREATE INDEX idx_service_types_org_id ON service_types(org_id);
CREATE INDEX idx_service_types_active ON service_types(org_id, active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS service_types;
-- +goose StatementEnd
