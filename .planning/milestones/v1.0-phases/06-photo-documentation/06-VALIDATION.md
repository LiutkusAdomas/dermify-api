---
phase: 6
slug: photo-documentation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (existing) |
| **Config file** | None — uses `go test ./...` |
| **Quick run command** | `go test ./internal/service/ -run TestPhoto -count=1 -v` |
| **Full suite command** | `make test` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/service/ -run TestPhoto -count=1 -v`
- **After every plan wave:** Run `make test`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | PHOTO-01 | unit | `go test ./internal/service/ -run TestUploadBeforePhoto -count=1 -v` | Wave 0 | pending |
| 06-01-02 | 01 | 1 | PHOTO-02 | unit | `go test ./internal/service/ -run TestUploadLabelPhoto -count=1 -v` | Wave 0 | pending |
| 06-01-03 | 01 | 1 | PHOTO-03 | unit | `go test ./internal/service/ -run TestPhotoFilePath -count=1 -v` | Wave 0 | pending |
| 06-01-04 | 01 | 1 | PHOTO-04 | unit | `go test ./internal/service/ -run TestPhotoConsentRequired -count=1 -v` | Wave 0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/service/photo_test.go` — stubs for PHOTO-01, PHOTO-02, PHOTO-03, PHOTO-04
- [ ] `internal/testutil/mock_photo.go` — mock PhotoRepository
- [ ] `internal/testutil/mock_filestore.go` — mock FileStore interface
- [ ] `internal/repository/postgres/photo_test.go` — repository tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Photo appears in filesystem with correct naming | PHOTO-03 | Requires real filesystem inspection | Upload a photo, check `{storage.base_path}/sessions/{id}/before/` for file |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
