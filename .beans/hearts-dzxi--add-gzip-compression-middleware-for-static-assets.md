---
# hearts-dzxi
title: Add gzip compression middleware for static assets and HTML
status: todo
type: feature
created_at: 2026-03-26T09:10:34Z
updated_at: 2026-03-26T09:10:34Z
parent: hearts-p6hh
---

Add middleware.Compress scoped to page and asset route groups (not WebSocket). SVG card files, HTML pages, CSS, and JS all compress well. Scope carefully so WebSocket upgrade requests are unaffected.
