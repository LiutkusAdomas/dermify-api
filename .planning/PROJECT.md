# Dermify

## What This Is

A clinical procedure documentation API for aesthetic dermatology clinics. Each clinic gets its own deployed instance where doctors and admin staff record treatment sessions — laser procedures, injectable treatments — with full device traceability, consent management, adverse event reporting, and immutable audit trails. The API serves a browser-based web app used in-clinic during and after procedures.

## Core Value

A clinician can document a complete treatment session — from patient selection through procedure parameters to sign-off — producing a locked, auditable medical record that meets medico-legal requirements.

## Requirements

### Validated

- ✓ User registration with email/password — existing
- ✓ JWT authentication with access/refresh token rotation — existing
- ✓ Token revocation (logout) — existing
- ✓ Authenticated user profile endpoint — existing
- ✓ Structured JSON error responses with machine-readable codes — existing
- ✓ Prometheus metrics and health check — existing
- ✓ CORS configuration — existing
- ✓ Swagger API documentation — existing

### Active

- [ ] Role-based access control (Doctor/Clinician, Admin/Receptionist)
- [ ] Patient management (manual CRUD)
- [ ] Treatment session lifecycle (create, draft, add modules, sign-off, lock)
- [ ] Session header fields (patient, clinician, timing, indication, skin type, context)
- [ ] Consent and safety checks (consent capture, contraindication screening)
- [ ] Device and product traceability (asset registry, batch/lot, expiry, UDI)
- [ ] IPL procedure module with device-specific parameters
- [ ] Nd:YAG / long-pulsed laser module with device-specific parameters
- [ ] CO2 / ablative laser module with device-specific parameters
- [ ] RF / RF microneedling module with device-specific parameters
- [ ] Filler injectable module with product-specific fields
- [ ] Botulinum toxin module with reconstitution and injection site mapping
- [ ] Outcome recording (immediate outcome, clinical endpoints)
- [ ] Adverse event capture (type, severity, details, reporting flag)
- [ ] Aftercare documentation (templated instructions with customisation)
- [ ] Follow-up scheduling (date/time, linked to session)
- [ ] Record sign-off with validation gate (block if required checks incomplete)
- [ ] Record locking with addendum-only amendment model
- [ ] Audit trail (created/updated/signed metadata, record versioning)
- [ ] Before photos and label photos (local filesystem storage)
- [ ] Seed data for controlled lists (devices, products, indication codes, clinical endpoints)

### Out of Scope

- AI-powered suggestions or risk scoring — future feature, not v1
- Patient-facing portal or self-service — clinician/admin tool only
- Multi-site organization features — single clinic per instance
- Integration with external EMR systems — manual entry only for v1
- Mobile native app — web app only
- Real-time notifications or WebSocket features — request/response API
- OAuth / social login — email/password auth sufficient
- Scheduling / appointment management — separate concern
- Billing or payment processing — out of scope entirely
- Nurse/assistant role — v1 has Doctor and Admin only

## Context

- **Existing codebase:** Go 1.23 REST API with Chi router, Cobra CLI, PostgreSQL (pgx v4), Goose embedded migrations, Prometheus metrics. Auth flow (register, login, logout, refresh, me) is complete and working.
- **Deployment model:** Single-tenant — each clinic gets its own deployed instance with its own database. No multi-tenancy at the application level.
- **Data model source:** Detailed field specifications from 12 spreadsheet tables covering core session fields, consent, devices, outcomes, and 6 procedure-specific module types (IPL, Nd:YAG, CO2/Erbium, RF, Fillers, Botulinum).
- **Flow diagram:** Session flow documented — start header, add procedure modules (energy-based or injectable), capture outcomes/aftercare, validate required checks, lock record with audit entry.
- **Controlled lists:** Device registry, product formulary, indication codes, and clinical endpoint lists ship as hardcoded seed data. Admin management UI deferred.
- **Sign-off model:** Once a record is signed off, it becomes immutable. Only addendums (additional notes) can be attached — original record cannot be modified.
- **Done criteria:** All API endpoints pass integration tests with test data covering the full session flow.

## Constraints

- **Tech stack**: Go 1.23, Chi, PostgreSQL, existing project conventions per CLAUDE.md — no framework changes
- **Architecture**: Layered monolith — handlers, routes, middleware pattern already established. Add service/repository layers as complexity grows.
- **Linting**: Strict golangci-lint config (~60 linters). All new code must pass `make lint`.
- **File storage**: Local filesystem for photos — no cloud storage dependencies in v1
- **Deployment**: Docker + docker-compose. Single binary, Alpine-based image.
- **Validation**: Field validation rules defined in spreadsheet — must honour required/conditional logic and value constraints per module type.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Single-tenant per clinic | Simplifies data isolation, deployment, and compliance | -- Pending |
| Addendum-only amendments | Medico-legal requirement — original records must be immutable | -- Pending |
| Local filesystem for photos | Avoid cloud dependencies in v1; can migrate to S3 later | -- Pending |
| Hardcoded seed data for controlled lists | Get to working state fast; admin UI for list management deferred | -- Pending |
| No service/repository layer yet | Existing code uses direct SQL in handlers; introduce layers as domains grow | -- Pending |
| 6 procedure module types | IPL, Nd:YAG, CO2, RF, Fillers, Botulinum — all required for full session flow | -- Pending |

---
*Last updated: 2026-03-07 after initialization*
