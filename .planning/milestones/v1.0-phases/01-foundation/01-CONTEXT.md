# Phase 1: Foundation - Context

**Gathered:** 2026-03-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Service/repository architecture, role-based access control (Doctor/Admin), patient management (CRUD with search), device/product registry (read-only seed data), and metadata tracking (created_at/by, updated_at/by, version). This phase establishes the layered architecture and foundational domains that all subsequent clinical phases build on.

</domain>

<decisions>
## Implementation Decisions

### Role Assignment
- Admin assigns roles -- users register without a role, an admin assigns Doctor or Admin afterward
- Single role per user (no multi-role)
- First registered user is auto-promoted to Admin (bootstrap mechanism)
- Unassigned users can authenticate (login, refresh, me) but are blocked from all clinical and admin endpoints with 403

### Patient Demographics
- Name fields: first_name, last_name (two fields)
- Contact: phone number and email address
- Biological sex: Male/Female/Other (labelled "sex" not "gender")
- Date of birth (required)
- Optional free-text external_reference field for clinic-specific identifiers (MRN, NHS number, chart number)

### Registry Seed Data
- Real-world device names and manufacturers (e.g., Lumenis M22, Candela GentleMax Pro)
- 2-3 devices per energy-based type (IPL, Nd:YAG, CO2, RF) and 2-3 products per injectable type (filler, botulinum toxin) -- approximately 15-20 entries total
- Curated indication codes and clinical endpoints grouped per module type (5-10 per type)
- Registry is strictly read-only in v1 -- no add/edit/delete endpoints (admin management deferred to v2 per ADMIN-01..03)

### Patient Search & List
- Searchable fields: name (partial/prefix match on first or last name) and phone/email
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

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- Handler closure pattern: `Handle{Action}(db, config, metrics) func(w, r)` -- all new handlers follow this
- Route Manager pattern: `{Domain}Routes` struct with `RegisterRoutes(chi.Router)` -- add PatientRoutes, RegistryRoutes, RoleRoutes
- `apierrors.WriteError()`: structured JSON errors with machine-readable codes -- extend with PATIENT_*, REGISTRY_*, ROLE_* codes
- Auth middleware `RequireAuth()`: JWT validation injecting Claims into context -- extend to check role claims
- `models.go`: shared response types -- add PatientResponse, DeviceResponse, ProductResponse, etc.
- Metrics Client: `Increment{Name}Count()` pattern -- add patient/registry operation counters

### Established Patterns
- Direct SQL with `database/sql` and pgx driver -- service/repository layers will wrap this
- Embedded Goose migrations with Up/Down sections -- new migrations for patients, devices, products, roles tables
- Config via Viper with `OVERRIDE_` env prefix -- no new config expected for Phase 1
- `log/slog` JSON logging -- carry forward for new domains
- Strict golangci-lint (~60 linters, max 100 lines/function, no globals, no init)

### Integration Points
- `internal/api/routes/manager.go`: Manager struct needs new route group fields and RegisterAllRoutes calls
- `internal/api/api.go`: App struct may need to pass role/user info to route constructors
- `internal/api/middleware/auth.go`: RequireAuth needs role-awareness (or new RequireRole middleware)
- `migrations/`: New SQL files for patients, devices, products, indication_codes, clinical_endpoints, user roles
- `internal/api/handlers/models.go`: New shared response types

</code_context>

<specifics>
## Specific Ideas

No specific requirements -- open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

None -- discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-03-07*
