---
# hearts-u8s0
title: Request timeout middleware for API endpoints
status: completed
type: feature
priority: low
created_at: 2026-03-26T09:10:49Z
updated_at: 2026-03-26T12:28:12Z
parent: hearts-p6hh
---

Add middleware.Timeout scoped to the API route group only (not WebSocket endpoints). Prevents hung /api/tables or /api/debug/bots requests from blocking goroutines indefinitely.

## Summary of Changes

Added `middleware.Timeout(10s)` to the `/api` route group in `routes_api.go`. This prevents hung API requests (`/api/tables`, `/api/debug/bots`) from blocking goroutines indefinitely, while leaving WebSocket endpoints unaffected.
