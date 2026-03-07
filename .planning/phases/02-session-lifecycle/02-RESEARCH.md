# Phase 2: Session Lifecycle - Research

**Researched:** 2026-03-07
**Domain:** Treatment session management, state machines, consent/safety gates, polymorphic procedure module slots
**Confidence:** HIGH

## Summary

Phase 2 builds the treatment session domain on top of the service/repository architecture established in Phase 1. The core challenge is threefold: (1) designing a session entity with header fields, lifecycle state machine, and metadata tracking that follows the existing patient/registry patterns exactly, (2) implementing consent and contraindication screening as gate conditions that block downstream operations (adding procedure modules), and (3) creating the polymorphic `session_modules` slot table that future phases (3, 4) will populate with type-specific detail tables.

The codebase already has well-established patterns: handler closures receiving `*service.XxxService` and `*metrics.Client`, route structs with `RegisterRoutes(chi.Router)`, repository interfaces defined in the `service` package, PostgreSQL implementations in `internal/repository/postgres/`, domain models in `internal/domain/`, apierrors constants, and strict golangci-lint (~60 linters). Phase 2 must follow these patterns exactly while introducing the new session, consent, and contraindication domains. The roadmap decision (from STATE.md) to use "hybrid polymorphism -- shared session_modules table + per-type detail tables" is critical and must be reflected in the schema design.

Key integration points with Phase 1: the patient list endpoint currently returns hardcoded `session_count=0` and `last_session_date=null` -- Phase 2 must wire these to real session data via a LEFT JOIN or subquery. The `domain.SessionSummary` struct already exists as a placeholder. The session must reference `patients(id)` and `users(id)` as foreign keys.

**Primary recommendation:** Create three new domain entities (Session, Consent, ContraindicationScreening) with corresponding service/repository layers. Implement the session state machine as a map of valid transitions enforced in the service layer. Use a junction table for session-indication-code many-to-many relationships. Create a `session_modules` table with `module_type` and `sort_order` columns as the polymorphic base that Phases 3-4 will extend with detail tables.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SESS-01 | Clinician can create a new treatment session linked to a patient | Session domain model + service with Create method, migration for sessions table with FK to patients |
| SESS-02 | Session captures header fields: patient, clinician, timing, indication codes, patient goal, Fitzpatrick skin type, context flags | Sessions table with all header columns + junction table for indication codes (many-to-many) |
| SESS-03 | Session follows lifecycle states: Draft -> In Progress -> Awaiting Sign-off -> Signed -> Locked | String status column with CHECK constraint, state transition map in service layer |
| SESS-04 | Server enforces valid state transitions (no skipping states) | Service method `TransitionState` validates against allowed transitions map, returns error on invalid |
| SESS-05 | Clinician can save a session as draft and return to it later | Default status is "draft" on create; GET endpoint returns draft sessions; UPDATE endpoint allows editing draft fields |
| SESS-06 | Clinician can add multiple procedure modules to a single session | `session_modules` table with FK to sessions, `module_type` column, `sort_order` for ordering |
| CONS-01 | Clinician can record consent (type, method, datetime, risks discussed flag) | Consent table or embedded fields on session with type, method, datetime, risks_discussed_flag |
| CONS-02 | System blocks adding procedure modules until consent is captured | Service layer check: before inserting into session_modules, verify consent record exists for the session |
| CONS-03 | Clinician can complete contraindication screening checklist | Contraindication screening table with checklist items as structured JSON or normalized columns |
| CONS-04 | System captures contraindication flags and mitigation notes | Screening record includes boolean flags and free-text mitigation_notes |
| CONS-05 | Clinician can record photo consent status (yes/no/limited) | Photo consent column on session or consent record with CHECK constraint |
</phase_requirements>

## Standard Stack

### Core (already in go.mod -- no new dependencies)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-chi/chi/v5 | 5.0.10 | HTTP router with groups/middleware | Route registration for session endpoints |
| database/sql (stdlib) | Go 1.23 | DB interface | All repository implementations |
| jackc/pgx/v4 | 4.18.3 | PostgreSQL driver | Underlying driver via database/sql |
| pressly/goose/v3 | 3.26.0 | Database migrations | Session, consent, screening, modules tables |
| stretchr/testify | 1.11.0 | Test assertions | Service and handler tests |
| log/slog (stdlib) | Go 1.23 | Structured logging | All new code |
| encoding/json (stdlib) | Go 1.23 | JSON encoding | Request/response serialization, JSON column handling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| time (stdlib) | Go 1.23 | Timestamps | Session timing, consent datetime |
| context (stdlib) | Go 1.23 | Request-scoped values | Passing user claims, request IDs |
| errors (stdlib) | Go 1.23 | Error wrapping | Sentinel errors for session domain |
| fmt (stdlib) | Go 1.23 | Error formatting | Repository error wrapping |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Service-layer state machine | PostgreSQL CHECK + trigger | DB triggers for state validation would couple logic to schema and are harder to test. Service-layer validation is testable with mocks and keeps business rules in Go. Use CHECK constraint only for valid values, not transition rules. |
| Separate consent/screening tables | Embedded JSON columns on sessions | Separate tables allow proper indexing, foreign key constraints, and independent queries. JSON columns would be simpler but make the consent gate check harder to express as a clean FK/EXISTS query. |
| Normalized contraindication checklist items | JSON array column | Normalized rows per checklist item add complexity with no benefit for a small fixed checklist. Use a single screening record with boolean flag columns (e.g., `pregnant`, `anticoagulants`, `active_infection`) for the common items plus a free-text `other_notes` field. This is simpler and still queryable. |

**Installation:**
```bash
# No new dependencies required -- all libraries already in go.mod
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
  domain/
    session.go              # NEW: Session, SessionStatus constants
    consent.go              # NEW: Consent domain model
    contraindication.go     # NEW: ContraindicationScreening model
    session_module.go       # NEW: SessionModule (polymorphic base)
  service/
    session.go              # NEW: SessionService + SessionRepository interface
    consent.go              # NEW: ConsentService + ConsentRepository interface
    contraindication.go     # NEW: ContraindicationService + repository interface
  repository/
    postgres/
      session.go            # NEW: PostgreSQL session repository
      consent.go            # NEW: PostgreSQL consent repository
      contraindication.go   # NEW: PostgreSQL contraindication repository
  api/
    handlers/
      sessions.go           # NEW: Session CRUD + state transition handlers
      consent.go            # NEW: Consent recording handler
      contraindication.go   # NEW: Screening handler
      models.go             # EXTEND: Session/consent/screening response types
    routes/
      sessions.go           # NEW: Session route registration
      manager.go            # EXTEND: Add sessionRoutes to Manager
    apierrors/
      apierrors.go          # EXTEND: Session/consent/screening error codes
    metrics/
      metrics.go            # EXTEND: Session-related counters
  testutil/
    mock_session.go         # NEW: Mock session repository
    mock_consent.go         # NEW: Mock consent repository
    mock_contraindication.go # NEW: Mock contraindication repository
migrations/
  20260308010000_create_sessions_table.sql
  20260308010001_create_session_indication_codes.sql
  20260308010002_create_consent_table.sql
  20260308010003_create_contraindication_screening.sql
  20260308010004_create_session_modules.sql
```

### Pattern 1: Session State Machine in Service Layer

**What:** A map of valid state transitions defines which status changes are allowed. The service method validates the requested transition before executing the update.

**When to use:** All session status transitions.

**Example:**
```go
// internal/service/session.go

// Valid session state transitions.
// Key = current state, Value = set of allowed next states.
var validTransitions = map[string][]string{ //nolint:gochecknoglobals // transition map
    domain.SessionStatusDraft:            {domain.SessionStatusInProgress},
    domain.SessionStatusInProgress:       {domain.SessionStatusAwaitingSignoff},
    domain.SessionStatusAwaitingSignoff:  {domain.SessionStatusSigned, domain.SessionStatusInProgress},
    domain.SessionStatusSigned:           {domain.SessionStatusLocked},
    // Locked is terminal -- no transitions out.
}

// TransitionState validates and applies a session state change.
func (s *SessionService) TransitionState(ctx context.Context, sessionID int64, newStatus string, userID int64) error {
    session, err := s.repo.GetByID(ctx, sessionID)
    if err != nil {
        return err
    }

    allowed, ok := validTransitions[session.Status]
    if !ok {
        return ErrInvalidStateTransition
    }

    valid := false
    for _, a := range allowed {
        if a == newStatus {
            valid = true
            break
        }
    }
    if !valid {
        return ErrInvalidStateTransition
    }

    return s.repo.UpdateStatus(ctx, sessionID, newStatus, session.Version, userID)
}
```

### Pattern 2: Consent Gate Check

**What:** Before allowing procedure module insertion, the service checks that a consent record exists for the session. This is a business rule enforced at the service layer, not a database constraint.

**When to use:** Any operation that adds a session_module.

**Example:**
```go
// internal/service/session.go (or a dedicated module service)

// AddModule validates consent exists before inserting a procedure module slot.
func (s *SessionService) AddModule(ctx context.Context, sessionID int64, moduleType string, userID int64) (*domain.SessionModule, error) {
    // Verify session exists and is in a mutable state.
    session, err := s.repo.GetByID(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    if session.Status != domain.SessionStatusDraft && session.Status != domain.SessionStatusInProgress {
        return nil, ErrSessionNotEditable
    }

    // Check consent gate (CONS-02).
    hasConsent, err := s.consentRepo.ExistsForSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    if !hasConsent {
        return nil, ErrConsentRequired
    }

    // Insert module slot.
    module := &domain.SessionModule{
        SessionID:  sessionID,
        ModuleType: moduleType,
        SortOrder:  0, // Calculated by repository.
    }
    if err := s.moduleRepo.Create(ctx, module, userID); err != nil {
        return nil, err
    }
    return module, nil
}
```

### Pattern 3: Handler Closure Following Existing Convention

**What:** Handlers follow the exact same pattern as `HandleCreatePatient` -- closure receiving service + metrics, extracting claims, parsing request, calling service, returning response.

**When to use:** All new session/consent/screening handlers.

**Example:**
```go
// internal/api/handlers/sessions.go

// HandleCreateSession creates a new treatment session linked to a patient.
func HandleCreateSession(svc *service.SessionService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        claims := middleware.GetUserClaims(r.Context())
        if claims == nil {
            apierrors.WriteError(w, http.StatusUnauthorized,
                apierrors.AuthNotAuthenticated, "not authenticated")
            return
        }

        var req createSessionRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            apierrors.WriteError(w, http.StatusBadRequest,
                apierrors.ValidationInvalidRequestBody, "invalid request body")
            return
        }

        // Build domain model from request, call svc.Create(), return response.
    }
}
```

### Pattern 4: Route Registration Following Manager Pattern

**What:** New `SessionRoutes` struct registered in the route Manager, following the exact pattern of `PatientRoutes` and `RegistryRoutes`.

**When to use:** Session, consent, and screening endpoints.

**Example:**
```go
// internal/api/routes/sessions.go

// SessionRoutes handles all session-related routes.
type SessionRoutes struct {
    sessionSvc *service.SessionService
    consentSvc *service.ConsentService
    screenSvc  *service.ContraindicationService
    config     *config.Configuration
    metrics    *metrics.Client
}

// RegisterRoutes registers all session routes under the /sessions prefix.
func (sr *SessionRoutes) RegisterRoutes(router chi.Router) {
    router.Route("/sessions", func(r chi.Router) {
        r.Use(middleware.RequireAuth(sr.config))
        r.Use(middleware.RequireRole(domain.RoleDoctor))

        r.Post("/", handlers.HandleCreateSession(sr.sessionSvc, sr.metrics))
        r.Get("/", handlers.HandleListSessions(sr.sessionSvc, sr.metrics))
        r.Get("/{id}", handlers.HandleGetSession(sr.sessionSvc, sr.metrics))
        r.Put("/{id}", handlers.HandleUpdateSession(sr.sessionSvc, sr.metrics))
        r.Post("/{id}/transition", handlers.HandleTransitionSession(sr.sessionSvc, sr.metrics))

        // Consent sub-resource.
        r.Post("/{id}/consent", handlers.HandleRecordConsent(sr.consentSvc, sr.metrics))
        r.Get("/{id}/consent", handlers.HandleGetConsent(sr.consentSvc, sr.metrics))

        // Contraindication screening sub-resource.
        r.Post("/{id}/screening", handlers.HandleRecordScreening(sr.screenSvc, sr.metrics))
        r.Get("/{id}/screening", handlers.HandleGetScreening(sr.screenSvc, sr.metrics))

        // Module slots.
        r.Post("/{id}/modules", handlers.HandleAddModule(sr.sessionSvc, sr.metrics))
        r.Get("/{id}/modules", handlers.HandleListModules(sr.sessionSvc, sr.metrics))
        r.Delete("/{id}/modules/{moduleId}", handlers.HandleRemoveModule(sr.sessionSvc, sr.metrics))
    })
}
```

### Anti-Patterns to Avoid

- **State transitions via direct field update:** Never allow the caller to set `status` directly in an update request body. State transitions must go through a dedicated endpoint that validates the transition.
- **Consent check at the database level via FK constraint:** A foreign key from session_modules to consent would create a rigid coupling. The consent gate is a business rule -- enforce it in the service layer where it can return a clear error message.
- **Single monolithic session handler file:** Split consent and screening into separate handler files following the project convention of "one file per domain."
- **Embedding all screening data as JSON:** While tempting, JSON blobs defeat the purpose of structured clinical data. Use proper columns for known checklist items.
- **Putting session state transition logic in the handler:** Handlers must remain thin. The state machine belongs in the service layer where it can be unit tested with mocks.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| State machine validation | Custom graph traversal | Simple map[string][]string lookup | Session lifecycle is a small fixed graph (5 states). A map of valid transitions is sufficient and testable. |
| UUID for module slots | Custom ID generation | BIGSERIAL (database auto-increment) | Consistent with all existing tables (patients, devices, etc.) which use BIGSERIAL. |
| Date/time parsing | Custom date format parser | `time.Parse("2006-01-02", ...)` and `time.Parse(time.RFC3339, ...)` | Already established in patient handler for date_of_birth. |
| JSON error responses | Custom error formatting | `apierrors.WriteError()` | Already established, provides consistent machine-readable error codes. |
| Pagination for session lists | New pagination implementation | Reuse `PaginatedResponse` struct from models.go | Already exists and tested. |
| Optimistic locking | Custom locking mechanism | `version` column with `WHERE version = $N` in UPDATE | Already established in patient repository. |

**Key insight:** Phase 2 introduces significant new domain complexity but zero new libraries or patterns. Every technical mechanism needed (CRUD, state validation, gate checks, FK relationships, optimistic locking, pagination) is already established in Phase 1. The challenge is domain modeling, not infrastructure.

## Common Pitfalls

### Pitfall 1: State Transition Race Condition
**What goes wrong:** Two concurrent requests try to transition the same session. Both read the current state as "draft", both validate the transition to "in_progress" as valid, both write. The session ends up in "in_progress" but skipped the concurrent validation.
**Why it happens:** Check-then-update without atomicity.
**How to avoid:** Use optimistic locking: include `AND version = $expected_version` in the UPDATE WHERE clause. The `UpdateStatus` repository method must return a conflict error if zero rows are affected, same pattern as `PatientRepository.Update`.
**Warning signs:** Version column not being checked in state transition queries.

### Pitfall 2: Consent Gate Bypass via Direct Module Insert
**What goes wrong:** A handler for adding modules bypasses the consent check, or a new code path is added later that does not include the check.
**Why it happens:** The consent gate is a business rule, not a database constraint. If enforcement is scattered across multiple code paths, one will be missed.
**How to avoid:** Centralize the consent check in a single service method (`SessionService.AddModule` or `ModuleService.Create`). All handlers that add modules must go through this method. Never expose the module repository directly to handlers.
**Warning signs:** Any handler that calls `moduleRepo.Create` directly instead of through the service.

### Pitfall 3: Indication Codes Many-to-Many Without Junction Table
**What goes wrong:** Storing indication codes as a JSON array or comma-separated string on the sessions table. This prevents indexing, referential integrity, and makes querying sessions by indication code impossible.
**Why it happens:** JSON arrays feel simpler for a "list of IDs."
**How to avoid:** Create a proper junction table `session_indication_codes` with FKs to both `sessions(id)` and `indication_codes(id)`. This allows the database to enforce referential integrity and enables queries like "find all sessions for indication code X."
**Warning signs:** JSON array columns for reference data.

### Pitfall 4: Missing Foreign Key Validation on Session Create
**What goes wrong:** Creating a session with a non-existent patient_id returns a cryptic database error instead of a clean 404.
**Why it happens:** The PostgreSQL FK constraint catches the violation, but the error message is not user-friendly.
**How to avoid:** In the service layer, verify the patient exists before creating the session. If not found, return a clear `ErrPatientNotFound` sentinel error that the handler maps to a 404. The FK constraint serves as a safety net but should not be the primary validation path.
**Warning signs:** Raw PostgreSQL FK violation errors in API responses.

### Pitfall 5: Editing Sessions in Non-Editable States
**What goes wrong:** A clinician can modify header fields on a session that is in "awaiting_signoff" or "signed" state, corrupting the record before locking.
**Why it happens:** The update handler does not check the session's current state before allowing field changes.
**How to avoid:** In the session service `Update` method, check that the session is in an editable state (draft or in_progress). Return `ErrSessionNotEditable` for sessions in any other state. Phase 5 will add database-level immutability triggers, but the application layer must enforce this now.
**Warning signs:** Being able to PUT changes to a session in "awaiting_signoff" state.

### Pitfall 6: golangci-lint Failures on New Domain Code
**What goes wrong:** The strict linter config (gomnd, godot, gochecknoglobals, funlen, goconst) rejects new code that looks correct.
**Why it happens:** Same as Phase 1 -- the config enables ~60 linters.
**How to avoid:** (1) Use named constants for magic numbers (e.g., Fitzpatrick types 1-6). (2) End all comments with a period. (3) Mark sentinel errors with `//nolint:gochecknoglobals // sentinel error`. (4) Keep functions under 100 lines. (5) Extract repeated strings into constants. (6) Use the existing `//nolint:errcheck // response write` pattern for `json.NewEncoder(w).Encode()` calls.
**Warning signs:** Lint failures on new files.

### Pitfall 7: Not Updating Patient Session Count/Date
**What goes wrong:** After Phase 2, the patient list still shows `session_count=0` and `last_session_date=null` because the patient repository was not updated to query real session data.
**Why it happens:** The Phase 1 code uses hardcoded placeholders. Easy to forget to wire up the real data.
**How to avoid:** As part of Phase 2, update `PostgresPatientRepository.List()` and the `PatientListItem` population to use a LEFT JOIN on the sessions table. Also update `GetSessionHistory` to return real session summaries instead of an empty slice.
**Warning signs:** Patient list response always showing zero sessions after sessions exist.

## Code Examples

### Database Schema: Sessions Table
```sql
-- Source: Designed for this project following existing migration patterns
-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    id                  BIGSERIAL PRIMARY KEY,
    patient_id          BIGINT NOT NULL REFERENCES patients(id),
    clinician_id        BIGINT NOT NULL REFERENCES users(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft', 'in_progress', 'awaiting_signoff', 'signed', 'locked')),
    scheduled_at        TIMESTAMPTZ,
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    patient_goal        TEXT,
    fitzpatrick_type    SMALLINT CHECK (fitzpatrick_type BETWEEN 1 AND 6),
    is_tanned           BOOLEAN NOT NULL DEFAULT false,
    is_pregnant         BOOLEAN NOT NULL DEFAULT false,
    on_anticoagulants   BOOLEAN NOT NULL DEFAULT false,
    photo_consent       VARCHAR(10) CHECK (photo_consent IN ('yes', 'no', 'limited')),
    notes               TEXT,
    version             INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by          BIGINT NOT NULL REFERENCES users(id),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by          BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_sessions_patient_id ON sessions(patient_id);
CREATE INDEX idx_sessions_clinician_id ON sessions(clinician_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
```

### Database Schema: Session-Indication-Codes Junction Table
```sql
-- Source: Many-to-many relationship for session header fields
-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_indication_codes (
    session_id       BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    indication_code_id BIGINT NOT NULL REFERENCES indication_codes(id),
    PRIMARY KEY (session_id, indication_code_id)
);

CREATE INDEX idx_session_indications_session ON session_indication_codes(session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_indication_codes;
-- +goose StatementEnd
```

### Database Schema: Consent Table
```sql
-- Source: Consent recording per CONS-01
-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_consents (
    id              BIGSERIAL PRIMARY KEY,
    session_id      BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    consent_type    VARCHAR(50) NOT NULL,
    consent_method  VARCHAR(50) NOT NULL,
    obtained_at     TIMESTAMPTZ NOT NULL,
    risks_discussed BOOLEAN NOT NULL DEFAULT false,
    notes           TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      BIGINT NOT NULL REFERENCES users(id),
    UNIQUE (session_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_consents;
-- +goose StatementEnd
```

### Database Schema: Contraindication Screening Table
```sql
-- Source: Contraindication checklist per CONS-03, CONS-04
-- +goose Up
-- +goose StatementBegin
CREATE TABLE contraindication_screenings (
    id                  BIGSERIAL PRIMARY KEY,
    session_id          BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    pregnant            BOOLEAN NOT NULL DEFAULT false,
    breastfeeding       BOOLEAN NOT NULL DEFAULT false,
    active_infection    BOOLEAN NOT NULL DEFAULT false,
    active_cold_sores   BOOLEAN NOT NULL DEFAULT false,
    isotretinoin        BOOLEAN NOT NULL DEFAULT false,
    photosensitivity    BOOLEAN NOT NULL DEFAULT false,
    autoimmune_disorder BOOLEAN NOT NULL DEFAULT false,
    keloid_history      BOOLEAN NOT NULL DEFAULT false,
    anticoagulants      BOOLEAN NOT NULL DEFAULT false,
    recent_tan          BOOLEAN NOT NULL DEFAULT false,
    has_flags           BOOLEAN NOT NULL DEFAULT false,
    mitigation_notes    TEXT,
    notes               TEXT,
    version             INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by          BIGINT NOT NULL REFERENCES users(id),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by          BIGINT NOT NULL REFERENCES users(id),
    UNIQUE (session_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS contraindication_screenings;
-- +goose StatementEnd
```

### Database Schema: Session Modules (Polymorphic Base)
```sql
-- Source: Hybrid polymorphism pattern per roadmap decision
-- +goose Up
-- +goose StatementBegin
CREATE TABLE session_modules (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    module_type VARCHAR(30) NOT NULL
                CHECK (module_type IN ('ipl', 'ndyag', 'co2', 'rf', 'filler', 'botulinum_toxin')),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by  BIGINT NOT NULL REFERENCES users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by  BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_session_modules_session ON session_modules(session_id);
CREATE INDEX idx_session_modules_type ON session_modules(module_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session_modules;
-- +goose StatementEnd
```

### Session Domain Model
```go
// internal/domain/session.go

package domain

import "time"

// Session status constants.
const (
    SessionStatusDraft           = "draft"
    SessionStatusInProgress      = "in_progress"
    SessionStatusAwaitingSignoff = "awaiting_signoff"
    SessionStatusSigned          = "signed"
    SessionStatusLocked          = "locked"
)

// FitzpatrickMin is the minimum valid Fitzpatrick skin type.
const FitzpatrickMin = 1

// FitzpatrickMax is the maximum valid Fitzpatrick skin type.
const FitzpatrickMax = 6

// Session represents a treatment session record.
type Session struct {
    ID               int64      `json:"id"`
    PatientID        int64      `json:"patient_id"`
    ClinicianID      int64      `json:"clinician_id"`
    Status           string     `json:"status"`
    ScheduledAt      *time.Time `json:"scheduled_at"`
    StartedAt        *time.Time `json:"started_at"`
    CompletedAt      *time.Time `json:"completed_at"`
    PatientGoal      *string    `json:"patient_goal"`
    FitzpatrickType  *int       `json:"fitzpatrick_type"`
    IsTanned         bool       `json:"is_tanned"`
    IsPregnant       bool       `json:"is_pregnant"`
    OnAnticoagulants bool       `json:"on_anticoagulants"`
    PhotoConsent     *string    `json:"photo_consent"`
    Notes            *string    `json:"notes"`
    IndicationCodes  []int64    `json:"indication_code_ids,omitempty"`
    Version          int        `json:"version"`
    CreatedAt        time.Time  `json:"created_at"`
    CreatedBy        int64      `json:"created_by"`
    UpdatedAt        time.Time  `json:"updated_at"`
    UpdatedBy        int64      `json:"updated_by"`
}
```

### Consent Domain Model
```go
// internal/domain/consent.go

package domain

import "time"

// Consent represents a treatment consent record for a session.
type Consent struct {
    ID             int64     `json:"id"`
    SessionID      int64     `json:"session_id"`
    ConsentType    string    `json:"consent_type"`
    ConsentMethod  string    `json:"consent_method"`
    ObtainedAt     time.Time `json:"obtained_at"`
    RisksDiscussed bool      `json:"risks_discussed"`
    Notes          *string   `json:"notes"`
    Version        int       `json:"version"`
    CreatedAt      time.Time `json:"created_at"`
    CreatedBy      int64     `json:"created_by"`
    UpdatedAt      time.Time `json:"updated_at"`
    UpdatedBy      int64     `json:"updated_by"`
}
```

### Session Service with Repository Interface
```go
// internal/service/session.go

package service

import (
    "context"
    "errors"
    "dermify-api/internal/domain"
)

var (
    ErrSessionNotFound          = errors.New("session not found")           //nolint:gochecknoglobals // sentinel error
    ErrSessionVersionConflict   = errors.New("session version conflict")   //nolint:gochecknoglobals // sentinel error
    ErrInvalidSessionData       = errors.New("invalid session data")       //nolint:gochecknoglobals // sentinel error
    ErrInvalidStateTransition   = errors.New("invalid state transition")   //nolint:gochecknoglobals // sentinel error
    ErrSessionNotEditable       = errors.New("session is not editable")    //nolint:gochecknoglobals // sentinel error
    ErrConsentRequired          = errors.New("consent required before adding modules") //nolint:gochecknoglobals // sentinel error
)

// SessionRepository defines the data access contract for sessions.
type SessionRepository interface {
    Create(ctx context.Context, session *domain.Session) error
    GetByID(ctx context.Context, id int64) (*domain.Session, error)
    Update(ctx context.Context, session *domain.Session) error
    UpdateStatus(ctx context.Context, id int64, status string, expectedVersion int, userID int64) error
    List(ctx context.Context, filter SessionFilter) (*SessionListResult, error)
    ListByPatient(ctx context.Context, patientID int64) ([]domain.SessionSummary, error)
    SetIndicationCodes(ctx context.Context, sessionID int64, codeIDs []int64) error
}
```

### Error Codes for New Domains
```go
// Extend internal/api/apierrors/apierrors.go

// Error codes for session-related failures.
const (
    SessionNotFound           = "SESSION_NOT_FOUND"
    SessionVersionConflict    = "SESSION_VERSION_CONFLICT"
    SessionInvalidData        = "SESSION_INVALID_DATA"
    SessionInvalidTransition  = "SESSION_INVALID_TRANSITION"
    SessionNotEditable        = "SESSION_NOT_EDITABLE"
    SessionCreationFailed     = "SESSION_CREATION_FAILED"
    SessionUpdateFailed       = "SESSION_UPDATE_FAILED"
    SessionLookupFailed       = "SESSION_LOOKUP_FAILED"
)

// Error codes for consent-related failures.
const (
    ConsentNotFound       = "CONSENT_NOT_FOUND"
    ConsentRequired       = "CONSENT_REQUIRED"
    ConsentAlreadyExists  = "CONSENT_ALREADY_EXISTS"
    ConsentInvalidData    = "CONSENT_INVALID_DATA"
    ConsentCreationFailed = "CONSENT_CREATION_FAILED"
)

// Error codes for screening-related failures.
const (
    ScreeningNotFound       = "SCREENING_NOT_FOUND"
    ScreeningAlreadyExists  = "SCREENING_ALREADY_EXISTS"
    ScreeningInvalidData    = "SCREENING_INVALID_DATA"
    ScreeningCreationFailed = "SCREENING_CREATION_FAILED"
)

// Error codes for module-related failures.
const (
    ModuleNotFound       = "MODULE_NOT_FOUND"
    ModuleInvalidData    = "MODULE_INVALID_DATA"
    ModuleCreationFailed = "MODULE_CREATION_FAILED"
    ModuleRemovalFailed  = "MODULE_REMOVAL_FAILED"
)
```

### Updating Patient Repository for Real Session Data
```go
// internal/repository/postgres/patient.go -- update List query

// Updated List query with real session data join:
baseQuery := `SELECT p.id, p.first_name, p.last_name, p.date_of_birth, p.sex,
    p.phone, p.email, p.external_reference, p.version,
    p.created_at, p.created_by, p.updated_at, p.updated_by,
    COALESCE(sess.session_count, 0) AS session_count,
    sess.last_session_date
FROM patients p
LEFT JOIN (
    SELECT patient_id,
           COUNT(*) AS session_count,
           MAX(created_at) AS last_session_date
    FROM sessions
    GROUP BY patient_id
) sess ON sess.patient_id = p.id`

// Updated GetSessionHistory to return real sessions:
func (r *PostgresPatientRepository) GetSessionHistory(ctx context.Context, patientID int64) ([]domain.SessionSummary, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, created_at, status FROM sessions
         WHERE patient_id = $1
         ORDER BY created_at DESC`, patientID,
    )
    // ... scan and return
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hardcoded session_count=0 in patient list | LEFT JOIN on sessions table | Phase 2 | Patient list shows real session counts |
| Empty GetSessionHistory placeholder | Real query on sessions table | Phase 2 | Patient detail shows real session history |
| SessionSummary as a placeholder struct | Fully populated from sessions table | Phase 2 | PAT-04 requirement fulfilled with real data |
| No clinical workflow entities | Sessions, consent, screening, modules | Phase 2 | Foundation for procedure documentation |

**Deprecated/outdated:**
- `domain.SessionSummary`: The existing placeholder struct from Phase 1 should remain but be populated with real data from the sessions table.
- `PostgresPatientRepository.GetSessionHistory`: Currently returns empty slice -- must be updated to query real sessions.
- Patient list `session_count=0` and `last_session_date=nil` hardcoding in `scanPatientListItems()` -- must be replaced with real LEFT JOIN data.

## API Endpoint Summary

| Method | Path | Handler | Purpose | Auth |
|--------|------|---------|---------|------|
| POST | /api/v1/sessions | HandleCreateSession | Create new session | Doctor only |
| GET | /api/v1/sessions | HandleListSessions | List sessions with filters | Doctor |
| GET | /api/v1/sessions/{id} | HandleGetSession | Get session by ID | Doctor |
| PUT | /api/v1/sessions/{id} | HandleUpdateSession | Update session header fields | Doctor |
| POST | /api/v1/sessions/{id}/transition | HandleTransitionSession | Change session state | Doctor |
| POST | /api/v1/sessions/{id}/consent | HandleRecordConsent | Record consent | Doctor |
| GET | /api/v1/sessions/{id}/consent | HandleGetConsent | Get consent for session | Doctor |
| PUT | /api/v1/sessions/{id}/consent | HandleUpdateConsent | Update consent record | Doctor |
| POST | /api/v1/sessions/{id}/screening | HandleRecordScreening | Record screening | Doctor |
| GET | /api/v1/sessions/{id}/screening | HandleGetScreening | Get screening for session | Doctor |
| PUT | /api/v1/sessions/{id}/screening | HandleUpdateScreening | Update screening record | Doctor |
| POST | /api/v1/sessions/{id}/modules | HandleAddModule | Add procedure module slot | Doctor |
| GET | /api/v1/sessions/{id}/modules | HandleListModules | List modules for session | Doctor |
| DELETE | /api/v1/sessions/{id}/modules/{moduleId} | HandleRemoveModule | Remove module slot | Doctor |

## Open Questions

1. **Consent type and method enum values**
   - What we know: CONS-01 requires "type, method, datetime, risks discussed flag." The exact enum values for consent_type (e.g., "informed", "verbal", "written") and consent_method (e.g., "in_person", "telehealth", "paper_form") are not specified.
   - What's unclear: The exact clinical vocabulary for consent types/methods in aesthetic dermatology.
   - Recommendation: Use VARCHAR with no CHECK constraint for now. Common values: consent_type ("informed", "procedure_specific"), consent_method ("verbal", "written", "electronic"). Let the frontend provide valid options. This can be tightened with a CHECK constraint later.

2. **Contraindication checklist items**
   - What we know: CONS-03 says "contraindication screening checklist." CONS-04 says "flags and mitigation notes."
   - What's unclear: The exact checklist items for aesthetic dermatology.
   - Recommendation: Use common dermatology/aesthetic contraindications as boolean columns: pregnant, breastfeeding, active_infection, active_cold_sores, isotretinoin (Accutane within 6 months), photosensitivity, autoimmune_disorder, keloid_history, anticoagulants, recent_tan. Plus a `has_flags` computed/set boolean and free-text `mitigation_notes`. This covers the most common screening items. Additional items can be added via migration later.

3. **Session list endpoint scope**
   - What we know: Clinicians need to find and return to draft sessions.
   - What's unclear: Whether session list should be filtered by clinician (my sessions only) or show all sessions.
   - Recommendation: Support both via query parameter: `?clinician_id=` for "my sessions", no filter for "all sessions." Default to all sessions. Also support `?patient_id=` and `?status=` filters.

4. **Photo consent placement**
   - What we know: CONS-05 requires "photo consent status (yes/no/limited)."
   - What's unclear: Whether this belongs on the session table or the consent table.
   - Recommendation: Put `photo_consent` as a column on the `sessions` table with `CHECK (photo_consent IN ('yes', 'no', 'limited'))`. It is a session-level flag, not a consent-type-specific field. Phase 6 (photo documentation) will check this field before allowing uploads.

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
| SESS-01 | Create session linked to patient | unit | `go test ./internal/service/ -run TestCreateSession -v` | No -- Wave 0 |
| SESS-02 | Session captures header fields | unit | `go test ./internal/service/ -run TestSessionHeaderFields -v` | No -- Wave 0 |
| SESS-03 | Session lifecycle states | unit | `go test ./internal/service/ -run TestSessionStateTransition -v` | No -- Wave 0 |
| SESS-04 | Reject invalid state transitions | unit | `go test ./internal/service/ -run TestInvalidStateTransition -v` | No -- Wave 0 |
| SESS-05 | Save session as draft | unit | `go test ./internal/service/ -run TestDraftSession -v` | No -- Wave 0 |
| SESS-06 | Add multiple procedure modules | unit | `go test ./internal/service/ -run TestAddModule -v` | No -- Wave 0 |
| CONS-01 | Record consent | unit | `go test ./internal/service/ -run TestRecordConsent -v` | No -- Wave 0 |
| CONS-02 | Block modules without consent | unit | `go test ./internal/service/ -run TestConsentGate -v` | No -- Wave 0 |
| CONS-03 | Contraindication screening | unit | `go test ./internal/service/ -run TestRecordScreening -v` | No -- Wave 0 |
| CONS-04 | Screening flags and notes | unit | `go test ./internal/service/ -run TestScreeningFlags -v` | No -- Wave 0 |
| CONS-05 | Photo consent status | unit | `go test ./internal/service/ -run TestPhotoConsent -v` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -count=1 -short`
- **Per wave merge:** `go test ./... -count=1 -v && make lint`
- **Phase gate:** Full suite green + lint clean before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/service/session_test.go` -- covers SESS-01 through SESS-06, CONS-02
- [ ] `internal/service/consent_test.go` -- covers CONS-01
- [ ] `internal/service/contraindication_test.go` -- covers CONS-03, CONS-04, CONS-05
- [ ] `internal/testutil/mock_session.go` -- mock session repository
- [ ] `internal/testutil/mock_consent.go` -- mock consent repository
- [ ] `internal/testutil/mock_contraindication.go` -- mock contraindication repository
- [ ] `internal/api/handlers/sessions_test.go` -- handler-level tests for RBAC (Doctor-only access)

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/domain/` -- domain model patterns (Patient, Device, Product, SessionSummary)
- Existing codebase: `internal/service/` -- service/repository interface pattern (PatientService, RegistryService)
- Existing codebase: `internal/repository/postgres/` -- PostgreSQL repository implementations
- Existing codebase: `internal/api/handlers/` -- handler closure pattern with service injection
- Existing codebase: `internal/api/routes/manager.go` -- route Manager registration pattern
- Existing codebase: `internal/api/apierrors/apierrors.go` -- error code constants and WriteError pattern
- Existing codebase: `migrations/` -- Goose SQL migration file structure
- Existing codebase: `golangci.yaml` -- exact linter configuration
- STATE.md: "Hybrid polymorphism for modules -- shared session_modules table + per-type detail tables"
- REQUIREMENTS.md: SESS-01 through SESS-06, CONS-01 through CONS-05

### Secondary (MEDIUM confidence)
- Phase 1 RESEARCH.md: Established patterns for service/repository, pagination, optimistic locking
- PostgreSQL documentation: CHECK constraints, TIMESTAMPTZ, BIGSERIAL, junction tables
- Clinical dermatology practice: Common contraindication screening checklist items (pregnancy, photosensitivity, isotretinoin, etc.)

### Tertiary (LOW confidence)
- Consent type/method vocabulary -- no authoritative source for aesthetic dermatology consent classification. Using common clinical terms.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all patterns established in Phase 1
- Architecture: HIGH -- direct code inspection of existing patterns, roadmap decisions documented
- Schema design: HIGH -- follows exact same patterns as patients/devices/products tables
- State machine: HIGH -- small fixed graph, well-understood pattern, testable with mocks
- Consent/screening fields: MEDIUM -- specific checklist items and consent type vocabulary based on general clinical knowledge, may need clinical review
- Pitfalls: HIGH -- identified from direct code inspection and Phase 1 experience

**Research date:** 2026-03-07
**Valid until:** 2026-04-07 (stable domain -- Go 1.23, pgx v4, chi v5 are not changing)
