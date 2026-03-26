package bot

import (
	"math/rand"
	"testing"

	"github.com/JHK/hearts/internal/game"
)

func BenchmarkMCEvaluate(b *testing.B) {
	mc := &mcEvaluator{
		rng:     rand.New(rand.NewSource(42)),
		samples: defaultMCSamples,
	}

	// Mid-game scenario: trick 5, following suit with multiple options.
	input := game.TurnInput{
		Hand: cards("KS", "7S", "QH", "3D", "9C", "4C", "JH", "6D"),
		Trick: []game.Play{
			{Seat: 1, Card: c("5S")},
		},
		HeartsBroken: true,
		FirstTrick:   false,
		PlayedCards: []game.Play{
			// 5 completed tricks (20 plays).
			{Seat: 0, Card: c("2C")}, {Seat: 1, Card: c("5C")}, {Seat: 2, Card: c("TC")}, {Seat: 3, Card: c("AC")},
			{Seat: 3, Card: c("KD")}, {Seat: 0, Card: c("8D")}, {Seat: 1, Card: c("2D")}, {Seat: 2, Card: c("AD")},
			{Seat: 2, Card: c("AS")}, {Seat: 3, Card: c("4S")}, {Seat: 0, Card: c("TS")}, {Seat: 1, Card: c("3S")},
			{Seat: 2, Card: c("KC")}, {Seat: 3, Card: c("7C")}, {Seat: 0, Card: c("8C")}, {Seat: 1, Card: c("JC")},
			{Seat: 1, Card: c("2H")}, {Seat: 2, Card: c("5H")}, {Seat: 3, Card: c("9H")}, {Seat: 0, Card: c("AH")},
		},
		RoundPoints: [4]game.Points{4, 0, 0, 0},
		MySeat:      0,
	}

	legal := game.LegalPlays(input.Hand, input.TrickCards(), input.HeartsBroken, input.FirstTrick)

	b.ResetTimer()
	for range b.N {
		mc.evaluate(input, legal)
	}
}
