-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    id                 BIGSERIAL PRIMARY KEY,
    patient_id         BIGINT NOT NULL REFERENCES patients(id),
    clinician_id       BIGINT NOT NULL REFERENCES users(id),
    status             VARCHAR(20) NOT NULL DEFAULT 'draft'
                       CHECK (status IN ('draft', 'in_progress', 'awaiting_signoff', 'signed', 'locked')),
    scheduled_at       TIMESTAMP WITH TIME ZONE,
    started_at         TIMESTAMP WITH TIME ZONE,
    completed_at       TIMESTAMP WITH TIME ZONE,
    patient_goal       TEXT,
    fitzpatrick_type   INTEGER CHECK (fitzpatrick_type IS NULL OR (fitzpatrick_type >= 1 AND fitzpatrick_type <= 6)),
    is_tanned          BOOLEAN NOT NULL DEFAULT false,
    is_pregnant        BOOLEAN NOT NULL DEFAULT false,
    on_anticoagulants  BOOLEAN NOT NULL DEFAULT false,
    photo_consent      VARCHAR(10) CHECK (photo_consent IS NULL OR photo_consent IN ('yes', 'no', 'limited')),
    notes              TEXT,
    version            INTEGER NOT NULL DEFAULT 1,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by         BIGINT NOT NULL REFERENCES users(id),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by         BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_sessions_patient_id ON sessions(patient_id);
CREATE INDEX idx_sessions_clinician_id ON sessions(clinician_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_scheduled_at ON sessions(scheduled_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
