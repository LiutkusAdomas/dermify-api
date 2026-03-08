---
phase: 01-foundation
plan: 03
subsystem: api
tags: [go, chi, postgres, crud, pagination, optimistic-locking, patients]

requires:
  - phase: 01-foundation/01-01
    provides: domain models (Patient, SessionSummary), service/repository pattern, RoleService
  - phase: 01-foundation/01-02
    provides: RequireAuth/RequireRole middleware, JWT claims with role, route Manager

provides:
  - Patient CRUD endpoints (create, list, get, update, session history)
  - PatientService with validation and pagination defaults
  - PatientRepository interface and PostgresPatientRepository
  - Patient search by name/phone/email with ILIKE prefix matching
  - Offset-based pagination with total count and page metadata
  - Optimistic locking via version column on patient updates
  - Metadata tracking (created_at/by, updated_at/by, version)
  - PaginatedResponse envelope for list endpoints
  - patient_created_total Prometheus counter

affects: [02-sessions, 03-energy-modules, 04-injectable-modules]

tech-stack:
  added: []
  patterns:
    - "Offset pagination with PaginatedResponse envelope (data, total, page, per_page, total_pages)"
    - "Optimistic locking via version column with 409 Conflict on mismatch"
    - "PatientListItem embedding domain.Patient with session metadata placeholders"
    - "Handler error mapping helpers (handlePatientCreateError, handlePatientLookupError, handlePatientUpdateError)"

key-files:
  created:
    - migrations/20260307140000_create_patients_table.sql
    - internal/service/patient.go
    - internal/repository/postgres/patient.go
    - internal/api/handlers/patients.go
    - internal/api/routes/patients.go
    - internal/service/patient_test.go
    - internal/testutil/mock_patient.go
  modified:
    - internal/api/apierrors/apierrors.go
    - internal/api/handlers/models.go
    - internal/api/routes/manager.go
    - internal/api/metrics/metrics.go
    - internal/api/metrics/prometheus.go
    - internal/testutil/mocks.go

key-decisions:
  - "ILIKE prefix search with LOWER() functional indexes for patient search (no pg_trgm extension needed)"
  - "Session count and last_session_date are hardcoded placeholders (0 and null) until Phase 2 sessions exist"
  - "Per-domain mock files (mock_patient.go, mock_role.go) instead of single monolithic mocks.go"
  - "Handler applies pagination defaults (page=1, per_page=20) before calling service for response accuracy"

patterns-established:
  - "PaginatedResponse: standard envelope for all paginated list endpoints"
  - "Optimistic locking: version column in WHERE clause, 409 on zero rows affected"
  - "Error mapping helpers: per-domain functions mapping service errors to HTTP status codes"
  - "parseIDParam/parseIntParam: reusable URL parameter parsing functions"

requirements-completed: [PAT-01, PAT-02, PAT-03, PAT-04, META-01, META-02, META-03]

duration: 8min
completed: 2026-03-07
---

# Phase 1 Plan 03: Patient Management Summary

**Patient CRUD with ILIKE search, offset pagination, optimistic version locking, and metadata tracking via service/repository pattern**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-07T18:55:37Z
- **Completed:** 2026-03-07T19:03:18Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- Full patient CRUD with 5 HTTP endpoints (create, list, get, update, session history)
- Patient search by name (prefix match), phone, and email with proper indexes
- Offset pagination with configurable page/per_page (defaults 1/20, max 100) and total count
- Optimistic locking via version column with 409 Conflict on concurrent modification
- Metadata tracking (created_at/by, updated_at/by, version) populated from JWT claims
- 8 passing unit tests covering validation, pagination, version conflict, and metadata

## Task Commits

Each task was committed atomically:

1. **Task 1: Patient migration, repository, service, and unit tests** - `a8c8ad4` (feat)
2. **Task 2: Patient handlers, routes, and wiring** - `97f2c8e` (feat)

_Note: Both commits were created by a concurrent Plan 01-04 executor that included Plan 01-03 files for build compatibility. Content verified identical to this plan's specifications._

## Files Created/Modified
- `migrations/20260307140000_create_patients_table.sql` - Patients table with metadata columns, CHECK constraint on sex, 4 search indexes
- `internal/service/patient.go` - PatientService with validation, pagination defaults, PatientRepository interface
- `internal/repository/postgres/patient.go` - PostgresPatientRepository with parameterized SQL, ILIKE search, optimistic locking
- `internal/service/patient_test.go` - 8 unit tests for service layer behavior
- `internal/testutil/mock_patient.go` - MockPatientRepository test double
- `internal/testutil/mocks.go` - Reduced to package declaration (mocks split into per-domain files)
- `internal/api/handlers/patients.go` - 5 handler functions with error mapping
- `internal/api/handlers/models.go` - PatientResponse and PaginatedResponse structs
- `internal/api/routes/patients.go` - PatientRoutes with RequireAuth + RequireRole(Doctor, Admin)
- `internal/api/routes/manager.go` - Added patientRoutes field and wiring
- `internal/api/apierrors/apierrors.go` - Patient error code constants
- `internal/api/metrics/metrics.go` - patient_created_total counter definition
- `internal/api/metrics/prometheus.go` - Counter registration and increment method

## Decisions Made
- Used ILIKE prefix search with LOWER() functional indexes instead of pg_trgm extension for patient name search. Simpler, no extension dependency, sufficient for prefix matching on small datasets.
- Session count and last_session_date are hardcoded as 0/null in the SQL query for Phase 1. Will be replaced with LEFT JOIN when sessions table exists in Phase 2.
- Split mock files per domain (mock_patient.go, mock_role.go) instead of a single mocks.go for independent build management.
- Handler applies pagination defaults before calling service so the response reflects actual values used.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Resolved MockRoleRepository redeclaration conflict**
- **Found during:** Task 1 (mocks.go refactor)
- **Issue:** MockRoleRepository was defined in both mocks.go (from Plan 01-00) and mock_role.go (from Plan 01-02), causing a compilation error when removing the build ignore tag
- **Fix:** Reduced mocks.go to a package declaration only; kept MockRoleRepository in mock_role.go and created MockPatientRepository in mock_patient.go
- **Files modified:** internal/testutil/mocks.go, internal/testutil/mock_patient.go
- **Verification:** `go build ./internal/testutil/` succeeds
- **Committed in:** a8c8ad4 (Task 1 commit)

**2. [Rule 1 - Bug] Fixed pagination defaults not reflected in list response**
- **Found during:** Task 2 (HandleListPatients implementation)
- **Issue:** Service applies pagination defaults by value, but handler's local filter variable retained pre-default values (0, 0), causing response to show page=0, per_page=0
- **Fix:** Applied defaults (page=1, per_page=20) in the handler's parseIntParam calls so response data is accurate
- **Files modified:** internal/api/handlers/patients.go
- **Verification:** Build succeeds, response will correctly show page=1, per_page=20 when not specified
- **Committed in:** 97f2c8e (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)
**Impact on plan:** Both fixes necessary for correctness. No scope creep.

## Issues Encountered
- Plan 01-04 was executed before Plan 01-03 (out of order), and that executor included all Plan 01-03 files in its commits for build compatibility. This plan verified the committed code matches the Plan 01-03 specifications exactly -- no re-commits were necessary.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Patient CRUD foundation complete, ready for session linking in Phase 2
- PaginatedResponse pattern established for reuse in registry and future list endpoints
- Metadata tracking pattern (created_at/by, updated_at/by, version) ready for all clinical entities

## Self-Check: PASSED

- All 13 created/modified files verified present on disk.
- Commit a8c8ad4 (Task 1) verified in git log.
- Commit 97f2c8e (Task 2) verified in git log.
- `go build ./...` passes.
- `go vet ./...` passes.
- All 8 patient service unit tests pass.

---
*Phase: 01-foundation*
*Completed: 2026-03-07*
