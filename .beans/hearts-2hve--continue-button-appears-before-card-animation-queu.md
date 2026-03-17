---
# hearts-2hve
title: '"Continue" button appears before card animation queue drains'
status: todo
type: bug
priority: normal
created_at: 2026-03-17T07:37:29Z
updated_at: 2026-03-17T07:37:37Z
---

Continue button can appear while card animations are still playing at end of round

## Context
Card play animations are already sequentialized so players can follow bot moves. However, the end-of-round "Continue" button is not part of this queue and can appear while animations are still playing.

## Current Behavior
Playing the final card (or watching bots finish quickly) shows the "Continue" button before all queued card animations have completed.

## Desired Behavior
The "Continue" button appears only after the animation queue has fully drained — consistent with how individual card plays are already sequenced.

## Acceptance Criteria
- [ ] The "Continue" button does not appear until all pending card animations have finished
- [ ] Behavior is consistent regardless of whether the final card was played by a human or a bot

## Out of Scope
- Changes to the card play animation sequencing itself
- Adding a skip-animation option
