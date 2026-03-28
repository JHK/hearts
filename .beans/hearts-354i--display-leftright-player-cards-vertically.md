---
# hearts-354i
title: Display left/right player cards vertically
status: todo
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-28T15:45:38Z
updated_at: 2026-03-28T15:45:49Z
parent: hearts-dfll
---

Rotate card backs for left and right seats to a vertical fan, simulating a real table where side players hold cards perpendicular to you.

## Context

Currently all player positions (top, left, right, bottom) display card backs in an identical horizontal row. On a real card table, the players to your left and right would hold their cards perpendicular to your viewpoint — their hand fans out vertically from your perspective.

## Higher Goal

Strengthen the "sitting at a real table" illusion. Positional cues make it easier to read the game state at a glance and give the table a more polished, immersive feel.

## Acceptance Criteria

- [ ] Left seat card backs are fanned vertically (stacked top-to-bottom with overlap, cards rotated 90°)
- [ ] Right seat card backs are mirrored — also vertical, fanned in the opposite direction
- [ ] Top seat remains horizontal (unchanged)
- [ ] Bottom seat (your hand) remains horizontal (unchanged)
- [ ] Layout works on mobile (single-column reflow) without breaking
- [ ] Trick-center played cards maintain their current orientation (no rotation)

## Out of Scope

- Showing actual card faces for opponents
- Animating the rotation or fanning
- Changing the bottom player's hand layout
