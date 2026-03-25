---
# hearts-tnu6
title: Add immutable cache headers for card SVGs and icons
status: todo
type: task
priority: normal
created_at: 2026-03-25T06:42:40Z
updated_at: 2026-03-25T06:42:45Z
parent: hearts-e5b4
---

Set Cache-Control: public, max-age=31536000, immutable on card SVG and icon routes


## Context

All assets are served from `embed.FS` without any caching headers. Card SVGs, favicons, and icons are static content baked into the binary — they cannot change without a redeploy. Browsers currently re-fetch them on every page load unnecessarily.

## Higher Goal

Reduce bandwidth and improve load times, especially on mobile/flaky connections (epic hearts-e5b4). These assets account for the majority of HTTP requests per page load (56 cards alone) and are the lowest-risk caching target.

## Acceptance Criteria

- [ ] `/assets/cards/*` responses include `Cache-Control: public, max-age=31536000, immutable`
- [ ] `/favicon.ico`, `/icon.svg`, `/apple-touch-icon.png` responses include the same header
- [ ] Test coverage verifies caching headers are present on these routes

## Out of Scope

- CSS and JS caching (separate ticket with fingerprinting)
- ETag / conditional request support
- HTML page caching
