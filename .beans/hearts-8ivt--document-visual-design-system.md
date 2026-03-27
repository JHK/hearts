---
# hearts-8ivt
title: Document visual design system
status: completed
type: task
priority: high
created_at: 2026-03-27T08:54:57Z
updated_at: 2026-03-27T10:23:52Z
parent: hearts-dfll
---

Establish and document the shared visual language for the Hearts game in a `design-system.md` file. This is the foundation that all other beans in the table UI redesign epic (and future design work) will reference.

## Context

The lobby redesign (hearts-c65d) introduced a modern aesthetic — dark ink tones, semi-transparent panels, subtle gradients, clean typography. The table page still uses an older visual approach. Before changing the table, we need to codify the visual direction so all changes are consistent and future tickets can reference a single source of truth.

## Higher Goal

Visual consistency across the entire game. Every future design bean should reference this document rather than inventing its own style.

## Acceptance Criteria

- [x] `design-system.md` exists at the repo root
- [x] Documents color palette (primary ink, muted text, accents, felt green, panel backgrounds) with hex values and CSS variable names
- [x] Documents typography: font families, sizes, weights used across lobby and table
- [x] Documents spacing and padding conventions
- [x] Documents component patterns: icon buttons, dropdowns, overlay vs inline controls, card surfaces
- [x] Documents the "visual mood" — the overarching aesthetic intent (lean, modern, casino-like)
- [x] CLAUDE.md updated to reference design-system.md in the documentation maintenance section

## Out of Scope

- Implementing any visual changes (that's the other beans)
- Creating new assets or icons
- Defining animations or transitions (can be added later)

## Summary of Changes

Created `design-system.md` at the repo root documenting the full visual language extracted from `styles.input.css` and HTML templates. Covers color palette (CSS variables and hex values for lobby/table scopes, felt, cards, badges, interactive states, chart colors), typography (font families, size scale, weights, effects), spacing conventions (base scale, common values, border-radius tiers), component patterns (buttons, card surfaces, settings panel/toggle, form inputs, overlays, scoreboard), play card sizing/animation, and animation timing variables. Updated CLAUDE.md documentation maintenance section to reference the new file.
