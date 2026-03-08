---
phase: 05-sign-off-and-compliance
plan: 03
subsystem: api, handlers, routes, metrics
tags: [sign-off, addendum, audit-trail, http-handlers, prometheus, chi-routes, dependency-injection]

# Dependency graph
requires:
  - phase: 05-sign-off-and-compliance
    plan: 01
    provides: "Database triggers, immutability constraints, audit trail tables"
  - phase: 05-sign-off-and-compliance
    plan: 02
    provides: "SignoffService, AddendumService, AuditService, Postgres repositories"

provides:
  - "HTTP handlers for sign-off readiness, sign-off, lock, addendum CRUD, and audit trail"
  - "Error codes for sign-off, addendum, and audit operations"
  - "Prometheus counters: session_signed_total, session_locked_total, addendum_created_total"
  - "Route wiring under /api/v1/sessions/{id}/ for 7 new endpoints"
  - "Complete DI wiring in Manager for signoff, addendum, and audit subsystems"

affects:
  - internal/api/apierrors/apierrors.go
  - internal/api/metrics/prometheus.go
  - internal/api/metrics/metrics.go
  - internal/api/routes/sessions.go
  - internal/api/routes/manager.go

# Tech stack
added: []
patterns:
  - "Shared handleSignOffError in dedicated signoff_errors.go (same as energy/injectable pattern)"
  - "Addendum error handlers local to addendum.go since single domain"
  - "Audit handler uses query params (entity_type, entity_id) not URL params"

# Key files
created:
  - internal/api/handlers/signoff.go
  - internal/api/handlers/signoff_errors.go
  - internal/api/handlers/addendum.go
  - internal/api/handlers/audit.go

modified:
  - internal/api/apierrors/apierrors.go
  - internal/api/metrics/prometheus.go
  - internal/api/metrics/metrics.go
  - internal/api/routes/sessions.go
  - internal/api/routes/manager.go

# Decisions
key-decisions:
  - "Shared handleSignOffError in dedicated file for clean separation across sign-off and lock handlers"
  - "Addendum create/get error handlers kept local to addendum.go (single-domain scope)"
  - "Audit trail uses query parameters for entity_type/entity_id filtering (not nested URL structure)"

# Metrics
duration: 3min
completed: "2026-03-08"
tasks_completed: 2
tasks_total: 2
files_changed: 9
---

# Phase 05 Plan 03: HTTP Layer for Sign-off, Addendum, and Audit Trail Summary

REST handlers, error codes, Prometheus metrics, and route wiring completing the Phase 5 HTTP layer for clinician session sign-off, addendum creation, and audit trail querying.

## What Was Built

### Error Codes (apierrors.go)
- Sign-off: SignoffSessionIncomplete, SignoffNotReady, SignoffFailed, SignoffLockFailed
- Addendum: AddendumNotFound, AddendumInvalidData, AddendumSessionNotLocked, AddendumCreationFailed
- Audit: AuditLookupFailed

### Prometheus Metrics
- `dermify_session_signed_total` -- incremented on successful sign-off
- `dermify_session_locked_total` -- incremented on successful lock
- `dermify_addendum_created_total` -- incremented on addendum creation

### Sign-off Handlers (signoff.go, signoff_errors.go)
- `HandleGetSignOffReadiness` -- returns ValidationResult with ready flag and missing items list
- `HandleSignOffSession` -- validates completeness, records clinician ID, transitions to signed
- `HandleLockSession` -- transitions signed session to locked (immutable)
- `handleSignOffError` -- maps ErrSessionNotFound, ErrSessionIncomplete, ErrSessionNotAwaitingSignoff, ErrInvalidStateTransition, ErrSessionVersionConflict to HTTP codes

### Addendum Handlers (addendum.go)
- `HandleCreateAddendum` -- creates addendum on locked session (enforces session lock check)
- `HandleListAddendums` -- lists all addendums for a session
- `HandleGetAddendum` -- retrieves single addendum by ID

### Audit Handler (audit.go)
- `HandleGetAuditTrail` -- queries audit entries by entity_type and entity_id query params

### Route Wiring (sessions.go, manager.go)
- 7 new endpoints registered under `/api/v1/sessions/{id}/`:
  - `GET /signoff/readiness`
  - `POST /signoff`
  - `POST /lock`
  - `POST /addendums`
  - `GET /addendums`
  - `GET /addendums/{addendumId}`
  - `GET /audit`
- Manager creates SignoffRepository, AddendumRepository, AuditRepository
- Manager wires SignoffService, AddendumService, AuditService via DI

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | e5e3da3 | Handler files, error codes, and metrics |
| 2 | fe730b6 | Route wiring and dependency injection |

## Deviations from Plan

None -- plan executed exactly as written.

## Verification

- `go build ./...` succeeds
- `go test ./... -count=1` passes (all existing + new tests)
- `go vet ./...` clean
- All 7 new endpoints registered in route tree

## Self-Check: PASSED

All 9 files verified on disk. Both commit hashes (e5e3da3, fe730b6) confirmed in git log.
