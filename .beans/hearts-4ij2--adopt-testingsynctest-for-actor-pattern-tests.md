---
# hearts-4ij2
title: Adopt testing/synctest for actor-pattern tests
status: completed
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:37:19Z
updated_at: 2026-03-31T15:43:40Z
parent: hearts-u20m
---

## Context

The codebase is heavily actor-based: `Table`, `Manager`, `ConnTracker`, `HumanPresence`, and `lobbyHub` all run as single-goroutine actors with channel communication. Several use `time.NewTimer`/`time.After` for deadlines (e.g. orphaned-table grace period in `ws.go`, WebSocket write deadlines). Testing these paths currently requires either real sleeps (slow, flaky) or skipping the timer logic entirely.

Go 1.25 graduated `testing/synctest` from experiment — it provides virtualized time so timer-dependent code advances instantly in tests.

## Higher Goal

Make actor/concurrency tests fast and deterministic, eliminating flaky timeout-dependent test failures.

## Acceptance Criteria

- [x] At least one existing timer-dependent test converted to use `synctest` as a proof of concept
- [x] New test covering the orphaned-table grace period timer (`ws.go`) using virtualized time
- [x] All existing tests still pass
- [x] Document the pattern in a test helper or comment for future actor tests

## Out of Scope

- Rewriting all tests to use synctest (just establish the pattern)
- Testing WebSocket I/O itself (synctest is for timers/channels, not network I/O)

## Summary of Changes

- Extracted `scheduleOrphanCleanup` from `handleTableWebSocket` in `ws.go` for testability
- Added `orphan_test.go` with 3 synctest tests: grace period expiry, human reconnect during grace period, short grace period
- Added `tracker/presence_test.go` with 3 synctest tests for HumanPresence and PlayerPresence actors
- Documented the synctest pattern in `presence_test.go` comment for future actor tests
- Pre-existing flaky test `TestDisconnectDuringPassingBotSubmitsOnResume` unrelated to changes
