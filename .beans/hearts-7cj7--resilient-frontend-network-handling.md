---
# hearts-7cj7
title: Resilient frontend network handling
status: todo
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-25T10:35:48Z
updated_at: 2026-03-25T10:35:54Z
parent: hearts-g7wu
---

Add try/catch, retries, and exponential backoff to frontend network operations so brief hiccups don't crash the UI or lose messages

## Context
The frontend has several network operations (WebSocket, fetch calls) with inconsistent or missing error handling. In hearts-2lkv we added initial WebSocket connect retries, but other operations still fail silently or crash on transient network issues.

## Higher Goal
Players on flaky connections (mobile, WiFi handoffs) should not lose their game or see broken UI states because of a brief network hiccup. Retries are invisible to the user — failures are logged to the console only.

## Acceptance Criteria
- [ ] `fetch /api/tables` (lobby polling) has try/catch and degrades gracefully on failure (shows stale data, logs to console, no crash)
- [ ] `POST /api/tables` (create table) retries up to 2 times on network failure before giving up; failures logged to console only
- [ ] WebSocket reconnection after disconnect uses exponential backoff instead of fixed 1s delay
- [ ] `send()` silently buffers or drops messages when WebSocket is not connected, with console logging instead of crashing
- [ ] No unhandled promise rejections from network operations in the browser console

## Out of Scope
- User-visible error banners or toast notifications
- Offline mode or service workers
- Server-side retry logic
- Persisting game state across server restarts (that's hearts-oeb4)
