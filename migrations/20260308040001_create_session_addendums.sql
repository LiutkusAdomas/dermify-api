-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_addendums (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id),
    author_id   BIGINT NOT NULL REFERENCES users(id),
    reason      TEXT NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_session_addendums_session_id ON session_addendums(session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_addendums;
-- +goose StatementEnd
