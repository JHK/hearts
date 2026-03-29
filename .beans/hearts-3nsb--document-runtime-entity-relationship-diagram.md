---
# hearts-3nsb
title: Document runtime entity relationship diagram
status: completed
type: task
priority: normal
created_at: 2026-03-26T14:37:10Z
updated_at: 2026-03-29T12:35:01Z
---

## Context

The session runtime has several interacting actors (Handler, Manager, LobbyHub, Table, playerState) plus tracker actors (HumanPresence, PlayerPresence, ConnTracker). Their cardinality and communication patterns are not documented anywhere — you have to read the code to piece it together.

## Higher Goal

Make the runtime architecture easy to reason about for future development. A developer should be able to look at one diagram/document and understand: what entities exist, how many of each, who owns whom, and how they communicate.

## Entities to Document

### Core Actors
- **Handler** (webui) — the HTTP/WS entry point. Singleton.
  - 1:1 → LobbyHub
  - 1:1 → HumanPresence
  - 1:1 → PlayerPresence
  - ref → Manager (passed in, not owned)
  - ref → ConnTracker (passed in, not owned)

- **Manager** (session) — table lifecycle.
  - 1:N → Table (map[string]*Table)
  - 1:N → subscriber channels (table list changes)

- **LobbyHub** (webui) — lobby presence tracking. Actor goroutine.
  - N tokens (player presence in lobby)
  - 1:N → subscriber channels (lobby snapshots)

- **Table** (session) — one game session. Actor goroutine.
  - 1:N → playerState (up to 4 seated players, via slice + maps)
  - 1:1 → game.Round (nil before start)
  - 1:1 → game.Game
  - 1:N → event subscriber channels
  - callback → Manager.notifyChange

- **playerState** (session) — one seated player.
  - 1:1 → Table (owned by)
  - optional 1:1 → bot.Bot (nil for humans)
  - token → used for reconnection and multi-tab

### Tracker Actors
- **HumanPresence** — counts human connections per table (map[tableID]int)
- **PlayerPresence** — counts connections per player-table pair (map["tableID\x00playerID"]int)
- **ConnTracker** — tracks all active WebSocket connections globally (for graceful shutdown)

### Communication Patterns
- All actors use channel-based message passing (actor pattern)
- Manager → Table: direct reference + callback injection (avoids circular dep)
- Table → subscribers: event channels
- WebSocket handlers → actors: command channels with reply channels
- Multi-tab: tracked via token ref-counting in LobbyHub and PlayerPresence

## Acceptance Criteria

- [x] Add an entity-relationship section to `architecture.md` documenting all runtime entities, their cardinalities (1:1, 1:N), ownership, and communication mechanisms
- [x] Include an ASCII or Mermaid diagram showing the entity graph
- [x] Cover both the "static" ownership tree and the "dynamic" communication patterns (channels, callbacks, subscriptions)
- [x] Document the WebSocket → Table → Player connection flow (lobby WS vs table WS)

## Out of Scope

- Changing any code or architecture
- Domain layer entities (game.Round internals, card types, etc.)
- Protocol message types

## Summary of Changes

Added a comprehensive "Runtime entity relationships" section to `architecture.md` covering:
- Mermaid entity graph showing all actors and their relationships
- Core actors table with package, goroutine, ownership, and channel info
- Static ownership tree with cardinalities
- Communication patterns table (all channel-based message flows)
- Lobby and table WebSocket connection flow diagrams
- Multi-tab handling and orphan cleanup documentation
