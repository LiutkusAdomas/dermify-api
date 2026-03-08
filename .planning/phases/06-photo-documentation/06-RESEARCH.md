# Phase 6: Photo Documentation - Research

**Researched:** 2026-03-08
**Domain:** Multipart file upload, filesystem storage, consent-gated photo management in Go
**Confidence:** HIGH

## Summary

Phase 6 adds photo documentation to the dermify-api. Clinicians upload "before" photos linked to sessions and product label photos linked to procedure modules. Photos are stored on the local filesystem with organized, predictable naming -- not in the database. The system enforces a consent gate: uploads are blocked unless the session's `photo_consent` field is set to "yes" or "limited".

The existing codebase already has the `photo_consent` field on the Session domain (`PhotoConsent *string` with constants `yes`, `no`, `limited`) and validation in `session.go`. This phase introduces a new `Photo` domain, a `PhotoService` with a `FileStore` interface for filesystem operations (enabling testability), a `photos` database table for metadata, corresponding repository/handler layers, and config for the storage base path. All of this follows the established service/repository/handler architecture used throughout Phases 1-5.

**Primary recommendation:** Use Go's stdlib `multipart`, `io`, `os`, and `crypto/rand` for file handling. No external libraries needed. Store metadata (path, MIME type, session/module links) in PostgreSQL; store binary files on the filesystem. Abstract filesystem operations behind a `FileStore` interface so service tests use an in-memory mock.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PHOTO-01 | Clinician can upload before photos linked to a session | Multipart upload handler, Photo domain with session_id FK, PhotoService with consent gate, filesystem storage |
| PHOTO-02 | Clinician can upload product label photos linked to a procedure module | Same upload mechanism with module_id FK, label photo type, module existence validation |
| PHOTO-03 | Photos are stored on the local filesystem with organized naming | FileStore interface, directory structure `{base}/{session_id}/{type}/{uuid}.{ext}`, config for base path |
| PHOTO-04 | Photos require consent flag to be set before upload | PhotoService checks session.PhotoConsent is "yes" or "limited" before accepting upload |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `net/http` | stdlib | Multipart form parsing, content type detection | Go stdlib handles multipart/form-data natively via `r.FormFile()` and `http.DetectContentType()` |
| `io` | stdlib | Streaming file copy from upload to disk | `io.Copy()` streams without loading entire file into memory |
| `os` | stdlib | File creation, directory creation | `os.MkdirAll()` for nested directories, `os.Create()` for file writes |
| `path/filepath` | stdlib | Cross-platform path construction | `filepath.Join()` ensures correct separators on Windows/Linux |
| `crypto/rand` | stdlib | Unique filename generation | Cryptographically secure random bytes for unique photo filenames |
| `encoding/hex` | stdlib | Convert random bytes to filename string | Hex encoding of 16 random bytes produces 32-char unique ID |
| `mime` | stdlib | File extension from MIME type | `mime.ExtensionsByType()` maps detected content type to extension |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `jackc/pgx` v4 | existing | Photo metadata table operations | Already in project for all PostgreSQL access |
| `go-chi/chi` v5 | existing | Route registration for photo endpoints | Already the project router |
| `stretchr/testify` | existing | Test assertions | Already the project test framework |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `crypto/rand` + hex | `google/uuid` | UUID package adds dependency; hex-encoded random bytes are sufficient and match project's minimal-deps philosophy |
| Filesystem storage | S3/MinIO | Out of scope -- requirements specify local filesystem; FileStore interface allows future swap |
| `http.DetectContentType` | `gabriel-vasile/mimetype` | External lib detects more types but stdlib covers JPEG/PNG which is all we need for clinical photos |

**Installation:**
No new dependencies required. Everything uses Go stdlib plus existing project libraries.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  domain/
    photo.go              # Photo domain type, PhotoType constants
  service/
    photo.go              # PhotoService, PhotoRepository interface, FileStore interface
    photo_test.go         # Service unit tests with mock FileStore and mock repo
  repository/postgres/
    photo.go              # PostgreSQL repository implementation
    photo_test.go         # Repository tests
  api/handlers/
    photo.go              # Upload and retrieval handlers
    photo_errors.go       # Error mapping (follows energy_module_errors.go pattern)
  api/routes/
    sessions.go           # Extended with photo routes (nested under /{id}/photos)
  testutil/
    mock_photo.go         # Mock PhotoRepository
    mock_filestore.go     # Mock FileStore
config/
  config.go              # Extended with StorageConfig
migrations/
  20260308050000_create_photos_table.sql
```

### Pattern 1: FileStore Interface for Testability
**What:** Abstract filesystem operations behind an interface so the service layer never calls `os` directly.
**When to use:** Always -- this is critical for unit testing without touching the real filesystem.
**Example:**
```go
// Source: project convention -- interface-based DI like SessionRepository
type FileStore interface {
    Save(ctx context.Context, path string, reader io.Reader) (int64, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
}
```

### Pattern 2: Consent Gate Check in Service Layer
**What:** PhotoService fetches the session and checks `PhotoConsent` before accepting any upload.
**When to use:** Every photo upload operation (both before-photos and label-photos).
**Example:**
```go
// Source: follows existing consent gate pattern in session.go AddModule()
func (s *PhotoService) UploadPhoto(ctx context.Context, photo *domain.Photo, reader io.Reader) error {
    session, err := s.sessionRepo.GetByID(ctx, photo.SessionID)
    if err != nil {
        return err
    }
    if session.PhotoConsent == nil ||
       *session.PhotoConsent == domain.PhotoConsentNo {
        return ErrPhotoConsentRequired
    }
    // ... validate, generate path, save file, persist metadata
}
```

### Pattern 3: Organized Filesystem Naming
**What:** Predictable directory hierarchy based on session ID, photo type, and unique filename.
**When to use:** All photo storage.
**Example:**
```
{storage.base_path}/
  sessions/
    {session_id}/
      before/
        {hex_id}.jpg
        {hex_id}.png
      labels/
        {hex_id}.jpg
```
This structure makes photos browsable by session and type, supports multiple photos per session, and avoids filename collisions via random hex IDs.

### Pattern 4: Multipart Upload Handler
**What:** Handler parses multipart form, validates file type via magic bytes, delegates to service.
**When to use:** POST endpoints for photo upload.
**Example:**
```go
// Source: Go stdlib multipart + project handler conventions
func HandleUploadBeforePhoto(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Limit request body size
        r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSize)

        file, header, err := r.FormFile("photo")
        if err != nil {
            // handle error
            return
        }
        defer file.Close()

        // Read first 512 bytes for content type detection
        buf := make([]byte, 512)
        n, _ := file.Read(buf)
        contentType := http.DetectContentType(buf[:n])
        // Seek back to start
        file.Seek(0, io.SeekStart)

        // Validate allowed types
        if !isAllowedImageType(contentType) {
            // reject
            return
        }

        // Build domain photo, call service
    }
}
```

### Pattern 5: Config-Driven Storage Path
**What:** Storage base path comes from config, not hardcoded. Overridable via `OVERRIDE_STORAGE_BASE_PATH`.
**When to use:** Application startup and service instantiation.
**Example:**
```go
// In config/config.go
type StorageConfig struct {
    BasePath string `mapstructure:"base_path"`
}

type Configuration struct {
    // ... existing fields
    Storage StorageConfig `mapstructure:"storage"`
}
```

### Anti-Patterns to Avoid
- **Storing binary data in PostgreSQL:** The requirements explicitly state filesystem storage. BYTEA columns waste DB resources and complicate backups for large files.
- **Trusting Content-Type header:** Always detect MIME type from file magic bytes using `http.DetectContentType`, never trust the client-provided Content-Type.
- **Calling os directly in service layer:** Makes unit testing require real filesystem operations. Use the FileStore interface.
- **Sequential numeric filenames:** Race conditions and predictability. Use crypto/rand hex IDs.
- **No size limit on uploads:** Always wrap with `http.MaxBytesReader` to prevent memory exhaustion.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| MIME type detection | Custom magic byte parser | `http.DetectContentType()` | Stdlib implements full WHATWG MIME sniffing spec; covers JPEG, PNG, GIF, WebP |
| Unique filenames | Timestamp-based naming | `crypto/rand` + hex encoding | Cryptographic randomness eliminates collisions even under concurrent uploads |
| Directory creation | Manual parent-by-parent creation | `os.MkdirAll()` | Handles full path creation atomically with proper permissions |
| Multipart parsing | Custom body parser | `r.FormFile()` / `r.ParseMultipartForm()` | Stdlib handles RFC 2046 multipart correctly with memory limits |
| Request body limiting | Manual byte counting | `http.MaxBytesReader()` | Returns 413 automatically, prevents DoS |

**Key insight:** Go's stdlib handles every aspect of file upload, validation, and storage. No external dependencies are needed for this phase.

## Common Pitfalls

### Pitfall 1: Forgetting to Seek After Content Type Detection
**What goes wrong:** After reading 512 bytes for `DetectContentType`, the file reader position is at byte 512. If you then copy the file to disk, you lose the first 512 bytes.
**Why it happens:** `multipart.File` is an `io.Reader` that advances its position on read.
**How to avoid:** Call `file.Seek(0, io.SeekStart)` after content type detection to reset the reader. `multipart.File` implements `io.Seeker`.
**Warning signs:** Uploaded images appear corrupted or cannot be opened.

### Pitfall 2: Path Traversal in Photo Retrieval
**What goes wrong:** If the download/serve endpoint takes a filename from the URL and concatenates it with the base path, an attacker can use `../` sequences to read arbitrary files.
**Why it happens:** URL parameters are untrusted input.
**How to avoid:** Use `filepath.Base()` to strip directory components. Better: look up the photo by database ID and construct the path server-side from trusted metadata.
**Warning signs:** Any endpoint that uses user-supplied strings in file paths.

### Pitfall 3: Not Limiting Upload Size
**What goes wrong:** A malicious or misconfigured client sends a multi-GB file, exhausting server memory or disk.
**Why it happens:** `r.ParseMultipartForm()` defaults to 32MB in-memory but still writes to temp files on disk.
**How to avoid:** Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, maxSize)` BEFORE parsing. This hard-limits the entire request body.
**Warning signs:** OOM errors under load, disk space exhaustion.

### Pitfall 4: Orphaned Files on Database Error
**What goes wrong:** File is saved to disk but database metadata insert fails. Now there is a file with no record pointing to it.
**Why it happens:** File write and DB insert are not in a single transaction.
**How to avoid:** Write metadata to DB first (with a "pending" status or a generated path), then write the file. On file write failure, delete the DB record. Alternatively, write file first, then DB, and on DB failure, delete the file. The latter is simpler -- just add cleanup in the error path.
**Warning signs:** Growing disk usage without matching database records.

### Pitfall 5: Windows Path Separator Issues
**What goes wrong:** Stored file paths use backslashes on Windows, causing lookup failures when served on Linux or compared in tests.
**Why it happens:** `filepath.Join` uses OS-native separators.
**How to avoid:** Store paths in the database using forward slashes (POSIX style). Use `filepath.Join` only for actual filesystem operations. Convert with `filepath.ToSlash()` before storing.
**Warning signs:** Tests pass on one OS but fail on another.

### Pitfall 6: Photo Consent "limited" Interpretation
**What goes wrong:** Blocking uploads when consent is "limited" when it should be allowed.
**Why it happens:** Misunderstanding the three-value photo consent: "yes" = all photos allowed, "limited" = some photos allowed (clinician discretion), "no" = no photos.
**How to avoid:** Only block uploads when `PhotoConsent` is nil or "no". Both "yes" and "limited" permit uploads.
**Warning signs:** Clinicians unable to upload photos despite having recorded limited consent.

## Code Examples

Verified patterns from official sources and project conventions:

### File Upload Handler (Multipart Parsing)
```go
// Source: Go stdlib net/http, adapted to project handler pattern
const maxPhotoSize = 10 << 20 // 10 MB

func HandleUploadBeforePhoto(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        claims := middleware.GetUserClaims(r.Context())
        if claims == nil {
            apierrors.WriteError(w, http.StatusUnauthorized,
                apierrors.AuthNotAuthenticated, "not authenticated")
            return
        }

        sessionID, err := parseIDParam(r)
        if err != nil {
            apierrors.WriteError(w, http.StatusBadRequest,
                apierrors.SessionInvalidData, "invalid session ID")
            return
        }

        r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSize)

        file, header, err := r.FormFile("photo")
        if err != nil {
            apierrors.WriteError(w, http.StatusBadRequest,
                apierrors.PhotoInvalidData, "invalid file upload")
            return
        }
        defer file.Close()

        // Detect content type from magic bytes
        buf := make([]byte, 512)
        n, _ := file.Read(buf)
        contentType := http.DetectContentType(buf[:n])
        if _, err := file.Seek(0, io.SeekStart); err != nil {
            apierrors.WriteError(w, http.StatusInternalServerError,
                apierrors.PhotoUploadFailed, "failed to process file")
            return
        }

        photo := &domain.Photo{
            SessionID:       sessionID,
            PhotoType:       domain.PhotoTypeBefore,
            OriginalName:    header.Filename,
            ContentType:     contentType,
            SizeBytes:       header.Size,
            CreatedBy:       claims.UserID,
            UpdatedBy:       claims.UserID,
        }

        if err := svc.UploadPhoto(r.Context(), photo, file); err != nil {
            handlePhotoUploadError(w, err)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(toPhotoResponse(photo)) //nolint:errcheck
    }
}
```

### Photo Domain Type
```go
// Source: project domain conventions (matches consent.go, outcome.go pattern)
package domain

import "time"

const (
    PhotoTypeBefore = "before"
    PhotoTypeLabel  = "label"
)

type Photo struct {
    ID           int64     `json:"id"`
    SessionID    int64     `json:"session_id"`
    ModuleID     *int64    `json:"module_id,omitempty"` // set for label photos
    PhotoType    string    `json:"photo_type"`          // "before" or "label"
    FilePath     string    `json:"file_path"`           // relative path from storage root
    OriginalName string    `json:"original_name"`
    ContentType  string    `json:"content_type"`
    SizeBytes    int64     `json:"size_bytes"`
    Version      int       `json:"version"`
    CreatedAt    time.Time `json:"created_at"`
    CreatedBy    int64     `json:"created_by"`
    UpdatedAt    time.Time `json:"updated_at"`
    UpdatedBy    int64     `json:"updated_by"`
}
```

### FileStore Implementation
```go
// Source: Go stdlib os/io, abstracted for testability
package service

type LocalFileStore struct {
    basePath string
}

func NewLocalFileStore(basePath string) *LocalFileStore {
    return &LocalFileStore{basePath: basePath}
}

func (fs *LocalFileStore) Save(ctx context.Context, relPath string, reader io.Reader) (int64, error) {
    fullPath := filepath.Join(fs.basePath, filepath.FromSlash(relPath))
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0o750); err != nil {
        return 0, fmt.Errorf("creating directory %s: %w", dir, err)
    }

    f, err := os.Create(fullPath)
    if err != nil {
        return 0, fmt.Errorf("creating file %s: %w", fullPath, err)
    }
    defer f.Close()

    n, err := io.Copy(f, reader)
    if err != nil {
        os.Remove(fullPath) // cleanup on error
        return 0, fmt.Errorf("writing file %s: %w", fullPath, err)
    }

    return n, nil
}

func (fs *LocalFileStore) Delete(ctx context.Context, relPath string) error {
    fullPath := filepath.Join(fs.basePath, filepath.FromSlash(relPath))
    return os.Remove(fullPath)
}

func (fs *LocalFileStore) Exists(ctx context.Context, relPath string) (bool, error) {
    fullPath := filepath.Join(fs.basePath, filepath.FromSlash(relPath))
    _, err := os.Stat(fullPath)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}
```

### Unique Filename Generation
```go
// Source: Go stdlib crypto/rand + encoding/hex
func generatePhotoFilename(contentType string) (string, error) {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("generating random bytes: %w", err)
    }

    ext := ".bin"
    switch contentType {
    case "image/jpeg":
        ext = ".jpg"
    case "image/png":
        ext = ".png"
    case "image/gif":
        ext = ".gif"
    case "image/webp":
        ext = ".webp"
    }

    return hex.EncodeToString(b) + ext, nil
}
```

### Photo Serving (Download) Handler
```go
// Source: Go stdlib net/http.ServeFile, chi URL params
func HandleGetPhotoFile(svc *service.PhotoService, basePath string) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        photoID, err := parsePhotoIDParam(r)
        if err != nil {
            apierrors.WriteError(w, http.StatusBadRequest,
                apierrors.PhotoInvalidData, "invalid photo ID")
            return
        }

        // Look up metadata from DB -- path is server-controlled, not user-supplied
        photo, err := svc.GetByID(r.Context(), photoID)
        if err != nil {
            handlePhotoLookupError(w, err)
            return
        }

        fullPath := filepath.Join(basePath, filepath.FromSlash(photo.FilePath))
        w.Header().Set("Content-Type", photo.ContentType)
        http.ServeFile(w, r, fullPath)
    }
}
```

### Migration SQL
```sql
-- +goose Up
CREATE TABLE session_photos (
    id              BIGSERIAL PRIMARY KEY,
    session_id      BIGINT NOT NULL REFERENCES sessions(id),
    module_id       BIGINT REFERENCES session_modules(id),
    photo_type      VARCHAR(20) NOT NULL CHECK (photo_type IN ('before', 'label')),
    file_path       VARCHAR(500) NOT NULL,
    original_name   VARCHAR(255) NOT NULL,
    content_type    VARCHAR(100) NOT NULL,
    size_bytes      BIGINT NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by      BIGINT NOT NULL,
    CONSTRAINT chk_label_requires_module
        CHECK (photo_type != 'label' OR module_id IS NOT NULL)
);

CREATE INDEX idx_session_photos_session_id ON session_photos(session_id);
CREATE INDEX idx_session_photos_module_id ON session_photos(module_id) WHERE module_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS session_photos;
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Store photos as BYTEA in PostgreSQL | Store metadata in DB, files on filesystem | Long-standing best practice | Keeps DB small, backups fast, serves files efficiently |
| Trust Content-Type header | Detect via magic bytes (`http.DetectContentType`) | OWASP recommendation | Prevents MIME type spoofing attacks |
| Sequential/timestamp filenames | Cryptographic random hex IDs | Security best practice | No collisions, no information leakage |
| Direct `os` calls in business logic | FileStore interface | Standard Go testing pattern | Enables unit testing without filesystem |

**Deprecated/outdated:**
- None specific to this domain. All patterns use Go stdlib which is stable.

## Open Questions

1. **Maximum photo dimensions/resolution**
   - What we know: File size limit (10MB is reasonable for clinical photos)
   - What's unclear: Whether there should be maximum pixel dimensions
   - Recommendation: Enforce file size limit only (10MB). Do not add image resizing -- keep it simple per v1 scope

2. **Photo deletion policy**
   - What we know: Signed/locked sessions are immutable; photos linked to locked sessions should not be deletable
   - What's unclear: Whether photos on draft/in-progress sessions can be deleted
   - Recommendation: Allow photo deletion only when session is editable (draft/in_progress), consistent with module deletion pattern

3. **Immutability trigger for photos**
   - What we know: Phase 5 added immutability triggers for sessions and addendums
   - What's unclear: Whether session_photos table needs its own immutability trigger
   - Recommendation: Yes -- add a trigger that prevents UPDATE/DELETE on photos belonging to signed/locked sessions, consistent with Phase 5 pattern

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (existing) |
| Config file | None -- uses `go test ./...` |
| Quick run command | `go test ./internal/service/ -run TestPhoto -count=1 -v` |
| Full suite command | `make test` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PHOTO-01 | Upload before photo linked to session | unit | `go test ./internal/service/ -run TestUploadBeforePhoto -count=1 -v` | Wave 0 |
| PHOTO-02 | Upload label photo linked to module | unit | `go test ./internal/service/ -run TestUploadLabelPhoto -count=1 -v` | Wave 0 |
| PHOTO-03 | Filesystem storage with organized naming | unit | `go test ./internal/service/ -run TestPhotoFilePath -count=1 -v` | Wave 0 |
| PHOTO-04 | Consent gate blocks upload without consent | unit | `go test ./internal/service/ -run TestPhotoConsentRequired -count=1 -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/service/ -run TestPhoto -count=1 -v`
- **Per wave merge:** `make test`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/service/photo_test.go` -- covers PHOTO-01, PHOTO-02, PHOTO-03, PHOTO-04
- [ ] `internal/testutil/mock_photo.go` -- mock PhotoRepository
- [ ] `internal/testutil/mock_filestore.go` -- mock FileStore interface
- [ ] `internal/repository/postgres/photo_test.go` -- repository tests

## Sources

### Primary (HIGH confidence)
- Go stdlib `net/http` -- multipart form parsing, `DetectContentType`, `MaxBytesReader`, `ServeFile`
- Go stdlib `os`, `io`, `path/filepath`, `crypto/rand` -- filesystem operations and secure random
- Project codebase -- existing service/repository/handler patterns from Phases 1-5
- [OWASP File Upload Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html) -- security best practices

### Secondary (MEDIUM confidence)
- [Go multipart package docs](https://pkg.go.dev/mime/multipart) -- RFC 2046 multipart handling
- [How to process file uploads in Go](https://freshman.tech/file-upload-golang/) -- practical upload patterns
- [Chi file server example](https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go) -- static file serving with chi

### Tertiary (LOW confidence)
- None -- all findings verified against Go stdlib docs or project codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Go stdlib only, no new dependencies, verified against docs
- Architecture: HIGH - follows exact patterns from Phases 1-5 (service/repo/handler/mock)
- Pitfalls: HIGH - well-documented in OWASP and Go community; verified with multiple sources
- Consent gate: HIGH - existing `PhotoConsent` field already in Session domain, validation already in session.go

**Research date:** 2026-03-08
**Valid until:** 2026-04-08 (stable -- Go stdlib and project patterns are settled)
