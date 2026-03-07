# Technology Stack

**Project:** Dermify API - Clinical Procedure Documentation
**Researched:** 2026-03-07
**Scope:** Additions to existing Go REST API stack for procedure documentation, audit trails, file uploads, and data validation

## Existing Stack (Do Not Replace)

These are already in the codebase and should remain:

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.23 | Language runtime |
| go-chi/chi | v5.0.10 | HTTP router |
| spf13/cobra + viper | v1.7.0 / v1.16.0 | CLI and configuration |
| jackc/pgx | v4.18.3 | PostgreSQL driver (via database/sql) |
| pressly/goose | v3.26.0 | Database migrations (embedded SQL) |
| golang-jwt/jwt | v4.5.2 | JWT authentication |
| google/uuid | v1.6.0 | UUID generation |
| prometheus/client_golang | v1.15.1 | Metrics |
| stretchr/testify | v1.11.0 | Testing assertions |
| swaggo/swag + http-swagger | v1.16.6 / v2.0.2 | API documentation |
| golang.org/x/crypto | v0.40.0 | Password hashing |

## Recommended New Stack Additions

### Data Validation: go-playground/validator v10

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/go-playground/validator/v10 | v10.30.1 | Struct-level request validation | HIGH |

**Why go-playground/validator over ozzo-validation:**
- Struct-tag based validation integrates with the existing swaggo annotation pattern (both use struct tags). The codebase already uses struct tags for JSON and swagger example annotations, so validator tags fit naturally.
- `required_if`, `required_unless`, `required_with` tags map directly to the conditional validation needs of medical forms (e.g., "reconstitution fields required only if procedure type is botulinum toxin").
- Cross-field validation (`gtfield`, `ltefield`, `eqfield`) handles dependent field rules like "session end time must be after start time."
- Most popular Go validation library by GitHub stars and usage. Ecosystem-standard choice.
- The `WithRequiredStructEnabled` option should be used during initialization (future default for v11).

**Why NOT ozzo-validation:**
- ozzo-validation v4 (v4.3.0) uses programmatic rule definitions, which would require building a separate validation layer rather than annotating the existing request structs. This adds boilerplate for the 50+ request struct types this project will need.
- Its code-based approach is better for dynamic validation logic, but the Dermify domain has mostly static rules defined in a spreadsheet -- struct tags are more maintainable for this use case.

**Custom validators needed (build on top of validator):**
- Skin type enum validation (Fitzpatrick I-VI)
- Device parameter range validation (per procedure module type)
- Batch/lot number format validation
- UDI (Unique Device Identifier) format validation

### File Upload: stdlib mime/multipart + gabriel-vasile/mimetype

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| net/http (stdlib) | - | Multipart form parsing | HIGH |
| github.com/gabriel-vasile/mimetype | v1.4.9 | Magic-number MIME type detection | HIGH |

**Why stdlib for file upload handling:**
- Go's `r.FormFile()` and `r.ParseMultipartForm()` are sufficient for the photo upload use case. The project stores photos on local filesystem -- no cloud streaming or resumable uploads needed.
- Chi router uses standard `http.Handler` interface, so stdlib multipart handling works directly.
- No dedicated upload middleware library needed -- the scope is "before/after photos" with a defined maximum size, not a general-purpose file service.

**Why gabriel-vasile/mimetype for type detection:**
- stdlib `http.DetectContentType` only recognizes ~30 MIME types and is too limited for reliable image type validation.
- mimetype checks actual file bytes (magic numbers) rather than trusting the Content-Type header, which is critical for security -- prevents upload of disguised executable files.
- Deterministic detection: same input always produces same result. Important for medical records where you need to verify file type at upload time.

**Why NOT gulter or other upload middleware:**
- Gulter adds abstraction over what is a 20-line stdlib pattern. Unnecessary dependency for local filesystem storage.

### Image Processing: disintegration/imaging

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/disintegration/imaging | v1.6.2 | Thumbnail generation for photos | MEDIUM |

**Why this library:**
- Pure Go, no CGO dependencies -- important for the Alpine Docker image build.
- Provides `Thumbnail()` and `Fit()` functions that handle aspect-ratio-aware resizing in a single call.
- Concurrent processing via goroutines for batch thumbnail generation.
- Supports JPEG, PNG, BMP, TIFF, GIF -- all likely clinical photo formats.

**Why MEDIUM confidence:**
- Last release was November 2021 (v1.6.2). Library is functionally complete but not actively developed.
- For v1 with local filesystem photos this is fine. If the project later moves to cloud storage, thumbnail generation might shift to a cloud function.

**Why NOT other options:**
- `nfnt/resize`: Archived/unmaintained, imaging is its spiritual successor.
- Standard library `image` package: Requires manual resize implementation, no thumbnail convenience functions.

### RBAC: Custom middleware (NOT Casbin)

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| Custom Chi middleware | - | Role-based access control | HIGH |

**Why custom over Casbin:**
- Dermify has exactly TWO roles: Doctor and Admin. The permission model is simple: Doctors can do everything clinical, Admins can manage patients/devices but not sign off records.
- Casbin (v2) is designed for complex multi-tenant RBAC/ABAC with policy files, model definitions, and enforcement engines. It is dramatically over-engineered for a two-role system on a single-tenant deployment.
- The existing auth middleware already extracts JWT claims with user ID. Adding a `role` claim and a `RequireRole(roles ...string)` middleware function is approximately 30 lines of code.
- No external dependency, no policy file management, no learning curve for future contributors.

**Why NOT Casbin:**
- Introduces model files (rbac_model.conf), policy storage, adapter configuration, and a concepts that make sense at 10+ roles or multi-tenant authorization. This project has 2 roles and single-tenant deployment.
- If role complexity grows in the future (e.g., Nurse role, per-clinic permissions), Casbin can be introduced then. Until that point, it adds complexity with no benefit.

### Audit Trail: PostgreSQL triggers + application-level audit table

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| PostgreSQL triggers (PL/pgSQL) | - | Immutability enforcement, automatic history capture | HIGH |
| crypto/sha256 (stdlib) | - | Hash chain for tamper detection | MEDIUM |

**Why database-level triggers for immutability:**
- PostgreSQL BEFORE UPDATE/DELETE triggers can prevent modifications to locked records at the database level, not just the application level. This is the correct enforcement point for medico-legal immutability -- even direct database access cannot modify a locked record.
- Trigger-based audit capture (recording old/new values on every change) runs inside the same transaction as the data change, ensuring audit completeness even if the application crashes mid-request.
- No external dependency. Raw SQL in Goose migration files, consistent with existing patterns.

**Why application-level audit metadata:**
- Session records should carry `created_by`, `created_at`, `updated_by`, `updated_at`, `signed_off_by`, `signed_off_at` columns directly on the record.
- A separate `audit_log` table captures the full change history as JSONB rows (who changed what, when, old/new values).
- This dual approach gives fast read access to "who last touched this" (columns on main table) and full history for compliance queries (audit_log table).

**Why hash chaining (MEDIUM confidence):**
- Storing SHA256(previous_hash + current_record_data) on each audit entry creates a tamper-evident chain. If any row is modified, the chain breaks.
- This is a "nice to have" for compliance demonstrations but adds complexity. The project can start without it and add hash chaining in a later phase if medico-legal requirements demand it.

**Why NOT event sourcing:**
- Full event sourcing (rebuilding state from events) is overkill for a clinical documentation app. The domain is CRUD with immutability-after-signoff, not a system where replaying events provides value.
- Event sourcing would require a completely different data access pattern incompatible with the existing direct-SQL handler approach.

**Why NOT temporal_tables extension:**
- Requires PostgreSQL extension installation, which adds deployment complexity.
- The project's append-only requirement is narrower than full temporal tables: only signed-off records become immutable, drafts can be freely edited. Custom triggers give precise control over this.

### Record Versioning: Addendum pattern in SQL

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| PostgreSQL (addendums table) | - | Post-signoff amendments to locked records | HIGH |

**Why a separate addendums table:**
- Once a session record is signed off, the original record is immutable (trigger-enforced). Addendums are new rows in a related table referencing the original session.
- Each addendum has its own `created_by`, `created_at`, `content` (text), and optional category.
- This is the standard pattern in medical records systems (EMR/EHR): original record is locked, corrections/additions are addendums that reference the original.

**Why NOT version numbers on the main record:**
- Version numbering implies the record itself changes. In this domain, the signed record MUST NOT change. Addendums are separate documents attached to the record.

### Integration Testing: testcontainers-go

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/testcontainers/testcontainers-go | v0.40.0 | PostgreSQL test containers for integration tests | MEDIUM |
| github.com/testcontainers/testcontainers-go/modules/postgres | v0.40.0 | PostgreSQL-specific test module | MEDIUM |

**Why testcontainers:**
- The procedure documentation domain has complex data relationships (sessions, modules, devices, outcomes, audit trails) that are best tested against a real PostgreSQL instance.
- Goose migrations can run against the test container, ensuring migration correctness.
- Trigger-based immutability enforcement can only be tested with real PostgreSQL.

**Why MEDIUM confidence:**
- The existing codebase uses testify assertions with manual setup. Introducing testcontainers is a pattern change that requires Docker-in-CI support.
- May not be needed for Phase 1 (basic CRUD). Becomes essential when testing signoff/locking/audit behavior.

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Validation | go-playground/validator v10 | ozzo-validation v4 | Code-based rules add boilerplate for 50+ static struct types; struct tags fit existing pattern |
| Validation | go-playground/validator v10 | Manual (current pattern) | Current `if field == ""` checks don't scale to 6 procedure module types with 20+ fields each |
| RBAC | Custom middleware | Casbin v2 | Only 2 roles in single-tenant app; Casbin is for complex multi-tenant authz |
| Upload | stdlib multipart | Gulter middleware | Unnecessary abstraction for local filesystem photo storage |
| Audit | PostgreSQL triggers | GORM callbacks | Project uses raw SQL, not GORM; triggers are database-level guarantees |
| Audit | PostgreSQL triggers | Event sourcing | Incompatible with existing CRUD data access; overkill for "immutable after signoff" |
| Image resize | disintegration/imaging | nfnt/resize | nfnt/resize is archived/unmaintained |
| Record versioning | Addendum table | Record version column | Medical record immutability requires original cannot change; versions imply mutation |
| Testing | testcontainers-go | Manual Docker setup | Testcontainers provide deterministic, isolated test databases per test suite |

## pgx v4 End-of-Life Warning

**CRITICAL:** jackc/pgx v4 reaches end-of-life on July 1, 2025. The project currently uses pgx v4.18.3. While this is not blocking for the procedure documentation milestone, upgrading to pgx v5 should be planned as a separate effort.

Key migration impact:
- Package reorganization: pgtype, pgconn, pgproto3 move into the pgx module
- Type system change: `Status` (Undefined/Present/Null) replaced with `Valid` boolean
- Logging replaced with tracing hooks
- Estimated effort: ~6 hours for a codebase this size (based on community migration reports)

**Recommendation:** Do NOT combine pgx upgrade with the procedure documentation work. Handle it as a separate prerequisite or parallel effort.

## JSONB Usage Pattern

The procedure modules (IPL, Nd:YAG, CO2, RF, Fillers, Botulinum) each have different parameter schemas. Two storage approaches are available:

**Recommended: Dedicated tables per module type** (not JSONB)
- Each module type gets its own table with typed columns matching the spreadsheet specification.
- PostgreSQL CHECK constraints enforce value ranges at the database level.
- Enables proper indexing, foreign keys, and query performance.
- More migration files, but each is straightforward.

**Why NOT JSONB for procedure parameters:**
- JSONB loses column-level NOT NULL enforcement, CHECK constraints, and type safety.
- The 6 module types have well-defined, stable schemas from the spreadsheet specification. This is not dynamic/user-defined data.
- JSONB makes sense for the audit log `changes` column (truly variable structure), not for the core domain data.

## Installation

```bash
# New core dependencies
go get github.com/go-playground/validator/v10@v10.30.1
go get github.com/gabriel-vasile/mimetype@v1.4.9

# Image processing (for photo thumbnails)
go get github.com/disintegration/imaging@v1.6.2

# Integration testing (dev dependency)
go get github.com/testcontainers/testcontainers-go@v0.40.0
go get github.com/testcontainers/testcontainers-go/modules/postgres@v0.40.0
```

## Enum/Controlled Value Pattern

For the many enumerated types in this domain (skin types, indication codes, device types, adverse event severities), use the string constant + validator pattern:

```go
// Define as string type with constants
type SkinType string

const (
    SkinTypeI   SkinType = "I"
    SkinTypeII  SkinType = "II"
    SkinTypeIII SkinType = "III"
    SkinTypeIV  SkinType = "IV"
    SkinTypeV   SkinType = "V"
    SkinTypeVI  SkinType = "VI"
)

// Register as custom validator
// validate:"skin_type" maps to a function checking against allowed values

// Enforce at DB level with CHECK constraint
// CHECK (skin_type IN ('I','II','III','IV','V','VI'))
```

This gives triple enforcement: Go type safety, validator tag checking, and database constraint.

## Sources

- [go-playground/validator v10 releases](https://github.com/go-playground/validator/releases) - v10.30.1 confirmed
- [go-playground/validator v10 docs](https://pkg.go.dev/github.com/go-playground/validator/v10) - conditional validation tags
- [ozzo-validation v4 releases](https://github.com/go-ozzo/ozzo-validation/releases) - v4.3.0 confirmed
- [gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) - v1.4.9, magic-number MIME detection
- [disintegration/imaging](https://github.com/disintegration/imaging) - v1.6.2, pure Go image processing
- [jackc/pgx v5](https://pkg.go.dev/github.com/jackc/pgx/v5) - v5.8.0, v4 EOL July 2025
- [jackc/pgx v4 EOL](https://github.com/jackc/pgx/blob/master/CHANGELOG.md) - end-of-life announced
- [Audit log patterns in Go/PostgreSQL](https://dev.to/akkaraponph/comprehensive-research-audit-log-paradigms-gopostgresqlgorm-design-patterns-1jmm)
- [PostgreSQL trigger documentation](https://www.postgresql.org/docs/current/plpgsql-trigger.html)
- [Tamper-evident audit trails in PostgreSQL](https://appmaster.io/blog/tamper-evident-audit-trails-postgresql)
- [Casbin RBAC library](https://github.com/casbin/casbin)
- [testcontainers-go PostgreSQL module](https://golang.testcontainers.org/modules/postgres/)
- [testcontainers-go releases](https://github.com/testcontainers/testcontainers-go/releases) - v0.40.0
- [Three Dots Labs: Safer Enums in Go](https://threedots.tech/post/safer-enums-in-go/)
- [Go multipart file upload patterns](https://www.mohitkhare.com/blog/file-upload-golang/)
- [Go MIME type detection](https://rnemeth90.github.io/posts/2024-03-27-golang-detect-file-type/)
- [PostgreSQL temporal tables](https://wiki.postgresql.org/wiki/SQL2011Temporal)
