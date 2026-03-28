---
# hearts-pn5w
title: Audit design system document against implementation
status: completed
type: task
priority: normal
tags:
    - frontend
    - docs
created_at: 2026-03-28T16:08:40Z
updated_at: 2026-03-28T16:14:56Z
parent: hearts-dfll
---

## Context

The design system (`design-system.md`) documents the visual language — colors, typography, spacing, components, animations — but it may have drifted from the actual CSS/templates over time. Mismatches between the doc and the code cause confusion: contributors reference the doc and build the wrong thing, or the doc becomes untrustworthy and gets ignored.

## Higher Goal

Keep the design system document authoritative. Every statement in it should either match the current implementation or be flagged as an intentional future target.

## Acceptance Criteria

- [x] Every CSS custom property value listed in the doc is compared against the actual stylesheet
- [x] Component patterns (buttons, page header, surfaces, overlays, scoreboard) are spot-checked against their templates and CSS
- [x] Animation timing values and custom property names are verified
- [x] Typography (font family, size tiers, weights) verified against CSS
- [x] Spacing and border-radius tiers spot-checked
- [x] Each mismatch is documented with: what the doc says, what the code does, and a recommendation (fix doc / fix code / needs discussion)
- [x] A follow-up bean is created for each actionable mismatch (or one grouped bean if mismatches are minor and related)

## Out of Scope

- Actually fixing any mismatches — this ticket only identifies them and creates follow-up tickets
- Evaluating whether the design system *should* change — that's for the follow-ups
- Non-visual concerns (architecture, API, game logic)

## Audit Findings

Audited `design-system.md` against `styles.input.css` and JS assets. Five mismatches found, all minor.

### 1. Table-page standard button gradient end color
- **Doc says**: Standard teal buttons use `linear-gradient(130deg, var(--accent), #1b8d8f)` everywhere except on felt
- **Code does**: Lobby buttons use `#1b8d8f` (correct), but `.table-page button` (line 999) uses `#1a8587`
- **Recommendation**: Fix code — unify to `#1b8d8f` to match doc and lobby buttons

### 2. `--anim-scoreboard-flip` defined but never used
- **Doc says**: Animation custom property `--anim-scoreboard-flip` at 400ms/200ms fast
- **Code does**: Property is defined in CSS (lines 317, 327) but never referenced in any CSS rule or JS file
- **Recommendation**: Either implement the scoreboard flip animation or remove the property from CSS and the doc

### 3. Icon button hover described as "lift" but no transform
- **Doc says**: Icon buttons "hover lifts with box-shadow"
- **Code does**: Hover only changes `background` and `box-shadow` — no `transform: translateY(...)` for actual lift
- **Recommendation**: Fix doc — change "lifts" to "gains box-shadow" (the visual effect is fine as-is)

### 4. Badge font size outside documented small tier
- **Doc says**: Small size tier is `0.78–0.88rem`
- **Code does**: `.badge` uses `0.75rem` (line 248), below the documented minimum
- **Recommendation**: Fix doc — widen small tier to `0.75–0.88rem`, or note badges as an exception

### 5. Font weight 300 not documented
- **Doc says**: Three weights: 400, 600, 700
- **Code does**: `.create-icon` uses `font-weight: 300` (line 280) for the "+" symbol
- **Recommendation**: Fix doc — either add 300 as a fourth weight or note it as a one-off exception for the create icon
