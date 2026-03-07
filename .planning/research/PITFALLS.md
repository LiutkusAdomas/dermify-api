# Domain Pitfalls

**Domain:** Clinical procedure documentation API for aesthetic dermatology
**Researched:** 2026-03-07

---

## Critical Pitfalls

Mistakes that cause rewrites, data integrity failures, or medico-legal exposure.

---

### Pitfall 1: Application-Only Immutability Enforcement

**What goes wrong:** Immutability of signed-off records is enforced only in Go handler code (e.g., an `if session.Status == "locked" { return 403 }` check). A bug in any handler, a new endpoint added without the check, or direct database access by an operator bypasses the protection entirely. The record appears immutable to the API consumer but is actually mutable at the data layer.

**Why it happens:** It feels natural to add a guard clause in the handler and move on. Database triggers feel like "magic" that is harder to test. Developers defer the trigger to later and forget.

**Consequences:** A signed-off clinical record gets silently modified. In a medico-legal dispute, the clinic cannot prove the record was unaltered since sign-off. The entire audit trail is undermined because the foundation -- true immutability -- was never enforced.

**Prevention:**
1. Create a PostgreSQL `BEFORE UPDATE` trigger on every table that holds clinical record data. The trigger checks the `status` column (or a `locked_at` timestamp) and raises an exception if the row is locked:
   ```sql
   CREATE OR REPLACE FUNCTION prevent_locked_record_update()
   RETURNS TRIGGER AS $$
   BEGIN
     IF OLD.status = 'signed_off' THEN
       RAISE EXCEPTION 'Cannot modify signed-off record (id=%)', OLD.id;
     END IF;
     RETURN NEW;
   END;
   $$ LANGUAGE plpgsql;
   ```
2. Apply this trigger to: `treatment_sessions`, every procedure module table (`ipl_modules`, `filler_modules`, etc.), `consent_records`, `outcome_records`.
3. Also enforce in the Go application layer -- defense in depth. The trigger is the last line of defense; the handler check provides a clean error response.
4. Write integration tests that attempt UPDATE and DELETE on locked records and assert they fail at the database level.

**Detection:** If you can run `UPDATE treatment_sessions SET notes = 'tampered' WHERE status = 'signed_off'` and it succeeds, the protection is missing.

**Phase mapping:** Must be implemented in the same phase as treatment session lifecycle and sign-off. Do not defer triggers to a later phase.

**Confidence:** HIGH -- this is a well-documented pattern in medical record systems. Multiple sources confirm database-level enforcement is required for medico-legal defensibility.

---

### Pitfall 2: Audit Trail That Only Captures "What" Not "Who/When/Why"

**What goes wrong:** The audit trail records that a field changed but omits: which user made the change, the timestamp with timezone, the IP address or request context, and crucially, the previous value. When an addendum is added to a locked record, there is no link back to the original record state that explains what is being amended.

**Why it happens:** Developers build a simple `audit_logs` table with `(record_id, action, timestamp)` and consider audit done. The detailed before/after state, actor identity, and amendment rationale get deferred.

**Consequences:** During a complaint investigation, the clinic cannot reconstruct who documented what and when. Addendums exist but lack context about what they are correcting or supplementing. Regulators consider the audit trail incomplete.

**Prevention:**
1. Design the audit table with mandatory fields from the start:
   ```sql
   CREATE TABLE audit_trail (
     id          BIGSERIAL PRIMARY KEY,
     entity_type VARCHAR(50) NOT NULL,  -- 'treatment_session', 'ipl_module', etc.
     entity_id   BIGINT NOT NULL,
     action      VARCHAR(20) NOT NULL,  -- 'created', 'updated', 'signed_off', 'addendum'
     actor_id    BIGINT NOT NULL REFERENCES users(id),
     actor_role  VARCHAR(30) NOT NULL,
     old_values  JSONB,                 -- snapshot of changed fields before
     new_values  JSONB,                 -- snapshot of changed fields after
     reason      TEXT,                  -- required for addendums
     ip_address  INET,
     request_id  UUID,
     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   ```
2. Make audit writes happen inside the same database transaction as the data change. Do NOT log asynchronously for clinical data -- a committed change without an audit entry is a compliance failure.
3. Use the existing request_id middleware value (already in the codebase) to correlate API requests with audit entries.
4. For addendums, require a `reason` field -- this is non-negotiable for medico-legal compliance.

**Detection:** Query for audit entries missing `actor_id` or `old_values`. If any exist, the implementation is incomplete.

**Phase mapping:** The audit table schema must be created in the same phase as treatment sessions. Every subsequent phase (modules, consent, outcomes) must write audit entries from day one.

**Confidence:** HIGH -- FDA 21 CFR Part 11 and HIPAA audit requirements are well-documented. Multiple sources confirm the need for actor, timestamp, and before/after capture.

---

### Pitfall 3: Polymorphic Procedure Modules as a Single Wide Table

**What goes wrong:** All six procedure module types (IPL, Nd:YAG, CO2, RF, Fillers, Botulinum) are crammed into one table with nullable columns for every possible field. The IPL-specific `pulse_width_ms` column is NULL for filler records; the filler-specific `injection_depth` column is NULL for laser records. The table accumulates 80+ nullable columns.

**Why it happens:** It seems simpler than managing six tables. Queries are "easier" because there are no joins. The developer thinks "I will just add columns as needed."

**Consequences:**
- Cannot use NOT NULL constraints on module-specific required fields. A filler record with no `product_id` silently passes database validation.
- CHECK constraints become nightmarishly complex: `CHECK (module_type = 'filler' AND product_id IS NOT NULL OR module_type != 'filler')` repeated for dozens of fields.
- PostgreSQL query planner loses statistics accuracy on sparse columns. Heap in their research found JSONB/wide-table patterns can cause 2000x slower queries due to planner misjudgments.
- Adding a seventh module type (e.g., chemical peels) requires altering the production table with even more nullable columns and adjusting all existing constraints.
- Go structs become unwieldy: a single `ProcedureModule` struct with 80+ pointer fields is painful to work with and test.

**Prevention:** Use the table-per-type (class table inheritance) pattern:
1. A shared `procedure_modules` base table holds common fields:
   ```sql
   CREATE TABLE procedure_modules (
     id              BIGSERIAL PRIMARY KEY,
     session_id      BIGINT NOT NULL REFERENCES treatment_sessions(id),
     module_type     VARCHAR(30) NOT NULL,  -- 'ipl', 'ndyag', 'co2', 'rf', 'filler', 'botulinum'
     treatment_area  VARCHAR(100),
     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   ```
2. Each module type gets its own table with a FK back to `procedure_modules`:
   ```sql
   CREATE TABLE ipl_modules (
     id                BIGINT PRIMARY KEY REFERENCES procedure_modules(id),
     device_id         BIGINT NOT NULL REFERENCES devices(id),
     wavelength_nm     INTEGER NOT NULL,
     pulse_width_ms    NUMERIC NOT NULL,
     fluence_j_cm2     NUMERIC NOT NULL,
     spot_size_mm      INTEGER,
     cooling_method    VARCHAR(50),
     number_of_passes  INTEGER NOT NULL DEFAULT 1
   );
   ```
3. Go code maps cleanly: one `ProcedureModule` base struct + one struct per type. No pointer fields for "maybe present" columns.
4. Adding a new module type = new migration + new table + new Go struct. Zero changes to existing tables.

**Detection:** If any procedure module table has more than 5 columns that are nullable only because they belong to a different module type, the design is wrong.

**Phase mapping:** This schema decision must be made before any procedure module is implemented. Changing from single-table to table-per-type after data exists requires a painful migration.

**Confidence:** HIGH -- PostgreSQL documentation, DoltHub analysis, and Hashrocket all recommend table-per-type or exclusive-belongs-to for this pattern. The anti-pattern is well-documented.

---

### Pitfall 4: Sign-Off Without Validation Gate

**What goes wrong:** The sign-off endpoint sets `status = 'signed_off'` without checking whether all required components are present: consent captured, contraindications screened, at least one procedure module attached, outcomes recorded. Records get locked in an incomplete state and can never be corrected (because they are now immutable).

**Why it happens:** Sign-off validation gets implemented as "we will add checks later" but the locking mechanism ships first. Or the validation list is incomplete -- consent is checked but adverse event fields are not.

**Consequences:** A signed-off record is missing required documentation. Because the record is locked, the clinician cannot go back and fill in the gaps. The only recourse is an addendum, which creates a messy paper trail and may not satisfy compliance reviewers. At scale, clinics discover months later that hundreds of records are locked with missing data.

**Prevention:**
1. Define a validation checklist as a first-class concept in the service layer:
   ```go
   type SignOffChecklist struct {
     HasConsent           bool
     ContraindictionsDone bool
     HasProcedureModule   bool
     HasOutcome           bool
     AllRequiredFields    bool
   }
   ```
2. The sign-off handler calls a `ValidateForSignOff(sessionID)` function that returns a list of blocking failures. If any fail, return HTTP 422 with the specific missing items.
3. Make validation rules data-driven, not hardcoded in handler logic. Store them in a config or a validation rules table so they can evolve without code changes.
4. Write tests that attempt sign-off with every combination of missing components and assert they are rejected.

**Detection:** If the sign-off endpoint does not return a structured list of validation failures, or if you can sign off a session with no procedure modules attached, the gate is broken.

**Phase mapping:** Validation gate must ship in the same phase as sign-off. Never deploy sign-off without the validation gate -- even for "testing purposes."

**Confidence:** HIGH -- this is a universal medical records requirement. The PROJECT.md explicitly calls out "block if required checks incomplete."

---

### Pitfall 5: Photo Storage Without Metadata Stripping and Access Control

**What goes wrong:** Clinical before/after photos are saved to the local filesystem with EXIF metadata intact (GPS coordinates, device serial number, timestamps that may not match server time). File paths are constructed from user-supplied filenames without sanitization. Photos are served via a static file handler without authentication checks, making them accessible to anyone who guesses the URL.

**Why it happens:** File upload is treated as a simple `os.Create` + `io.Copy` operation. The developer focuses on getting the upload working and defers security hardening. EXIF stripping requires an image processing library, which feels like scope creep.

**Consequences:**
- EXIF GPS data exposes the clinic's exact location or, worse, a patient's home location if photos were taken on a mobile device. Research shows 82-90% of clinicians are noncompliant with photo handling best practices.
- Path traversal attacks via filenames like `../../../etc/passwd` can read or overwrite arbitrary files. Go 1.23 (project version) does not have `os.Root` -- that shipped in Go 1.24.
- Unauthenticated photo access means clinical images are exposed as public URLs. Patient photos are PHI.
- Photos stored alongside the binary in the container filesystem are lost on container restart if the volume mount is misconfigured.

**Prevention:**
1. **Never use user-supplied filenames.** Generate UUIDs for storage: `{session_id}/{uuid}.jpg`. Store the original filename in the database, not the filesystem.
2. **Strip EXIF metadata** on upload. Use a Go image library (e.g., `disintegration/imaging` for decode/re-encode, or `rwcarlsen/goexif` to read and strip). Re-encoding the image as a new JPEG inherently strips EXIF.
3. **Serve photos through an authenticated endpoint**, not a static file handler. The handler verifies the JWT, checks the user has access to the session, then streams the file. Never expose the filesystem path in the URL.
4. **Validate file type** by reading magic bytes (`image/jpeg`, `image/png`), not by trusting the file extension or Content-Type header.
5. **Sanitize the storage path:** use `filepath.Clean`, reject any path containing `..`, and ensure the resolved path is within the configured upload directory.
6. **Configure the Docker volume mount** for the photo storage directory and document it as a required deployment step.

**Detection:** Upload a JPEG with GPS EXIF data, download it, and check if the GPS data is still present. Try uploading a file named `../../etc/passwd`. Try accessing a photo URL without an auth token.

**Phase mapping:** Photo upload phase. These protections are not "nice to have" -- they must ship with the initial photo feature.

**Confidence:** HIGH -- PMC clinical photography guidelines, HIPAA photography rules, and OWASP path traversal documentation all confirm these requirements.

---

## Moderate Pitfalls

---

### Pitfall 6: Validation Rule Explosion in Handler Code

**What goes wrong:** Each of the six procedure module types has 15-30 fields with complex conditional validation rules (e.g., "if device type is X, then fluence range is Y-Z"; "if reconstitution_method is saline, then dilution_ratio is required"). These rules end up as nested if/else chains inside handler functions, mixing HTTP concerns with domain validation.

**Why it happens:** The existing codebase pattern (see `HandleRegister`) does inline validation: `if req.Field == "" { return error }`. This works for 3 fields but does not scale to 30 conditional fields across 6 module types.

**Consequences:**
- Handlers grow to 200+ lines of validation logic, violating the CLAUDE.md guideline to "keep handlers thin."
- Validation rules are duplicated: the same fluence range check appears in the create handler, update handler, and sign-off validation gate.
- Changing a validation rule (e.g., new device supports a wider fluence range) requires finding every place the rule is checked.
- Testing is painful: each combination of conditional fields needs a test case, and the tests are tightly coupled to HTTP handler code.

**Prevention:**
1. Use `go-ozzo/ozzo-validation` instead of struct tags for complex conditional rules. Ozzo supports programmatic rule construction:
   ```go
   func (m *IPLModuleInput) Validate() error {
     return validation.ValidateStruct(m,
       validation.Field(&m.WavelengthNM, validation.Required, validation.Min(400), validation.Max(1200)),
       validation.Field(&m.FluenceJCM2, validation.Required, validation.Min(m.minFluence()), validation.Max(m.maxFluence())),
       validation.Field(&m.CoolingMethod, validation.When(m.FluenceJCM2 > 20, validation.Required)),
     )
   }
   ```
2. Put validation in the domain/service layer, not the handler. The handler decodes JSON, calls `input.Validate()`, and returns structured errors. The service layer can also call `Validate()` during sign-off gate checks.
3. Define device-specific parameter ranges in seed data (the project already plans hardcoded seed data). Validation rules reference the seed data rather than hardcoding numbers.
4. Return all validation errors at once (not fail-fast). Clinicians filling out a 20-field form should see all problems, not fix them one at a time.

**Detection:** If any handler function is longer than 50 lines of validation code, or if you see the same numeric range check in more than one file, the pattern is wrong.

**Phase mapping:** Establish the validation pattern in the first procedure module phase (IPL). All subsequent modules follow the same pattern. Retrofitting validation architecture after six modules exist is painful.

**Confidence:** HIGH -- ozzo-validation documentation confirms conditional validation support. The existing codebase pattern will not scale to the documented field complexity.

---

### Pitfall 7: Consent Workflow That Ignores Temporal Edge Cases

**What goes wrong:** The consent record is a simple boolean: `consent_given: true`. The system does not track when consent was given, who witnessed it, whether it was for the specific procedure being documented, or what happens if consent is withdrawn mid-session. A consent record created yesterday for "laser treatment" is treated as valid for today's filler injection.

**Why it happens:** Consent feels like a checkbox -- the patient said yes, record it, move on. The temporal and scope dimensions are overlooked because the immediate use case is "document that consent was obtained."

**Consequences:**
- A patient consents to IPL treatment but the clinician also performs RF treatment in the same session. The consent record does not cover RF. In a dispute, the clinic cannot prove informed consent for the RF procedure.
- Consent is recorded after the procedure is documented (out of order). The timestamp proves the procedure happened before consent was given.
- A patient verbally withdraws consent during a session. The system has no mechanism to record withdrawal; the original consent flag remains `true`.
- Consent scope is unclear: does it cover photos? Data retention? The specific procedure type?

**Prevention:**
1. Model consent as a rich record, not a boolean:
   ```sql
   CREATE TABLE consent_records (
     id              BIGSERIAL PRIMARY KEY,
     session_id      BIGINT NOT NULL REFERENCES treatment_sessions(id),
     consent_type    VARCHAR(50) NOT NULL, -- 'procedure', 'photography', 'data_retention'
     procedure_types TEXT[],               -- which module types this consent covers
     consented_at    TIMESTAMPTZ NOT NULL,
     consented_by    BIGINT NOT NULL REFERENCES users(id), -- clinician recording
     patient_confirmation TEXT,            -- how confirmed: 'verbal', 'signed_form', 'digital'
     withdrawn_at    TIMESTAMPTZ,
     withdrawn_by    BIGINT REFERENCES users(id),
     withdrawal_reason TEXT,
     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   ```
2. Sign-off validation gate checks: does a valid (non-withdrawn) consent record exist for each procedure module type in the session?
3. Enforce temporal ordering: consent `consented_at` must be before or equal to the session's first procedure module `created_at`. Log a warning if consent is recorded after procedure documentation begins.
4. Support consent withdrawal: set `withdrawn_at` and `withdrawn_by`. If a session has withdrawn consent and the record is not yet signed off, block sign-off until the clinician addresses it (removes the unconsented module or obtains new consent).

**Detection:** If the consent table has no `consent_type` column or no `withdrawn_at` column, the model is too simplistic.

**Phase mapping:** Consent and safety checks phase. The consent model must be designed before the first procedure module is implemented, because the sign-off gate needs to cross-reference consent scope against attached modules.

**Confidence:** MEDIUM -- consent requirements vary by jurisdiction. The specific fields above are based on general medico-legal best practice, not a specific regulatory framework. Validate with the clinic's legal requirements.

---

### Pitfall 8: Addendum Model That Allows Stealth Corrections

**What goes wrong:** Addendums are implemented as free-text notes appended to a locked record, but there is no structured link between the addendum and the specific field or section being amended. Or worse, addendums are allowed to modify the original record's data rather than being stored separately.

**Why it happens:** The simplest addendum implementation is a `notes` text column or a linked `addendums` table with just a `text` field. It ships fast but lacks the structure needed for compliance.

**Consequences:**
- An addendum says "corrected fluence to 15 J/cm2" but there is no programmatic way to identify which field was amended or what the original value was.
- If addendums can overwrite original data, the record is no longer immutable -- the addendum mechanism defeats the locking mechanism.
- Multiple addendums on the same record with no ordering or categorization create confusion about the current correct interpretation of the record.

**Prevention:**
1. Addendums are always separate rows in an `addendums` table, never modifications to the original record:
   ```sql
   CREATE TABLE addendums (
     id              BIGSERIAL PRIMARY KEY,
     session_id      BIGINT NOT NULL REFERENCES treatment_sessions(id),
     addendum_type   VARCHAR(30) NOT NULL, -- 'correction', 'addition', 'clarification'
     section         VARCHAR(50),          -- 'ipl_module', 'outcome', 'consent', etc.
     field_reference VARCHAR(100),         -- optional: specific field being amended
     content         TEXT NOT NULL,
     author_id       BIGINT NOT NULL REFERENCES users(id),
     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   ```
2. The addendums table has its own immutability trigger: once created, an addendum cannot be modified or deleted.
3. When retrieving a session, the API returns the original record plus all addendums in chronological order. The client renders them visually distinct (e.g., addendum callouts below the original record).
4. Every addendum creation writes to the audit trail with full actor/timestamp/content capture.

**Detection:** If you can modify an addendum after creation, or if addendums are stored as an array field on the session row, the design is wrong.

**Phase mapping:** Must be designed alongside the sign-off/locking mechanism. Addendums are the "escape valve" for locked records; deploying locking without addendums means clinicians have no way to correct errors.

**Confidence:** HIGH -- the addendum-only amendment model is explicitly called out in PROJECT.md as a key decision. The structured approach is standard in medical record systems.

---

### Pitfall 9: Direct SQL in Handlers Does Not Scale to This Domain

**What goes wrong:** The existing codebase has handlers that call `db.QueryRow` directly (see `HandleRegister`, `HandleGetProfile`). This pattern is acceptable for simple auth flows but breaks down when a single operation (e.g., sign-off) needs to: validate across multiple tables, write to the session table, write to the audit table, check consent records, and do it all in a single transaction.

**Why it happens:** The current code works. There is no service layer or repository layer because the existing endpoints are simple CRUD. The developer continues the pattern for treatment sessions, and suddenly has 150-line handlers with nested transaction management.

**Consequences:**
- Transaction management (`db.Begin`, `tx.Commit`, `tx.Rollback`) ends up in handler code, mixing HTTP concerns with data integrity concerns.
- The same SQL query appears in multiple handlers (e.g., "check if session is locked" appears in every module update handler).
- Testing requires a live database for every handler test because there is no repository interface to mock.
- Error handling becomes inconsistent: some handlers roll back on error, others do not, leading to partial writes.

**Prevention:**
1. Introduce a repository layer before building treatment session features. Each domain gets a repository interface:
   ```go
   type SessionRepository interface {
     Create(ctx context.Context, input CreateSessionInput) (*Session, error)
     GetByID(ctx context.Context, id int64) (*Session, error)
     SignOff(ctx context.Context, id int64, actorID int64) error
   }
   ```
2. Introduce a service layer that orchestrates cross-repository operations within transactions:
   ```go
   type SessionService struct {
     sessions SessionRepository
     consent  ConsentRepository
     audit    AuditRepository
     db       *sql.DB // for transaction management
   }
   ```
3. The handler's only job: decode request, call service, encode response. All business logic (validation, authorization, audit logging) lives in the service layer.
4. This is explicitly anticipated in CLAUDE.md: "business logic should live in service/repository layers as they are introduced" and PROJECT.md: "introduce layers as domains grow."

**Detection:** If any handler function imports `database/sql` or calls `db.QueryRow` directly after treatment sessions are implemented, the layering is missing.

**Phase mapping:** The service/repository pattern should be established in the first phase that introduces treatment sessions. Retroactively extracting layers from 10+ handlers is a major refactor.

**Confidence:** HIGH -- the project documentation explicitly calls for this transition. The existing code confirms the current pattern will not scale.

---

### Pitfall 10: Race Condition on Concurrent Session Modification

**What goes wrong:** Two clinicians (or two browser tabs) modify the same treatment session simultaneously. Clinician A adds an IPL module; Clinician B updates the session header. Without concurrency control, Clinician B's save overwrites Clinician A's module attachment, or both succeed but the session is left in an inconsistent state.

**Why it happens:** Optimistic locking is not the default in Go/PostgreSQL. Without an explicit version column or `SELECT FOR UPDATE`, concurrent writes silently succeed based on last-write-wins.

**Consequences:** Data loss -- a clinician's documented procedure module disappears. Or a session gets signed off while another clinician is still adding data to it. In a medical context, lost procedure documentation is a patient safety issue, not just a UX annoyance.

**Prevention:**
1. Add a `version` column (integer, default 1) to `treatment_sessions` and all module tables.
2. Every update includes `WHERE id = $1 AND version = $2` and increments the version. If `RowsAffected() == 0`, return HTTP 409 Conflict with the current version.
3. For the sign-off operation specifically, use `SELECT ... FOR UPDATE` (pessimistic locking) because sign-off is a critical irreversible operation that must not race with concurrent modifications.
4. Return the current `version` in every GET response so the client can include it in subsequent updates.

**Detection:** Open two browser windows, load the same session, make different changes, submit both. If both succeed without conflict, locking is missing.

**Phase mapping:** Implement in the treatment session phase. The version column must exist from the first migration.

**Confidence:** HIGH -- HackerNoon Go/PostgreSQL locking comparison and multiple sources confirm this is a standard requirement for concurrent record modification.

---

## Minor Pitfalls

---

### Pitfall 11: Seed Data as Go Constants Instead of Migrations

**What goes wrong:** Controlled lists (devices, products, indication codes, clinical endpoints) are defined as Go constants or in-memory maps. They exist in application code but not in the database. Foreign key relationships to device or product records cannot be enforced because there are no rows to reference.

**Why it happens:** The PROJECT.md says "hardcoded seed data" which can be interpreted as "Go constants." The developer puts device names in a `var devices = map[string]Device{...}` and validates against it in code.

**Prevention:**
1. Seed data goes into the database via Goose migration files. Use `INSERT ... ON CONFLICT DO NOTHING` for idempotency.
2. Create proper tables (`devices`, `products`, `indication_codes`, `clinical_endpoints`) with foreign keys from procedure modules.
3. Go code can still have a constants file for documentation or for populating migrations, but the source of truth is the database.

**Detection:** If you cannot write `SELECT * FROM devices` and get results, the seed data is not in the database.

**Phase mapping:** Seed data tables and initial data should be created in the phase before procedure modules, so foreign keys can reference them from day one.

**Confidence:** HIGH -- relational integrity requires the referenced data to exist in the database.

---

### Pitfall 12: Treating `database/sql` + pgx as Equivalent to pgx Native

**What goes wrong:** The project uses pgx v4 through the `database/sql` compatibility layer (`pgx/v4/stdlib`). This works but sacrifices pgx-native features: PostgreSQL-specific types (`pgtype.JSONB`, `pgtype.Inet`, arrays, composite types), connection pool management, LISTEN/NOTIFY, and COPY protocol for bulk inserts.

**Why it happens:** The `database/sql` interface is familiar and portable. The developer does not realize they are leaving performance and type safety on the table.

**Consequences:** For this project specifically:
- Scanning JSONB audit `old_values`/`new_values` into Go requires manual `json.Unmarshal` from `[]byte` instead of using pgx's native JSONB scanning.
- PostgreSQL `TEXT[]` arrays (useful for consent `procedure_types`) cannot be scanned directly; they require a custom scanner or string parsing.
- Bulk seed data inserts are slow (row-by-row INSERT instead of COPY).

**Prevention:** This is a moderate concern, not a blocker. The `database/sql` interface is adequate for v1. If performance or type-handling becomes painful, consider migrating to pgx v5 native (which the project could adopt since it is Go 1.23+). Do not change the database layer mid-feature -- pick a migration point between phases.

**Detection:** If you find yourself writing custom `sql.Scanner` implementations for PostgreSQL-specific types in more than 3 places, consider the native driver.

**Phase mapping:** Not urgent. Evaluate after the first two procedure module types are implemented.

**Confidence:** MEDIUM -- the current approach works but may cause friction with complex PostgreSQL types.

---

### Pitfall 13: Missing `DOWN` Migrations for Clinical Tables

**What goes wrong:** The existing migrations have simple `DROP TABLE` in the down section. For clinical data tables, a careless `goose down` in production drops patient records. Additionally, developers write complex `UP` migrations but skip or write incomplete `DOWN` migrations, making rollbacks impossible.

**Prevention:**
1. Every migration must have a complete `DOWN` section (existing project convention).
2. For production deployments, consider adding a safety check: the down migration for tables containing clinical data should RAISE an error rather than dropping:
   ```sql
   -- +goose Down
   -- SAFETY: This migration cannot be rolled back in production.
   -- To roll back, manually verify no clinical data exists, then run:
   -- DROP TABLE treatment_sessions CASCADE;
   DO $$ BEGIN
     IF EXISTS (SELECT 1 FROM treatment_sessions LIMIT 1) THEN
       RAISE EXCEPTION 'Cannot drop treatment_sessions: table contains data';
     END IF;
     DROP TABLE treatment_sessions;
   END $$;
   ```
3. Test migrations both up and down in CI before deploying.

**Detection:** Run `goose up` then `goose down` then `goose up` in a test environment. If it fails on the second `up`, the `down` migration is incomplete.

**Phase mapping:** Every phase that creates tables. Enforce from the first treatment session migration.

**Confidence:** HIGH -- the existing migration pattern confirms this is a project convention. The clinical data protection aspect is the new concern.

---

### Pitfall 14: Forgetting DELETE Protection on Locked Records

**What goes wrong:** The immutability trigger prevents UPDATE on signed-off records but does not prevent DELETE. An API endpoint or database operation deletes a locked record entirely.

**Prevention:** The immutability trigger must also handle the `BEFORE DELETE` event:
```sql
CREATE TRIGGER prevent_locked_record_delete
BEFORE DELETE ON treatment_sessions
FOR EACH ROW
WHEN (OLD.status = 'signed_off')
EXECUTE FUNCTION prevent_locked_record_modification();
```

Apply to all clinical data tables. There should be no code path that can delete a signed-off record.

**Detection:** Attempt `DELETE FROM treatment_sessions WHERE status = 'signed_off'` -- it must fail.

**Phase mapping:** Same phase as Pitfall 1 (immutability triggers).

**Confidence:** HIGH -- trivially verifiable and commonly overlooked.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Treatment session lifecycle | Pitfall 4 (sign-off without validation gate), Pitfall 9 (SQL in handlers), Pitfall 10 (race conditions) | Establish service/repository layer and version column from the start |
| Record locking / sign-off | Pitfall 1 (app-only immutability), Pitfall 14 (DELETE not blocked), Pitfall 8 (addendum model) | Database triggers + addendum table must ship together |
| Procedure modules (IPL, Nd:YAG, etc.) | Pitfall 3 (single wide table), Pitfall 6 (validation explosion) | Table-per-type schema + ozzo-validation pattern established on first module |
| Consent and safety | Pitfall 7 (temporal edge cases), Pitfall 4 (consent not checked at sign-off) | Rich consent model with type/scope/withdrawal from the start |
| Photo handling | Pitfall 5 (EXIF, path traversal, unauthenticated access) | UUID filenames, EXIF stripping, authenticated serving endpoint |
| Audit trail | Pitfall 2 (incomplete audit data) | Full audit table schema designed before any clinical data is written |
| Device/product seed data | Pitfall 11 (seed data in Go code not database) | Migration-based seed data with proper FK tables |
| All phases | Pitfall 13 (incomplete down migrations) | Test up/down/up cycle in CI |

---

## Sources

- [Immutable Audit Trails Guide](https://www.hubifi.com/blog/immutable-audit-log-basics)
- [Tamper-Proof Audit Logs for Health SaaS - DEV Community](https://dev.to/beck_moulton/immutable-by-design-building-tamper-proof-audit-logs-for-health-saas-22dc)
- [Tamper-Evident Audit Trails in PostgreSQL - AppMaster](https://appmaster.io/blog/tamper-evident-audit-trails-postgresql)
- [Audit Log Paradigms - Go/PostgreSQL - DEV Community](https://dev.to/akkaraponph/comprehensive-research-audit-log-paradigms-gopostgresqlgorm-design-patterns-1jmm)
- [Modeling Polymorphic Associations - Hashrocket](https://hashrocket.com/blog/posts/modeling-polymorphic-associations-in-a-relational-database)
- [Polymorphic Database Schema Choices - DoltHub](https://www.dolthub.com/blog/2024-06-25-polymorphic-associations/)
- [When to Avoid JSONB in PostgreSQL - Heap](https://www.heap.io/blog/when-to-avoid-jsonb-in-a-postgresql-schema)
- [Optimistic vs Pessimistic Locking Go/PostgreSQL - HackerNoon](https://hackernoon.com/comparing-optimistic-and-pessimistic-locking-with-go-and-postgresql)
- [Safe Clinical Photography Guidelines - PMC](https://pmc.ncbi.nlm.nih.gov/articles/PMC8143941/)
- [HIPAA Photography Rules - HIPAA Journal](https://www.hipaajournal.com/hipaa-photography-rules/)
- [Path Traversal Prevention - OWASP](https://owasp.org/www-community/attacks/Path_Traversal)
- [Go 1.24 os.Root Security - Contrast Security](https://www.contrastsecurity.com/security-influencers/navigating-os.root-and-path-traversal-vulnerabilities-go-1.24-detection-and-protection-methods-contrast-security)
- [ozzo-validation Library](https://github.com/go-ozzo/ozzo-validation)
- [E-Signature Audit Trail Schema - Anvil](https://www.useanvil.com/blog/engineering/e-signature-audit-trail-schema-events-json-checklist/)
- [Consent Revocation in Healthcare - Simbo AI](https://www.simbo.ai/blog/revocation-of-consent-in-healthcare-implications-for-patient-records-and-provider-responsibilities-2472679/)
- [FDA Electronic Signatures - 21 CFR Part 11](https://www.ecfr.gov/current/title-21/chapter-I/subchapter-A/part-11)
- [PostgreSQL Immutable Store Pattern - Medium](https://medium.com/@_jba/using-postgresql-as-an-immutable-store-ac1819ca464d)
- [Dermatology Practice HIPAA Photo Tips - Practical Dermatology](https://practicaldermatology.com/topics/practice-management/keep-your-dermatology-practice-hipaa-compliant-with-these-5-photo-tips/20684/)
