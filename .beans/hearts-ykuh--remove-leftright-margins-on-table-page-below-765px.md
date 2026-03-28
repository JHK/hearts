---
# hearts-ykuh
title: Remove left/right margins on table page below 765px
status: todo
type: bug
priority: normal
tags:
    - frontend
created_at: 2026-03-28T16:17:10Z
updated_at: 2026-03-28T16:17:19Z
parent: hearts-dfll
---

## Context

The in-game table view has layout issues at viewport widths below 765px. Elements with left/right margins cause content to overflow or compress awkwardly at narrow widths.

## Current Behavior

At viewports narrower than ~765px, the table page UI breaks — elements with horizontal margins consume too much space, causing layout problems.

## Desired Behavior

Below 765px, left/right margins on table page elements are removed so the layout uses the full viewport width.

## Acceptance Criteria

- [ ] At 765px and below, table page elements have zero left/right margins
- [ ] Layout remains visually correct at common narrow widths (e.g. 375px, 414px, 480px, 768px)
- [ ] No regressions at wider viewports — existing margins preserved above the breakpoint
- [ ] Design system updated if this introduces a new responsive breakpoint or changes documented spacing

## Out of Scope

- Redesigning the table layout itself — this is just margin removal
- Other responsive issues unrelated to horizontal margins
