---
# hearts-9bqi
title: Fix flaky TestDisconnectDuringPassingBotSubmitsOnResume
status: todo
type: bug
priority: high
created_at: 2026-03-31T15:49:37Z
updated_at: 2026-03-31T15:49:37Z
---

## Context

`TestDisconnectDuringPassingBotSubmitsOnResume` in `internal/session/table_test.go:781` is flaky. It expects all 4 passes to be submitted after a disconnect/resume cycle during the passing phase, but intermittently sees only 2–3.

Observed during hearts-4ij2 work — confirmed pre-existing by running against the unchanged codebase.

## Higher Goal

Deterministic test suite with no flaky failures.

## Acceptance Criteria

- [ ] Root-cause the race (likely bot pass commands not fully draining before assertion)
- [ ] Fix the test or the underlying timing issue
- [ ] Verify the fix is stable across repeated runs (`go test -count=50`)
