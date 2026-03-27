---
# hearts-pl10
title: Convert HumanPresence and PlayerPresence to actor pattern
status: completed
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:55:08Z
updated_at: 2026-03-26T13:00:46Z
parent: hearts-6e2o
---

## Context

`HumanPresence` and `PlayerPresence` in `tracker/presence.go` are mutex-protected map counters tracking per-table connection counts. They're simple types (Join/Leave/Count), but converting them aligns the `tracker` package to a single concurrency style — all three trackers (`ConnTracker`, `HumanPresence`, `PlayerPresence`) would then be actors.

Consider whether these two can be merged into a single presence actor or should remain separate.

## Higher Goal

Consistent actor-based concurrency across the codebase (epic hearts-6e2o).

## Acceptance Criteria

- [x] `HumanPresence` uses a goroutine + channel instead of `sync.Mutex`
- [x] `PlayerPresence` uses a goroutine + channel instead of `sync.Mutex`
- [x] Public API is preserved (kept separate -- they track different dimensions)
- [x] Graceful shutdown support (Shutdown method on both)
- [x] Existing tests pass; no race detector warnings

## Out of Scope

- Changing what presence information is tracked
- Adding persistence


## Summary of Changes

Converted HumanPresence and PlayerPresence from sync.Mutex-protected maps to actor-pattern goroutines with channels, matching ConnTracker concurrency style. Kept them as separate types since they track different dimensions (per-table vs per-table-per-player). Added Shutdown method to both for graceful teardown. Public API (Join/Leave/Count) unchanged. Updated tracker.md.
