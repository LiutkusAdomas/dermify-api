# Testing Patterns

**Analysis Date:** 2026-03-07

## Test Framework

**Runner:**
- Go standard `testing` package
- No separate test runner -- uses `go test`

**Assertion Library:**
- `github.com/stretchr/testify` v1.11.0
- Uses `testify/assert` sub-package (not `require`)

**Run Commands:**
```bash
go test ./...              # Run all tests
go test ./... -v           # Verbose output
go test ./... -count=1     # No cache
go test -cover ./...       # With coverage summary
go test -coverprofile=coverage.out ./...  # Generate coverage file
```

**Note:** There is no `make test` target defined in the `Makefile`. Tests are run directly with `go test`.

## Test File Organization

**Location:**
- Co-located with source code (same directory as the code under test)
- Test files use `_test.go` suffix

**Naming:**
- Test files: `{source_file}_test.go` -- e.g., `config.go` -> `config_test.go`
- Test functions: `Test{FunctionName}` -- e.g., `TestConfigure`

**Current test files:**
- `config/config_test.go` -- tests for configuration loading

**Structure:**
```
config/
  config.go
  config_test.go     # Co-located test file
```

## Test Structure

**Suite Organization:**
```go
func TestConfigure(t *testing.T) {
    type input struct {
        fileInput string
        envVar    string
    }

    testCases := []struct {
        name     string
        input    input
        expected config.Configuration
    }{
        {
            name: "file input with no env var",
            input: input{
                fileInput: `
environment: test123
port: 12345`,
            },
            expected: config.Configuration{
                Environment: "test123",
                Port:        12345,
            },
        },
        // ... more cases
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            // setup
            // action
            // assertion
        })
    }
}
```

**Patterns:**
- Table-driven tests are the preferred pattern
- Use `t.Run()` with descriptive test case names for sub-tests
- Loop variable capture: `tc := tc` before `t.Run()` (Go <1.22 safety, though project uses Go 1.23)
- Test case struct with `name`, typed `input`, and `expected` fields
- Use the `_test` package suffix (enforced by `testpackage` linter) -- tests use the external package API

**Setup Pattern:**
- Create temporary files with `os.CreateTemp(".", "*.yaml")`
- Clean up with `defer os.Remove(testFile.Name())`
- Set environment variables with `os.Setenv()` for override testing

**Assertion Pattern:**
- `assert.NoError(t, err)` for error checks
- `assert.Equal(t, expected, actual)` for value comparison

## Mocking

**Framework:** No mocking framework currently in use

**Current approach:**
- Tests use real filesystem operations (temp files) rather than mocks
- No mock implementations exist for `*sql.DB`, metrics client, or other dependencies
- The `exhaustruct` linter config excludes `github.com/stretchr/testify/mock.Mock`, suggesting mock support is planned

**What to Mock (when adding mocks):**
- Database connections (`*sql.DB`) -- use interfaces or `sqlmock`
- External service calls
- Time-dependent operations

**What NOT to Mock:**
- Configuration loading (use real temp files as shown in existing tests)
- Pure functions (test them directly)

## Fixtures and Factories

**Test Data:**
```go
// Inline YAML for config testing:
fileInput: `
environment: test123
port: 12345`

// Temp file creation for test fixtures:
testFile, err := os.CreateTemp(".", "*.yaml")
assert.NoError(t, err)
defer os.Remove(testFile.Name())

err = os.WriteFile(testFile.Name(), []byte(tc.input.fileInput), 6)
assert.NoError(t, err)
```

**Location:**
- No dedicated fixtures directory
- Test data is defined inline within test cases
- Temp files are created in the current directory and cleaned up via `defer`

## Coverage

**Requirements:** No coverage thresholds are enforced

**View Coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out        # Open HTML report
go tool cover -func=coverage.out        # Print per-function coverage
```

## Test Types

**Unit Tests:**
- Only `config/config_test.go` exists currently
- Tests configuration loading from YAML files and environment variable overrides
- Uses real filesystem (temp files) rather than mocks

**Integration Tests:**
- Not present. No database integration tests exist
- The `docker-compose.yml` provides a PostgreSQL instance for manual testing

**E2E Tests:**
- Not present. No HTTP-level or end-to-end tests exist
- No test helpers for spinning up the HTTP server

## Linter Rules for Tests

The golangci-lint config in `golangci.yaml` relaxes several rules for test files (`_test.go`):
- `bodyclose` -- disabled (HTTP response body close checks)
- `dupl` -- disabled (duplicate code allowed in tests)
- `funlen` -- disabled (test functions can be long)
- `goconst` -- disabled (repeated strings allowed)
- `gosec` -- disabled (security checks relaxed)
- `noctx` -- disabled (HTTP requests without context OK)
- `wrapcheck` -- disabled (error wrapping not required)

The `tenv` linter is enabled and configured with `all: true`, meaning it checks entire test files for `os.Setenv` usage (recommends `t.Setenv` instead). Note: the existing test in `config/config_test.go` uses `os.Setenv` which may trigger this linter.

The `testpackage` linter is enabled, enforcing that test files use a separate `_test` package (e.g., `package config_test` not `package config`).

## Writing New Tests

**Template for a new handler test:**
```go
package handlers_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "dermify-api/internal/api/handlers"

    "github.com/stretchr/testify/assert"
)

func TestHandleSomething(t *testing.T) {
    testCases := []struct {
        name           string
        expectedStatus int
        expectedBody   string
    }{
        {
            name:           "successful response",
            expectedStatus: http.StatusOK,
            expectedBody:   `{"key":"value"}`,
        },
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/path", nil)
            rec := httptest.NewRecorder()

            handler := handlers.HandleSomething()
            handler(rec, req)

            assert.Equal(t, tc.expectedStatus, rec.Code)
        })
    }
}
```

**Guidelines for new tests:**
- Place test file next to source file: `internal/api/handlers/auth_test.go`
- Use `package handlers_test` (external test package, enforced by linter)
- Use table-driven tests with `t.Run()`
- Use `testify/assert` for assertions
- Create temp files for filesystem-dependent tests; clean up with `defer`
- Do not rely on external state (running database, running server)
- Use `httptest.NewRequest` and `httptest.NewRecorder` for handler tests

## Gaps and Recommendations

**Critical gaps:**
- Only 1 test file exists (`config/config_test.go`) -- no handler, middleware, auth, or metrics tests
- No HTTP-level tests using `httptest`
- No database integration tests
- No mock infrastructure for `*sql.DB` or metrics client

**When adding tests, prioritize:**
1. Auth functions in `internal/api/auth/auth.go` (pure functions like `HashPassword`, `CheckPassword`, `GenerateAccessToken`, `ValidateAccessToken`)
2. Handler tests using `httptest` for `internal/api/handlers/`
3. Middleware tests for `internal/api/middleware/`
4. API error response formatting in `internal/api/apierrors/apierrors.go`

---

*Testing analysis: 2026-03-07*
