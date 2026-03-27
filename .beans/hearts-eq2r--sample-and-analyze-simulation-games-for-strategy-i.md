---
# hearts-eq2r
title: Sample and analyze simulation games for strategy improvements
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:48:18Z
updated_at: 2026-03-27T13:04:10Z
parent: hearts-8j8z
---

Add the ability to sample individual games from `cmd/sim` runs, then analyze those samples to identify strategy weaknesses, potential improvements, and implementation bugs. Use findings to create follow-up tasks with concrete improvements.

## Approach

**You MUST follow the Validation section in the parent epic (`hearts-8j8z`).** Read it before starting. The key rules:
- Establish a baseline with a single 250k run first.
- Measure each change with a 50k run. An improvement must be ≥1pp to keep.
- Run simulations sequentially — never in parallel. Wait for each run to complete before proceeding.
- If a change does not meet the threshold, revert it and try something else.

### Phase 1: Sampling infrastructure
- [x] Add game sampling to `cmd/sim` — capture full game logs (hands dealt, passes, plays, tricks, scores) for a configurable subset of games
- [x] Output sampled games in a format suitable for analysis (e.g. JSON or structured text)

### Phase 2: Analysis
- [x] Analyze sampled games where the hard bot lost — identify patterns: what went wrong, what could have been played differently
- [x] Look for cases where the hard bot made clearly suboptimal plays (e.g. dumping points unnecessarily, failing to block a moon shot, poor pass choices)
- [x] Check for implementation bugs where the bot's behavior diverges from what `strategies.md` describes

### Phase 3: Follow-up
- [x] Create follow-up beans for each concrete improvement identified (hearts-v1pc, hearts-f7ek, hearts-573t), with clear descriptions of the problem and proposed fix
- [x] Each follow-up should be individually measurable via the sim validation approach described above

### Important: Follow-up bean instructions
When creating follow-up beans under this epic, each one MUST include the instruction to follow the Validation section in the parent epic (`hearts-8j8z`). Copy the approach rules (baseline, 50k measurement, sequential runs, revert threshold) into every new bean you create.

## Summary of Changes

### Baseline (250k games, MCSamples=3)
- Hard: 36.7%, Medium: 36.5%, Easy: 28.4%, Random: 1.2%

### Sampling infrastructure
- Added `internal/sim/sample.go` with `GameLog`, `RoundLog`, `TrickLog` types capturing full game state
- Added `RunWithSamples()` method to `Simulation` for capturing game logs during sim runs
- Added `-sample` and `-sample-file` flags to `cmd/sim`
- Added `cmd/analyze` tool for automated analysis of sampled game logs

### Analysis findings (1000-game sample)
- No implementation bugs found — hard bot behavior matches `strategies.md`
- **Failed moon shots**: 202 instances (20-25 raw pts) in 650 losses — largest catastrophic round source
- **Unblocked random bot moon shots**: 158 instances — detection thresholds too high for accidental shooters  
- **QS dump vulnerability**: 766 of 886 QS catches were dumps (not led) — 16.8% of all QS tricks caught by hard bot
- **Comeback losses**: 365 of 650 losses were games where hard bot led at some point

### Follow-up beans created
- hearts-v1pc: Tighten moon-shot abort logic
- hearts-f7ek: Improve moon-shot blocking against accidental shooters
- hearts-573t: Reduce QS dump vulnerability when following spades
