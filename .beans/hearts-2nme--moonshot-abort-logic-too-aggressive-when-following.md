---
# hearts-2nme
title: Moonshot abort logic too aggressive when following
status: todo
type: bug
priority: normal
created_at: 2026-03-25T17:20:17Z
updated_at: 2026-03-25T17:26:08Z
parent: hearts-8j8z
blocked_by:
    - hearts-gavv
---

Bot kills moonshot when another player leads, even when it holds all penalty points and strong cards

## Context
The smart bot's moonshot tracking in `internal/game/bot/smart.go` aborts the moonshot whenever another player leads a trick (lines 79-86). The dynamic re-activation check (lines 99-108) only runs when the bot is leading (`len(input.Trick) == 0`), so a bot following in a trick can never re-activate moonshot mid-trick.

## Current Behavior
When another player leads a trick, `moonShotAborted` is set to `true` unconditionally (line 83-84). The re-activation path requires `len(input.Trick) == 0`, so it never fires when the bot is following. This means a bot with 13 round points (e.g. from QS) and AH+KH in hand has its moonshot killed just because someone else led — even though it's in a strong position to collect all remaining penalty cards.

Observed in game: Edsger had 13 round points, held AC/JD/3H/KH/AH with 5 tricks remaining, and `moonShotActive` was `false` because Margaret led 6H.

## Desired Behavior
The bot should maintain or re-activate moonshot pursuit when it has captured all penalty points so far and still holds cards capable of winning remaining penalty-card tricks. The abort-on-follow heuristic should be refined — either:
1. Run the dynamic re-activation check when following too (not just when leading), or
2. Factor in current penalty-point ownership before aborting, or
3. Both

## Acceptance Criteria
- [ ] A bot that has collected all penalty points so far and holds safe high cards for remaining tricks keeps moonshot active when following
- [ ] The fix does not cause false-positive moonshot pursuit (bot chasing moon when it's clearly lost penalty cards to others)
- [ ] A simulation run (`cmd/sim`, 50k+ iterations) shows no decreased win rates compared to baseline
- [ ] Existing `smart_test.go` moonshot tests pass; new test covers the follow-abort scenario

## Out of Scope
- Reworking the initial `evaluateMoonShot` hand-evaluation heuristic
- Changing moonshot pass strategy
