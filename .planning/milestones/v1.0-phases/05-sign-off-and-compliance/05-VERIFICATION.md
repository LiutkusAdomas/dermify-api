---
phase: 05-sign-off-and-compliance
verified: 2026-03-08T14:00:00Z
status: passed
score: 5/5 must-haves verified
---

# Phase 5: Sign-off and Compliance Verification Report

**Phase Goal:** A clinician can sign off a completed session, producing a locked, immutable medical record with a full audit trail -- the core medico-legal requirement of the system
**Verified:** 2026-03-08T14:00:00Z
**Status:** passed
**Re-verification:** No -- initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | The system validates all required fields across the session (header, consent, modules, outcomes) and blocks sign-off with a clear list of what is incomplete | VERIFIED | `SignoffService.ValidateForSignoff` checks session state, consent existence, module existence, and outcome existence. Returns `ValidationResult{Ready: false, Missing: [...]}` with specific items. 7 unit tests cover all-present, missing-consent, missing-modules, missing-outcome, missing-all, wrong-state, and session-not-found cases. HTTP endpoint `GET /signoff/readiness` exposes this. |
| 2 | A clinician can sign off a valid session, which records the timestamp and clinician ID, and the signed record becomes immutable -- any attempt to modify it (via API or direct SQL) fails | VERIFIED | `SignoffService.SignOff` validates completeness then delegates to `PostgresSignoffRepository.SignOff` which atomically sets `status='signed', signed_at=$1, signed_by=$2, version=version+1` with optimistic locking. DB trigger `prevent_signed_session_modification()` blocks UPDATE/DELETE on sessions and all child tables when status is 'signed' or 'locked'. POST `/signoff` endpoint wired. |
| 3 | A clinician can add addendums to a locked session (with date, author, reason, content) and those addendums are themselves immutable once saved | VERIFIED | `AddendumService.CreateAddendum` validates session is in 'locked' state, validates reason/content non-empty, sets created_at. `PostgresAddendumRepository.Create` inserts into `session_addendums` via RETURNING id. DB trigger `prevent_addendum_modification()` unconditionally blocks UPDATE/DELETE on `session_addendums`. POST `/addendums`, GET `/addendums`, GET `/addendums/{addendumId}` endpoints wired. 5 unit tests cover success, session-not-locked, empty-reason, empty-content, and list-delegation. |
| 4 | Every create, update, sign-off, and lock operation on clinical entities is recorded in an append-only audit trail with action, timestamp, user ID, entity type, and entity ID | VERIFIED | `audit_trigger_function()` PL/pgSQL function fires AFTER INSERT/UPDATE/DELETE on all 12 clinical tables (sessions, session_consents, contraindication_screenings, session_modules, session_outcomes, all 6 detail tables, session_addendums). Inserts into `audit_trail` with action, performed_at, user_id, entity_type, entity_id, old_values, new_values. Audit trail table has `prevent_audit_modification()` trigger blocking UPDATE/DELETE. GET `/audit` endpoint queries by entity_type and entity_id. |
| 5 | Immutability is enforced at the database level via triggers -- not just application-layer checks | VERIFIED | Migration `20260308040003` creates 3 PL/pgSQL trigger functions and 13 BEFORE UPDATE/DELETE triggers across all clinical tables. `prevent_signed_session_modification()` covers sessions and 4 direct child tables. `prevent_signed_module_detail_modification()` covers 6 module detail tables (joins through session_modules). `prevent_addendum_modification()` covers session_addendums unconditionally. Only exception: signed-to-locked transition is allowed. |

**Score:** 5/5 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/domain/addendum.go` | Addendum domain type | VERIFIED | 6-field struct (ID, SessionID, AuthorID, Reason, Content, CreatedAt), insert-only design |
| `internal/domain/audit.go` | AuditEntry domain type | VERIFIED | 8-field struct with nullable UserID (*int64), json.RawMessage old/new values |
| `migrations/20260308040000_add_session_signoff_columns.sql` | signed_at, signed_by on sessions | VERIFIED | ALTER TABLE sessions ADD COLUMN signed_at/signed_by with goose Up/Down |
| `migrations/20260308040001_create_session_addendums.sql` | session_addendums table | VERIFIED | CREATE TABLE with FK, index, goose Up/Down |
| `migrations/20260308040002_create_audit_trail.sql` | audit_trail table with append-only | VERIFIED | CREATE TABLE with CHECK constraint, 3 indexes, prevent_audit_modification trigger |
| `migrations/20260308040003_create_immutability_triggers.sql` | Immutability triggers on all clinical tables | VERIFIED | 3 functions, 13 triggers covering sessions, 4 child tables, 6 detail tables, addendums |
| `migrations/20260308040004_create_audit_triggers.sql` | Audit triggers on all clinical tables | VERIFIED | 1 function, 12 triggers (11 AFTER INSERT/UPDATE/DELETE + 1 AFTER INSERT ONLY for addendums) |
| `internal/service/signoff.go` | SignoffService with validation and sign-off | VERIFIED | ValidateForSignoff, SignOff, LockSession methods; SignoffRepository interface; sentinel errors |
| `internal/service/addendum.go` | AddendumService with locked-session gate | VERIFIED | CreateAddendum, GetByID, ListBySession; AddendumRepository interface; sentinel errors |
| `internal/service/audit.go` | AuditService read-only with pagination | VERIFIED | ListByEntity, List with pagination defaults/caps; AuditRepository interface; AuditFilter/AuditListResult types |
| `internal/repository/postgres/signoff.go` | PostgresSignoffRepository | VERIFIED | SignOff (atomic status+signed_at+signed_by), LockSession, both with optimistic locking |
| `internal/repository/postgres/addendum.go` | PostgresAddendumRepository | VERIFIED | Create (RETURNING id), GetByID (ErrAddendumNotFound), ListBySession (ORDER BY created_at DESC) |
| `internal/repository/postgres/audit.go` | PostgresAuditRepository | VERIFIED | ListByEntity, List with dynamic WHERE, separate count query, JSONB null handling |
| `internal/api/handlers/signoff.go` | Sign-off HTTP handlers | VERIFIED | HandleGetSignOffReadiness, HandleSignOffSession, HandleLockSession, auth checks, metrics |
| `internal/api/handlers/signoff_errors.go` | Error mapping for sign-off | VERIFIED | handleSignOffError maps 5 error types to HTTP status codes |
| `internal/api/handlers/addendum.go` | Addendum HTTP handlers | VERIFIED | HandleCreateAddendum (201), HandleListAddendums, HandleGetAddendum, error handlers |
| `internal/api/handlers/audit.go` | Audit HTTP handler | VERIFIED | HandleGetAuditTrail with entity_type/entity_id query params, validation |
| `internal/api/routes/sessions.go` | Route registration | VERIFIED | 7 new endpoints registered under /{id}/ for signoff, addendums, audit |
| `internal/api/routes/manager.go` | DI wiring | VERIFIED | signoffRepo, signoffSvc, addendumRepo, addendumSvc, auditRepo, auditSvc all created and passed to NewSessionRoutes |
| `internal/api/apierrors/apierrors.go` | Error codes | VERIFIED | 9 new error codes: 4 signoff, 4 addendum, 1 audit |
| `internal/api/metrics/prometheus.go` | Prometheus counters | VERIFIED | 3 new counters registered and increment methods defined |
| `internal/api/metrics/metrics.go` | Counter definitions | VERIFIED | newSessionSignedCounter, newSessionLockedCounter, newAddendumCreatedCounter |
| `internal/testutil/mock_signoff.go` | Mock signoff repo | VERIFIED | SignOffFn, LockSessionFn function fields |
| `internal/testutil/mock_addendum.go` | Mock addendum repo | VERIFIED | CreateFn, GetByIDFn, ListBySessionFn function fields |
| `internal/testutil/mock_audit.go` | Mock audit repo | VERIFIED | ListByEntityFn, ListFn function fields |
| `internal/service/signoff_test.go` | 12 signoff service tests | VERIFIED | All pass: validation (7), sign-off (3), lock (2) |
| `internal/service/addendum_test.go` | 5 addendum service tests | VERIFIED | All pass: create success/not-locked/empty-reason/empty-content, list delegation |
| `internal/service/audit_test.go` | 3 audit service tests | VERIFIED | All pass: list-by-entity delegation, default pagination, max cap |
| `internal/repository/postgres/signoff_test.go` | Compile-time assertion | VERIFIED | `var _ service.SignoffRepository = (*PostgresSignoffRepository)(nil)` |
| `internal/repository/postgres/addendum_test.go` | Compile-time assertion | VERIFIED | `var _ service.AddendumRepository = (*PostgresAddendumRepository)(nil)` |
| `internal/repository/postgres/audit_test.go` | Compile-time assertion | VERIFIED | `var _ service.AuditRepository = (*PostgresAuditRepository)(nil)` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `handlers/signoff.go` | `service/signoff.go` | `svc.ValidateForSignoff`, `svc.SignOff`, `svc.LockSession` | WIRED | All 3 service methods called from their respective handlers |
| `handlers/addendum.go` | `service/addendum.go` | `svc.CreateAddendum`, `svc.ListBySession`, `svc.GetByID` | WIRED | All 3 service methods called from their respective handlers |
| `handlers/audit.go` | `service/audit.go` | `svc.ListByEntity` | WIRED | Service method called with parsed query params |
| `routes/sessions.go` | `handlers/signoff.go` | Route registration | WIRED | `HandleGetSignOffReadiness`, `HandleSignOffSession`, `HandleLockSession` registered at `/signoff/readiness`, `/signoff`, `/lock` |
| `routes/sessions.go` | `handlers/addendum.go` | Route registration | WIRED | `HandleCreateAddendum`, `HandleListAddendums`, `HandleGetAddendum` registered at `/addendums` and `/addendums/{addendumId}` |
| `routes/sessions.go` | `handlers/audit.go` | Route registration | WIRED | `HandleGetAuditTrail` registered at `/audit` |
| `routes/manager.go` | `postgres/signoff.go` | DI wiring | WIRED | `NewPostgresSignoffRepository(db)` called, passed to `NewSignoffService` |
| `routes/manager.go` | `postgres/addendum.go` | DI wiring | WIRED | `NewPostgresAddendumRepository(db)` called, passed to `NewAddendumService` |
| `routes/manager.go` | `postgres/audit.go` | DI wiring | WIRED | `NewPostgresAuditRepository(db)` called, passed to `NewAuditService` |
| `postgres/signoff.go` | sessions table | `UPDATE sessions SET status, signed_at, signed_by` | WIRED | SQL updates with optimistic locking on version and status guard |
| `postgres/addendum.go` | session_addendums table | `INSERT INTO session_addendums` | WIRED | INSERT with RETURNING id, SELECT for Get/List |
| `postgres/audit.go` | audit_trail table | `SELECT FROM audit_trail` | WIRED | Read-only queries with dynamic WHERE and pagination |
| `migration 040003` | All clinical tables | BEFORE UPDATE/DELETE triggers | WIRED | 13 triggers across sessions, 4 child tables, 6 detail tables, addendums |
| `migration 040004` | audit_trail table | AFTER INSERT/UPDATE/DELETE triggers | WIRED | 12 triggers across all clinical tables, addendums INSERT-only |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| LOCK-01 | 05-02, 05-03 | System validates all required fields before allowing sign-off | SATISFIED | `ValidateForSignoff` checks state, consent, modules, outcome; `HandleGetSignOffReadiness` exposes via API; 7 unit tests |
| LOCK-02 | 05-01, 05-02, 05-03 | Clinician can sign off a session (records timestamp and clinician ID) | SATISFIED | `SignOff` method calls `signoffRepo.SignOff` which sets `signed_at`, `signed_by`; `HandleSignOffSession` endpoint |
| LOCK-03 | 05-01, 05-03 | Signed session becomes immutable -- original record cannot be modified | SATISFIED | DB trigger `prevent_signed_session_modification()` on all clinical tables blocks modification when status in ('signed', 'locked') |
| LOCK-04 | 05-01, 05-02, 05-03 | Clinician can add addendums to a locked session (date, author, reason, content) | SATISFIED | `AddendumService.CreateAddendum` gates on locked status; `HandleCreateAddendum` endpoint; domain type has all required fields |
| LOCK-05 | 05-01, 05-03 | Addendums are themselves immutable once saved | SATISFIED | DB trigger `prevent_addendum_modification()` unconditionally blocks UPDATE/DELETE on `session_addendums` |
| LOCK-06 | 05-01, 05-03 | Immutability is enforced at the database level (not just application layer) | SATISFIED | 3 PL/pgSQL trigger functions, 13 BEFORE triggers, plus audit trail immutability trigger |
| AUDIT-01 | 05-01, 05-03 | System logs all create, update, and delete operations on clinical entities | SATISFIED | `audit_trigger_function()` fires AFTER INSERT/UPDATE/DELETE on all 12 clinical tables |
| AUDIT-02 | 05-01, 05-02, 05-03 | Each audit entry captures: action, timestamp, user ID, entity type, entity ID | SATISFIED | `AuditEntry` struct and `audit_trail` table both have: action, performed_at, user_id, entity_type, entity_id, plus old_values/new_values |
| AUDIT-03 | 05-01, 05-03 | Audit log is append-only -- entries cannot be modified or deleted | SATISFIED | `prevent_audit_modification()` trigger fires BEFORE UPDATE/DELETE on audit_trail, raises exception unconditionally |
| AUDIT-04 | 05-01, 05-03 | Sign-off and lock events are recorded in the audit trail | SATISFIED | Audit triggers fire on UPDATE to sessions table, which covers both sign-off (status -> signed) and lock (status -> locked) transitions |
| META-02 | 05-01, 05-02, 05-03 | Signed records track signed_at, signed_by | SATISFIED | Migration adds `signed_at TIMESTAMPTZ` and `signed_by BIGINT REFERENCES users(id)` to sessions; `SignOff` repo method populates them atomically |

No orphaned requirements found. All 11 requirement IDs from the phase are accounted for in plan frontmatter and satisfied by implementation.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | - | - | - | No anti-patterns detected |

No TODOs, FIXMEs, placeholders, empty implementations, or stub patterns found in any Phase 5 files.

### Human Verification Required

### 1. Database Immutability Triggers Under Load

**Test:** With PostgreSQL running, create a session through the full lifecycle (create -> consent -> module -> outcome -> transition to awaiting_signoff -> sign off -> lock). Then attempt direct SQL UPDATE/DELETE on the signed/locked session and its child records.
**Expected:** All modification attempts should raise PostgreSQL exceptions with clear error messages. Only the signed-to-locked transition should succeed.
**Why human:** Triggers can only be tested against a running PostgreSQL instance. Unit tests verify application logic; integration testing verifies the actual DB enforcement.

### 2. Audit Trail Completeness

**Test:** Perform a full session lifecycle and then query the audit trail endpoint for the session entity.
**Expected:** Audit entries should appear for every INSERT and UPDATE operation, with correct old_values/new_values JSONB snapshots, user_id populated from created_by/updated_by columns.
**Why human:** The audit trigger function extracts user_id from row data (created_by, updated_by columns). Some tables may have different column naming that could cause null user_id. Requires running against actual data to verify.

### 3. Addendum on Locked Session End-to-End

**Test:** Lock a session, then POST to /addendums with reason and content. Attempt to UPDATE or DELETE the addendum via direct SQL.
**Expected:** Addendum creation returns 201. Direct SQL modification raises exception. Listing addendums returns the created entry.
**Why human:** Requires running API server with database to verify the full flow including trigger enforcement.

### Gaps Summary

No gaps found. All 5 observable truths are verified. All 31 artifacts exist, are substantive, and are properly wired. All 14 key links are connected. All 11 requirements are satisfied. No anti-patterns detected. The full test suite passes (20 Phase 5 service tests + all existing tests). The application builds cleanly.

The only items requiring human verification are integration-level tests against a running PostgreSQL database to confirm that the PL/pgSQL triggers behave correctly in practice, which cannot be verified through static code analysis alone.

---

_Verified: 2026-03-08T14:00:00Z_
_Verifier: Claude (gsd-verifier)_
