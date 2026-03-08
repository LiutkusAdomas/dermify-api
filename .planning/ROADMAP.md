# Roadmap: Dermify

## Overview

Dermify is a clinical procedure documentation API for aesthetic dermatology. The existing codebase provides auth, health checks, and metrics. This roadmap delivers the remaining 63 v1 requirements in 6 phases: establishing the service/repository architecture and foundational domains (RBAC, patients, registry), building the session lifecycle with consent gates, implementing all 6 procedure module types across two phases (energy-based then injectable), adding medico-legal compliance (sign-off, locking, audit trail), and finishing with photo documentation. Each phase delivers a coherent, verifiable capability that unblocks the next.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation** - Service/repository architecture, RBAC, patient management, device/product registry, and metadata tracking (completed 2026-03-07)
- [ ] **Phase 2: Session Lifecycle** - Treatment session creation, header fields, state machine, and consent/safety gates
- [ ] **Phase 3: Energy-Based Modules** - IPL, Nd:YAG, CO2/ablative, and RF procedure modules with device linkage
- [ ] **Phase 4: Injectable Modules and Outcomes** - Filler and botulinum toxin modules with product traceability, plus outcomes and aftercare
- [ ] **Phase 5: Sign-off and Compliance** - Validation gates, record sign-off, database-enforced immutability, addendums, and audit trail
- [ ] **Phase 6: Photo Documentation** - Before photos, label photos, filesystem storage, and consent-gated uploads

## Phase Details

### Phase 1: Foundation
**Goal**: Doctors and admins can authenticate with role-appropriate permissions, manage patient records, and browse the device/product registry -- all built on a service/repository architecture that supports the clinical domains ahead
**Depends on**: Nothing (first phase)
**Requirements**: RBAC-01, RBAC-02, RBAC-03, RBAC-04, PAT-01, PAT-02, PAT-03, PAT-04, REG-01, REG-02, REG-03, REG-04, META-01, META-03
**Success Criteria** (what must be TRUE):
  1. A Doctor user can access clinical endpoints (create patients, view registry) while an Admin user can manage patients but is blocked from clinical-only actions
  2. A user can create a patient record, search/filter the patient list, update patient details, and view a patient's session history (empty at this stage)
  3. The system returns pre-loaded device entries (energy-based devices with manufacturers, models, handpieces) and product entries (fillers, botulinum toxins) from seed data
  4. All created/updated clinical records carry created_at, created_by, updated_at, updated_by metadata automatically
  5. Unauthorized role access to any protected endpoint returns a structured JSON error with appropriate HTTP status
**Plans**: 6 plans

Plans:
- [x] 01-00-PLAN.md -- Test infrastructure (Wave 0: mock repositories, test stubs, Makefile test target)
- [x] 01-01-PLAN.md -- Domain models and service/repository scaffold (domain types, RoleService, PostgresRoleRepository)
- [x] 01-02-PLAN.md -- RBAC system (role migration, JWT claims, RequireRole middleware, role assignment endpoint, first-user bootstrap)
- [x] 01-03-PLAN.md -- Patient management with metadata tracking (patient CRUD, search/pagination, optimistic locking, metadata)
- [x] 01-04-PLAN.md -- Device/product registry with seed data (schema tables, seed migrations, read-only list/detail endpoints)
- [x] 01-05-PLAN.md -- Gap closure: handler-level RBAC tests and META-02 requirement tracking fix

### Phase 2: Session Lifecycle
**Goal**: A clinician can create a treatment session for a patient, fill in clinical header fields, record consent and safety screening, and move the session through its lifecycle states -- producing a structured draft record ready for procedure modules
**Depends on**: Phase 1
**Requirements**: SESS-01, SESS-02, SESS-03, SESS-04, SESS-05, SESS-06, CONS-01, CONS-02, CONS-03, CONS-04, CONS-05
**Success Criteria** (what must be TRUE):
  1. A clinician can create a new session linked to an existing patient, filling in header fields (indication codes, patient goal, Fitzpatrick skin type, context flags) and saving it as a draft to return to later
  2. A session progresses through Draft, In Progress, Awaiting Sign-off, Signed, and Locked states -- and the server rejects any invalid state transition attempt with a clear error
  3. The system blocks adding procedure modules to a session until consent is recorded (type, method, datetime, risks discussed)
  4. A clinician can complete contraindication screening (checklist, flags, mitigation notes) and record photo consent status before proceeding
  5. A session supports multiple procedure modules (the slots exist even though module types are delivered in later phases)
**Plans**: 4 plans

Plans:
- [ ] 02-01-PLAN.md -- Domain models, migrations, service interfaces, and mock repositories (contracts foundation)
- [ ] 02-02-PLAN.md -- Session service with state machine and PostgreSQL repository
- [ ] 02-03-PLAN.md -- Consent, contraindication, and module services with PostgreSQL repositories
- [ ] 02-04-PLAN.md -- HTTP handlers, route wiring, metrics, and patient history integration

### Phase 3: Energy-Based Modules
**Goal**: A clinician can document energy-based procedures (IPL, Nd:YAG, CO2/ablative, RF) within a treatment session, selecting devices from the registry and recording all device-specific parameters
**Depends on**: Phase 2
**Requirements**: IPL-01, IPL-02, IPL-03, YAG-01, YAG-02, YAG-03, CO2-01, CO2-02, CO2-03, RF-01, RF-02, RF-03
**Success Criteria** (what must be TRUE):
  1. A clinician can add an IPL module to a session, selecting a device and handpiece from the registry, and recording all IPL-specific parameters (filter/band, lightguide, fluence, pulse duration, pulse delay, pulse count, passes, cooling mode)
  2. A clinician can add Nd:YAG, CO2/ablative, and RF modules to a session with their respective device-specific parameter sets fully captured
  3. Each module links to a device from the registry -- attempting to reference a non-existent device returns an error
  4. A single session can contain multiple modules of different types (e.g., IPL treatment on one area plus RF on another)
**Plans**: 3 plans

Plans:
- [ ] 03-01-PLAN.md -- Domain types, migrations, repository interfaces, service scaffold, and mock repositories (contracts foundation)
- [ ] 03-02-PLAN.md -- Postgres repositories for all 4 module types and EnergyModuleService unit tests
- [ ] 03-03-PLAN.md -- HTTP handlers, route wiring, metrics, and dependency injection for all 4 module types

### Phase 4: Injectable Modules and Outcomes
**Goal**: A clinician can document injectable procedures (fillers, botulinum toxin) with full product traceability, then record outcomes, aftercare instructions, and follow-up scheduling for the complete session
**Depends on**: Phase 3
**Requirements**: FILL-01, FILL-02, FILL-03, TOX-01, TOX-02, TOX-03, OUT-01, OUT-02, OUT-03, OUT-04, OUT-05
**Success Criteria** (what must be TRUE):
  1. A clinician can add a filler module to a session with product selection from the registry, batch/lot number, expiry date, syringe volume, injection details, and anatomical sites
  2. A clinician can add a botulinum toxin module with product, batch tracking, reconstitution details (diluent, volume, concentration), total units, and per-site injection mapping
  3. A clinician can record the session outcome (completed/partial/aborted), select observed clinical endpoints from a module-specific list, and document aftercare with templated instructions including mandatory red flags
  4. A clinician can set a follow-up date/time linked to the session
**Plans**: 3 plans

Plans:
- [x] 04-01-PLAN.md -- Domain types, migrations, service scaffolds with product validation and outcome logic, mock repositories
- [ ] 04-02-PLAN.md -- Postgres repositories for filler, botulinum, and outcome tables plus service unit tests
- [ ] 04-03-PLAN.md -- HTTP handlers, route wiring, metrics, and dependency injection for injectable modules and outcomes

### Phase 5: Sign-off and Compliance
**Goal**: A clinician can sign off a completed session, producing a locked, immutable medical record with a full audit trail -- the core medico-legal requirement of the system
**Depends on**: Phase 4
**Requirements**: LOCK-01, LOCK-02, LOCK-03, LOCK-04, LOCK-05, LOCK-06, AUDIT-01, AUDIT-02, AUDIT-03, AUDIT-04, META-02
**Success Criteria** (what must be TRUE):
  1. The system validates all required fields across the session (header, consent, modules, outcomes) and blocks sign-off with a clear list of what is incomplete
  2. A clinician can sign off a valid session, which records the timestamp and clinician ID, and the signed record becomes immutable -- any attempt to modify it (via API or direct SQL) fails
  3. A clinician can add addendums to a locked session (with date, author, reason, content) and those addendums are themselves immutable once saved
  4. Every create, update, sign-off, and lock operation on clinical entities is recorded in an append-only audit trail with action, timestamp, user ID, entity type, and entity ID
  5. Immutability is enforced at the database level via triggers -- not just application-layer checks
**Plans**: 3 plans

Plans:
- [ ] 05-01-PLAN.md -- Domain types, SQL migrations (signed_at/signed_by, addendums, audit trail, immutability triggers, audit triggers), service interfaces, mock repositories
- [ ] 05-02-PLAN.md -- Postgres repositories for signoff, addendum, and audit, plus service unit tests
- [ ] 05-03-PLAN.md -- HTTP handlers, route wiring, metrics, and DI for sign-off, addendum, and audit endpoints

### Phase 6: Photo Documentation
**Goal**: A clinician can attach before photos and product label photos to sessions, stored on the local filesystem with proper consent gating
**Depends on**: Phase 5
**Requirements**: PHOTO-01, PHOTO-02, PHOTO-03, PHOTO-04
**Success Criteria** (what must be TRUE):
  1. A clinician can upload before photos linked to a session and product label photos linked to a specific procedure module
  2. Photos are stored on the local filesystem with organized, predictable naming (not in the database)
  3. The system blocks photo uploads if the session's photo consent flag has not been set
**Plans**: 3 plans

Plans:
- [ ] 06-01-PLAN.md -- Photo domain type, SQL migration, PhotoService with consent gate and FileStore interface, mock repositories
- [ ] 06-02-PLAN.md -- PostgreSQL photo repository, StorageConfig, and PhotoService unit tests
- [ ] 06-03-PLAN.md -- HTTP handlers, LocalFileStore, error codes, metrics, route wiring, and DI for photo endpoints

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 6/6 | Complete   | 2026-03-07 |
| 2. Session Lifecycle | 4/4 | Complete |  |
| 3. Energy-Based Modules | 3/3 | Complete | 2026-03-08 |
| 4. Injectable Modules and Outcomes | 2/3 | In Progress|  |
| 5. Sign-off and Compliance | 0/3 | Not started | - |
| 6. Photo Documentation | 1/3 | In Progress|  |
