---
phase: 03-energy-based-modules
plan: 01
subsystem: api
tags: [go, domain-types, sql-migrations, repository-interfaces, service-scaffold]

# Dependency graph
requires:
  - phase: 02-session-lifecycle
    provides: "SessionModule base table, SessionService.AddModule, RegistryService.GetDeviceByID"
provides:
  - "IPLModuleDetail, NdYAGModuleDetail, CO2ModuleDetail, RFModuleDetail domain types"
  - "Four detail table migrations (ipl/ndyag/co2/rf_module_details) with FK constraints"
  - "Repository interfaces (IPLModuleRepository, NdYAGModuleRepository, CO2ModuleRepository, RFModuleRepository)"
  - "EnergyModuleService with Create/Get/Update x 4 types and device validation"
  - "Mock repositories for all four interfaces"
  - "Sentinel errors: ErrModuleDetailNotFound, ErrModuleDetailVersionConflict, ErrDeviceTypeMismatch, ErrHandpieceMismatch, ErrInvalidModuleData"
affects: [03-02-PLAN, 03-03-PLAN]

# Tech tracking
tech-stack:
  added: []
  patterns: ["per-type detail tables with FK to session_modules (hybrid polymorphism)", "validateDeviceForModule private helper for cross-entity validation"]

key-files:
  created:
    - internal/domain/ipl_module.go
    - internal/domain/ndyag_module.go
    - internal/domain/co2_module.go
    - internal/domain/rf_module.go
    - migrations/20260308020000_create_ipl_module_details.sql
    - migrations/20260308020001_create_ndyag_module_details.sql
    - migrations/20260308020002_create_co2_module_details.sql
    - migrations/20260308020003_create_rf_module_details.sql
    - internal/service/energy_module.go
    - internal/testutil/mock_energy_module.go
  modified: []

key-decisions:
  - "Pointer types for all nullable clinical fields; DeviceID is required (int64, not pointer)"
  - "DECIMAL(6,2) for fluence/energy, DECIMAL(8,2) for duration/rate, DECIMAL(5,2) for percentage values"
  - "validateDeviceForModule iterates device.Handpieces slice rather than separate DB query"
  - "Create methods delegate to SessionService.AddModule for consent gate and editability enforcement"

patterns-established:
  - "Per-type detail domain struct: ID, ModuleID, DeviceID, HandpieceID, clinical fields, Notes, Version, metadata"
  - "Per-type detail migration: UNIQUE FK to session_modules, FK to devices, nullable handpiece_id FK"
  - "Mock per-domain file pattern (mock_energy_module.go) with function fields for each method"

requirements-completed: [IPL-01, IPL-02, IPL-03, YAG-01, YAG-02, YAG-03, CO2-01, CO2-02, CO2-03, RF-01, RF-02, RF-03]

# Metrics
duration: 4min
completed: 2026-03-08
---

# Phase 03 Plan 01: Energy Module Contracts Summary

**Domain types, SQL migrations, repository interfaces, service scaffold with device validation, and mock repositories for IPL, NdYAG, CO2, and RF energy-based modules**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-08T09:11:20Z
- **Completed:** 2026-03-08T09:14:48Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Four domain types with full clinical parameter fields matching requirements (IPL-02, YAG-02, CO2-02, RF-02)
- Four Goose SQL migrations creating detail tables with FK constraints to session_modules and devices (IPL-03, YAG-03, CO2-03, RF-03)
- EnergyModuleService with 12 CRUD methods (Create/Get/Update x 4 types) and device validation helper (IPL-01, YAG-01, CO2-01, RF-01)
- Mock repositories ready for unit testing in Plan 02

## Task Commits

Each task was committed atomically:

1. **Task 1: Domain types, migrations, and repository interfaces** - `146bc45` (feat)
2. **Task 2: Service scaffold with repository interfaces, sentinel errors, and mock repositories** - `5ef7180` (feat)

## Files Created/Modified
- `internal/domain/ipl_module.go` - IPLModuleDetail type with filter_band, lightguide_size, fluence, pulse_duration, pulse_delay, pulse_count, passes, total_pulses, cooling_mode
- `internal/domain/ndyag_module.go` - NdYAGModuleDetail type with wavelength, spot_size, fluence, pulse_duration, repetition_rate, cooling_type, total_pulses
- `internal/domain/co2_module.go` - CO2ModuleDetail type with mode, scanner_pattern, power, pulse_energy, pulse_duration, density, pattern, passes, anaesthesia_used
- `internal/domain/rf_module.go` - RFModuleDetail type with rf_mode, tip_type, depth, energy_level, overlap, pulses_per_zone, total_pulses
- `migrations/20260308020000_create_ipl_module_details.sql` - IPL detail table with FK constraints and indexes
- `migrations/20260308020001_create_ndyag_module_details.sql` - NdYAG detail table with FK constraints and indexes
- `migrations/20260308020002_create_co2_module_details.sql` - CO2 detail table with FK constraints and indexes
- `migrations/20260308020003_create_rf_module_details.sql` - RF detail table with FK constraints and indexes
- `internal/service/energy_module.go` - EnergyModuleService with 4 repository interfaces, 5 sentinel errors, 12 CRUD methods, and device validation
- `internal/testutil/mock_energy_module.go` - Mock implementations for all four repository interfaces

## Decisions Made
- Pointer types for all nullable clinical fields; DeviceID is required (int64, not pointer)
- DECIMAL(6,2) for fluence/energy, DECIMAL(8,2) for duration/rate, DECIMAL(5,2) for percentage values in SQL
- validateDeviceForModule iterates device.Handpieces slice from RegistryService rather than separate DB query
- Create methods delegate to SessionService.AddModule for consent gate and editability enforcement

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All contracts (types, interfaces, errors) are stable for Plan 02 (repository implementations + service tests)
- Mock repositories are ready for unit testing
- Migrations are ready for database application

## Self-Check: PASSED

All 10 created files verified present. Both task commits (146bc45, 5ef7180) verified in git log.

---
*Phase: 03-energy-based-modules*
*Completed: 2026-03-08*
