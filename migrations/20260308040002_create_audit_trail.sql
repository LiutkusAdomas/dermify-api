-- +goose Up
-- +goose StatementBegin
CREATE TABLE audit_trail (
    id           BIGSERIAL PRIMARY KEY,
    action       VARCHAR(10) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id      BIGINT,
    entity_type  VARCHAR(100) NOT NULL,
    entity_id    BIGINT NOT NULL,
    old_values   JSONB,
    new_values   JSONB
);

CREATE INDEX idx_audit_trail_entity ON audit_trail(entity_type, entity_id);
CREATE INDEX idx_audit_trail_performed_at ON audit_trail(performed_at);
CREATE INDEX idx_audit_trail_user_id ON audit_trail(user_id);

-- Make audit trail append-only
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit trail entries cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_trail_immutable
    BEFORE UPDATE OR DELETE ON audit_trail
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audit_trail_immutable ON audit_trail;
DROP FUNCTION IF EXISTS prevent_audit_modification();
DROP TABLE IF EXISTS audit_trail;
-- +goose StatementEnd
