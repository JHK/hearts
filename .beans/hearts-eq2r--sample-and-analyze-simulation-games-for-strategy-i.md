---
# hearts-eq2r
title: Sample and analyze simulation games for strategy improvements
status: todo
type: task
priority: normal
created_at: 2026-03-27T09:48:18Z
updated_at: 2026-03-27T09:54:06Z
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
- [ ] Add game sampling to `cmd/sim` — capture full game logs (hands dealt, passes, plays, tricks, scores) for a configurable subset of games
- [ ] Output sampled games in a format suitable for analysis (e.g. JSON or structured text)

### Phase 2: Analysis
- [ ] Analyze sampled games where the hard bot lost — identify patterns: what went wrong, what could have been played differently
- [ ] Look for cases where the hard bot made clearly suboptimal plays (e.g. dumping points unnecessarily, failing to block a moon shot, poor pass choices)
- [ ] Check for implementation bugs where the bot's behavior diverges from what `strategies.md` describes

### Phase 3: Follow-up
- [ ] Create follow-up beans for each concrete improvement identified, with clear descriptions of the problem and proposed fix
- [ ] Each follow-up should be individually measurable via the sim validation approach described above

### Important: Follow-up bean instructions
When creating follow-up beans under this epic, each one MUST include the instruction to follow the Validation section in the parent epic (`hearts-8j8z`). Copy the approach rules (baseline, 50k measurement, sequential runs, revert threshold) into every new bean you create.
