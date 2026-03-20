package sim

import (
	"math/rand"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
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
	strategies [game.PlayersPerTable]bot.StrategyKind
	iterations int
}

// New creates a Simulation that will run iterations games between the given strategies.
func New(strategies [game.PlayersPerTable]bot.StrategyKind, iterations int) *Simulation {
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
	var players [game.PlayersPerTable]bot.Bot
	for i, strategy := range s.strategies {
		players[i] = strategy.NewBot()
	}
	var moonShots [game.PlayersPerTable]int

	for roundIndex := 0; ; roundIndex++ {
		s.runRound(rng, players, roundIndex, &moonShots)

		for _, p := range players {
			if p.CumulativePoints() >= gameOverThreshold {
				return winners(players), moonShots
			}
		}
	}
}

func (s *Simulation) runRound(rng *rand.Rand, players [game.PlayersPerTable]bot.Bot, roundIndex int, moonShots *[game.PlayersPerTable]int) {
	hands := game.Deal(rng)
	for i, p := range players {
		p.DealCards(hands[i])
	}

	dir := game.PassDirectionForRound(roundIndex)
	if dir != game.PassDirectionHold {
		applyPasses(players, dir)
	}

	playRound(players)

	var rawPoints [game.PlayersPerTable]game.Points
	for i, p := range players {
		rawPoints[i] = p.RoundPoints()
	}
	adjusted := game.ApplyShootTheMoon(rawPoints)
	for i, p := range players {
		if p.RoundPoints() == game.ShootTheMoonPoints {
			moonShots[i]++
		}
		p.FinalizeRound(adjusted[i])
	}
}

// applyPasses asks each bot which cards to pass, then delegates the exchange to the game.
func applyPasses(players [game.PlayersPerTable]bot.Bot, dir game.PassDirection) {
	var passes [game.PlayersPerTable][]game.Card
	for i, p := range players {
		passed, err := p.ChoosePass(bot.PassInput{Hand: p.Hand(), Direction: dir})
		if err != nil || len(passed) != 3 {
			passed = p.Hand()[:3]
		}
		p.SubmitPass(passed)
		passes[i] = p.PassSent()
	}
	received := game.ExchangePasses(passes, dir)
	for i, p := range players {
		p.SendPassCards()
		p.ReceivePassCards(received[i])
	}
}

// playRound plays all 13 tricks, accumulating trick points on each player.
func playRound(players [game.PlayersPerTable]bot.Bot) {
	twoClubs := game.Card{Suit: game.SuitClubs, Rank: 2}
	currentSeat := 0
	for i, p := range players {
		if game.ContainsCard(p.Hand(), twoClubs) {
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
			player := players[seat]
			trickCards := playsToCards(plays)

			legal := game.LegalPlays(player.Hand(), trickCards, heartsBroken, firstTrick)

			card, err := player.ChoosePlay(bot.TurnInput{
				Hand:         player.Hand(),
				Trick:        trickCards,
				HeartsBroken: heartsBroken,
				FirstTrick:   firstTrick,
				PlayedCards:  playedCards,
			})
			if err != nil || !game.ContainsCard(legal, card) {
				card = legal[0]
			}

			player.PlayCard(card)
			plays = append(plays, game.Play{Player: player, Card: card})

			if card.Suit == game.SuitHearts {
				heartsBroken = true
			}
		}

		winner, points, _ := game.TrickWinner(plays)
		winner.AddTrickPoints(points)

		for _, p := range plays {
			playedCards = append(playedCards, p.Card)
		}

		for i, p := range players {
			if p == winner {
				currentSeat = i
				break
			}
		}
	}
}

func playsToCards(plays []game.Play) []game.Card {
	cards := make([]game.Card, len(plays))
	for i, p := range plays {
		cards[i] = p.Card
	}
	return cards
}

func winners(players [game.PlayersPerTable]bot.Bot) []int {
	min := players[0].CumulativePoints()
	for _, p := range players {
		if p.CumulativePoints() < min {
			min = p.CumulativePoints()
		}
	}
	var out []int
	for i, p := range players {
		if p.CumulativePoints() == min {
			out = append(out, i)
		}
	}
	return out
}
