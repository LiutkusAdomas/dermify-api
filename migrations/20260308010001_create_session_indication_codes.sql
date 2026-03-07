-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_indication_codes (
    session_id         BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    indication_code_id BIGINT NOT NULL REFERENCES indication_codes(id),
    PRIMARY KEY (session_id, indication_code_id)
);

CREATE INDEX idx_session_indication_codes_session ON session_indication_codes(session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_indication_codes;
-- +goose StatementEnd
