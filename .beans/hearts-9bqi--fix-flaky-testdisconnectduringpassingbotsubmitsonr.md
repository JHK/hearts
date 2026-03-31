---
# hearts-9bqi
title: Fix flaky TestDisconnectDuringPassingBotSubmitsOnResume
status: completed
type: bug
priority: high
created_at: 2026-03-31T15:49:37Z
updated_at: 2026-03-31T15:57:10Z
---

## Context

`TestDisconnectDuringPassingBotSubmitsOnResume` in `internal/session/table_test.go:781` is flaky. It expects all 4 passes to be submitted after a disconnect/resume cycle during the passing phase, but intermittently sees only 2–3.

Observed during hearts-4ij2 work — confirmed pre-existing by running against the unchanged codebase.

## Higher Goal

Deterministic test suite with no flaky failures.

## Acceptance Criteria

- [x] Root-cause the race (likely bot pass commands not fully draining before assertion)
- [x] Fix the test or the underlying timing issue
- [x] Verify the fix is stable across repeated runs (`go test -count=50`)


## Summary of Changes

Root cause: `Drain()` was a heuristic (10 channel round-trips) that could not guarantee bot pass goroutines had been scheduled by the Go runtime. Replaced with `testing/synctest` which deterministically waits for all goroutines in the bubble to block. Removed the now-unused `Drain()` helper.
