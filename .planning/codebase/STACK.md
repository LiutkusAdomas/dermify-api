# Technology Stack

**Analysis Date:** 2026-03-07

## Languages

**Primary:**
- Go 1.23 - All application code (`go.mod` specifies `go 1.23.0`)

**Secondary:**
- SQL - Database migrations in `migrations/*.sql`
- YAML - Configuration files (`config.yaml`, `golangci.yaml`, `docker-compose.yml`)

## Runtime

**Environment:**
- Go 1.23 (Alpine-based Docker image for production: `golang:1.23-alpine`)
- Binary compiled with `-ldflags="-X main.Commit=$(git rev-parse HEAD)"` to embed git commit

**Package Manager:**
- Go Modules
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- `github.com/go-chi/chi/v5` v5.0.10 - HTTP router and middleware framework
- `github.com/spf13/cobra` v1.7.0 - CLI command framework (serve, version subcommands)
- `github.com/spf13/viper` v1.16.0 - Configuration management with YAML + env var overrides

**Testing:**
- `github.com/stretchr/testify` v1.11.0 - Test assertions

**Build/Dev:**
- `golangci-lint` - Linting with strict config in `golangci.yaml` (~60 linters enabled)
- `github.com/swaggo/swag` - Swagger/OpenAPI doc generation from Go annotations
- `make` - Build system (`Makefile`)
- Docker - Multi-stage build (`Dockerfile`)

## Key Dependencies

**Critical (direct dependencies in `go.mod`):**
- `github.com/go-chi/chi/v5` v5.0.10 - HTTP router; all request handling flows through this
- `github.com/jackc/pgx/v4` v4.18.3 - PostgreSQL driver (used via `database/sql` interface with `pgx` driver name)
- `github.com/golang-jwt/jwt/v4` v4.5.2 - JWT access token generation and validation (HS256 signing)
- `golang.org/x/crypto` v0.40.0 - bcrypt password hashing (`bcrypt.GenerateFromPassword`, `bcrypt.CompareHashAndPassword`)
- `github.com/pressly/goose/v3` v3.26.0 - Database migration runner with embedded SQL filesystem support
- `github.com/prometheus/client_golang` v1.15.1 - Prometheus metrics (counters, histograms)
- `github.com/google/uuid` v1.6.0 - UUID generation for request IDs

**Infrastructure (indirect but important):**
- `github.com/swaggo/http-swagger/v2` v2.0.2 - Swagger UI HTTP handler served at `/swagger/*`
- `github.com/swaggo/swag` v1.16.6 - Swagger doc generation from Go annotations
- `github.com/go-chi/chi/v5/middleware` - Used for `WrapResponseWriter` in logging and metrics middleware

## Configuration

**Environment:**
- Configuration loaded via Viper from `config.yaml` file
- Environment variable overrides use `OVERRIDE_` prefix (e.g., `OVERRIDE_DATABASE_HOST`, `OVERRIDE_PORT`)
- Nested config keys use underscore separator: `database.host` -> `OVERRIDE_DATABASE_HOST`
- Config struct defined in `config/config.go`

**Key config sections:**
```yaml
environment: local          # Runtime environment name
port: 8080                  # HTTP listen port
database:                   # PostgreSQL connection
  host, port, user, password, dbname, sslmode
auth:                       # JWT settings
  jwt_secret                # HMAC signing key
  access_token_expiry: 15m  # Access token TTL
  refresh_token_expiry: 168h # Refresh token TTL (7 days)
cors:                       # CORS policy
  allowed_origins            # List of allowed origins
  allowed_methods            # List of allowed HTTP methods
  allowed_headers            # List of allowed headers
```

**Config file location:**
- Default: `config.yaml` in working directory
- Override via CLI flag: `--config` / `-c`
- Config loading: `config/config.go` -> `Configure(filepath string)`

**Build:**
- `Makefile` - Build targets: `build`, `build-image`, `run-image`, `lint`, `lint-fix`, `swagger`
- `Dockerfile` - Multi-stage: builder (golang:1.23-alpine) -> runtime (alpine)
- `docker-compose.yml` - PostgreSQL 15 + API service with health checks

## Platform Requirements

**Development:**
- Go 1.23+
- `golangci-lint` installed globally (for `make lint`)
- `swag` CLI installed globally (for `make swagger`, installed in Docker build)
- PostgreSQL 15 (via `docker-compose up` or local install)
- Make

**Production:**
- Alpine Linux container
- PostgreSQL 15+ (external, connection via config/env vars)
- Port 8080 exposed
- `config.yaml` must be present in working directory (or path specified)

## CLI Interface

**Entry point:** `main.go` -> `cmd.Execute()` (Cobra)

**Commands:**
- `dermify-api serve [--config config.yaml]` - Start the HTTP server (`cmd/serve.go`)
- `dermify-api version` - Print git commit hash (`cmd/version.go`)

## Linting

**Tool:** golangci-lint with strict configuration

**Config:** `golangci.yaml` (300 lines)

**Key settings:**
- ~60 linters enabled (explicit enable list, `disable-all: true`)
- Cyclomatic complexity max: 30 (`cyclop`)
- Cognitive complexity max: 20 (`gocognit`)
- Function length max: 100 lines / 50 statements (`funlen`)
- No global variables (`gocheckcompilerdirectives`)
- No init functions (`gochecknoinits`)
- No naked returns (`nakedret`)
- Security checks enabled (`gosec`)
- Prometheus metric naming enforced (`promlinter`)
- Test files excluded from: `bodyclose`, `dupl`, `funlen`, `goconst`, `gosec`, `noctx`, `wrapcheck`
- `docs/` directory skipped

---

*Stack analysis: 2026-03-07*
