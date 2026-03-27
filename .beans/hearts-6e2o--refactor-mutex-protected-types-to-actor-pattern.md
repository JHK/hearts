---
# hearts-6e2o
title: Refactor mutex-protected types to actor pattern
status: completed
type: epic
priority: normal
tags:
    - backend
created_at: 2026-03-26T12:54:48Z
updated_at: 2026-03-27T10:46:28Z
---

## Vision

All shared mutable state in the server is managed through the actor pattern (single goroutine + channel), eliminating mutexes and making data races structurally impossible. `connTracker` already demonstrates this pattern; the remaining mutex-protected types should follow suit.

## Context

The codebase has two concurrency styles: `connTracker` and `session.Table` use the actor pattern (goroutine owns state, external callers send messages via channels), while several other types use `sync.Mutex`/`sync.RWMutex`. Standardizing on actors simplifies reasoning about concurrency and removes an entire class of bugs.

**Current mutex holders:**

| Type | File | Lock | Notes |
|------|------|------|-------|
| `HumanPresence` | `tracker/presence.go` | Mutex | Trivial map counter |
| `PlayerPresence` | `tracker/presence.go` | Mutex | Trivial map counter |
| `session.Manager` | `session/manager.go` | RWMutex | Table registry; uses read concurrency for `Get`/`List` |
| `lobbyHub` | `webui/lobby_hub.go` | Mutex | Broadcasts to subscriber channels under lock |
| `Table.subsMu` | `session/table.go` | RWMutex | Subscriber registry on an otherwise actor-based type |

**Reference implementation:** `connTracker` in `tracker/conn.go` — all state lives in `run()`, callers send ops through a channel.

## Out of Scope

- Changing the `session.Table` game-command actor loop (already correct)
- The local `playerMu` in `ws.go` (scoped to a single connection handler, not a shared type)
- Persisting state across restarts

## Summary of Changes

All mutex-protected types have been converted to the actor pattern:
- HumanPresence + PlayerPresence (hearts-pl10)
- lobbyHub (hearts-x3g6)
- session.Manager (hearts-yywx)
- Table.subsMu (hearts-2hsh)

The codebase now uses actors consistently for all shared mutable state.
