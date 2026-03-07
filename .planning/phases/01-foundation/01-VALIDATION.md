---
phase: 1
slug: foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-07
---

# Phase 1 — Validation Strategy

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
| 01-01-01 | 01 | 1 | RBAC-01 | unit | `go test ./internal/api/middleware/ -run TestRequireRole -v` | No -- Wave 0 | pending |
| 01-01-02 | 01 | 1 | RBAC-04 | unit | `go test ./internal/api/middleware/ -run TestRequireRole -v` | No -- Wave 0 | pending |
| 01-02-01 | 02 | 1 | PAT-01 | unit | `go test ./internal/service/ -run TestCreatePatient -v` | No -- Wave 0 | pending |
| 01-02-02 | 02 | 1 | PAT-02 | unit | `go test ./internal/service/ -run TestListPatients -v` | No -- Wave 0 | pending |
| 01-02-03 | 02 | 1 | PAT-03 | unit | `go test ./internal/service/ -run TestUpdatePatient -v` | No -- Wave 0 | pending |
| 01-02-04 | 02 | 1 | PAT-04 | unit | `go test ./internal/service/ -run TestPatientSessionHistory -v` | No -- Wave 0 | pending |
| 01-02-05 | 02 | 2 | RBAC-02 | integration | `go test ./internal/api/handlers/ -run TestDoctorAccess -v` | No -- Wave 0 | pending |
| 01-02-06 | 02 | 2 | RBAC-03 | integration | `go test ./internal/api/handlers/ -run TestAdminAccess -v` | No -- Wave 0 | pending |
| 01-03-01 | 03 | 1 | REG-01 | smoke | `go test ./internal/repository/postgres/ -run TestDeviceSeedData -v` | No -- Wave 0 | pending |
| 01-03-02 | 03 | 1 | REG-02 | smoke | `go test ./internal/repository/postgres/ -run TestProductSeedData -v` | No -- Wave 0 | pending |
| 01-03-03 | 03 | 1 | REG-03 | smoke | `go test ./internal/repository/postgres/ -run TestIndicationCodeSeedData -v` | No -- Wave 0 | pending |
| 01-03-04 | 03 | 1 | REG-04 | unit | `go test ./internal/service/ -run TestListDevices -v` | No -- Wave 0 | pending |
| 01-03-05 | 03 | 1 | META-01 | unit | `go test ./internal/service/ -run TestMetadataTracking -v` | No -- Wave 0 | pending |
| 01-03-06 | 03 | 1 | META-03 | unit | `go test ./internal/service/ -run TestVersionIncrement -v` | No -- Wave 0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/service/patient_test.go` -- stubs for PAT-01, PAT-02, PAT-03, PAT-04, META-01, META-03
- [ ] `internal/service/registry_test.go` -- stubs for REG-04
- [ ] `internal/service/role_test.go` -- stubs for RBAC-01
- [ ] `internal/api/middleware/auth_test.go` -- stubs for RBAC-01, RBAC-04
- [ ] `internal/api/handlers/patients_test.go` -- stubs for RBAC-02, RBAC-03
- [ ] Test infrastructure: mock implementations of repository interfaces for service-level unit tests
- [ ] Makefile `test` target: `go test ./... -count=1 -v`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Signed records track signed_at/by | META-02 | Schema columns exist but populated in Phase 5 | Verify columns exist in migration SQL |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
