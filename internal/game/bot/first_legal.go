package bot

import "github.com/JHK/hearts/internal/game"

type FirstLegal struct{}

var firstLegalBotNames = []string{"Fritz"}

func (f *FirstLegal) Kind() StrategyKind { return StrategyFirstLegal }

// NewFirstLegalBot creates a first-legal bot for testing.
func NewFirstLegalBot() *FirstLegal {
	return &FirstLegal{}
}

func (f *FirstLegal) ChoosePlay(input game.TurnInput) (game.Card, error) {
	for _, card := range input.Hand {
		err := game.ValidatePlay(game.ValidatePlayInput{
			Hand:         input.Hand,
			Card:         card,
			Trick:        input.TrickCards(),
			HeartsBroken: input.HeartsBroken,
			FirstTrick:   input.FirstTrick,
		})
		if err == nil {
			return card, nil
		}
	}

	return game.Card{}, ErrNoLegalPlays
}

func (f *FirstLegal) ChoosePass(input game.PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, ErrNotEnoughCards
	}

	return append([]game.Card(nil), input.Hand[:3]...), nil
}
