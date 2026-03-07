# Architecture Research

**Domain:** Clinical procedure documentation API for aesthetic dermatology
**Researched:** 2026-03-07
**Confidence:** HIGH

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          HTTP Layer (Chi)                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  Auth     │ │ Patient  │ │ Session  │ │ Device/  │ │  Photo   │ │
│  │ Handlers  │ │ Handlers │ │ Handlers │ │ Product  │ │ Handlers │ │
│  │ (exists)  │ │          │ │          │ │ Handlers │ │          │ │
│  └─────┬─────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ │
│        │             │            │            │            │      │
├────────┴─────────────┴────────────┴────────────┴────────────┴──────┤
│                        Service Layer (NEW)                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  Auth    │ │ Patient  │ │ Session  │ │ Module   │ │  Audit   │ │
│  │ Service  │ │ Service  │ │ Service  │ │ Registry │ │ Service  │ │
│  │ (exists  │ │          │ │          │ │          │ │          │ │
│  │ inline)  │ │          │ │          │ │          │ │          │ │
│  └─────┬────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ │
│        │             │            │            │            │      │
├────────┴─────────────┴────────────┴────────────┴────────────┴──────┤
│                      Repository Layer (NEW)                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  User    │ │ Patient  │ │ Session  │ │ Module   │ │  Audit   │ │
│  │  Repo    │ │  Repo    │ │  Repo    │ │  Repos   │ │  Repo    │ │
│  └─────┬────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ │
│        │             │            │            │            │      │
├────────┴─────────────┴────────────┴────────────┴────────────┴──────┤
│                    PostgreSQL (pgx v4)                              │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐      │
│  │   users    │ │  patients  │ │  sessions  │ │audit_trail │      │
│  │ (exists)   │ │            │ │ + modules  │ │            │      │
│  └────────────┘ └────────────┘ └────────────┘ └────────────┘      │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐                     │
│  │  devices   │ │  products  │ │   photos   │                     │
│  │  (seed)    │ │  (seed)    │ │ (metadata) │                     │
│  └────────────┘ └────────────┘ └────────────┘                     │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                         ┌──────────┴──────────┐
                         │  Local Filesystem    │
                         │  (photo storage)     │
                         └─────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| Handlers | HTTP decode/encode, auth extraction, call service, return response | One file per domain in `internal/api/handlers/` |
| Services | Business logic, validation, state transitions, transaction orchestration | One file per domain in `internal/service/` |
| Repositories | SQL queries, data mapping, no business logic | One file per domain in `internal/repository/` |
| Module Registry | Route procedure module creation/validation to correct type handler | Single registry mapping module_type to validator/repo |
| Audit Service | Write-only append to audit trail, hash chaining optional | Cross-cutting, called by session service |
| Middleware (Auth) | JWT validation, role extraction, RBAC enforcement | Existing pattern in `internal/api/middleware/` |

## Recommended Project Structure

```
internal/
├── api/
│   ├── handlers/           # HTTP handlers (thin - decode, delegate, encode)
│   │   ├── auth.go         # (exists) Login, register, refresh, logout
│   │   ├── patients.go     # Patient CRUD
│   │   ├── sessions.go     # Session lifecycle (create, get, list, sign-off)
│   │   ├── modules.go      # Add/update procedure modules to sessions
│   │   ├── outcomes.go     # Outcomes, adverse events, aftercare
│   │   ├── addendums.go    # Addendums on locked records
│   │   ├── devices.go      # Device/product registry (read + admin)
│   │   ├── photos.go       # Photo upload/retrieve
│   │   └── health.go       # (exists)
│   ├── middleware/          # (exists) Auth, CORS, logging, request_id
│   │   ├── auth.go         # (exists) JWT validation
│   │   └── rbac.go         # NEW: role-based access control
│   ├── routes/             # (exists) Route registration per domain
│   │   ├── manager.go      # (exists) Central route wiring
│   │   ├── auth.go         # (exists)
│   │   ├── patients.go     # Patient routes
│   │   ├── sessions.go     # Session + module + outcome routes
│   │   └── admin.go        # Device/product management routes
│   └── metrics/            # (exists) Prometheus definitions
├── service/                # NEW: Business logic layer
│   ├── patient.go          # Patient validation, CRUD logic
│   ├── session.go          # Session lifecycle, state machine, sign-off
│   ├── module.go           # Module registry, polymorphic dispatch
│   ├── validation.go       # Per-module-type conditional validation rules
│   ├── audit.go            # Audit trail writing
│   ├── consent.go          # Consent capture, contraindication checks
│   └── photo.go            # Photo storage, metadata management
├── repository/             # NEW: Data access layer
│   ├── patient.go          # Patient SQL queries
│   ├── session.go          # Session + header SQL queries
│   ├── module_ipl.go       # IPL module queries
│   ├── module_ndyag.go     # Nd:YAG module queries
│   ├── module_co2.go       # CO2/ablative module queries
│   ├── module_rf.go        # RF/microneedling module queries
│   ├── module_filler.go    # Filler module queries
│   ├── module_botox.go     # Botulinum toxin module queries
│   ├── device.go           # Device/product registry queries
│   ├── audit.go            # Audit trail queries (append-only)
│   ├── photo.go            # Photo metadata queries
│   └── db.go               # Transaction helper, DBTX interface
├── model/                  # NEW: Domain types shared across layers
│   ├── patient.go          # Patient struct
│   ├── session.go          # Session, SessionStatus enum
│   ├── module.go           # Module interface + per-type structs
│   ├── device.go           # Device, Product structs
│   ├── audit.go            # AuditEntry struct
│   └── photo.go            # Photo metadata struct
└── pkg/                    # (exists) Shared utilities
    └── postgres.go         # (exists) DB connection + migrations
```

### Structure Rationale

- **`internal/service/`:** Extracts business logic from handlers. The existing codebase has SQL directly in handlers (e.g., `HandleRegister` runs `INSERT INTO users`). Session lifecycle, module validation, and sign-off logic is too complex for handlers. Services own transaction boundaries.
- **`internal/repository/`:** Isolates SQL from business logic. Each module type gets its own repository file because the column sets differ significantly. A shared `DBTX` interface lets repositories work with either a `*sql.DB` or `*sql.Tx`, enabling services to wrap multiple repository calls in a single transaction.
- **`internal/model/`:** Shared domain types prevent circular imports between service and repository packages. Handlers, services, and repositories all reference these types. Keeps structs separate from HTTP request/response types (which stay in handlers).
- **One module repo file per type:** Each of the 6 procedure types (IPL, Nd:YAG, CO2, RF, Filler, Botulinum) has different columns. Separate files keep each manageable at ~100-200 lines rather than one 1200-line monolith.

## Architectural Patterns

### Pattern 1: Hybrid Polymorphism (Shared Header + Separate Detail Tables)

**What:** A `session_modules` table stores the shared fields common to all procedure types (session_id, module_type, treatment_area, created_at), while six separate detail tables (`module_ipl`, `module_ndyag`, `module_co2`, `module_rf`, `module_filler`, `module_botox`) store type-specific parameters. Each detail table has a foreign key back to `session_modules`.

**When to use:** When polymorphic types share some fields but have significantly different type-specific fields -- exactly the case here. IPL has wavelength/filter, pulse duration, spot size. Fillers have product volume, injection depth, cannula vs needle. These share almost nothing.

**Trade-offs:**
- PRO: Full referential integrity via foreign keys (unlike JSONB or single-table approaches)
- PRO: PostgreSQL query planner has proper column statistics for each table
- PRO: Adding a new procedure type means adding a new table, not altering existing ones
- PRO: Each table can have strict NOT NULL constraints appropriate to its type
- CON: Queries that list "all modules for a session" require LEFT JOINs or UNION queries
- CON: More tables to manage in migrations

**Why not alternatives:**
- Single table with nullable columns: 6 types with 8-15 unique fields each = 60+ nullable columns. Unmaintainable.
- JSONB for type-specific fields: Loses column statistics, cannot enforce NOT NULL on nested fields, harder to query and index. PostgreSQL's JSONB query planner flies blind without statistics.
- PostgreSQL table inheritance: Cannot enforce foreign keys between parent and child tables. The PostgreSQL docs themselves discourage this for production data modeling.

**Schema sketch:**

```sql
CREATE TABLE session_modules (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id),
    module_type VARCHAR(20) NOT NULL CHECK (module_type IN (
        'ipl', 'ndyag', 'co2', 'rf', 'filler', 'botox'
    )),
    treatment_area  TEXT NOT NULL,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT NOT NULL REFERENCES users(id)
);

-- Example: IPL-specific detail table
CREATE TABLE module_ipl (
    id              BIGSERIAL PRIMARY KEY,
    session_module_id BIGINT NOT NULL UNIQUE REFERENCES session_modules(id),
    device_id       BIGINT NOT NULL REFERENCES devices(id),
    wavelength_nm   INTEGER NOT NULL,
    filter_nm       INTEGER,
    fluence_j_cm2   DECIMAL(6,2) NOT NULL,
    pulse_duration_ms DECIMAL(8,2) NOT NULL,
    spot_size_mm    DECIMAL(5,2) NOT NULL,
    cooling_type    VARCHAR(50),
    pass_count      INTEGER NOT NULL DEFAULT 1,
    endpoint_reached TEXT
);

-- Example: Filler-specific detail table (very different fields)
CREATE TABLE module_filler (
    id              BIGSERIAL PRIMARY KEY,
    session_module_id BIGINT NOT NULL UNIQUE REFERENCES session_modules(id),
    product_id      BIGINT NOT NULL REFERENCES products(id),
    batch_number    VARCHAR(100) NOT NULL,
    expiry_date     DATE NOT NULL,
    volume_ml       DECIMAL(5,2) NOT NULL,
    injection_depth VARCHAR(50) NOT NULL,
    technique       VARCHAR(50) NOT NULL CHECK (technique IN ('needle', 'cannula')),
    cannula_gauge   INTEGER,
    needle_gauge    INTEGER,
    aspiration_performed BOOLEAN NOT NULL DEFAULT TRUE
);
```

**Go interface for polymorphic dispatch:**

```go
// ModuleValidator validates type-specific fields
type ModuleValidator interface {
    ValidateForSignOff() []ValidationError
    ModuleType() string
}

// ModuleRepository handles type-specific persistence
type ModuleRepository interface {
    Create(ctx context.Context, tx DBTX, moduleID int64, data interface{}) error
    GetByModuleID(ctx context.Context, db DBTX, moduleID int64) (interface{}, error)
}
```

### Pattern 2: Session Lifecycle State Machine

**What:** Sessions follow a strict state machine enforced at both the application layer (service) and the database layer (CHECK constraint + transition validation). States flow: `draft` -> `in_progress` -> `pending_review` -> `signed` -> `locked`. The `locked` state is terminal -- no further transitions allowed.

**When to use:** Any workflow where records pass through defined stages and certain operations are only valid in certain states. Treatment sessions are the canonical case.

**Trade-offs:**
- PRO: Prevents invalid state transitions (cannot go from `draft` to `locked`)
- PRO: Database-level enforcement catches bugs the application layer misses
- PRO: State stored as a simple VARCHAR column, queried efficiently
- CON: State machine logic must be duplicated (Go service + SQL constraint)

**State machine definition:**

```
                ┌──────────┐
                │  draft   │  Session created, header fields filled
                └────┬─────┘
                     │ AddModule / UpdateHeader
                     v
              ┌──────────────┐
              │ in_progress  │  Modules being added, data entry active
              └──────┬───────┘
                     │ RequestReview (all required checks pass)
                     v
            ┌────────────────┐
            │ pending_review │  Ready for clinician sign-off
            └───────┬────────┘
                    │ SignOff (validation gate passes)
                    v
              ┌──────────┐
              │  signed  │  Clinician has signed, short grace period
              └────┬─────┘
                   │ Lock (automatic or manual, after grace period)
                   v
              ┌──────────┐
              │  locked  │  Immutable. Addendums only.
              └──────────┘
```

**Enforcement in SQL:**

```sql
-- Transition validation via UPDATE ... WHERE state = expected
UPDATE sessions
SET status = 'signed', signed_at = NOW(), signed_by = $2
WHERE id = $1 AND status = 'pending_review'
RETURNING id;
-- Returns 0 rows if session wasn't in pending_review = transition rejected
```

**Enforcement in Go service:**

```go
var validTransitions = map[string][]string{
    "draft":          {"in_progress"},
    "in_progress":    {"pending_review"},
    "pending_review": {"signed"},
    "signed":         {"locked"},
    "locked":         {}, // terminal state
}

func (s *SessionService) TransitionState(ctx context.Context, sessionID int64,
    targetState string) error {
    session, err := s.repo.GetByID(ctx, sessionID)
    if err != nil {
        return err
    }
    allowed := validTransitions[session.Status]
    if !slices.Contains(allowed, targetState) {
        return fmt.Errorf("cannot transition from %s to %s", session.Status, targetState)
    }
    // ... proceed with transition + audit entry
}
```

### Pattern 3: Immutable Record with Addendum-Only Amendments

**What:** Once a session reaches `locked` status, the original record cannot be modified. Any corrections or additions are stored as separate `addendum` records linked to the session. Each addendum captures who wrote it, when, and the content. The original session row and all its module rows become read-only.

**When to use:** Medico-legal compliance. Medical records must preserve the original documentation. Corrections happen via addendums that are themselves immutable once created.

**Trade-offs:**
- PRO: Original record is always recoverable -- no "what did it say before?" questions
- PRO: Meets medico-legal requirements for clinical documentation
- PRO: Simple to implement -- just prevent UPDATEs on locked records
- CON: Cannot fix typos in-place; addendum must reference what was wrong

**Implementation layers:**

1. **Database triggers:** Prevent UPDATE/DELETE on sessions and modules where status = 'locked'
2. **Application guards:** Service layer checks status before any mutation
3. **Addendum table:** Separate append-only table

```sql
-- Prevent modification of locked sessions
CREATE OR REPLACE FUNCTION prevent_locked_session_update()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status = 'locked' AND NEW.status = 'locked' THEN
        RAISE EXCEPTION 'Cannot modify a locked session record';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_session_immutability
BEFORE UPDATE ON sessions
FOR EACH ROW
EXECUTE FUNCTION prevent_locked_session_update();

-- Addendum table
CREATE TABLE session_addendums (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id),
    content     TEXT NOT NULL,
    reason      TEXT NOT NULL,  -- why the addendum was needed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  BIGINT NOT NULL REFERENCES users(id)
);
-- No UPDATE or DELETE operations exposed for addendums
```

### Pattern 4: Conditional Validation via Module Registry

**What:** A central module registry maps each `module_type` string to its validator, repository, and request decoder. When a module is created or a session is submitted for sign-off, the registry dispatches to the correct type-specific validator. Each validator enforces its own required fields and conditional rules.

**When to use:** When you have N procedure types with different validation rules and need a single entry point that routes to the correct logic.

**Trade-offs:**
- PRO: Adding a new procedure type = implement the interface + register it
- PRO: Sign-off validation is a single loop over all modules, each validating itself
- CON: Registry must be kept in sync with database CHECK constraints

```go
// registry.go
type ModuleHandler struct {
    Validator  func(data json.RawMessage) (ModuleValidator, error)
    Repository ModuleRepository
}

var moduleRegistry = map[string]ModuleHandler{
    "ipl":    {Validator: decodeIPL, Repository: &IPLRepo{}},
    "ndyag":  {Validator: decodeNdYAG, Repository: &NdYAGRepo{}},
    "co2":    {Validator: decodeCO2, Repository: &CO2Repo{}},
    "rf":     {Validator: decodeRF, Repository: &RFRepo{}},
    "filler": {Validator: decodeFiller, Repository: &FillerRepo{}},
    "botox":  {Validator: decodeBotox, Repository: &BotoxRepo{}},
}

// Example: Filler-specific conditional validation
// cannula_gauge is required only when technique = "cannula"
func (f *FillerModule) ValidateForSignOff() []ValidationError {
    var errs []ValidationError
    if f.Volume <= 0 {
        errs = append(errs, ValidationError{Field: "volume_ml", Msg: "must be positive"})
    }
    if f.Technique == "cannula" && f.CannulaGauge == nil {
        errs = append(errs, ValidationError{Field: "cannula_gauge",
            Msg: "required when technique is cannula"})
    }
    if f.Technique == "needle" && f.NeedleGauge == nil {
        errs = append(errs, ValidationError{Field: "needle_gauge",
            Msg: "required when technique is needle"})
    }
    return errs
}
```

### Pattern 5: Audit Trail as Append-Only Event Log

**What:** Every state change, data modification, and sign-off action writes an entry to an `audit_trail` table. Entries are append-only (INSERT only, never UPDATE or DELETE). Each entry records who, what, when, the entity affected, and a snapshot of what changed.

**When to use:** Any system requiring medico-legal traceability. This is not optional for clinical documentation.

**Trade-offs:**
- PRO: Complete history of every change, who made it, and when
- PRO: Simple implementation -- just INSERT on every mutation
- PRO: No need for event sourcing complexity; this is a side-effect log, not the source of truth
- CON: Table grows indefinitely (acceptable for single-clinic deployment)
- CON: Must be disciplined about writing audit entries everywhere

**Schema:**

```sql
CREATE TABLE audit_trail (
    id          BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,  -- 'session', 'module', 'patient', etc.
    entity_id   BIGINT NOT NULL,
    action      VARCHAR(50) NOT NULL,  -- 'created', 'updated', 'signed', 'locked', 'addendum'
    actor_id    BIGINT NOT NULL REFERENCES users(id),
    details     JSONB,                 -- snapshot of what changed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for looking up history of a specific entity
CREATE INDEX idx_audit_trail_entity ON audit_trail(entity_type, entity_id);
-- Index for looking up all actions by a specific user
CREATE INDEX idx_audit_trail_actor ON audit_trail(actor_id);
```

**Note on cryptographic chaining:** For a single-tenant, single-clinic deployment where the database admin is the clinic itself, hash chaining adds complexity without meaningful security benefit. The simpler append-only model with database-level access controls (revoke UPDATE/DELETE on audit_trail from the application role) is sufficient. If regulatory requirements change, hash chaining can be retrofitted.

## Data Flow

### Request Flow (Standard CRUD)

```
[HTTP Request]
    │
    v
[Chi Router] ──> [Auth Middleware] ──> [RBAC Middleware]
                                           │
                                           v
                                      [Handler]
                                           │ decode request body
                                           │ extract user claims from context
                                           v
                                      [Service]
                                           │ validate business rules
                                           │ begin transaction
                                           v
                                    [Repository]
                                           │ execute SQL
                                           v
                                    [PostgreSQL]
                                           │
                                    [Repository] returns model
                                           │
                                    [Service] writes audit entry
                                           │ commit transaction
                                           v
                                      [Handler]
                                           │ encode response
                                           v
                                   [HTTP Response]
```

### Session Sign-Off Flow (Most Complex)

```
[POST /api/v1/sessions/{id}/sign-off]
    │
    v
[Handler] ──> extract clinician ID from JWT claims
    │
    v
[SessionService.SignOff(ctx, sessionID, clinicianID)]
    │
    ├── 1. Load session (check status = 'pending_review')
    │
    ├── 2. Load all modules for session
    │
    ├── 3. For each module:
    │       ├── Look up ModuleHandler from registry by type
    │       ├── Load type-specific data from detail table
    │       └── Call ValidateForSignOff()
    │
    ├── 4. Check consent captured
    │
    ├── 5. Check contraindication screening completed
    │
    ├── 6. If any validation errors ──> return 422 with list of issues
    │
    ├── 7. BEGIN TRANSACTION
    │       ├── UPDATE session status to 'signed', set signed_at, signed_by
    │       ├── INSERT audit_trail entry (action: 'signed')
    │       └── COMMIT
    │
    └── 8. Return signed session
```

### Photo Upload Flow

```
[POST /api/v1/sessions/{id}/photos]
    │
    v
[Handler] ──> parse multipart form, validate file type/size
    │
    v
[PhotoService]
    │
    ├── 1. Verify session exists and is NOT locked
    ├── 2. Generate unique filename (UUID + extension)
    ├── 3. Write file to local filesystem ({base_path}/{session_id}/{filename})
    ├── 4. INSERT photo metadata into database (path, type, label, session_id)
    └── 5. Write audit entry
```

### Key Data Flows

1. **Session creation:** Handler -> SessionService (validate patient exists, create session in draft status, write audit entry) -> SessionRepo (INSERT session) -> Response with session ID
2. **Module addition:** Handler -> SessionService (verify session is draft/in_progress, transition to in_progress if draft) -> ModuleRegistry (dispatch to correct type) -> ModuleRepo (INSERT into shared + detail table) -> AuditService (log module addition)
3. **Record locking:** SessionService (verify status = signed) -> SessionRepo (UPDATE to locked) -> Trigger prevents future UPDATEs -> AuditService (log lock event)
4. **Addendum creation:** Handler -> SessionService (verify session IS locked) -> AddendumRepo (INSERT addendum) -> AuditService (log addendum)

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1 clinic (target) | Current design: single PostgreSQL, local filesystem, no caching needed. Sessions table will have hundreds to low thousands of rows per year. No optimization required. |
| 5-10 clinics (separate instances) | Each gets its own deployment. Consider a shared Docker image registry and deployment automation. Database per instance means zero cross-contamination. |
| 50+ clinics | If consolidation is ever needed, the single-tenant model becomes operationally expensive. Would need multi-tenant redesign with tenant_id columns. Out of scope for v1. |

### Scaling Priorities

1. **First bottleneck: Photo storage.** Local filesystem works for one clinic. If photo volume grows or backups become painful, migrate to S3-compatible storage. The photo service layer already abstracts the storage backend -- swap the implementation without changing handlers.
2. **Second bottleneck: Audit trail table size.** Append-only audit logs grow indefinitely. For a single clinic doing 20-50 sessions/week, this is years before it matters. If needed, partition by created_at or archive old entries.

## Anti-Patterns

### Anti-Pattern 1: Fat Handlers with Inline SQL

**What people do:** Continue the existing pattern of writing SQL directly in handler functions (as seen in `HandleRegister` today).
**Why it's wrong:** Session sign-off requires loading a session, loading all its modules across 6 different tables, running conditional validation on each, checking consent status, and writing an audit trail -- all in a single transaction. Putting this in a handler creates untestable 200+ line functions. Transaction management in handlers also makes it impossible to reuse business logic.
**Do this instead:** Introduce the service layer now, before building session logic. Handlers decode requests and encode responses. Services own business rules and transactions. Repositories execute SQL.

### Anti-Pattern 2: Single JSONB Column for All Module Data

**What people do:** Store all procedure-specific fields as a single JSONB column on `session_modules` to avoid creating 6 detail tables.
**Why it's wrong:** Loses PostgreSQL query planner statistics (it cannot see inside JSONB). Cannot enforce NOT NULL on type-specific required fields. Cannot create proper foreign keys from device_id/product_id inside JSONB to reference tables. Updates rewrite the entire JSONB value. Makes conditional validation a JSON schema problem instead of a Go struct problem.
**Do this instead:** Use the hybrid model: shared `session_modules` table + separate detail tables per type. Yes, it is more tables. But each table has proper types, constraints, and statistics. The extra migration effort is a one-time cost; the query and validation benefits are permanent.

### Anti-Pattern 3: Soft Deletes Instead of True Immutability

**What people do:** Add `deleted_at` columns and filter them in queries instead of actually preventing mutation.
**Why it's wrong:** Soft deletes still allow UPDATE/DELETE at the database level. A bug or direct SQL access can silently modify "locked" records. For medico-legal compliance, the defense is "the database physically prevents modification," not "our application code filters deleted records."
**Do this instead:** Use database triggers to reject UPDATEs on locked records. Revoke DELETE permission on session-related tables from the application database role. The application user should only have INSERT/SELECT on audit_trail.

### Anti-Pattern 4: Validating Only at the Handler Level

**What people do:** Put all validation in the HTTP handler using struct tags (`required`, `min`, `max`).
**Why it's wrong:** Struct tag validation handles field presence and basic type constraints, but cannot express rules like "cannula_gauge is required when technique = cannula" or "session cannot be signed unless at least one module exists and consent is captured." These are business rules that belong in the service layer.
**Do this instead:** Use struct tags for basic field validation (non-empty, valid ranges) at the handler level. Use service-layer validation for cross-field rules, cross-entity rules (consent exists, modules exist), and state-dependent rules (cannot add modules to locked session).

### Anti-Pattern 5: Shared Module Table with Type Column and Nullable Everything

**What people do:** Create one `procedure_data` table with columns for every possible field across all 6 types, using a `type` discriminator column.
**Why it's wrong:** Results in 60+ columns where any given row uses only 10-15. Impossible to tell which fields are actually required for which type from the schema alone. CHECK constraints become a maze of `(type = 'ipl' AND wavelength IS NOT NULL AND fluence IS NOT NULL) OR (type = 'filler' AND product_id IS NOT NULL AND volume IS NOT NULL) OR ...` for every type. New developers cannot understand the schema.
**Do this instead:** Separate detail tables. Each table is self-documenting: its columns are exactly the fields that type needs, with appropriate NOT NULL constraints.

## Integration Points

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| Handler <-> Service | Direct function calls | Service methods accept context + typed params, return model types + errors |
| Service <-> Repository | Direct function calls via interface | Repository interface defined in service package, implemented in repository package. Enables testing with mocks. |
| Service <-> Audit | Direct function call within same transaction | Audit writes happen inside the service transaction. If the business operation fails, the audit entry rolls back too. |
| Session Service <-> Module Registry | Registry lookup by module_type string | Returns type-specific validator and repository. Service does not import individual module packages. |
| Photo Handler <-> Filesystem | `os.Create` / `io.Copy` | Abstracted behind a `PhotoStorage` interface for future S3 migration |

### DBTX Interface (Critical Pattern)

The repository layer needs to work with both `*sql.DB` (normal queries) and `*sql.Tx` (within transactions). Define a shared interface:

```go
// internal/repository/db.go
type DBTX interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
```

Services start transactions and pass the `*sql.Tx` (which satisfies `DBTX`) to repository methods. This keeps transaction management in the service layer where it belongs.

## Suggested Build Order

The dependency graph dictates the build order. Each phase builds on the previous one.

### Phase 1: Foundation (Service/Repository Layer + RBAC + Patients)

Must be built first because everything else depends on these patterns being established.

**Dependencies:** Existing auth system, existing user table.

1. Introduce `internal/repository/db.go` with DBTX interface
2. Introduce `internal/model/` package with base types
3. Add role column to users table (doctor/admin), RBAC middleware
4. Patient CRUD (first domain to use the new service/repository pattern -- establishes the pattern for all subsequent domains)
5. Seed data migration for devices and products

**Rationale:** Patients must exist before sessions can reference them. The service/repository pattern must exist before session logic can be built properly. RBAC must exist before any endpoint access rules can be enforced. Device/product seed data must exist before modules can reference them.

### Phase 2: Session Lifecycle Core

**Dependencies:** Phase 1 (patients, RBAC, service layer pattern, device/product seed data).

1. Sessions table with state machine (draft -> in_progress -> pending_review -> signed -> locked)
2. Session CRUD (create, get, list, update header fields)
3. `session_modules` shared table
4. Module registry scaffolding
5. First 2 module types implemented (e.g., IPL + Filler -- one energy-based, one injectable, to prove the polymorphic pattern works for both categories)
6. Audit trail table and service
7. State transition enforcement (application + database trigger)

**Rationale:** Building 2 module types first validates the polymorphic architecture before committing to all 6. Choose one energy-based (IPL) and one injectable (Filler) because they have the most different field sets. If the pattern works for these two, it works for the rest.

### Phase 3: Remaining Modules + Clinical Workflow

**Dependencies:** Phase 2 (proven module pattern, session lifecycle, audit trail).

1. Remaining 4 module types (Nd:YAG, CO2, RF, Botulinum)
2. Consent capture and contraindication screening
3. Outcome recording (immediate outcomes, clinical endpoints)
4. Adverse event capture
5. Aftercare documentation
6. Sign-off validation gate (all required checks must pass)

**Rationale:** Once the pattern is proven with 2 modules, adding 4 more is mechanical. Clinical workflow features (consent, outcomes, adverse events) are needed before sign-off validation can be implemented.

### Phase 4: Immutability, Photos, Polish

**Dependencies:** Phase 3 (sign-off working, all modules, clinical workflow complete).

1. Record locking (signed -> locked transition)
2. Database triggers preventing mutation of locked records
3. Addendum system for locked records
4. Photo upload/retrieval with local filesystem storage
5. Follow-up scheduling
6. Integration tests covering full session flow (create -> modules -> consent -> sign-off -> lock -> addendum)

**Rationale:** Immutability enforcement is the final layer -- it wraps everything built in phases 2-3. Photos and follow-ups are independent features that don't block the core flow. Integration tests validate the entire chain end-to-end.

## Sources

- [Choosing a Database Schema for Polymorphic Data (DoltHub)](https://www.dolthub.com/blog/2024-06-25-polymorphic-associations/)
- [Modeling Polymorphic Associations in a Relational Database (Hashrocket)](https://hashrocket.com/blog/posts/modeling-polymorphic-associations-in-a-relational-database)
- [Implementing State Machines in PostgreSQL (Felix Geisendorfer)](https://felixge.de/2017/07/27/implementing-state-machines-in-postgresql/)
- [Use your database to power state machines (Lawrence Jones)](https://blog.lawrencejones.dev/state-machines/)
- [Immutable by Design: Building Tamper-Proof Audit Logs for Health SaaS (DEV)](https://dev.to/beck_moulton/immutable-by-design-building-tamper-proof-audit-logs-for-health-saas-22dc)
- [Immutable Audit Trails: A Complete Guide (HubiFi)](https://www.hubifi.com/blog/immutable-audit-log-basics)
- [The Repository Pattern in Go (Three Dots Labs)](https://threedots.tech/post/repository-pattern-in-go/)
- [Database Transactions in Go with Layered Architecture (Three Dots Labs)](https://threedots.tech/post/database-transactions-in-go/)
- [The Fat Service Pattern for Go Web Applications (Alex Edwards)](https://www.alexedwards.net/blog/the-fat-service-pattern)
- [When to Avoid JSONB in a PostgreSQL Schema (Heap)](https://www.heap.io/blog/when-to-avoid-jsonb-in-a-postgresql-schema)
- [go-playground/validator (GitHub)](https://github.com/go-playground/validator)
- [Architectural Patterns for Health Information Systems (Frontiers)](https://www.frontiersin.org/journals/digital-health/articles/10.3389/fdgth.2025.1694839/full)

---
*Architecture research for: Dermify clinical procedure documentation API*
*Researched: 2026-03-07*
