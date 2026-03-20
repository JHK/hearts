package bot

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JHK/hearts/internal/game"
)

type Random struct {
	*game.Player
	rng *rand.Rand
}

var randomBotNames = []string{"Lucky", "Dice", "Chance", "Jinx", "Hazard", "Wild", "Shuffle", "Rando"}

func (r *Random) Kind() StrategyKind   { return StrategyRandom }
func (r *Random) BotName() string      { return randomFrom(randomBotNames) }
func (r *Random) Unwrap() *game.Player { return r.Player }

func newRandomBot(p *game.Player, rng *rand.Rand) *Random {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &Random{Player: p, rng: rng}
}

// NewRandomBot creates a random bot for testing with a seeded RNG.
func NewRandomBot(rng *rand.Rand) *Random {
	return newRandomBot(game.NewPlayer(), rng)
}

func (r *Random) ChoosePlay(input TurnInput) (game.Card, error) {
	legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, fmt.Errorf("no legal plays")
	}

	return legal[r.rng.Intn(len(legal))], nil
}

func (r *Random) ChoosePass(input PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, fmt.Errorf("not enough cards to pass")
	}

	perm := r.rng.Perm(len(input.Hand))
	selected := make([]game.Card, 0, 3)
	for i := 0; i < 3; i++ {
		selected = append(selected, input.Hand[perm[i]])
	}

	return selected, nil
}
