---
# hearts-2hsh
title: Fold Table.subsMu into the Table actor loop
status: completed
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:55:11Z
updated_at: 2026-03-27T09:17:29Z
parent: hearts-6e2o
---

## Context

`session.Table` is already an actor for game commands, but it uses a separate `sync.RWMutex` (`subsMu`) to protect its subscriber registry (`subs` map). Subscribers are added/removed by WebSocket handlers concurrently with the game loop broadcasting events.

Moving subscribe/unsubscribe into the command channel eliminates the last mutex on Table, making it a pure actor. The tradeoff is that subscribe/unsubscribe becomes serialized with game commands, but these are infrequent operations.

## Higher Goal

Consistent actor-based concurrency across the codebase (epic hearts-6e2o).

## Acceptance Criteria

- [x] `Table.subsMu` is removed
- [x] Subscribe and Unsubscribe operations go through the command channel
- [x] Event broadcasting reads the subscriber map without a lock (only the actor goroutine touches it)
- [x] Existing tests pass; no race detector warnings

## Out of Scope

- Refactoring the Table command processing loop beyond what's needed for this change

## Summary of Changes

Removed `subsMu` RWMutex from `Table`. Subscribe and unsubscribe now go through the command channel as `subscribeCommand`/`unsubscribeCommand`, making `Table` a pure actor. The `emit` method no longer needs a lock since only the actor goroutine reads `subs`. `Close` cleanup is also lock-free since it waits for the actor to exit first.
