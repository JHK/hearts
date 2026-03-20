package bot

import (
	"fmt"

	"github.com/JHK/hearts/internal/game"
)

type FirstLegal struct {
	*game.Player
}

func (f *FirstLegal) Kind() StrategyKind   { return StrategyFirstLegal }
func (f *FirstLegal) BotName() string      { return "Fritz" }
func (f *FirstLegal) Unwrap() *game.Player { return f.Player }

// NewFirstLegalBot creates a first-legal bot for testing.
func NewFirstLegalBot() *FirstLegal {
	return &FirstLegal{Player: game.NewPlayer()}
}

func (f *FirstLegal) ChoosePlay(input TurnInput) (game.Card, error) {
	for _, card := range input.Hand {
		err := game.ValidatePlay(game.ValidatePlayInput{
			Hand:         input.Hand,
			Card:         card,
			Trick:        input.Trick,
			HeartsBroken: input.HeartsBroken,
			FirstTrick:   input.FirstTrick,
		})
		if err == nil {
			return card, nil
		}
	}

	return game.Card{}, fmt.Errorf("no legal plays")
}

func (f *FirstLegal) ChoosePass(input PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, fmt.Errorf("not enough cards to pass")
	}

	return append([]game.Card(nil), input.Hand[:3]...), nil
}
