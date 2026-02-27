package bot

import "github.com/JHK/hearts/internal/game"

type TurnInput struct {
	Hand         []game.Card
	Trick        []game.Card
	HeartsBroken bool
	FirstTrick   bool
}

type Strategy interface {
	ChoosePlay(input TurnInput) (game.Card, error)
}
