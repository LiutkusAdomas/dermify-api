---
phase: 3
slug: energy-based-modules
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (assert/require) |
| **Config file** | None (Go standard test runner) |
| **Quick run command** | `go test ./internal/service/... -count=1 -run TestEnergy -v` |
| **Full suite command** | `make test` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/service/... -count=1 -v`
- **After every plan wave:** Run `make test`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | IPL-01 | unit | `go test ./internal/service/... -count=1 -run TestCreateIPLModule -v` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | IPL-02 | unit | `go test ./internal/service/... -count=1 -run TestIPLModule_Parameters -v` | ❌ W0 | ⬜ pending |
| 03-01-03 | 01 | 1 | IPL-03 | unit | `go test ./internal/service/... -count=1 -run TestIPLModule_DeviceValidation -v` | ❌ W0 | ⬜ pending |
| 03-01-04 | 01 | 1 | YAG-01 | unit | `go test ./internal/service/... -count=1 -run TestCreateNdYAGModule -v` | ❌ W0 | ⬜ pending |
| 03-01-05 | 01 | 1 | YAG-02 | unit | `go test ./internal/service/... -count=1 -run TestNdYAGModule_Parameters -v` | ❌ W0 | ⬜ pending |
| 03-01-06 | 01 | 1 | YAG-03 | unit | `go test ./internal/service/... -count=1 -run TestNdYAGModule_DeviceValidation -v` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 1 | CO2-01 | unit | `go test ./internal/service/... -count=1 -run TestCreateCO2Module -v` | ❌ W0 | ⬜ pending |
| 03-02-02 | 02 | 1 | CO2-02 | unit | `go test ./internal/service/... -count=1 -run TestCO2Module_Parameters -v` | ❌ W0 | ⬜ pending |
| 03-02-03 | 02 | 1 | CO2-03 | unit | `go test ./internal/service/... -count=1 -run TestCO2Module_DeviceValidation -v` | ❌ W0 | ⬜ pending |
| 03-02-04 | 02 | 1 | RF-01 | unit | `go test ./internal/service/... -count=1 -run TestCreateRFModule -v` | ❌ W0 | ⬜ pending |
| 03-02-05 | 02 | 1 | RF-02 | unit | `go test ./internal/service/... -count=1 -run TestRFModule_Parameters -v` | ❌ W0 | ⬜ pending |
| 03-02-06 | 02 | 1 | RF-03 | unit | `go test ./internal/service/... -count=1 -run TestRFModule_DeviceValidation -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/service/energy_module_test.go` — stubs for IPL-01 through RF-03
- [ ] `internal/testutil/mock_energy_module.go` — mock repos for all 4 detail types
- [ ] `internal/testutil/mock_registry.go` — update for device-type validation if needed

*Existing test infrastructure (Go test runner, testify) covers all phase requirements.*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
