# Phase 3: Energy-Based Modules - Research

**Researched:** 2026-03-08
**Domain:** PostgreSQL polymorphic detail tables, Go service/repository/handler patterns for energy-based procedure modules (IPL, Nd:YAG, CO2/ablative, RF)
**Confidence:** HIGH

## Summary

Phase 3 implements the "detail half" of the hybrid polymorphism pattern established in Phase 2. The `session_modules` table already exists as the polymorphic base, storing module type and sort order. This phase adds four per-type detail tables (`ipl_module_details`, `ndyag_module_details`, `co2_module_details`, `rf_module_details`) each with a 1:1 foreign key back to `session_modules.id` and containing the type-specific clinical parameters from the requirements.

The existing codebase provides all scaffolding needed: `SessionModule` domain type with module type constants, `ModuleRepository` interface for CRUD on the base table, `SessionService.AddModule()` with consent gate and session editability checks, and working handler/route patterns. Phase 3 extends this by: (1) adding detail domain types, (2) creating detail tables via migrations, (3) building a new `EnergyModuleService` (or extending `SessionService`) that orchestrates base module creation + detail insertion in a single operation, (4) adding new repository interfaces for detail CRUD, and (5) wiring new HTTP handlers for creating/reading/updating energy module details.

The key architectural decision is that the existing `AddModule` creates a base `session_modules` row, and Phase 3 must add detail data either in the same operation (extended AddModule) or as a follow-up call. The cleaner approach is a new dedicated service per module type (or a single `EnergyModuleService`) that wraps AddModule + detail insertion, validates device existence via the registry, and returns a composite response.

**Primary recommendation:** Create one `EnergyModuleService` with methods per module type (`CreateIPLModule`, `CreateNdYAGModule`, etc.) that validates device/handpiece references against the registry, calls the existing `ModuleRepository.Create` for the base row, then inserts detail data into the type-specific table. Use separate detail tables with FK constraints to `session_modules(id)` and `devices(id)`.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| IPL-01 | Clinician can add an IPL procedure module to a session | Extend AddModule flow: create base session_module row + ipl_module_details row. Handler at POST /sessions/{id}/modules/ipl |
| IPL-02 | Module captures: filter/band, lightguide size, fluence, pulse duration, pulse delay, pulse count (MSP), passes, total pulses, cooling mode | ipl_module_details table with all fields as nullable columns. Domain type IPLModuleDetail |
| IPL-03 | Module links to a device from the registry with handpiece selection | FK device_id -> devices(id), handpiece_id -> handpieces(id). Validate existence via RegistryRepository.GetDeviceByID |
| YAG-01 | Clinician can add an Nd:YAG procedure module to a session | Same pattern as IPL. Handler at POST /sessions/{id}/modules/ndyag |
| YAG-02 | Module captures: wavelength, spot size, fluence, pulse duration, repetition rate, cooling type, total pulses | ndyag_module_details table. Domain type NdYAGModuleDetail |
| YAG-03 | Module links to a device from the registry | FK device_id -> devices(id). Validate via registry |
| CO2-01 | Clinician can add a CO2/ablative procedure module to a session | Same pattern. Handler at POST /sessions/{id}/modules/co2 |
| CO2-02 | Module captures: mode, handpiece/scanner, power, pulse energy, pulse duration, density, pattern, passes, anaesthesia used | co2_module_details table. Domain type CO2ModuleDetail |
| CO2-03 | Module links to a device from the registry | FK device_id -> devices(id). Validate via registry |
| RF-01 | Clinician can add an RF/RF microneedling procedure module to a session | Same pattern. Handler at POST /sessions/{id}/modules/rf |
| RF-02 | Module captures: RF mode, tip type, depth, energy level, overlap, pulses per zone, total pulses | rf_module_details table. Domain type RFModuleDetail |
| RF-03 | Module links to a device from the registry | FK device_id -> devices(id). Validate via registry |
</phase_requirements>

## Standard Stack

### Core (already in project)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-chi/chi | v5 | HTTP routing | Already used for all routes |
| database/sql | stdlib | DB access | Project uses stdlib sql with pgx driver |
| jackc/pgx | v4 | PostgreSQL driver | Already configured as DB driver |
| pressly/goose | v3 | Migrations | Embedded SQL migrations already in place |
| stretchr/testify | latest | Testing | assert/require used in all tests |
| prometheus/client_golang | latest | Metrics | Counter pattern established |

### Supporting (no new dependencies needed)
This phase requires **zero new dependencies**. All patterns use existing Go stdlib + project libraries.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Separate detail tables per type | Single JSONB column on session_modules | JSONB loses FK constraints, schema validation, and query-ability. Separate tables are the established decision. |
| One handler per module type | Generic handler with type switch | Per-type handlers are cleaner, match project convention (one file per domain), and avoid complex generic JSON unmarshaling. |
| Separate service per module type | Single EnergyModuleService | Single service reduces constructor count while keeping methods type-specific. Best balance. |

## Architecture Patterns

### Recommended Project Structure
```
internal/
  domain/
    ipl_module.go           # IPLModuleDetail type
    ndyag_module.go         # NdYAGModuleDetail type
    co2_module.go           # CO2ModuleDetail type
    rf_module.go            # RFModuleDetail type
  service/
    energy_module.go        # EnergyModuleService with Create/Get/Update per type
    energy_module_test.go   # Unit tests
  repository/postgres/
    ipl_module.go           # IPLModuleRepository
    ndyag_module.go         # NdYAGModuleRepository
    co2_module.go           # CO2ModuleRepository
    rf_module.go            # RFModuleRepository
  api/handlers/
    ipl_module.go           # HandleCreateIPLModule, HandleGetIPLModule, HandleUpdateIPLModule
    ndyag_module.go         # HandleCreateNdYAGModule, HandleGetNdYAGModule, HandleUpdateNdYAGModule
    co2_module.go           # HandleCreateCO2Module, HandleGetCO2Module, HandleUpdateCO2Module
    rf_module.go            # HandleCreateRFModule, HandleGetRFModule, HandleUpdateRFModule
  api/routes/
    sessions.go             # EXTENDED: add module-type-specific sub-routes
  testutil/
    mock_energy_module.go   # Mock repositories for energy module detail tables
migrations/
    20260308020000_create_ipl_module_details.sql
    20260308020001_create_ndyag_module_details.sql
    20260308020002_create_co2_module_details.sql
    20260308020003_create_rf_module_details.sql
```

### Pattern 1: Hybrid Polymorphism -- Base + Detail Tables
**What:** `session_modules` is the polymorphic base (already exists). Each module type gets a detail table with a 1:1 FK to `session_modules.id`. The detail table holds type-specific clinical parameters and the `device_id` FK to the device registry.
**When to use:** Always, for all four module types in this phase.
**Example:**
```sql
-- Detail table pattern (IPL example, others follow same structure)
CREATE TABLE ipl_module_details (
    id              BIGSERIAL PRIMARY KEY,
    module_id       BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    device_id       BIGINT NOT NULL REFERENCES devices(id),
    handpiece_id    BIGINT REFERENCES handpieces(id),
    filter_band     VARCHAR(100),
    lightguide_size VARCHAR(50),
    fluence         DECIMAL(6,2),        -- J/cm2
    pulse_duration  DECIMAL(8,2),        -- ms
    pulse_delay     DECIMAL(8,2),        -- ms
    pulse_count     INTEGER,             -- MSP pulse count
    passes          INTEGER,
    total_pulses    INTEGER,
    cooling_mode    VARCHAR(50),
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id)
);
```

### Pattern 2: Composite Service Operation
**What:** The `EnergyModuleService` orchestrates base module + detail creation in a single service method. It reuses existing `SessionService.AddModule` logic (consent gate, editability check) but adds device validation and detail insertion.
**When to use:** For all Create operations on energy modules.
**Example:**
```go
// EnergyModuleService orchestrates energy-based module operations.
type EnergyModuleService struct {
    sessionSvc   *SessionService
    registrySvc  *RegistryService
    iplRepo      IPLModuleRepository
    ndyagRepo    NdYAGModuleRepository
    co2Repo      CO2ModuleRepository
    rfRepo       RFModuleRepository
}

func (s *EnergyModuleService) CreateIPLModule(ctx context.Context, sessionID int64, detail *domain.IPLModuleDetail, userID int64) (*domain.IPLModuleDetail, error) {
    // 1. Validate device exists in registry
    // 2. Validate handpiece belongs to device (if provided)
    // 3. Create base session_module via SessionService.AddModule
    // 4. Insert detail row into ipl_module_details
    // 5. Return composite result
}
```

### Pattern 3: Per-Type Handler Files
**What:** Each module type gets its own handler file following the project convention of one file per domain.
**When to use:** For all four module types.
**Example:**
```go
// Handler follows existing pattern: closure returning http.HandlerFunc
func HandleCreateIPLModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Parse session ID from URL
        // 2. Get user claims from context
        // 3. Decode type-specific request body
        // 4. Call service.CreateIPLModule
        // 5. Return 201 with response
    }
}
```

### Pattern 4: Route Registration Under Sessions
**What:** Energy module routes nest under the existing session routes at `/sessions/{id}/modules/{type}`.
**When to use:** For all module type endpoints.
**Example:**
```go
// Inside SessionRoutes.RegisterRoutes, within the /{id} Route block:
r.Route("/modules", func(r chi.Router) {
    // Existing generic module endpoints
    r.Post("/", handlers.HandleAddModule(sr.sessionSvc, sr.metrics))
    r.Get("/", handlers.HandleListModules(sr.sessionSvc, sr.metrics))
    r.Delete("/{moduleId}", handlers.HandleRemoveModule(sr.sessionSvc, sr.metrics))

    // Type-specific energy module endpoints
    r.Route("/ipl", func(r chi.Router) {
        r.Post("/", handlers.HandleCreateIPLModule(sr.energySvc, sr.metrics))
        r.Get("/{moduleId}", handlers.HandleGetIPLModule(sr.energySvc, sr.metrics))
        r.Put("/{moduleId}", handlers.HandleUpdateIPLModule(sr.energySvc, sr.metrics))
    })
    r.Route("/ndyag", func(r chi.Router) { /* same pattern */ })
    r.Route("/co2", func(r chi.Router) { /* same pattern */ })
    r.Route("/rf", func(r chi.Router) { /* same pattern */ })
})
```

### Anti-Patterns to Avoid
- **Single giant handler with type switch:** Avoid routing all module types through one handler that switches on type. Per-type handlers are cleaner, testable in isolation, and match project convention.
- **Storing detail data as JSON on session_modules:** Loses FK integrity to devices table, makes queries harder, and contradicts the hybrid polymorphism decision from the roadmap.
- **Skipping device validation:** The device_id FK constraint alone is not sufficient for good error messages. The service layer must validate device existence and return a clear error before attempting the insert.
- **Creating detail without base module:** Always create the `session_modules` row first (which enforces consent gate), then the detail row. Never insert directly into detail tables without the base.
- **Coupling module detail to session status checks:** Session editability is already checked in `SessionService.AddModule`. The detail service should delegate to it, not re-implement status checks.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Device existence check | Custom SQL lookup | `RegistryRepository.GetDeviceByID` | Already exists, returns proper sentinel error |
| Session editability + consent gate | Custom status checks | `SessionService.AddModule` | Existing method handles all validation |
| Optimistic locking | Manual version checks | Existing `WHERE version = $expected` pattern from session/consent repos | Battle-tested pattern in codebase |
| JSON error responses | Custom error formatting | `apierrors.WriteError` | Consistent error format project-wide |
| Pagination | Custom page logic | Existing `PaginatedResponse` pattern | Already standardized |

**Key insight:** Phase 3 is almost entirely an extension of existing patterns. The only new concept is the 1:1 detail table + composite service. Everything else (handlers, routes, repos, mocks, tests) follows Phase 2 patterns exactly.

## Common Pitfalls

### Pitfall 1: Orphaned Detail Rows
**What goes wrong:** If the service creates the base `session_modules` row but crashes before inserting the detail row, you get an orphaned base row with no detail.
**Why it happens:** Two separate INSERT operations without a transaction.
**How to avoid:** Wrap base + detail creation in a single database transaction. Use `db.BeginTx` in the repository layer or pass a `*sql.Tx` to both operations.
**Warning signs:** Base module rows with no corresponding detail row in any detail table.

### Pitfall 2: Handpiece-Device Mismatch
**What goes wrong:** A clinician selects handpiece_id=5 which belongs to device_id=2, but passes device_id=3. The FK constraints pass individually but the handpiece doesn't belong to the device.
**Why it happens:** The FK only validates existence, not the parent-child relationship.
**How to avoid:** In the service layer, when handpiece_id is provided, query the handpiece and verify `handpiece.device_id == request.device_id`. Return a clear validation error if mismatched.
**Warning signs:** Clinically nonsensical data (e.g., "IPL 515nm Filter" handpiece linked to a CO2 device).

### Pitfall 3: Device Type Mismatch
**What goes wrong:** A clinician creates an IPL module but references an RF device from the registry.
**Why it happens:** The `devices.device_type` column has a CHECK constraint with type values, but nothing prevents an IPL module from referencing an RF device.
**How to avoid:** In the service layer, after fetching the device, validate `device.DeviceType == moduleType`. For IPL modules, device_type must be "ipl".
**Warning signs:** IPL module detail rows referencing CO2 or RF devices.

### Pitfall 4: golangci-lint Exhaustive Struct Check
**What goes wrong:** The `exhaustruct` linter requires all struct fields to be initialized explicitly, which creates verbose constructors.
**Why it happens:** The project uses a strict golangci config with `exhaustruct` enabled.
**How to avoid:** Add new domain types to the `exhaustruct.exclude` list in `golangci.yaml` if they have many optional fields, OR initialize all fields explicitly in constructors and tests.
**Warning signs:** Lint failures on struct initialization.

### Pitfall 5: Decimal Precision for Clinical Parameters
**What goes wrong:** Using `float64` for fluence/energy values introduces floating-point precision errors. A clinician enters 15.5 J/cm2 and the system stores 15.499999999.
**Why it happens:** Go float64 has IEEE 754 limitations.
**How to avoid:** Use `DECIMAL(p,s)` in PostgreSQL and scan into `*float64` in Go. For display, format to the expected decimal places. Alternatively, use string representation for exact decimals. The project does not currently use a decimal library -- stick with float64 and DECIMAL column (pgx handles the conversion correctly for typical clinical ranges).
**Warning signs:** Values like 15.4999999 in API responses.

### Pitfall 6: Module Detail Update Without Version Check
**What goes wrong:** Two clinicians update the same IPL module detail simultaneously, and one overwrites the other's changes.
**Why it happens:** Missing optimistic locking on detail table updates.
**How to avoid:** Include `version` column on detail tables and use `WHERE id = $1 AND version = $2` in UPDATE queries, same pattern as sessions and consent.
**Warning signs:** Silent data loss on concurrent updates.

## Code Examples

### Domain Type: IPL Module Detail
```go
package domain

import "time"

// IPLModuleDetail holds IPL-specific parameters for a session module.
type IPLModuleDetail struct {
    ID            int64     `json:"id"`
    ModuleID      int64     `json:"module_id"`
    DeviceID      int64     `json:"device_id"`
    HandpieceID   *int64    `json:"handpiece_id"`
    FilterBand    *string   `json:"filter_band"`
    LightguideSize *string  `json:"lightguide_size"`
    Fluence       *float64  `json:"fluence"`         // J/cm2
    PulseDuration *float64  `json:"pulse_duration"`   // ms
    PulseDelay    *float64  `json:"pulse_delay"`      // ms
    PulseCount    *int      `json:"pulse_count"`      // MSP
    Passes        *int      `json:"passes"`
    TotalPulses   *int      `json:"total_pulses"`
    CoolingMode   *string   `json:"cooling_mode"`
    Notes         *string   `json:"notes"`
    Version       int       `json:"version"`
    CreatedAt     time.Time `json:"created_at"`
    CreatedBy     int64     `json:"created_by"`
    UpdatedAt     time.Time `json:"updated_at"`
    UpdatedBy     int64     `json:"updated_by"`
}
```

### Domain Type: Nd:YAG Module Detail
```go
// NdYAGModuleDetail holds Nd:YAG-specific parameters for a session module.
type NdYAGModuleDetail struct {
    ID             int64     `json:"id"`
    ModuleID       int64     `json:"module_id"`
    DeviceID       int64     `json:"device_id"`
    HandpieceID    *int64    `json:"handpiece_id"`
    Wavelength     *string   `json:"wavelength"`      // e.g. "1064nm", "755nm"
    SpotSize       *string   `json:"spot_size"`        // e.g. "8mm", "12mm"
    Fluence        *float64  `json:"fluence"`           // J/cm2
    PulseDuration  *float64  `json:"pulse_duration"`    // ms
    RepetitionRate *float64  `json:"repetition_rate"`   // Hz
    CoolingType    *string   `json:"cooling_type"`
    TotalPulses    *int      `json:"total_pulses"`
    Notes          *string   `json:"notes"`
    Version        int       `json:"version"`
    CreatedAt      time.Time `json:"created_at"`
    CreatedBy      int64     `json:"created_by"`
    UpdatedAt      time.Time `json:"updated_at"`
    UpdatedBy      int64     `json:"updated_by"`
}
```

### Domain Type: CO2 Module Detail
```go
// CO2ModuleDetail holds CO2/ablative-specific parameters for a session module.
type CO2ModuleDetail struct {
    ID              int64     `json:"id"`
    ModuleID        int64     `json:"module_id"`
    DeviceID        int64     `json:"device_id"`
    HandpieceID     *int64    `json:"handpiece_id"`
    Mode            *string   `json:"mode"`             // e.g. "fractional", "ablative", "combo"
    ScannerPattern  *string   `json:"scanner_pattern"`   // e.g. "square", "hexagonal"
    Power           *float64  `json:"power"`             // W
    PulseEnergy     *float64  `json:"pulse_energy"`      // mJ
    PulseDuration   *float64  `json:"pulse_duration"`    // ms/us
    Density         *float64  `json:"density"`           // %
    Pattern         *string   `json:"pattern"`
    Passes          *int      `json:"passes"`
    AnaesthesiaUsed *string   `json:"anaesthesia_used"`
    Notes           *string   `json:"notes"`
    Version         int       `json:"version"`
    CreatedAt       time.Time `json:"created_at"`
    CreatedBy       int64     `json:"created_by"`
    UpdatedAt       time.Time `json:"updated_at"`
    UpdatedBy       int64     `json:"updated_by"`
}
```

### Domain Type: RF Module Detail
```go
// RFModuleDetail holds RF/RF microneedling-specific parameters for a session module.
type RFModuleDetail struct {
    ID            int64     `json:"id"`
    ModuleID      int64     `json:"module_id"`
    DeviceID      int64     `json:"device_id"`
    HandpieceID   *int64    `json:"handpiece_id"`
    RFMode        *string   `json:"rf_mode"`           // e.g. "monopolar", "bipolar", "microneedling"
    TipType       *string   `json:"tip_type"`           // e.g. "49-pin", "16-pin"
    Depth         *float64  `json:"depth"`              // mm
    EnergyLevel   *float64  `json:"energy_level"`       // mJ or W
    Overlap       *float64  `json:"overlap"`            // %
    PulsesPerZone *int      `json:"pulses_per_zone"`
    TotalPulses   *int      `json:"total_pulses"`
    Notes         *string   `json:"notes"`
    Version       int       `json:"version"`
    CreatedAt     time.Time `json:"created_at"`
    CreatedBy     int64     `json:"created_by"`
    UpdatedAt     time.Time `json:"updated_at"`
    UpdatedBy     int64     `json:"updated_by"`
}
```

### Migration Pattern
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE ipl_module_details (
    id              BIGSERIAL PRIMARY KEY,
    module_id       BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    device_id       BIGINT NOT NULL REFERENCES devices(id),
    handpiece_id    BIGINT REFERENCES handpieces(id),
    filter_band     VARCHAR(100),
    lightguide_size VARCHAR(50),
    fluence         DECIMAL(6,2),
    pulse_duration  DECIMAL(8,2),
    pulse_delay     DECIMAL(8,2),
    pulse_count     INTEGER,
    passes          INTEGER,
    total_pulses    INTEGER,
    cooling_mode    VARCHAR(50),
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_ipl_module_details_module_id ON ipl_module_details(module_id);
CREATE INDEX idx_ipl_module_details_device_id ON ipl_module_details(device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ipl_module_details;
-- +goose StatementEnd
```

### Repository Pattern
```go
// IPLModuleRepository defines data access for IPL module details.
type IPLModuleRepository interface {
    Create(ctx context.Context, detail *domain.IPLModuleDetail) error
    GetByModuleID(ctx context.Context, moduleID int64) (*domain.IPLModuleDetail, error)
    Update(ctx context.Context, detail *domain.IPLModuleDetail) error
}
```

### Service Validation Pattern
```go
func (s *EnergyModuleService) validateDeviceForModule(ctx context.Context, deviceID int64, handpieceID *int64, expectedType string) error {
    device, err := s.registrySvc.GetDeviceByID(ctx, deviceID)
    if err != nil {
        return err // ErrDeviceNotFound propagates
    }

    if device.DeviceType != expectedType {
        return ErrDeviceTypeMismatch
    }

    if handpieceID != nil {
        found := false
        for _, hp := range device.Handpieces {
            if hp.ID == *handpieceID {
                found = true
                break
            }
        }
        if !found {
            return ErrHandpieceNotFound
        }
    }

    return nil
}
```

### Handler Request/Response Pattern
```go
type createIPLModuleRequest struct {
    DeviceID       int64    `json:"device_id"`
    HandpieceID    *int64   `json:"handpiece_id"`
    FilterBand     *string  `json:"filter_band"`
    LightguideSize *string  `json:"lightguide_size"`
    Fluence        *float64 `json:"fluence"`
    PulseDuration  *float64 `json:"pulse_duration"`
    PulseDelay     *float64 `json:"pulse_delay"`
    PulseCount     *int     `json:"pulse_count"`
    Passes         *int     `json:"passes"`
    TotalPulses    *int     `json:"total_pulses"`
    CoolingMode    *string  `json:"cooling_mode"`
    Notes          *string  `json:"notes"`
}
```

### API Endpoints Summary
```
POST   /api/v1/sessions/{id}/modules/ipl            Create IPL module
GET    /api/v1/sessions/{id}/modules/ipl/{moduleId}  Get IPL module detail
PUT    /api/v1/sessions/{id}/modules/ipl/{moduleId}  Update IPL module detail

POST   /api/v1/sessions/{id}/modules/ndyag           Create Nd:YAG module
GET    /api/v1/sessions/{id}/modules/ndyag/{moduleId} Get Nd:YAG module detail
PUT    /api/v1/sessions/{id}/modules/ndyag/{moduleId} Update Nd:YAG module detail

POST   /api/v1/sessions/{id}/modules/co2             Create CO2 module
GET    /api/v1/sessions/{id}/modules/co2/{moduleId}   Get CO2 module detail
PUT    /api/v1/sessions/{id}/modules/co2/{moduleId}   Update CO2 module detail

POST   /api/v1/sessions/{id}/modules/rf              Create RF module
GET    /api/v1/sessions/{id}/modules/rf/{moduleId}    Get RF module detail
PUT    /api/v1/sessions/{id}/modules/rf/{moduleId}    Update RF module detail
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single table inheritance (STI) | Hybrid polymorphism (base + detail tables) | Project decision at roadmap | Each module type gets type-safe columns, FK constraints, and clean queries |
| JSONB blob for parameters | Typed columns with DECIMAL precision | Project decision | Schema-enforced data integrity, SQL-queryable parameters |
| Generic module creation | Type-specific creation with device validation | Phase 3 (this phase) | Prevents invalid device-module associations |

## Open Questions

1. **Transaction handling for composite inserts**
   - What we know: The project uses `database/sql` with pgx driver. Both `db.BeginTx` and manual transaction management are available.
   - What's unclear: Whether the existing `ModuleRepository.Create` accepts a `*sql.Tx` or only `*sql.DB`. Currently it uses `r.db.QueryRowContext` which works with both `*sql.DB` and `*sql.Tx` (both implement the query interface).
   - Recommendation: The `*sql.DB` methods (`QueryRowContext`, `ExecContext`) also work on `*sql.Tx`. Create a new method in the energy module repository that takes a `*sql.Tx`, or have the service manage the transaction and pass it through. Alternatively, since `database/sql.DB` connection pool is used, the simplest approach is to have the energy module repository do both inserts (base + detail) in a single transactional method.

2. **Clinical parameter fields: nullable vs required**
   - What we know: Requirements list the parameters each module captures but don't specify which are mandatory vs optional.
   - What's unclear: In clinical practice, a clinician may document a session incrementally -- filling in some parameters now and the rest later.
   - Recommendation: Make all clinical parameters nullable (pointers in Go, nullable columns in SQL). Only `device_id` should be required (NOT NULL) since the module must link to a device. This allows saving partial data and completing it later while the session is in draft/in_progress.

3. **Whether existing `HandleAddModule` endpoint should still work independently**
   - What we know: The existing `POST /sessions/{id}/modules` creates a bare `session_modules` row with just a type.
   - What's unclear: Should clinicians still be able to create a bare module slot (without detail), or should they always go through the type-specific endpoint?
   - Recommendation: Keep the existing generic endpoint working for backward compatibility, but have the type-specific endpoints be the primary creation path. The generic endpoint creates a "skeleton" that can have detail attached later, while the type-specific endpoint creates both in one step. This preserves the existing test suite and allows Phase 4 to follow the same pattern for injectables.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (assert/require) |
| Config file | None (Go standard test runner) |
| Quick run command | `go test ./internal/service/... -count=1 -run TestEnergy -v` |
| Full suite command | `make test` (runs `go test ./... -count=1 -v`) |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| IPL-01 | Create IPL module in session | unit | `go test ./internal/service/... -count=1 -run TestCreateIPLModule -v` | Wave 0 |
| IPL-02 | IPL parameter capture | unit | `go test ./internal/service/... -count=1 -run TestIPLModule_Parameters -v` | Wave 0 |
| IPL-03 | IPL device linkage | unit | `go test ./internal/service/... -count=1 -run TestIPLModule_DeviceValidation -v` | Wave 0 |
| YAG-01 | Create Nd:YAG module | unit | `go test ./internal/service/... -count=1 -run TestCreateNdYAGModule -v` | Wave 0 |
| YAG-02 | Nd:YAG parameter capture | unit | `go test ./internal/service/... -count=1 -run TestNdYAGModule_Parameters -v` | Wave 0 |
| YAG-03 | Nd:YAG device linkage | unit | `go test ./internal/service/... -count=1 -run TestNdYAGModule_DeviceValidation -v` | Wave 0 |
| CO2-01 | Create CO2 module | unit | `go test ./internal/service/... -count=1 -run TestCreateCO2Module -v` | Wave 0 |
| CO2-02 | CO2 parameter capture | unit | `go test ./internal/service/... -count=1 -run TestCO2Module_Parameters -v` | Wave 0 |
| CO2-03 | CO2 device linkage | unit | `go test ./internal/service/... -count=1 -run TestCO2Module_DeviceValidation -v` | Wave 0 |
| RF-01 | Create RF module | unit | `go test ./internal/service/... -count=1 -run TestCreateRFModule -v` | Wave 0 |
| RF-02 | RF parameter capture | unit | `go test ./internal/service/... -count=1 -run TestRFModule_Parameters -v` | Wave 0 |
| RF-03 | RF device linkage | unit | `go test ./internal/service/... -count=1 -run TestRFModule_DeviceValidation -v` | Wave 0 |

### Cross-Cutting Tests
| Behavior | Test Type | Automated Command |
|----------|-----------|-------------------|
| Non-existent device returns error | unit | `go test ./internal/service/... -count=1 -run TestEnergyModule_DeviceNotFound -v` |
| Device type mismatch returns error | unit | `go test ./internal/service/... -count=1 -run TestEnergyModule_DeviceTypeMismatch -v` |
| Handpiece-device mismatch returns error | unit | `go test ./internal/service/... -count=1 -run TestEnergyModule_HandpieceMismatch -v` |
| Multiple module types in one session | unit | `go test ./internal/service/... -count=1 -run TestEnergyModule_MultipleTypes -v` |
| Non-editable session rejected | unit | `go test ./internal/service/... -count=1 -run TestEnergyModule_NonEditableSession -v` |

### Sampling Rate
- **Per task commit:** `go test ./internal/service/... -count=1 -v`
- **Per wave merge:** `make test`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/service/energy_module_test.go` -- covers IPL-01 through RF-03
- [ ] `internal/testutil/mock_energy_module.go` -- mock repos for all 4 detail types
- [ ] `internal/testutil/mock_registry.go` -- already exists, may need update for device-type validation

## Sources

### Primary (HIGH confidence)
- Existing codebase analysis -- all patterns derived from reading the actual Phase 1 and Phase 2 implementation
- `internal/domain/session_module.go` -- base module type with constants for all 6 module types
- `internal/service/session.go` -- AddModule flow with consent gate and editability checks
- `internal/repository/postgres/session_module.go` -- base module CRUD pattern
- `migrations/20260308010004_create_session_modules.sql` -- base table schema with CHECK constraint
- `migrations/20260307150000_create_devices_tables.sql` -- device and handpiece schema
- `migrations/20260307160000_seed_devices.sql` -- actual device seed data showing device types and handpieces

### Secondary (MEDIUM confidence)
- `.planning/ROADMAP.md` -- "Hybrid polymorphism for modules -- shared session_modules table + per-type detail tables" decision
- `.planning/STATE.md` -- accumulated decisions from Phases 1 and 2

### Tertiary (LOW confidence)
- Clinical parameter field names -- derived from REQUIREMENTS.md descriptions. Exact column types (DECIMAL precision, VARCHAR lengths) are reasonable defaults but could be refined with clinical input.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies, all patterns exist in codebase
- Architecture: HIGH -- hybrid polymorphism decision is locked, implementation follows existing patterns exactly
- Pitfalls: HIGH -- transaction handling and device validation are well-understood concerns with clear solutions
- Clinical parameters: MEDIUM -- field names and types derived from requirements text, may need clinical review for precision/units

**Research date:** 2026-03-08
**Valid until:** 2026-04-08 (stable -- project conventions unlikely to change mid-milestone)
