---
# hearts-1ph8
title: GPU powered animations
status: todo
type: task
priority: normal
created_at: 2026-04-08T13:29:22Z
updated_at: 2026-04-10T10:17:26Z
---

## Context

The lobby page renders 22 floating background cards using CSS `@keyframes` with `translate`. Anecdotally, this causes ~20–30% sustained CPU usage on a single core even when the tab is idle. Paint flashing confirms no repaints are happening, so the cost is likely in compositing: the cards lack explicit layer promotion (`will-change`), and `opacity`/`filter` sit on a child element rather than the animated wrapper, which may prevent the browser from efficiently compositing each card as a single GPU texture.

## Higher Goal

The lobby should be lightweight enough to sit in a background tab without noticeable CPU drain.

## Acceptance Criteria

- [ ] `will-change: translate` is set on `.card-bg-card` and each card gets its own compositor layer (verify in DevTools → Layers panel)
- [ ] `opacity` and `filter` are moved to `.card-bg-card` so the entire subtree rasterizes as one layer
- [ ] CPU usage on the lobby page is measurably reduced (compare before/after in DevTools → Performance)
- [ ] Visual appearance of the card background is unchanged
- [ ] `prefers-reduced-motion: reduce` still disables the animation

## Out of Scope

- Floating player names animation (separate concern)
- Table card flip animation (already GPU-accelerated)
- Reducing the number of cards or changing the visual design
