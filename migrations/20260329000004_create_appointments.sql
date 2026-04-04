-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS appointments (
    id                  BIGSERIAL PRIMARY KEY,
    org_id              BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    patient_id          BIGINT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_type_id     BIGINT NOT NULL REFERENCES service_types(id),
    start_time          TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time            TIMESTAMP WITH TIME ZONE NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'scheduled'
                        CHECK (status IN ('scheduled','confirmed','checked_in','in_progress','completed','cancelled','no_show')),
    notes               TEXT NOT NULL DEFAULT '',
    cancellation_reason TEXT NOT NULL DEFAULT '',
    session_id          BIGINT REFERENCES sessions(id),
    created_by          BIGINT NOT NULL REFERENCES users(id),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    version             INTEGER NOT NULL DEFAULT 1,
    CHECK (start_time < end_time)
);

CREATE INDEX idx_appointments_org_doctor_time ON appointments(org_id, doctor_id, start_time);
CREATE INDEX idx_appointments_org_patient ON appointments(org_id, patient_id);
CREATE INDEX idx_appointments_org_status ON appointments(org_id, status);
CREATE INDEX idx_appointments_start_time ON appointments(start_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appointments;
-- +goose StatementEnd
