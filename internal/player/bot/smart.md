# Smart Bot Strategy

## Current Strategy Overview

The smart bot plays a defensive Hearts strategy: avoid penalty cards (hearts and Q♠), dump high-risk cards when void in the led suit, and opportunistically attempt moon shots when holding enough guaranteed tricks.

### Passing

Cards are scored by `passRisk` (shared with dumb bot). The three highest-risk cards are passed. Key heuristics:
- Q♠ gets a massive penalty (+240) — pass it immediately if held
- A♠/K♠ get elevated penalties (+120) — high chance of being forced to take Q♠
- Short suits get bonuses (+45 for singleton, +25 for doubleton) — being void lets you discard freely
- Low clubs/diamonds get discounts — safe to hold, low threat

### Leading

Pool is restricted to non-hearts cards unless only hearts remain. Within the pool:
1. **Void suits are avoided** — if opponents have shown voids (by discarding off-suit in prior tricks), don't lead that suit since they'll dump penalty cards on you.
2. **Safe high cards are preferred** — if you hold the highest remaining card in a suit (all higher cards played or in your hand), lead it: you control the trick.
3. **Crowded suits otherwise** — prefer suits where opponents still hold many cards (`opponentCardCount >= 2`), to avoid wasting leads.
4. **Lowest card in chosen suit** — minimises risk of accidentally winning.

Tiebreaking follows `compareLeadCard`: danger penalty (spades with Q♠ in hand), then suit length, then rank, then suit priority.

### Following Suit

When you have the led suit:
1. **Moon-shot active** — play highest card to win every trick.
2. **Penalty in trick** — avoid taking it:
   - If you have cards that lose (under), play the highest under that is NOT a guaranteed safe-high-card (prefer to shed unsafe high cards). Q♠ is always preferred within unsafe candidates — dump it onto whoever is winning.
   - If forced to win (no under cards), prefer lowest non-penalty card; only play a penalty card as last resort.
3. **Clean trick, last to play** — win cheaply: lowest non-penalty card that beats the winner. Avoid playing Q♠ when a cheaper winning card exists.
4. **Not last to play, all cards beat winner** — prefer lowest non-penalty card; Q♠ avoided when alternatives exist.
5. **Clean trick, not last** — play highest unsafe-under card (shed dangerous cards while avoiding accidental wins).

### Discarding (Off-Suit)

Priority: Q♠ first → any penalty card (by `highestRiskCard`) → high spades → highest remaining card.

### Moon-Shot State Machine

Activation condition: `guaranteedHeartTricks(hand) >= 3 && guaranteedTricks(hand) >= 8`.

"Guaranteed tricks" = consecutive top-card sequences from Ace down (e.g., A♥K♥Q♥ = 3 guaranteed heart tricks). "Guaranteed total tricks" sums across all suits.

When active:
- **Lead**: always highest card in hand.
- **Follow**: always highest card to win tricks.
- **Discard**: highest card to accumulate penalties onto self.

**Dynamic re-activation**: if moon shot was aborted but at some later point `allSafeHighCards >= remainingTricks`, re-activate (late-game opportunity).

**Abort**: if an opponent wins a trick while moon shot is active, abort and switch back to defensive play.

---

## What Was Attempted

### Baseline
- Smart beat dumb **54.6%** head-to-head (200k games).

### Q♠ Avoidance in Follow (+1.8pp total)

Three bugs where `lowestRankedCard` accidentally played Q♠ (rank 12) over a cheaper winning card:

| Bug | Fix | Delta |
|-----|-----|-------|
| Penalty trick, no under cards → played Q♠ | Prefer non-penalty legal cards | +0.6pp |
| Last to play clean trick, over=[Q♠,K♠] → played Q♠ | Prefer non-penalty overs | +0.7pp |
| All cards over winner, not last → played Q♠ | Prefer non-penalty overs | +0.5pp |

Q♠ in `highestUnsafeUnder`: when under=[Q♠,K♠] both unsafe, `highestRankedCard` returned K♠ (rank 13), missing the chance to offload Q♠ (13 pts) onto the trick winner. Fixed: always prefer Q♠ within unsafe candidates.

**Result: 56.4%**

### Crowded-Suit Threshold Tuning (neutral)

`opponentCardCount >= N` swept over {1, 2, 3, 4, 5}. Threshold=2 gave highest measured result at 50k (56.2%) vs default 4. Confirmed combined result stable at 56.4% — within noise of the Q♠ fixes.

### Moon-Shot Threshold Tuning (neutral/marginal)

| Threshold | Result |
|-----------|--------|
| >=2 hearts, >=7 total | 55.9% (no gain) |
| >=3 hearts, >=8 total (current) | 56.4% |
| >=3 hearts, >=9 total | 56.1% (slight loss) |
| >=4 hearts, >=10 total | 55.8% (no gain) |

### Near-Moon-Shot (HARMFUL, −3.3pp)

Activate moon shot when `allSafeHighCards >= remainingTricks - 1` (one trick short of guaranteed). Result: **53.1%**. The bot was too aggressive, committing to moon shots it couldn't complete, gifting opponents 26 pts instead of blocking the moon.

### A♠/K♠ Retention in Pass (HARMFUL, −1.1pp)

When not holding Q♠, retain A♠ and K♠ (don't pass them — they're "safe" if you don't have Q♠). Backfired: these high spades still forced costly trick wins. Result: **54.8%**.

### Leading Threshold Conservative: rank < 10 (SLIGHT LOSS)

Restrict safe pool to rank < 10 (exclude Jacks/Tens from leading candidates). Meant to be more conservative about leading high cards. Result: **55.6%** (−0.8pp).

### `lowestRankedCard(under)` in Follow (CATASTROPHIC, −17pp)

Replace `highestUnsafeUnder` with `lowestRankedCard(under)` everywhere. The bot stopped shedding its own dangerous high cards, accumulating penalty points. Result: **39.2%**.

### `lowestRankedCard(under)` in Not-Last, No-Penalty Path Only (CATASTROPHIC, −13pp)

Same experiment scoped to just the "not last to play, clean trick" path. Still catastrophic because this is the most common follow scenario. Result: **43.3%**.
