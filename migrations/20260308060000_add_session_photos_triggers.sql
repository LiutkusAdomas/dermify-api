-- +goose Up
-- +goose StatementBegin

-- Immutability trigger: block UPDATE/DELETE on session_photos when parent session is signed/locked
CREATE TRIGGER enforce_photo_immutability
    BEFORE UPDATE OR DELETE ON session_photos
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

-- Audit trigger: record all INSERT/UPDATE/DELETE operations in audit_trail
CREATE TRIGGER audit_session_photos
    AFTER INSERT OR UPDATE OR DELETE ON session_photos
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS enforce_photo_immutability ON session_photos;
DROP TRIGGER IF EXISTS audit_session_photos ON session_photos;
-- +goose StatementEnd
