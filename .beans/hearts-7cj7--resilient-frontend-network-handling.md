---
# hearts-7cj7
title: Resilient frontend network handling
status: completed
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-25T10:35:48Z
updated_at: 2026-03-25T10:42:55Z
parent: hearts-g7wu
---

Add try/catch, retries, and exponential backoff to frontend network operations so brief hiccups don't crash the UI or lose messages

## Context
The frontend has several network operations (WebSocket, fetch calls) with inconsistent or missing error handling. In hearts-2lkv we added initial WebSocket connect retries, but other operations still fail silently or crash on transient network issues.

## Higher Goal
Players on flaky connections (mobile, WiFi handoffs) should not lose their game or see broken UI states because of a brief network hiccup. Retries are invisible to the user — failures are logged to the console only.

## Acceptance Criteria
- [x] `fetch /api/tables` (lobby polling) has try/catch and degrades gracefully on failure (shows stale data, logs to console, no crash)
- [x] `POST /api/tables` (create table) retries up to 2 times on network failure before giving up; failures logged to console only
- [x] WebSocket reconnection after disconnect uses exponential backoff instead of fixed 1s delay
- [x] `send()` silently buffers or drops messages when WebSocket is not connected, with console logging instead of crashing
- [x] No unhandled promise rejections from network operations in the browser console

## Out of Scope
- User-visible error banners or toast notifications
- Offline mode or service workers
- Server-side retry logic
- Persisting game state across server restarts (that's hearts-oeb4)

## Summary of Changes

Added resilient error handling to all frontend network operations:

- **Lobby polling** (`fetchTables`): wrapped in try/catch, shows stale data on failure
- **Table creation** (`createTable`): retries up to 2 times on network failure with console logging
- **WebSocket reconnect**: exponential backoff (1s, 2s, 4s... up to 30s) with jitter, instead of fixed 1s
- **`send()`**: logs dropped messages to console when WebSocket is not connected
- **`onmessage`**: wrapped `JSON.parse` in try/catch to prevent unhandled exceptions
