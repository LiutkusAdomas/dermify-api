---
phase: 5
slug: sign-off-and-compliance
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify v1 |
| **Config file** | None (Go stdlib test runner) |
| **Quick run command** | `go test ./internal/service/... -run TestSignoff -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/service/... -count=1 && go test ./internal/repository/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | META-02 | migration | `go test ./... -count=1` | No - W0 | pending |
| 05-01-02 | 01 | 1 | LOCK-04 | migration | `go test ./... -count=1` | No - W0 | pending |
| 05-01-03 | 01 | 1 | AUDIT-02, AUDIT-03 | migration | `go test ./... -count=1` | No - W0 | pending |
| 05-01-04 | 01 | 1 | LOCK-03, LOCK-05, LOCK-06 | migration + manual-DB | Manual: attempt UPDATE on signed session via psql | No - W0 | pending |
| 05-01-05 | 01 | 1 | AUDIT-01, AUDIT-04 | migration + manual-DB | Manual: perform operations, query audit_trail | No - W0 | pending |
| 05-02-01 | 02 | 2 | LOCK-01 | unit | `go test ./internal/service/ -run TestValidateForSignoff -count=1` | No - W0 | pending |
| 05-02-02 | 02 | 2 | LOCK-02, META-02 | unit | `go test ./internal/service/ -run TestSignOff -count=1` | No - W0 | pending |
| 05-02-03 | 02 | 2 | LOCK-04 | unit | `go test ./internal/service/ -run TestCreateAddendum -count=1` | No - W0 | pending |
| 05-02-04 | 02 | 2 | AUDIT-02 | unit | `go test ./internal/service/ -run TestAuditEntry -count=1` | No - W0 | pending |
| 05-03-01 | 03 | 3 | LOCK-01, LOCK-02 | unit | `go test ./internal/api/handlers/ -run TestSignOff -count=1` | No - W0 | pending |
| 05-03-02 | 03 | 3 | LOCK-04 | unit | `go test ./internal/api/handlers/ -run TestAddendum -count=1` | No - W0 | pending |
| 05-03-03 | 03 | 3 | AUDIT-01 | unit | `go test ./internal/api/handlers/ -run TestAudit -count=1` | No - W0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/service/signoff_test.go` — stubs for LOCK-01, LOCK-02, LOCK-03
- [ ] `internal/service/addendum_test.go` — stubs for LOCK-04
- [ ] `internal/service/audit_test.go` — stubs for AUDIT-02
- [ ] `internal/testutil/mock_addendum.go` — mock addendum repository
- [ ] `internal/testutil/mock_audit.go` — mock audit repository

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Signed session immutable at DB level | LOCK-03, LOCK-06 | Requires live PostgreSQL with triggers installed | Connect via psql, attempt UPDATE on signed session row, verify RAISE EXCEPTION |
| Addendum immutable at DB level | LOCK-05 | Requires live PostgreSQL with triggers installed | Connect via psql, attempt UPDATE on session_addendums row, verify RAISE EXCEPTION |
| Audit trail append-only | AUDIT-03 | Requires live PostgreSQL with triggers installed | Connect via psql, attempt UPDATE/DELETE on audit_trail row, verify RAISE EXCEPTION |
| All CRUD operations logged | AUDIT-01 | Requires live PostgreSQL with triggers installed | Perform INSERT/UPDATE/DELETE on clinical tables, query audit_trail for entries |
| Sign-off/lock events in audit trail | AUDIT-04 | Requires live PostgreSQL with triggers installed | Sign off a session, query audit_trail for status transition entries |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
