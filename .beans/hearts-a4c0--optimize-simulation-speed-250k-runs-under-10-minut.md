---
# hearts-a4c0
title: Optimize simulation speed (250k runs under 10 minutes)
status: todo
type: task
priority: high
created_at: 2026-03-27T12:41:03Z
updated_at: 2026-03-27T12:41:20Z
---

250k sim runs take >1h; profile to confirm MC is the bottleneck, then reduce MC samples/threshold for sim runs while keeping live gameplay parameters unchanged

## Context

Running 250k games with `./sim -n 250000` currently takes over 1 hour. The bottleneck is almost certainly the Monte Carlo evaluator (`mc.go`): it runs 50 samples per candidate play, triggers when hand ≤ 7 cards (or ≤ 9 near game-over), meaning it fires for roughly the last 6–8 tricks of every round, for every Hard bot, across every game.

The MC parameters (50 samples, threshold 7/9) are tuned for live gameplay quality. In sim mode, we already know those values produce a better win rate — we don't need the same fidelity per decision to get statistically valid results.

## Higher Goal

Sim runs are the feedback loop for bot strategy development. A >1h cycle kills iteration speed. Bringing 250k runs under 10 minutes makes it practical to run sims after every change.

## Acceptance Criteria

- [ ] Profile with small batch sizes (e.g. 1k, 5k) to confirm MC is the bottleneck and measure per-game cost
- [ ] If MC is the bottleneck: reduce `defaultMCSamples` and/or `mcThreshold` for sim runs (e.g. via a sim-specific bot config or a lower sample count passed through)
- [ ] 250k runs complete in ≤ 10 minutes on the dev machine
- [ ] Live gameplay MC parameters remain unchanged (sim-only reduction)
- [ ] `strategies.md` updated if MC sim parameters are documented

## Out of Scope

- Improving MC algorithm quality (smarter sampling, pruning)
- Changing the sim worker pool / concurrency model (it already uses all CPUs)
- Benchmarking MC decision quality at reduced samples (we accept the trade-off)
