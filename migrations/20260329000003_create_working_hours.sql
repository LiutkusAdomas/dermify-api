-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS working_hours (
    id         BIGSERIAL PRIMARY KEY,
    org_id     BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    doctor_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time   TIME NOT NULL,
    UNIQUE (org_id, doctor_id, day_of_week),
    CHECK (start_time < end_time)
);

CREATE INDEX idx_working_hours_doctor ON working_hours(org_id, doctor_id);

CREATE TABLE IF NOT EXISTS schedule_overrides (
    id         BIGSERIAL PRIMARY KEY,
    org_id     BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    doctor_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date       DATE NOT NULL,
    start_time TIME,
    end_time   TIME,
    reason     VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (org_id, doctor_id, date),
    CHECK (
        (start_time IS NULL AND end_time IS NULL) OR
        (start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    )
);

CREATE INDEX idx_schedule_overrides_doctor_date ON schedule_overrides(org_id, doctor_id, date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schedule_overrides;
DROP TABLE IF EXISTS working_hours;
-- +goose StatementEnd
