---
# hearts-8ivt
title: Document visual design system
status: todo
type: task
priority: high
created_at: 2026-03-27T08:54:57Z
updated_at: 2026-03-27T08:54:57Z
parent: hearts-dfll
---

Establish and document the shared visual language for the Hearts game in a `design-system.md` file. This is the foundation that all other beans in the table UI redesign epic (and future design work) will reference.

## Context

The lobby redesign (hearts-c65d) introduced a modern aesthetic — dark ink tones, semi-transparent panels, subtle gradients, clean typography. The table page still uses an older visual approach. Before changing the table, we need to codify the visual direction so all changes are consistent and future tickets can reference a single source of truth.

## Higher Goal

Visual consistency across the entire game. Every future design bean should reference this document rather than inventing its own style.

## Acceptance Criteria

- [ ] `design-system.md` exists at the repo root
- [ ] Documents color palette (primary ink, muted text, accents, felt green, panel backgrounds) with hex values and CSS variable names
- [ ] Documents typography: font families, sizes, weights used across lobby and table
- [ ] Documents spacing and padding conventions
- [ ] Documents component patterns: icon buttons, dropdowns, overlay vs inline controls, card surfaces
- [ ] Documents the "visual mood" — the overarching aesthetic intent (lean, modern, casino-like)
- [ ] CLAUDE.md updated to reference design-system.md in the documentation maintenance section

## Out of Scope

- Implementing any visual changes (that's the other beans)
- Creating new assets or icons
- Defining animations or transitions (can be added later)
