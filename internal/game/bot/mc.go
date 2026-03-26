package bot

import (
	"math"
	"math/rand"

	"github.com/JHK/hearts/internal/game"
)

const defaultMCSamples = 20

// mcEvaluator uses Monte Carlo hand sampling to evaluate candidate plays.
// For each legal play it samples hypothetical opponent hands consistent with
// known voids and played cards, simulates remaining tricks using easy-bot
// heuristics, and picks the move that minimizes expected penalty points.
type mcEvaluator struct {
	rng     *rand.Rand
	samples int
}

func newMCEvaluator(samples int) *mcEvaluator {
	return &mcEvaluator{
		rng:     rand.New(rand.NewSource(rand.Int63())),
		samples: samples,
	}
}

// evaluate returns the legal play that minimizes expected adjusted penalty
// points across sampled opponent hands. Caller must ensure len(legal) >= 1.
func (mc *mcEvaluator) evaluate(input game.TurnInput, legal []game.Card) game.Card {
	if len(legal) <= 1 {
		return legal[0]
	}

	voids := detectSeatVoids(input.PlayedCards, input.Trick)

	known := make(map[game.Card]bool, len(input.Hand)+len(input.PlayedCards)+len(input.Trick))
	for _, c := range input.Hand {
		known[c] = true
	}
	for _, p := range input.PlayedCards {
		known[p.Card] = true
	}
	for _, p := range input.Trick {
		known[p.Card] = true
	}

	remaining := make([]game.Card, 0, 52-len(known))
	for _, c := range game.BuildDeck() {
		if !known[c] {
			remaining = append(remaining, c)
		}
	}

	var handSizes [game.PlayersPerTable]int
	var playedBySeat [game.PlayersPerTable]int
	for _, p := range input.PlayedCards {
		playedBySeat[p.Seat]++
	}
	for _, p := range input.Trick {
		playedBySeat[p.Seat]++
	}
	for seat := range game.PlayersPerTable {
		if seat != input.MySeat {
			handSizes[seat] = 13 - playedBySeat[seat]
		}
	}

	trickNumber := len(input.PlayedCards) / game.PlayersPerTable

	bestCard := legal[0]
	bestAvg := math.MaxFloat64

	for _, candidate := range legal {
		total := 0.0
		for range mc.samples {
			hands, ok := sampleOpponentHands(mc.rng, input.MySeat, remaining, handSizes, voids)
			if !ok {
				var noVoids [game.PlayersPerTable]map[game.Suit]bool
				hands, _ = sampleOpponentHands(mc.rng, input.MySeat, remaining, handSizes, noVoids)
			}

			myHand := make([]game.Card, len(input.Hand))
			copy(myHand, input.Hand)
			myHand, _ = game.RemoveCard(myHand, candidate)
			hands[input.MySeat] = myHand

			trick := make([]game.Play, len(input.Trick), len(input.Trick)+1)
			copy(trick, input.Trick)
			trick = append(trick, game.Play{Seat: input.MySeat, Card: candidate})

			heartsBroken := input.HeartsBroken || candidate.Suit == game.SuitHearts
			roundPoints := input.RoundPoints
			nextSeat := (input.MySeat + 1) % game.PlayersPerTable
			tn := trickNumber
			ft := input.FirstTrick

			if len(trick) == game.PlayersPerTable {
				winner, pts, _ := game.TrickWinner(trick)
				roundPoints[winner] += pts
				trick = nil
				tn++
				ft = false
				nextSeat = winner
			}

			final := simulateRemaining(mcState{
				hands:        hands,
				heartsBroken: heartsBroken,
				trickNumber:  tn,
				roundPoints:  roundPoints,
				currentTrick: trick,
				turnSeat:     nextSeat,
				firstTrick:   ft,
			})
			adjusted := game.ApplyShootTheMoon(final)
			total += float64(adjusted[input.MySeat])
		}

		avg := total / float64(mc.samples)
		if avg < bestAvg {
			bestAvg = avg
			bestCard = candidate
		}
	}

	return bestCard
}

// mcState holds the state needed to simulate remaining tricks.
type mcState struct {
	hands        [game.PlayersPerTable][]game.Card
	heartsBroken bool
	trickNumber  int
	roundPoints  [game.PlayersPerTable]game.Points
	currentTrick []game.Play
	turnSeat     int
	firstTrick   bool
}

// simulateRemaining plays out all remaining tricks using easy-bot heuristics
// and returns the final round points (before shoot-the-moon adjustment).
func simulateRemaining(state mcState) [game.PlayersPerTable]game.Points {
	for {
		seat := state.turnSeat
		if len(state.hands[seat]) == 0 && len(state.currentTrick) == 0 {
			break
		}

		trick := game.CardsFrom(state.currentTrick)
		legal := game.LegalPlays(state.hands[seat], trick, state.heartsBroken, state.firstTrick)
		if len(legal) == 0 {
			break
		}

		var card game.Card
		if len(state.currentTrick) == 0 {
			card = chooseSmartLead(state.hands[seat], legal)
		} else {
			leadSuit := state.currentTrick[0].Card.Suit
			if game.HasSuit(state.hands[seat], leadSuit) {
				card = chooseSmartFollow(trick, legal)
			} else {
				card = chooseSmartDiscard(legal)
			}
		}

		state.hands[seat], _ = game.RemoveCard(state.hands[seat], card)
		if card.Suit == game.SuitHearts {
			state.heartsBroken = true
		}
		state.currentTrick = append(state.currentTrick, game.Play{Seat: seat, Card: card})

		if len(state.currentTrick) == game.PlayersPerTable {
			winner, points, _ := game.TrickWinner(state.currentTrick)
			state.roundPoints[winner] += points
			state.currentTrick = nil
			state.trickNumber++
			state.firstTrick = false
			state.turnSeat = winner
		} else {
			state.turnSeat = (state.turnSeat + 1) % game.PlayersPerTable
		}
	}

	return state.roundPoints
}

// detectSeatVoids returns per-seat void information by scanning completed
// tricks and the current in-progress trick for off-suit follows.
func detectSeatVoids(completedPlays []game.Play, currentTrick []game.Play) [game.PlayersPerTable]map[game.Suit]bool {
	var voids [game.PlayersPerTable]map[game.Suit]bool

	addVoid := func(seat int, suit game.Suit) {
		if voids[seat] == nil {
			voids[seat] = make(map[game.Suit]bool)
		}
		voids[seat][suit] = true
	}

	for i := 0; i+3 < len(completedPlays); i += game.PlayersPerTable {
		leadSuit := completedPlays[i].Card.Suit
		for j := 1; j < game.PlayersPerTable; j++ {
			if completedPlays[i+j].Card.Suit != leadSuit {
				addVoid(completedPlays[i+j].Seat, leadSuit)
			}
		}
	}

	if len(currentTrick) > 1 {
		leadSuit := currentTrick[0].Card.Suit
		for _, p := range currentTrick[1:] {
			if p.Card.Suit != leadSuit {
				addVoid(p.Seat, leadSuit)
			}
		}
	}

	return voids
}

// sampleOpponentHands generates random opponent hands from remaining cards,
// respecting void constraints and required hand sizes. Returns false if
// dealing fails after max retries (rare with sparse constraints).
func sampleOpponentHands(rng *rand.Rand, mySeat int, remaining []game.Card, handSizes [game.PlayersPerTable]int, voids [game.PlayersPerTable]map[game.Suit]bool) ([game.PlayersPerTable][]game.Card, bool) {
	shuffled := make([]game.Card, len(remaining))

	for range 50 {
		copy(shuffled, remaining)
		rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

		var hands [game.PlayersPerTable][]game.Card
		sizes := handSizes

		ok := true
		for _, card := range shuffled {
			placed := false
			start := rng.Intn(game.PlayersPerTable)
			for d := range game.PlayersPerTable {
				seat := (start + d) % game.PlayersPerTable
				if seat == mySeat || sizes[seat] == 0 {
					continue
				}
				if voids[seat] != nil && voids[seat][card.Suit] {
					continue
				}
				hands[seat] = append(hands[seat], card)
				sizes[seat]--
				placed = true
				break
			}
			if !placed {
				ok = false
				break
			}
		}
		if ok {
			return hands, true
		}
	}

	return [game.PlayersPerTable][]game.Card{}, false
}
