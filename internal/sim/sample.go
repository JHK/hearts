package sim

import (
	"fmt"
	"math/rand"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
)

// GameLog captures the full history of a simulated game for analysis.
type GameLog struct {
	// Strategies maps strategy slot index to strategy name.
	Strategies [game.PlayersPerTable]string `json:"strategies"`
	// SeatToStrategy maps seat index to strategy slot index.
	SeatToStrategy [game.PlayersPerTable]int `json:"seat_to_strategy"`
	Rounds         []RoundLog                `json:"rounds"`
	FinalScores    [game.PlayersPerTable]int `json:"final_scores"`
	Winners        []int                     `json:"winners"` // strategy slot indices
}

// RoundLog captures the full history of a single round.
type RoundLog struct {
	RoundNumber   int                        `json:"round_number"`
	PassDirection string                     `json:"pass_direction"`
	DealtHands    [game.PlayersPerTable]Hand `json:"dealt_hands"`
	Passes        [game.PlayersPerTable]Hand `json:"passes,omitempty"`
	PostPassHands [game.PlayersPerTable]Hand `json:"post_pass_hands"`
	Tricks        []TrickLog                 `json:"tricks"`
	RawScores     [game.PlayersPerTable]int  `json:"raw_scores"`
	AdjScores     [game.PlayersPerTable]int  `json:"adj_scores"`
	CumScores     [game.PlayersPerTable]int  `json:"cum_scores"` // cumulative after this round
}

// TrickLog captures one trick.
type TrickLog struct {
	TrickNumber int      `json:"trick_number"`
	Plays       []string `json:"plays"` // "seat:card" e.g. "2:AH"
	Winner      int      `json:"winner"`
	Points      int      `json:"points"`
}

// Hand is a list of card strings for JSON output.
type Hand = []string

func cardStrs(cards []game.Card) []string {
	out := make([]string, len(cards))
	for i, c := range cards {
		out[i] = c.String()
	}
	return out
}

func playStrs(plays []game.Play) []string {
	out := make([]string, len(plays))
	for i, p := range plays {
		out[i] = fmt.Sprintf("%d:%s", p.Seat, p.Card)
	}
	return out
}

// runGameSampled runs a single game and returns the full game log.
func (s *Simulation) runGameSampled(rng *rand.Rand) GameLog {
	perm := rng.Perm(game.PlayersPerTable)
	var bots [game.PlayersPerTable]bot.Bot
	for stratIdx, seat := range perm {
		bots[seat] = s.strategies[stratIdx].NewBotWithOptions(s.botOpts)
	}

	seatToStrat := [game.PlayersPerTable]int{}
	for stratIdx, seat := range perm {
		seatToStrat[seat] = stratIdx
	}

	var gl GameLog
	for i, sk := range s.strategies {
		gl.Strategies[i] = string(sk)
	}
	gl.SeatToStrategy = seatToStrat

	g := game.NewGame()
	roundNum := 0

	for {
		hands := game.Deal(rng)
		passDir := g.NextPassDirection()
		round := game.NewRound(hands, passDir)

		rl := RoundLog{
			RoundNumber:   roundNum,
			PassDirection: string(passDir),
		}
		for i := range hands {
			rl.DealtHands[i] = cardStrs(hands[i])
		}

		if passDir != game.PassDirectionHold {
			for i := range bots {
				cards, err := bots[i].ChoosePass(round.PassInput(i, g.Scores()))
				if err != nil || len(cards) != 3 {
					cards = round.Hand(i)[:3]
				}
				_ = round.SubmitPass(i, cards)
				rl.Passes[i] = cardStrs(cards)
			}
			_ = round.ApplyPasses()
			for i := range bots {
				_ = round.MarkReady(i)
			}
		}

		for i := range bots {
			rl.PostPassHands[i] = cardStrs(round.Hand(i))
		}

		_ = round.StartPlaying()

		var currentTrickPlays []game.Play
		for round.Phase() == game.PhasePlaying {
			seat := round.TurnSeat()
			input := round.TurnInput(seat, g.Scores())
			legal := game.LegalPlays(input.Hand, input.TrickCards(), input.HeartsBroken, input.FirstTrick)

			card, err := bots[seat].ChoosePlay(input)
			if err != nil || !game.ContainsCard(legal, card) {
				card = legal[0]
			}

			currentTrickPlays = append(currentTrickPlays, game.Play{Seat: seat, Card: card})
			tr, _ := round.Play(seat, card)

			if tr != nil {
				tl := TrickLog{
					TrickNumber: tr.TrickNumber,
					Plays:       playStrs(currentTrickPlays),
					Winner:      tr.WinnerSeat,
					Points:      int(tr.Points),
				}
				rl.Tricks = append(rl.Tricks, tl)
				currentTrickPlays = nil
			}
		}

		scores := round.Scores()
		for i := range game.PlayersPerTable {
			rl.RawScores[i] = int(scores.Raw[i])
			rl.AdjScores[i] = int(scores.Adjusted[i])
		}

		g.AddRoundScores(scores.Adjusted)

		for i := range game.PlayersPerTable {
			rl.CumScores[i] = int(g.Score(i))
		}

		gl.Rounds = append(gl.Rounds, rl)
		roundNum++

		if g.IsOver() {
			seatWinners := g.Winners()
			gl.Winners = make([]int, len(seatWinners))
			for i, seat := range seatWinners {
				gl.Winners[i] = seatToStrat[seat]
			}
			for i := range game.PlayersPerTable {
				gl.FinalScores[i] = int(g.Score(i))
			}
			return gl
		}
	}
}
