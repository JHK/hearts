---
# hearts-pl10
title: Convert HumanPresence and PlayerPresence to actor pattern
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:55:08Z
updated_at: 2026-03-26T12:55:26Z
parent: hearts-6e2o
---

## Context

`HumanPresence` and `PlayerPresence` in `tracker/presence.go` are mutex-protected map counters tracking per-table connection counts. They're simple types (Join/Leave/Count), but converting them aligns the `tracker` package to a single concurrency style — all three trackers (`ConnTracker`, `HumanPresence`, `PlayerPresence`) would then be actors.

Consider whether these two can be merged into a single presence actor or should remain separate.

## Higher Goal

Consistent actor-based concurrency across the codebase (epic hearts-6e2o).

## Acceptance Criteria

- [ ] `HumanPresence` uses a goroutine + channel instead of `sync.Mutex`
- [ ] `PlayerPresence` uses a goroutine + channel instead of `sync.Mutex`
- [ ] Public API is preserved (or simplified if merging)
- [ ] Graceful shutdown support
- [ ] Existing tests pass; no race detector warnings

## Out of Scope

- Changing what presence information is tracked
- Adding persistence
