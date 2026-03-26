---
# hearts-yywx
title: Convert session.Manager from RWMutex to actor pattern
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:55:11Z
updated_at: 2026-03-26T12:55:32Z
parent: hearts-6e2o
---

## Context

`session.Manager` in `session/manager.go` uses `sync.RWMutex` to protect its `map[string]*Table` registry. The RWMutex allows concurrent reads for `Get` and `List` (called from HTTP handlers), while writes (`Create`, `CloseTable`, `Close`) take exclusive locks.

Converting to an actor serializes all operations including reads. This is likely fine — table lookups are fast map reads, and the current read concurrency isn't load-bearing — but should be verified.

## Higher Goal

Consistent actor-based concurrency across the codebase (epic hearts-6e2o).

## Acceptance Criteria

- [ ] `Manager` uses a goroutine + channel instead of `sync.RWMutex`
- [ ] `Get`, `Create`, `List`, `CloseTable`, `Close` all work through the channel
- [ ] Synchronous callers (HTTP handlers calling `Get`) still get a response (use reply channels)
- [ ] Graceful shutdown: actor exits after closing all tables
- [ ] Existing tests pass; no race detector warnings

## Out of Scope

- Changing Manager's public API
- Adding table persistence
