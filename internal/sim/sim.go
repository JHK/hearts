package sim

import (
	"math/rand"
	"runtime"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
)

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

// Run plays all games concurrently using a worker pool and returns cumulative win counts per slot.
func (s *Simulation) Run() Result {
	workers := runtime.NumCPU()
	perWorker := s.iterations / workers
	remainder := s.iterations % workers

	results := make(chan Result, workers)
	for w := 0; w < workers; w++ {
		n := perWorker
		if w < remainder {
			n++
		}
		go s.runWorker(n, results)
	}

	var total Result
	for range workers {
		partial := <-results
		for i := range total.Wins {
			total.Wins[i] += partial.Wins[i]
			total.MoonShots[i] += partial.MoonShots[i]
		}
	}
	return total
}

func (s *Simulation) runWorker(n int, results chan<- Result) {
	rng := rand.New(rand.NewSource(rand.Int63()))
	var result Result
	for range n {
		wins, moonShots := s.runGame(rng)
		for _, w := range wins {
			result.Wins[w]++
		}
		for slot, count := range moonShots {
			result.MoonShots[slot] += count
		}
	}
	results <- result
}

func (s *Simulation) runGame(rng *rand.Rand) ([]int, [game.PlayersPerTable]int) {
	// Randomly permute strategy-to-seat assignment each game to eliminate
	// positional bias from fixed neighbor relationships (passing, trick order).
	perm := rng.Perm(game.PlayersPerTable)
	var bots [game.PlayersPerTable]bot.Bot
	for stratIdx, seat := range perm {
		bots[seat] = s.strategies[stratIdx].NewBot()
	}
	g := game.NewGame()
	var moonShots [game.PlayersPerTable]int

	for {
		hands := game.Deal(rng)
		passDir := g.NextPassDirection()
		round := game.NewRound(hands, passDir)

		if passDir != game.PassDirectionHold {
			for i := range bots {
				cards, err := bots[i].ChoosePass(round.PassInput(i, g.Scores()))
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
			input := round.TurnInput(seat, g.Scores())
			legal := game.LegalPlays(input.Hand, input.TrickCards(), input.HeartsBroken, input.FirstTrick)

			card, err := bots[seat].ChoosePlay(input)
			if err != nil || !game.ContainsCard(legal, card) {
				card = legal[0]
			}
			_, _ = round.Play(seat, card)
		}

		scores := round.Scores()
		for i := range game.PlayersPerTable {
			if scores.Raw[i] == game.ShootTheMoonPoints {
				moonShots[i]++
			}
		}

		if g.AddRoundScores(scores.Adjusted) {
			// Map seat-based winners back to strategy indices.
			seatToStrat := [game.PlayersPerTable]int{}
			for stratIdx, seat := range perm {
				seatToStrat[seat] = stratIdx
			}
			seatWinners := g.Winners()
			stratWinners := make([]int, len(seatWinners))
			for i, seat := range seatWinners {
				stratWinners[i] = seatToStrat[seat]
			}
			var stratMoonShots [game.PlayersPerTable]int
			for seat, count := range moonShots {
				stratMoonShots[seatToStrat[seat]] += count
			}
			return stratWinners, stratMoonShots
		}
	}
}
