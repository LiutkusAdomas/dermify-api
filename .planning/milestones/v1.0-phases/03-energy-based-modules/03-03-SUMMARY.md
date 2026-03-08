---
phase: 03-energy-based-modules
plan: 03
subsystem: api
tags: [go, http-handlers, chi-routes, prometheus-metrics, dependency-injection]

# Dependency graph
requires:
  - phase: 03-energy-based-modules
    plan: 01
    provides: "Domain types, EnergyModuleService with Create/Get/Update x 4 types, sentinel errors"
  - phase: 03-energy-based-modules
    plan: 02
    provides: "Postgres repositories for IPL/NdYAG/CO2/RF detail tables"
provides:
  - "HandleCreate/Get/Update IPL/NdYAG/CO2/RF Module HTTP handlers (12 endpoints)"
  - "IPL/NdYAG/CO2/RF ModuleDetailResponse types in models.go"
  - "Energy module detail API error codes (5 new constants)"
  - "dermify_energy_module_created_total Prometheus counter"
  - "Route registration under /sessions/{id}/modules/{type} with POST/GET/PUT"
  - "Manager DI wiring for EnergyModuleService with all 4 detail repositories"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: ["Shared handleEnergyModuleError function across all 4 energy module handler files", "Per-type handler files following one-file-per-domain convention"]

key-files:
  created:
    - internal/api/handlers/ipl_module.go
    - internal/api/handlers/ndyag_module.go
    - internal/api/handlers/co2_module.go
    - internal/api/handlers/rf_module.go
    - internal/api/handlers/energy_module_errors.go
  modified:
    - internal/api/handlers/models.go
    - internal/api/apierrors/apierrors.go
    - internal/api/metrics/prometheus.go
    - internal/api/metrics/metrics.go
    - internal/api/routes/sessions.go
    - internal/api/routes/manager.go

key-decisions:
  - "Shared handleEnergyModuleError in dedicated energy_module_errors.go file for clean separation"
  - "Module ID parsed from chi.URLParam moduleId in Get/Update handlers rather than from request body"
  - "Update handlers re-fetch after update to return full updated record (consistent with session/consent pattern)"

patterns-established:
  - "Energy module handler pattern: closure over EnergyModuleService and metrics.Client, claims check, URL param parsing, request decode, service call, error handler, response write"
  - "Shared error handler pattern: one handleEnergyModuleError function covering all service sentinel errors for the energy module domain"

requirements-completed: [IPL-01, IPL-02, IPL-03, YAG-01, YAG-02, YAG-03, CO2-01, CO2-02, CO2-03, RF-01, RF-02, RF-03]

# Metrics
duration: 5min
completed: 2026-03-08
---

# Phase 03 Plan 03: Energy Module HTTP Layer Summary

**HTTP handlers, route registration, and DI wiring for 12 energy module endpoints (Create/Get/Update x IPL/NdYAG/CO2/RF) with shared error handling and Prometheus metrics**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-08T09:25:46Z
- **Completed:** 2026-03-08T09:31:14Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- 12 HTTP endpoints wired: POST/GET/PUT for ipl, ndyag, co2, rf under /api/v1/sessions/{id}/modules/{type}
- 4 handler files with request types, response converters, and shared error handler mapping all energy module sentinel errors
- Response models, error codes, and Prometheus counter added following existing conventions
- Route manager properly constructs EnergyModuleService with all 4 detail repositories and passes to SessionRoutes

## Task Commits

Each task was committed atomically:

1. **Task 1: Handler files, response models, error codes, and metrics** - `ecfe857` (feat)
2. **Task 2: Route registration and dependency injection wiring** - `d6c1999` (feat)

## Files Created/Modified
- `internal/api/handlers/ipl_module.go` - HandleCreateIPLModule, HandleGetIPLModule, HandleUpdateIPLModule with request types and response converter
- `internal/api/handlers/ndyag_module.go` - HandleCreateNdYAGModule, HandleGetNdYAGModule, HandleUpdateNdYAGModule with request types and response converter
- `internal/api/handlers/co2_module.go` - HandleCreateCO2Module, HandleGetCO2Module, HandleUpdateCO2Module with request types and response converter
- `internal/api/handlers/rf_module.go` - HandleCreateRFModule, HandleGetRFModule, HandleUpdateRFModule with request types and response converter
- `internal/api/handlers/energy_module_errors.go` - Shared handleEnergyModuleError mapping 10 sentinel errors to HTTP responses
- `internal/api/handlers/models.go` - Added IPLModuleDetailResponse, NdYAGModuleDetailResponse, CO2ModuleDetailResponse, RFModuleDetailResponse
- `internal/api/apierrors/apierrors.go` - Added ModuleDetailNotFound, ModuleDetailVersionConflict, ModuleDeviceTypeMismatch, ModuleHandpieceMismatch, ModuleDetailUpdateFailed
- `internal/api/metrics/prometheus.go` - Added energyModuleCreatedCounterMetric constant and IncrementEnergyModuleCreatedCount method
- `internal/api/metrics/metrics.go` - Added newEnergyModuleCreatedCounter factory (dermify_energy_module_created_total)
- `internal/api/routes/sessions.go` - Added energySvc field, 4 module type route groups with POST/GET/PUT each
- `internal/api/routes/manager.go` - Create 4 Postgres detail repos, EnergyModuleService, pass to NewSessionRoutes

## Decisions Made
- Shared handleEnergyModuleError placed in dedicated energy_module_errors.go file for clean separation (rather than embedding in one of the handler files)
- Module ID parsed from chi.URLParam "moduleId" in Get/Update handlers, consistent with existing HandleRemoveModule pattern
- Update handlers re-fetch after update to return full updated record, consistent with existing session and consent handler patterns

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 03 (Energy-Based Modules) is complete: domain types, migrations, repositories, service, and HTTP handlers all wired
- All 12 energy module endpoints are registered and the full project compiles with all 89+ tests passing
- Ready for Phase 04 (Injectable Modules) or other follow-on work

## Self-Check: PASSED

All 11 files verified present. Both task commits (ecfe857, d6c1999) verified in git log.

---
*Phase: 03-energy-based-modules*
*Completed: 2026-03-08*
