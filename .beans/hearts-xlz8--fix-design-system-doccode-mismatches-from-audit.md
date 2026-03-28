---
# hearts-xlz8
title: Fix design system doc/code mismatches from audit
status: completed
type: task
priority: normal
created_at: 2026-03-28T16:14:49Z
updated_at: 2026-03-28T16:18:39Z
parent: hearts-dfll
---

Five minor mismatches found in hearts-pn5w audit. Group fix:

1. **Button gradient**: Unify table-page button gradient end color from `#1a8587` to `#1b8d8f` (styles.input.css line 999)
2. **`--anim-scoreboard-flip`**: Remove the unused CSS custom property (lines 317, 327) and its doc entry — or implement the animation. Needs decision.
3. **Icon button hover doc**: Change 'lifts with box-shadow' to 'gains box-shadow' in design-system.md
4. **Badge font size**: Widen small tier in doc from 0.78–0.88rem to 0.75–0.88rem
5. **Font weight 300**: Note the create-icon exception in design-system.md

## Summary of Changes

1. Unified table-page button gradient end color to `#1b8d8f` (was `#1a8587`)
2. Removed unused `--anim-scoreboard-flip` CSS custom property from both standard and fast-mode declarations, and from design-system.md
3. Changed icon button hover description from "lifts" to "gains box-shadow" in design-system.md
4. Widened small font-size tier from `0.78–0.88rem` to `0.75–0.88rem` in design-system.md
5. Added weight `300` (create icon) to typography weights in design-system.md
