-- +goose Up
-- +goose StatementBegin

-- Generic audit trigger function
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id BIGINT;
    v_entity_id BIGINT;
    v_old JSONB;
    v_new JSONB;
BEGIN
    -- Extract entity ID
    IF TG_OP = 'DELETE' THEN
        v_entity_id := OLD.id;
    ELSE
        v_entity_id := NEW.id;
    END IF;

    -- Extract user ID from row data
    IF TG_OP = 'INSERT' THEN
        v_user_id := NEW.created_by;
        v_old := NULL;
        v_new := row_to_json(NEW)::JSONB;
    ELSIF TG_OP = 'UPDATE' THEN
        v_user_id := NEW.updated_by;
        v_old := row_to_json(OLD)::JSONB;
        v_new := row_to_json(NEW)::JSONB;
    ELSIF TG_OP = 'DELETE' THEN
        v_user_id := OLD.updated_by;
        v_old := row_to_json(OLD)::JSONB;
        v_new := NULL;
    END IF;

    INSERT INTO audit_trail (action, performed_at, user_id, entity_type, entity_id, old_values, new_values)
    VALUES (TG_OP, CURRENT_TIMESTAMP, v_user_id, TG_TABLE_NAME, v_entity_id, v_old, v_new);

    RETURN NULL;  -- AFTER trigger, return value is ignored
END;
$$ LANGUAGE plpgsql;

-- Attach to all clinical tables
CREATE TRIGGER audit_sessions
    AFTER INSERT OR UPDATE OR DELETE ON sessions
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_session_consents
    AFTER INSERT OR UPDATE OR DELETE ON session_consents
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_contraindication_screenings
    AFTER INSERT OR UPDATE OR DELETE ON contraindication_screenings
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_session_modules
    AFTER INSERT OR UPDATE OR DELETE ON session_modules
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_session_outcomes
    AFTER INSERT OR UPDATE OR DELETE ON session_outcomes
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_ipl_module_details
    AFTER INSERT OR UPDATE OR DELETE ON ipl_module_details
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_ndyag_module_details
    AFTER INSERT OR UPDATE OR DELETE ON ndyag_module_details
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_co2_module_details
    AFTER INSERT OR UPDATE OR DELETE ON co2_module_details
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_rf_module_details
    AFTER INSERT OR UPDATE OR DELETE ON rf_module_details
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_filler_module_details
    AFTER INSERT OR UPDATE OR DELETE ON filler_module_details
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_botulinum_module_details
    AFTER INSERT OR UPDATE OR DELETE ON botulinum_module_details
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Addendums: AFTER INSERT ONLY (immutable, no UPDATE/DELETE audit needed)
CREATE TRIGGER audit_session_addendums
    AFTER INSERT ON session_addendums
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audit_sessions ON sessions;
DROP TRIGGER IF EXISTS audit_session_consents ON session_consents;
DROP TRIGGER IF EXISTS audit_contraindication_screenings ON contraindication_screenings;
DROP TRIGGER IF EXISTS audit_session_modules ON session_modules;
DROP TRIGGER IF EXISTS audit_session_outcomes ON session_outcomes;
DROP TRIGGER IF EXISTS audit_ipl_module_details ON ipl_module_details;
DROP TRIGGER IF EXISTS audit_ndyag_module_details ON ndyag_module_details;
DROP TRIGGER IF EXISTS audit_co2_module_details ON co2_module_details;
DROP TRIGGER IF EXISTS audit_rf_module_details ON rf_module_details;
DROP TRIGGER IF EXISTS audit_filler_module_details ON filler_module_details;
DROP TRIGGER IF EXISTS audit_botulinum_module_details ON botulinum_module_details;
DROP TRIGGER IF EXISTS audit_session_addendums ON session_addendums;

DROP FUNCTION IF EXISTS audit_trigger_function();
-- +goose StatementEnd
