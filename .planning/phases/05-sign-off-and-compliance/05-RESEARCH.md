# Phase 5: Sign-off and Compliance - Research

**Researched:** 2026-03-08
**Domain:** Session sign-off, immutability enforcement (DB triggers), addendums, audit trail
**Confidence:** HIGH

## Summary

Phase 5 delivers the core medico-legal requirement: a clinician signs off a completed session, producing a locked, immutable medical record with a full audit trail. This phase builds on the existing session lifecycle (Draft -> In Progress -> Awaiting Sign-off -> Signed -> Locked) already implemented in Phases 2-4, and adds three major capabilities: (1) pre-sign-off validation that checks completeness across all session components, (2) database-level immutability enforcement via PostgreSQL triggers that prevent modification of signed/locked records, and (3) an append-only audit trail that records every create, update, sign-off, and lock operation on clinical entities.

The codebase already has the session status constants (`signed`, `locked`) and valid state transitions defined in `service/session.go`. The Session domain type already tracks `created_by`, `updated_by`, `created_at`, `updated_at`, and `version`. What is missing is: the `signed_at`/`signed_by` columns (META-02), the validation logic that gates the AwaitingSignoff->Signed transition, the addendum domain/table, the audit_trail table and triggers, and the immutability triggers. No new external dependencies are needed -- all work uses stdlib `database/sql`, existing pgx/goose stack, and native PostgreSQL PL/pgSQL trigger functions.

**Primary recommendation:** Implement in three plans: (1) database layer (migrations for signed_at/signed_by, addendums table, audit_trail table, immutability triggers, audit triggers), (2) service and repository layer (sign-off validation service, addendum service/repo, audit repo), (3) HTTP handlers and route wiring.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| LOCK-01 | System validates all required fields before allowing sign-off (blocks if incomplete) | Sign-off validation service checks consent, screening, modules, outcome exist; returns structured list of missing items |
| LOCK-02 | Clinician can sign off a session (records timestamp and clinician ID) | Add signed_at, signed_by columns to sessions table; transition AwaitingSignoff->Signed sets these fields |
| LOCK-03 | Signed session becomes immutable -- original record cannot be modified | BEFORE UPDATE/DELETE triggers on sessions, session_consents, session_modules, all detail tables, session_outcomes raise exception when session status is 'signed' or 'locked' |
| LOCK-04 | Clinician can add addendums to a locked session (date, author, reason, content) | New session_addendums table with FK to sessions, new domain type, service, repository, handler |
| LOCK-05 | Addendums are themselves immutable once saved | BEFORE UPDATE/DELETE trigger on session_addendums raises exception unconditionally |
| LOCK-06 | Immutability is enforced at the database level (not just application layer) | PL/pgSQL trigger functions using RAISE EXCEPTION; cannot be bypassed by direct SQL |
| AUDIT-01 | System logs all create, update, and delete operations on clinical entities | Generic audit trigger function attached to all clinical tables; fires AFTER INSERT/UPDATE/DELETE |
| AUDIT-02 | Each audit entry captures: action, timestamp, user ID, entity type, entity ID | audit_trail table with action, performed_at, user_id, entity_type, entity_id, plus optional old/new JSONB |
| AUDIT-03 | Audit log is append-only -- entries cannot be modified or deleted | BEFORE UPDATE/DELETE trigger on audit_trail raises exception; also REVOKE UPDATE, DELETE on table |
| AUDIT-04 | Sign-off and lock events are recorded in the audit trail | Audit trigger captures status transitions; sign-off service also writes explicit audit entry |
| META-02 | Signed records track signed_at, signed_by | Add nullable signed_at (TIMESTAMPTZ), signed_by (BIGINT FK users) columns to sessions table |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| database/sql | stdlib | Database operations | Already used throughout; all repos use `*sql.DB` |
| PL/pgSQL | PostgreSQL built-in | Trigger functions | Native PostgreSQL procedural language for triggers; no extensions needed |
| pressly/goose v3 | v3 | SQL migrations | Already used for all schema changes; embedded SQL via `//go:embed` |
| stretchr/testify | v1 | Testing assertions | Already used in all service tests |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | JSONB serialization for audit old/new values | Audit trail stores before/after snapshots |
| go-chi/chi v5 | v5 | HTTP routing | New endpoints for sign-off, addendums, audit trail |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| PL/pgSQL triggers | Application-layer checks only | LOCK-06 explicitly requires DB-level enforcement; app-layer alone is insufficient |
| JSONB for audit old/new | hstore extension | JSONB is built-in, no extension install needed, richer querying |
| Separate audit service | In-process audit repo | Unnecessary complexity; audit writes are synchronous trigger-based at DB level |

**Installation:**
No new dependencies required. All work uses existing stack.

## Architecture Patterns

### Recommended Project Structure
```
migrations/
  20260308040000_add_session_signoff_columns.sql
  20260308040001_create_session_addendums.sql
  20260308040002_create_audit_trail.sql
  20260308040003_create_immutability_triggers.sql
  20260308040004_create_audit_triggers.sql
internal/
  domain/
    addendum.go                  # Addendum domain type
    audit.go                     # AuditEntry domain type
  service/
    signoff.go                   # SignoffService: validation + sign-off orchestration
    signoff_test.go              # Sign-off validation and transition tests
    addendum.go                  # AddendumService
    addendum_test.go             # Addendum service tests
    audit.go                     # AuditService (read-only query)
    audit_test.go                # Audit service tests
  repository/postgres/
    addendum.go                  # AddendumRepository postgres impl
    addendum_test.go             # Addendum repo tests
    audit.go                     # AuditRepository postgres impl (read-only)
    audit_test.go                # Audit repo tests
  testutil/
    mock_addendum.go             # Mock addendum repo
    mock_audit.go                # Mock audit repo
  api/
    handlers/
      signoff.go                 # HandleSignOffSession, HandleGetSignOffReadiness
      addendum.go                # HandleCreateAddendum, HandleListAddendums, HandleGetAddendum
      audit.go                   # HandleGetAuditTrail
      signoff_errors.go          # Error mapping for sign-off operations
    routes/
      sessions.go                # Extended with sign-off, addendum, audit routes
```

### Pattern 1: Sign-off Validation Service
**What:** A dedicated `SignoffService` that aggregates completeness checks across all session sub-entities and orchestrates the Signed transition.
**When to use:** When transitioning from AwaitingSignoff to Signed.
**Example:**
```go
// SignoffService validates session completeness and performs sign-off.
type SignoffService struct {
    sessionRepo  SessionRepository
    consentRepo  ConsentRepository
    moduleRepo   ModuleRepository
    outcomeRepo  OutcomeRepository
    auditRepo    AuditRepository
}

// ValidationResult holds the completeness check results.
type ValidationResult struct {
    Ready    bool     `json:"ready"`
    Missing  []string `json:"missing,omitempty"`
}

// ValidateForSignoff checks all required components exist.
func (s *SignoffService) ValidateForSignoff(ctx context.Context, sessionID int64) (*ValidationResult, error) {
    var missing []string
    // Check session is in awaiting_signoff state
    // Check consent exists
    // Check at least one module exists
    // Check outcome exists
    // Return structured result
    return &ValidationResult{Ready: len(missing) == 0, Missing: missing}, nil
}

// SignOff validates completeness, transitions to Signed, sets signed_at/signed_by.
func (s *SignoffService) SignOff(ctx context.Context, sessionID int64, clinicianID int64) error {
    result, err := s.ValidateForSignoff(ctx, sessionID)
    if err != nil { return err }
    if !result.Ready { return ErrSessionIncomplete }
    // Set signed_at, signed_by, transition to Signed status
    return nil
}
```

### Pattern 2: Database-Level Immutability via Triggers
**What:** PL/pgSQL BEFORE UPDATE/DELETE trigger functions that RAISE EXCEPTION when the session is in a protected state (signed or locked).
**When to use:** On all clinical entity tables that reference sessions.
**Example:**
```sql
-- Trigger function: prevent modification of signed/locked sessions
CREATE OR REPLACE FUNCTION prevent_signed_session_modification()
RETURNS TRIGGER AS $$
DECLARE
    session_status VARCHAR(20);
BEGIN
    -- For the sessions table itself, check OLD status
    IF TG_TABLE_NAME = 'sessions' THEN
        IF OLD.status IN ('signed', 'locked') THEN
            -- Allow only the transition from signed -> locked
            IF TG_OP = 'UPDATE' AND OLD.status = 'signed'
               AND NEW.status = 'locked' THEN
                RETURN NEW;
            END IF;
            RAISE EXCEPTION 'Cannot modify session in % state', OLD.status;
        END IF;
        RETURN NEW;
    END IF;

    -- For child tables, look up the session status
    SELECT status INTO session_status FROM sessions WHERE id = OLD.session_id;
    IF session_status IN ('signed', 'locked') THEN
        RAISE EXCEPTION 'Cannot modify % when session is %', TG_TABLE_NAME, session_status;
    END IF;

    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### Pattern 3: Append-Only Audit Trail via Triggers
**What:** A generic AFTER trigger function that writes to an `audit_trail` table on every INSERT/UPDATE/DELETE of clinical entities.
**When to use:** Attached to all clinical tables (sessions, session_consents, session_modules, detail tables, session_outcomes, session_addendums).
**Example:**
```sql
CREATE TABLE audit_trail (
    id           BIGSERIAL PRIMARY KEY,
    action       VARCHAR(10) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id      BIGINT,  -- extracted from updated_by/created_by
    entity_type  VARCHAR(100) NOT NULL,
    entity_id    BIGINT NOT NULL,
    old_values   JSONB,
    new_values   JSONB
);

-- Prevent any modification to audit entries
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit trail entries cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_trail_immutable
    BEFORE UPDATE OR DELETE ON audit_trail
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();
```

### Pattern 4: Addendum as Immutable Append
**What:** Addendums are separate records linked to a locked session. They are insert-only: once created, they cannot be modified.
**When to use:** When a clinician needs to correct or supplement a locked medical record.
**Example:**
```go
type Addendum struct {
    ID        int64     `json:"id"`
    SessionID int64     `json:"session_id"`
    AuthorID  int64     `json:"author_id"`
    Reason    string    `json:"reason"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Anti-Patterns to Avoid
- **Application-only immutability:** Never rely solely on Go code to prevent modifications. A developer with DB access could bypass application checks. Triggers enforce at the storage layer.
- **Mutable audit entries:** Never allow UPDATE or DELETE on the audit_trail table. Use both triggers AND SQL REVOKE to enforce.
- **Storing audit in the same row:** Don't add audit columns to clinical tables. Use a separate append-only table so history cannot be rewritten.
- **Complex trigger logic:** Keep trigger functions simple. The immutability trigger just checks status and raises exception. The audit trigger just captures old/new values and writes a row.
- **Signed-to-Locked gap:** Don't leave a period where a signed session can still be modified. The immutability trigger must protect BOTH signed and locked states.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Audit trail capture | Custom application-level audit logging in each handler | PostgreSQL AFTER trigger function | Triggers capture ALL changes including direct SQL; impossible to bypass; no code duplication across handlers |
| Immutability enforcement | Application-layer `if session.Status == "locked" { return error }` only | PostgreSQL BEFORE UPDATE/DELETE trigger with RAISE EXCEPTION | Requirement LOCK-06 explicitly demands DB-level enforcement; triggers protect against direct SQL access |
| JSON serialization of rows for audit | Custom column-by-column serialization | `row_to_json(OLD)` / `row_to_json(NEW)` in PL/pgSQL | Built-in PostgreSQL function, handles all column types, no maintenance when schema changes |

**Key insight:** The medico-legal requirement (LOCK-06) makes database-level enforcement non-negotiable. Application-layer checks are still needed for good UX (clear error messages before hitting DB triggers), but they are the second line of defense, not the primary one.

## Common Pitfalls

### Pitfall 1: Trigger Order and Circular Dependencies
**What goes wrong:** Immutability triggers fire before audit triggers, or audit triggers try to write to a table that has its own triggers causing unexpected behavior.
**Why it happens:** PostgreSQL fires triggers in alphabetical order within the same timing (BEFORE/AFTER). Immutability triggers must be BEFORE (to prevent the operation), audit triggers must be AFTER (to capture what happened).
**How to avoid:** Name immutability triggers with a prefix like `enforce_` and audit triggers with `audit_`. Ensure immutability triggers are BEFORE UPDATE/DELETE and audit triggers are AFTER INSERT/UPDATE/DELETE.
**Warning signs:** Audit entries appearing for operations that should have been blocked.

### Pitfall 2: Session Table Self-Reference in Trigger
**What goes wrong:** The immutability trigger on the sessions table itself needs special handling for the Signed->Locked transition. A naive trigger that blocks all updates on signed sessions would prevent the lock transition.
**Why it happens:** The sessions table is both the entity being protected AND the entity whose status determines protection.
**How to avoid:** The trigger function must explicitly allow the Signed->Locked transition on the sessions table while blocking all other modifications to signed/locked sessions.
**Warning signs:** Unable to transition from Signed to Locked after immutability triggers are installed.

### Pitfall 3: Child Tables Without session_id
**What goes wrong:** Module detail tables (ipl_module_details, etc.) reference session_modules.id, not sessions.id directly. The immutability trigger needs to join through session_modules to find the session status.
**Why it happens:** The polymorphic module design uses a two-table pattern (session_modules -> *_module_details).
**How to avoid:** For detail tables, the trigger function must SELECT sessions.status FROM sessions JOIN session_modules ON sessions.id = session_modules.session_id WHERE session_modules.id = OLD.module_id.
**Warning signs:** Detail tables remain editable even when their parent session is signed/locked.

### Pitfall 4: Audit Trail user_id Extraction
**What goes wrong:** PostgreSQL triggers don't have access to the application-level user ID by default. The trigger must extract it from the row's created_by/updated_by columns.
**Why it happens:** Unlike application code that has the JWT claims context, triggers only see table data and session variables.
**How to avoid:** Use `NEW.created_by` for INSERT, `NEW.updated_by` for UPDATE, `OLD.updated_by` for DELETE. All clinical tables already have these columns.
**Warning signs:** Audit entries with NULL user_id.

### Pitfall 5: Addendum Allowed Only on Locked Sessions
**What goes wrong:** Allowing addendums on signed (not yet locked) sessions, or on draft/in_progress sessions.
**Why it happens:** The requirement says "add addendums to a locked session" specifically.
**How to avoid:** The addendum service must check session.Status == "locked" before allowing creation. The DB can also enforce this via a trigger or check constraint.
**Warning signs:** Addendums appearing on sessions that haven't been locked.

### Pitfall 6: Missing Signed_at/Signed_by on Lock
**What goes wrong:** The session transitions to Signed but signed_at and signed_by are not populated.
**Why it happens:** The existing `UpdateStatus` repository method only sets status, version, updated_at, updated_by. It doesn't set the new signed_at/signed_by columns.
**How to avoid:** Either extend `UpdateStatus` to accept optional signed fields, or create a dedicated `SignOff` repository method that sets status, signed_at, signed_by atomically.
**Warning signs:** signed_at is NULL on signed sessions.

## Code Examples

### Migration: Add signed_at/signed_by to sessions
```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN signed_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sessions ADD COLUMN signed_by BIGINT REFERENCES users(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN IF EXISTS signed_by;
ALTER TABLE sessions DROP COLUMN IF EXISTS signed_at;
-- +goose StatementEnd
```

### Migration: Create session_addendums table
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_addendums (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id),
    author_id   BIGINT NOT NULL REFERENCES users(id),
    reason      TEXT NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_session_addendums_session_id ON session_addendums(session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_addendums;
-- +goose StatementEnd
```

### Migration: Create audit_trail table
```sql
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
```

### Migration: Immutability triggers
```sql
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
```

### Migration: Audit triggers
```sql
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
```

### Domain: Addendum Type
```go
package domain

import "time"

// Addendum represents an immutable addendum to a locked session.
type Addendum struct {
    ID        int64     `json:"id"`
    SessionID int64     `json:"session_id"`
    AuthorID  int64     `json:"author_id"`
    Reason    string    `json:"reason"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Domain: AuditEntry Type
```go
package domain

import (
    "encoding/json"
    "time"
)

// AuditEntry represents a single entry in the append-only audit trail.
type AuditEntry struct {
    ID          int64            `json:"id"`
    Action      string           `json:"action"`
    PerformedAt time.Time        `json:"performed_at"`
    UserID      *int64           `json:"user_id,omitempty"`
    EntityType  string           `json:"entity_type"`
    EntityID    int64            `json:"entity_id"`
    OldValues   json.RawMessage  `json:"old_values,omitempty"`
    NewValues   json.RawMessage  `json:"new_values,omitempty"`
}
```

### Service: Sign-off Validation
```go
// ValidateForSignoff checks that all required session components exist.
func (s *SignoffService) ValidateForSignoff(ctx context.Context, sessionID int64) (*ValidationResult, error) {
    session, err := s.sessionRepo.GetByID(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    var missing []string

    if session.Status != domain.SessionStatusAwaitingSignoff {
        return &ValidationResult{Ready: false, Missing: []string{"session must be in awaiting_signoff state"}}, nil
    }

    // Check consent exists
    hasConsent, err := s.consentRepo.ExistsForSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    if !hasConsent {
        missing = append(missing, "consent record")
    }

    // Check at least one module exists
    modules, err := s.moduleRepo.ListBySession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    if len(modules) == 0 {
        missing = append(missing, "at least one procedure module")
    }

    // Check outcome exists
    hasOutcome, err := s.outcomeRepo.ExistsForSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    if !hasOutcome {
        missing = append(missing, "outcome record")
    }

    return &ValidationResult{Ready: len(missing) == 0, Missing: missing}, nil
}
```

### Handler: Sign-off Endpoint
```go
// HandleSignOffSession signs off a session after validation.
func HandleSignOffSession(svc *service.SignoffService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        claims := middleware.GetUserClaims(r.Context())
        if claims == nil {
            apierrors.WriteError(w, http.StatusUnauthorized,
                apierrors.AuthNotAuthenticated, "not authenticated")
            return
        }

        id, err := parseIDParam(r)
        if err != nil {
            apierrors.WriteError(w, http.StatusBadRequest,
                apierrors.SessionInvalidData, "invalid session ID")
            return
        }

        if err := svc.SignOff(r.Context(), id, claims.UserID); err != nil {
            handleSignOffError(w, err)
            return
        }

        m.IncrementSessionSignedCount()

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(MessageResponse{Message: "session signed off"})
    }
}
```

### Session Repository: Extended UpdateStatus for Sign-off
```go
// SignOff transitions a session to 'signed' status, setting signed_at and signed_by.
func (r *PostgresSessionRepository) SignOff(ctx context.Context, id int64, clinicianID int64, expectedVersion int) error {
    now := time.Now()

    result, err := r.db.ExecContext(ctx,
        `UPDATE sessions SET status = 'signed', signed_at = $1, signed_by = $2,
            version = version + 1, updated_at = $1, updated_by = $2
        WHERE id = $3 AND version = $4 AND status = 'awaiting_signoff'`,
        now, clinicianID, id, expectedVersion,
    )
    if err != nil {
        return fmt.Errorf("signing off session: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("checking rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return service.ErrSessionVersionConflict
    }

    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Application-only immutability | DB triggers + application checks | Standard practice | DB triggers prevent bypass via direct SQL; application provides good error UX |
| Flat text audit logs | JSONB old/new values with row_to_json() | PostgreSQL 9.4+ (JSONB) | Rich queryable audit data; no need for hstore extension |
| Manual audit insert in each handler | Generic AFTER trigger on tables | Standard practice | Zero application code for audit capture; impossible to forget |

**Deprecated/outdated:**
- hstore-based audit: JSONB is preferred since PostgreSQL 9.4; richer type support, no extension needed
- pgaudit extension: Useful for DDL/statement auditing but too low-level for entity-level clinical audit requirements

## Open Questions

1. **Screening table name**
   - What we know: The migration file is `create_contraindication_screening.sql` but need to verify exact table name
   - What's unclear: Could be `contraindication_screenings` or `contraindication_screening`
   - Recommendation: Check the migration file during planning; use the exact table name from the CREATE TABLE statement

2. **Should audit trail capture session_indication_codes and session_outcome_endpoints junction tables?**
   - What we know: These are junction tables without created_by/updated_by columns
   - What's unclear: Whether changes to indication codes constitute "clinical entity" changes worth auditing
   - Recommendation: Skip junction tables for now; they are managed via DELETE+INSERT loops and the parent entity's update is already audited

3. **Should the Lock transition (Signed -> Locked) be automatic after sign-off or manual?**
   - What we know: Requirements show Signed and Locked as separate states; the transition map allows Signed -> Locked
   - What's unclear: Whether there is a delay between signing and locking (e.g., review period)
   - Recommendation: Keep them as separate explicit transitions. SignOff moves to Signed, a separate Lock endpoint moves to Locked. This matches the existing transition map and provides flexibility for a future review period.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1 |
| Config file | None (Go stdlib test runner) |
| Quick run command | `go test ./internal/service/... -run TestSignoff -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LOCK-01 | Validation blocks sign-off when incomplete | unit | `go test ./internal/service/ -run TestValidateForSignoff -count=1` | No - Wave 0 |
| LOCK-02 | Sign-off records timestamp and clinician ID | unit | `go test ./internal/service/ -run TestSignOff -count=1` | No - Wave 0 |
| LOCK-03 | Signed session is immutable | unit + manual-DB | `go test ./internal/service/ -run TestSignedSessionImmutable -count=1` | No - Wave 0 |
| LOCK-04 | Addendum creation on locked session | unit | `go test ./internal/service/ -run TestCreateAddendum -count=1` | No - Wave 0 |
| LOCK-05 | Addendum is immutable once saved | manual-DB | Manual: attempt UPDATE on session_addendums via psql | No - DB trigger verification |
| LOCK-06 | DB-level immutability via triggers | manual-DB | Manual: attempt UPDATE on signed session via psql, verify RAISE EXCEPTION | No - DB trigger verification |
| AUDIT-01 | All CRUD operations logged | manual-DB | Manual: perform operations, query audit_trail table | No - DB trigger verification |
| AUDIT-02 | Audit entries have required fields | unit | `go test ./internal/service/ -run TestAuditEntry -count=1` | No - Wave 0 |
| AUDIT-03 | Audit log is append-only | manual-DB | Manual: attempt UPDATE/DELETE on audit_trail via psql | No - DB trigger verification |
| AUDIT-04 | Sign-off/lock events in audit trail | manual-DB | Manual: sign off session, query audit_trail for action | No - DB trigger verification |
| META-02 | Signed records have signed_at, signed_by | unit | `go test ./internal/service/ -run TestSignOff -count=1` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/service/... -count=1 && go test ./internal/repository/... -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green + manual DB trigger verification before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/service/signoff_test.go` -- covers LOCK-01, LOCK-02, LOCK-03
- [ ] `internal/service/addendum_test.go` -- covers LOCK-04
- [ ] `internal/service/audit_test.go` -- covers AUDIT-02
- [ ] `internal/testutil/mock_addendum.go` -- mock addendum repository
- [ ] `internal/testutil/mock_audit.go` -- mock audit repository
- [ ] DB trigger verification is manual (triggers tested by attempting forbidden SQL; not unit-testable without live DB)

## Sources

### Primary (HIGH confidence)
- PostgreSQL official docs: [Trigger Functions (PL/pgSQL)](https://www.postgresql.org/docs/current/plpgsql-trigger.html) - trigger syntax, RAISE EXCEPTION, TG_OP, TG_TABLE_NAME, BEFORE/AFTER timing
- PostgreSQL official docs: [Overview of Trigger Behavior](https://www.postgresql.org/docs/current/trigger-definition.html) - trigger execution order, BEFORE vs AFTER semantics
- Codebase analysis: `internal/service/session.go` - existing state transition map, session lifecycle
- Codebase analysis: `internal/repository/postgres/session.go` - existing UpdateStatus pattern
- Codebase analysis: `migrations/*.sql` - all existing table schemas, column patterns

### Secondary (MEDIUM confidence)
- [PostgreSQL Wiki: Audit trigger 91plus](https://wiki.postgresql.org/wiki/Audit_trigger_91plus) - audit trigger design patterns, row_to_json approach
- [How to Implement Audit Trails with Triggers in PostgreSQL](https://oneuptime.com/blog/post/2026-01-25-postgresql-audit-trails-triggers/view) - modern audit trail patterns
- [Working with Postgres Audit Triggers | EDB](https://www.enterprisedb.com/postgres-tutorials/working-postgres-audit-triggers) - enterprise audit trigger patterns

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all existing stack, no new dependencies
- Architecture: HIGH - follows established codebase patterns (service/repo/handler layers), PostgreSQL trigger syntax is well-documented
- Pitfalls: HIGH - identified through codebase analysis (child table joins, self-referential session trigger, user_id extraction)
- Migrations: HIGH - exact table names and column patterns verified from existing migration files

**Research date:** 2026-03-08
**Valid until:** 2026-04-08 (stable domain, no moving dependencies)
