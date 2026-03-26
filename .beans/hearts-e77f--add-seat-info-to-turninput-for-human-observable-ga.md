---
# hearts-e77f
title: Add seat info to TurnInput for human-observable game state
status: completed
type: task
created_at: 2026-03-26T06:41:09Z
updated_at: 2026-03-26T06:41:09Z
---

Change TurnInput.Trick and TurnInput.PlayedCards from []Card to []Play so bots (and future UIs) can see who played each card — matching what a human player can observe and remember. Add CardsFrom helper and TrickCards/PlayedCardsList convenience methods for backward-compatible card-only access.
