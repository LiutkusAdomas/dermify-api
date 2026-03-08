-- +goose Up
-- +goose StatementBegin

-- Generic function for tables with direct session_id FK
CREATE OR REPLACE FUNCTION prevent_signed_session_modification()
RETURNS TRIGGER AS $$
DECLARE
    v_session_status VARCHAR(20);
BEGIN
    -- Handle the sessions table itself
    IF TG_TABLE_NAME = 'sessions' THEN
        IF OLD.status IN ('signed', 'locked') THEN
            -- Allow ONLY signed -> locked transition
            IF TG_OP = 'UPDATE' AND OLD.status = 'signed'
               AND NEW.status = 'locked' THEN
                RETURN NEW;
            END IF;
            RAISE EXCEPTION 'Cannot modify session in % state (session_id: %)', OLD.status, OLD.id;
        END IF;
        IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
        RETURN NEW;
    END IF;

    -- For child tables with session_id column
    SELECT status INTO v_session_status FROM sessions WHERE id = OLD.session_id;
    IF v_session_status IN ('signed', 'locked') THEN
        RAISE EXCEPTION 'Cannot modify % when session is in % state', TG_TABLE_NAME, v_session_status;
    END IF;

    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function for module detail tables (join through session_modules)
CREATE OR REPLACE FUNCTION prevent_signed_module_detail_modification()
RETURNS TRIGGER AS $$
DECLARE
    v_session_status VARCHAR(20);
BEGIN
    SELECT s.status INTO v_session_status
    FROM sessions s
    JOIN session_modules sm ON s.id = sm.session_id
    WHERE sm.id = OLD.module_id;

    IF v_session_status IN ('signed', 'locked') THEN
        RAISE EXCEPTION 'Cannot modify % when session is in % state', TG_TABLE_NAME, v_session_status;
    END IF;

    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Addendum immutability (always immutable, no exceptions)
CREATE OR REPLACE FUNCTION prevent_addendum_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Addendums cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

-- Apply to sessions table
CREATE TRIGGER enforce_session_immutability
    BEFORE UPDATE OR DELETE ON sessions
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

-- Apply to direct child tables
CREATE TRIGGER enforce_consent_immutability
    BEFORE UPDATE OR DELETE ON session_consents
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

CREATE TRIGGER enforce_screening_immutability
    BEFORE UPDATE OR DELETE ON contraindication_screenings
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

CREATE TRIGGER enforce_module_immutability
    BEFORE UPDATE OR DELETE ON session_modules
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

CREATE TRIGGER enforce_outcome_immutability
    BEFORE UPDATE OR DELETE ON session_outcomes
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

-- Apply to module detail tables (via session_modules join)
CREATE TRIGGER enforce_ipl_detail_immutability
    BEFORE UPDATE OR DELETE ON ipl_module_details
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_module_detail_modification();

CREATE TRIGGER enforce_ndyag_detail_immutability
    BEFORE UPDATE OR DELETE ON ndyag_module_details
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_module_detail_modification();

CREATE TRIGGER enforce_co2_detail_immutability
    BEFORE UPDATE OR DELETE ON co2_module_details
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_module_detail_modification();

CREATE TRIGGER enforce_rf_detail_immutability
    BEFORE UPDATE OR DELETE ON rf_module_details
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_module_detail_modification();

CREATE TRIGGER enforce_filler_detail_immutability
    BEFORE UPDATE OR DELETE ON filler_module_details
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_module_detail_modification();

CREATE TRIGGER enforce_botulinum_detail_immutability
    BEFORE UPDATE OR DELETE ON botulinum_module_details
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_module_detail_modification();

-- Apply to addendums (always immutable)
CREATE TRIGGER enforce_addendum_immutability
    BEFORE UPDATE OR DELETE ON session_addendums
    FOR EACH ROW EXECUTE FUNCTION prevent_addendum_modification();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop all immutability triggers
DROP TRIGGER IF EXISTS enforce_session_immutability ON sessions;
DROP TRIGGER IF EXISTS enforce_consent_immutability ON session_consents;
DROP TRIGGER IF EXISTS enforce_screening_immutability ON contraindication_screenings;
DROP TRIGGER IF EXISTS enforce_module_immutability ON session_modules;
DROP TRIGGER IF EXISTS enforce_outcome_immutability ON session_outcomes;
DROP TRIGGER IF EXISTS enforce_ipl_detail_immutability ON ipl_module_details;
DROP TRIGGER IF EXISTS enforce_ndyag_detail_immutability ON ndyag_module_details;
DROP TRIGGER IF EXISTS enforce_co2_detail_immutability ON co2_module_details;
DROP TRIGGER IF EXISTS enforce_rf_detail_immutability ON rf_module_details;
DROP TRIGGER IF EXISTS enforce_filler_detail_immutability ON filler_module_details;
DROP TRIGGER IF EXISTS enforce_botulinum_detail_immutability ON botulinum_module_details;
DROP TRIGGER IF EXISTS enforce_addendum_immutability ON session_addendums;

DROP FUNCTION IF EXISTS prevent_signed_session_modification();
DROP FUNCTION IF EXISTS prevent_signed_module_detail_modification();
DROP FUNCTION IF EXISTS prevent_addendum_modification();
-- +goose StatementEnd
