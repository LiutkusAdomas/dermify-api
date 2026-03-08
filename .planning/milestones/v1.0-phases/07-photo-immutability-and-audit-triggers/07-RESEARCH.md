# Phase 7: Photo Immutability and Audit Triggers - Research

**Researched:** 2026-03-08
**Domain:** PostgreSQL triggers, database-level immutability enforcement, audit logging
**Confidence:** HIGH

## Summary

Phase 7 is a narrowly-scoped gap closure phase. The v1.0 milestone audit identified that `session_photos` (added in Phase 6) was not included in the immutability triggers (migration `20260308040003`) or audit triggers (migration `20260308040004`) created in Phase 5. This means LOCK-06 (DB-level immutability) and AUDIT-01 (all clinical operations logged) are only partially satisfied -- photo records can be modified or deleted at the database level on signed/locked sessions, and photo operations are not recorded in the audit trail.

The fix is straightforward: a single new Goose migration that attaches the existing `prevent_signed_session_modification()` trigger function to `session_photos` for immutability, and attaches the existing `audit_trigger_function()` to `session_photos` for audit logging. No new PL/pgSQL functions are needed. No Go application code changes are required.

**Primary recommendation:** Create one migration file with two `CREATE TRIGGER` statements reusing the existing trigger functions. No application code changes needed.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| LOCK-06 | Immutability is enforced at the database level (not just application layer) | Existing `prevent_signed_session_modification()` function handles tables with `session_id` FK. `session_photos` has `session_id BIGINT NOT NULL REFERENCES sessions(id)`. Attaching a BEFORE UPDATE OR DELETE trigger on `session_photos` using this function closes the gap. |
| AUDIT-01 | System logs all create, update, and delete operations on clinical entities | Existing `audit_trigger_function()` extracts entity_id, user_id (from created_by/updated_by), and row snapshots. `session_photos` has `id`, `created_by`, and `updated_by` columns. Attaching an AFTER INSERT OR UPDATE OR DELETE trigger on `session_photos` using this function closes the gap. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| PostgreSQL | 14+ | Database with PL/pgSQL triggers | Already in use; all trigger functions exist |
| pressly/goose v3 | v3 | Migration management | Project standard; embedded SQL via `//go:embed` |

### Supporting
No additional libraries needed. This phase creates only SQL migration files.

## Architecture Patterns

### Recommended Project Structure
```
migrations/
  20260308060000_add_session_photos_triggers.sql   # NEW: the only file this phase creates
```

No changes to any Go source files under `internal/`.

### Pattern 1: Reuse Existing Trigger Functions
**What:** The Phase 5 migrations created two reusable trigger functions: `prevent_signed_session_modification()` for immutability and `audit_trigger_function()` for audit logging. Both are designed to work with any table that follows the project's conventions.
**When to use:** When a new clinical table is added that has a `session_id` FK column, `created_by`, and `updated_by` columns.
**Source:** `migrations/20260308040003_create_immutability_triggers.sql`, `migrations/20260308040004_create_audit_triggers.sql`

The immutability function `prevent_signed_session_modification()` has two code paths:
1. For the `sessions` table itself (checks `TG_TABLE_NAME = 'sessions'`)
2. For child tables with `session_id` column -- looks up `SELECT status FROM sessions WHERE id = OLD.session_id`

`session_photos` has a direct `session_id` FK, so it hits code path 2 -- no modifications needed to the function.

The audit function `audit_trigger_function()` extracts:
- `entity_id` from `OLD.id` (DELETE) or `NEW.id` (INSERT/UPDATE)
- `user_id` from `NEW.created_by` (INSERT), `NEW.updated_by` (UPDATE), `OLD.updated_by` (DELETE)
- Row snapshots via `row_to_json(OLD/NEW)::JSONB`

`session_photos` has all required columns: `id`, `created_by`, `updated_by`. The function will work without modification.

### Pattern 2: Exact Trigger Naming Convention
**What:** All existing triggers follow a consistent naming pattern.
**Convention:**

For immutability triggers:
```sql
CREATE TRIGGER enforce_{table_short_name}_immutability
    BEFORE UPDATE OR DELETE ON {table_name}
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();
```

For audit triggers:
```sql
CREATE TRIGGER audit_{table_name}
    AFTER INSERT OR UPDATE OR DELETE ON {table_name}
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();
```

Existing examples:
- `enforce_consent_immutability` on `session_consents`
- `enforce_outcome_immutability` on `session_outcomes`
- `audit_session_consents` on `session_consents`
- `audit_session_outcomes` on `session_outcomes`

For `session_photos`, the names should be:
- `enforce_photo_immutability` on `session_photos`
- `audit_session_photos` on `session_photos`

### Pattern 3: Goose Migration File Naming
**What:** Migration files follow `YYYYMMDDHHMMSS_description.sql` naming.
**Convention:** The existing migrations use a scheme where the first 8 digits encode the date and the next 6 digits encode a series number within that date. Phase 5 uses `040000`-`040004`, Phase 6 uses `050000`. Phase 7 should use `060000`.
**Example:** `20260308060000_add_session_photos_triggers.sql`

### Pattern 4: Migration Up/Down Completeness
**What:** Every migration has both `-- +goose Up` and `-- +goose Down` sections.
**Convention:** Down sections reverse the Up section. For triggers, the Down section drops the triggers but does NOT drop the shared functions (since those were created in migration `040003`/`040004` and would be dropped by those migrations' Down sections).

### Anti-Patterns to Avoid
- **Modifying existing migrations:** Never alter `20260308040003` or `20260308040004`. These are already applied. Create a new migration that adds the missing triggers.
- **Creating new trigger functions:** The existing functions are generic and reusable. Do not duplicate them.
- **Changing Go application code:** No Go code changes are needed. The application-layer checks in `PhotoService` already exist and will continue to work. The DB triggers add defense-in-depth.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Session status check in trigger | New PL/pgSQL function | `prevent_signed_session_modification()` | Already handles session_id FK tables generically |
| Audit logging in trigger | New PL/pgSQL function | `audit_trigger_function()` | Already handles created_by/updated_by extraction generically |

**Key insight:** Both trigger functions were designed to be generic. They use `TG_TABLE_NAME` for dynamic table identification and `OLD.session_id` / `NEW.created_by` / `NEW.updated_by` for data extraction. No custom code is needed.

## Common Pitfalls

### Pitfall 1: Modifying Already-Applied Migrations
**What goes wrong:** Editing `20260308040003` or `20260308040004` to add the session_photos triggers would cause checksum mismatches in Goose's migration tracking table on any database that has already run those migrations.
**Why it happens:** It seems simpler to add lines to existing files than create a new one.
**How to avoid:** Always create a new migration file with a timestamp after the last existing migration.
**Warning signs:** goose reporting "checksum mismatch" or "already applied" errors.

### Pitfall 2: Forgetting the Down Section
**What goes wrong:** If the Down section is missing or incomplete, rollbacks will fail, leaving the database in an inconsistent state.
**Why it happens:** Down sections feel optional during development.
**How to avoid:** Always write `DROP TRIGGER IF EXISTS` statements in the Down section. Do NOT drop the shared functions -- those belong to migrations `040003` and `040004`.

### Pitfall 3: Wrong Trigger Timing (BEFORE vs AFTER)
**What goes wrong:** Using AFTER for immutability triggers means the modification has already been applied before the trigger fires. Using BEFORE for audit triggers means the trigger could prevent the operation from completing.
**Why it happens:** Confusing the two trigger types.
**How to avoid:** Immutability triggers are BEFORE (to block the operation). Audit triggers are AFTER (to record what happened).

### Pitfall 4: INSERT Blocking by Immutability Trigger
**What goes wrong:** If the immutability trigger fires on INSERT, it would block adding photos to sessions in draft/in_progress state.
**Why it happens:** Accidentally including INSERT in the trigger events.
**How to avoid:** The immutability trigger must fire on UPDATE OR DELETE only (not INSERT). Adding photos to an editable session is allowed; only modifying or deleting them on signed/locked sessions should be blocked. The existing trigger function references `OLD.session_id` which is only available for UPDATE/DELETE, not INSERT.

## Code Examples

### The Complete Migration File

```sql
-- +goose Up
-- +goose StatementBegin

-- Immutability: block UPDATE/DELETE on session_photos when parent session is signed or locked.
-- Reuses the existing prevent_signed_session_modification() function from migration 20260308040003.
CREATE TRIGGER enforce_photo_immutability
    BEFORE UPDATE OR DELETE ON session_photos
    FOR EACH ROW EXECUTE FUNCTION prevent_signed_session_modification();

-- Audit: record all INSERT/UPDATE/DELETE on session_photos in the audit_trail table.
-- Reuses the existing audit_trigger_function() from migration 20260308040004.
CREATE TRIGGER audit_session_photos
    AFTER INSERT OR UPDATE OR DELETE ON session_photos
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS enforce_photo_immutability ON session_photos;
DROP TRIGGER IF EXISTS audit_session_photos ON session_photos;
-- +goose StatementEnd
```

Source: Follows exact patterns from `migrations/20260308040003_create_immutability_triggers.sql` and `migrations/20260308040004_create_audit_triggers.sql`.

### Verification: How the Triggers Work

When `enforce_photo_immutability` fires on `session_photos`:
1. `TG_TABLE_NAME` = `'session_photos'` (not `'sessions'`), so it skips the first IF block
2. Executes: `SELECT status INTO v_session_status FROM sessions WHERE id = OLD.session_id`
3. If status is `'signed'` or `'locked'`, raises: `Cannot modify session_photos when session is in signed state`
4. Otherwise, allows the operation

When `audit_session_photos` fires on `session_photos`:
1. For INSERT: `v_entity_id = NEW.id`, `v_user_id = NEW.created_by`, captures `row_to_json(NEW)::JSONB`
2. For UPDATE: `v_entity_id = NEW.id`, `v_user_id = NEW.updated_by`, captures both old and new
3. For DELETE: `v_entity_id = OLD.id`, `v_user_id = OLD.updated_by`, captures `row_to_json(OLD)::JSONB`
4. Inserts into `audit_trail` with `entity_type = 'session_photos'`

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| App-layer only photo protection | App-layer + DB triggers | Phase 7 (this phase) | Defense-in-depth: even direct SQL cannot bypass immutability |

No deprecated approaches -- this phase simply extends existing patterns to cover a table that was added after the triggers were created.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify/assert |
| Config file | Makefile `test` target |
| Quick run command | `go test ./internal/... -count=1 -short` |
| Full suite command | `go test ./... -count=1 -v` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LOCK-06 | UPDATE/DELETE on session_photos blocked when session signed/locked | integration (DB required) | Manual: requires running PostgreSQL | N/A - DB trigger, not unit testable |
| AUDIT-01 | INSERT/UPDATE/DELETE on session_photos recorded in audit_trail | integration (DB required) | Manual: requires running PostgreSQL | N/A - DB trigger, not unit testable |

### Sampling Rate
- **Per task commit:** `go build ./...` (migration files are embedded; build ensures they parse)
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Build green + migration applies cleanly on running PostgreSQL

### Wave 0 Gaps
None -- this phase creates only a SQL migration file. No new Go test infrastructure is needed. The migration's correctness is verified by:
1. Build passes (embedded FS includes the migration)
2. Migration applies without error on PostgreSQL
3. Manual verification: attempt UPDATE/DELETE on session_photos when session is signed/locked
4. Manual verification: check audit_trail entries after photo INSERT

The existing application-level tests in `internal/service/photo_test.go` continue to cover the Go service layer behavior. DB trigger testing requires integration tests against a running PostgreSQL instance, which is consistent with how Phase 5 triggers are verified (see Phase 5 VERIFICATION.md "Human Verification Required" section).

## Open Questions

None. This phase is fully defined by the existing patterns and the audit findings. The implementation is mechanical -- two CREATE TRIGGER statements reusing existing functions.

## Sources

### Primary (HIGH confidence)
- `migrations/20260308040003_create_immutability_triggers.sql` - existing immutability trigger functions and patterns
- `migrations/20260308040004_create_audit_triggers.sql` - existing audit trigger function and patterns
- `migrations/20260308050000_create_session_photos.sql` - session_photos table schema (has session_id, created_by, updated_by)
- `migrations/20260308040002_create_audit_trail.sql` - audit_trail table schema
- `.planning/v1.0-MILESTONE-AUDIT.md` - gap identification for LOCK-06 and AUDIT-01
- `.planning/REQUIREMENTS.md` - requirement definitions

### Secondary (MEDIUM confidence)
None needed -- all information comes from the existing codebase.

### Tertiary (LOW confidence)
None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new libraries, just SQL using existing patterns
- Architecture: HIGH - exact replication of existing trigger attachment pattern
- Pitfalls: HIGH - all pitfalls are well-understood (migration immutability, trigger timing)

**Research date:** 2026-03-08
**Valid until:** Indefinite (patterns are stable and project-specific)
