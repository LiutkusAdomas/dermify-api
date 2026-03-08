# Phase 4: Injectable Modules and Outcomes - Research

**Researched:** 2026-03-08
**Domain:** Injectable procedure module detail tables (filler, botulinum toxin), session outcomes, aftercare, and follow-up scheduling in a Go/PostgreSQL REST API
**Confidence:** HIGH

## Summary

Phase 4 adds two new injectable procedure module types (filler and botulinum toxin) and a session-level outcomes/aftercare/follow-up system. The injectable modules follow the exact same hybrid polymorphism pattern established in Phase 3 for energy modules: a 1:1 detail table linked to the existing `session_modules` polymorphic base. The key difference is that injectable modules reference **products** from the product registry (not devices), requiring a new product validation path analogous to the device validation in `EnergyModuleService`.

The outcomes/aftercare/follow-up domain is structurally different from modules. It is a **session-level singleton** (one outcome record per session, similar to how consent works) rather than a per-module entity. The outcome record captures the session status (completed/partial/aborted), links to clinical endpoints observed (a many-to-many relationship filtered by module type), stores aftercare instructions with mandatory red flags, and records a follow-up date/time. This maps closely to the existing `ConsentService` + `ConsentRepository` pattern: Create/Get/Update on a single record per session.

All infrastructure needed already exists: `session_modules` base table with CHECK constraint already allows `'filler'` and `'botulinum_toxin'` types, product registry and seed data are in place (3 fillers, 3 botulinum toxins), clinical endpoints for both types are seeded, and the `RegistryService.GetProductByID` method exists. Phase 4 requires zero new dependencies.

**Primary recommendation:** Create an `InjectableModuleService` (parallel to `EnergyModuleService`) with `validateProductForModule` that checks product existence and type match, then follow the identical Create/Get/Update pattern per module type. For outcomes, create a separate `OutcomeService` following the consent service pattern (one record per session with Create/Get/Update).

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| FILL-01 | Clinician can add a filler procedure module to a session | Extend AddModule flow: create base session_module row + filler_module_details row. Handler at POST /sessions/{id}/modules/filler. Same pattern as IPL/NdYAG/CO2/RF. |
| FILL-02 | Module captures: product, syringe volume, total volume, batch/lot, expiry, needle/cannula, injection planes, anatomical sites, endpoint | filler_module_details table with product_id FK + type-specific clinical fields. All clinical params nullable for incremental documentation. |
| FILL-03 | Module links to product from registry with batch/lot and expiry tracking | FK product_id -> products(id). Service validates product exists via RegistryService.GetProductByID and product_type == 'filler'. Batch/lot and expiry are detail-table columns (not FK -- they are per-procedure instance data). |
| TOX-01 | Clinician can add a botulinum toxin procedure module to a session | Same pattern as filler. Handler at POST /sessions/{id}/modules/botulinum. |
| TOX-02 | Module captures: product, batch number, reconstitution details (diluent, volume, concentration), total units, injection sites with units per site | botulinum_module_details table with reconstitution fields + injection_sites JSONB for per-site unit mapping. |
| TOX-03 | Module links to product from registry with batch tracking | FK product_id -> products(id). Service validates product_type == 'botulinum_toxin'. Batch number is a detail-table column. |
| OUT-01 | Clinician can record immediate outcome (completed/partial/aborted) | Session-level outcome table with status enum column. One outcome per session (like consent). |
| OUT-02 | Clinician can select clinical endpoints observed (module-specific list) | Junction table session_outcome_endpoints linking outcome to clinical_endpoints. Endpoint list filtered by module types present in session. |
| OUT-03 | Clinician can record aftercare provided with templated instructions | Aftercare text column(s) on outcome table. Templated instructions are application-level (the API stores the final text). |
| OUT-04 | Aftercare includes mandatory red flags and contact section | Separate red_flags_text and contact_info columns on outcome table to enforce presence. Service validates these are non-empty when aftercare is provided. |
| OUT-05 | Clinician can set follow-up date/time linked to session | follow_up_at TIMESTAMPTZ column on outcome table. Nullable -- only set if clinician schedules follow-up. |
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
This phase requires **zero new dependencies**. All patterns use existing Go stdlib + project libraries. The JSONB type for injection sites is handled natively by pgx/database/sql.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Separate filler + botulinum detail tables | Single injectables table with nullable columns for both | Separate tables keep columns clean and type-safe. Filler has syringe volume/injection planes; botulinum has reconstitution/units-per-site. Very different schemas. |
| JSONB for injection sites | Separate injection_sites child table | JSONB is simpler for a list of {site, units} pairs, avoids extra join, and Phase 5 locking will freeze the whole detail row anyway. A child table adds complexity without clear benefit for read-heavy documentation. |
| Outcome as part of session table | Separate session_outcomes table | Separate table keeps session table clean, allows outcome to have its own version/audit trail, and matches the "one sub-record per session" pattern established by consent. |
| Single OutcomeService for everything | Separate aftercare/follow-up services | Outcome, aftercare, and follow-up are tightly coupled (all recorded at session completion). One service with one table is simpler and matches the workflow. |

## Architecture Patterns

### Recommended Project Structure
```
internal/
  domain/
    filler_module.go          # FillerModuleDetail type
    botulinum_module.go       # BotulinumModuleDetail type
    outcome.go                # SessionOutcome type
  service/
    injectable_module.go      # InjectableModuleService with Create/Get/Update per type
    injectable_module_test.go # Unit tests
    outcome.go                # OutcomeService with Create/Get/Update
    outcome_test.go           # Unit tests
  repository/postgres/
    filler_module.go          # FillerModuleRepository
    filler_module_test.go     # Compile-time interface check
    botulinum_module.go       # BotulinumModuleRepository
    botulinum_module_test.go  # Compile-time interface check
    outcome.go                # OutcomeRepository
    outcome_test.go           # Compile-time interface check
  api/handlers/
    filler_module.go          # HandleCreateFillerModule, HandleGetFillerModule, HandleUpdateFillerModule
    botulinum_module.go       # HandleCreateBotulinumModule, HandleGetBotulinumModule, HandleUpdateBotulinumModule
    injectable_module_errors.go # Shared error mapper for injectable modules
    outcome.go                # HandleRecordOutcome, HandleGetOutcome, HandleUpdateOutcome
  api/routes/
    sessions.go               # EXTENDED: add filler/botulinum/outcome sub-routes
  testutil/
    mock_injectable_module.go # Mock repos for filler + botulinum detail tables
    mock_outcome.go           # Mock repo for outcome
migrations/
    20260308030000_create_filler_module_details.sql
    20260308030001_create_botulinum_module_details.sql
    20260308030002_create_session_outcomes.sql
```

### Pattern 1: Injectable Module Detail Tables (Hybrid Polymorphism)
**What:** Same 1:1 base-detail pattern as energy modules. `session_modules` is the polymorphic base (already supports `'filler'` and `'botulinum_toxin'` types). Each injectable type gets a detail table with FK to `session_modules(id)` and FK to `products(id)` instead of `devices(id)`.
**When to use:** For both filler and botulinum toxin module types.
**Key difference from energy modules:** References `products(id)` not `devices(id)`. No handpiece concept. Has batch/lot and expiry tracking fields that energy modules lack.
**Example:**
```sql
CREATE TABLE filler_module_details (
    id              BIGSERIAL PRIMARY KEY,
    module_id       BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    product_id      BIGINT NOT NULL REFERENCES products(id),
    batch_number    VARCHAR(100),
    expiry_date     DATE,
    syringe_volume  DECIMAL(6,2),         -- mL per syringe
    total_volume    DECIMAL(6,2),         -- mL total used
    needle_type     VARCHAR(100),          -- e.g. "27G needle", "25G cannula"
    injection_plane VARCHAR(100),          -- e.g. "subdermal", "supraperiosteal"
    anatomical_sites TEXT,                 -- comma-separated or structured
    endpoint        VARCHAR(200),          -- clinical endpoint observed
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id)
);
```

### Pattern 2: Botulinum Toxin with Reconstitution and Injection Site Mapping
**What:** Botulinum toxin has unique clinical requirements: reconstitution details (what diluent, how much, resulting concentration) and a per-site injection mapping (which anatomical site, how many units).
**When to use:** For botulinum toxin module only.
**Design decision -- injection sites:** Use JSONB column for injection sites rather than a child table. The data is always read/written as a unit with the module detail, and Phase 5 locking will freeze the entire row. JSONB avoids join complexity.
**Example:**
```sql
CREATE TABLE botulinum_module_details (
    id                    BIGSERIAL PRIMARY KEY,
    module_id             BIGINT NOT NULL UNIQUE REFERENCES session_modules(id) ON DELETE CASCADE,
    product_id            BIGINT NOT NULL REFERENCES products(id),
    batch_number          VARCHAR(100),
    expiry_date           DATE,
    diluent               VARCHAR(100),          -- e.g. "0.9% NaCl"
    dilution_volume       DECIMAL(6,2),          -- mL of diluent added
    resulting_concentration VARCHAR(100),         -- e.g. "4 units/0.1mL"
    total_units           DECIMAL(8,2),          -- total units administered
    injection_sites       JSONB,                 -- [{site: "glabella", units: 20}, ...]
    notes                 TEXT,
    version               INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by            BIGINT NOT NULL REFERENCES users(id),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by            BIGINT NOT NULL REFERENCES users(id)
);
```

### Pattern 3: Session Outcome as Singleton (Consent-like Pattern)
**What:** One outcome record per session, following the exact same Create/Get/Update pattern as consent. Captures outcome status, clinical endpoints observed, aftercare text, red flags, and follow-up scheduling.
**When to use:** For the session outcome domain (OUT-01 through OUT-05).
**Key structural elements:**
- `session_outcomes` table: 1:1 with sessions, holds status/aftercare/follow-up
- `session_outcome_endpoints` junction table: many-to-many linking outcomes to clinical_endpoints (the module-specific list from the registry)
**Example:**
```sql
CREATE TABLE session_outcomes (
    id                BIGSERIAL PRIMARY KEY,
    session_id        BIGINT NOT NULL UNIQUE REFERENCES sessions(id) ON DELETE CASCADE,
    outcome_status    VARCHAR(50) NOT NULL
                      CHECK (outcome_status IN ('completed', 'partial', 'aborted')),
    aftercare_notes   TEXT,
    red_flags_text    TEXT,
    contact_info      TEXT,
    follow_up_at      TIMESTAMPTZ,
    notes             TEXT,
    version           INTEGER NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by        BIGINT NOT NULL REFERENCES users(id),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by        BIGINT NOT NULL REFERENCES users(id)
);

CREATE TABLE session_outcome_endpoints (
    outcome_id  BIGINT NOT NULL REFERENCES session_outcomes(id) ON DELETE CASCADE,
    endpoint_id BIGINT NOT NULL REFERENCES clinical_endpoints(id),
    PRIMARY KEY (outcome_id, endpoint_id)
);
```

### Pattern 4: Product Validation (Analogous to Device Validation)
**What:** The `InjectableModuleService` validates product existence and type match before creating a module, exactly as `EnergyModuleService.validateDeviceForModule` does for devices.
**When to use:** For all injectable module Create operations.
**Example:**
```go
func (s *InjectableModuleService) validateProductForModule(
    ctx context.Context, productID int64, expectedProductType string,
) error {
    product, err := s.registrySvc.GetProductByID(ctx, productID)
    if err != nil {
        return fmt.Errorf("validating product: %w", err)
    }
    if product.ProductType != expectedProductType {
        return ErrProductTypeMismatch
    }
    return nil
}
```

### Pattern 5: Route Registration Extension
**What:** Injectable and outcome routes nest under existing session routes, following the same pattern as energy modules.
**When to use:** For all new endpoints in this phase.
**Example:**
```go
// Inside SessionRoutes, within the /{id} Route block:
r.Route("/modules/filler", func(r chi.Router) {
    r.Post("/", handlers.HandleCreateFillerModule(sr.injectableSvc, sr.metrics))
    r.Get("/{moduleId}", handlers.HandleGetFillerModule(sr.injectableSvc, sr.metrics))
    r.Put("/{moduleId}", handlers.HandleUpdateFillerModule(sr.injectableSvc, sr.metrics))
})
r.Route("/modules/botulinum", func(r chi.Router) {
    r.Post("/", handlers.HandleCreateBotulinumModule(sr.injectableSvc, sr.metrics))
    r.Get("/{moduleId}", handlers.HandleGetBotulinumModule(sr.injectableSvc, sr.metrics))
    r.Put("/{moduleId}", handlers.HandleUpdateBotulinumModule(sr.injectableSvc, sr.metrics))
})

// Outcome routes (session-level, like consent)
r.Post("/outcome", handlers.HandleRecordOutcome(sr.outcomeSvc, sr.metrics))
r.Get("/outcome", handlers.HandleGetOutcome(sr.outcomeSvc, sr.metrics))
r.Put("/outcome", handlers.HandleUpdateOutcome(sr.outcomeSvc, sr.metrics))
```

### Anti-Patterns to Avoid
- **Extending EnergyModuleService for injectables:** Injectables validate products, not devices. Separate service keeps concerns clean and avoids a monolith service with mixed device/product dependencies.
- **Storing batch/lot/expiry on the products table:** Batch and expiry are per-procedure-instance data, not registry data. The product registry says "Juvederm Ultra XC exists." The filler detail says "I used batch ABC123, expiring 2027-03-15." These are different concerns.
- **Using a normalized child table for botulinum injection sites:** JSONB is appropriate here -- the data is always read/written atomically with the detail row, rarely queried independently, and Phase 5 locking freezes the whole record.
- **Putting aftercare/follow-up on the session table:** Session table is already large. Outcome is a distinct lifecycle concept (recorded after treatment, not during creation).
- **Skipping red flags validation:** OUT-04 requires mandatory red flags. The service layer must validate that `red_flags_text` is non-empty when `aftercare_notes` is provided.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Product existence check | Custom SQL lookup | `RegistryService.GetProductByID` | Already exists, returns `ErrProductNotFound` |
| Session editability + consent gate | Custom status checks | `SessionService.AddModule` | Existing method handles all validation |
| Optimistic locking | Manual version checks | Existing `WHERE version = $expected` pattern | Battle-tested in every repository |
| JSON error responses | Custom formatting | `apierrors.WriteError` | Consistent error format project-wide |
| Module type validation | Custom check | `validModuleTypes` map in session.go | Already includes `filler` and `botulinum_toxin` |
| JSONB scanning in Go | Custom JSON parsing | `json.RawMessage` or `[]byte` with pgx | pgx handles JSONB natively |

**Key insight:** Phase 4 injectables are a straightforward parallel to Phase 3 energy modules. The only new structural concept is the outcomes/aftercare domain, which follows the existing consent singleton pattern. Everything else is a direct application of established patterns.

## Common Pitfalls

### Pitfall 1: Product Type Mismatch (Analogous to Device Type Mismatch)
**What goes wrong:** A clinician creates a filler module but references a botulinum_toxin product from the registry.
**Why it happens:** The FK constraint only checks existence, not type match.
**How to avoid:** `validateProductForModule` checks `product.ProductType == expectedType`. For filler modules, product_type must be `'filler'`. For botulinum modules, product_type must be `'botulinum_toxin'`.
**Warning signs:** Filler detail rows referencing botulinum toxin products.

### Pitfall 2: JSONB Injection Sites Schema Drift
**What goes wrong:** Injection sites JSONB has no schema enforcement at the database level. Different requests could store `{site: "x", units: 5}` vs `{location: "x", amount: 5}`.
**Why it happens:** JSONB is schema-free by nature.
**How to avoid:** Define a Go struct (`InjectionSite` with `Site string` and `Units float64`) and marshal/unmarshal consistently in the service layer. Validate the structure before INSERT. The handler should reject malformed injection site arrays.
**Warning signs:** Inconsistent field names in stored JSONB data.

### Pitfall 3: Outcome Without Required Red Flags
**What goes wrong:** A clinician records aftercare but omits the mandatory red flags section, violating OUT-04.
**Why it happens:** No validation on the relationship between aftercare and red flags fields.
**How to avoid:** Service-layer validation: if `aftercare_notes` is non-empty, then `red_flags_text` must also be non-empty. Return `ErrInvalidOutcomeData` if violated.
**Warning signs:** Outcome records with aftercare text but empty/null red flags.

### Pitfall 4: Outcome on Non-Completable Session
**What goes wrong:** Outcome is recorded for a session that is still in draft state with no modules.
**Why it happens:** No guard on session status when recording outcomes.
**How to avoid:** The service should check that the session is at least `in_progress` before allowing outcome recording. Draft sessions should not have outcomes.
**Warning signs:** Outcome records on sessions with no modules or still in draft.

### Pitfall 5: Clinical Endpoints Not Filtered by Module Type
**What goes wrong:** The outcome allows selecting an IPL clinical endpoint for a session that only has filler modules.
**Why it happens:** The junction table does not enforce module-type filtering.
**How to avoid:** In the service layer, when setting outcome endpoints, verify each endpoint's `module_type` matches one of the module types present in the session. Query the session's modules to get the type list, then validate each endpoint against it.
**Warning signs:** Session outcomes with clinically irrelevant endpoint selections.

### Pitfall 6: Orphaned Detail Rows (Same as Phase 3)
**What goes wrong:** Base `session_modules` row created but detail insertion fails, leaving an orphan.
**Why it happens:** Two separate INSERT operations without a transaction.
**How to avoid:** Same approach as Phase 3 -- the current codebase does not wrap these in a transaction (Phase 3 established this pattern). Since failure after base-create but before detail-create is extremely rare in a single-user clinical app, and the base row is essentially harmless, this is acceptable for v1. Document as a known limitation.
**Warning signs:** Base module rows with no corresponding detail row.

### Pitfall 7: golangci-lint exhaustruct for New Types
**What goes wrong:** The `exhaustruct` linter requires all struct fields to be initialized.
**Why it happens:** New domain types with many nullable fields.
**How to avoid:** Phase 3 established the pattern -- use `omitempty` on JSON tags and pointer types for nullable fields. The linter config may need the new types added to the exclude list if they cause issues.
**Warning signs:** Lint failures on struct initialization in tests.

## Code Examples

### Domain Type: Filler Module Detail
```go
package domain

import "time"

// FillerModuleDetail holds filler-specific parameters for a session module.
type FillerModuleDetail struct {
    ID             int64      `json:"id"`
    ModuleID       int64      `json:"module_id"`
    ProductID      int64      `json:"product_id"`
    BatchNumber    *string    `json:"batch_number,omitempty"`
    ExpiryDate     *time.Time `json:"expiry_date,omitempty"`
    SyringeVolume  *float64   `json:"syringe_volume,omitempty"`   // mL per syringe
    TotalVolume    *float64   `json:"total_volume,omitempty"`     // mL total used
    NeedleType     *string    `json:"needle_type,omitempty"`      // e.g. "27G needle"
    InjectionPlane *string    `json:"injection_plane,omitempty"`  // e.g. "subdermal"
    AnatomicalSites *string   `json:"anatomical_sites,omitempty"` // structured text
    Endpoint       *string    `json:"endpoint,omitempty"`
    Notes          *string    `json:"notes,omitempty"`
    Version        int        `json:"version"`
    CreatedAt      time.Time  `json:"created_at"`
    CreatedBy      int64      `json:"created_by"`
    UpdatedAt      time.Time  `json:"updated_at"`
    UpdatedBy      int64      `json:"updated_by"`
}
```

### Domain Type: Botulinum Module Detail
```go
package domain

import (
    "encoding/json"
    "time"
)

// InjectionSite represents a single injection site with units administered.
type InjectionSite struct {
    Site  string  `json:"site"`
    Units float64 `json:"units"`
}

// BotulinumModuleDetail holds botulinum toxin-specific parameters for a session module.
type BotulinumModuleDetail struct {
    ID                     int64           `json:"id"`
    ModuleID               int64           `json:"module_id"`
    ProductID              int64           `json:"product_id"`
    BatchNumber            *string         `json:"batch_number,omitempty"`
    ExpiryDate             *time.Time      `json:"expiry_date,omitempty"`
    Diluent                *string         `json:"diluent,omitempty"`                  // e.g. "0.9% NaCl"
    DilutionVolume         *float64        `json:"dilution_volume,omitempty"`           // mL
    ResultingConcentration *string         `json:"resulting_concentration,omitempty"`   // e.g. "4 units/0.1mL"
    TotalUnits             *float64        `json:"total_units,omitempty"`
    InjectionSites         json.RawMessage `json:"injection_sites,omitempty"`           // []InjectionSite as JSONB
    Notes                  *string         `json:"notes,omitempty"`
    Version                int             `json:"version"`
    CreatedAt              time.Time       `json:"created_at"`
    CreatedBy              int64           `json:"created_by"`
    UpdatedAt              time.Time       `json:"updated_at"`
    UpdatedBy              int64           `json:"updated_by"`
}
```

### Domain Type: Session Outcome
```go
package domain

import "time"

// Outcome status constants.
const (
    OutcomeStatusCompleted = "completed"
    OutcomeStatusPartial   = "partial"
    OutcomeStatusAborted   = "aborted"
)

// SessionOutcome represents the outcome, aftercare, and follow-up for a session.
type SessionOutcome struct {
    ID             int64      `json:"id"`
    SessionID      int64      `json:"session_id"`
    OutcomeStatus  string     `json:"outcome_status"`
    EndpointIDs    []int64    `json:"endpoint_ids,omitempty"`
    AftercareNotes *string    `json:"aftercare_notes,omitempty"`
    RedFlagsText   *string    `json:"red_flags_text,omitempty"`
    ContactInfo    *string    `json:"contact_info,omitempty"`
    FollowUpAt     *time.Time `json:"follow_up_at,omitempty"`
    Notes          *string    `json:"notes,omitempty"`
    Version        int        `json:"version"`
    CreatedAt      time.Time  `json:"created_at"`
    CreatedBy      int64      `json:"created_by"`
    UpdatedAt      time.Time  `json:"updated_at"`
    UpdatedBy      int64      `json:"updated_by"`
}
```

### Service: InjectableModuleService
```go
// InjectableModuleService handles business logic for injectable treatment
// module details (filler, botulinum toxin).
type InjectableModuleService struct {
    sessionSvc  *SessionService
    registrySvc *RegistryService
    fillerRepo  FillerModuleRepository
    botulinumRepo BotulinumModuleRepository
}

func (s *InjectableModuleService) CreateFillerModule(
    ctx context.Context, sessionID int64, detail *domain.FillerModuleDetail, userID int64,
) (*domain.FillerModuleDetail, error) {
    // 1. Validate product exists and is type "filler"
    // 2. Create base session_module via SessionService.AddModule (consent gate + editability)
    // 3. Insert detail row into filler_module_details
    // 4. Return composite result
}
```

### Service: OutcomeService
```go
// OutcomeService handles outcome, aftercare, and follow-up business logic.
type OutcomeService struct {
    repo       OutcomeRepository
    sessionSvc *SessionService
}

func (s *OutcomeService) RecordOutcome(ctx context.Context, outcome *domain.SessionOutcome) error {
    // 1. Validate session exists and is at least in_progress
    // 2. Check no outcome already exists (one per session)
    // 3. Validate required fields (outcome_status)
    // 4. Validate aftercare/red_flags coupling (if aftercare given, red flags required)
    // 5. Insert outcome record
    // 6. If endpoint_ids provided, insert junction rows
}
```

### API Endpoints Summary
```
# Injectable modules (same pattern as energy modules)
POST   /api/v1/sessions/{id}/modules/filler              Create filler module
GET    /api/v1/sessions/{id}/modules/filler/{moduleId}    Get filler module detail
PUT    /api/v1/sessions/{id}/modules/filler/{moduleId}    Update filler module detail

POST   /api/v1/sessions/{id}/modules/botulinum            Create botulinum module
GET    /api/v1/sessions/{id}/modules/botulinum/{moduleId}  Get botulinum module detail
PUT    /api/v1/sessions/{id}/modules/botulinum/{moduleId}  Update botulinum module detail

# Outcomes (session-level, consent-like pattern)
POST   /api/v1/sessions/{id}/outcome                      Record session outcome
GET    /api/v1/sessions/{id}/outcome                      Get session outcome
PUT    /api/v1/sessions/{id}/outcome                      Update session outcome
```

### Error Codes to Add
```go
// In apierrors package:
const (
    OutcomeNotFound        = "OUTCOME_NOT_FOUND"
    OutcomeAlreadyExists   = "OUTCOME_ALREADY_EXISTS"
    OutcomeInvalidData     = "OUTCOME_INVALID_DATA"
    OutcomeCreationFailed  = "OUTCOME_CREATION_FAILED"
    OutcomeUpdateFailed    = "OUTCOME_UPDATE_FAILED"
    ModuleProductTypeMismatch = "MODULE_PRODUCT_TYPE_MISMATCH"
    InjectableModuleCreationFailed = "INJECTABLE_MODULE_CREATION_FAILED"
)
```

### Metrics to Add
```go
// New counters:
// dermify_injectable_module_created_total
// dermify_outcome_recorded_total
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Device-centric module detail | Product-centric module detail for injectables | Phase 4 (this phase) | Different FK target, product type validation instead of device type |
| No session outcome tracking | Structured outcome with aftercare and follow-up | Phase 4 (this phase) | Completes the clinical documentation lifecycle before sign-off |
| Normalized child tables for all sub-data | JSONB for injection sites | Phase 4 (this phase) | Simpler read/write for atomically managed sub-data |

## Open Questions

1. **Injection sites: JSONB vs text vs child table**
   - What we know: Botulinum toxin requires per-site unit mapping. The data is always read/written as a unit. Phase 5 locking freezes the entire record.
   - What's unclear: Whether injection sites will ever need independent querying (e.g., "show all sessions where glabella was injected").
   - Recommendation: Use JSONB. It provides structured storage with ability to query via PostgreSQL JSON operators if needed later, without the join complexity of a child table. Define a Go struct for type safety in the application layer.

2. **Aftercare templating**
   - What we know: OUT-03 says "templated instructions." OUT-04 says "mandatory red flags."
   - What's unclear: Whether the API provides templates or just stores final text.
   - Recommendation: The API stores the final aftercare text. Templating is a frontend concern (the frontend shows templates, the clinician picks/edits, the API receives the final text). The API enforces that if aftercare_notes is provided, red_flags_text must also be provided. This keeps the API simple and avoids template management complexity.

3. **Clinical endpoint validation against session modules**
   - What we know: OUT-02 says "module-specific list." The clinical_endpoints table has a module_type column.
   - What's unclear: Whether the API should strictly enforce that selected endpoints match the module types present in the session.
   - Recommendation: Validate at the service layer. Query the session's modules to get the set of module types, then reject any endpoint whose module_type is not in that set. This prevents clinically nonsensical selections.

4. **Outcome timing relative to session state**
   - What we know: Outcomes are recorded "after" treatment. Sessions go through Draft -> In Progress -> Awaiting Sign-off -> Signed -> Locked.
   - What's unclear: At which state(s) should outcome recording be allowed?
   - Recommendation: Allow outcome recording when session is `in_progress` or `awaiting_signoff`. This lets clinicians document outcomes while still editing the session (in_progress) or after flagging it for review (awaiting_signoff). Do NOT allow on draft (no treatment has started) or signed/locked (already finalized).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (assert/require) |
| Config file | None (Go standard test runner) |
| Quick run command | `go test ./internal/service/... -count=1 -run "TestInjectable\|TestOutcome" -v` |
| Full suite command | `make test` (runs `go test ./... -count=1 -v`) |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FILL-01 | Create filler module in session | unit | `go test ./internal/service/... -count=1 -run TestCreateFillerModule -v` | Wave 0 |
| FILL-02 | Filler parameter capture | unit | `go test ./internal/service/... -count=1 -run TestFillerModule_Parameters -v` | Wave 0 |
| FILL-03 | Filler product linkage | unit | `go test ./internal/service/... -count=1 -run TestFillerModule_ProductValidation -v` | Wave 0 |
| TOX-01 | Create botulinum module | unit | `go test ./internal/service/... -count=1 -run TestCreateBotulinumModule -v` | Wave 0 |
| TOX-02 | Botulinum parameter + injection sites | unit | `go test ./internal/service/... -count=1 -run TestBotulinumModule_Parameters -v` | Wave 0 |
| TOX-03 | Botulinum product linkage | unit | `go test ./internal/service/... -count=1 -run TestBotulinumModule_ProductValidation -v` | Wave 0 |
| OUT-01 | Record outcome status | unit | `go test ./internal/service/... -count=1 -run TestRecordOutcome -v` | Wave 0 |
| OUT-02 | Select clinical endpoints | unit | `go test ./internal/service/... -count=1 -run TestOutcome_Endpoints -v` | Wave 0 |
| OUT-03 | Record aftercare instructions | unit | `go test ./internal/service/... -count=1 -run TestOutcome_Aftercare -v` | Wave 0 |
| OUT-04 | Mandatory red flags validation | unit | `go test ./internal/service/... -count=1 -run TestOutcome_RedFlagsRequired -v` | Wave 0 |
| OUT-05 | Set follow-up date | unit | `go test ./internal/service/... -count=1 -run TestOutcome_FollowUp -v` | Wave 0 |

### Cross-Cutting Tests
| Behavior | Test Type | Automated Command |
|----------|-----------|-------------------|
| Product not found returns error | unit | `go test ./internal/service/... -count=1 -run TestInjectableModule_ProductNotFound -v` |
| Product type mismatch returns error | unit | `go test ./internal/service/... -count=1 -run TestInjectableModule_ProductTypeMismatch -v` |
| Non-editable session rejected for modules | unit | `go test ./internal/service/... -count=1 -run TestInjectableModule_NonEditableSession -v` |
| Consent required before injectable modules | unit | `go test ./internal/service/... -count=1 -run TestInjectableModule_ConsentRequired -v` |
| Outcome duplicate rejected | unit | `go test ./internal/service/... -count=1 -run TestOutcome_AlreadyExists -v` |
| Outcome on draft session rejected | unit | `go test ./internal/service/... -count=1 -run TestOutcome_DraftSessionRejected -v` |
| Invalid injection sites JSON rejected | unit | `go test ./internal/service/... -count=1 -run TestBotulinumModule_InvalidInjectionSites -v` |

### Sampling Rate
- **Per task commit:** `go test ./internal/service/... -count=1 -v`
- **Per wave merge:** `make test`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/service/injectable_module_test.go` -- covers FILL-01 through TOX-03
- [ ] `internal/service/outcome_test.go` -- covers OUT-01 through OUT-05
- [ ] `internal/testutil/mock_injectable_module.go` -- mock repos for filler + botulinum
- [ ] `internal/testutil/mock_outcome.go` -- mock repo for outcome

## Sources

### Primary (HIGH confidence)
- Existing codebase analysis -- all patterns derived from reading the actual Phase 1-3 implementation
- `internal/domain/session_module.go` -- module type constants already include `filler` and `botulinum_toxin`
- `internal/domain/product.go` -- Product type with `ProductType` field (`filler` or `botulinum_toxin`)
- `internal/service/energy_module.go` -- EnergyModuleService pattern to replicate for injectables
- `internal/service/consent.go` -- ConsentService singleton pattern to replicate for outcomes
- `internal/service/registry.go` -- `GetProductByID` and `ErrProductNotFound` already exist
- `migrations/20260307150001_create_products_table.sql` -- products schema with CHECK constraint
- `migrations/20260307160001_seed_products.sql` -- seed data: 3 fillers + 3 botulinum toxins
- `migrations/20260307160002_seed_indication_codes.sql` -- clinical endpoints seeded for filler and botulinum_toxin module types
- `migrations/20260308010004_create_session_modules.sql` -- CHECK constraint already includes `'filler', 'botulinum_toxin'`
- `internal/api/apierrors/apierrors.go` -- existing error code patterns
- `internal/api/metrics/prometheus.go` -- metrics Client pattern

### Secondary (MEDIUM confidence)
- `.planning/REQUIREMENTS.md` -- FILL-01 through OUT-05 requirement text
- `.planning/STATE.md` -- accumulated decisions from Phases 1-3

### Tertiary (LOW confidence)
- Clinical parameter field names for injectables -- derived from REQUIREMENTS.md descriptions. Exact units (mL, units, concentrations) are clinical standards but column precision choices could be refined with clinical input.
- JSONB schema for injection sites -- reasonable design but exact field names ({site, units} vs {location, dose}) would benefit from clinical review.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies, all patterns exist in codebase
- Injectable modules architecture: HIGH -- direct parallel to energy modules, all infrastructure (products, session_modules, registry) already in place
- Outcomes architecture: HIGH -- follows consent singleton pattern, clinical_endpoints table and seed data already exist
- Injectable clinical parameters: MEDIUM -- field names derived from requirements text, may need clinical review
- JSONB injection sites schema: MEDIUM -- reasonable design, but exact structure is a design choice

**Research date:** 2026-03-08
**Valid until:** 2026-04-08 (stable -- project conventions unlikely to change mid-milestone)
