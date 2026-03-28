---
# hearts-u969
title: Card selection fails in passing phase after navigating away and back
status: todo
type: bug
priority: normal
tags:
    - frontend
created_at: 2026-03-28T14:18:57Z
updated_at: 2026-03-28T14:19:09Z
---

Clicking cards to select for passing stops working after leaving the table page and returning. Desktop/mouse. Likely related to stale render state or event handler rebinding on reconnect.

## Context

Affects the passing phase on desktop. Triggered by navigating away from the table page and returning (e.g., going to the lobby and back). Card clicks stop registering for pass selection.

## Current Behavior

After leaving a table page and navigating back, clicking cards during the passing phase does not toggle their selection. The cards appear in the hand but are unresponsive to clicks.

## Desired Behavior

Card selection should work reliably in the passing phase regardless of navigation history. Returning to a table mid-pass should restore full interactivity.

## Acceptance Criteria

- [ ] Cards are clickable for pass selection after navigating away and back to the table
- [ ] Investigate whether bfcache restoration or `handRenderKey` staleness causes the DOM/event handler mismatch
- [ ] Fix verified in Chrome and Firefox on desktop

## Out of Scope

- Playing phase card interaction
- Mobile/touch-specific issues
- WebSocket reconnection reliability
