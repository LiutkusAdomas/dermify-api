# Codebase Concerns

**Analysis Date:** 2026-03-07

## Tech Debt

**Stub/Hardcoded User Handlers (Critical):**
- Issue: All user CRUD handlers in `internal/api/handlers/users.go` return hardcoded data. `HandleListUsers` returns a static array of two fake users, `HandleCreateUser` returns a hardcoded ID of 3, `HandleGetUser` fabricates a user from the URL param, `HandleUpdateUser` and `HandleDeleteUser` do nothing to the database. These endpoints are fully exposed in the API with no indication they are stubs.
- Files: `internal/api/handlers/users.go` (lines 19-132)
- Impact: Any client calling `/api/v1/users` endpoints gets fake data and no persistence. Deletes and updates are no-ops. This is deceptive in production.
- Fix approach: Rewrite all handlers to use `*sql.DB` (already available through route wiring but not passed to these handlers -- `internal/api/routes/users.go` has `db` field but never passes it). Add proper database queries against the `users` table. Add `RequireAuth` middleware to the user routes group.

**Hardcoded Health Check Response:**
- Issue: `HandleHealth` in `internal/api/handlers/health.go` returns a hardcoded timestamp (`"2024-01-01T00:00:00Z"`) and version (`"1.0.0"`) instead of real values.
- Files: `internal/api/handlers/health.go` (lines 20-24)
- Impact: Health checks do not reflect actual server state. Timestamp is always stale. No database connectivity check is performed, so the endpoint cannot detect a degraded system.
- Fix approach: Use `time.Now()` for timestamp, inject the build version (available via `debug.ReadBuildInfo`), and add a `db.Ping()` check to verify database connectivity.

**No Service/Repository Layer:**
- Issue: Handlers directly call `db.QueryRow` and `db.Exec` with inline SQL. Auth logic (token storage, validation, revocation) lives in `internal/api/auth/auth.go` as package-level functions accepting `*sql.DB`. There is no separation between business logic and data access.
- Files: `internal/api/handlers/auth.go`, `internal/api/handlers/login.go`, `internal/api/auth/auth.go`
- Impact: Handlers are tightly coupled to the database, making unit testing impossible without a real database. Business logic cannot be reused outside HTTP handlers. SQL is scattered across multiple packages.
- Fix approach: Introduce a `internal/repository/` package for data access and an `internal/service/` package for business logic. Handlers should call services; services call repositories. This enables mocking in tests.

**Duplicate CORS Middleware:**
- Issue: Two CORS middleware implementations exist: `middleware.CORS` (hardcoded `Access-Control-Allow-Origin: *`) in `internal/api/middleware/cors.go` and `middleware.CORSWithConfig` (config-driven) in `internal/api/middleware/cors_config.go`. Only `CORSWithConfig` is used in `internal/api/api.go` line 70. The `CORS` function is dead code.
- Files: `internal/api/middleware/cors.go`, `internal/api/middleware/cors_config.go`
- Impact: Dead code causes confusion. If someone mistakenly uses `CORS` instead of `CORSWithConfig`, they get a wide-open wildcard origin policy.
- Fix approach: Delete `internal/api/middleware/cors.go` entirely. Only `CORSWithConfig` should exist.

**Empty Directories (Dead Scaffolding):**
- Issue: `cmd/server/` and `cmd/server/http/` are empty directories. `dermify/` is also an empty directory. These appear to be remnants of an earlier project structure.
- Files: `cmd/server/`, `cmd/server/http/`, `dermify/`
- Impact: Confusing for developers navigating the project. Suggests abandoned refactoring.
- Fix approach: Delete the empty directories.

**Stray Test Config Files in `config/`:**
- Issue: Multiple files with numeric names (`1290504905.yaml`, `1323721463.yaml`, etc.) exist in `config/`. These are leftover temp files from `config_test.go` that create temp files in the current directory with `os.CreateTemp(".", "*.yaml")`. The test has `defer os.Remove` but some were apparently not cleaned up.
- Files: `config/1290504905.yaml`, `config/1323721463.yaml`, `config/2299117044.yaml`, `config/3211288114.yaml`, `config/3671142822.yaml`, `config/3959122956.yaml`
- Impact: Clutter in the repo. These are untracked but pollute `git status`.
- Fix approach: Delete the stray files. Fix the test in `config/config_test.go` to use `t.TempDir()` instead of `os.CreateTemp(".", ...)` to guarantee cleanup.

**Committed Binary Artifacts:**
- Issue: `dermify-api.exe` (34MB) and `dermify-api.exe~` (34MB) are present in the repo root and appear as untracked files.
- Files: `dermify-api.exe`, `dermify-api.exe~`
- Impact: Bloats the repo. Should never be committed.
- Fix approach: Delete the files and add `*.exe` and `*.exe~` to `.gitignore`. The current `.gitignore` only has `dermify-api` (no `.exe` extension).

**Test Config Pollution Across Test Runs:**
- Issue: `config_test.go` sets `OVERRIDE_ENVIRONMENT` via `os.Setenv` but never calls `os.Unsetenv` afterward. Since `viper` uses global state (`viper.SetConfigFile`, `viper.AutomaticEnv`), test cases can bleed into each other.
- Files: `config/config_test.go` (lines 59-61)
- Impact: Tests may pass or fail depending on execution order. The env var from one test case persists into subsequent cases.
- Fix approach: Add `defer os.Unsetenv("OVERRIDE_ENVIRONMENT")` after each `Setenv` call. Better: use `t.Setenv()` which auto-restores after the test. Also reset viper state between test cases.

## Security Considerations

**Hardcoded JWT Secret in Committed Config:**
- Risk: `config.yaml` contains `jwt_secret: "change-me-in-production"` and `password: postgres`. This file is committed to git. If deployed without overriding via `OVERRIDE_AUTH_JWT_SECRET`, the API uses a guessable JWT secret.
- Files: `config.yaml` (line 11)
- Current mitigation: The `OVERRIDE_` env var prefix mechanism allows overriding at runtime.
- Recommendations: Add a startup check in `internal/api/api.go` that refuses to start if `jwt_secret` is still the default value when `environment` is not `local` or `test`. Log a warning even in local mode.

**No Rate Limiting:**
- Risk: Login endpoint (`/api/v1/auth/login`), registration (`/api/v1/auth/register`), and token refresh (`/api/v1/auth/refresh`) have no rate limiting. An attacker can brute-force credentials or flood the registration endpoint.
- Files: `internal/api/handlers/login.go`, `internal/api/handlers/auth.go`, `internal/api/routes/auth.go`
- Current mitigation: Login failures are counted via Prometheus metrics (`login_failure_total`), but nothing acts on the count.
- Recommendations: Add per-IP rate limiting middleware. Use `golang.org/x/time/rate` or a chi-compatible rate limiter like `go-chi/httprate`. Apply stricter limits to auth endpoints.

**No Request Body Size Limit:**
- Risk: `json.NewDecoder(r.Body).Decode(...)` is used without `http.MaxBytesReader`. An attacker can send arbitrarily large request bodies to exhaust server memory.
- Files: `internal/api/handlers/auth.go` (lines 47, 103, 147), `internal/api/handlers/login.go` (line 46)
- Current mitigation: None.
- Recommendations: Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, maxBytes)` in a global middleware or at the start of each handler that reads a body. A reasonable limit is 1MB for most endpoints.

**No Input Validation Beyond Empty Checks:**
- Risk: Registration accepts any non-empty string for email, username, and password. No email format validation, no minimum password length, no username character restrictions. SQL injection is prevented by parameterized queries, but invalid data gets stored.
- Files: `internal/api/handlers/auth.go` (lines 52-55)
- Current mitigation: PostgreSQL UNIQUE constraints prevent duplicate emails/usernames.
- Recommendations: Validate email format with `net/mail.ParseAddress` or regex. Enforce minimum password length (8+ characters). Restrict username to alphanumeric characters. Consider a validation library like `go-playground/validator`.

**User Endpoints Have No Authentication:**
- Risk: All `/api/v1/users` endpoints (list, create, get, update, delete) have no `RequireAuth` middleware. Any unauthenticated client can call them.
- Files: `internal/api/routes/users.go` (lines 27-34)
- Current mitigation: The handlers are stubs returning fake data, so no real damage currently occurs. But when real implementations are added, this will be a critical security gap.
- Recommendations: Add `r.Use(middleware.RequireAuth(cfg))` to the users route group before implementing real handlers.

**CORS Allows Credentials with Potential Wildcard Origin:**
- Risk: In `internal/api/middleware/cors_config.go`, when `cfg.CORS.AllowedOrigins` is empty, the middleware sets `Access-Control-Allow-Origin: *` while also setting `Access-Control-Allow-Credentials: true`. Browsers reject this combination, but the intent is unclear and the fallback to `*` may mask misconfiguration.
- Files: `internal/api/middleware/cors_config.go` (lines 18-20)
- Current mitigation: `config.yaml` has specific origins configured.
- Recommendations: Never fall back to `*` when credentials are allowed. If origins list is empty, deny CORS rather than opening it wide.

**Logout Silently Ignores Errors:**
- Risk: `HandleLogout` discards the error from `auth.RevokeRefreshToken` with `_ = auth.RevokeRefreshToken(db, oldHash)`. If the revocation fails, the token remains valid but the user believes they are logged out.
- Files: `internal/api/handlers/auth.go` (line 110)
- Current mitigation: None.
- Recommendations: Return an error response if revocation fails instead of silently succeeding.

## Performance Bottlenecks

**No Database Connection Pool Configuration:**
- Problem: `internal/pkg/postgres.go` calls `sql.Open` but never configures `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, or `SetConnMaxIdleTime`. The defaults (`MaxOpenConns=0` meaning unlimited, `MaxIdleConns=2`) are suboptimal.
- Files: `internal/pkg/postgres.go` (lines 19-30)
- Cause: Missing pool tuning means either too many connections under load (exhausting PostgreSQL `max_connections`) or too few idle connections (causing reconnection latency).
- Improvement path: Add pool configuration to `DatabaseConfig` in `config/config.go` and apply it in `Open()`. Recommended starting point: `MaxOpenConns=25`, `MaxIdleConns=10`, `ConnMaxLifetime=5m`.

**Prometheus Metrics Use Raw URL Paths (Cardinality Explosion):**
- Problem: The metrics middleware in `internal/api/middleware/metrics.go` (lines 57-59) uses `r.URL.Path` as a label value. For routes like `/api/v1/users/{id}`, each unique ID creates a new time series (e.g., `/api/v1/users/1`, `/api/v1/users/2`, etc.).
- Files: `internal/api/middleware/metrics.go` (lines 57-59)
- Cause: Using raw path instead of the route pattern. Chi provides `chi.RouteContext(r.Context()).RoutePattern()` for this purpose.
- Improvement path: Replace `r.URL.Path` with `chi.RouteContext(r.Context()).RoutePattern()` in the metrics middleware to use route templates (e.g., `/api/v1/users/{id}`) as labels.

**No Graceful Shutdown:**
- Problem: `http.ListenAndServe` in `internal/api/api.go` (line 82) blocks until an error occurs. There is no signal handling. When the process is killed, in-flight requests are dropped and the database connection may not be closed cleanly despite the `defer a.db.Close()`.
- Files: `internal/api/api.go` (lines 82-86)
- Cause: Using `http.ListenAndServe` instead of `http.Server` with `Shutdown(ctx)`.
- Improvement path: Create an `http.Server`, listen for `SIGINT`/`SIGTERM` with `signal.Notify`, and call `server.Shutdown(ctx)` with a timeout to drain connections gracefully.

## Fragile Areas

**Auth Package Mixes Concerns:**
- Files: `internal/api/auth/auth.go`
- Why fragile: This single file handles password hashing (bcrypt), JWT generation/validation, refresh token generation, SHA-256 hashing, and all database operations for refresh tokens. Changes to token storage logic risk breaking JWT logic and vice versa.
- Safe modification: Split into separate files: `password.go` (hashing), `jwt.go` (token generation/validation), `refresh.go` (refresh token operations including DB).
- Test coverage: Zero tests. No tests exist for any auth functions.

**Config Uses Global Viper State:**
- Files: `config/config.go` (lines 56-72)
- Why fragile: `Configure()` calls `viper.SetConfigFile`, `viper.SetEnvPrefix`, and `viper.AutomaticEnv` on the global viper instance. If called multiple times (e.g., in tests or if the app is restructured), settings from previous calls persist and interfere.
- Safe modification: Use `viper.New()` to create an isolated instance instead of the global singleton.
- Test coverage: `config/config_test.go` exists but has env var leakage issues (see Tech Debt section).

**Metrics Client Uses Type Assertions:**
- Files: `internal/api/metrics/prometheus.go` (lines 41-50)
- Why fragile: `IncrementFooCount` and similar methods retrieve metrics from a `map[string]prometheus.Metric` and type-assert to `prometheus.Counter`. If a metric is registered with the wrong type or a key is misspelled, this panics at runtime with no compile-time safety.
- Safe modification: Store typed fields (`fooCounter prometheus.Counter`, etc.) directly on the `Client` struct instead of using a generic map with string keys and type assertions.
- Test coverage: No tests exist for the metrics package.

## Scaling Limits

**Refresh Token Table Growth:**
- Current capacity: The `refresh_tokens` table grows with every login and token refresh. Revoked tokens are marked with `revoked_at` but never deleted.
- Limit: Over time, the table accumulates rows indefinitely. Lookups against `token_hash` are indexed, but the table will grow unbounded.
- Scaling path: Add a periodic cleanup job (cron or background goroutine) that deletes rows where `revoked_at IS NOT NULL` or `expires_at < NOW()`. Alternatively, add a TTL-based partition strategy.

**Single-Instance Architecture:**
- Current capacity: The API runs as a single process. JWT validation is stateless, but refresh token operations hit the database on every call.
- Limit: Horizontal scaling is not blocked by architecture, but there is no health-check that verifies DB connectivity (the `/health` endpoint is hardcoded). Load balancers cannot detect a degraded instance.
- Scaling path: Implement a real health check with DB ping. Consider connection pooling with PgBouncer for multi-instance deployments.

## Dependencies at Risk

**pgx v4 (Outdated Major Version):**
- Risk: `github.com/jackc/pgx/v4` is used but pgx v5 has been the current major version since 2022. v4 receives limited maintenance.
- Impact: Missing performance improvements, context-first API, and pgx v5's native `pgxpool` (better than `database/sql` pool). Potential security fixes may only land in v5.
- Migration plan: Upgrade to `github.com/jackc/pgx/v5/stdlib` (drop-in for `database/sql` usage) or migrate fully to `pgxpool` for better connection management.

**Deprecated Protobuf Dependency:**
- Risk: `github.com/golang/protobuf v1.5.3` is an indirect dependency (via Prometheus). The `golang/protobuf` module is deprecated in favor of `google.golang.org/protobuf`.
- Impact: Low immediate risk since it is indirect, but it will eventually stop receiving updates.
- Migration plan: This resolves automatically when `prometheus/client_golang` is upgraded to a newer version that uses the modern protobuf module.

## Test Coverage Gaps

**No Handler Tests:**
- What's not tested: All HTTP handlers (auth, login, users, health, foo, hello) have zero tests. No request/response cycle testing exists anywhere.
- Files: `internal/api/handlers/auth.go`, `internal/api/handlers/login.go`, `internal/api/handlers/users.go`, `internal/api/handlers/health.go`, `internal/api/handlers/foo.go`, `internal/api/handlers/hello.go`
- Risk: Any handler change can break the API with no automated detection. Auth logic (registration, login, token refresh, logout) is entirely untested.
- Priority: High

**No Auth Package Tests:**
- What's not tested: Password hashing, JWT generation/validation, refresh token operations, token hashing.
- Files: `internal/api/auth/auth.go`
- Risk: Security-critical code has no tests. A regression in JWT validation or bcrypt hashing would go undetected.
- Priority: High

**No Middleware Tests:**
- What's not tested: Auth middleware, CORS, logging, metrics, request ID generation.
- Files: `internal/api/middleware/auth.go`, `internal/api/middleware/cors.go`, `internal/api/middleware/cors_config.go`, `internal/api/middleware/logging.go`, `internal/api/middleware/metrics.go`, `internal/api/middleware/request_id.go`
- Risk: Middleware bugs (e.g., auth bypass, CORS misconfiguration) would not be caught.
- Priority: Medium

**No Route Registration Tests:**
- What's not tested: Route wiring, middleware application to route groups, URL patterns.
- Files: `internal/api/routes/manager.go`, `internal/api/routes/auth.go`, `internal/api/routes/users.go`, `internal/api/routes/api.go`
- Risk: A route could be accidentally removed or middleware detached without detection.
- Priority: Medium

**No Integration Tests:**
- What's not tested: Full request lifecycle with database. Migration execution. Auth flow end-to-end.
- Files: None exist.
- Risk: Individual unit tests (when added) may pass while the system fails when components are wired together.
- Priority: Medium

**Only One Test File Exists:**
- What's tested: `config/config_test.go` tests config loading from file and env var override. Two test cases only.
- Files: `config/config_test.go`
- Risk: This is the only automated test in the entire project. The test itself has env var leakage issues.
- Priority: Low (test exists but needs improvement)

---

*Concerns audit: 2026-03-07*
