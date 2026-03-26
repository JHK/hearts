---
# hearts-aazx
title: Decompose NewHandler into per-concern route registration files
status: todo
type: task
priority: low
created_at: 2026-03-26T09:10:56Z
updated_at: 2026-03-26T09:10:56Z
parent: hearts-p6hh
---

Split the monolithic NewHandler() into separate registration functions per route group (pages, API, WebSocket, assets, dev) in their own files. Keeps server.go focused on wiring and makes each concern independently navigable.
