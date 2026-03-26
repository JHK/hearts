# Bot Strategies Reference

This document describes every heuristic and decision rule used by the bot strategy tiers. Update it whenever bot logic changes.

## Strategy Tiers

| Tier | Struct | State | Description |
|---|---|---|---|
| Easy | `Easy` | Stateless | Rule-based defensive play. No moon-shot pursuit. No card counting beyond suit lengths. |
| Medium | `Medium` | Stateful | Adds moon-shot pursuit/detection, card counting, safe-high-card tracking, void detection. |
| Hard | `Hard` | Stateful | Currently identical structure to Medium, with a relaxed moon-shot threshold and improved moon-shot pass/lead/follow logic. |
| Random | `Random` | Stateless | Picks a uniformly random legal play. Baseline for benchmarks. |
| FirstLegal | `FirstLegal` | Stateless | Picks the first legal play. Testing only. |

## Decision Context

Bots receive two input types from the game engine:

- **`PassInput`**: hand (13 cards), pass direction (left / right / across / hold).
- **`TurnInput`**: hand, current trick cards, played cards (all completed tricks), round points per seat, seat index, hearts-broken flag, first-trick flag.

Bots have **no access** to: cumulative game scores, other players' hands, or what was passed to/from them in previous rounds.

---

## Passing Strategies

### Defensive Pass (Easy, Medium, Hard)

Used when neither moon-shot nor void-creation is selected.

1. Score each card with `passRisk(card, suitCounts)`:
   - Base: `rank * 4`
   - Q♠: +240 (extreme danger — 13 penalty points if taken)
   - K♠ / A♠: +120 (can force you to take Q♠)
   - High spades (J♠+): +30
   - Hearts: +20 + rank * 2
   - Singleton suit: +45 (opponents can void you)
   - Doubleton suit: +25
   - Low clubs (2-4): -25 (safe to keep, useful for opening trick)
   - Low diamonds (2-4): -10 (safe to keep)
2. Sort descending by risk; pass the top 3.

### Void-Creation Pass (Medium, Hard)

Triggered when a non-heart suit has only 1-2 cards.

1. Find the shortest non-heart suit (1-2 cards).
2. Pass all cards from that suit.
3. Fill remaining slots with highest-risk cards (using `passRisk`).

Rationale: a void lets you discard penalty cards whenever that suit is led.

### Moon-Shot Pass (Medium)

Triggered when `evaluateMoonShot(hand)` returns true (requires 3+ guaranteed heart tricks AND 8+ total guaranteed tricks).

1. Score each card with `moonShotSupport(card, hand)`:
   - Base: `rank * 2`
   - Card in consecutive top-card run for its suit: +50 (backbone of trick control)
2. Sort ascending by support score; pass the 3 least-supportive cards.

### Moon-Shot Pass — Hard Variant

Same trigger as Medium but with relaxed threshold: 2+ guaranteed hearts with 7+ total guaranteed tricks also qualifies (EV analysis shows moon-shot is profitable at ~31% success rate).

Additional logic:
1. Same `moonShotSupport` scoring as Medium.
2. **Void-suit bonus**: if an off-suit (clubs, diamonds, spades) has all non-run cards and count <= 3, those cards get -40 to their support score, making them more likely to be passed. This creates a void for free discards during play.

---

## Round Strategy Selection

### Medium

```
if guaranteedHeartTricks >= 3 AND guaranteedTricks >= 8 → moonShot
else if any non-heart suit has 1-2 cards              → voidCreation
else                                                   → defensive
```

### Hard

```
if guaranteedHeartTricks >= 3 AND guaranteedTricks >= 8 → moonShot  (standard)
if guaranteedHeartTricks >= 2 AND guaranteedTricks >= 7 → moonShot  (relaxed)
else if any non-heart suit has 1-2 cards                → voidCreation
else                                                     → defensive
```

### Guaranteed Trick Counting

A "guaranteed trick" is a card in a consecutive run from the top of its suit. Example: holding A♠ K♠ Q♠ gives 3 guaranteed spade tricks (each beats all remaining cards). Holding A♠ Q♠ gives only 1 (A♠), because K♠ is missing and could beat Q♠.

---

## Leading Strategies

### Easy: `chooseSmartLead`

1. Prefer non-hearts.
2. Among the pool, pick the card with the lowest **lead danger** (`leadDangerPenalty`):
   - Spades while holding Q♠: penalty increases with spade rank and decreases if only 1-2 spades remain.
3. Tiebreaker: shortest suit first, then lowest rank.

### Medium/Hard: `smartChooseLead`

**When pursuing moon-shot**: lead the highest safe-high-card (guaranteed winner).

**Defensive leading** (layered filtering):

1. **Prefer non-hearts.**
2. **Avoid void suits**: filter out suits where an opponent has been observed discarding off-suit (they'll dump penalties).
3. **Prefer crowded suits**: suits where opponents still hold 2+ cards (`opponentCardCount`). Opponents with cards must follow suit, preventing penalty dumps.
4. **Prefer low non-winners**: cards with rank < 11 that are NOT safe-high-cards. Guaranteed winners invite Q♠ dumps; high uncertain cards risk uncontrolled trick-taking.
5. **Prefer very low cards**: rank < 9. If none exist in the filtered pool, widen to the full legal set (a low heart beats leading K♠).
6. **Tiebreaker** (`compareDefensiveLeadCard`): shortest suit, then **highest** rank (shed borderline cards early while opponents still hold higher ones), with Q♠ danger penalty.

### Hard Moon-Shot Lead: `hardMoonShotLead`

1. Filter to safe-high-cards (guaranteed winners).
2. Among non-heart safe leads, prefer the suit where the bot holds the **most** cards (run the longest suit first to drain opponents' holdings and create opponent voids that yield free hearts).
3. If only heart safe leads remain, play the highest.
4. Fallback: highest legal card.

---

## Following Strategies

### Easy: `chooseSmartFollow`

1. **Penalty in trick**: play highest under-card (duck as high as safely possible). If forced to take, play lowest card.
2. **Last to play, no penalty**: win cheaply (lowest over-card).
3. **Default**: play highest under-card (duck high).

### Medium/Hard: `smartChooseFollow`

**When pursuing moon-shot**: win cheaply (lowest over-card). If can't win, preserve high cards (lowest under-card).

**Defensive — penalty in trick**:
1. Play highest *unsafe* under-card (`highestUnsafeUnder`):
   - Filter to under-cards that are NOT safe-high-cards.
   - Prefer Q♠ among unsafe cards (shed the 13-point liability onto the trick winner).
   - If all under-cards are safe, shed the highest (safe high cards are future trick-winning liabilities).
2. If forced to take (no under-cards): prefer non-penalty cards so Q♠ isn't added. Then lowest card.

**Defensive — no penalty, last to play**:
- Win cheaply with lowest non-penalty over-card (avoid playing Q♠ as a "cheap win" when K♠/A♠ exists).

**Defensive — no penalty, not last**:
- Shed highest unsafe under-card (same logic as penalty case, including Q♠ priority).
- If all cards beat the winner: play lowest non-penalty card (avoid Q♠ when unnecessary).

### Hard Moon-Shot Follow: `hardMoonShotFollow`

1. If can win and penalty cards are in the trick and players remain after us: play **highest** over-card (maximize chance of winning — a lost penalty trick kills the moon-shot).
2. Otherwise if can win: win cheaply (lowest over-card).
3. If can't win: preserve high cards (lowest legal card).

---

## Discarding Strategies (Void in Led Suit)

### Easy: `chooseSmartDiscard`

Priority order:
1. Q♠ (13 points, highest single-card risk)
2. Highest-risk penalty card (hearts, by `discardRisk`)
3. Highest-risk spade
4. Highest-risk card overall

### Medium/Hard: `smartChooseDiscard`

**When pursuing moon-shot**: discard lowest non-penalty card (preserve high hearts/Q♠ for collecting).

**Defensive**: same priority as Easy (Q♠ → highest penalty → highest spade → highest risk).

### Discard Risk Scoring

`discardRisk(card)`:
- Base: `rank * 4`
- Hearts: +40
- Spades: +20, and K♠/A♠ get an additional +20

---

## Moon-Shot State Machine (Medium/Hard)

The moon-shot is tracked across tricks with two flags: `moonShotActive` and `moonShotAborted`.

### Lifecycle

1. **Pass phase**: evaluate hand. If qualified, set `moonShotActive = true`.
2. **Trick 1 (post-pass)**: re-evaluate with actual post-pass hand. Received cards may complete a qualifying run.
3. **Hold round start**: re-evaluate since `ChoosePass` wasn't called.
4. **Each subsequent trick**:
   - **Medium**: abort if someone else leads (they won the last trick, meaning we lost control).
   - **Hard**: abort only if we've lost penalty points to other players. If we still hold all scored penalties, keep pursuing — we may regain control.

### Dynamic Activation (Medium/Hard)

At each lead opportunity, if the bot holds safe-high-cards for ALL remaining tricks, activate moon-shot regardless of prior state. This covers late-game scenarios where early tricks were messy but the bot now controls everything.

### Hard-Only: Soft Re-activation

If one safe-high-card short of full coverage, count "near-safe" cards (exactly one higher card unaccounted for). If any exist, activate — the odds favor winning even with one gap.

### Self-Abort

If pursuing and no safe-high-cards exist anywhere in hand, abort.

---

## Card Analysis Helpers

### Safe High Card (`isSafeHighCard`)

A card is "safe" (guaranteed to win its suit) if every card of higher rank in that suit is either already played or in the bot's hand. Used by Medium/Hard for leading, following, and moon-shot evaluation.

### Void Detection (`detectSuitVoids`)

Scans completed tricks and the current in-progress trick. When a player follows with an off-suit card, that suit is marked as having a void. Used by Medium/Hard to avoid leading void suits (opponents will dump penalties).

### Opponent Card Count (`opponentCardCount`)

Estimates cards opponents still hold in a suit: `13 - inHand - played`. Used to prefer "crowded" suits when leading (opponents must follow, preventing penalty dumps).

### Guaranteed Trick Counting

- `guaranteedHeartTricks`: consecutive top-card run in hearts (A♥, K♥, Q♥...).
- `guaranteedNonHeartTricks`: sum of consecutive top-card runs across clubs, diamonds, spades.
- `guaranteedTricks`: heart + non-heart total.

### Near-Safe Cards (Hard only)

`countNearSafeCards`: cards where exactly one higher card in their suit is unaccounted for. Used for soft re-activation of moon-shot with marginal hands.

---

## Strategies Not Yet Implemented

- **Opponent moon-shot detection**: no mechanism to notice an opponent collecting all penalties and save a high heart to block.
- **Spade flushing**: no deliberate low-spade leads to smoke out Q♠.
- **Pass direction weighting**: `passRisk` ignores whether passing left (dangerous) or right (safer).
- **Cumulative score awareness**: bots optimize each round independently; no adjustment for leading/trailing.
- **Monte Carlo evaluation**: all decisions are heuristic; no sampling of opponent hands to compare outcomes.
- **Suit establishment**: no deliberate strategy of leading low from long suits to establish later winners.
- **Cooperative play**: bots act independently; no implicit coordination against a shooter or leader.

---

## References

Strategy research that informed the current implementation and planned improvements.

### Competitive Play Guides

- [Mark's Advanced Hearts Strategies](https://mark.random-article.com/hearts/advanced.html): spade bleeding sequences (lead A♠ → K♠ → J♠), counting tricks by suit cycles, contractual passing tactics, moon-shot material passing to trailing players.
- [Wikibooks — Card Games/Hearts/Strategy](https://en.wikibooks.org/wiki/Card_Games/Hearts/Strategy): never short-suit spades below Q♠, void clubs/diamonds first, keep low spades to control Q♠, hold A♥ to block moon-shots, save 2-3 strong cards for final tricks.
- [247 Solitaire Hearts Strategy Guide](https://www.247solitaire.com/news/hearts-card-game-strategy-guide/): flush the Queen early with low spade leads, void clubs or diamonds first, pass direction effects.
- [VIP Hearts Tips](https://viphearts.com/blog/hearts-tips/): track who's close to winning, adjust strategy based on game standings, long-game score awareness.
- [Solitaired — Hearts Strategies and Tips](https://solitaired.com/guides/hearts-strategies-and-tips-to-win): observation and adaptation, detecting opponent void suits, long-game strategy.
- [Hearts.co — Most Effective Strategies](https://hearts.co/hearts-strategy): voiding suits, Q♠ management, moon-shot execution.
- [MobilityWare Hearts — Advanced Strategy](https://mobilityware.helpshift.com/hc/en/42-hearts-card-game/faq/2629-advanced-strategy/): first-trick tactics, pass disruption, observation signals.
- [MobilityWare Hearts — Shoot the Moon](https://mobilityware.helpshift.com/hc/en/42-hearts-card-game/faq/2628-what-s-a-good-shoot-the-moon-strategy/): detection signals (opponent passes low cards, leads high repeatedly), blocking tactics.
- [Arkadium — Hearts Strategy](https://www.arkadium.com/blog/hearts-card-game-strategy/): defensive fundamentals, card management priorities.
- [PokerNews — Hearts Strategy Tips](https://www.pokernews.com/card-games/hearts/hearts-strategy-tips.htm): passing heuristics, trick avoidance patterns.

### AI and Algorithm Research

- [Moving AI Lab — Hearts](https://www.movingai.com/hearts.html): Monte Carlo sampling with UCT algorithm; their bot averaged 55.8 points/game vs. 75.1 for Hearts Deluxe. Key finding: rule-based "strong heuristic" agents can be competitive with MC approaches. Opponent modeling identified as the next improvement area.
- [Sturtevant & White — Feature Construction for RL in Hearts](https://webdocs.cs.ualberta.ca/~nathanst/papers/heartslearning.pdf): feature engineering for reinforcement learning, Monte Carlo learning with multilayer perceptron. Hearts is harder to learn than perfect-information games due to hidden state.
- [HeartsAI Framework (GitHub)](https://github.com/Devking/HeartsAI): open-source framework for Hearts AI development and algorithm comparison.
- [Stanford CS230 — Reinforcement Learning on Hearts](http://cs230.stanford.edu/projects_spring_2021/reports/9.pdf): RL approaches to Hearts, policy gradient methods.
- [World of Card Games — When AI Learned to Play Hearts](https://worldofcardgames.com/blog/2025/07/when-ai-learned-to-play-hearts-study): study using real game data; strong heuristic agents outperformed MCTS, with void-creation being the most impactful single strategy.
