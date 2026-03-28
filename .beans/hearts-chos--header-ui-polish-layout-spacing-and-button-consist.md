---
# hearts-chos
title: 'Header UI polish: layout, spacing, and button consistency'
status: completed
type: bug
priority: normal
tags:
    - frontend
created_at: 2026-03-27T15:37:01Z
updated_at: 2026-03-28T13:06:23Z
---

## Context
Several small visual issues in the table page header (`page-header` in `table.html`) that hurt polish on both mobile and desktop.

## Current Behavior
1. Lobby headline text doesn't adjust size on mobile viewports
2. Turn indicator shows redundant "pass 3 cards (right)" — the direction is already communicated elsewhere
3. Header icons have insufficient margin between them
4. "Back to Lobby" button sits on the right, causing the config button to shift position when it appears/disappears
5. Back button uses the same `icon-btn` style as menu buttons (settings), despite performing a navigation action rather than opening a menu. It also has a hover animation the others don't — inconsistency in both directions.

## Desired Behavior
1. Headline font size scales down on narrow viewports
2. Turn indicator shows "pass 3 cards" without the direction parenthetical
3. Icon buttons have enough spacing to not feel cramped
4. Back button is positioned on the left so other buttons stay in place
5. Back button is visually distinct from menu-opening icon buttons (different style or treatment), and hover/animation behavior is consistent across all header buttons

## Acceptance Criteria
- [x] Lobby headline is responsive on mobile
- [x] Pass phase turn indicator omits direction label
- [x] Header icon buttons have adequate margin/gap
- [x] Back to Lobby button is on the left side of the header
- [x] Back button is visually distinguishable from menu buttons
- [x] Hover/animation behavior is consistent across header buttons

## Out of Scope
- Redesigning the entire header layout
- Changing header button functionality

## Summary of Changes

- Moved back button to the left side of the header so icon buttons stay in place
- Restyled back button as a subtle ghost link (no border/background) distinct from icon-btn menu buttons
- Added responsive font sizing for headline on mobile (<480px)
- Removed redundant pass direction parenthetical from turn indicator
- Increased icon button gap from 6px to 10px
- Unified hover behavior: icon-btns get shadow, back button gets subtle background tint

## Follow-up Changes

- Removed table ID, status emoji, and observer badge from table header h1 to match lobby header
- Added Page Header section to design-system.md documenting shared structure and the rule that page-specific state doesn't belong in the header
- Documented back (navigation) vs icon (action) button distinction in design system
