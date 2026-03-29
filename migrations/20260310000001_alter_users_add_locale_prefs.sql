-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
  ADD COLUMN language VARCHAR(10) NOT NULL DEFAULT 'en',
  ADD COLUMN timezone VARCHAR(50) NOT NULL DEFAULT 'UTC';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
  DROP COLUMN language,
  DROP COLUMN timezone;
-- +goose StatementEnd
