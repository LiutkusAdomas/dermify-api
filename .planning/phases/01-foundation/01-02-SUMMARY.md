---
phase: 01-foundation
plan: 02
subsystem: api
tags: [go, rbac, jwt, middleware, chi, postgres, migration]

# Dependency graph
requires:
  - phase: 01-foundation-01
    provides: Domain model types (Role constants), RoleService, RoleRepository interface, PostgresRoleRepository
provides:
  - Role migration adding role column to users table with CHECK constraint
  - RequireRole middleware gating endpoints by role
  - Role assignment endpoint (POST /api/v1/roles/assign, Admin-only)
  - First-user auto-promotion to Admin on registration
  - JWT claims carrying role field with backward-compatible omitempty tag
  - Role in login, refresh, and profile responses
  - role_assignment_total Prometheus metric
affects: [01-foundation-03, 01-foundation-04]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "RequireRole middleware chained after RequireAuth for RBAC enforcement"
    - "First-user bootstrap via IsFirstUser check after INSERT in registration handler"
    - "Role carried in JWT claims with omitempty for backward compatibility"

key-files:
  created:
    - migrations/20260307130000_add_role_to_users.sql
    - internal/api/handlers/roles.go
    - internal/api/routes/roles.go
  modified:
    - internal/api/auth/auth.go
    - internal/api/middleware/auth.go
    - internal/api/middleware/auth_test.go
    - internal/api/apierrors/apierrors.go
    - internal/api/handlers/auth.go
    - internal/api/handlers/login.go
    - internal/api/routes/auth.go
    - internal/api/routes/manager.go
    - internal/api/metrics/metrics.go
    - internal/api/metrics/prometheus.go

key-decisions:
  - "Role field uses json omitempty tag for backward compatibility with pre-migration tokens"
  - "Login and refresh handlers updated in Task 1 (Rule 3 blocking fix) since GenerateAccessToken signature change required all callers to pass role"
  - "First-user bootstrap uses RoleService.IsFirstUser (CountUsers <= 1) after INSERT, non-fatal on error"

patterns-established:
  - "RequireRole middleware: variadic allowed roles, returns 401 if no claims, 403 if role not in allowed set"
  - "Handler closure receiving service struct (RoleService) instead of raw *sql.DB for new handlers"
  - "Route groups with RequireAuth + RequireRole middleware chain for protected endpoints"

requirements-completed: [RBAC-01, RBAC-02, RBAC-03, RBAC-04]

# Metrics
duration: 5min
completed: 2026-03-07
---

# Phase 1 Plan 02: RBAC System Summary

**RequireRole middleware with JWT role claims, admin-only role assignment endpoint, and first-user auto-admin bootstrap**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-07T18:46:25Z
- **Completed:** 2026-03-07T18:52:05Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- RequireRole middleware correctly gates endpoints by role (4 passing unit tests)
- JWT claims carry role with backward-compatible omitempty tag
- POST /api/v1/roles/assign endpoint restricted to Admin users via RequireAuth + RequireRole chain
- First registered user is auto-promoted to Admin via RoleService.IsFirstUser
- Login and refresh handlers include user role in generated tokens
- Profile endpoint returns role field

## Task Commits

Each task was committed atomically:

1. **Task 1: Create RBAC migration, update JWT claims, and build RequireRole middleware** - `fe29445` (feat)
2. **Task 2: Build role handlers, routes, first-user bootstrap, and wire to app** - `3910a7a` (feat)

## Files Created/Modified
- `migrations/20260307130000_add_role_to_users.sql` - Role column migration with CHECK constraint
- `internal/api/auth/auth.go` - Added Role field to Claims, updated GenerateAccessToken signature
- `internal/api/middleware/auth.go` - Added RequireRole middleware function
- `internal/api/middleware/auth_test.go` - Four RequireRole unit tests (removed build ignore tag)
- `internal/api/apierrors/apierrors.go` - Added AuthInsufficientRole, RoleInvalidRole, RoleAssignmentFailed, RoleUserNotFound error codes
- `internal/api/handlers/roles.go` - HandleAssignRole handler with validation and error handling
- `internal/api/handlers/auth.go` - HandleRegister with first-user bootstrap, HandleGetProfile with role
- `internal/api/handlers/login.go` - Role query and role-aware token generation
- `internal/api/routes/roles.go` - RoleRoutes with RequireAuth + RequireRole(admin) middleware
- `internal/api/routes/auth.go` - Added RoleService dependency for HandleRegister
- `internal/api/routes/manager.go` - Wired PostgresRoleRepository, RoleService, RoleRoutes
- `internal/api/metrics/metrics.go` - Added newRoleAssignmentCounter
- `internal/api/metrics/prometheus.go` - Registered role_assignment_total metric and IncrementRoleAssignmentCount method

## Decisions Made
- Used `json:"role,omitempty"` on Claims.Role to ensure backward compatibility with existing tokens that have no role field. Existing tokens parse successfully with empty string role, and RequireRole treats empty role as "no role" (403).
- Updated login and refresh handler callers of GenerateAccessToken in Task 1 as a Rule 3 blocking fix, since the signature change broke compilation. The plan allocated this to Task 2, but the build had to pass for Task 1 verification.
- First-user bootstrap is non-fatal on error (logs error, continues registration) to avoid blocking user creation if the role check fails.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated GenerateAccessToken callers in Task 1**
- **Found during:** Task 1 (JWT Claims update)
- **Issue:** Changing GenerateAccessToken signature to accept role parameter broke compilation of HandleLogin and HandleRefreshToken, which were planned for Task 2 updates.
- **Fix:** Updated both callers in Task 1 to pass the role parameter (querying role from database), making them fully role-aware ahead of schedule.
- **Files modified:** internal/api/handlers/login.go, internal/api/handlers/auth.go
- **Verification:** `go build ./internal/api/...` passes
- **Committed in:** fe29445 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 3 blocking)
**Impact on plan:** No scope creep. The work was planned for Task 2 but had to be done in Task 1 due to compilation dependency. Task 2 proceeded without needing to re-do these changes.

## Issues Encountered
- golangci-lint not installed in execution environment. Used `go vet` as fallback. Code follows all known lint rules (no globals except sentinel errors with nolint, comments end with period, functions under 100 lines, no naked returns).

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- RBAC system complete with RequireRole middleware ready for use by Plans 01-03 (Patient CRUD) and 01-04 (Registry)
- Both Doctor and Admin roles structurally supported; full differentiation exercised when clinical-only endpoints arrive in Phase 2+
- Route Manager wires RoleService automatically from database connection
- role_assignment_total metric tracks admin role operations

## Self-Check: PASSED

All created files verified present. Both commit hashes (fe29445, 3910a7a) confirmed in git log.

---
*Phase: 01-foundation*
*Completed: 2026-03-07*
