---
phase: 06-photo-documentation
plan: 03
subsystem: api
tags: [photo-upload, filestore, http-handlers, prometheus, chi-router]

# Dependency graph
requires:
  - phase: 06-photo-documentation plan 01
    provides: domain types, service layer, FileStore interface, photo repository contract
  - phase: 06-photo-documentation plan 02
    provides: PostgresPhotoRepository, StorageConfig, photo service tests
provides:
  - LocalFileStore implementing FileStore for local filesystem storage
  - 6 HTTP handlers for photo upload/list/get/serve/delete
  - Photo error codes and error-to-HTTP mapping
  - Prometheus photo_uploaded counter
  - Complete route wiring under /api/v1/sessions/{id}/photos
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "LocalFileStore with directory creation on Save and cleanup on error"
    - "Content-type detection via http.DetectContentType on first 512 bytes"
    - "Server-controlled file paths from DB record to prevent path traversal"

key-files:
  created:
    - internal/service/filestore.go
    - internal/api/handlers/photo.go
    - internal/api/handlers/photo_errors.go
  modified:
    - internal/api/apierrors/apierrors.go
    - internal/api/metrics/prometheus.go
    - internal/api/metrics/metrics.go
    - internal/api/routes/sessions.go
    - internal/api/routes/manager.go

key-decisions:
  - "parsePhotoIDParam and parseModuleIDParam as local helpers in photo.go (existing modules inline the parsing)"
  - "HandleServePhotoFile uses server-controlled DB path preventing path traversal attacks"
  - "Label photo route uses POST /photos/label/{moduleId} for explicit module binding"

patterns-established:
  - "LocalFileStore pattern: basePath + filepath.FromSlash(relPath) for cross-platform path handling"
  - "Photo handler content detection: read 512 bytes, detect type, seek back to start"

requirements-completed: [PHOTO-01, PHOTO-02, PHOTO-03, PHOTO-04]

# Metrics
duration: 4min
completed: 2026-03-08
---

# Phase 6 Plan 03: HTTP Layer Summary

**LocalFileStore, 6 photo HTTP handlers, error mapping, Prometheus counter, and full route wiring under /sessions/{id}/photos**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-08T14:09:37Z
- **Completed:** 2026-03-08T14:13:16Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- LocalFileStore with Save (creates dirs, writes files, cleans up on error), Delete, and Exists
- 6 HTTP handlers: upload before photo, upload label photo, list session photos, get photo metadata, serve photo file, delete photo
- Photo error codes (7 constants) and handlePhotoError mapping all 6 service errors to HTTP statuses
- Prometheus counter dermify_photo_uploaded_total registered and incremented on uploads
- Routes wired under /api/v1/sessions/{id}/photos with proper nesting and auth/role middleware

## Task Commits

Each task was committed atomically:

1. **Task 1: LocalFileStore, HTTP handlers, error codes, and metrics** - `8ed3475` (feat)
2. **Task 2: Route wiring and dependency injection** - `c161952` (feat)

## Files Created/Modified
- `internal/service/filestore.go` - LocalFileStore implementing FileStore interface with Save/Delete/Exists
- `internal/api/handlers/photo.go` - 6 HTTP handlers for photo operations plus parsePhotoIDParam/parseModuleIDParam helpers
- `internal/api/handlers/photo_errors.go` - handlePhotoError mapping service errors to HTTP responses
- `internal/api/apierrors/apierrors.go` - Added 7 photo error code constants
- `internal/api/metrics/prometheus.go` - Added photoUploadedCounterMetric and IncrementPhotoUploadedCount method
- `internal/api/metrics/metrics.go` - Added newPhotoUploadedCounter constructor
- `internal/api/routes/sessions.go` - Added photoSvc/storagePath fields and photo route registration
- `internal/api/routes/manager.go` - Creates PhotoService with LocalFileStore and injects into SessionRoutes

## Decisions Made
- parsePhotoIDParam and parseModuleIDParam as local helpers in photo.go (existing modules inline the parsing)
- HandleServePhotoFile uses server-controlled DB path preventing path traversal attacks
- Label photo route uses POST /photos/label/{moduleId} for explicit module binding

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 6 Photo Documentation is now complete (all 3 plans executed)
- All PHOTO requirements (PHOTO-01 through PHOTO-04) satisfied
- Full photo lifecycle: upload, list, get metadata, serve file, delete
- Photo consent enforcement via service layer (403 when consent not set or declined)

## Self-Check: PASSED

All 8 files verified present. Commits `8ed3475` and `c161952` verified in git log.

---
*Phase: 06-photo-documentation*
*Completed: 2026-03-08*
