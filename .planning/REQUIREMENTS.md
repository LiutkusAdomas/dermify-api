# Requirements: Dermify

**Defined:** 2026-03-07
**Core Value:** A clinician can document a complete treatment session producing a locked, auditable medical record

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Access Control

- [x] **RBAC-01**: System supports Doctor and Admin roles with distinct permissions
- [x] **RBAC-02**: Doctor can perform all clinical operations (create sessions, add modules, sign off)
- [x] **RBAC-03**: Admin can manage patients and view sessions but cannot sign off or modify clinical data
- [x] **RBAC-04**: Endpoints enforce role-based authorization via middleware

### Patient Management

- [x] **PAT-01**: User can create a patient record with demographics (name, date of birth, contact)
- [x] **PAT-02**: User can search and list patients with filtering
- [x] **PAT-03**: User can update patient records
- [x] **PAT-04**: User can view a patient's session history

### Device & Product Registry

- [x] **REG-01**: System ships with seed data for energy-based devices (manufacturer, model, handpieces)
- [x] **REG-02**: System ships with seed data for injectable products (fillers, botulinum toxins with concentrations)
- [x] **REG-03**: System ships with seed data for indication codes and clinical endpoints
- [x] **REG-04**: Clinician can select devices and products from controlled lists when documenting procedures

### Treatment Sessions

- [x] **SESS-01**: Clinician can create a new treatment session linked to a patient
- [x] **SESS-02**: Session captures header fields: patient, clinician, timing, indication codes, patient goal, Fitzpatrick skin type, context flags (tan, pregnancy, anticoagulants)
- [x] **SESS-03**: Session follows lifecycle states: Draft -> In Progress -> Awaiting Sign-off -> Signed -> Locked
- [x] **SESS-04**: Server enforces valid state transitions (no skipping states)
- [x] **SESS-05**: Clinician can save a session as draft and return to it later
- [x] **SESS-06**: Clinician can add multiple procedure modules to a single session

### Consent & Safety

- [x] **CONS-01**: Clinician can record consent (type, method, datetime, risks discussed flag)
- [x] **CONS-02**: System blocks adding procedure modules until consent is captured
- [x] **CONS-03**: Clinician can complete contraindication screening checklist
- [x] **CONS-04**: System captures contraindication flags and mitigation notes
- [x] **CONS-05**: Clinician can record photo consent status (yes/no/limited)

### IPL Module

- [x] **IPL-01**: Clinician can add an IPL procedure module to a session
- [x] **IPL-02**: Module captures: filter/band, lightguide size, fluence, pulse duration, pulse delay, pulse count (MSP), passes, total pulses, cooling mode
- [x] **IPL-03**: Module links to a device from the registry with handpiece selection

### Nd:YAG Module

- [x] **YAG-01**: Clinician can add an Nd:YAG procedure module to a session
- [x] **YAG-02**: Module captures: wavelength, spot size, fluence, pulse duration, repetition rate, cooling type, total pulses
- [x] **YAG-03**: Module links to a device from the registry

### CO2/Ablative Module

- [x] **CO2-01**: Clinician can add a CO2/ablative procedure module to a session
- [x] **CO2-02**: Module captures: mode, handpiece/scanner, power, pulse energy, pulse duration, density, pattern, passes, anaesthesia used
- [x] **CO2-03**: Module links to a device from the registry

### RF Module

- [x] **RF-01**: Clinician can add an RF/RF microneedling procedure module to a session
- [x] **RF-02**: Module captures: RF mode, tip type, depth, energy level, overlap, pulses per zone, total pulses
- [x] **RF-03**: Module links to a device from the registry

### Filler Module

- [x] **FILL-01**: Clinician can add a filler procedure module to a session
- [x] **FILL-02**: Module captures: product, syringe volume, total volume used, batch/lot, expiry, needle/cannula, injection planes, anatomical sites, endpoint
- [x] **FILL-03**: Module links to a product from the registry with batch/lot and expiry tracking

### Botulinum Module

- [x] **TOX-01**: Clinician can add a botulinum toxin procedure module to a session
- [x] **TOX-02**: Module captures: product, batch number, reconstitution details (diluent, volume, concentration), total units administered, injection sites with units per site
- [x] **TOX-03**: Module links to a product from the registry with batch tracking

### Outcomes & Aftercare

- [x] **OUT-01**: Clinician can record immediate outcome (completed/partial/aborted)
- [x] **OUT-02**: Clinician can select clinical endpoints observed (module-specific list)
- [x] **OUT-03**: Clinician can record aftercare provided with templated instructions
- [x] **OUT-04**: Aftercare includes mandatory red flags and contact section
- [x] **OUT-05**: Clinician can set follow-up date/time linked to session

### Sign-off & Locking

- [x] **LOCK-01**: System validates all required fields before allowing sign-off (blocks if incomplete)
- [x] **LOCK-02**: Clinician can sign off a session (records timestamp and clinician ID)
- [x] **LOCK-03**: Signed session becomes immutable — original record cannot be modified
- [x] **LOCK-04**: Clinician can add addendums to a locked session (date, author, reason, content)
- [x] **LOCK-05**: Addendums are themselves immutable once saved
- [x] **LOCK-06**: Immutability is enforced at the database level (not just application layer)

### Audit Trail

- [x] **AUDIT-01**: System logs all create, update, and delete operations on clinical entities
- [x] **AUDIT-02**: Each audit entry captures: action, timestamp, user ID, entity type, entity ID
- [x] **AUDIT-03**: Audit log is append-only — entries cannot be modified or deleted
- [x] **AUDIT-04**: Sign-off and lock events are recorded in the audit trail

### Photo Documentation

- [ ] **PHOTO-01**: Clinician can upload before photos linked to a session
- [ ] **PHOTO-02**: Clinician can upload product label photos linked to a procedure module
- [ ] **PHOTO-03**: Photos are stored on the local filesystem with organized naming
- [ ] **PHOTO-04**: Photos require consent flag to be set before upload

### Metadata

- [x] **META-01**: All clinical records track created_at, created_by, updated_at, updated_by
- [x] **META-02**: Signed records track signed_at, signed_by
- [x] **META-03**: Records maintain an incrementing version number for medico-legal defensibility

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Safety

- **AE-01**: Clinician can record adverse events with structured type, severity, and details
- **AE-02**: System flags events that meet mandatory reporting thresholds
- **AE-03**: Adverse events link to specific procedure modules within a session

### Validation

- **VAL-01**: Per-module parameter validation against device-specific constraints (fluence ranges, depth limits)
- **VAL-02**: Skin type-aware parameter warnings (Fitzpatrick type constrains safe ranges)

### Traceability

- **UDI-01**: Device registry supports UDI-DI and UDI-PI capture
- **VER-01**: Record versioning with before/after snapshots at state transitions

### Administration

- **ADMIN-01**: Admin can manage device registry (add, edit, deactivate devices)
- **ADMIN-02**: Admin can manage product formulary
- **ADMIN-03**: Admin can manage indication codes and clinical endpoint lists

## Out of Scope

| Feature | Reason |
|---------|--------|
| AI documentation / speech-to-text | Future feature — requires ML infrastructure |
| Injection site visual mapping (diagram annotation) | Frontend concern — API stores structured data |
| Patient-facing portal | Clinician/admin tool only |
| Billing / CPT coding | Separate domain entirely |
| Appointment scheduling | Separate concern — follow-up is just a date reference |
| Multi-tenant / multi-site | Single-tenant per clinic deployment model |
| Real-time notifications / WebSocket | Request/response API sufficient for clinical documentation |
| External EMR integration (HL7/FHIR) | Manual entry only for v1 |
| OAuth / social login | Email/password auth sufficient |
| Mobile native app | Web app only |
| Complex reporting / analytics | Frontend or separate service concern |
| Treatment plan / multi-session course management | Sessions are atomic — history shown via patient view |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| RBAC-01 | Phase 1 | Complete |
| RBAC-02 | Phase 1 | Complete |
| RBAC-03 | Phase 1 | Complete |
| RBAC-04 | Phase 1 | Complete |
| PAT-01 | Phase 1 | Complete |
| PAT-02 | Phase 1 | Complete |
| PAT-03 | Phase 1 | Complete |
| PAT-04 | Phase 1 | Complete |
| REG-01 | Phase 1 | Complete |
| REG-02 | Phase 1 | Complete |
| REG-03 | Phase 1 | Complete |
| REG-04 | Phase 1 | Complete |
| META-01 | Phase 1 | Complete |
| META-02 | Phase 5 | Complete |
| META-03 | Phase 1 | Complete |
| SESS-01 | Phase 2 | Complete |
| SESS-02 | Phase 2 | Complete |
| SESS-03 | Phase 2 | Complete |
| SESS-04 | Phase 2 | Complete |
| SESS-05 | Phase 2 | Complete |
| SESS-06 | Phase 2 | Complete |
| CONS-01 | Phase 2 | Complete |
| CONS-02 | Phase 2 | Complete |
| CONS-03 | Phase 2 | Complete |
| CONS-04 | Phase 2 | Complete |
| CONS-05 | Phase 2 | Complete |
| IPL-01 | Phase 3 | Complete |
| IPL-02 | Phase 3 | Complete |
| IPL-03 | Phase 3 | Complete |
| YAG-01 | Phase 3 | Complete |
| YAG-02 | Phase 3 | Complete |
| YAG-03 | Phase 3 | Complete |
| CO2-01 | Phase 3 | Complete |
| CO2-02 | Phase 3 | Complete |
| CO2-03 | Phase 3 | Complete |
| RF-01 | Phase 3 | Complete |
| RF-02 | Phase 3 | Complete |
| RF-03 | Phase 3 | Complete |
| FILL-01 | Phase 4 | Complete |
| FILL-02 | Phase 4 | Complete |
| FILL-03 | Phase 4 | Complete |
| TOX-01 | Phase 4 | Complete |
| TOX-02 | Phase 4 | Complete |
| TOX-03 | Phase 4 | Complete |
| OUT-01 | Phase 4 | Complete |
| OUT-02 | Phase 4 | Complete |
| OUT-03 | Phase 4 | Complete |
| OUT-04 | Phase 4 | Complete |
| OUT-05 | Phase 4 | Complete |
| LOCK-01 | Phase 5 | Complete |
| LOCK-02 | Phase 5 | Complete |
| LOCK-03 | Phase 5 | Complete |
| LOCK-04 | Phase 5 | Complete |
| LOCK-05 | Phase 5 | Complete |
| LOCK-06 | Phase 5 | Complete |
| AUDIT-01 | Phase 5 | Complete |
| AUDIT-02 | Phase 5 | Complete |
| AUDIT-03 | Phase 5 | Complete |
| AUDIT-04 | Phase 5 | Complete |
| PHOTO-01 | Phase 6 | Pending |
| PHOTO-02 | Phase 6 | Pending |
| PHOTO-03 | Phase 6 | Pending |
| PHOTO-04 | Phase 6 | Pending |

**Coverage:**
- v1 requirements: 63 total
- Mapped to phases: 63
- Unmapped: 0

---
*Requirements defined: 2026-03-07*
*Last updated: 2026-03-07 after roadmap creation*
