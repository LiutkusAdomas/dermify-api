# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 -- Clinical Procedure Documentation API

**Shipped:** 2026-03-08
**Phases:** 7 | **Plans:** 23

### What Was Built
- Complete clinical procedure documentation API with 63 requirements across RBAC, patients, sessions, 6 module types, sign-off, and photos
- Service/repository layered architecture with interface-based DI across all domains
- Database-enforced immutability and audit trail via PL/pgSQL triggers on all 13 clinical tables
- Hybrid polymorphism for procedure modules (shared session_modules + per-type detail tables)

### What Worked
- Phase-by-phase delivery with each phase building cleanly on the previous -- no rework needed
- 3-layer plan structure (domain+migrations -> repositories+tests -> handlers+routes) applied consistently across all phases
- Service/repository pattern established in Phase 1 scaled perfectly through 6 more phases
- Mock repositories with function fields enabled fast, focused unit testing
- Database triggers for immutability and audit provided defense-in-depth without coupling to application code

### What Was Inefficient
- SUMMARY.md frontmatter (requirements_completed) never populated -- metadata gap carried across all 23 plans
- Module/outcome update methods lack application-level editability checks -- identified late, compensated by DB triggers
- ROADMAP.md plan checkboxes not consistently updated during execution (some phases show unchecked despite completion)

### Patterns Established
- Handler naming: `Handle{Action}()` (e.g., HandleCreateSession, HandleSignOff)
- Error handling: shared `handle{Domain}Error()` in dedicated files for multi-handler domains, local for single-handler domains
- Test helpers: `{domain}TestDeps` struct with setup methods (e.g., energyTestDeps.setupEditableSession)
- Mock pattern: function-field mocks (not testify/mock) for repository test doubles
- Route structure: nested `chi.Route` for sub-resources under `/{id}`
- Migration naming: `YYYYMMDDHHMMSS_description.sql` with both Up and Down sections
- Domain separation: one handler file, one route file, one mock file per domain

### Key Lessons
1. Establishing architecture patterns early (Phase 1) pays compound returns -- every subsequent phase follows the same recipe with minimal decisions
2. Database-level enforcement (triggers) is worth the upfront investment -- application bugs cannot bypass immutability or audit
3. Hybrid polymorphism (shared table + detail tables) handles procedure module diversity well without excessive abstraction
4. Consent gates and editability checks at the service layer prevent invalid operations cleanly across all handler entry points

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v1.0 | 7 | 23 | Established layered architecture and consistent 3-plan phase structure |

### Cumulative Quality

| Milestone | Tests | LOC | Migrations |
|-----------|-------|-----|------------|
| v1.0 | 100+ | 17,400 Go | 19 SQL |

### Top Lessons (Verified Across Milestones)

1. Architecture patterns set in first phase compound through all subsequent phases
2. Database-level enforcement (triggers) provides defense-in-depth that application code alone cannot guarantee
