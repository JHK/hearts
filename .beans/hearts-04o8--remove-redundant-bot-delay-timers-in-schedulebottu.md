---
# hearts-04o8
title: Remove redundant bot delay timers in scheduleBotTurn/scheduleBotPass
status: todo
type: task
priority: normal
created_at: 2026-03-24T17:43:40Z
updated_at: 2026-03-24T17:48:02Z
---

\`scheduleBotTurn\` and \`scheduleBotPass\` in \`internal/session/table.go\` each spin up a goroutine with a 300-350ms timer and a \`r.stop\` select guard before calling \`r.submit\`. Both are unnecessary:

1. **Delay**: The frontend already queues trick events with animations (card transitions, trick-winner stagger), so bots don't appear instant to humans regardless of backend timing.
2. **Shutdown guard**: \`r.submit\` itself already selects on \`r.stop\`, so the goroutine-level guard is redundant.

Replace both functions' goroutine bodies with a bare \`go r.submit(...)\`.
