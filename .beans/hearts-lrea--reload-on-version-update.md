---
# hearts-lrea
title: Reload on version update
status: draft
type: task
priority: normal
created_at: 2026-04-08T13:28:17Z
updated_at: 2026-04-08T13:28:17Z
---

When a new version gets released players should reload their browser. Note that the reload should only happen if the server version is newer than the browser one. Communication of that version can happen via websocket.
A time based UUID injected at build time or if missing (in development) at runtime could be used, so we don't rely on git commits or git tags. This should be validated in ticket refinement or even research.
