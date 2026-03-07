-- +goose Up
-- +goose StatementBegin
CREATE TABLE patients (
    id                BIGSERIAL PRIMARY KEY,
    first_name        VARCHAR(100) NOT NULL,
    last_name         VARCHAR(100) NOT NULL,
    date_of_birth     DATE NOT NULL,
    sex               VARCHAR(10) NOT NULL CHECK (sex IN ('male', 'female', 'other')),
    phone             VARCHAR(50),
    email             VARCHAR(255),
    external_reference TEXT,
    version           INTEGER NOT NULL DEFAULT 1,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by        BIGINT NOT NULL REFERENCES users(id),
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by        BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_patients_last_name_lower ON patients (LOWER(last_name) varchar_pattern_ops);
CREATE INDEX idx_patients_first_name_lower ON patients (LOWER(first_name) varchar_pattern_ops);
CREATE INDEX idx_patients_email_lower ON patients (LOWER(email));
CREATE INDEX idx_patients_phone ON patients (phone);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS patients;
-- +goose StatementEnd
