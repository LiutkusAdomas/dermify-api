# External Integrations

**Analysis Date:** 2026-03-07

## APIs & External Services

**No external third-party API integrations detected.** The API is self-contained with no calls to external services (no Stripe, no email providers, no cloud APIs, etc.).

**Internal API Endpoints (served by this application):**

| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| GET | `/api/v1/health` | `HandleHealth` | Health check |
| GET | `/api/v1/hello` | `HandleHello` | Hello endpoint |
| GET | `/api/v1/foo` | `HandleFoo` | Foo endpoint (increments metric) |
| POST | `/api/v1/auth/register` | `HandleRegister` | User registration |
| POST | `/api/v1/auth/login` | `HandleLogin` | User login (returns JWT + refresh token) |
| POST | `/api/v1/auth/logout` | `HandleLogout` | Revoke refresh token |
| POST | `/api/v1/auth/refresh` | `HandleRefreshToken` | Rotate refresh token, get new access token |
| GET | `/api/v1/auth/me` | `HandleGetProfile` | Get authenticated user profile (requires JWT) |
| GET | `/api/v1/users` | `HandleListUsers` | List users (stub/hardcoded) |
| POST | `/api/v1/users` | `HandleCreateUser` | Create user (stub/hardcoded) |
| GET | `/api/v1/users/{id}` | `HandleGetUser` | Get user by ID (stub/hardcoded) |
| PUT | `/api/v1/users/{id}` | `HandleUpdateUser` | Update user (stub/hardcoded) |
| DELETE | `/api/v1/users/{id}` | `HandleDeleteUser` | Delete user (stub/hardcoded) |
| GET | `/metrics` | `HandleMetrics` | Prometheus metrics scrape endpoint |
| GET | `/swagger/*` | `httpSwagger.WrapHandler` | Swagger UI |

## Data Storage

**Database:**
- PostgreSQL 15 (via `docker-compose.yml`)
- Driver: `github.com/jackc/pgx/v4` v4.18.3 (used through `database/sql` standard interface)
- Connection: DSN built from config fields in `config/config.go` -> `DatabaseConfig.DSN()`
- Connection string format: `host=%s user=%s password=%s dbname=%s port=%d sslmode=%s`
- Connection helper: `internal/pkg/postgres.go` -> `Open()` with retry logic (10 attempts, 2s delay)
- Connection pool: Uses Go stdlib `database/sql` default pool settings (no custom pool config)

**Environment variables for database:**
- `OVERRIDE_DATABASE_HOST` - PostgreSQL host (default: `localhost`)
- `OVERRIDE_DATABASE_PORT` - PostgreSQL port (default: `5432`)
- `OVERRIDE_DATABASE_USER` - PostgreSQL user (default: `postgres`)
- `OVERRIDE_DATABASE_PASSWORD` - PostgreSQL password
- `OVERRIDE_DATABASE_DBNAME` - Database name (default: `dermify`)
- `OVERRIDE_DATABASE_SSLMODE` - SSL mode (default: `disable`)

**Database Tables:**
- `users` - User accounts (`migrations/20251213115406_users_table.sql`)
  - Columns: `id` (BIGSERIAL PK), `username` (VARCHAR(50) UNIQUE), `email` (VARCHAR(255) UNIQUE), `password_hash` (VARCHAR(255)), `bio` (TEXT nullable), `created_at`, `updated_at`
- `refresh_tokens` - JWT refresh token storage (`migrations/20260226120000_refresh_tokens_table.sql`)
  - Columns: `id` (BIGSERIAL PK), `user_id` (BIGINT FK->users), `token_hash` (VARCHAR(255) UNIQUE), `expires_at`, `created_at`, `revoked_at` (nullable)
  - Indexes: `idx_refresh_tokens_user_id`, `idx_refresh_tokens_token_hash`

**Migrations:**
- Tool: Goose v3 (`github.com/pressly/goose/v3`)
- Location: `migrations/*.sql` embedded via `//go:embed` in `migrations/fs.go`
- Run automatically on server start in `internal/api/api.go` -> `Start()`
- Migration runner: `internal/pkg/postgres.go` -> `MigrateFS()` / `Migrate()`

**File Storage:**
- None (no file upload/storage integration)

**Caching:**
- None (no Redis, Memcached, or in-memory cache)

## Authentication & Identity

**Auth Provider:** Custom (self-implemented JWT-based authentication)

**Implementation details:**
- Auth logic: `internal/api/auth/auth.go`
- Auth middleware: `internal/api/middleware/auth.go`
- Structured error codes: `internal/api/apierrors/apierrors.go`

**Access Tokens:**
- Type: JWT (JSON Web Token)
- Signing: HMAC-SHA256 (`jwt.SigningMethodHS256`)
- Library: `github.com/golang-jwt/jwt/v4`
- Default expiry: 15 minutes (configurable via `auth.access_token_expiry`)
- Claims: `user_id` (int64), `email` (string), standard JWT registered claims
- Transmitted via: `Authorization: Bearer <token>` header

**Refresh Tokens:**
- Type: 32-byte cryptographically random hex string
- Storage: SHA-256 hash stored in `refresh_tokens` PostgreSQL table
- Default expiry: 168 hours / 7 days (configurable via `auth.refresh_token_expiry`)
- Rotation: On refresh, old token is revoked and new token pair issued
- Revocation: Soft-delete via `revoked_at` timestamp column

**Password Handling:**
- Hashing: bcrypt via `golang.org/x/crypto/bcrypt` with `bcrypt.DefaultCost` (10)
- Functions: `HashPassword()`, `CheckPassword()` in `internal/api/auth/auth.go`

**Protected Routes:**
- `GET /api/v1/auth/me` - Uses `middleware.RequireAuth(cfg)` middleware
- Middleware extracts JWT claims and stores in request context via `userClaimsKey{}`
- Claims retrieval: `middleware.GetUserClaims(ctx)` returns `*auth.Claims`

**Environment variables for auth:**
- `OVERRIDE_AUTH_JWT_SECRET` - JWT signing secret (default: `change-me-in-production`)
- `OVERRIDE_AUTH_ACCESS_TOKEN_EXPIRY` - Access token TTL (default: `15m`)
- `OVERRIDE_AUTH_REFRESH_TOKEN_EXPIRY` - Refresh token TTL (default: `168h`)

## Monitoring & Observability

**Metrics:**
- Prometheus via `github.com/prometheus/client_golang` v1.15.1
- Custom registry (not default global) - `prometheus.NewRegistry()`
- Metrics client: `internal/api/metrics/metrics.go` -> `Client` struct
- Metric definitions: `internal/api/metrics/prometheus.go`
- Scrape endpoint: `GET /metrics` (served by `promhttp.HandlerFor`)

**Custom metrics:**
- `foo_request_total` (Counter) - Requests to foo endpoint
- `login_success_total` (Counter) - Successful logins
- `login_failure_total` (Counter) - Failed login attempts

**HTTP middleware metrics (per-request):**
- `chi_requests_total` (CounterVec) - Total requests by `code`, `method`, `path`
- `chi_request_duration_milliseconds` (HistogramVec) - Request latency by `code`, `method`, `path`
  - Default buckets: 300ms, 1200ms, 5000ms
- Middleware: `internal/api/middleware/metrics.go` -> `NewPrometheusMiddleware()`

**Error Tracking:**
- None (no Sentry, Datadog, Bugsnag, etc.)

**Logging:**
- `log/slog` with JSON handler to stdout
- Logger created in `internal/api/api.go` -> `slog.NewJSONHandler(os.Stdout, nil)`
- Request logging middleware: `internal/api/middleware/logging.go`
  - Logs: `request_id`, `method`, `path`, `status`, `duration_ms`
- Request ID middleware: `internal/api/middleware/request_id.go`
  - Generates UUID via `github.com/google/uuid`
  - Propagates upstream `X-Request-ID` header if present

## CI/CD & Deployment

**Hosting:**
- Docker container (Alpine-based, multi-stage build)
- No cloud-specific deployment configuration detected
- No Kubernetes manifests, Helm charts, or Terraform files detected

**CI Pipeline:**
- None detected (no `.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`, etc.)

**Docker:**
- `Dockerfile` - Multi-stage build:
  1. Builder stage: `golang:1.23-alpine` - installs make, git, swag; generates swagger docs; compiles binary
  2. Runtime stage: `alpine` - copies binary and `config.yaml`
  3. Entrypoint: `/app/dermify-api serve`
- `docker-compose.yml` - Development environment:
  - `dermify-db`: PostgreSQL 15 with health check (`pg_isready`)
  - `dermify-api`: Application built from Dockerfile, depends on healthy db
  - Volume: `postgres_data` for database persistence

## API Documentation

**Swagger/OpenAPI:**
- Generated by `swag init` from Go annotation comments
- Source annotations: handler functions in `internal/api/handlers/*.go` and `internal/api/api.go`
- Generated files: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
- UI served at: `GET /swagger/*` via `github.com/swaggo/http-swagger/v2`
- Base path: `/api/v1`
- Security definition: `BearerAuth` (apikey in Authorization header)

**Generate command:**
```bash
make swagger
# Runs: swag init -g internal/api/api.go --parseDependency --parseInternal -o docs
```

## Environment Configuration

**Required env vars for production override:**
- `OVERRIDE_AUTH_JWT_SECRET` - **Critical**: Must change from default `change-me-in-production`
- `OVERRIDE_DATABASE_HOST` - PostgreSQL host
- `OVERRIDE_DATABASE_PASSWORD` - PostgreSQL password
- `OVERRIDE_DATABASE_SSLMODE` - Should be `require` or `verify-full` in production

**Optional env vars:**
- `OVERRIDE_PORT` - HTTP listen port (default: 8080)
- `OVERRIDE_ENVIRONMENT` - Environment name (default: `local`)
- `OVERRIDE_CORS_ALLOWED_ORIGINS` - CORS allowed origins

**Secrets location:**
- No `.env` file detected
- Secrets passed via environment variables with `OVERRIDE_` prefix
- Default (insecure) values in `config.yaml` for local development

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None

## CORS Configuration

**Implementation:** Custom middleware in `internal/api/middleware/cors_config.go`
- Config-driven: reads `cors.allowed_origins`, `cors.allowed_methods`, `cors.allowed_headers` from config
- Fallback: If no origins configured, allows all (`*`) - development mode
- Preflight: Handles OPTIONS requests with 200 OK
- Max age: 300 seconds
- Credentials: Allowed (`Access-Control-Allow-Credentials: true`)

**Default allowed origins (local dev):**
- `http://localhost:3000`, `http://localhost:3001`, `http://localhost:5173`
- `http://127.0.0.1:3000`, `http://127.0.0.1:3001`, `http://127.0.0.1:5173`

---

*Integration audit: 2026-03-07*
