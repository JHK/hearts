---
# hearts-yipb
title: Show player presence in lobby
status: todo
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:13:06Z
updated_at: 2026-03-25T09:13:13Z
parent: hearts-g7wu
---

Display names of other players currently in the lobby to create a sense of liveness

## Context
The lobby currently gives no indication of whether other humans are online and looking for a game. A player arriving at an empty-looking lobby doesn't know if anyone else is around.

## Higher Goal
Create a sense of liveness so players know it's worth waiting or creating a table.

## Acceptance Criteria
- [ ] The lobby shows the names of other players currently browsing it
- [ ] When there are too many players to display, overflow is summarized (e.g. "Alice, Bob, Carol and 5 others are waiting")
- [ ] The list updates in real-time as players arrive and leave
- [ ] Players joining a table are removed from the lobby presence list

## Out of Scope
- Chat or direct interaction between lobby players
- "Looking for game" status or matchmaking queue
