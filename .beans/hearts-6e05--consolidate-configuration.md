---
# hearts-6e05
title: Consolidate settings panels across lobby and table
status: completed
type: feature
priority: normal
created_at: 2026-03-26T14:32:26Z
updated_at: 2026-03-27T10:53:45Z
parent: hearts-5ceo
---

## Context
The lobby and table pages each have their own settings panel, but they expose
different options. The lobby panel only has the player name input; the table
panel only has animation speed, sound, and notification toggles. A player who
wants to change their name mid-game has to leave the table, and a player in the
lobby can't pre-configure animation speed before joining.

All settings already use a shared `hearts.*` localStorage namespace, so the
underlying storage is unified — only the UI is fragmented.

## Higher Goal
Part of the In-Game User Settings epic (hearts-5ceo). Players should be able to
customize their full experience from wherever they are, without navigating away
from their current context.

## Acceptance Criteria
- [x] Table settings panel includes a player name input
- [x] Changing name in-game sends an update to the server (other players see the new name)
- [x] Lobby settings panel includes animation speed, sound, and notification toggles
- [x] Both panels read/write the same localStorage keys (no drift)
- [x] Settings markup and styles are shared (not duplicated per page)

## Out of Scope
- Adding new settings beyond what already exists
- Server-side preference storage
- Redesigning the settings panel layout or visual style

## Summary of Changes

### Server-side
- Added `rename` command to the table WebSocket protocol
- Added `EventPlayerRenamed` event type and `PlayerRenamedData` contract
- Added `handleRename` handler in session runtime (validates seated human player, broadcasts rename event)
- Added WebSocket routing for `rename` command in ws.go

### Frontend
- Created shared `_settings_panel.html` Go template partial with all settings (name, speed, sound, notifications)
- Both `index.html` (lobby) and `table.html` use `{{template "settings_panel"}}` — no duplicated markup
- Lobby JS now initializes and persists speed, sound, and notification toggles to localStorage
- Table JS adds name input that sends `rename` commands (debounced 300ms) and persists to localStorage
- Table JS handles `rename_result` and `player_renamed` events
- Updated Go template parsing in `routes_pages.go` to prepend the shared partial
