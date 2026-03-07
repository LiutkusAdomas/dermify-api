-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_modules (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    module_type VARCHAR(50) NOT NULL
                CHECK (module_type IN ('ipl', 'ndyag', 'co2', 'rf', 'filler', 'botulinum_toxin')),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by  BIGINT NOT NULL REFERENCES users(id),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by  BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_session_modules_session_id ON session_modules(session_id);
CREATE INDEX idx_session_modules_type ON session_modules(module_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_modules;
-- +goose StatementEnd
