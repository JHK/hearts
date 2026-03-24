---
# hearts-akrr
title: Fix 2 failing webui integration tests
status: completed
type: bug
priority: high
created_at: 2026-03-24T14:41:33Z
updated_at: 2026-03-24T14:49:20Z
---

Two tests failing in `internal/webui/server_integration_test.go`:

1. **TestWebSocketJoinReusesPlayerByToken** (line 184) — reconnecting with the same token creates a new player (`p-2`) instead of reusing the existing one (`p-1`).
2. **TestTableClosesWhenLastHumanLeaves** (line 276) — table doesn't close within the 2s timeout after the last human disconnects.

All other packages pass (`game`, `bot`, `table`).

## Summary of Changes

Two fixes in the webui integration tests:

1. **Token reuse on reconnect**: Added `departedTokens` map to `tableState` that remembers player IDs from pre-round leavers. When a player rejoins with the same token, their original player ID is reused instead of generating a new one.

2. **Table close on empty**: Reduced the orphaned-table grace period from 60s to 500ms for tables where no round was ever started. Tables with active/completed rounds still get the full 60s grace period for reconnection.
