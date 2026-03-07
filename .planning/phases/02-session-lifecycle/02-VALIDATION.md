---
phase: 02
slug: session-lifecycle
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-07
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + testify/assert |
| **Config file** | Makefile (test targets from Phase 1) |
| **Quick run command** | `go test ./internal/... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1 -v` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -count=1 -short`
- **After every plan wave:** Run `go test ./... -count=1 -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | SESS-01 | unit | `go test ./internal/service/ -run TestSession -v -count=1` | W0 | pending |
| 02-01-02 | 01 | 1 | SESS-02, SESS-03 | unit | `go test ./internal/service/ -run TestSession -v -count=1` | W0 | pending |
| 02-02-01 | 02 | 2 | SESS-04, SESS-05, SESS-06 | unit | `go test ./internal/service/ -run TestSession -v -count=1` | W0 | pending |
| 02-02-02 | 02 | 2 | CONS-01, CONS-02, CONS-03 | unit | `go test ./internal/service/ -run TestConsent -v -count=1` | W0 | pending |
| 02-03-01 | 03 | 2 | CONS-04, CONS-05 | unit | `go test ./internal/service/ -run TestContraindication -v -count=1` | W0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/service/session_test.go` — stubs for SESS-01 through SESS-06
- [ ] `internal/service/consent_test.go` — stubs for CONS-01 through CONS-05
- [ ] `internal/testutil/mock_session.go` — MockSessionRepository
- [ ] `internal/testutil/mock_consent.go` — MockConsentRepository

*Existing test infrastructure (testify, Makefile targets, testutil package) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Session state machine with database | SESS-02 | Requires running PostgreSQL | Create session, transition through states, verify invalid transitions return 400/409 |
| Consent gate blocks module addition | CONS-01 | Requires database + session with no consent | Create session, attempt to add module without consent, verify 403/422 |
| Patient session history integration | SESS-06 | Requires Phase 1 patient + Phase 2 sessions | Create patient, create sessions, verify GET /patients/:id/sessions returns them |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
