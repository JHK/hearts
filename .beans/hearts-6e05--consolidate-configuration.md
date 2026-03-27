---
# hearts-6e05
title: Consolidate settings panels across lobby and table
status: todo
type: feature
priority: normal
created_at: 2026-03-26T14:32:26Z
updated_at: 2026-03-27T10:37:25Z
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
- [ ] Table settings panel includes a player name input
- [ ] Changing name in-game sends an update to the server (other players see the new name)
- [ ] Lobby settings panel includes animation speed, sound, and notification toggles
- [ ] Both panels read/write the same localStorage keys (no drift)
- [ ] Settings markup and styles are shared (not duplicated per page)

## Out of Scope
- Adding new settings beyond what already exists
- Server-side preference storage
- Redesigning the settings panel layout or visual style
