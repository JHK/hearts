---
# hearts-5fnr
title: Extract presence trackers into own file or sub-package
status: completed
type: task
priority: low
created_at: 2026-03-26T09:55:32Z
updated_at: 2026-03-26T10:36:35Z
parent: hearts-aazx
---

## Context

server.go contains two presence trackers: `humanPresenceTracker` (humans per table) and `playerPresenceTracker` (connections per player per table, with multi-tab ref counting and grace periods). These are self-contained state machines with clear interfaces, called from WebSocket connect/disconnect paths.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [x] Presence trackers live in their own file or sub-package
- [x] All existing tests pass without modification
- [x] Re-evaluated current state of server.go before extracting
- [x] Decision on file vs sub-package is documented in a code comment or commit message

## Guidance

- **Strong sub-package candidate.** The presence trackers have a small API surface (join/leave/count) but ~100 lines of internal logic (ref counting, grace periods, timers). If the interface is as narrow as it appears, extract to `internal/webui/presence/` or similar.
- Note the connection tracker coming in hearts-eefe (graceful shutdown) — that work may expand the presence tracking surface. Coordinate if both are in-flight.

## Out of Scope

- Changing presence tracking behavior or grace period logic
- The lobby hub (lobby_hub.go is already separate)

## Summary of Changes

Extracted all three trackers (`ConnTracker`, `HumanPresence`, `PlayerPresence`) from `server.go` into `internal/webui/tracker/` sub-package. Each file exports its own API with no inner-package dependencies — see `tracker.md` for rationale. server.go reduced from 878 to ~630 lines. Updated hearts-3ub1 (WS extraction) guidance to note the new package layout.
