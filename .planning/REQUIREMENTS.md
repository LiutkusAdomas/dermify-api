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

- [ ] **PAT-01**: User can create a patient record with demographics (name, date of birth, contact)
- [ ] **PAT-02**: User can search and list patients with filtering
- [ ] **PAT-03**: User can update patient records
- [ ] **PAT-04**: User can view a patient's session history

### Device & Product Registry

- [x] **REG-01**: System ships with seed data for energy-based devices (manufacturer, model, handpieces)
- [x] **REG-02**: System ships with seed data for injectable products (fillers, botulinum toxins with concentrations)
- [x] **REG-03**: System ships with seed data for indication codes and clinical endpoints
- [x] **REG-04**: Clinician can select devices and products from controlled lists when documenting procedures

### Treatment Sessions

- [ ] **SESS-01**: Clinician can create a new treatment session linked to a patient
- [ ] **SESS-02**: Session captures header fields: patient, clinician, timing, indication codes, patient goal, Fitzpatrick skin type, context flags (tan, pregnancy, anticoagulants)
- [ ] **SESS-03**: Session follows lifecycle states: Draft -> In Progress -> Awaiting Sign-off -> Signed -> Locked
- [ ] **SESS-04**: Server enforces valid state transitions (no skipping states)
- [ ] **SESS-05**: Clinician can save a session as draft and return to it later
- [ ] **SESS-06**: Clinician can add multiple procedure modules to a single session

### Consent & Safety

- [ ] **CONS-01**: Clinician can record consent (type, method, datetime, risks discussed flag)
- [ ] **CONS-02**: System blocks adding procedure modules until consent is captured
- [ ] **CONS-03**: Clinician can complete contraindication screening checklist
- [ ] **CONS-04**: System captures contraindication flags and mitigation notes
- [ ] **CONS-05**: Clinician can record photo consent status (yes/no/limited)

### IPL Module

- [ ] **IPL-01**: Clinician can add an IPL procedure module to a session
- [ ] **IPL-02**: Module captures: filter/band, lightguide size, fluence, pulse duration, pulse delay, pulse count (MSP), passes, total pulses, cooling mode
- [ ] **IPL-03**: Module links to a device from the registry with handpiece selection

### Nd:YAG Module

- [ ] **YAG-01**: Clinician can add an Nd:YAG procedure module to a session
- [ ] **YAG-02**: Module captures: wavelength, spot size, fluence, pulse duration, repetition rate, cooling type, total pulses
- [ ] **YAG-03**: Module links to a device from the registry

### CO2/Ablative Module

- [ ] **CO2-01**: Clinician can add a CO2/ablative procedure module to a session
- [ ] **CO2-02**: Module captures: mode, handpiece/scanner, power, pulse energy, pulse duration, density, pattern, passes, anaesthesia used
- [ ] **CO2-03**: Module links to a device from the registry

### RF Module

- [ ] **RF-01**: Clinician can add an RF/RF microneedling procedure module to a session
- [ ] **RF-02**: Module captures: RF mode, tip type, depth, energy level, overlap, pulses per zone, total pulses
- [ ] **RF-03**: Module links to a device from the registry

### Filler Module

- [ ] **FILL-01**: Clinician can add a filler procedure module to a session
- [ ] **FILL-02**: Module captures: product, syringe volume, total volume used, batch/lot, expiry, needle/cannula, injection planes, anatomical sites, endpoint
- [ ] **FILL-03**: Module links to a product from the registry with batch/lot and expiry tracking

### Botulinum Module

- [ ] **TOX-01**: Clinician can add a botulinum toxin procedure module to a session
- [ ] **TOX-02**: Module captures: product, batch number, reconstitution details (diluent, volume, concentration), total units administered, injection sites with units per site
- [ ] **TOX-03**: Module links to a product from the registry with batch tracking

### Outcomes & Aftercare

- [ ] **OUT-01**: Clinician can record immediate outcome (completed/partial/aborted)
- [ ] **OUT-02**: Clinician can select clinical endpoints observed (module-specific list)
- [ ] **OUT-03**: Clinician can record aftercare provided with templated instructions
- [ ] **OUT-04**: Aftercare includes mandatory red flags and contact section
- [ ] **OUT-05**: Clinician can set follow-up date/time linked to session

### Sign-off & Locking

- [ ] **LOCK-01**: System validates all required fields before allowing sign-off (blocks if incomplete)
- [ ] **LOCK-02**: Clinician can sign off a session (records timestamp and clinician ID)
- [ ] **LOCK-03**: Signed session becomes immutable — original record cannot be modified
- [ ] **LOCK-04**: Clinician can add addendums to a locked session (date, author, reason, content)
- [ ] **LOCK-05**: Addendums are themselves immutable once saved
- [ ] **LOCK-06**: Immutability is enforced at the database level (not just application layer)

### Audit Trail

- [ ] **AUDIT-01**: System logs all create, update, and delete operations on clinical entities
- [ ] **AUDIT-02**: Each audit entry captures: action, timestamp, user ID, entity type, entity ID
- [ ] **AUDIT-03**: Audit log is append-only — entries cannot be modified or deleted
- [ ] **AUDIT-04**: Sign-off and lock events are recorded in the audit trail

### Photo Documentation

- [ ] **PHOTO-01**: Clinician can upload before photos linked to a session
- [ ] **PHOTO-02**: Clinician can upload product label photos linked to a procedure module
- [ ] **PHOTO-03**: Photos are stored on the local filesystem with organized naming
- [ ] **PHOTO-04**: Photos require consent flag to be set before upload

### Metadata

- [x] **META-01**: All clinical records track created_at, created_by, updated_at, updated_by
- [ ] **META-02**: Signed records track signed_at, signed_by
- [ ] **META-03**: Records maintain an incrementing version number for medico-legal defensibility

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
| PAT-01 | Phase 1 | Pending |
| PAT-02 | Phase 1 | Pending |
| PAT-03 | Phase 1 | Pending |
| PAT-04 | Phase 1 | Pending |
| REG-01 | Phase 1 | Complete |
| REG-02 | Phase 1 | Complete |
| REG-03 | Phase 1 | Complete |
| REG-04 | Phase 1 | Complete |
| META-01 | Phase 1 | Complete |
| META-02 | Phase 1 | Pending |
| META-03 | Phase 1 | Pending |
| SESS-01 | Phase 2 | Pending |
| SESS-02 | Phase 2 | Pending |
| SESS-03 | Phase 2 | Pending |
| SESS-04 | Phase 2 | Pending |
| SESS-05 | Phase 2 | Pending |
| SESS-06 | Phase 2 | Pending |
| CONS-01 | Phase 2 | Pending |
| CONS-02 | Phase 2 | Pending |
| CONS-03 | Phase 2 | Pending |
| CONS-04 | Phase 2 | Pending |
| CONS-05 | Phase 2 | Pending |
| IPL-01 | Phase 3 | Pending |
| IPL-02 | Phase 3 | Pending |
| IPL-03 | Phase 3 | Pending |
| YAG-01 | Phase 3 | Pending |
| YAG-02 | Phase 3 | Pending |
| YAG-03 | Phase 3 | Pending |
| CO2-01 | Phase 3 | Pending |
| CO2-02 | Phase 3 | Pending |
| CO2-03 | Phase 3 | Pending |
| RF-01 | Phase 3 | Pending |
| RF-02 | Phase 3 | Pending |
| RF-03 | Phase 3 | Pending |
| FILL-01 | Phase 4 | Pending |
| FILL-02 | Phase 4 | Pending |
| FILL-03 | Phase 4 | Pending |
| TOX-01 | Phase 4 | Pending |
| TOX-02 | Phase 4 | Pending |
| TOX-03 | Phase 4 | Pending |
| OUT-01 | Phase 4 | Pending |
| OUT-02 | Phase 4 | Pending |
| OUT-03 | Phase 4 | Pending |
| OUT-04 | Phase 4 | Pending |
| OUT-05 | Phase 4 | Pending |
| LOCK-01 | Phase 5 | Pending |
| LOCK-02 | Phase 5 | Pending |
| LOCK-03 | Phase 5 | Pending |
| LOCK-04 | Phase 5 | Pending |
| LOCK-05 | Phase 5 | Pending |
| LOCK-06 | Phase 5 | Pending |
| AUDIT-01 | Phase 5 | Pending |
| AUDIT-02 | Phase 5 | Pending |
| AUDIT-03 | Phase 5 | Pending |
| AUDIT-04 | Phase 5 | Pending |
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
