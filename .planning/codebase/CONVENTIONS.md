# Coding Conventions

**Analysis Date:** 2026-03-07

## Naming Patterns

**Files:**
- Use lowercase `snake_case` for Go source files: `cors_config.go`, `request_id.go`, `postgres.go`
- One handler file per domain: `auth.go`, `login.go`, `users.go`, `health.go`
- Shared response types in a dedicated `models.go` file: `internal/api/handlers/models.go`
- Package-level docs in `doc.go` per package (every package has one)
- Migration files use goose timestamp naming: `YYYYMMDDHHMMSS_description.sql`

**Functions:**
- Handler functions: `Handle{Action}()` -- always PascalCase verb prefix
  - Examples: `HandleHealth()`, `HandleLogin()`, `HandleRegister()`, `HandleListUsers()`, `HandleGetUser()`
- Constructor functions: `New{Type}()` or `new{type}()` for unexported
  - Examples: `NewManager()`, `NewAuthRoutes()`, `newFooCounter()`
- Middleware: standalone functions returning `func(http.Handler) http.Handler`
  - Examples: `RequestID`, `NewLoggingMiddleware()`, `RequireAuth()`, `CORSWithConfig()`
- Context extractors: `Get{Value}()` -- e.g., `GetRequestID()`, `GetUserClaims()`
- Metrics increment methods: `Increment{Name}Count()` -- e.g., `IncrementFooCount()`, `IncrementLoginSuccessCount()`

**Variables:**
- Use `camelCase` for local variables and struct fields
- Context keys use unexported empty struct types: `type requestIDKey struct{}`
- Constants use `camelCase` for unexported: `fooCounterMetric`, `requestName`
- Constants use `PascalCase` for exported: `RequestIDHeader`, `ValidationRequiredFields`

**Types:**
- Exported structs: `PascalCase` -- `App`, `Client`, `Manager`, `Configuration`
- Request/response types: lowercase `camelCase` for handler-local types (`loginRequest`, `loginResponse`, `registerRequest`)
- Exported response types in `models.go`: `PascalCase` (`HealthResponse`, `UserResponse`, `MessageResponse`)
- Error code constants: `SCREAMING_SNAKE_CASE` strings grouped by category (`AUTH_INVALID_CREDENTIALS`, `VALIDATION_REQUIRED_FIELDS`)

**Packages:**
- All lowercase, single word: `handlers`, `middleware`, `metrics`, `auth`, `apierrors`, `config`, `cmd`

## Code Style

**Formatting:**
- `goimports` enforced via golangci-lint (handles both formatting and import ordering)
- No additional formatter configuration (relies on standard Go formatting)

**Linting:**
- golangci-lint with strict config in `golangci.yaml`
- Run: `make lint` or `make lint-fix`
- 50+ linters enabled including: `errcheck`, `govet`, `staticcheck`, `gosec`, `funlen`, `gocognit`, `cyclop`, `bodyclose`, `godot`, `nakedret`
- Key constraints:
  - Max function length: 100 lines (ignoring comments), 50 statements
  - Max cyclomatic complexity: 30 (cyclop), 20 (gocognit)
  - No global variables (`gochecknoglobals`) -- except in `main` and test files
  - No init functions (`gochecknoinits`)
  - No naked returns (`nakedret`)
  - No named returns (`nonamedreturns`)
  - Comments must end in a period (`godot`)
  - `nolint` directives require specific linter name and explanation
- Test files are excluded from: `bodyclose`, `dupl`, `funlen`, `goconst`, `gosec`, `noctx`, `wrapcheck`
- `docs/` directory is skipped entirely

**Nolint usage pattern:**
- Use `//nolint:lintername // reason` with a short explanation
- Example: `//nolint:errcheck // response write` for `json.NewEncoder(w).Encode()`

## Import Organization

**Order:**
1. Standard library packages (`database/sql`, `encoding/json`, `net/http`, `time`, etc.)
2. Internal project packages (`dermify-api/config`, `dermify-api/internal/api/...`)
3. Third-party packages (`github.com/go-chi/chi/v5`, `github.com/prometheus/...`)

**Path Aliases:**
- Named import for `internal/pkg`: `postgres "dermify-api/internal/pkg"` (in `internal/api/api.go`)
- Named import for swagger: `httpSwagger "github.com/swaggo/http-swagger/v2"` (in `internal/api/routes/manager.go`)
- Blank import for docs: `_ "dermify-api/docs"` (in `internal/api/routes/manager.go`)
- Blank import for pgx driver: `_ "github.com/jackc/pgx/v4/stdlib"` (in `internal/pkg/postgres.go`)

## Handler Pattern

**All handlers follow the closure pattern** -- they return `func(w http.ResponseWriter, r *http.Request)`:

```go
func HandleSomething(db *sql.DB, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        // handler logic
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response) //nolint:errcheck // response write
    }
}
```

**Key conventions within handlers:**
- Set `Content-Type: application/json` at the top of every handler
- Dependencies (`db`, `config`, `metrics`) are injected via closure parameters
- Decode request body with `json.NewDecoder(r.Body).Decode(&req)`
- Extract URL params with `chi.URLParam(r, "paramName")`
- Use `apierrors.WriteError()` for all error responses (never raw `http.Error()`)
- Return early on errors (guard clauses)
- Encode response with `json.NewEncoder(w).Encode()` -- suppress errcheck with nolint

## Middleware Pattern

**Middleware functions return `func(http.Handler) http.Handler`:**

```go
// Simple middleware (no dependencies):
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // logic
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Middleware with dependencies (returns factory):
func NewLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // logic using logger
            next.ServeHTTP(ww, r)
        })
    }
}
```

**Context value pattern:**
- Define an unexported empty struct type as key: `type requestIDKey struct{}`
- Store with `context.WithValue(r.Context(), requestIDKey{}, value)`
- Retrieve with exported helper: `GetRequestID(ctx context.Context) string`

## Route Registration Pattern

**Routes use a Manager pattern** in `internal/api/routes/`:

```go
// Each domain has its own route struct:
type AuthRoutes struct {
    db      *sql.DB
    config  *config.Configuration
    metrics *metrics.Client
}

func NewAuthRoutes(db *sql.DB, cfg *config.Configuration, metrics *metrics.Client) *AuthRoutes {
    return &AuthRoutes{db: db, config: cfg, metrics: metrics}
}

func (ar *AuthRoutes) RegisterRoutes(router chi.Router) {
    router.Route("/auth", func(r chi.Router) {
        r.Post("/register", handlers.HandleRegister(ar.db, ar.metrics))
        // Protected routes use middleware group:
        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireAuth(ar.config))
            r.Get("/me", handlers.HandleGetProfile(ar.db, ar.metrics))
        })
    })
}
```

- The `Manager` struct in `internal/api/routes/manager.go` aggregates all route groups
- API routes are versioned under `/api/v1/`
- Metrics endpoint `/metrics` is unversioned
- Swagger UI at `/swagger/*` is unversioned

## Error Handling

**HTTP Error Responses:**
- Use `apierrors.WriteError(w, statusCode, errorCode, message)` from `internal/api/apierrors/apierrors.go`
- Every error response has a machine-readable `code` string and human-readable `error` message
- Error codes are grouped by category as constants:
  - `VALIDATION_*` for request validation failures
  - `AUTH_*` for authentication/authorization failures
  - `USER_*` for user-related failures
  - `INTERNAL_*` for server-side failures
- Response format: `{"code": "AUTH_INVALID_CREDENTIALS", "error": "invalid credentials"}`

**Internal Error Handling:**
- Wrap errors with context using `fmt.Errorf("action: %w", err)` -- e.g., `fmt.Errorf("hashing password: %w", err)`
- Non-handler code returns errors; only initialization code uses `os.Exit(1)` or `log.Fatalln`
- Handlers never panic; they return appropriate HTTP status codes

**Error suppression:**
- `json.NewEncoder(w).Encode()` return values are ignored with `//nolint:errcheck // response write`
- Discarded errors explicitly use `_ =` assignment: `_ = auth.RevokeRefreshToken(db, oldHash)`

## Logging

**Framework:** `log/slog` with JSON handler

**Initialization:**
```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

**Patterns:**
- Use structured logging with `slog.String()`, `slog.Int()`, `slog.Float64()` attributes
- Log levels: `Info` for normal operations, `Warn` for retries, `Error` for failures
- Include `request_id` in request-scoped logs
- Never use `fmt.Println` or `log.Println` in application code (only `log.Fatalln` during config init)

**Example:**
```go
logger.Info("request completed",
    slog.String("request_id", requestID),
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
    slog.Int("status", ww.Status()),
    slog.Float64("duration_ms", durationMs),
)
```

## Comments

**When to Comment:**
- Every exported function and type requires a comment starting with the symbol name (godoc convention)
- Package-level documentation in `doc.go` for every package
- Comments end with a period (enforced by `godot` linter)
- Use `//nolint` with explanation for intentional lint suppressions

**Swagger/API Documentation:**
- Handler functions include swaggo annotation comments for API documentation
- Annotations: `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`, `@Security`
- Generate with: `make swagger` (runs `swag init`)

**Example:**
```go
// HandleRegister creates a new user account with a hashed password.
//
//	@Summary		Register a new user
//	@Description	Creates a new user account with a hashed password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerRequest		true	"Registration details"
//	@Success		201		{object}	registerResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Router			/auth/register [post]
```

## Function Design

**Size:** Max 100 lines (excluding comments), enforced by `funlen` linter. Max 50 statements.

**Parameters:** Dependencies injected via closure for handlers. Pass `*sql.DB`, `*config.Configuration`, `*metrics.Client` explicitly -- no global state.

**Return Values:** Non-handler functions return `(value, error)`. No named returns (enforced by `nonamedreturns`). No naked returns (enforced by `nakedret`).

## Module Design

**Exports:**
- `internal/` package ensures nothing is importable externally
- Each package exports only what is needed by other internal packages
- Handler-local request/response types are unexported (`loginRequest`, `loginResponse`)
- Shared response types are exported in `internal/api/handlers/models.go`

**Barrel Files:** Not used. Each package exposes types and functions directly.

**Package Documentation:**
- Every package has a `doc.go` with a single comment line describing the package purpose
- Example: `// Package handlers contains code for the API endpoint handlers`

## Struct Tags

**JSON tags:** Always lowercase snake_case with `json:"field_name"`
**Swagger example tags:** Include `example:"value"` for API documentation
**Mapstructure tags:** Use `mapstructure:"field_name"` for config structs

```go
type loginRequest struct {
    Email    string `json:"email" example:"johndoe@example.com"`
    Password string `json:"password" example:"secretpassword"`
}
```

## Metrics Conventions

- Metric names use `snake_case` with `_total` suffix for counters
- Examples: `foo_request_total`, `login_success_total`, `login_failure_total`
- HTTP metrics: `chi_requests_total`, `chi_request_duration_milliseconds`
- All metrics are registered with a custom `prometheus.Registry` (not the default global)
- Metric definitions live in `internal/api/metrics/metrics.go`
- Metric client methods use `Increment{Name}Count()` pattern

## Configuration Conventions

- Default values in `config.yaml`
- Environment variable overrides use `OVERRIDE_` prefix (e.g., `OVERRIDE_PORT`, `OVERRIDE_DATABASE_HOST`)
- Nested config keys use `_` separator in env vars: `database.host` -> `OVERRIDE_DATABASE_HOST`
- Config struct uses `mapstructure` tags for Viper unmarshaling
- Instantiate via `config.Configure(filepath)` -- never direct struct instantiation

## Database Conventions

- Use `database/sql` with pgx driver (blank import `_ "github.com/jackc/pgx/v4/stdlib"`)
- Raw SQL queries (no ORM) -- inline SQL strings in handler/auth functions
- Use `$1`, `$2` positional parameters for PostgreSQL
- Migrations are embedded SQL files with goose Up/Down sections
- Migration naming: `YYYYMMDDHHMMSS_description.sql`
- Connection with retry logic in `internal/pkg/postgres.go`

---

*Convention analysis: 2026-03-07*
