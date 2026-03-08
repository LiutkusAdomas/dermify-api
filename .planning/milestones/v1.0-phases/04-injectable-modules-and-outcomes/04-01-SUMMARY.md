---
phase: 04-injectable-modules-and-outcomes
plan: 01
subsystem: api
tags: [go, postgres, filler, botulinum, outcome, jsonb, service-layer]

# Dependency graph
requires:
  - phase: 03-energy-based-modules
    provides: "Energy module service pattern (validateDeviceForModule, Create/Get/Update), mock pattern"
  - phase: 02-session-management
    provides: "SessionService.AddModule, ConsentService singleton pattern, session status lifecycle"
  - phase: 01-foundation
    provides: "RegistryService.GetProductByID, Product domain type, migration pattern"
provides:
  - "FillerModuleDetail and BotulinumModuleDetail domain types"
  - "SessionOutcome domain type with status constants"
  - "SQL migrations for filler_module_details, botulinum_module_details, session_outcomes, session_outcome_endpoints"
  - "InjectableModuleService with product validation and injection site validation"
  - "OutcomeService with singleton pattern, session status guard, aftercare/red-flags coupling"
  - "Mock repositories for filler, botulinum, and outcome"
affects: [04-02, 04-03, 05-signoff-and-locking]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Product validation (validateProductForModule) analogous to device validation in energy modules"
    - "Injection site JSONB validation with typed Go struct unmarshalling"
    - "Outcome singleton pattern following consent service design"
    - "Aftercare/red-flags mandatory coupling (OUT-04)"
    - "Session outcome endpoints junction table for clinical endpoint associations"

key-files:
  created:
    - internal/domain/filler_module.go
    - internal/domain/botulinum_module.go
    - internal/domain/outcome.go
    - migrations/20260308030000_create_filler_module_details.sql
    - migrations/20260308030001_create_botulinum_module_details.sql
    - migrations/20260308030002_create_session_outcomes.sql
    - internal/service/injectable_module.go
    - internal/service/outcome.go
    - internal/testutil/mock_injectable_module.go
    - internal/testutil/mock_outcome.go
  modified: []

key-decisions:
  - "ProductID is required (int64, not pointer) matching DeviceID pattern from Phase 3"
  - "InjectionSites stored as json.RawMessage for flexible JSONB mapping"
  - "Outcome validation enforces aftercare-to-red-flags coupling per OUT-04"
  - "Session status guard allows outcomes only in in_progress or awaiting_signoff"
  - "validateInjectionSites is a package-level function (not method) for testability"

patterns-established:
  - "validateProductForModule: product type validation analogous to validateDeviceForModule"
  - "Junction table pattern for outcome-to-endpoint associations (same as session-to-indication-codes)"

requirements-completed: [FILL-01, FILL-02, FILL-03, TOX-01, TOX-02, TOX-03, OUT-01, OUT-02, OUT-03, OUT-04, OUT-05]

# Metrics
duration: 2min
completed: 2026-03-08
---

# Phase 4 Plan 1: Injectable Modules and Outcomes Domain/Service Layer Summary

**Filler and botulinum domain types with product validation services, session outcome singleton with aftercare/red-flags coupling, and three SQL migrations**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-08T10:14:27Z
- **Completed:** 2026-03-08T10:16:47Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Domain types for filler modules, botulinum modules, and session outcomes with all clinical fields
- Three SQL migrations with FK constraints, indexes, CHECK constraint on outcome status, and junction table
- InjectableModuleService with product type validation and injection site JSONB validation
- OutcomeService following consent singleton pattern with session status guard and aftercare/red-flags coupling
- Mock repositories for all three new domain types ready for unit testing

## Task Commits

Each task was committed atomically:

1. **Task 1: Domain types, SQL migrations** - `69fee6a` (feat)
2. **Task 2: Service scaffolds, repository interfaces, sentinel errors, and mock repositories** - `ce801e1` (feat)

## Files Created/Modified
- `internal/domain/filler_module.go` - FillerModuleDetail struct with all clinical fields as pointers
- `internal/domain/botulinum_module.go` - InjectionSite and BotulinumModuleDetail with json.RawMessage
- `internal/domain/outcome.go` - SessionOutcome with status constants and EndpointIDs
- `migrations/20260308030000_create_filler_module_details.sql` - Filler detail table with FK to session_modules and products
- `migrations/20260308030001_create_botulinum_module_details.sql` - Botulinum detail table with JSONB injection_sites
- `migrations/20260308030002_create_session_outcomes.sql` - Session outcomes + endpoints junction table
- `internal/service/injectable_module.go` - InjectableModuleService with product and injection site validation
- `internal/service/outcome.go` - OutcomeService with singleton pattern and session status guard
- `internal/testutil/mock_injectable_module.go` - MockFillerModuleRepository and MockBotulinumModuleRepository
- `internal/testutil/mock_outcome.go` - MockOutcomeRepository with all 6 interface methods

## Decisions Made
- ProductID is required (int64, not pointer) matching DeviceID pattern from Phase 3
- InjectionSites stored as json.RawMessage for flexible JSONB mapping
- Outcome validation enforces aftercare-to-red-flags coupling per OUT-04
- Session status guard allows outcomes only in in_progress or awaiting_signoff
- validateInjectionSites is a package-level function (not method) for testability

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All domain types, migrations, services, and mocks are ready for Plan 02 (repositories + tests)
- Plan 03 (HTTP handlers + wiring) can build against the full contract surface
- All 10 planned files exist and compile cleanly
- Existing tests continue to pass

## Self-Check: PASSED

All 10 files verified present. Both task commits (69fee6a, ce801e1) verified in git log.

---
*Phase: 04-injectable-modules-and-outcomes*
*Completed: 2026-03-08*
