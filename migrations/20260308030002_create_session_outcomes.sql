-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_outcomes (
    id              BIGSERIAL PRIMARY KEY,
    session_id      BIGINT NOT NULL UNIQUE REFERENCES sessions(id) ON DELETE CASCADE,
    outcome_status  VARCHAR(50) NOT NULL CHECK (outcome_status IN ('completed', 'partial', 'aborted')),
    aftercare_notes TEXT,
    red_flags_text  TEXT,
    contact_info    TEXT,
    follow_up_at    TIMESTAMP WITH TIME ZONE,
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id)
);

CREATE TABLE session_outcome_endpoints (
    outcome_id  BIGINT NOT NULL REFERENCES session_outcomes(id) ON DELETE CASCADE,
    endpoint_id BIGINT NOT NULL REFERENCES clinical_endpoints(id),
    PRIMARY KEY (outcome_id, endpoint_id)
);

CREATE INDEX idx_session_outcomes_session_id ON session_outcomes(session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_outcome_endpoints;
DROP TABLE IF EXISTS session_outcomes;
-- +goose StatementEnd
