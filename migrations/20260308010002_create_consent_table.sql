-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_consents (
    id              BIGSERIAL PRIMARY KEY,
    session_id      BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    consent_type    VARCHAR(50) NOT NULL,
    consent_method  VARCHAR(50) NOT NULL,
    obtained_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    risks_discussed BOOLEAN NOT NULL DEFAULT false,
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id),
    UNIQUE (session_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_consents;
-- +goose StatementEnd
