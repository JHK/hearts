---
# hearts-cm2r
title: Enhance debugBot to snapshot full bot decision context
status: todo
type: feature
priority: normal
tags:
    - backend
    - dx
created_at: 2026-03-25T16:05:42Z
updated_at: 2026-03-25T16:22:28Z
---

## Context

The current `debugBot()` console command only shows each bot's hand (name, seat, cards).
When observing a questionable bot play (e.g. bot plays K♠ instead of 9♦), there's no way
to capture the decision context at that moment. To reason about the play, you'd need:
trick state, played cards history, whether Q♠ is still out, hearts broken status, scores,
trick number, bot strategy, etc.

## Higher Goal

Make bot strategy issues debuggable without adding logging or stepping through code.
A single command should produce a markdown snapshot that can be pasted into a Claude session
for analysis.

## Acceptance Criteria

- [ ] `debugBot()` returns a structured snapshot including:
  - Each bot's hand, seat, name, and strategy type (smart/dumb/etc.)
  - Current trick number (0-12) and cards played in the current trick (with who played each)
  - Led suit for the current trick
  - All previously played cards (completed tricks)
  - Hearts broken status
  - First trick flag
  - Current scores (round + cumulative)
  - Whose turn it is
  - Pass direction for the round
  - Smart bot's moon-shot status (if applicable)
- [ ] Output is a pre-formatted markdown text block designed for pasting into a Claude conversation — labeled sections, not raw JSON
- [ ] JSON is also available (e.g. `debugBot({json: true})` or via the raw API endpoint)

## Out of Scope

- Logging every bot decision automatically (this is on-demand only)
- Changing bot strategy logic itself
- Persisting snapshots to disk or replaying games
- Preserving the old debugBot hand-only output
