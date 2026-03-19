package bot

import "github.com/JHK/hearts/internal/game"

type TurnInput struct {
	Hand         []game.Card
	Trick        []game.Card
	HeartsBroken bool
	FirstTrick   bool
	PlayedCards  []game.Card // all cards played in completed tricks this round
}

type PassInput struct {
	Hand      []game.Card
	Direction game.PassDirection
}

type Strategy interface {
	Kind() StrategyKind
	BotName() string
	ChoosePlay(input TurnInput) (game.Card, error)
	ChoosePass(input PassInput) ([]game.Card, error)
}
