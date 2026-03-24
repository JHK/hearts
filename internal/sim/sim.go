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

func (s *Simulation) runGame(rng *rand.Rand) ([]int, [game.PlayersPerTable]int) {
	var bots [game.PlayersPerTable]bot.Bot
	for i, strategy := range s.strategies {
		bots[i] = strategy.NewBot()
	}
	var cumulative [game.PlayersPerTable]game.Points
	var moonShots [game.PlayersPerTable]int

	for roundIndex := 0; ; roundIndex++ {
		hands := game.Deal(rng)
		passDir := game.PassDirectionForRound(roundIndex)
		round := game.NewRound(hands, passDir)

		if passDir != game.PassDirectionHold {
			for i := range bots {
				cards, err := bots[i].ChoosePass(round.PassInput(i))
				if err != nil || len(cards) != 3 {
					cards = round.Hand(i)[:3]
				}
				_ = round.SubmitPass(i, cards)
			}
			_ = round.ApplyPasses()
			for i := range bots {
				_ = round.MarkReady(i)
			}
		}
		_ = round.StartPlaying()

		for round.Phase() == game.PhasePlaying {
			seat := round.TurnSeat()
			input := round.TurnInput(seat)
			legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)

			card, err := bots[seat].ChoosePlay(input)
			if err != nil || !game.ContainsCard(legal, card) {
				card = legal[0]
			}
			_, _ = round.Play(seat, card)
		}

		scores := round.Scores()
		for i := range cumulative {
			if scores.Raw[i] == game.ShootTheMoonPoints {
				moonShots[i]++
			}
			cumulative[i] += scores.Adjusted[i]
		}

		for _, pts := range cumulative {
			if pts >= gameOverThreshold {
				return winners(cumulative), moonShots
			}
		}
	}
}

func winners(cumulative [game.PlayersPerTable]game.Points) []int {
	min := cumulative[0]
	for _, pts := range cumulative {
		if pts < min {
			min = pts
		}
	}
	var out []int
	for i, pts := range cumulative {
		if pts == min {
			out = append(out, i)
		}
	}
	return out
}
