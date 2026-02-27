package bot

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JHK/hearts/internal/game"
)

type Random struct {
	rng *rand.Rand
}

func NewRandomBot(rng *rand.Rand) Strategy {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &Random{rng: rng}
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
