# CLAUDE.md — dermify-api

## Project Overview

Backend REST API for the Dermify application, built in Go 1.23.

## Tech Stack

- **Router:** go-chi/chi v5
- **CLI:** spf13/cobra + viper
- **Database:** PostgreSQL (jackc/pgx v4)
- **Migrations:** pressly/goose v3 (embedded SQL via `//go:embed`)
- **Metrics:** Prometheus (client_golang)
- **Testing:** stretchr/testify
- **Linting:** golangci-lint (strict config in `golangci.yaml`)

## Project Structure

```
cmd/              CLI commands (root, serve, version)
config/           Viper-based config loading (config.yaml + OVERRIDE_ env vars)
internal/
  api/
    api.go        App struct, server startup
    routes.go     Top-level route wiring
    middleware.go Middleware application
    handlers/     HTTP handlers — one file per domain (auth, users, health, etc.)
    metrics/      Prometheus metric definitions and client
    middleware/   Individual middleware (cors, logging, request_id, metrics)
    routes/       Modular route registration (manager pattern)
  pkg/
    postgres.go   DB connection + migration runner
migrations/       Goose SQL migration files + embedded FS
```

## Conventions

### Code Style
- All application code goes in `internal/` — nothing should be importable externally.
- Handler functions: `Handle{Action}()` (e.g. `HandleCreateUser`, `HandleLogin`).
- Middleware: standalone functions returning `func(http.Handler) http.Handler`.
- Routes are grouped by domain in `internal/api/routes/` using a Manager pattern.
- API paths are versioned: `/api/v1/...`. Metrics live at `/metrics` (unversioned).
- Use `log/slog` for structured logging (JSON output) — not `log` or `fmt.Println`.
- Metric names use snake_case with `_total` suffix for counters.
- Each package has a `doc.go` for package-level documentation.

### Configuration
- Default values live in `config.yaml`.
- Environment variable overrides use `OVERRIDE_` prefix (e.g. `OVERRIDE_PORT`).
- Config struct is in `config/config.go`.

### Database
- Connection helper in `internal/pkg/postgres.go`.
- Migrations are embedded SQL files in `migrations/`.
- Create new migrations with goose naming: `YYYYMMDDHHMMSS_description.sql`.

### Testing
- Use `testify/assert` for assertions.
- Table-driven tests preferred.
- Create temp files/fixtures in tests, don't rely on external state.

### Error Handling
- Return proper HTTP status codes with JSON error responses.
- Log fatal only during initialization; handlers should never panic.

## Build & Run

```bash
make build          # Build binary with git commit embedded
make lint           # Run golangci-lint
make lint-fix       # Auto-fix lint issues
make build-image    # Docker image
make run-image      # Run via Docker on :8080

# Database
docker-compose up   # Start PostgreSQL
./dermify-api serve --config config.yaml
```

## Rules for Claude

- Always place new application code under `internal/`.
- When adding a new endpoint: create a handler in `internal/api/handlers/`, register the route in the appropriate file under `internal/api/routes/`, and follow existing naming patterns.
- When adding a new domain (e.g. products, appointments): create separate handler, route, and migration files — don't pile everything into one file.
- Write migrations as raw SQL with both `-- +goose Up` and `-- +goose Down` sections.
- Do not modify `main.go` unless absolutely necessary — it should stay minimal.
- Respect the strict golangci-lint config. Run `make lint` mentally before suggesting code.
- Keep handlers thin — business logic should live in service/repository layers as they're introduced.
- Use context for request-scoped values (request ID, auth info).
- Never hardcode secrets or connection strings — use config or environment variables.
- Prefer returning errors over logging-and-continuing in non-handler code.
- When suggesting new dependencies, justify why they're needed over stdlib.
