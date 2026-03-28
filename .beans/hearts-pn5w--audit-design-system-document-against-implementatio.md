---
# hearts-pn5w
title: Audit design system document against implementation
status: todo
type: task
priority: normal
tags:
    - frontend
    - docs
created_at: 2026-03-28T16:08:40Z
updated_at: 2026-03-28T16:08:53Z
parent: hearts-dfll
---

## Context

The design system (`design-system.md`) documents the visual language — colors, typography, spacing, components, animations — but it may have drifted from the actual CSS/templates over time. Mismatches between the doc and the code cause confusion: contributors reference the doc and build the wrong thing, or the doc becomes untrustworthy and gets ignored.

## Higher Goal

Keep the design system document authoritative. Every statement in it should either match the current implementation or be flagged as an intentional future target.

## Acceptance Criteria

- [ ] Every CSS custom property value listed in the doc is compared against the actual stylesheet
- [ ] Component patterns (buttons, page header, surfaces, overlays, scoreboard) are spot-checked against their templates and CSS
- [ ] Animation timing values and custom property names are verified
- [ ] Typography (font family, size tiers, weights) verified against CSS
- [ ] Spacing and border-radius tiers spot-checked
- [ ] Each mismatch is documented with: what the doc says, what the code does, and a recommendation (fix doc / fix code / needs discussion)
- [ ] A follow-up bean is created for each actionable mismatch (or one grouped bean if mismatches are minor and related)

## Out of Scope

- Actually fixing any mismatches — this ticket only identifies them and creates follow-up tickets
- Evaluating whether the design system *should* change — that's for the follow-ups
- Non-visual concerns (architecture, API, game logic)
