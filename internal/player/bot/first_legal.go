package bot

import (
	"fmt"

	"github.com/JHK/hearts/internal/game"
)

type FirstLegal struct{}

func NewFirstLegalBot() Strategy {
	return &FirstLegal{}
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
