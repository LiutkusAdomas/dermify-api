-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN signed_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sessions ADD COLUMN signed_by BIGINT REFERENCES users(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN IF EXISTS signed_by;
ALTER TABLE sessions DROP COLUMN IF EXISTS signed_at;
-- +goose StatementEnd
