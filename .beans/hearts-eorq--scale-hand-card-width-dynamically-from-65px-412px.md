---
# hearts-eorq
title: Scale hand card width dynamically from 65px (412px viewport) to 68px default
status: todo
type: task
priority: normal
created_at: 2026-03-18T09:58:55Z
updated_at: 2026-03-18T09:59:32Z
---

Replace hard-coded 65px card width with CSS clamp() that scales from 65px at 412px viewport to 68px at 490px, keeping small-screen chrome unchanged

## Context

The `max-width: 490px` media query hard-codes `.seat.bottom .hand-card` to `65px` — a size calculated to fit 13 cards on a 412px-wide viewport (hearts-dfph). On slightly wider phones (e.g. 430px or 480px) the cards are unnecessarily small. The 68px default is the right upper bound; 412px is the minimum supported width.

## Higher Goal

Use available screen real estate to make cards as large as possible without sacrificing the existing small-screen layout optimizations (borderless sections, no padding, no border-radius).

## Acceptance Criteria

- [ ] At 412px viewport width, hand cards are 65px wide (minimum, same as today)
- [ ] At or above the point where the default layout applies, hand cards are 68px wide
- [ ] Card width scales smoothly between the two bounds as viewport width increases
- [ ] All other `max-width: 490px` rules (no padding, no border-radius, borderless trick section) are unchanged

## Out of Scope

- Scaling any card dimension other than width (aspect ratio is fixed via existing CSS)
- Changing the overlap/margin between cards (those stay fixed)
- Layout changes for screens wider than the current default breakpoint
