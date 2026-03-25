---
# hearts-fmia
title: ETag support for HTML pages
status: todo
type: task
priority: normal
created_at: 2026-03-25T06:56:12Z
updated_at: 2026-03-25T06:56:17Z
parent: hearts-e5b4
---

Add ETag and Cache-Control: no-cache to HTML routes for conditional-request support


## Context

HTML pages (`/` and `/table/{id}`) are served from embedded templates. Their content changes between builds but can't use URL fingerprinting since they're navigated to directly. Currently no conditional-request support exists, so browsers always download the full response.

## Higher Goal

Save bandwidth on page reloads and navigations by allowing browsers to validate cached HTML with a lightweight `If-None-Match` round-trip instead of re-downloading the full page (epic hearts-e5b4).

## Acceptance Criteria

- [ ] `/` and `/table/{id}` responses include an `ETag` header derived from content hash
- [ ] Responses include `Cache-Control: no-cache` (forces revalidation, allows caching)
- [ ] Requests with matching `If-None-Match` receive 304 Not Modified
- [ ] Test coverage for ETag generation and 304 responses

## Out of Scope

- Caching of API responses (`/api/tables`)
- WebSocket caching considerations
- HTML fingerprinting via URL rewriting
