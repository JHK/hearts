---
# hearts-akrr
title: Fix 2 failing webui integration tests
status: todo
type: bug
priority: high
created_at: 2026-03-24T14:41:33Z
updated_at: 2026-03-24T14:41:33Z
---

Two tests failing in `internal/webui/server_integration_test.go`:

1. **TestWebSocketJoinReusesPlayerByToken** (line 184) — reconnecting with the same token creates a new player (`p-2`) instead of reusing the existing one (`p-1`).
2. **TestTableClosesWhenLastHumanLeaves** (line 276) — table doesn't close within the 2s timeout after the last human disconnects.

All other packages pass (`game`, `bot`, `table`).
