# Feature Landscape

**Domain:** Clinical procedure documentation API for aesthetic dermatology clinics
**Researched:** 2026-03-07

## Table Stakes

Features users expect from any credible aesthetic procedure documentation system. Missing any of these makes the product feel incomplete or unusable for clinical practice.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Patient management (CRUD)** | Every session links to a patient; can't document without one | Low | Demographics, contact, medical history basics. Single-tenant so no cross-clinic sharing needed. |
| **Treatment session lifecycle** | Core unit of work -- clinicians think in "sessions" not individual records | High | States: Draft -> In Progress -> Awaiting Sign-off -> Signed/Locked. Must enforce state transitions server-side. |
| **Session header fields** | Every competitor captures patient, clinician, date/time, indication, skin type | Medium | Fitzpatrick skin type (I-VI), clinical indication, treatment context. Drives downstream parameter validation (e.g., skin type constrains safe fluence ranges). |
| **Consent capture** | Medico-legal requirement -- no consent = no treatment. Every aesthetic EMR has this. | Medium | Consent type, method (written/verbal/digital), timestamp, and link to session. Must be captured before procedure modules can be added. Blocking gate. |
| **Contraindication screening** | Standard of care -- failing to screen is a liability. Regulatory expectation. | Medium | Checklist-style screening questions. Must be answered before procedure. Results stored with session. Can block session progression if contraindications detected. |
| **Energy-based device procedure modules** | Core clinical documentation -- parameters differ per device type and must be captured precisely | High | Six module types required (IPL, Nd:YAG, CO2/ablative, RF/RF microneedling, fillers, botulinum). Each has device-specific parameter schemas. See "Procedure Module Detail" section below. |
| **Injectable procedure modules** | Fillers and botulinum toxin are the bread and butter of aesthetic clinics | High | Product identification (name, manufacturer, batch/lot, expiry), reconstitution details (for botulinum), injection sites with units/volume per site, total dose calculation. |
| **Device and product registry** | Clinicians must select from known devices/products -- free-text is unacceptable for traceability | Medium | Seed data for devices (manufacturer, model, type, UDI) and products (filler/toxin name, manufacturer, concentration). Admin management deferred -- hardcoded seed data for v1. |
| **Batch/lot and expiry tracking** | Regulatory requirement (FDA 21 CFR 803, EU MDR). Every aesthetic EMR competitor tracks this. | Medium | Batch number, lot number, expiry date linked to each procedure module. Required for adverse event traceability back to specific product batches. |
| **Record sign-off with validation gate** | Medico-legal requirement -- unsigned records have no legal standing | Medium | Clinician digitally signs the session. System validates all required fields are populated, consent is captured, and mandatory checks are complete before allowing sign-off. Timestamp + clinician ID recorded. |
| **Record locking (immutability)** | Medico-legal standard -- signed records must not be modifiable | Medium | Once signed, the session record becomes read-only. Original content preserved permanently. Only addendums (new notes attached to the record) permitted after lock. |
| **Addendum-only amendment model** | Legal standard for medical record corrections (Noridian guidelines, HIPAA) | Low | Addendums include: current date/time, author, reason, content. Original record untouched. Each addendum is itself immutable once saved. |
| **Audit trail** | HIPAA requirement. Medico-legal expectation. Every EHR must have this. | Medium | Who accessed/created/modified what, when, from where. Append-only log. Must capture: action type, timestamp, user ID, entity affected, before/after state for modifications. |
| **Outcome recording** | Clinicians need to document what happened -- immediate clinical endpoints | Low | Immediate outcome (e.g., erythema grade, patient-reported pain score), clinical endpoints achieved (e.g., blanching for vascular lesions, frosting for tattoo removal). Structured fields, not free text. |
| **Aftercare documentation** | Standard of care -- patients must receive post-procedure instructions | Low | Templated aftercare instructions selected per procedure type, with customization (free-text additions). Linked to session and included in locked record. |
| **Follow-up scheduling** | Every competitor tracks next appointment linked to session | Low | Date/time, type (review, next session, emergency), linked to session. Simple reference field -- actual scheduling is out of scope. |
| **Role-based access control** | Two roles minimum (Doctor, Admin). Regulatory requirement for access control. | Medium | Doctor: full clinical CRUD, sign-off. Admin/Receptionist: patient management, session viewing (read-only clinical), no sign-off. Must restrict who can sign off records and who can view clinical details. |

## Differentiators

Features that set the product apart from generic aesthetic EMR platforms. Not expected by every user, but valued by practices that care about compliance, efficiency, and audit readiness.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **UDI (Unique Device Identifier) capture** | Full FDA/EU MDR traceability chain -- most aesthetic EMRs only track batch/lot, not UDI | Low | Store the device's UDI (DI + PI components) alongside batch/lot. Future-proofs for regulatory tightening. Simple string field on device registry and procedure module. |
| **Structured adverse event capture with regulatory flag** | Goes beyond "complications noted" to structured severity grading and regulatory reporting triggers | Medium | Type (categorized), severity (mild/moderate/severe/life-threatening), onset timing, description, treatment given, outcome, and a flag for whether it meets mandatory reporting thresholds (FDA MDR 30-day, EU 2-15 day). Does NOT submit reports -- just flags and captures data for manual submission. |
| **Photo documentation with session linkage** | Clinical photography is the gold standard but most small-clinic EMRs handle it poorly | Medium | Before/after photos linked to specific session. Metadata: timestamp, anatomical region, photo type (before/label/after). Local filesystem storage for v1. Naming convention enforces organization. Not annotation/markup -- just capture and link. |
| **Per-module parameter validation** | Most EMRs accept any value; validating parameters against device-specific constraints prevents documentation errors | High | Enforce ranges: fluence within device spec, pulse duration within device capability, needle depth within handpiece limits. Requires device registry to carry parameter constraints. Significantly reduces documentation errors. |
| **Session flow enforcement (state machine)** | Prevents skipping required steps -- consent before procedure, screening before parameters | Medium | Server-enforced state machine prevents adding procedure modules without consent, prevents sign-off without outcomes, prevents locking without sign-off. Most competitors rely on UI warnings, not API enforcement. |
| **Seed data with controlled vocabularies** | Structured indication codes, clinical endpoints, and device catalogs reduce free-text variability | Low | Ship with curated seed data for: indication codes, clinical endpoint definitions, device catalog, product formulary. Makes data consistent and queryable. Admin management UI deferred. |
| **Record versioning with diff capability** | Most EHRs log access but don't preserve document versions for comparison | Medium | Store version snapshots at key transitions (draft saved, signed, addendum added). Enables "what changed" review for audits. Not real-time -- version captured at state transitions only. |

## Anti-Features

Features to explicitly NOT build. These are either out of scope, premature, or harmful to the project's focus.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **AI-powered documentation / speech-to-text** | Major undertaking, requires ML infrastructure, not v1 scope. Competitors like Clinicminds and Edvak have dedicated AI teams for this. | Provide well-structured templates and controlled vocabularies that make manual entry fast. AI can layer on top later. |
| **Injection site visual mapping (diagram annotation)** | Requires image manipulation, canvas rendering, coordinate storage -- purely frontend concern that doesn't belong in an API | Store injection site data as structured fields (anatomical zone, depth, units/volume per site). Frontend can render this onto face diagrams. API stores data, not images of diagrams. |
| **Photo annotation/markup tools** | Image editing is a frontend/client concern. API should store and serve photos, not manipulate them. | Store photo metadata and binary. Frontend handles annotation overlays. Keep API scope to upload, retrieve, link-to-session. |
| **Patient-facing portal** | Out of scope per PROJECT.md. Adds massive security surface. Different user type entirely. | Clinician/admin-only tool. Aftercare instructions can be printed or emailed manually. |
| **Billing/CPT coding** | Out of scope. Extremely complex domain (insurance, coding rules, claim submission). Entire industry of billing-specific software exists. | Record procedure details accurately so they CAN be used for billing by external systems later. Don't try to code or calculate fees. |
| **Appointment scheduling** | Separate concern per PROJECT.md. Many clinics already have scheduling software. | Follow-up field on session is just a date/time reference, not a calendar system. |
| **Multi-tenant / multi-site** | Single-tenant model per PROJECT.md. Multi-tenancy adds enormous complexity to data isolation, migrations, and compliance. | Each clinic gets its own deployed instance with its own database. Keep the deployment model simple. |
| **Real-time notifications / WebSocket** | Over-engineering for a clinical documentation tool used during in-person procedures | Standard request/response REST API. Polling for status changes is fine given the usage pattern. |
| **Integration with external EMR systems (HL7/FHIR)** | Massive scope. Standards are complex. Deferred to future versions. | Manual entry only for v1. API is well-structured so FHIR mapping can be added later. |
| **Complex reporting / analytics dashboards** | Analytics is a presentation concern. API should expose data, not generate reports. | Provide listing/filtering endpoints with date ranges, clinician filters, procedure type filters. Reports are a frontend or separate service concern. |
| **Automated regulatory submission** | Filing adverse event reports to FDA (eMDR) or EU authorities is a specialized workflow with strict format requirements | Capture all data needed for reports and flag events that meet thresholds. Actual submission is manual. |
| **Treatment plan / multi-session course management** | Tempting but adds significant complexity. Sessions are the atomic unit. | Each session stands alone. Frontend can display session history for a patient to show treatment progression. No formal "treatment plan" entity in v1. |

## Procedure Module Detail

Each procedure module type has specific parameter schemas. This is the most complex part of the feature set.

### Energy-Based Device Modules

**Common fields across all energy-based modules:**
- Device selection (from registry)
- Handpiece/tip selection
- Treatment area (anatomical zone)
- Number of passes
- Cooling method (contact/cryogen/air/none)
- Skin response observed (erythema, edema, etc.)
- Clinician notes (free text)

**IPL Module:**
- Wavelength/filter (nm)
- Fluence (J/cm2)
- Pulse duration (ms)
- Pulse mode (single/double/triple)
- Spot size (mm)
- Repetition rate (Hz)

**Nd:YAG / Long-Pulsed Laser Module:**
- Wavelength (typically 1064nm)
- Fluence (J/cm2)
- Pulse duration (ms)
- Spot size (mm)
- Repetition rate (Hz)
- Target chromophore

**CO2 / Ablative Laser Module:**
- Mode (continuous wave / single pulse / ultra pulse / fractional)
- Power (W)
- Fluence (J/cm2) -- for fractional mode
- Dwell time (ms)
- Pitch/density (for fractional scanner)
- Spot size (mm)
- Stacking (number of passes at same spot)

**RF / RF Microneedling Module:**
- RF type (monopolar/bipolar)
- Needle depth (mm)
- Energy per needle/pulse (mJ)
- Needle cartridge type (pin count)
- Mode (fixed/cycle/burst)
- Number of passes
- Direction pattern (horizontal/vertical/crosshatch)
- Temperature target (if applicable)

### Injectable Modules

**Filler Module:**
- Product selection (from formulary)
- Batch/lot number
- Expiry date
- Injection sites (list of: anatomical zone, depth, technique, volume per site)
- Total volume (mL) -- calculated from sites
- Needle/cannula type and gauge
- Anesthesia used

**Botulinum Toxin Module:**
- Product selection (from formulary)
- Batch/lot number
- Expiry date
- Reconstitution details (diluent volume, resulting concentration, time of reconstitution)
- Injection sites (list of: muscle name, units per site)
- Total units -- calculated from sites
- Needle gauge
- Technique notes

## Feature Dependencies

```
Role-Based Access Control
  |
  v
Patient Management --> Treatment Session Lifecycle
                         |
                         +--> Session Header (patient, clinician, timing, skin type)
                         |      |
                         |      v
                         +--> Consent Capture ----+
                         |                         |
                         +--> Contraindication     |
                         |    Screening -----------+
                         |                         |
                         |                         v
                         +--> Procedure Modules <--+-- Device/Product Registry
                         |    (all 6 types)            (seed data)
                         |      |
                         |      +--> Batch/Lot Tracking
                         |      |
                         |      +--> Photo Documentation
                         |
                         +--> Outcome Recording
                         |
                         +--> Adverse Event Capture
                         |
                         +--> Aftercare Documentation
                         |
                         +--> Follow-up Scheduling
                         |
                         v
                    Record Sign-off (validation gate)
                         |
                         v
                    Record Locking (immutability)
                         |
                         v
                    Addendum-only Amendments
                         |
                    Audit Trail (cross-cutting, applies to all above)
```

Key dependency chains:

1. **RBAC must come first** -- every other feature needs to know who the user is and what they can do
2. **Patient management before sessions** -- sessions require a patient reference
3. **Device/product registry before procedure modules** -- modules reference devices and products from the registry
4. **Consent + screening before procedure modules** -- session flow enforcement blocks procedure creation without consent
5. **All session content before sign-off** -- validation gate checks completeness
6. **Sign-off before locking** -- locking is the consequence of sign-off
7. **Locking before addendums** -- addendums only exist on locked records
8. **Audit trail is cross-cutting** -- implemented early, applied to all operations

## MVP Recommendation

Prioritize in this order:

1. **RBAC (Doctor + Admin roles)** -- foundation for all authorization decisions. Without this, nothing else can be properly secured.
2. **Patient management** -- simple CRUD. Required before any session can be created. Low complexity, high dependency.
3. **Device and product registry with seed data** -- must exist before procedure modules can reference them. Ship with hardcoded seed data. Low complexity.
4. **Treatment session lifecycle with session header** -- the core entity. Create, populate header fields, manage state transitions. High complexity but everything builds on this.
5. **Consent capture + contraindication screening** -- blocking gates in the session flow. Must work before procedure modules can be added.
6. **One energy-based module (IPL) + one injectable module (filler)** -- prove the module architecture works with two representative types before building all six.
7. **Outcome recording + aftercare** -- low complexity, completes the session flow from start to finish.
8. **Record sign-off + locking + addendums** -- the medico-legal capstone. Must work correctly to produce legally valid records.
9. **Audit trail** -- cross-cutting concern. Can be added incrementally but should be in place before sign-off/locking goes live.
10. **Remaining procedure modules** (Nd:YAG, CO2, RF, botulinum) -- follow the pattern established by IPL and filler modules.
11. **Adverse event capture** -- structured but not blocking for basic session flow.
12. **Photo documentation** -- valuable but not blocking. Local filesystem storage. Can be added after core flow works.

**Defer to post-MVP:**
- Follow-up scheduling: simple reference field, not critical path
- Per-module parameter validation against device specs: high value but high complexity, can be added after modules work with basic validation
- Record versioning with diffs: nice for audit but not required for medico-legal compliance (audit trail covers the requirement)
- UDI capture: simple field addition, can be added to device registry any time

## Sources

- [Pabau - 7 Best Aesthetic Clinic Software Solutions (2026)](https://pabau.com/blog/best-aesthetic-clinic-software/)
- [Pabau - 7 Best EMR Software for Medical & Aesthetic Practices in 2026](https://pabau.com/blog/best-emr-software/)
- [Edvak - Best Dermatology EHR for US Clinics in 2026](https://edvak.com/blogs/best-dermatology-ehr-us/)
- [LegendEHR - Cosmetic Dermatology EMR: Botox & Aesthetics in 2026](https://legendehr.com/cosmetic-dermatology-emr-botox-aesthetics-2026/)
- [Clinicminds - Aesthetic Clinic Software](https://www.clinicminds.com/)
- [Aesthetic Record EMR](https://www.aestheticrecord.com/)
- [Calysta EMR - Best Aesthetic EMR for Injector](https://calystaemr.com/best-aesthetic-emr-for-injector/)
- [PMC - Standard operating protocol for utilizing energy-based devices in aesthetic practice](https://pmc.ncbi.nlm.nih.gov/articles/PMC11626368/)
- [PMC - Updated Standards of Photographic Documentation in Aesthetic Medicine](https://pmc.ncbi.nlm.nih.gov/articles/PMC5585426/)
- [NCBI - Laser Fitzpatrick Skin Type Recommendations](https://www.ncbi.nlm.nih.gov/books/NBK557626/)
- [FDA - Medical Device Reporting (MDR)](https://www.fda.gov/medical-devices/medical-device-safety/medical-device-reporting-mdr-how-report-medical-device-problems)
- [FDA - UDI Basics](https://www.fda.gov/medical-devices/unique-device-identification-system-udi-system/udi-basics)
- [Noridian - Documentation Guidelines for Amended Records](https://med.noridianmedicare.com/web/jeb/cert-reviews/mr/documentation-guidelines-for-amended-records)
- [AccountableHQ - EHR Audit Trail Explained](https://www.accountablehq.com/post/ehr-audit-trail-explained-what-it-is-compliance-requirements-and-best-practices)
- [Jackson LLP - Digital Signatures in Healthcare](https://jacksonllp.com/digital-signatures/)
- [HubiFi - Immutable Audit Trails: A Complete Guide](https://www.hubifi.com/blog/immutable-audit-log-basics)
- [John Hoopman - The 5 Key Laser Parameters Every Aesthetic Professional Must Master](https://www.johnhoopman.com/laser-parameters-training/)
- [Pabau - Botox Face Mapping Guide + Template](https://pabau.com/blog/botox-face-mapping/)
- [Pabau - Filler Face Mapping Guide and Template](https://pabau.com/blog/filler-face-mapping/)
- [RxPhoto - Simplify Botox Charting](https://rxphoto.com/resources/blog/botox-charting-rxphoto/)
- [Getsolum - Role-Based Access Control in Healthcare](https://getsolum.com/glossary/role-based-access-control-healthcare)
