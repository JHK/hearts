---
# hearts-dzxi
title: Add gzip compression middleware for static assets and HTML
status: completed
type: feature
priority: normal
created_at: 2026-03-26T09:10:34Z
updated_at: 2026-03-26T13:08:45Z
parent: hearts-p6hh
---

Add middleware.Compress scoped to page and asset route groups (not WebSocket). SVG card files, HTML pages, CSS, and JS all compress well. Scope carefully so WebSocket upgrade requests are unaffected.

## Summary of Changes

Added gzip compression (level 5) via chi's built-in `middleware.Compress` to page, asset, and API route groups. WebSocket routes are excluded by scoping the middleware to a `r.Group` that only contains non-WS route registrations.
