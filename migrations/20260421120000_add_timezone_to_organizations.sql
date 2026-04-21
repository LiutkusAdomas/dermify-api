-- +goose Up
-- +goose StatementBegin
ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS timezone VARCHAR(64) NOT NULL DEFAULT 'UTC';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE organizations
DROP COLUMN IF EXISTS timezone;
-- +goose StatementEnd
