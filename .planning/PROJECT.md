# Dermify

## What This Is

A clinical procedure documentation API for aesthetic dermatology clinics. Each clinic gets its own deployed instance where doctors and admin staff record treatment sessions -- laser procedures (IPL, Nd:YAG, CO2, RF), injectable treatments (fillers, botulinum toxin) -- with full device/product traceability, consent management, photo documentation, and immutable audit trails. The API serves a browser-based web app used in-clinic during and after procedures.

## Core Value

A clinician can document a complete treatment session -- from patient selection through procedure parameters to sign-off -- producing a locked, auditable medical record that meets medico-legal requirements.

## Requirements

### Validated

- ✓ User registration with email/password -- existing
- ✓ JWT authentication with access/refresh token rotation -- existing
- ✓ Token revocation (logout) -- existing
- ✓ Authenticated user profile endpoint -- existing
- ✓ Structured JSON error responses with machine-readable codes -- existing
- ✓ Prometheus metrics and health check -- existing
- ✓ CORS configuration -- existing
- ✓ Swagger API documentation -- existing
- ✓ Role-based access control (Doctor/Admin) -- v1.0
- ✓ Patient management (CRUD with search/pagination) -- v1.0
- ✓ Device and product registry with seed data -- v1.0
- ✓ Treatment session lifecycle with 5-state machine -- v1.0
- ✓ Session header fields (indication codes, skin type, context flags) -- v1.0
- ✓ Consent and safety checks (consent capture, contraindication screening) -- v1.0
- ✓ IPL procedure module with device-specific parameters -- v1.0
- ✓ Nd:YAG laser module with device-specific parameters -- v1.0
- ✓ CO2/ablative laser module with device-specific parameters -- v1.0
- ✓ RF/RF microneedling module with device-specific parameters -- v1.0
- ✓ Filler injectable module with product traceability -- v1.0
- ✓ Botulinum toxin module with reconstitution and injection site mapping -- v1.0
- ✓ Outcome recording (immediate outcome, clinical endpoints) -- v1.0
- ✓ Aftercare documentation with mandatory red flags -- v1.0
- ✓ Follow-up scheduling linked to session -- v1.0
- ✓ Record sign-off with validation gate -- v1.0
- ✓ Record locking with addendum-only amendment model -- v1.0
- ✓ Database-enforced immutability via PL/pgSQL triggers -- v1.0
- ✓ Append-only audit trail on all clinical tables -- v1.0
- ✓ Before photos and label photos with consent gating -- v1.0
- ✓ Seed data for controlled lists (devices, products, indications, endpoints) -- v1.0

### Active

- [ ] Adverse event capture (type, severity, details, reporting flag)
- [ ] Per-module parameter validation against device-specific constraints
- [ ] Skin type-aware parameter warnings (Fitzpatrick type constrains safe ranges)
- [ ] UDI-DI and UDI-PI capture for device registry
- [ ] Record versioning with before/after snapshots at state transitions
- [ ] Admin management of device registry, product formulary, and controlled lists
- [ ] Application-level editability checks on module/outcome update paths (clean 409 instead of DB 500)

### Out of Scope

- AI-powered suggestions or risk scoring -- future feature, not v1
- Patient-facing portal or self-service -- clinician/admin tool only
- Multi-site organization features -- single clinic per instance
- Integration with external EMR systems -- manual entry only
- Mobile native app -- web app only
- Real-time notifications or WebSocket features -- request/response API
- OAuth / social login -- email/password auth sufficient
- Scheduling / appointment management -- separate concern
- Billing or payment processing -- out of scope entirely
- Nurse/assistant role -- Doctor and Admin only for now

## Context

- **Shipped:** v1.0 on 2026-03-08. All 63 requirements satisfied across 7 phases (23 plans).
- **Codebase:** 17,400 LOC Go + 1,028 LOC SQL migrations. 221 files. 100+ service tests.
- **Tech stack:** Go 1.23, Chi router, Cobra CLI, PostgreSQL (pgx v4), Goose embedded migrations, Prometheus metrics.
- **Architecture:** Layered monolith -- handlers -> services -> repositories with interface-based DI. Hybrid polymorphism for procedure modules (shared session_modules + per-type detail tables).
- **Database:** 13 clinical tables with PL/pgSQL immutability triggers and audit triggers. Seed data for devices, products, indications, and clinical endpoints.
- **Known tech debt:** pgx v4 is EOL (July 2025), migration to v5 deferred. Module/outcome update methods lack app-level editability checks.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Single-tenant per clinic | Simplifies data isolation, deployment, and compliance | ✓ Good |
| Addendum-only amendments | Medico-legal requirement -- original records must be immutable | ✓ Good |
| Local filesystem for photos | Avoid cloud dependencies in v1; can migrate to S3 later | ✓ Good |
| Hardcoded seed data for controlled lists | Get to working state fast; admin UI for list management deferred | ✓ Good |
| Service/repository layer from Phase 1 | Clean architecture from the start; every phase follows same pattern | ✓ Good |
| 6 procedure module types with hybrid polymorphism | Shared session_modules table + per-type detail tables; extensible | ✓ Good |
| DB-level immutability via PL/pgSQL triggers | Defense in depth -- cannot bypass via direct SQL | ✓ Good |
| DB-level audit trail via AFTER triggers | Automatic, complete -- no way to forget audit logging | ✓ Good |
| Energy modules before injectables | Validated pattern with 4 module types before applying to 2 more | ✓ Good |
| pgx v4 (EOL) kept for v1 | Avoid migration risk during feature development | ⚠️ Revisit |

## Constraints

- **Tech stack**: Go 1.23, Chi, PostgreSQL, existing project conventions per CLAUDE.md -- no framework changes
- **Architecture**: Layered monolith -- handlers, routes, middleware pattern established. Service/repository layers for all domains.
- **Linting**: Strict golangci-lint config (~60 linters). All new code must pass `make lint`.
- **File storage**: Local filesystem for photos -- no cloud storage dependencies
- **Deployment**: Docker + docker-compose. Single binary, Alpine-based image.
- **Validation**: Field validation rules per module type -- required/conditional logic and value constraints.

---
*Last updated: 2026-03-08 after v1.0 milestone*
