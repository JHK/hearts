---
# hearts-x3g6
title: Convert lobbyHub from mutex to actor pattern
status: completed
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:55:07Z
updated_at: 2026-03-26T13:42:42Z
parent: hearts-6e2o
---

## Context

`lobbyHub` in `webui/lobby_hub.go` manages lobby player presence and broadcasts snapshots to subscriber channels. All public methods acquire a single `sync.Mutex`, and several (`Join`, `Leave`, `UpdateName`) broadcast to subscriber channels while holding the lock. This is the strongest candidate for actor conversion because it already has a pub/sub pattern.

## Higher Goal

Consistent actor-based concurrency across the codebase (epic hearts-6e2o).

## Acceptance Criteria

- [x] `lobbyHub` uses a single goroutine + channel instead of `sync.Mutex`
- [x] All public methods (`Join`, `Leave`, `UpdateName`, `Subscribe`, `Snapshot`) work through the channel
- [x] Subscriber broadcast happens inside the actor loop, not under a lock
- [x] Graceful shutdown (actor exits when server stops)
- [x] Existing tests pass; no race detector warnings (`go test -race`)

## Out of Scope

- Changing the lobby snapshot format or subscriber API

## Summary of Changes

Converted `lobbyHub` from `sync.Mutex` to actor pattern (single goroutine + channel). All public methods (`Join`, `Leave`, `UpdateName`, `Subscribe`, `Snapshot`) now serialize through an ops channel. Added `Broadcast()` method to decouple broadcast from `Join`, fixing a message ordering race between `lobby_self` and `lobby_presence`. Added `Shutdown()` for graceful actor termination, wired into the server shutdown sequence via a new `Handler` wrapper type. All existing tests pass with `-race`.
