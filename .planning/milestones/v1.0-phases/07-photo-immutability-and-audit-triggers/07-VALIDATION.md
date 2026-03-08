---
phase: 7
slug: photo-immutability-and-audit-triggers
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify/assert |
| **Config file** | Makefile `test` target |
| **Quick run command** | `go build ./...` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go build ./...`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | LOCK-06 | build | `go build ./...` | N/A (SQL) | ⬜ pending |
| 07-01-02 | 01 | 1 | AUDIT-01 | build | `go build ./...` | N/A (SQL) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. This phase creates only a SQL migration file — no new Go test files needed.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| UPDATE/DELETE on session_photos blocked when session is signed/locked | LOCK-06 | Requires running PostgreSQL with triggers applied | 1. Apply migration 2. Create session, add photo, sign & lock session 3. Attempt UPDATE on session_photos row — expect error 4. Attempt DELETE — expect error |
| INSERT/UPDATE/DELETE on session_photos recorded in audit_trail | AUDIT-01 | Requires running PostgreSQL with triggers applied | 1. Apply migration 2. INSERT a photo row 3. Check audit_trail for INSERT entry 4. UPDATE photo row (on editable session) 5. Check audit_trail for UPDATE entry |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
