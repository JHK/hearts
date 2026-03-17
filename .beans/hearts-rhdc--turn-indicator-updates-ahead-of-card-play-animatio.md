---
# hearts-rhdc
title: Turn indicator updates ahead of card play animations
status: todo
type: bug
priority: normal
created_at: 2026-03-17T13:44:06Z
updated_at: 2026-03-17T13:44:14Z
---

Turn indicator jumps to next player while card-play animation is still in progress

## Context

The turn indicator (`#turnIndicator` text + `.turn` seat highlight) reflects `snapshot.turn_player_id` from the latest server snapshot. When a card is played, the backend publishes `EventCardPlayed` followed by `EventTurnChanged`. The frontend schedules a state refresh (`scheduleStateRefresh`) both during and after queue processing, so `renderState()` — which drives the turn indicator — can fire while the animation queue is still mid-flight.

## Current Behavior

When a player plays a card, the turn indicator (text label + seat highlight) immediately jumps to the next player while the card-play animation is still in progress. The visual order is:

1. Card play animation starts (520 ms + 80 ms buffer)
2. Turn indicator switches to next player ← happens here, too early
3. Animation finishes

## Desired Behavior

The turn indicator should only advance to the next player after the animation queue is fully drained — consistent with how the Continue button was gated in hearts-2hve. Visual order should be:

1. Card play animation finishes
2. (Trick collection animation finishes, if applicable)
3. Turn indicator switches to next player

## Acceptance Criteria

- [ ] Turn indicator text and seat `.turn` highlight do not change while the trick event queue is processing
- [ ] After the queue drains, the turn indicator reflects the correct next player
- [ ] No regression on the Continue button delay fix (hearts-2hve)

## Out of Scope

- Fixing any other desync issues (e.g. hand rendering during animations) — those are separate bugs
