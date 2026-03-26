---
# hearts-u8s0
title: Request timeout middleware for API endpoints
status: todo
type: feature
priority: low
created_at: 2026-03-26T09:10:49Z
updated_at: 2026-03-26T09:10:49Z
parent: hearts-p6hh
---

Add middleware.Timeout scoped to the API route group only (not WebSocket endpoints). Prevents hung /api/tables or /api/debug/bots requests from blocking goroutines indefinitely.
