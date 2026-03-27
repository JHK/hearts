---
# hearts-yywx
title: Convert session.Manager from RWMutex to actor pattern
status: completed
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:55:11Z
updated_at: 2026-03-27T10:37:02Z
parent: hearts-6e2o
---

## Context

`session.Manager` in `session/manager.go` uses `sync.RWMutex` to protect its `map[string]*Table` registry. The RWMutex allows concurrent reads for `Get` and `List` (called from HTTP handlers), while writes (`Create`, `CloseTable`, `Close`) take exclusive locks.

Converting to an actor serializes all operations including reads. This is likely fine — table lookups are fast map reads, and the current read concurrency isn't load-bearing — but should be verified.

## Higher Goal

Consistent actor-based concurrency across the codebase (epic hearts-6e2o).

## Acceptance Criteria

- [x] `Manager` uses a goroutine + channel instead of `sync.RWMutex`
- [x] `Get`, `Create`, `List`, `CloseTable`, `Close` all work through the channel
- [x] Synchronous callers (HTTP handlers calling `Get`) still get a response (use reply channels)
- [x] Graceful shutdown: actor exits after closing all tables
- [x] Existing tests pass; no race detector warnings

## Out of Scope

- Changing Manager's public API
- Adding table persistence


## Summary of Changes

Converted `session.Manager` from `sync.RWMutex`/`sync.Mutex` to an actor pattern with a goroutine + `chan func()`. All public methods (`Get`, `Create`, `List`, `CloseTable`, `Close`, `Subscribe`) now serialize through the actor via reply channels. The `notifyChange` callback passed to tables sends a fire-and-forget command to the actor. Graceful shutdown collects tables inside the actor, stops the actor goroutine, then closes tables outside it to avoid blocking other commands. All tests pass with `-race`.
