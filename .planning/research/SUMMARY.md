# Research Summary: Dermify Procedure Documentation

**Domain:** Clinical procedure documentation API for aesthetic dermatology
**Researched:** 2026-03-07
**Overall confidence:** HIGH

## Executive Summary

Building a clinical procedure documentation system on top of the existing Dermify Go REST API requires surprisingly few new external dependencies. The existing stack (Chi router, pgx/PostgreSQL, Goose migrations, JWT auth) provides a solid foundation. The primary additions are: go-playground/validator v10 for struct-level validation of the 50+ request types needed, gabriel-vasile/mimetype for secure photo upload validation, and disintegration/imaging for thumbnail generation. The most critical architectural change is not a library addition but a structural one: introducing service and repository layers to handle the complex business logic (state machines, sign-off validation gates, immutability enforcement) that cannot live in handlers.

The domain's defining characteristic is medico-legal immutability. Once a clinician signs off on a treatment session, the record must be physically unmodifiable at the database level -- not just application-level access control. This requirement drives the key technical decisions: PostgreSQL triggers for immutability enforcement, append-only audit trail tables, and a separate addendums table for post-signoff amendments. These are standard patterns in healthcare IT, well-supported by PostgreSQL's trigger system, and require no external libraries.

The most complex feature set is the 6 procedure module types (IPL, Nd:YAG, CO2, RF, Fillers, Botulinum), each with 15-25 device-specific parameter fields. The recommended approach is a hybrid polymorphism pattern: a shared `session_modules` table for common fields plus separate detail tables per module type. This preserves PostgreSQL's type safety, NOT NULL enforcement, and query planner statistics -- advantages lost if JSONB is used for the varied parameter schemas.

One infrastructure concern surfaces prominently: jackc/pgx v4 reached end-of-life in July 2025. The project currently depends on v4.18.3. While this does not block procedure documentation work, it creates technical debt that grows with every new database query site. The migration to pgx v5 should be handled as a separate, focused effort -- not interleaved with feature development.

## Key Findings

**Stack:** Only 3 new external dependencies needed -- go-playground/validator v10.30.1, gabriel-vasile/mimetype v1.4.9, and disintegration/imaging v1.6.2. Everything else is stdlib or PostgreSQL features.

**Architecture:** Introduce service and repository layers. Use hybrid polymorphism (shared header table + per-type detail tables) for procedure modules. Session lifecycle as a state machine enforced at both application and database levels.

**Critical pitfall:** Application-only immutability enforcement is the highest-risk mistake. Database triggers MUST prevent modification of locked records. A bug in Go code should not be able to alter a signed medical record.

## Implications for Roadmap

Based on research, suggested phase structure:

1. **Foundation: Service/Repository Layer + RBAC + Patients** - Establish the architectural pattern before building complex domains
   - Addresses: RBAC (FEATURES.md table stakes), Patient CRUD (dependency for sessions), service layer pattern
   - Avoids: Fat handler anti-pattern (PITFALLS.md #3), scattered validation (PITFALLS.md #8)

2. **Core Session Flow: Lifecycle + First Two Modules + Audit** - Build the product's core value with validation of the polymorphic module architecture
   - Addresses: Session lifecycle, session header, consent capture, IPL module, Filler module, audit trail
   - Avoids: Single table for all module types (PITFALLS.md #4), untested state machine (PITFALLS.md #9)

3. **Clinical Completeness: Remaining Modules + Safety Features** - Expand to all 6 module types and add clinical workflow gates
   - Addresses: Nd:YAG, CO2, RF, Botulinum modules, contraindication screening, adverse events, outcomes, aftercare
   - Avoids: Copy-paste inconsistencies across modules (PITFALLS.md phase warning)

4. **Immutability and Compliance: Sign-off + Locking + Addendums** - The medico-legal capstone
   - Addresses: Validation gate, sign-off, record locking (trigger-enforced), addendum system, full audit history
   - Avoids: Application-only immutability (PITFALLS.md #1), missing audit entries (PITFALLS.md #5)

5. **Polish: Photos + Follow-ups + Integration Tests** - Non-blocking enhancements and end-to-end validation
   - Addresses: Photo upload/storage, thumbnails, follow-up scheduling, integration test suite
   - Avoids: Storing photos in database (PITFALLS.md #6), trusting Content-Type header (PITFALLS.md #7)

**Phase ordering rationale:**
- Foundation must come first because every subsequent phase depends on the service/repository pattern and RBAC being in place.
- Session lifecycle before modules because modules belong to sessions.
- Two modules before six modules to validate the polymorphic architecture with minimal investment.
- Sign-off and locking after all modules because the validation gate must check all module types.
- Photos last because they are independent of the core clinical flow and can be added without affecting existing endpoints.

**Research flags for phases:**
- Phase 2: The state machine implementation needs careful design -- enumeration of all valid/invalid transitions. Likely needs targeted research on PostgreSQL state machine enforcement patterns during planning.
- Phase 3: Botulinum module has unique complexity (reconstitution tracking, time-sensitive discard calculations). May need domain-specific research during planning.
- Phase 4: Trigger-based immutability across multiple tables (sessions, session_modules, 6 detail tables, consent, outcomes) is intricate. Integration test strategy for triggers needs planning.
- Phase 5: Photo storage directory structure and naming convention should be decided before implementation. Migration path to S3 should be considered in the interface design.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All library versions verified via official sources. go-playground/validator and mimetype are ecosystem standard choices. |
| Features | HIGH | Feature landscape derived directly from PROJECT.md requirements with domain knowledge of aesthetic dermatology documentation standards. |
| Architecture | HIGH | Hybrid polymorphism, state machines, and service/repository layers are well-established patterns with extensive Go community documentation. |
| Pitfalls | HIGH | Pitfalls sourced from medical records compliance standards, Go community best practices, and PostgreSQL documentation. pgx v4 EOL confirmed via official changelog. |
| Image processing | MEDIUM | disintegration/imaging is functionally complete but last released November 2021. Acceptable for v1 but may need re-evaluation if requirements grow. |
| Integration testing | MEDIUM | testcontainers-go recommendation is sound but represents a pattern change from existing test approach. Docker-in-CI requirements need validation. |

## Gaps to Address

- **pgx v4 to v5 migration timing:** Should this be a prerequisite phase (Phase 0) or a parallel effort? The current codebase is small enough to migrate now (~6 hours). Delaying makes it harder.
- **Exact field specifications per module type:** The PROJECT.md references "12 spreadsheet tables" with detailed field specs. The research covers the general patterns but the exact column definitions for each module type need the spreadsheet data during implementation planning.
- **Photo storage directory structure:** The research recommends local filesystem with `{base_path}/{session_id}/{uuid}.{ext}` but the exact configuration (base path from config, directory creation strategy, cleanup policy) needs to be decided during Phase 5 planning.
- **Seed data content:** The device/product registry needs actual seed data (real device models, product names, manufacturers, indication codes). This is domain data, not a technology decision -- needs clinical input.
- **golangci-lint compatibility:** The project has ~60 linters enabled. New dependencies (validator, mimetype, imaging) need to pass the strict lint config. No known incompatibilities but should be verified early.

## Files Created

| File | Key Content |
|------|-------------|
| .planning/research/SUMMARY.md | This file -- executive summary with roadmap implications |
| .planning/research/STACK.md | 3 new dependencies + patterns for audit/RBAC/versioning |
| .planning/research/FEATURES.md | 12 table stakes, 10 differentiators, 10 anti-features, dependency graph |
| .planning/research/ARCHITECTURE.md | Service/repository layers, hybrid polymorphism, state machine, audit trail patterns |
| .planning/research/PITFALLS.md | 14 pitfalls (4 critical, 6 moderate, 4 minor) with phase-specific warnings |
