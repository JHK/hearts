---
# hearts-04o8
title: Remove redundant bot delay timers in scheduleBotTurn/scheduleBotPass
status: completed
type: task
priority: normal
created_at: 2026-03-24T17:43:40Z
updated_at: 2026-03-24T17:54:50Z
---

\`scheduleBotTurn\` and \`scheduleBotPass\` in \`internal/session/table.go\` each spin up a goroutine with a 300-350ms timer and a \`r.stop\` select guard before calling \`r.submit\`. Both are unnecessary:

1. **Delay**: The frontend already queues trick events with animations (card transitions, trick-winner stagger), so bots don't appear instant to humans regardless of backend timing.
2. **Shutdown guard**: \`r.submit\` itself already selects on \`r.stop\`, so the goroutine-level guard is redundant.

Replace both functions' goroutine bodies with a bare \`go r.submit(...)\`.

## Summary of Changes

Removed redundant bot delay timers and shutdown guards from `scheduleBotTurn` and `scheduleBotPass` in `internal/session/table.go`. Both goroutine bodies are now a bare `go r.submit(...)` since `submit` already selects on `r.stop` and the frontend handles animation timing.
