# Phase 1: Foundation - Research

**Researched:** 2026-03-07
**Domain:** Service/repository architecture, RBAC, patient management, device/product registry, metadata tracking
**Confidence:** HIGH

## Summary

Phase 1 introduces the service/repository layered architecture into an existing Go 1.23 / Chi v5 REST API that already has JWT authentication, user registration/login, Goose migrations, and structured error handling. The codebase currently has handlers that talk directly to `database/sql` -- this phase creates the abstraction layers (repository interfaces, service structs) and builds three new domains on top of them: role-based access control (adding a `role` column to users and a `RequireRole` middleware), patient management (full CRUD with search/filter/pagination), and a read-only device/product registry (populated via seed migrations).

The existing patterns are well-established and consistent: handler closures with `Handle{Action}(db, cfg, metrics)`, route Manager pattern, `apierrors.WriteError()` for structured errors, embedded Goose SQL migrations, and strict golangci-lint (~60 linters). The phase must preserve all of these while introducing the service/repository layer beneath handlers. The key challenge is making the transition clean -- introducing interfaces that handlers depend on rather than raw `*sql.DB`, while keeping the strict linter happy (no globals, exhaustive struct init, max 100-line functions, no naked returns).

**Primary recommendation:** Introduce repository interfaces per domain (defined alongside the service that uses them), concrete PostgreSQL implementations, and service structs that handlers receive via closure injection. Use the existing `*sql.DB` for all data access (no ORM). Add a `role` field to JWT claims and create a `RequireRole(roles ...string)` middleware that wraps `RequireAuth`. Seed registry data via a single Goose SQL migration.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Admin assigns roles -- users register without a role, an admin assigns Doctor or Admin afterward
- Single role per user (no multi-role)
- First registered user is auto-promoted to Admin (bootstrap mechanism)
- Unassigned users can authenticate (login, refresh, me) but are blocked from all clinical and admin endpoints with 403
- Patient demographics: first_name, last_name, phone, email, sex (Male/Female/Other), date of birth (required), optional external_reference
- Registry is strictly read-only in v1 -- no add/edit/delete endpoints
- Real-world device names and manufacturers (Lumenis M22, Candela GentleMax Pro, etc.)
- 2-3 devices per energy-based type (IPL, Nd:YAG, CO2, RF) and 2-3 products per injectable type (filler, botulinum toxin) -- ~15-20 entries
- Curated indication codes and clinical endpoints grouped per module type (5-10 per type)
- Patient search: name (partial/prefix match on first or last name) and phone/email
- Offset-based pagination: ?page=&per_page= with total count in response
- Default sort: last_name ascending (A-Z)
- List response includes session_count and last_session_date alongside demographics

### Claude's Discretion
- Service/repository layer organization and interface design
- Database schema details (column types, indexes, constraints)
- Seed data migration structure (single migration vs per-category)
- Search implementation approach (ILIKE, trigram, etc.)
- Error code naming for new domains (patients, registry, roles)
- Swagger annotation details

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| RBAC-01 | System supports Doctor and Admin roles with distinct permissions | Add `role` column to users table, `RequireRole` middleware, role in JWT claims |
| RBAC-02 | Doctor can perform all clinical operations | Route groups with `RequireRole("doctor")` middleware |
| RBAC-03 | Admin can manage patients and view sessions but cannot sign off or modify clinical data | Route groups with `RequireRole("admin")` middleware, separate permission sets |
| RBAC-04 | Endpoints enforce role-based authorization via middleware | Chi `r.Group()` with `RequireRole` middleware per route group |
| PAT-01 | User can create a patient record with demographics | Patient repository + service, migration for patients table |
| PAT-02 | User can search and list patients with filtering | LOWER+text_pattern_ops index, ILIKE prefix search, offset pagination |
| PAT-03 | User can update patient records | Patient service with version increment, updated_by tracking |
| PAT-04 | User can view a patient's session history | Patient endpoint returning empty session list (sessions not yet built) |
| REG-01 | System ships with seed data for energy-based devices | Goose seed migration with real device data |
| REG-02 | System ships with seed data for injectable products | Goose seed migration with real product data |
| REG-03 | System ships with seed data for indication codes and clinical endpoints | Goose seed migration with curated codes per module type |
| REG-04 | Clinician can select devices and products from controlled lists | Read-only list/detail endpoints for devices, products, handpieces |
| META-01 | All clinical records track created_at, created_by, updated_at, updated_by | Database columns + application-layer population from auth context |
| META-02 | Signed records track signed_at, signed_by | Schema columns (populated in Phase 5 when sign-off is implemented) |
| META-03 | Records maintain an incrementing version number | `version` INTEGER DEFAULT 1, incremented on each UPDATE |
</phase_requirements>

## Standard Stack

### Core (already in go.mod)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-chi/chi/v5 | 5.0.10 | HTTP router | Already in use, group/middleware pattern for RBAC |
| jackc/pgx/v4 | 4.18.3 | PostgreSQL driver via database/sql | Already in use, stable for this phase |
| pressly/goose/v3 | 3.26.0 | Database migrations | Already in use with embedded SQL |
| golang-jwt/jwt/v4 | 4.5.2 | JWT tokens | Already in use, add role claim |
| google/uuid | 1.6.0 | UUID generation | Already in go.mod, use for patient external IDs if needed |
| stretchr/testify | 1.11.0 | Test assertions | Already in use |
| prometheus/client_golang | 1.15.1 | Metrics | Already in use |

### Supporting (no new dependencies needed)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| database/sql (stdlib) | Go 1.23 | DB interface | All repository implementations use this |
| encoding/json (stdlib) | Go 1.23 | JSON encode/decode | Request/response serialization |
| log/slog (stdlib) | Go 1.23 | Structured logging | All new code |
| context (stdlib) | Go 1.23 | Request-scoped values | Pass user claims, request ID |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Raw SQL via database/sql | sqlc or sqlx | sqlc generates type-safe code from SQL but adds a build step; sqlx simplifies scanning but adds dependency. Raw SQL is consistent with existing codebase and keeps dependency count low. Stick with raw SQL. |
| Application-level metadata | Database triggers for updated_at | Triggers are standard for `updated_at` timestamps but `created_by`/`updated_by` require session-level user context which PostgreSQL does not have natively. Use application-level population for all metadata columns for consistency. |
| pg_trgm for search | ILIKE with LOWER index | Trigram is more powerful (fuzzy, infix) but adds a PostgreSQL extension dependency. For prefix/partial name matching on a small patient dataset, ILIKE with a functional index is sufficient. |

**Installation:**
```bash
# No new dependencies required -- all libraries already in go.mod
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
  api/
    handlers/
      auth.go           # Existing (modify to include role in JWT)
      patients.go       # NEW: patient CRUD handlers
      registry.go       # NEW: device/product list handlers
      roles.go          # NEW: role assignment handler
      models.go         # Extend with new response types
    middleware/
      auth.go           # Existing (extend with RequireRole)
    routes/
      manager.go        # Extend with new route groups
      auth.go           # Existing
      patients.go       # NEW: patient routes
      registry.go       # NEW: registry routes
      roles.go          # NEW: role routes
    apierrors/
      apierrors.go      # Extend with new error codes
    metrics/
      metrics.go        # Extend with new counters
  service/
    patient.go          # NEW: patient business logic + repository interface
    registry.go         # NEW: registry service + repository interface
    role.go             # NEW: role service + repository interface
  repository/
    postgres/
      patient.go        # NEW: PostgreSQL patient repository
      registry.go       # NEW: PostgreSQL registry repository
      role.go           # NEW: PostgreSQL role repository
  domain/
    patient.go          # NEW: Patient domain model
    device.go           # NEW: Device, Handpiece domain models
    product.go          # NEW: Product domain model
    registry.go         # NEW: IndicationCode, ClinicalEndpoint models
    role.go             # NEW: Role constants and types
migrations/
  20260307..._add_role_to_users.sql
  20260307..._create_patients_table.sql
  20260307..._create_devices_tables.sql
  20260307..._create_products_table.sql
  20260307..._create_indication_codes.sql
  20260307..._seed_devices.sql
  20260307..._seed_products.sql
  20260307..._seed_indication_codes.sql
```

### Pattern 1: Service/Repository Layer with Interface-Based Injection

**What:** Each domain defines a repository interface in the service package. The service struct depends on the interface. Concrete PostgreSQL implementations live in `internal/repository/postgres/`. Handlers receive service structs via closure injection.

**When to use:** All new data access code. Existing auth code can remain as-is for now (refactoring it is not in scope).

**Example:**
```go
// internal/service/patient.go
package service

import (
    "context"
    "dermify-api/internal/domain"
)

// PatientRepository defines the data access contract for patients.
type PatientRepository interface {
    Create(ctx context.Context, patient *domain.Patient) error
    GetByID(ctx context.Context, id int64) (*domain.Patient, error)
    Update(ctx context.Context, patient *domain.Patient) error
    List(ctx context.Context, filter PatientFilter) (*PatientListResult, error)
    GetSessionHistory(ctx context.Context, patientID int64) ([]domain.SessionSummary, error)
}

// PatientService handles patient business logic.
type PatientService struct {
    repo PatientRepository
}

// NewPatientService creates a new PatientService.
func NewPatientService(repo PatientRepository) *PatientService {
    return &PatientService{repo: repo}
}
```

### Pattern 2: RequireRole Middleware via Chi Groups

**What:** A `RequireRole(roles ...string)` middleware that reads the authenticated user's role from JWT claims in context and returns 403 if the role is not in the allowed set. Applied per chi route group.

**When to use:** Every protected endpoint beyond basic auth.

**Example:**
```go
// internal/api/middleware/auth.go (extend existing file)

// RequireRole returns middleware that checks whether the authenticated user
// has one of the specified roles. Must be used after RequireAuth.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := GetUserClaims(r.Context())
            if claims == nil {
                apierrors.WriteError(w, http.StatusUnauthorized,
                    apierrors.AuthNotAuthenticated, "not authenticated")
                return
            }
            for _, role := range allowed {
                if claims.Role == role {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            apierrors.WriteError(w, http.StatusForbidden,
                apierrors.AuthInsufficientRole, "insufficient role permissions")
        })
    }
}
```

### Pattern 3: Handler Closure with Service Injection

**What:** Handlers receive services (not `*sql.DB`) as dependencies via the closure pattern already established in the codebase.

**When to use:** All new handlers in this phase.

**Example:**
```go
// internal/api/handlers/patients.go

// HandleCreatePatient creates a new patient record.
func HandleCreatePatient(
    svc *service.PatientService,
    m *metrics.Client,
) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract user claims for created_by
        claims := middleware.GetUserClaims(r.Context())
        // Parse request, call svc.Create(), return response
    }
}
```

### Pattern 4: Offset Pagination Response Envelope

**What:** Standard pagination response with total count, current page, per-page size.

**When to use:** Patient list endpoint, registry list endpoints.

**Example:**
```go
// internal/api/handlers/models.go

// PaginatedResponse wraps a list response with pagination metadata.
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int         `json:"total"`
    Page       int         `json:"page"`
    PerPage    int         `json:"per_page"`
    TotalPages int         `json:"total_pages"`
}
```

### Anti-Patterns to Avoid
- **Passing `*sql.DB` directly to handlers:** This is the existing pattern but new handlers should receive services. Do not propagate the anti-pattern.
- **God service that handles multiple domains:** Each domain (patient, registry, role) gets its own service and repository. Do not combine them.
- **Repository interfaces in the repository package:** Define repository interfaces in the `service` package (where they are consumed), not alongside implementations. This follows the dependency inversion principle.
- **Storing role as a separate table for single-role-per-user:** Unnecessary complexity. A `role` VARCHAR column on the `users` table with a CHECK constraint is sufficient for the single-role-per-user model.
- **Using `init()` or global variables:** The golangci-lint config enforces `gochecknoglobals` and `gochecknoinits`. All initialization must be explicit.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT token generation/validation | Custom token library | `golang-jwt/jwt/v4` (already in use) | Cryptographic correctness, standard claims handling |
| Password hashing | Custom hash function | `golang.org/x/crypto/bcrypt` (already in use) | Industry-standard, timing-safe comparison |
| UUID generation | Custom UUID function | `google/uuid` (already in go.mod) | RFC 4122 compliant, fast |
| SQL query building for search | String concatenation | Parameterized queries with `$1, $2` placeholders | SQL injection prevention |
| Pagination math | Manual offset/limit calculation | Helper function computing offset from page+per_page | Off-by-one errors, consistent across endpoints |
| Structured error responses | Ad-hoc JSON error formatting | `apierrors.WriteError()` (already in use) | Consistent machine-readable codes across all endpoints |

**Key insight:** The existing codebase already has all foundational libraries. This phase adds no new dependencies -- only new code using existing tools.

## Common Pitfalls

### Pitfall 1: JWT Claims Backward Compatibility
**What goes wrong:** Adding a `Role` field to the `auth.Claims` struct breaks existing tokens that were issued without a role. Users with valid pre-migration tokens get parse errors.
**Why it happens:** `golang-jwt/jwt` uses strict claim parsing by default.
**How to avoid:** Make the `Role` field a string with `json:"role,omitempty"`. Existing tokens without a role claim will parse successfully with an empty string `Role` field. The `RequireRole` middleware treats empty role as "no role assigned" and returns 403. Auth-only endpoints (login, refresh, me) do not use `RequireRole` so they work for all users.
**Warning signs:** 401 errors after migration on previously valid tokens.

### Pitfall 2: First-User Bootstrap Race Condition
**What goes wrong:** Two users register simultaneously, both see zero existing users, both get auto-promoted to Admin.
**Why it happens:** Check-then-insert without serialization.
**How to avoid:** Use a database-level approach: count users inside the INSERT transaction or use a serializable check. The simplest approach is a single SQL statement: after inserting the new user, check if the user count is exactly 1, and if so, set their role to 'admin'. Alternatively, use `SELECT COUNT(*) FROM users FOR UPDATE` inside a transaction before deciding.
**Warning signs:** Multiple admin accounts in a fresh deployment.

### Pitfall 3: golangci-lint Strict Mode Violations
**What goes wrong:** New code fails `make lint` due to unfamiliar linter rules: `gochecknoglobals` rejects package-level vars, `funlen` rejects functions over 100 lines, `goconst` flags repeated string literals, `exhaustive` requires all switch cases, `godot` requires periods at end of comments.
**Why it happens:** The golangci config enables ~60 linters. Developers unfamiliar with the config write code that passes `go vet` but fails the extended checks.
**How to avoid:** Follow these rules for all new code: (1) No package-level variables except constants. (2) Keep functions under 100 lines (ignore comments). (3) Define string constants for repeated values. (4) End all comments with a period. (5) Handle all switch cases explicitly or use default. (6) No naked returns. (7) Use `//nolint:lintername // reason` sparingly with a reason when truly needed.
**Warning signs:** CI/lint failures on otherwise correct code.

### Pitfall 4: Search Query SQL Injection via ILIKE
**What goes wrong:** Constructing ILIKE patterns by concatenating user input directly into the SQL query.
**Why it happens:** ILIKE requires pattern characters (`%`) as part of the value, tempting developers to use string formatting.
**How to avoid:** Pass the search term as a parameterized value and construct the pattern in Go: `searchPattern := strings.ToLower(term) + "%"`, then use `WHERE LOWER(last_name) LIKE $1` with `searchPattern` as the parameter. Never put user input in the SQL string itself.
**Warning signs:** Any SQL string containing `%` + a variable concatenation.

### Pitfall 5: Missing version Increment on Patient Updates
**What goes wrong:** The version column stays at 1 forever because the UPDATE query does not include `version = version + 1`.
**Why it happens:** Forgetting to include the version column in the UPDATE SQL.
**How to avoid:** The repository's `Update` method must include `version = version + 1` in the SET clause and `AND version = $N` (the expected version) in the WHERE clause. If no rows are updated, return a conflict error (409) indicating concurrent modification.
**Warning signs:** Version column always reads 1 in the database.

### Pitfall 6: N+1 Query on Patient List with session_count
**What goes wrong:** For each patient in the list, a separate query fetches the session count, causing O(N) database round trips.
**Why it happens:** Implementing the list as a loop with per-record queries.
**How to avoid:** Use a single query with a LEFT JOIN or subquery: `SELECT p.*, COALESCE(COUNT(s.id), 0) as session_count, MAX(s.created_at) as last_session_date FROM patients p LEFT JOIN sessions s ON s.patient_id = p.id GROUP BY p.id`. Since sessions do not exist yet, use `0` and `NULL` as constants in Phase 1, and replace with the real join when sessions are built in Phase 2.
**Warning signs:** Slow patient list response times proportional to patient count.

## Code Examples

### Database Schema: Users Table Migration (add role column)
```sql
-- Source: Designed for this project based on CONTEXT.md decisions
-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT NULL;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'doctor') OR role IS NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users DROP COLUMN IF EXISTS role;
-- +goose StatementEnd
```

### Database Schema: Patients Table
```sql
-- Source: Designed for this project based on CONTEXT.md decisions
-- +goose Up
-- +goose StatementBegin
CREATE TABLE patients (
    id          BIGSERIAL PRIMARY KEY,
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    sex         VARCHAR(10) NOT NULL CHECK (sex IN ('male', 'female', 'other')),
    phone       VARCHAR(50),
    email       VARCHAR(255),
    external_reference TEXT,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by  BIGINT NOT NULL REFERENCES users(id),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by  BIGINT NOT NULL REFERENCES users(id)
);

-- Indexes for search performance
CREATE INDEX idx_patients_last_name_lower ON patients (LOWER(last_name) varchar_pattern_ops);
CREATE INDEX idx_patients_first_name_lower ON patients (LOWER(first_name) varchar_pattern_ops);
CREATE INDEX idx_patients_email_lower ON patients (LOWER(email));
CREATE INDEX idx_patients_phone ON patients (phone);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS patients;
-- +goose StatementEnd
```

### Database Schema: Devices Table
```sql
-- Source: Designed for this project based on CONTEXT.md decisions
-- +goose Up
-- +goose StatementBegin
CREATE TABLE devices (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(200) NOT NULL,
    manufacturer VARCHAR(200) NOT NULL,
    model        VARCHAR(200) NOT NULL,
    device_type  VARCHAR(50) NOT NULL CHECK (device_type IN ('ipl', 'ndyag', 'co2', 'rf')),
    active       BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE handpieces (
    id         BIGSERIAL PRIMARY KEY,
    device_id  BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name       VARCHAR(200) NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_handpieces_device_id ON handpieces(device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS handpieces;
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd
```

### Database Schema: Products Table
```sql
-- Source: Designed for this project based on CONTEXT.md decisions
-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(200) NOT NULL,
    manufacturer    VARCHAR(200) NOT NULL,
    product_type    VARCHAR(50) NOT NULL CHECK (product_type IN ('filler', 'botulinum_toxin')),
    concentration   VARCHAR(100),
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
```

### Database Schema: Indication Codes and Clinical Endpoints
```sql
-- Source: Designed for this project based on CONTEXT.md decisions
-- +goose Up
-- +goose StatementBegin
CREATE TABLE indication_codes (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    module_type VARCHAR(50) NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE clinical_endpoints (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    module_type VARCHAR(50) NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_indication_codes_module ON indication_codes(module_type);
CREATE INDEX idx_clinical_endpoints_module ON clinical_endpoints(module_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinical_endpoints;
DROP TABLE IF EXISTS indication_codes;
-- +goose StatementEnd
```

### Extended JWT Claims with Role
```go
// Source: Extending existing internal/api/auth/auth.go
// Claims defines the JWT claims for access tokens.
type Claims struct {
    UserID int64  `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role,omitempty"`
    jwt.RegisteredClaims
}
```

### Patient Search Query Pattern
```go
// Source: Application pattern for ILIKE prefix search with pagination
// In the repository implementation:
func (r *PostgresPatientRepository) List(
    ctx context.Context,
    filter service.PatientFilter,
) (*service.PatientListResult, error) {
    baseQuery := `SELECT id, first_name, last_name, date_of_birth, sex,
                         phone, email, external_reference, version,
                         created_at, created_by, updated_at, updated_by,
                         0 AS session_count, NULL::timestamptz AS last_session_date
                  FROM patients`
    countQuery := `SELECT COUNT(*) FROM patients`
    whereClause := ""
    args := []interface{}{}
    argIndex := 1

    if filter.Search != "" {
        searchPattern := strings.ToLower(filter.Search) + "%"
        whereClause = fmt.Sprintf(` WHERE LOWER(last_name) LIKE $%d
            OR LOWER(first_name) LIKE $%d
            OR LOWER(email) LIKE $%d
            OR phone LIKE $%d`, argIndex, argIndex, argIndex, argIndex)
        args = append(args, searchPattern)
        argIndex++
    }

    // Count total matching records
    var total int
    err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&total)
    // ...

    // Apply ordering and pagination
    orderClause := " ORDER BY last_name ASC"
    offset := (filter.Page - 1) * filter.PerPage
    limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
    args = append(args, filter.PerPage, offset)

    rows, err := r.db.QueryContext(ctx, baseQuery+whereClause+orderClause+limitClause, args...)
    // ... scan rows into domain.Patient slice
}
```

### Route Registration Pattern
```go
// Source: Extending existing internal/api/routes/manager.go pattern

// In manager.go:
type Manager struct {
    authRoutes     *AuthRoutes
    userRoutes     *UserRoutes
    apiRoutes      *APIRoutes
    patientRoutes  *PatientRoutes
    registryRoutes *RegistryRoutes
    roleRoutes     *RoleRoutes
    metrics        *metrics.Client
}

// In patients.go route file:
func (pr *PatientRoutes) RegisterRoutes(router chi.Router) {
    router.Route("/patients", func(r chi.Router) {
        r.Use(middleware.RequireAuth(pr.config))
        r.Use(middleware.RequireRole("doctor", "admin"))

        r.Get("/", handlers.HandleListPatients(pr.patientService, pr.metrics))
        r.Post("/", handlers.HandleCreatePatient(pr.patientService, pr.metrics))
        r.Get("/{id}", handlers.HandleGetPatient(pr.patientService, pr.metrics))
        r.Put("/{id}", handlers.HandleUpdatePatient(pr.patientService, pr.metrics))
        r.Get("/{id}/sessions", handlers.HandleGetPatientSessions(pr.patientService, pr.metrics))
    })
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Handlers query DB directly | Service/repository layers | Introduced this phase | Testability, separation of concerns |
| No role awareness in JWT | Role field in claims + middleware | Introduced this phase | RBAC enforcement |
| Users table only has basic fields | Role column + metadata columns | Introduced this phase | Authorization capability |
| No clinical data tables | Patients + registry tables | Introduced this phase | Foundation for all clinical phases |

**Deprecated/outdated:**
- pgx/v4: Reached EOL July 2025 per STATE.md blocker. However, migration to v5 is explicitly deferred -- not interleaved with feature work. Continue using v4/stdlib for this phase.
- The placeholder handler functions in `users.go` (HandleListUsers, etc.) return hardcoded data. These should be replaced or deprecated as the service/repository pattern supersedes them.

## Open Questions

1. **exhaustruct linter and new domain structs**
   - What we know: The golangci config has `exhaustruct` in the commented-out "you may want to enable" section -- it is NOT currently enabled. The `exhaustive` linter (switch/map exhaustiveness) IS enabled.
   - What's unclear: Whether the team intends to enable `exhaustruct` soon.
   - Recommendation: Proceed without `exhaustruct`. But initialize all struct fields explicitly as a best practice in case it gets enabled later.

2. **Session count placeholder in patient list**
   - What we know: The patient list must include `session_count` and `last_session_date` per the locked decisions. Sessions do not exist until Phase 2.
   - What's unclear: Whether to add a LEFT JOIN placeholder now or hardcode zeros.
   - Recommendation: Return `0` for session_count and `null` for last_session_date as literal values in the SQL query for Phase 1. Replace with a real LEFT JOIN when sessions are introduced in Phase 2.

3. **Seed data accuracy for clinical devices**
   - What we know: STATE.md flags "Seed data content (real device models, product names) needs clinical input" as a concern.
   - What's unclear: Exact product concentrations, handpiece specifications.
   - Recommendation: Use publicly available data from manufacturer websites. Concentrations and specs can be refined later -- registry is read-only so corrections are just a new migration.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.0 |
| Config file | None (uses `go test` defaults) |
| Quick run command | `go test ./internal/... -count=1 -short` |
| Full suite command | `go test ./... -count=1 -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RBAC-01 | Doctor and Admin roles with distinct permissions | unit | `go test ./internal/api/middleware/ -run TestRequireRole -v` | No -- Wave 0 |
| RBAC-02 | Doctor can perform clinical operations | integration | `go test ./internal/api/handlers/ -run TestDoctorAccess -v` | No -- Wave 0 |
| RBAC-03 | Admin can manage patients but not clinical-only | integration | `go test ./internal/api/handlers/ -run TestAdminAccess -v` | No -- Wave 0 |
| RBAC-04 | Endpoints enforce role-based authorization | unit | `go test ./internal/api/middleware/ -run TestRequireRole -v` | No -- Wave 0 |
| PAT-01 | Create patient with demographics | unit | `go test ./internal/service/ -run TestCreatePatient -v` | No -- Wave 0 |
| PAT-02 | Search and list patients with filtering | unit | `go test ./internal/service/ -run TestListPatients -v` | No -- Wave 0 |
| PAT-03 | Update patient records | unit | `go test ./internal/service/ -run TestUpdatePatient -v` | No -- Wave 0 |
| PAT-04 | View patient session history | unit | `go test ./internal/service/ -run TestPatientSessionHistory -v` | No -- Wave 0 |
| REG-01 | Seed data for energy-based devices | smoke | `go test ./internal/repository/postgres/ -run TestDeviceSeedData -v` | No -- Wave 0 |
| REG-02 | Seed data for injectable products | smoke | `go test ./internal/repository/postgres/ -run TestProductSeedData -v` | No -- Wave 0 |
| REG-03 | Seed data for indication codes | smoke | `go test ./internal/repository/postgres/ -run TestIndicationCodeSeedData -v` | No -- Wave 0 |
| REG-04 | Select devices/products from controlled lists | unit | `go test ./internal/service/ -run TestListDevices -v` | No -- Wave 0 |
| META-01 | Records track created_at/by, updated_at/by | unit | `go test ./internal/service/ -run TestMetadataTracking -v` | No -- Wave 0 |
| META-02 | Signed records track signed_at/by | manual-only | Schema columns exist, populated in Phase 5 | N/A |
| META-03 | Version number increments | unit | `go test ./internal/service/ -run TestVersionIncrement -v` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -count=1 -short`
- **Per wave merge:** `go test ./... -count=1 -v && make lint`
- **Phase gate:** Full suite green + lint clean before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/service/patient_test.go` -- covers PAT-01, PAT-02, PAT-03, PAT-04, META-01, META-03
- [ ] `internal/service/registry_test.go` -- covers REG-04
- [ ] `internal/service/role_test.go` -- covers RBAC-01
- [ ] `internal/api/middleware/auth_test.go` -- covers RBAC-01, RBAC-04
- [ ] `internal/api/handlers/patients_test.go` -- covers RBAC-02, RBAC-03 (handler-level integration)
- [ ] Test infrastructure: mock implementations of repository interfaces for service-level unit tests
- [ ] Makefile `test` target: currently missing from Makefile, needs `go test ./... -count=1 -v`

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/api/` handlers, middleware, routes, auth -- direct code inspection
- Existing codebase: `migrations/` -- current schema and patterns
- Existing codebase: `golangci.yaml` -- exact linter configuration
- Existing codebase: `go.mod` -- exact dependency versions
- CONTEXT.md -- locked user decisions for Phase 1

### Secondary (MEDIUM confidence)
- [Three Dots Labs - Repository Pattern in Go](https://threedots.tech/post/repository-pattern-in-go/) -- service/repository architecture patterns
- [DEV.to - How to Optimise PostgreSQL ILIKE Queries](https://dev.to/tdournet/how-to-optimise-postgresql-like-and-ilike-queries-494i) -- ILIKE indexing with varchar_pattern_ops
- [PostgreSQL Wiki - Audit Trigger](https://wiki.postgresql.org/wiki/Audit_trigger) -- metadata column patterns
- [go-chi/chi GitHub](https://github.com/go-chi/chi) -- route group middleware patterns
- [PostgreSQL Wiki - Hibernate oplocks](https://wiki.postgresql.org/wiki/Hibernate_oplocks) -- version column trigger pattern

### Tertiary (LOW confidence)
- WebSearch results on first-user bootstrap patterns -- no authoritative source found, recommendation based on standard transactional patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries already in go.mod with known versions, no new dependencies
- Architecture: HIGH -- service/repository is well-documented Go pattern, codebase patterns directly inspected
- Pitfalls: HIGH -- linter config directly read, JWT backward compatibility is a known Go pattern
- Schema design: MEDIUM -- schema patterns are standard PostgreSQL, but exact index strategy for patient search may need tuning under load
- Seed data content: MEDIUM -- real device names from manufacturer websites, concentrations may need clinical review

**Research date:** 2026-03-07
**Valid until:** 2026-04-07 (stable domain -- Go 1.23, pgx v4, chi v5 are not changing)
