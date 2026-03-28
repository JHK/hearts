---
# hearts-354i
title: Display left/right player cards vertically
status: completed
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-28T15:45:38Z
updated_at: 2026-03-28T16:04:19Z
parent: hearts-dfll
---

Rotate card backs for left and right seats to a vertical fan, simulating a real table where side players hold cards perpendicular to you.

## Context

Currently all player positions (top, left, right, bottom) display card backs in an identical horizontal row. On a real card table, the players to your left and right would hold their cards perpendicular to your viewpoint — their hand fans out vertically from your perspective.

## Higher Goal

Strengthen the "sitting at a real table" illusion. Positional cues make it easier to read the game state at a glance and give the table a more polished, immersive feel.

## Acceptance Criteria

- [x] Left seat card backs are fanned vertically (stacked top-to-bottom with overlap, cards rotated 90°)
- [x] Right seat card backs are mirrored — also vertical, fanned in the opposite direction
- [x] Top seat remains horizontal (unchanged)
- [x] Bottom seat (your hand) remains horizontal (unchanged)
- [x] Layout works on mobile (single-column reflow) without breaking
- [x] Trick-center played cards maintain their current orientation (no rotation)

## Out of Scope

- Showing actual card faces for opponents
- Animating the rotation or fanning
- Changing the bottom player's hand layout

## Summary of Changes

Modified `styles.input.css` to display left/right seat card backs vertically:
- Left and right `.seat-hand` containers use `flex-direction: column` with centered alignment
- Card backs inside left/right seats are rotated 90° via `transform: rotate(90deg)`
- Overlap changed from horizontal (`margin-left: -8px`) to vertical (`margin-top: -20px`)
- Mobile breakpoint (≤700px) reverts to horizontal layout: removes rotation, restores row direction and horizontal overlap
- Top seat, bottom seat, and trick-center cards are completely unchanged
