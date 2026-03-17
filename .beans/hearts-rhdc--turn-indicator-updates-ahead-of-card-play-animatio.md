---
# hearts-rhdc
title: Turn indicator updates ahead of card play animations
status: done
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

The turn indicator advances in sync with each card-play animation: the player whose card is currently flying is highlighted, so visually the card "comes from" the highlighted seat. After the trick collection animation, the winner is highlighted. Visual order per card:

1. Player X's seat gets `.turn` highlight
2. Player X's card flies to the centre (animation)
3. Next player's turn begins

## Acceptance Criteria

- [x] Turn indicator does not jump to the next player before the current player's card animation plays
- [x] Each card animation begins with the playing player's seat highlighted
- [x] After trick collection animation, the trick winner's seat is highlighted
- [x] After the queue drains, the turn indicator syncs to the server's authoritative turn player
- [x] No regression on the Continue button delay fix (hearts-2hve)

## Out of Scope

- Fixing any other desync issues (e.g. hand rendering during animations) — those are separate bugs
