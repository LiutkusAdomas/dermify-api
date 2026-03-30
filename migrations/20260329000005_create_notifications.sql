-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notifications (
    id           BIGSERIAL PRIMARY KEY,
    org_id       BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    patient_id   BIGINT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    channel      VARCHAR(20) NOT NULL CHECK (channel IN ('sms', 'email')),
    type         VARCHAR(50) NOT NULL,
    recipient    VARCHAR(255) NOT NULL,
    subject      VARCHAR(500) NOT NULL DEFAULT '',
    body         TEXT NOT NULL DEFAULT '',
    status       VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    sent_at      TIMESTAMP WITH TIME ZONE,
    error        TEXT NOT NULL DEFAULT '',
    reference_id BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_org ON notifications(org_id);
CREATE INDEX idx_notifications_patient ON notifications(patient_id);
CREATE INDEX idx_notifications_reference ON notifications(reference_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
