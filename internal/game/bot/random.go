package bot

import (
	"math/rand"
	"time"

	"github.com/JHK/hearts/internal/game"
)

type Random struct {
	rng *rand.Rand
}

var randomBotNames = []string{"Lucky", "Dice", "Chance", "Jinx", "Hazard", "Wild", "Shuffle", "Rando"}

func (r *Random) Kind() StrategyKind { return StrategyRandom }


func newRandomBot(rng *rand.Rand) *Random {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &Random{rng: rng}
}

// NewRandomBot creates a random bot for testing with a seeded RNG.
func NewRandomBot(rng *rand.Rand) *Random {
	return newRandomBot(rng)
}

func (r *Random) ChoosePlay(input game.TurnInput) (game.Card, error) {
	legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, ErrNoLegalPlays
	}

	return legal[r.rng.Intn(len(legal))], nil
}

func (r *Random) ChoosePass(input game.PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, ErrNotEnoughCards
	}

	perm := r.rng.Perm(len(input.Hand))
	selected := make([]game.Card, 0, 3)
	for i := 0; i < 3; i++ {
		selected = append(selected, input.Hand[perm[i]])
	}

	return selected, nil
}
