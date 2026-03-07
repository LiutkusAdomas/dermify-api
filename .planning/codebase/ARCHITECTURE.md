# Architecture

**Analysis Date:** 2026-03-07

## Pattern Overview

**Overall:** Layered monolithic REST API with CLI bootstrapping

**Key Characteristics:**
- Cobra CLI entrypoint delegates to a central `App` struct that owns server lifecycle
- Chi router with middleware chain handles HTTP concerns
- Route Manager pattern groups domain routes into separate registration units
- Handlers are closures receiving dependencies (db, config, metrics) via function parameters -- no dependency injection framework
- All application code lives under `internal/` to prevent external imports
- Database access happens directly in handlers and auth utilities (no repository/service layer yet)
- Embedded SQL migrations run automatically on startup

## Layers

**CLI Layer:**
- Purpose: Parse commands and flags, bootstrap the application
- Location: `cmd/`
- Contains: Cobra command definitions (`root.go`, `serve.go`, `version.go`)
- Depends on: `internal/api` (for `api.New()` and `api.Start()`)
- Used by: `main.go`

**Application Layer (App struct):**
- Purpose: Wire together config, database, logger, metrics, router, and middleware; manage server lifecycle
- Location: `internal/api/api.go`
- Contains: `App` struct with `New()` constructor and `Start()` method
- Depends on: `config`, `internal/api/metrics`, `internal/api/middleware`, `internal/api/routes`, `internal/pkg` (postgres), `migrations`
- Used by: `cmd/serve.go`

**Route Registration Layer:**
- Purpose: Group and register HTTP routes by domain using the Manager pattern
- Location: `internal/api/routes/`
- Contains: `Manager` struct (`manager.go`), domain route structs (`api.go`, `auth.go`, `users.go`)
- Depends on: `internal/api/handlers`, `internal/api/middleware`, `internal/api/metrics`, `config`
- Used by: `internal/api/api.go` via `App.createRoutes()`

**Handler Layer:**
- Purpose: HTTP request/response logic -- decode requests, call business logic, encode responses
- Location: `internal/api/handlers/`
- Contains: Handler functions (`auth.go`, `login.go`, `users.go`, `health.go`, `hello.go`, `foo.go`, `prometheus.go`), shared response models (`models.go`)
- Depends on: `internal/api/auth`, `internal/api/apierrors`, `internal/api/metrics`, `internal/api/middleware`, `config`, `database/sql`
- Used by: `internal/api/routes/`

**Auth Utilities Layer:**
- Purpose: JWT generation/validation, password hashing, refresh token lifecycle
- Location: `internal/api/auth/auth.go`
- Contains: `Claims` struct, password hashing (bcrypt), JWT operations (HS256), refresh token CRUD (direct SQL)
- Depends on: `database/sql`, `golang-jwt/jwt`, `golang.org/x/crypto/bcrypt`
- Used by: `internal/api/handlers/`, `internal/api/middleware/auth.go`

**Middleware Layer:**
- Purpose: Cross-cutting HTTP concerns applied to the request pipeline
- Location: `internal/api/middleware/`
- Contains: CORS (`cors.go`, `cors_config.go`), logging (`logging.go`), request ID (`request_id.go`), Prometheus metrics (`metrics.go`), JWT auth (`auth.go`)
- Depends on: `config`, `internal/api/auth`, `internal/api/apierrors`, `prometheus/client_golang`, `chi/v5/middleware`, `google/uuid`
- Used by: `internal/api/api.go` (global middleware), `internal/api/routes/auth.go` (per-route auth middleware)

**Error Response Layer:**
- Purpose: Structured JSON error responses with machine-readable error codes
- Location: `internal/api/apierrors/apierrors.go`
- Contains: `ErrorResponse` struct, error code constants, `WriteError()` helper
- Depends on: `net/http`, `encoding/json`
- Used by: `internal/api/handlers/`, `internal/api/middleware/auth.go`

**Metrics Layer:**
- Purpose: Define and expose Prometheus metrics
- Location: `internal/api/metrics/`
- Contains: `Client` struct (`prometheus.go`), counter definitions (`metrics.go`)
- Depends on: `prometheus/client_golang`, `log/slog`
- Used by: `internal/api/handlers/`, `internal/api/routes/`, `internal/api/api.go`

**Database Layer:**
- Purpose: PostgreSQL connection with retry logic and Goose migration runner
- Location: `internal/pkg/postgres.go`
- Contains: `Open()` (connection with retry), `MigrateFS()` / `Migrate()` (goose wrapper)
- Depends on: `jackc/pgx/v4/stdlib`, `pressly/goose/v3`
- Used by: `internal/api/api.go`

**Migrations:**
- Purpose: Schema evolution via embedded SQL files
- Location: `migrations/`
- Contains: `fs.go` (embed directive), SQL files (`20251213115406_users_table.sql`, `20260226120000_refresh_tokens_table.sql`)
- Depends on: Nothing (raw SQL)
- Used by: `internal/pkg/postgres.go` via `MigrateFS()`

**Configuration Layer:**
- Purpose: Load and merge YAML config with environment variable overrides
- Location: `config/config.go`
- Contains: `Configuration`, `DatabaseConfig`, `AuthConfig`, `CORSConfig` structs; `Configure()` loader
- Depends on: `spf13/viper`
- Used by: `internal/api/api.go`, `internal/api/routes/`, `internal/api/handlers/`, `internal/api/middleware/`

## Data Flow

**HTTP Request Lifecycle:**

1. Incoming HTTP request hits `net/http.ListenAndServe` on configured port
2. Chi router applies global middleware chain in order: CORS -> RequestID -> Logging -> Prometheus Metrics
3. Router matches path to a registered handler under `/api/v1/...`, `/metrics`, or `/swagger/*`
4. For protected routes (e.g., `/auth/me`), the `RequireAuth` middleware validates the JWT Bearer token and injects `Claims` into request context
5. Handler closure executes: decodes JSON body, performs validation, runs business logic (including direct SQL), encodes JSON response
6. Logging middleware logs the completed request with status code and duration
7. Prometheus metrics middleware records request count and latency

**Authentication Flow (Login):**

1. Client POSTs `{email, password}` to `/api/v1/auth/login`
2. `HandleLogin` queries `users` table for password hash by email
3. bcrypt comparison validates the password
4. On success: generates JWT access token (HS256, configurable expiry) and random refresh token
5. Refresh token hash (SHA-256) stored in `refresh_tokens` table with expiry
6. Returns `{access_token, refresh_token, expires_in}` to client

**Token Refresh Flow:**

1. Client POSTs `{refresh_token}` to `/api/v1/auth/refresh`
2. `HandleRefreshToken` hashes the token, validates against DB (not expired, not revoked)
3. Old refresh token is revoked (token rotation)
4. New access token and refresh token pair generated and returned

**State Management:**
- No in-memory application state -- all persistent state in PostgreSQL
- Request-scoped state (request ID, auth claims) stored in `context.Context`
- Metrics counters held in-memory by Prometheus registry (not persisted across restarts)

## Key Abstractions

**App (Application Container):**
- Purpose: Central struct owning all server dependencies and lifecycle
- Examples: `internal/api/api.go`
- Pattern: Constructor (`New()`) builds the struct, `Start()` runs the server. Dependencies (logger, config, metrics, db) are struct fields.

**Route Manager:**
- Purpose: Organize route registration by domain, inject dependencies into route groups
- Examples: `internal/api/routes/manager.go`, `internal/api/routes/auth.go`, `internal/api/routes/users.go`, `internal/api/routes/api.go`
- Pattern: `Manager` holds domain-specific route structs (`AuthRoutes`, `UserRoutes`, `APIRoutes`). Each route struct has a `RegisterRoutes(chi.Router)` method. `Manager.RegisterAllRoutes()` calls each one.

**Handler Closures:**
- Purpose: Bind dependencies to HTTP handlers without global state
- Examples: `internal/api/handlers/auth.go`, `internal/api/handlers/login.go`, `internal/api/handlers/users.go`
- Pattern: Each handler is a function returning `func(http.ResponseWriter, *http.Request)`. The outer function receives dependencies (`*sql.DB`, `*config.Configuration`, `*metrics.Client`), and the inner closure captures them.

**Metrics Client:**
- Purpose: Encapsulate Prometheus metric registration and increment operations
- Examples: `internal/api/metrics/prometheus.go`, `internal/api/metrics/metrics.go`
- Pattern: Factory function `New()` creates a `Client` with a custom `prometheus.Registry` and pre-registered counters stored in a map. Named increment methods provide type-safe access.

**Structured Error Responses:**
- Purpose: Provide machine-readable error codes alongside human messages
- Examples: `internal/api/apierrors/apierrors.go`
- Pattern: `WriteError(w, statusCode, code, message)` writes a JSON `{code, error}` response. Error codes are string constants grouped by category (validation, auth, user, internal).

## Entry Points

**Binary Entrypoint:**
- Location: `main.go`
- Triggers: Running the compiled binary
- Responsibilities: Calls `cmd.Execute()` which sets up Cobra CLI

**Serve Command:**
- Location: `cmd/serve.go`
- Triggers: `./dermify-api serve --config config.yaml`
- Responsibilities: Creates `api.App` with config path, calls `App.Start()` which opens DB, runs migrations, sets up router, starts HTTP server

**Version Command:**
- Location: `cmd/version.go`
- Triggers: `./dermify-api version`
- Responsibilities: Prints git commit hash from build info

**API Routes (all under `/api/v1`):**
- `/api/v1/health` - GET - Health check
- `/api/v1/hello` - GET - Hello world test endpoint
- `/api/v1/foo` - GET - Foo test endpoint with metrics
- `/api/v1/auth/register` - POST - User registration
- `/api/v1/auth/login` - POST - Authentication
- `/api/v1/auth/logout` - POST - Token revocation
- `/api/v1/auth/refresh` - POST - Token rotation
- `/api/v1/auth/me` - GET - Authenticated user profile (requires JWT)
- `/api/v1/users` - GET/POST - List/create users (stub)
- `/api/v1/users/{id}` - GET/PUT/DELETE - User CRUD (stub)

**Infrastructure Endpoints:**
- `/metrics` - GET - Prometheus metrics scrape endpoint
- `/swagger/*` - GET - Swagger UI for API documentation

## Error Handling

**Strategy:** Return structured JSON errors with HTTP status codes and machine-readable error codes. Never panic in handlers.

**Patterns:**
- Handlers use `apierrors.WriteError(w, statusCode, code, message)` for all error responses
- Error codes follow `CATEGORY_SPECIFIC_ERROR` naming (e.g., `AUTH_INVALID_CREDENTIALS`, `VALIDATION_REQUIRED_FIELDS`)
- Fatal errors during startup (`main.go`, `config.Configure()`, `App.Start()`) call `log.Fatalln()` or `os.Exit(1)`
- Database and auth utility functions wrap errors with `fmt.Errorf("context: %w", err)` for error chain propagation
- Handlers typically map specific errors to specific HTTP status codes (400, 401, 404, 409, 500)

## Cross-Cutting Concerns

**Logging:**
- `log/slog` with JSON handler writing to stdout
- Logger created in `App.New()` and passed to middleware and metrics
- Request logging middleware records method, path, status, duration, and request ID
- Structured fields via `slog.String()`, `slog.Int()`, `slog.Float64()`

**Validation:**
- Manual validation in handlers (check for empty fields, decode errors)
- No validation framework -- handlers check required fields directly
- Database constraints (UNIQUE on username/email) provide secondary validation

**Authentication:**
- JWT Bearer tokens (HS256) in `Authorization` header
- `RequireAuth` middleware validates tokens and injects `Claims` into context
- `middleware.GetUserClaims(ctx)` extracts claims downstream
- Refresh tokens stored as SHA-256 hashes in PostgreSQL with expiry and revocation support
- Token rotation on refresh (old token revoked, new pair issued)

**CORS:**
- Config-driven CORS middleware (`CORSWithConfig`) applied globally as first middleware
- Falls back to allow-all if no origins configured (development mode)
- Preflight (OPTIONS) requests handled automatically

**Metrics:**
- Custom Prometheus registry (not default global)
- Per-request metrics middleware tracks `chi_requests_total` and `chi_request_duration_milliseconds`
- Business metrics: `foo_request_total`, `login_success_total`, `login_failure_total`
- Exposed at `/metrics` endpoint

**Request Tracing:**
- UUID-based request IDs generated by `RequestID` middleware
- Honors upstream `X-Request-ID` header if present
- Stored in context, included in response header and log entries

---

*Architecture analysis: 2026-03-07*
