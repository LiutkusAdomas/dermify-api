---
phase: 05-sign-off-and-compliance
plan: 02
subsystem: database, api
tags: [postgres, repository, signoff, addendum, audit-trail, optimistic-locking, unit-tests]

# Dependency graph
requires:
  - phase: 05-sign-off-and-compliance
    provides: "SignoffService, AddendumService, AuditService interfaces, domain types, and mock repositories"
provides:
  - "PostgresSignoffRepository with SignOff and LockSession using optimistic locking"
  - "PostgresAddendumRepository with Create, GetByID, ListBySession (insert-only + read)"
  - "PostgresAuditRepository with ListByEntity and paginated List with dynamic filters"
  - "20 service-layer unit tests covering sign-off validation, addendum creation, and audit pagination"
affects: [05-03]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Separate count query + data query for audit pagination (consistent with session List pattern)"
    - "Dynamic WHERE clause building with argIndex tracking for audit filter"
    - "JSONB null handling via *[]byte intermediate for audit old_values/new_values"

key-files:
  created:
    - internal/repository/postgres/signoff.go
    - internal/repository/postgres/signoff_test.go
    - internal/repository/postgres/addendum.go
    - internal/repository/postgres/addendum_test.go
    - internal/repository/postgres/audit.go
    - internal/repository/postgres/audit_test.go
    - internal/service/signoff_test.go
    - internal/service/addendum_test.go
    - internal/service/audit_test.go
  modified: []

key-decisions:
  - "Separate count query for audit pagination (not COUNT(*) OVER()) matching existing session List pattern for consistency"
  - "signoffTestDeps helper struct with setupAllComplete for DRY test setup across validation and sign-off tests"

patterns-established:
  - "signoffTestDeps pattern: helper struct with setupAwaitingSignoffSession/setupAllComplete for signoff test DRY"
  - "addendumTestDeps pattern: helper struct with setupLockedSession for addendum test DRY"

requirements-completed: [LOCK-01, LOCK-02, LOCK-04, AUDIT-02, META-02]

# Metrics
duration: 3min
completed: 2026-03-08
---

# Phase 5 Plan 2: Postgres Repositories and Service Unit Tests Summary

**Postgres repositories for signoff/addendum/audit with optimistic locking and 20 service-layer unit tests covering validation, state guards, and pagination**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-08T13:19:54Z
- **Completed:** 2026-03-08T13:23:36Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- PostgresSignoffRepository with SignOff (sets signed_at/signed_by atomically) and LockSession, both using optimistic locking via version check and status guard
- PostgresAddendumRepository with insert-only Create (RETURNING id), GetByID (ErrAddendumNotFound mapping), and ListBySession (ordered by created_at DESC)
- PostgresAuditRepository with ListByEntity (read-only, performed_at DESC) and paginated List with dynamic WHERE clause building for entity_type/entity_id/user_id filters
- 20 service unit tests: 12 for SignoffService (validation completeness, wrong state, sign-off success/failure, lock success/wrong-state), 5 for AddendumService (locked-session gate, empty field validation, delegation), 3 for AuditService (delegation, pagination defaults, max cap)

## Task Commits

Each task was committed atomically:

1. **Task 1: Postgres repositories for signoff, addendum, and audit** - `8fce714` (feat)
2. **Task 2: Service unit tests for signoff, addendum, and audit** - `3108b8c` (test)

## Files Created/Modified
- `internal/repository/postgres/signoff.go` - PostgresSignoffRepository with SignOff and LockSession methods
- `internal/repository/postgres/signoff_test.go` - Compile-time interface assertion for SignoffRepository
- `internal/repository/postgres/addendum.go` - PostgresAddendumRepository with Create, GetByID, ListBySession
- `internal/repository/postgres/addendum_test.go` - Compile-time interface assertion for AddendumRepository
- `internal/repository/postgres/audit.go` - PostgresAuditRepository with ListByEntity and paginated List
- `internal/repository/postgres/audit_test.go` - Compile-time interface assertion for AuditRepository
- `internal/service/signoff_test.go` - 12 unit tests for SignoffService validation and execution
- `internal/service/addendum_test.go` - 5 unit tests for AddendumService creation and validation
- `internal/service/audit_test.go` - 3 unit tests for AuditService pagination behavior

## Decisions Made
- **Separate count query for audit pagination:** Used separate COUNT(*) + data query pattern (consistent with existing PostgresSessionRepository.List) rather than COUNT(*) OVER() window function, keeping the codebase consistent.
- **signoffTestDeps helper struct:** Created setupAllComplete helper that wires up session, consent, module, and outcome mocks for DRY setup across validation and sign-off tests, following the established outcomeTestDeps pattern.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All repository implementations complete, ready for Plan 03 (HTTP handler layer)
- Service tests provide confidence in business logic correctness
- Compile-time interface assertions verify all method signatures match

## Self-Check: PASSED

All 9 created files verified present. Both task commits (8fce714, 3108b8c) verified in git log.

---
*Phase: 05-sign-off-and-compliance*
*Completed: 2026-03-08*
