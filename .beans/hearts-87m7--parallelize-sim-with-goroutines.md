---
# hearts-87m7
title: Parallelize sim with goroutines
status: todo
type: task
priority: normal
created_at: 2026-03-24T13:11:27Z
updated_at: 2026-03-24T13:11:41Z
---

Use a worker pool of runtime.NumCPU() goroutines to run sim games in parallel


## Context

`cmd/sim` runs all game iterations sequentially in a single goroutine. On a 50 000-iteration run this leaves all but one CPU core idle. The game logic in `internal/game/` is pure and stateless per game, so individual games are embarrassingly parallel — the only shared state is the final `Result` struct (two `[4]int` arrays).

## Higher Goal

Faster feedback loop when tuning bot strategies or validating rule changes.

## Acceptance Criteria

- [ ] Games run concurrently using a worker-pool of `runtime.NumCPU()` goroutines
- [ ] Each worker uses its own `*rand.Rand` (seeded independently) — no shared RNG
- [ ] Results are aggregated without data races (`go test -race` passes)
- [ ] `time go run ./cmd/sim -n 50000` is measurably faster than before (run before/after, include timings in PR summary)
- [ ] Existing tests pass (`mise run test`)

## Out of Scope

- Changing bot logic or game rules
- Making the number of worker goroutines configurable (just use `runtime.NumCPU()`)
- Deterministic / reproducible output (each run already varies by seed)
