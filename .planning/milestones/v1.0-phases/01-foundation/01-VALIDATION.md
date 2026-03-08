---
phase: 1
slug: foundation
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-07
---

# Phase 1 -- Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + stretchr/testify v1.11.0 |
| **Config file** | None (uses `go test` defaults) |
| **Quick run command** | `go test ./internal/... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1 -v` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -count=1 -short`
- **After every plan wave:** Run `go test ./... -count=1 -v && make lint`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-00-01 | 00 | 0 | META-01 | scaffold | `make test` | Created in 01-00 | pending |
| 01-01-01 | 01 | 1 | RBAC-01 | unit | `go test ./internal/service/ -run TestAssignRole -v` | Created in 01-00, activated in 01-01 | pending |
| 01-02-01 | 02 | 2 | RBAC-01,04 | unit | `go test ./internal/api/middleware/ -run TestRequireRole -v` | Created in 01-00, activated in 01-02 | pending |
| 01-03-01 | 03 | 3 | PAT-01 | unit | `go test ./internal/service/ -run TestCreatePatient -v` | Created in 01-00, activated in 01-03 | pending |
| 01-03-02 | 03 | 3 | PAT-02 | unit | `go test ./internal/service/ -run TestListPatients -v` | Created in 01-00, activated in 01-03 | pending |
| 01-03-03 | 03 | 3 | PAT-03 | unit | `go test ./internal/service/ -run TestUpdatePatient -v` | Created in 01-00, activated in 01-03 | pending |
| 01-03-04 | 03 | 3 | PAT-04 | unit | `go test ./internal/service/ -run TestGetPatientSessions -v` | Created in 01-00, activated in 01-03 | pending |
| 01-04-01 | 04 | 3 | REG-04 | unit | `go test ./internal/service/ -run TestListDevices -v` | Created in 01-00, activated in 01-04 | pending |
| 01-04-02 | 04 | 3 | REG-01 | unit | `go test ./internal/service/ -run TestGetDeviceByID -v` | Created in 01-00, activated in 01-04 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements (addressed by Plan 01-00)

- [x] `internal/service/patient_test.go` -- stubs for PAT-01, PAT-02, PAT-03, PAT-04, META-01, META-03
- [x] `internal/service/registry_test.go` -- stubs for REG-04
- [x] `internal/service/role_test.go` -- stubs for RBAC-01
- [x] `internal/api/middleware/auth_test.go` -- stubs for RBAC-01, RBAC-04
- [x] `internal/api/handlers/patients_test.go` -- stubs for RBAC-02, RBAC-03
- [x] Test infrastructure: mock implementations of repository interfaces for service-level unit tests
- [x] Makefile `test` target: `go test ./... -count=1 -v`

All Wave 0 items are addressed by Plan 01-00-PLAN.md (Wave 0).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Signed records track signed_at/by | META-02 | Schema columns exist but populated in Phase 5 | Verify columns exist in migration SQL |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved
