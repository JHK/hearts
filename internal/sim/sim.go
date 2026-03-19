package sim

import (
	"fmt"
	"math/rand"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/player/bot"
)

const gameOverThreshold = game.Points(100)

// Result holds win counts per strategy slot (index 0–3).
// Ties count as a win for each tied player.
type Result struct {
	Wins      [game.PlayersPerTable]int
	MoonShots [game.PlayersPerTable]int
}

// Simulation runs N complete games between 4 fixed bot strategies.
type Simulation struct {
	strategies [game.PlayersPerTable]bot.Strategy
	iterations int
}

// New creates a Simulation that will run iterations games between the given strategies.
func New(strategies [game.PlayersPerTable]bot.Strategy, iterations int) *Simulation {
	return &Simulation{strategies: strategies, iterations: iterations}
}

// Run plays all games and returns cumulative win counts per slot.
func (s *Simulation) Run() Result {
	var result Result
	rng := rand.New(rand.NewSource(rand.Int63()))

	for i := 0; i < s.iterations; i++ {
		wins, moonShots := s.runGame(rng)
		for _, w := range wins {
			result.Wins[w]++
		}
		for slot, n := range moonShots {
			result.MoonShots[slot] += n
		}
	}

	return result
}

// runGame plays one complete game and returns the winning slot index/indices and
// moon-shot counts per slot for the game.
func (s *Simulation) runGame(rng *rand.Rand) ([]int, [game.PlayersPerTable]int) {
	var totals [game.PlayersPerTable]game.Points
	var moonShots [game.PlayersPerTable]int

	for roundIndex := 0; ; roundIndex++ {
		s.runRound(rng, &totals, roundIndex, &moonShots)

		for _, pts := range totals {
			if pts >= gameOverThreshold {
				return winners(totals), moonShots
			}
		}
	}
}

func (s *Simulation) runRound(rng *rand.Rand, totals *[game.PlayersPerTable]game.Points, roundIndex int, moonShots *[game.PlayersPerTable]int) {
	hands := game.Deal(rng)

	dir := game.PassDirectionForRound(roundIndex)
	if dir != game.PassDirectionHold {
		applyPasses(&hands, s.strategies, dir)
	}

	roundPoints := playRound(&hands, s.strategies)
	adjusted := game.ApplyShootTheMoon(roundPoints)
	for i := range game.PlayersPerTable {
		pid := game.PlayerID(fmt.Sprintf("p%d", i))
		if roundPoints[pid] == game.ShootTheMoonPoints {
			moonShots[i]++
		}
		totals[i] += adjusted[pid]
	}
}

// applyPasses asks each strategy which cards to pass, then delegates the exchange to the game.
func applyPasses(hands *[game.PlayersPerTable][]game.Card, strategies [game.PlayersPerTable]bot.Strategy, dir game.PassDirection) {
	var passes [game.PlayersPerTable][]game.Card
	for i, s := range strategies {
		passed, err := s.ChoosePass(bot.PassInput{Hand: hands[i], Direction: dir})
		if err != nil || len(passed) != 3 {
			passed = hands[i][:3]
		}
		passes[i] = passed
	}
	received := game.ExchangePasses(passes, dir)
	for dst, cards := range received {
		for _, card := range cards {
			hands[dst] = append(hands[dst], card)
		}
	}
	for src, cards := range passes {
		for _, card := range cards {
			hands[src], _ = game.RemoveCard(hands[src], card)
		}
	}
}

// playRound plays all 13 tricks and returns the raw round points per player.
func playRound(hands *[game.PlayersPerTable][]game.Card, strategies [game.PlayersPerTable]bot.Strategy) map[game.PlayerID]game.Points {
	roundPoints := map[game.PlayerID]game.Points{
		"p0": 0, "p1": 0, "p2": 0, "p3": 0,
	}

	twoClubs := game.Card{Suit: game.SuitClubs, Rank: 2}
	currentSeat := 0
	for i, hand := range hands {
		if game.ContainsCard(hand, twoClubs) {
			currentSeat = i
			break
		}
	}

	heartsBroken := false
	var playedCards []game.Card

	for trick := 0; trick < 13; trick++ {
		firstTrick := trick == 0
		var plays []game.Play

		for p := 0; p < game.PlayersPerTable; p++ {
			seat := (currentSeat + p) % game.PlayersPerTable
			pid := game.PlayerID(fmt.Sprintf("p%d", seat))
			trickCards := playsToCards(plays)

			legal := game.LegalPlays(hands[seat], trickCards, heartsBroken, firstTrick)

			card, err := strategies[seat].ChoosePlay(bot.TurnInput{
				Hand:         hands[seat],
				Trick:        trickCards,
				HeartsBroken: heartsBroken,
				FirstTrick:   firstTrick,
				PlayedCards:  playedCards,
			})
			if err != nil || !game.ContainsCard(legal, card) {
				card = legal[0]
			}

			hands[seat], _ = game.RemoveCard(hands[seat], card)
			plays = append(plays, game.Play{PlayerID: pid, Card: card})

			if card.Suit == game.SuitHearts {
				heartsBroken = true
			}
		}

		winnerID, points, _ := game.TrickWinner(plays)
		roundPoints[winnerID] += points

		for _, p := range plays {
			playedCards = append(playedCards, p.Card)
		}

		for i := range game.PlayersPerTable {
			if game.PlayerID(fmt.Sprintf("p%d", i)) == winnerID {
				currentSeat = i
				break
			}
		}
	}

	return roundPoints
}

func playsToCards(plays []game.Play) []game.Card {
	cards := make([]game.Card, len(plays))
	for i, p := range plays {
		cards[i] = p.Card
	}
	return cards
}

func winners(totals [game.PlayersPerTable]game.Points) []int {
	min := totals[0]
	for _, pts := range totals {
		if pts < min {
			min = pts
		}
	}
	var out []int
	for i, pts := range totals {
		if pts == min {
			out = append(out, i)
		}
	}
	return out
}
