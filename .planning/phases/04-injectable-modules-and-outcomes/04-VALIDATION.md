---
phase: 4
slug: injectable-modules-and-outcomes
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (assert/require) |
| **Config file** | None (Go standard test runner) |
| **Quick run command** | `go test ./internal/service/... -count=1 -run "TestInjectable\|TestOutcome" -v` |
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
| 04-01-01 | 01 | 1 | FILL-01 | unit | `go test ./internal/service/... -count=1 -run TestCreateFillerModule -v` | ❌ W0 | ⬜ pending |
| 04-01-02 | 01 | 1 | FILL-02 | unit | `go test ./internal/service/... -count=1 -run TestFillerModule_Parameters -v` | ❌ W0 | ⬜ pending |
| 04-01-03 | 01 | 1 | FILL-03 | unit | `go test ./internal/service/... -count=1 -run TestFillerModule_ProductValidation -v` | ❌ W0 | ⬜ pending |
| 04-02-01 | 02 | 1 | TOX-01 | unit | `go test ./internal/service/... -count=1 -run TestCreateBotulinumModule -v` | ❌ W0 | ⬜ pending |
| 04-02-02 | 02 | 1 | TOX-02 | unit | `go test ./internal/service/... -count=1 -run TestBotulinumModule_Parameters -v` | ❌ W0 | ⬜ pending |
| 04-02-03 | 02 | 1 | TOX-03 | unit | `go test ./internal/service/... -count=1 -run TestBotulinumModule_ProductValidation -v` | ❌ W0 | ⬜ pending |
| 04-03-01 | 03 | 2 | OUT-01 | unit | `go test ./internal/service/... -count=1 -run TestRecordOutcome -v` | ❌ W0 | ⬜ pending |
| 04-03-02 | 03 | 2 | OUT-02 | unit | `go test ./internal/service/... -count=1 -run TestOutcome_Endpoints -v` | ❌ W0 | ⬜ pending |
| 04-03-03 | 03 | 2 | OUT-03 | unit | `go test ./internal/service/... -count=1 -run TestOutcome_Aftercare -v` | ❌ W0 | ⬜ pending |
| 04-03-04 | 03 | 2 | OUT-04 | unit | `go test ./internal/service/... -count=1 -run TestOutcome_RedFlagsRequired -v` | ❌ W0 | ⬜ pending |
| 04-03-05 | 03 | 2 | OUT-05 | unit | `go test ./internal/service/... -count=1 -run TestOutcome_FollowUp -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/service/injectable_module_test.go` — stubs for FILL-01 through TOX-03
- [ ] `internal/service/outcome_test.go` — stubs for OUT-01 through OUT-05
- [ ] `internal/testutil/mock_injectable_module.go` — mock repos for filler + botulinum
- [ ] `internal/testutil/mock_outcome.go` — mock repo for outcome

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
