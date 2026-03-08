---
phase: 05-sign-off-and-compliance
plan: 01
subsystem: database, api
tags: [postgres, triggers, plpgsql, audit-trail, immutability, sign-off, addendum]

# Dependency graph
requires:
  - phase: 04-injectable-modules-and-outcomes
    provides: "All clinical tables (sessions, consents, screenings, modules, outcomes, detail tables)"
provides:
  - "Addendum and AuditEntry domain types"
  - "5 SQL migrations: signed_at/signed_by columns, session_addendums table, audit_trail table, immutability triggers, audit triggers"
  - "SignoffService, AddendumService, AuditService interfaces and sentinel errors"
  - "SignoffRepository interface with SignOff and LockSession contracts"
  - "MockAddendumRepository, MockAuditRepository, MockSignoffRepository test doubles"
affects: [05-02, 05-03]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "PL/pgSQL BEFORE triggers for immutability enforcement on signed/locked sessions"
    - "PL/pgSQL AFTER triggers for automatic audit trail capture via row_to_json"
    - "Separate SignoffRepository interface for atomic sign-off (status + signed_at + signed_by)"
    - "Read-only AuditService (audit entries created by DB triggers, not application code)"

key-files:
  created:
    - internal/domain/addendum.go
    - internal/domain/audit.go
    - migrations/20260308040000_add_session_signoff_columns.sql
    - migrations/20260308040001_create_session_addendums.sql
    - migrations/20260308040002_create_audit_trail.sql
    - migrations/20260308040003_create_immutability_triggers.sql
    - migrations/20260308040004_create_audit_triggers.sql
    - internal/service/signoff.go
    - internal/service/addendum.go
    - internal/service/audit.go
    - internal/testutil/mock_addendum.go
    - internal/testutil/mock_audit.go
    - internal/testutil/mock_signoff.go
  modified: []

key-decisions:
  - "SignoffRepository as separate interface from SessionRepository for atomic signed_at/signed_by update"
  - "AuditService is read-only; audit entries created exclusively by DB triggers"
  - "Mock files follow existing testutil pattern (no //go:build ignore tags) for consistency"
  - "Audit pagination uses auditDefaultPerPage=50 and auditMaxPerPage=100 as separate constants to avoid collisions"

patterns-established:
  - "Immutability trigger pattern: BEFORE UPDATE/DELETE with status check, special-case signed->locked transition"
  - "Audit trigger pattern: AFTER INSERT/UPDATE/DELETE using row_to_json for old/new values"
  - "Append-only table pattern: BEFORE UPDATE/DELETE trigger that unconditionally raises exception"

requirements-completed: [LOCK-02, LOCK-03, LOCK-04, LOCK-05, LOCK-06, AUDIT-01, AUDIT-02, AUDIT-03, AUDIT-04, META-02]

# Metrics
duration: 2min
completed: 2026-03-08
---

# Phase 5 Plan 1: Domain Types, Migrations, and Service Interfaces Summary

**Addendum and AuditEntry domain types, 5 SQL migrations (schema + immutability/audit triggers), 3 service interfaces with sentinel errors, and 3 mock repositories for sign-off compliance**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-08T13:14:36Z
- **Completed:** 2026-03-08T13:17:23Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- Addendum domain type (6 fields, insert-only) and AuditEntry domain type (8 fields with nullable UserID and json.RawMessage old/new values)
- 5 SQL migrations establishing complete sign-off schema: signed_at/signed_by columns, session_addendums table, append-only audit_trail table, immutability triggers on all 12 clinical tables, audit triggers on all 12 clinical tables
- SignoffService with ValidateForSignoff (checks consent, modules, outcome), SignOff, and LockSession methods
- AddendumService with CreateAddendum (validates locked session status), GetByID, ListBySession methods
- AuditService (read-only) with ListByEntity and paginated List methods
- 3 mock repositories (addendum, audit, signoff) for service testing

## Task Commits

Each task was committed atomically:

1. **Task 1: Domain types and SQL migrations** - `11e5e02` (feat)
2. **Task 2: Service interfaces, sentinel errors, and mock repositories** - `ff32a0b` (feat)

## Files Created/Modified
- `internal/domain/addendum.go` - Addendum domain type (insert-only, 6 fields)
- `internal/domain/audit.go` - AuditEntry domain type with json.RawMessage old/new values
- `migrations/20260308040000_add_session_signoff_columns.sql` - signed_at/signed_by columns on sessions
- `migrations/20260308040001_create_session_addendums.sql` - session_addendums table with FK and index
- `migrations/20260308040002_create_audit_trail.sql` - audit_trail table with append-only trigger
- `migrations/20260308040003_create_immutability_triggers.sql` - BEFORE UPDATE/DELETE triggers on all clinical tables
- `migrations/20260308040004_create_audit_triggers.sql` - AFTER INSERT/UPDATE/DELETE triggers on all clinical tables
- `internal/service/signoff.go` - SignoffService, SignoffRepository interface, ValidationResult, sentinel errors
- `internal/service/addendum.go` - AddendumService, AddendumRepository interface, sentinel errors
- `internal/service/audit.go` - AuditService, AuditRepository interface, AuditFilter, AuditListResult
- `internal/testutil/mock_addendum.go` - MockAddendumRepository test double
- `internal/testutil/mock_audit.go` - MockAuditRepository test double
- `internal/testutil/mock_signoff.go` - MockSignoffRepository test double

## Decisions Made
- **SignoffRepository as separate interface:** The existing SessionRepository.UpdateStatus does not set signed_at/signed_by. A dedicated SignoffRepository with SignOff and LockSession methods keeps the sign-off contract focused and the existing SessionRepository unchanged.
- **Read-only AuditService:** Audit entries are created exclusively by PostgreSQL AFTER triggers, so the service layer has no Create method. The service provides query/filter capabilities only.
- **Mock files without build tags:** Existing mock files in testutil do not use `//go:build ignore` tags, so new mocks follow the same pattern for consistency.
- **Separate audit pagination constants:** Used `auditDefaultPerPage` and `auditMaxPerPage` to avoid redeclaring the `maxPerPage` constant already defined in patient.go within the same package.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Domain types and service interfaces ready for Plan 02 (repository implementations and service tests)
- All SQL migrations ready for DB deployment
- Mock repositories ready for unit testing in Plan 02
- Immutability and audit triggers will activate once migrations run against PostgreSQL

## Self-Check: PASSED

All 13 created files verified present. Both task commits (11e5e02, ff32a0b) verified in git log.

---
*Phase: 05-sign-off-and-compliance*
*Completed: 2026-03-08*
